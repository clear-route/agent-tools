package meetings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"teams-assistant/graph"
)

const cacheFile = ".teams-assistant-meeting-cache.json"

type CachedMeeting struct {
	JoinURL         string `json:"joinUrl"`
	OnlineMeetingID string `json:"onlineMeetingId,omitempty"`
	Subject         string `json:"subject"`
}

func cachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, cacheFile), nil
}

func loadCache() ([]CachedMeeting, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	var cache []CachedMeeting
	return cache, json.Unmarshal(b, &cache)
}

func saveCache(meetings []CachedMeeting) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	b, err := json.Marshal(meetings)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func ResolveRef(ref string) (CachedMeeting, error) {
	if idx, err := strconv.Atoi(ref); err == nil {
		cache, cacheErr := loadCache()
		if cacheErr != nil || cache == nil {
			return CachedMeeting{}, fmt.Errorf("no meeting cache found — run 'list' first")
		}
		if idx < 1 || idx > len(cache) {
			return CachedMeeting{}, fmt.Errorf("index %d out of range (1–%d)", idx, len(cache))
		}
		return cache[idx-1], nil
	}
	return CachedMeeting{OnlineMeetingID: ref}, nil
}

func ResolveMeetingID(ctx context.Context, client *graph.Client, m *CachedMeeting) error {
	if m.OnlineMeetingID != "" {
		return nil
	}
	if m.JoinURL == "" {
		return fmt.Errorf("meeting has no join URL and no online meeting ID")
	}

	escaped := strings.ReplaceAll(m.JoinURL, "'", "''")
	filter := fmt.Sprintf("JoinWebUrl eq '%s'", escaped)
	params := url.Values{"$filter": {filter}}
	path := "/me/onlineMeetings?" + params.Encode()

	var result struct {
		Value []struct {
			ID string `json:"id"`
		} `json:"value"`
	}
	if err := client.GetJSON(ctx, path, &result); err != nil {
		return fmt.Errorf("resolving meeting ID: %w", err)
	}
	if len(result.Value) == 0 {
		return fmt.Errorf("no online meeting found for join URL: %s", m.JoinURL)
	}
	m.OnlineMeetingID = result.Value[0].ID
	return nil
}

type MeetingSummary struct {
	Index         int    `json:"index"`
	Subject       string `json:"subject"`
	Start         string `json:"start"`
	End           string `json:"end"`
	Organizer     string `json:"organizer"`
	JoinURL       string `json:"joinUrl,omitempty"`
	HasTranscript bool   `json:"hasTranscript"`
}

func List(ctx context.Context, client *graph.Client, count int32, since, before string, jsonOut bool) error {
	now := time.Now()
	startDT := now.AddDate(0, 0, -30)
	endDT := now.AddDate(0, 0, 30)

	if since != "" {
		t, err := parseDate(since)
		if err != nil {
			return err
		}
		startDT = t
	}
	if before != "" {
		t, err := parseDate(before)
		if err != nil {
			return err
		}
		endDT = t
	}

	if count < 1 {
		count = 1
	} else if count > 200 {
		count = 200
	}
	fetch := count * 3
	if fetch < 50 {
		fetch = 50
	}
	path := fmt.Sprintf(
		"/me/calendarView?startDateTime=%s&endDateTime=%s&$select=subject,start,end,organizer,isOnlineMeeting,onlineMeeting&$top=%d&$orderby=start/dateTime desc",
		url.QueryEscape(startDT.UTC().Format(time.RFC3339)),
		url.QueryEscape(endDT.UTC().Format(time.RFC3339)),
		fetch,
	)

	summaries, cached, err := fetchMeetings(ctx, client, path, count)
	if err != nil {
		return err
	}

	checkTranscripts(ctx, client, summaries, cached)

	if err := saveCache(cached); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save meeting cache: %v\n", err)
	}

	return printMeetings(client, summaries, jsonOut, "Teams meeting")
}

func Search(ctx context.Context, client *graph.Client, query string, count int32, since, before string, jsonOut bool) error {
	now := time.Now()
	startDT := now.AddDate(0, -3, 0)
	endDT := now.AddDate(0, 0, 30)

	if since != "" {
		t, err := parseDate(since)
		if err != nil {
			return err
		}
		startDT = t
	}
	if before != "" {
		t, err := parseDate(before)
		if err != nil {
			return err
		}
		endDT = t
	}

	path := fmt.Sprintf(
		"/me/calendarView?startDateTime=%s&endDateTime=%s&$select=subject,start,end,organizer,isOnlineMeeting,onlineMeeting&$top=50&$orderby=start/dateTime desc",
		url.QueryEscape(startDT.UTC().Format(time.RFC3339)),
		url.QueryEscape(endDT.UTC().Format(time.RFC3339)),
	)

	allSummaries, _, err := fetchMeetings(ctx, client, path, 500)
	if err != nil {
		return err
	}

	lowerQuery := strings.ToLower(query)
	filtered := make([]MeetingSummary, 0)
	cached := make([]CachedMeeting, 0)
	idx := 1
	for _, m := range allSummaries {
		if strings.Contains(strings.ToLower(m.Subject), lowerQuery) {
			m.Index = idx
			filtered = append(filtered, m)
			cached = append(cached, CachedMeeting{JoinURL: m.JoinURL, Subject: m.Subject})
			idx++
			if int32(idx-1) >= count {
				break
			}
		}
	}

	checkTranscripts(ctx, client, filtered, cached)

	if err := saveCache(cached); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save meeting cache: %v\n", err)
	}

	return printMeetings(client, filtered, jsonOut, "matching meeting")
}

func checkTranscripts(ctx context.Context, client *graph.Client, summaries []MeetingSummary, cached []CachedMeeting) {
	if len(summaries) == 0 {
		return
	}

	const maxWorkers = 5
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for i := range summaries {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			m := &cached[idx]
			if err := ResolveMeetingID(ctx, client, m); err != nil {
				return
			}

			cached[idx].OnlineMeetingID = m.OnlineMeetingID

			meetingID := url.PathEscape(m.OnlineMeetingID)
			path := fmt.Sprintf("/me/onlineMeetings/%s/transcripts?$top=1", meetingID)
			var result struct {
				Value []struct {
					ID string `json:"id"`
				} `json:"value"`
			}
			if err := client.GetJSON(ctx, path, &result); err != nil {
				return
			}
			summaries[idx].HasTranscript = len(result.Value) > 0
		}(i)
	}
	wg.Wait()
}

func Transcript(ctx context.Context, client *graph.Client, meetingRef string, jsonOut bool) error {
	meeting, err := ResolveRef(meetingRef)
	if err != nil {
		return err
	}
	if err := ResolveMeetingID(ctx, client, &meeting); err != nil {
		return err
	}

	meetingID := url.PathEscape(meeting.OnlineMeetingID)
	listPath := fmt.Sprintf("/me/onlineMeetings/%s/transcripts?$top=1", meetingID)
	var listResult struct {
		Value []struct {
			ID string `json:"id"`
		} `json:"value"`
	}
	if err := client.GetJSON(ctx, listPath, &listResult); err != nil {
		return fmt.Errorf("listing transcripts: %w", err)
	}
	if len(listResult.Value) == 0 {
		return fmt.Errorf("no transcripts found for this meeting")
	}
	transcriptID := listResult.Value[0].ID

	path := fmt.Sprintf("/me/onlineMeetings/%s/transcripts/%s/content?$format=text/vtt", meetingID, url.PathEscape(transcriptID))

	body, err := client.GetBody(ctx, path)
	if err != nil {
		return fmt.Errorf("fetching transcript content: %w", err)
	}

	utterances := parseVTT(string(body))
	if len(utterances) == 0 {
		fmt.Fprintln(os.Stderr, "Transcript is empty.")
		return nil
	}

	speakerLines := map[string]int{}
	for _, u := range utterances {
		speakerLines[u.Speaker]++
	}
	type speakerStat struct {
		Name       string `json:"name"`
		Utterances int    `json:"utterances"`
	}
	var speakers []speakerStat
	for name, count := range speakerLines {
		speakers = append(speakers, speakerStat{Name: name, Utterances: count})
	}
	sort.Slice(speakers, func(i, j int) bool {
		return speakers[i].Utterances > speakers[j].Utterances
	})

	duration := ""
	if utterances[len(utterances)-1].EndSec > 0 {
		totalSec := int(utterances[len(utterances)-1].EndSec)
		duration = fmt.Sprintf("%dm%ds", totalSec/60, totalSec%60)
	}

	const chunkSeconds = 600
	type chunkInfo struct {
		Index    int    `json:"index"`
		Path     string `json:"path"`
		TimeFrom string `json:"timeFrom"`
		TimeTo   string `json:"timeTo"`
		Lines    int    `json:"lines"`
	}

	tmpDir, err := os.MkdirTemp("", "teams-transcript-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}

	var chunks []chunkInfo
	chunkIdx := 0
	chunkStart := 0.0
	var chunkBuf strings.Builder
	chunkLines := 0
	chunkTimeFrom := formatSeconds(0)

	flushChunk := func(timeTo string) error {
		if chunkBuf.Len() == 0 {
			return nil
		}
		chunkIdx++
		fpath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%02d.txt", chunkIdx))
		if err := os.WriteFile(fpath, []byte(chunkBuf.String()), 0600); err != nil {
			return err
		}
		chunks = append(chunks, chunkInfo{
			Index:    chunkIdx,
			Path:     fpath,
			TimeFrom: chunkTimeFrom,
			TimeTo:   timeTo,
			Lines:    chunkLines,
		})
		chunkBuf.Reset()
		chunkLines = 0
		return nil
	}

	for _, u := range utterances {

		if u.StartSec >= chunkStart+chunkSeconds && chunkBuf.Len() > 0 {
			if err := flushChunk(formatSeconds(u.StartSec)); err != nil {
				return err
			}
			chunkStart = u.StartSec
			chunkTimeFrom = formatSeconds(u.StartSec)
		}
		fmt.Fprintf(&chunkBuf, "[%s --> %s] %s: %s\n",
			formatSeconds(u.StartSec), formatSeconds(u.EndSec), u.Speaker, u.Text)
		chunkLines++
	}

	if err := flushChunk(formatSeconds(utterances[len(utterances)-1].EndSec)); err != nil {
		return err
	}

	if jsonOut {
		return writeJSON(client, map[string]any{
			"meetingId":    meeting.OnlineMeetingID,
			"transcriptId": transcriptID,
			"duration":     duration,
			"speakers":     speakers,
			"totalLines":   len(utterances),
			"chunks":       chunks,
		})
	}

	fmt.Fprintf(os.Stderr, "Transcript: %d utterances, %d speakers, %d chunks\n",
		len(utterances), len(speakers), len(chunks))
	for _, c := range chunks {
		fmt.Printf("  chunk %d: %s -> %s (%d lines) %s\n", c.Index, c.TimeFrom, c.TimeTo, c.Lines, c.Path)
	}
	return nil
}

type calendarPage struct {
	NextLink string `json:"@odata.nextLink"`
	Value    []struct {
		Subject         string `json:"subject"`
		IsOnlineMeeting bool   `json:"isOnlineMeeting"`
		Start           struct {
			DateTime string `json:"dateTime"`
		} `json:"start"`
		End struct {
			DateTime string `json:"dateTime"`
		} `json:"end"`
		Organizer struct {
			EmailAddress struct {
				Name string `json:"name"`
			} `json:"emailAddress"`
		} `json:"organizer"`
		OnlineMeeting struct {
			JoinURL string `json:"joinUrl"`
		} `json:"onlineMeeting"`
	} `json:"value"`
}

func fetchMeetings(ctx context.Context, client *graph.Client, path string, limit int32) ([]MeetingSummary, []CachedMeeting, error) {
	summaries := make([]MeetingSummary, 0)
	cached := make([]CachedMeeting, 0)
	idx := 0

	const maxPages = 50
	currentPath := path
	for pageNum := 0; currentPath != "" && pageNum < maxPages; pageNum++ {
		var page calendarPage
		if err := client.GetJSON(ctx, currentPath, &page); err != nil {
			return nil, nil, err
		}

		for _, ev := range page.Value {
			if !ev.IsOnlineMeeting || ev.OnlineMeeting.JoinURL == "" {
				continue
			}
			if int32(idx) >= limit {
				return summaries, cached, nil
			}
			idx++
			cached = append(cached, CachedMeeting{
				JoinURL: ev.OnlineMeeting.JoinURL,
				Subject: ev.Subject,
			})
			summaries = append(summaries, MeetingSummary{
				Index:     idx,
				Subject:   ev.Subject,
				Start:     ev.Start.DateTime,
				End:       ev.End.DateTime,
				Organizer: ev.Organizer.EmailAddress.Name,
				JoinURL:   ev.OnlineMeeting.JoinURL,
			})
		}

		currentPath = page.NextLink
	}
	return summaries, cached, nil
}

func writeJSON(client *graph.Client, data any) error {
	me, _ := client.Me()
	out := struct {
		User graph.UserProfile `json:"user"`
		Data any               `json:"data"`
	}{User: me, Data: data}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printMeetings(client *graph.Client, summaries []MeetingSummary, jsonOut bool, label string) error {
	if jsonOut {
		if summaries == nil {
			summaries = make([]MeetingSummary, 0)
		}
		return writeJSON(client, summaries)
	}

	if len(summaries) == 0 {
		fmt.Fprintf(os.Stderr, "No %ss found in the date range.\n", label)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d %s(s)\n\n", len(summaries), label)
	for _, m := range summaries {
		tr := " "
		if m.HasTranscript {
			tr = "T"
		}
		fmt.Printf("%3d %s %-50s  %s  %s\n", m.Index, tr, truncate(m.Subject, 50), formatDT(m.Start), m.Organizer)
	}
	return nil
}

func parseDate(s string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04", "2006-01-02T15:04"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date %q (expected YYYY-MM-DD or YYYY-MM-DD HH:MM)", s)
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 3 {
		return "..."
	}
	return string(r[:max-3]) + "..."
}

func formatDT(s string) string {
	for _, layout := range []string{
		time.RFC3339Nano, time.RFC3339,
		"2006-01-02T15:04:05.0000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:00",
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Local().Format("2006-01-02 15:04")
		}
	}
	return s
}

type utterance struct {
	StartSec float64
	EndSec   float64
	Speaker  string
	Text     string
}

func parseVTT(vtt string) []utterance {
	var out []utterance
	var curStart, curEnd float64

	for _, line := range strings.Split(vtt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "WEBVTT" {
			continue
		}

		if strings.Contains(line, " --> ") {
			parts := strings.Split(line, " --> ")
			if len(parts) == 2 {
				curStart = parseTimestamp(parts[0])
				curEnd = parseTimestamp(parts[1])
			}
			continue
		}

		if strings.HasPrefix(line, "<v ") {
			closeTag := strings.Index(line, ">")
			if closeTag > 3 {
				speaker := line[3:closeTag]
				text := strings.TrimSuffix(line[closeTag+1:], "</v>")
				out = append(out, utterance{
					StartSec: curStart,
					EndSec:   curEnd,
					Speaker:  speaker,
					Text:     text,
				})
			}
			continue
		}

	}
	return out
}

func parseTimestamp(ts string) float64 {
	ts = strings.TrimSpace(ts)

	sec := 0.0
	if dot := strings.Index(ts, "."); dot != -1 {
		ms, _ := strconv.ParseFloat("0"+ts[dot:], 64)
		sec = ms
		ts = ts[:dot]
	}
	parts := strings.Split(ts, ":")
	if len(parts) == 3 {
		h, _ := strconv.ParseFloat(parts[0], 64)
		m, _ := strconv.ParseFloat(parts[1], 64)
		s, _ := strconv.ParseFloat(parts[2], 64)
		sec += h*3600 + m*60 + s
	}
	return sec
}

func formatSeconds(s float64) string {
	total := int(s)
	h := total / 3600
	m := (total % 3600) / 60
	sec := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, sec)
	}
	return fmt.Sprintf("%02d:%02d", m, sec)
}

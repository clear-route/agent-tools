// Package calendar provides functions for interacting with Outlook Calendar
// via the Microsoft Graph API.
package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

// ---------- JSON output types ----------

// EventSummary is the JSON representation of a calendar event.
type EventSummary struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Subject  string `json:"subject"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Location string `json:"location"`
	IsAllDay bool   `json:"isAllDay"`
	Organizer string `json:"organizer"`
}

// EventCreated is the JSON response after creating an event.
type EventCreated struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	WebLink string `json:"webLink"`
}

// ---------- List ----------

// List prints calendar events within a time range.
// since and before are optional ISO date strings (YYYY-MM-DD or YYYY-MM-DD HH:MM).
// Default range: 30 days ago → 30 days from now.
func List(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, count int32, since, before string, jsonOutput bool) error {
	var startTime, endTime time.Time

	if since != "" {
		t, err := parseDateTime(since)
		if err != nil {
			return fmt.Errorf("invalid --since: %w", err)
		}
		startTime = t.UTC()
	} else {
		startTime = time.Now().UTC().AddDate(0, 0, -30)
	}

	if before != "" {
		t, err := parseDateTime(before)
		if err != nil {
			return fmt.Errorf("invalid --before: %w", err)
		}
		endTime = t.UTC()
	} else {
		endTime = time.Now().UTC().AddDate(0, 0, 30)
	}

	startStr := startTime.Format(time.RFC3339)
	endStr := endTime.Format(time.RFC3339)

	requestParams := &users.ItemCalendarViewRequestBuilderGetQueryParameters{
		StartDateTime: &startStr,
		EndDateTime:   &endStr,
		Select:        []string{"id", "subject", "start", "end", "location", "organizer", "isAllDay"},
		Top:           &count,
		Orderby:       []string{"start/dateTime ASC"},
	}
	config := &users.ItemCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParams,
	}

	result, err := client.Me().CalendarView().Get(ctx, config)
	if err != nil {
		return fmt.Errorf("listing calendar events: %w", err)
	}

	events := result.GetValue()

	if jsonOutput {
		summaries := make([]EventSummary, 0, len(events))
		for i, event := range events {
			location := ""
			if event.GetLocation() != nil {
				location = deref(event.GetLocation().GetDisplayName(), "")
			}
			organizer := ""
			if event.GetOrganizer() != nil && event.GetOrganizer().GetEmailAddress() != nil {
				organizer = deref(event.GetOrganizer().GetEmailAddress().GetAddress(), "")
			}
			isAllDay := event.GetIsAllDay() != nil && *event.GetIsAllDay()
			summaries = append(summaries, EventSummary{
				Index:     i + 1,
				ID:        deref(event.GetId(), ""),
				Subject:   deref(event.GetSubject(), ""),
				Start:     formatEventTime(event.GetStart()),
				End:       formatEventTime(event.GetEnd()),
				Location:  location,
				IsAllDay:  isAllDay,
				Organizer: organizer,
			})
		}
		return printJSON(summaries)
	}

	if len(events) == 0 {
		fmt.Println("No events found in the specified date range.")
		return nil
	}

	fmt.Printf("\n%-3s  %-40s  %-20s  %-20s  %s\n", "#", "Subject", "Start", "End", "Location")
	fmt.Println(strings.Repeat("-", 110))
	for i, event := range events {
		location := ""
		if event.GetLocation() != nil {
			location = deref(event.GetLocation().GetDisplayName(), "")
		}
		fmt.Printf("%-3d  %-40s  %-20s  %-20s  %s\n",
			i+1,
			truncate(deref(event.GetSubject(), "(no subject)"), 40),
			formatEventTime(event.GetStart()),
			formatEventTime(event.GetEnd()),
			truncate(location, 30),
		)
	}

	return nil
}

// ---------- Create ----------

// Create creates a new calendar event from explicit arguments — no interactive prompts.
// startStr and endStr accept: "2006-01-02 15:04" or "2006-01-02T15:04".
// attendees is a comma-separated list of email addresses (may be empty).
func Create(
	ctx context.Context,
	client *msgraphsdkgo.GraphServiceClient,
	title, startStr, endStr, location, attendees string,
	jsonOutput bool,
) error {
	if title == "" {
		return fmt.Errorf("--title is required")
	}
	if startStr == "" {
		return fmt.Errorf("--start is required (format: 2006-01-02 15:04)")
	}
	if endStr == "" {
		return fmt.Errorf("--end is required (format: 2006-01-02 15:04)")
	}

	startTime, err := parseDateTime(startStr)
	if err != nil {
		return fmt.Errorf("invalid --start: %w", err)
	}
	endTime, err := parseDateTime(endStr)
	if err != nil {
		return fmt.Errorf("invalid --end: %w", err)
	}

	event := models.NewEvent()
	event.SetSubject(&title)

	tz := "UTC"
	startDT := models.NewDateTimeTimeZone()
	startFormatted := startTime.Format("2006-01-02T15:04:05")
	startDT.SetDateTime(&startFormatted)
	startDT.SetTimeZone(&tz)
	event.SetStart(startDT)

	endDT := models.NewDateTimeTimeZone()
	endFormatted := endTime.Format("2006-01-02T15:04:05")
	endDT.SetDateTime(&endFormatted)
	endDT.SetTimeZone(&tz)
	event.SetEnd(endDT)

	if location != "" {
		loc := models.NewLocation()
		loc.SetDisplayName(&location)
		event.SetLocation(loc)
	}

	if attendees != "" {
		var attendeeList []models.Attendeeable
		for _, email := range strings.Split(attendees, ",") {
			email = strings.TrimSpace(email)
			if email == "" {
				continue
			}
			addr := models.NewEmailAddress()
			addr.SetAddress(&email)
			attendee := models.NewAttendee()
			attendee.SetEmailAddress(addr)
			attendeeType := models.REQUIRED_ATTENDEETYPE
			attendee.SetTypeEscaped(&attendeeType)
			attendeeList = append(attendeeList, attendee)
		}
		event.SetAttendees(attendeeList)
	}

	created, err := client.Me().Events().Post(ctx, event, nil)
	if err != nil {
		return fmt.Errorf("creating event: %w", err)
	}

	if jsonOutput {
		return printJSON(EventCreated{
			ID:      deref(created.GetId(), ""),
			Subject: deref(created.GetSubject(), title),
			WebLink: deref(created.GetWebLink(), ""),
		})
	}

	fmt.Fprintf(os.Stderr, "Event created: %s\n", deref(created.GetSubject(), title))
	if created.GetWebLink() != nil {
		fmt.Fprintf(os.Stderr, "Open in Outlook: %s\n", deref(created.GetWebLink(), ""))
	}
	return nil
}

// ---------- Helpers ----------

func formatEventTime(dt models.DateTimeTimeZoneable) string {
	if dt == nil {
		return ""
	}
	s := deref(dt.GetDateTime(), "")
	if s == "" {
		return ""
	}
	t, err := time.Parse("2006-01-02T15:04:05.9999999", s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", s)
		if err != nil {
			return s
		}
	}
	return t.Format("Jan 02 15:04")
}

func parseDateTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"02/01/2006 15:04",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse %q — use format: 2006-01-02 15:04", s)
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func deref(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

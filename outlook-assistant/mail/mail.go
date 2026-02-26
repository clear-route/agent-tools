// Package mail provides functions for interacting with Outlook mail
// via the Microsoft Graph API.
package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

// ---------- JSON output types ----------

// MessageSummary is the JSON representation of a message in a list or search result.
type MessageSummary struct {
	Index            int      `json:"index"`
	ID               string   `json:"id"`
	Subject          string   `json:"subject"`
	From             string   `json:"from"`
	ReceivedDateTime string   `json:"receivedDateTime"`
	IsRead           bool     `json:"isRead"`
	BodyPreview      string   `json:"bodyPreview"`
	Categories       []string `json:"categories,omitempty"`
}

// MessageDetail is the JSON representation of a fully-read message.
type MessageDetail struct {
	ID               string   `json:"id"`
	Subject          string   `json:"subject"`
	From             string   `json:"from"`
	To               []string `json:"to"`
	ReceivedDateTime string   `json:"receivedDateTime"`
	Body             string   `json:"body"`
	Categories       []string `json:"categories,omitempty"`
}

// FolderSummary is the JSON representation of a mail folder.
type FolderSummary struct {
	Index       int    `json:"index"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	TotalItems  int32  `json:"totalItems"`
	UnreadItems int32  `json:"unreadItems"`
}

// ---------- ID cache (stored in home directory) ----------

func idCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".outlook-assistant-mail-cache.json")
}

func saveIDCache(ids []string) {
	data, _ := json.Marshal(ids)
	_ = os.WriteFile(idCachePath(), data, 0600)
}

// appendIDCache merges new IDs onto the existing cache (used when paginating).
// IDs already present are skipped so duplicate pages don't corrupt the index.
func appendIDCache(newIDs []string) {
	existing := LoadIDCache()
	existingSet := make(map[string]bool, len(existing))
	for _, id := range existing {
		existingSet[id] = true
	}
	for _, id := range newIDs {
		if !existingSet[id] {
			existing = append(existing, id)
		}
	}
	saveIDCache(existing)
}

// LoadIDCache reads cached message IDs. Returns nil if no cache exists.
func LoadIDCache() []string {
	data, err := os.ReadFile(idCachePath())
	if err != nil {
		return nil
	}
	var ids []string
	_ = json.Unmarshal(data, &ids)
	return ids
}

func resolveMessageID(ref string) (string, error) {
	if n, err := strconv.Atoi(ref); err == nil {
		ids := LoadIDCache()
		if ids == nil {
			return "", fmt.Errorf("no cached message list — run `mail list` first")
		}
		if n < 1 || n > len(ids) {
			return "", fmt.Errorf("index %d out of range (last list had %d messages)", n, len(ids))
		}
		return ids[n-1], nil
	}
	return ref, nil
}

// ---------- List ----------

// ListOptions holds optional filter parameters for List.
type ListOptions struct {
	Since      string // RFC3339 or "2006-01-02" lower bound on receivedDateTime
	Before     string // RFC3339 or "2006-01-02" upper bound on receivedDateTime
	From       string // filter by sender email address
	UnreadOnly bool   // only return unread messages
	Folder     string // folder name or well-known name (default: inbox)
	Subject    string // client-side subject substring filter (case-insensitive)
}

// List prints inbox emails for the given page with optional filters.
// Page is 1-based; page 1 resets the ID cache, subsequent pages append to it
// so that index references remain valid across multi-page fetches.
func List(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, count int32, page int, opts ListOptions, jsonOutput bool) error {
	// Build $filter expression from options.
	var filters []string
	if opts.Since != "" {
		t, err := parseFlexibleDate(opts.Since)
		if err != nil {
			return fmt.Errorf("--since: %w", err)
		}
		filters = append(filters, "receivedDateTime ge "+t.UTC().Format(time.RFC3339))
	}
	if opts.Before != "" {
		t, err := parseFlexibleDate(opts.Before)
		if err != nil {
			return fmt.Errorf("--before: %w", err)
		}
		filters = append(filters, "receivedDateTime le "+t.UTC().Format(time.RFC3339))
	}
	if opts.From != "" {
		filters = append(filters, fmt.Sprintf("from/emailAddress/address eq '%s'", opts.From))
	}
	if opts.UnreadOnly {
		filters = append(filters, "isRead eq false")
	}

	var filterPtr *string
	if len(filters) > 0 {
		s := strings.Join(filters, " and ")
		filterPtr = &s
	}

	skip := int32((page - 1) * int(count))

	// sentitems uses sentDateTime; all other folders use receivedDateTime.
	// Using the wrong field causes Graph to return 0 results.
	orderField := "receivedDateTime"
	if strings.ToLower(opts.Folder) == "sentitems" {
		orderField = "sentDateTime"
	}

	requestParams := &users.ItemMailFoldersItemMessagesRequestBuilderGetQueryParameters{
		Select:  []string{"id", "subject", "from", "receivedDateTime", "isRead", "bodyPreview", "categories"},
		Top:     &count,
		Skip:    &skip,
		Orderby: []string{orderField + " DESC"},
		Filter:  filterPtr,
	}
	config := &users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParams,
	}

	folderID := "inbox"
	if opts.Folder != "" {
		var ferr error
		folderID, ferr = resolveFolderID(ctx, client, opts.Folder)
		if ferr != nil {
			return ferr
		}
	}

	result, err := client.Me().MailFolders().ByMailFolderId(folderID).Messages().Get(ctx, config)
	if err != nil {
		return fmt.Errorf("listing messages: %w", err)
	}

	messages := result.GetValue()

	// Client-side subject filter (Graph does not support subject $filter reliably).
	if opts.Subject != "" {
		lower := strings.ToLower(opts.Subject)
		filtered := make([]models.Messageable, 0, len(messages))
		for _, msg := range messages {
			if strings.Contains(strings.ToLower(deref(msg.GetSubject(), "")), lower) {
				filtered = append(filtered, msg)
			}
		}
		messages = filtered
	}

	// Update ID cache: page 1 resets it; subsequent pages accumulate so that
	// index references stay valid across multi-page fetches of the same query.
	ids := make([]string, 0, len(messages))
	for _, msg := range messages {
		ids = append(ids, deref(msg.GetId(), ""))
	}
	if page == 1 {
		saveIDCache(ids)
	} else {
		appendIDCache(ids)
	}

	// Indicate whether more pages exist.
	hasMore := result.GetOdataNextLink() != nil

	if jsonOutput {
		summaries := make([]MessageSummary, 0, len(messages))
		for i, msg := range messages {
			summaries = append(summaries, MessageSummary{
				Index:            i + 1,
				ID:               deref(msg.GetId(), ""),
				Subject:          deref(msg.GetSubject(), ""),
				From:             senderAddress(msg),
				ReceivedDateTime: formatMsgTime(msg.GetReceivedDateTime()),
				IsRead:           msg.GetIsRead() != nil && *msg.GetIsRead(),
				BodyPreview:      deref(msg.GetBodyPreview(), ""),
				Categories:       msg.GetCategories(),
			})
		}
		type listResult struct {
			Page     int              `json:"page"`
			Count    int              `json:"count"`
			HasMore  bool             `json:"hasMore"`
			Messages []MessageSummary `json:"messages"`
		}
		return printJSON(listResult{Page: page, Count: len(summaries), HasMore: hasMore, Messages: summaries})
	}

	if len(messages) == 0 {
		fmt.Println("No messages found.")
		return nil
	}

	fmt.Printf("\nPage %d  (showing %d messages)\n", page, len(messages))
	fmt.Printf("%-3s  %-50s  %-30s  %s\n", "#", "Subject", "From", "Received")
	fmt.Println(strings.Repeat("-", 110))
	for i, msg := range messages {
		read := " "
		if msg.GetIsRead() != nil && !*msg.GetIsRead() {
			read = "*"
		}
		cats := ""
		if len(msg.GetCategories()) > 0 {
			cats = " [" + strings.Join(msg.GetCategories(), ", ") + "]"
		}
		fmt.Printf("%s%-3d  %-50s  %-30s  %s%s\n",
			read, i+1,
			truncate(deref(msg.GetSubject(), "(no subject)"), 50),
			truncate(senderAddress(msg), 30),
			formatMsgTime(msg.GetReceivedDateTime()),
			cats,
		)
	}
	fmt.Println("\n(* = unread)")
	if hasMore {
		fmt.Fprintf(os.Stderr, "More messages available — use --page=%d to continue.\n", page+1)
	}
	return nil
}

// ---------- Read ----------

// Read fetches and prints a single message.
// ref may be a 1-based list index or a raw Graph message ID.
func Read(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, ref string, jsonOutput bool) error {
	messageID, err := resolveMessageID(ref)
	if err != nil {
		return err
	}

	config := &users.ItemMessagesMessageItemRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMessagesMessageItemRequestBuilderGetQueryParameters{
			Select: []string{"id", "subject", "from", "toRecipients", "receivedDateTime", "body", "isRead", "categories"},
		},
	}

	msg, err := client.Me().Messages().ByMessageId(messageID).Get(ctx, config)
	if err != nil {
		return fmt.Errorf("reading message: %w", err)
	}

	body := extractBody(msg)

	if jsonOutput {
		to := []string{}
		for _, r := range msg.GetToRecipients() {
			if r.GetEmailAddress() != nil {
				to = append(to, deref(r.GetEmailAddress().GetAddress(), ""))
			}
		}
		return printJSON(MessageDetail{
			ID:               deref(msg.GetId(), ""),
			Subject:          deref(msg.GetSubject(), ""),
			From:             senderAddress(msg),
			To:               to,
			ReceivedDateTime: formatMsgTime(msg.GetReceivedDateTime()),
			Body:             body,
			Categories:       msg.GetCategories(),
		})
	}

	fmt.Printf("\nSubject : %s\n", deref(msg.GetSubject(), "(no subject)"))
	if msg.GetFrom() != nil && msg.GetFrom().GetEmailAddress() != nil {
		fmt.Printf("From    : %s <%s>\n",
			deref(msg.GetFrom().GetEmailAddress().GetName(), ""),
			deref(msg.GetFrom().GetEmailAddress().GetAddress(), ""),
		)
	}
	if msg.GetReceivedDateTime() != nil {
		fmt.Printf("Date    : %s\n", msg.GetReceivedDateTime().Format("Mon, 02 Jan 2006 15:04:05"))
	}
	to := []string{}
	for _, r := range msg.GetToRecipients() {
		if r.GetEmailAddress() != nil {
			to = append(to, deref(r.GetEmailAddress().GetAddress(), ""))
		}
	}
	fmt.Printf("To      : %s\n", strings.Join(to, ", "))
	if len(msg.GetCategories()) > 0 {
		fmt.Printf("Categories: %s\n", strings.Join(msg.GetCategories(), ", "))
	}
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(body)
	return nil
}

// ---------- Send ----------

// Send composes and sends an email from flag arguments — no interactive prompts.
// to, cc, and bcc accept comma-separated email addresses; cc and bcc may be empty.
func Send(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, to, cc, bcc, subject, body string, format BodyFormat) error {
	if to == "" {
		return fmt.Errorf("--to is required")
	}
	if subject == "" {
		return fmt.Errorf("--subject is required")
	}

	message := models.NewMessage()
	message.SetSubject(&subject)

	htmlBody := RenderBody(body, format)
	bodyContent := models.NewItemBody()
	contentType := models.HTML_BODYTYPE
	bodyContent.SetContentType(&contentType)
	bodyContent.SetContent(&htmlBody)
	message.SetBody(bodyContent)

	message.SetToRecipients(parseRecipients(to))
	if cc != "" {
		message.SetCcRecipients(parseRecipients(cc))
	}
	if bcc != "" {
		message.SetBccRecipients(parseRecipients(bcc))
	}

	sendMailBody := users.NewItemSendMailPostRequestBody()
	saveToSentItems := true
	sendMailBody.SetSaveToSentItems(&saveToSentItems)
	sendMailBody.SetMessage(message)

	if err := client.Me().SendMail().Post(ctx, sendMailBody, nil); err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Email sent to %s\n", to)
	return nil
}

// parseRecipients splits a comma-separated list of email addresses into Recipientable values.
func parseRecipients(addresses string) []models.Recipientable {
	var recipients []models.Recipientable
	for _, addr := range strings.Split(addresses, ",") {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		ea := models.NewEmailAddress()
		ea.SetAddress(&addr)
		r := models.NewRecipient()
		r.SetEmailAddress(ea)
		recipients = append(recipients, r)
	}
	return recipients
}

// ---------- Reply ----------

// Reply sends a reply to a message identified by ref (list index or Graph ID).
// Uses createReply → patch body → send so that HTML formatting is preserved.
func Reply(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, ref, body string, format BodyFormat) error {
	if body == "" {
		return fmt.Errorf("--body is required")
	}

	messageID, err := resolveMessageID(ref)
	if err != nil {
		return err
	}

	// Step 1: create a draft reply.
	createReplyReqBody := users.NewItemMessagesItemCreateReplyPostRequestBody()
	draft, err := client.Me().Messages().ByMessageId(messageID).CreateReply().Post(ctx, createReplyReqBody, nil)
	if err != nil {
		return fmt.Errorf("creating reply draft: %w", err)
	}

	draftID := deref(draft.GetId(), "")

	// Step 2: patch the draft with our HTML body so formatting is preserved.
	htmlBody := RenderBody(body, format)
	patch := models.NewMessage()
	itemBody := models.NewItemBody()
	contentType := models.HTML_BODYTYPE
	itemBody.SetContentType(&contentType)
	itemBody.SetContent(&htmlBody)
	patch.SetBody(itemBody)

	if _, err := client.Me().Messages().ByMessageId(draftID).Patch(ctx, patch, nil); err != nil {
		return fmt.Errorf("updating reply draft body: %w", err)
	}

	// Step 3: send the draft.
	if err := client.Me().Messages().ByMessageId(draftID).Send().Post(ctx, nil); err != nil {
		return fmt.Errorf("sending reply draft: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Reply sent")
	return nil
}

// ---------- Forward ----------

// Forward creates a forwarded copy of a message and sends it to new recipients.
// Uses createForward → patch body → send so that HTML formatting is preserved.
// ref may be a 1-based list index or a raw Graph message ID.
// body is optional prepend text; if empty only the original message is forwarded.
func Forward(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, ref, to, cc, bcc, body string, format BodyFormat) error {
	if to == "" {
		return fmt.Errorf("--to is required for mail forward")
	}

	messageID, err := resolveMessageID(ref)
	if err != nil {
		return err
	}

	// Step 1: create a forward draft with the recipients already set.
	fwdBody := users.NewItemMessagesItemCreateForwardPostRequestBody()
	fwdBody.SetToRecipients(parseRecipients(to))

	draft, err := client.Me().Messages().ByMessageId(messageID).CreateForward().Post(ctx, fwdBody, nil)
	if err != nil {
		return fmt.Errorf("creating forward draft: %w", err)
	}

	draftID := deref(draft.GetId(), "")

	// Step 2: patch the draft — set CC/BCC and optionally prepend a custom body.
	patch := models.NewMessage()

	if cc != "" {
		patch.SetCcRecipients(parseRecipients(cc))
	}
	if bcc != "" {
		patch.SetBccRecipients(parseRecipients(bcc))
	}

	// Only patch the body if custom text was provided (otherwise the original
	// forwarded content created by Graph is preserved untouched).
	if body != "" {
		// Fetch the current draft body so we can prepend our text above it.
		draftMsg, err := client.Me().Messages().ByMessageId(draftID).Get(ctx,
			&users.ItemMessagesMessageItemRequestBuilderGetRequestConfiguration{
				QueryParameters: &users.ItemMessagesMessageItemRequestBuilderGetQueryParameters{
					Select: []string{"body"},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("reading forward draft body: %w", err)
		}

		originalHTML := ""
		if draftMsg.GetBody() != nil {
			originalHTML = deref(draftMsg.GetBody().GetContent(), "")
		}

		// Prepend our custom HTML above the quoted original.
		// RenderBodyInner gives inner HTML only (no html/body wrapper), so we
		// can safely splice it above the quoted message without creating nested
		// or malformed HTML documents. ExtractBodyContent strips the outer
		// html/body tags from Graph's original before combining.
		prepend := RenderBodyInner(body, format)
		quotedContent := ExtractBodyContent(originalHTML)
		combined := wrapEmailHTML(prepend + "\n<hr>\n" + quotedContent)

		itemBody := models.NewItemBody()
		contentType := models.HTML_BODYTYPE
		itemBody.SetContentType(&contentType)
		itemBody.SetContent(&combined)
		patch.SetBody(itemBody)
	}

	if _, err := client.Me().Messages().ByMessageId(draftID).Patch(ctx, patch, nil); err != nil {
		return fmt.Errorf("updating forward draft: %w", err)
	}

	// Step 3: send the draft.
	if err := client.Me().Messages().ByMessageId(draftID).Send().Post(ctx, nil); err != nil {
		return fmt.Errorf("sending forward draft: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Message forwarded to %s\n", to)
	return nil
}

// ---------- MarkRead ----------

// MarkRead sets or clears the isRead flag on a message.
// ref may be a 1-based list index or a raw Graph message ID.
func MarkRead(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, ref string, isRead bool) error {
	messageID, err := resolveMessageID(ref)
	if err != nil {
		return err
	}

	patch := models.NewMessage()
	patch.SetIsRead(&isRead)

	if _, err := client.Me().Messages().ByMessageId(messageID).Patch(ctx, patch, nil); err != nil {
		return fmt.Errorf("updating read state: %w", err)
	}

	if isRead {
		fmt.Fprintln(os.Stderr, "Message marked as read")
	} else {
		fmt.Fprintln(os.Stderr, "Message marked as unread")
	}
	return nil
}

// ---------- Delete ----------

// Delete permanently deletes a message (moves to Recoverable Items).
// ref may be a 1-based list index or a raw Graph message ID.
func Delete(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, ref string) error {
	messageID, err := resolveMessageID(ref)
	if err != nil {
		return err
	}

	if err := client.Me().Messages().ByMessageId(messageID).Delete(ctx, nil); err != nil {
		return fmt.Errorf("deleting message: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Message deleted")
	return nil
}

// ---------- Search ----------

// SearchOptions holds optional post-filter parameters for Search.
// Graph does not allow combining $search with $filter, so filtering is client-side.
type SearchOptions struct {
	Since  string // client-side lower bound on receivedDateTime (YYYY-MM-DD)
	Before string // client-side upper bound on receivedDateTime (YYYY-MM-DD)
}

// Search finds messages matching query.
// Note: Graph's $search does not support $skip — use -n to increase result size.
func Search(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, query string, count int32, opts SearchOptions, jsonOutput bool) error {
	if query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	quoted := `"` + query + `"`
	requestParams := &users.ItemMessagesRequestBuilderGetQueryParameters{
		Search: &quoted,
		Select: []string{"id", "subject", "from", "receivedDateTime", "isRead", "bodyPreview", "categories"},
		Top:    &count,
	}
	config := &users.ItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParams,
	}

	result, err := client.Me().Messages().Get(ctx, config)
	if err != nil {
		return fmt.Errorf("searching messages: %w", err)
	}

	messages := result.GetValue()

	// Client-side date filtering ($search + $filter cannot be combined in Graph).
	if opts.Since != "" || opts.Before != "" {
		var sinceT, beforeT time.Time
		if opts.Since != "" {
			if t, err := parseFlexibleDate(opts.Since); err == nil {
				sinceT = t
			}
		}
		if opts.Before != "" {
			if t, err := parseFlexibleDate(opts.Before); err == nil {
				beforeT = t
			}
		}
		filtered := make([]models.Messageable, 0, len(messages))
		for _, msg := range messages {
			if msg.GetReceivedDateTime() == nil {
				continue
			}
			msgTime := *msg.GetReceivedDateTime()
			if !sinceT.IsZero() && msgTime.Before(sinceT) {
				continue
			}
			if !beforeT.IsZero() && msgTime.After(beforeT) {
				continue
			}
			filtered = append(filtered, msg)
		}
		messages = filtered
	}

	// Cache IDs so results can be referenced by index.
	ids := make([]string, 0, len(messages))
	for _, msg := range messages {
		ids = append(ids, deref(msg.GetId(), ""))
	}
	saveIDCache(ids)

	if jsonOutput {
		summaries := make([]MessageSummary, 0, len(messages))
		for i, msg := range messages {
			summaries = append(summaries, MessageSummary{
				Index:            i + 1,
				ID:               deref(msg.GetId(), ""),
				Subject:          deref(msg.GetSubject(), ""),
				From:             senderAddress(msg),
				ReceivedDateTime: formatMsgTime(msg.GetReceivedDateTime()),
				IsRead:           msg.GetIsRead() != nil && *msg.GetIsRead(),
				BodyPreview:      deref(msg.GetBodyPreview(), ""),
				Categories:       msg.GetCategories(),
			})
		}
		return printJSON(summaries)
	}

	if len(messages) == 0 {
		fmt.Printf("No messages found for %q.\n", query)
		return nil
	}

	fmt.Printf("\nSearch results for %q:\n\n", query)
	fmt.Printf("%-3s  %-50s  %-30s  %s\n", "#", "Subject", "From", "Received")
	fmt.Println(strings.Repeat("-", 110))
	for i, msg := range messages {
		read := " "
		if msg.GetIsRead() != nil && !*msg.GetIsRead() {
			read = "*"
		}
		fmt.Printf("%s%-3d  %-50s  %-30s  %s\n",
			read, i+1,
			truncate(deref(msg.GetSubject(), "(no subject)"), 50),
			truncate(senderAddress(msg), 30),
			formatMsgTime(msg.GetReceivedDateTime()),
		)
	}
	fmt.Println("\n(* = unread)")
	return nil
}

// ---------- Archive ----------

// Archive moves a message to the Archive folder.
// ref may be a 1-based list index or a raw Graph message ID.
func Archive(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, ref string) error {
	return Move(ctx, client, ref, "archive")
}

// ---------- Move ----------

// Move moves a message to the named folder.
// folderName may be a well-known name (inbox, archive, deleteditems, drafts, sentitems, junkemail)
// or a display name that will be resolved against the user's folder list.
func Move(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, ref, folderName string) error {
	if folderName == "" {
		return fmt.Errorf("--folder is required")
	}

	messageID, err := resolveMessageID(ref)
	if err != nil {
		return err
	}

	// Resolve folder name to an ID. Well-known names work directly as IDs.
	folderID, err := resolveFolderID(ctx, client, folderName)
	if err != nil {
		return err
	}

	moveBody := users.NewItemMessagesItemMovePostRequestBody()
	moveBody.SetDestinationId(&folderID)

	if _, err := client.Me().Messages().ByMessageId(messageID).Move().Post(ctx, moveBody, nil); err != nil {
		return fmt.Errorf("moving message: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Message moved to %q\n", folderName)
	return nil
}

// resolveFolderID returns a folder ID for the given name.
// If the name is a well-known Outlook folder name it is used directly.
// Otherwise the user's folders are searched by display name (case-insensitive).
func resolveFolderID(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, name string) (string, error) {
	wellKnown := map[string]bool{
		"inbox": true, "archive": true, "deleteditems": true,
		"drafts": true, "sentitems": true, "junkemail": true,
		"outbox": true, "recoverableitemsdeletions": true,
	}
	lower := strings.ToLower(strings.ReplaceAll(name, " ", ""))
	if wellKnown[lower] {
		return lower, nil
	}

	// Search user folders by display name.
	top := int32(100)
	result, err := client.Me().MailFolders().Get(ctx, &users.ItemMailFoldersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMailFoldersRequestBuilderGetQueryParameters{
			Select: []string{"id", "displayName"},
			Top:    &top,
		},
	})
	if err != nil {
		return "", fmt.Errorf("listing folders: %w", err)
	}

	for _, f := range result.GetValue() {
		if strings.EqualFold(deref(f.GetDisplayName(), ""), name) {
			return deref(f.GetId(), ""), nil
		}
	}
	return "", fmt.Errorf("folder %q not found — use `mail folders` to list available folders", name)
}

// ---------- Categorize ----------

// Categorize sets (or clears) Outlook categories on a message.
// set is a comma-separated list of category names to apply; pass empty to clear all.
func Categorize(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, ref, set string) error {
	messageID, err := resolveMessageID(ref)
	if err != nil {
		return err
	}

	var cats []string
	if set != "" {
		for _, c := range strings.Split(set, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				cats = append(cats, c)
			}
		}
	}

	patch := models.NewMessage()
	patch.SetCategories(cats)

	if _, err := client.Me().Messages().ByMessageId(messageID).Patch(ctx, patch, nil); err != nil {
		return fmt.Errorf("categorizing message: %w", err)
	}

	if len(cats) == 0 {
		fmt.Fprintln(os.Stderr, "Categories cleared")
	} else {
		fmt.Fprintf(os.Stderr, "Categories set: %s\n", strings.Join(cats, ", "))
	}
	return nil
}

// ---------- Folders ----------

// Folders lists the user's mail folders.
func Folders(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, jsonOutput bool) error {
	top := int32(100)
	result, err := client.Me().MailFolders().Get(ctx, &users.ItemMailFoldersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMailFoldersRequestBuilderGetQueryParameters{
			Select: []string{"id", "displayName", "totalItemCount", "unreadItemCount"},
			Top:    &top,
		},
	})
	if err != nil {
		return fmt.Errorf("listing folders: %w", err)
	}

	folders := result.GetValue()

	if jsonOutput {
		summaries := make([]FolderSummary, 0, len(folders))
		for i, f := range folders {
			total := int32(0)
			if f.GetTotalItemCount() != nil {
				total = *f.GetTotalItemCount()
			}
			unread := int32(0)
			if f.GetUnreadItemCount() != nil {
				unread = *f.GetUnreadItemCount()
			}
			summaries = append(summaries, FolderSummary{
				Index:       i + 1,
				ID:          deref(f.GetId(), ""),
				Name:        deref(f.GetDisplayName(), ""),
				TotalItems:  total,
				UnreadItems: unread,
			})
		}
		return printJSON(summaries)
	}

	fmt.Printf("\n%-3s  %-35s  %8s  %8s\n", "#", "Folder", "Total", "Unread")
	fmt.Println(strings.Repeat("-", 60))
	for i, f := range folders {
		total := int32(0)
		if f.GetTotalItemCount() != nil {
			total = *f.GetTotalItemCount()
		}
		unread := int32(0)
		if f.GetUnreadItemCount() != nil {
			unread = *f.GetUnreadItemCount()
		}
		fmt.Printf("%-3d  %-35s  %8d  %8d\n", i+1, deref(f.GetDisplayName(), ""), total, unread)
	}
	return nil
}

// ---------- Helpers ----------

func senderAddress(msg models.Messageable) string {
	if msg.GetFrom() != nil && msg.GetFrom().GetEmailAddress() != nil {
		return deref(msg.GetFrom().GetEmailAddress().GetAddress(), "")
	}
	return ""
}

func extractBody(msg models.Messageable) string {
	if msg.GetBody() == nil {
		return ""
	}
	body := deref(msg.GetBody().GetContent(), "")
	if msg.GetBody().GetContentType() != nil {
		if strings.ToLower(msg.GetBody().GetContentType().String()) == "html" {
			return stripHTML(body)
		}
	}
	return body
}

func formatMsgTime(t interface{ Format(string) string }) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04")
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

// stripHTML removes HTML tags and decodes common entities for plain-text rendering.
func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, ch := range s {
		switch {
		case ch == '<':
			inTag = true
		case ch == '>':
			inTag = false
		case !inTag:
			result.WriteRune(ch)
		}
	}
	text := result.String()

	// Decode the most common HTML entities.
	replacer := strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&apos;", "'",
		"&mdash;", "—",
		"&ndash;", "–",
		"&hellip;", "…",
		"&laquo;", "«",
		"&raquo;", "»",
		"&#160;", " ",
		"&#8203;", "", // zero-width space
	)
	text = replacer.Replace(text)

	// Strip invisible Unicode characters that survive HTML entity decoding.
	text = stripInvisibleUnicode(text)

	// Collapse whitespace and trim blank lines.
	lines := strings.Split(text, "\n")
	var cleaned []string
	blanks := 0
	for _, l := range lines {
		l = strings.TrimRight(l, " \t\r")
		// Collapse runs of spaces/tabs down to a single space.
		l = collapseSpaces(l)
		if l == "" {
			blanks++
			if blanks <= 1 {
				cleaned = append(cleaned, l)
			}
		} else {
			blanks = 0
			cleaned = append(cleaned, l)
		}
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// stripInvisibleUnicode removes zero-width and formatting Unicode characters
// that survive HTML entity decoding and pollute plain-text output.
func stripInvisibleUnicode(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\u200b', // zero-width space
			'\u200c', // zero-width non-joiner
			'\u200d', // zero-width joiner
			'\u200e', // left-to-right mark
			'\u200f', // right-to-left mark
			'\u034f', // combining grapheme joiner
			'\ufeff', // BOM / zero-width no-break space
			'\u00ad': // soft hyphen
			// drop
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// collapseSpaces replaces runs of whitespace (space/tab) with a single space.
func collapseSpaces(s string) string {
	var b strings.Builder
	prevSpace := false
	for _, ch := range s {
		if ch == ' ' || ch == '\t' {
			if !prevSpace {
				b.WriteRune(' ')
			}
			prevSpace = true
		} else {
			prevSpace = false
			b.WriteRune(ch)
		}
	}
	return b.String()
}
// Body rendering is handled by RenderBody / RenderBodyInner in formatting.go.
// Accepted: "2006-01-02", "2006-01-02 15:04", "2006-01-02T15:04:05Z07:00".
func parseFlexibleDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date format %q — use YYYY-MM-DD or YYYY-MM-DD HH:MM", s)
}

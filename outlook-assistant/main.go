package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"

	"outlook-assistant/auth"
	"outlook-assistant/calendar"
	"outlook-assistant/mail"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load credentials — try multiple locations so the tool works from any CWD.
	// Priority: binary's own directory → ~/.outlook-assistant.env → CWD .env
	loadEnv()

	clientID := os.Getenv("CLIENT_ID")
	tenantID := os.Getenv("TENANT_ID")
	if clientID == "" || tenantID == "" {
		return fmt.Errorf("CLIENT_ID and TENANT_ID must be set in environment or .env file")
	}

	// ── Structural flags ──────────────────────────────────────────────────────
	group  := flag.String("group", "mail", "Command group: mail | calendar (default: mail)")
	action := flag.String("action", "", "Action: list | read | send | reply | forward | search | archive | move | categorize | markread | delete | folders | create")
	ref    := flag.String("ref", "", "Message reference: list index (e.g. 3) or raw Graph message ID")
	query  := flag.String("query", "", "Search query string (mail search)")

	// ── Shared output flag ────────────────────────────────────────────────────
	jsonOut := flag.Bool("json", false, "Output results as JSON to stdout")

	// ── List / filter flags ───────────────────────────────────────────────────
	count   := flag.Int("n", 20, "Number of messages or events to fetch")
	page    := flag.Int("page", 1, "Page number, 1-based (mail list)")
	since   := flag.String("since", "", "Only messages received on or after date: YYYY-MM-DD or YYYY-MM-DD HH:MM")
	before  := flag.String("before", "", "Only messages received on or before date: YYYY-MM-DD or YYYY-MM-DD HH:MM")
	from    := flag.String("from", "", "Only messages from this sender email address")
	unread  := flag.Bool("unread", false, "mail list: only unread messages. mail markread: mark as unread instead of read")
	folder  := flag.String("folder", "inbox", "Folder name or well-known name (mail list, mail move). Default: inbox")
	subject := flag.String("subject", "", "Email subject — filter substring for mail list, subject line for mail send")

	// ── Send / reply flags ────────────────────────────────────────────────────
	to   := flag.String("to", "", "Recipient address(es), comma-separated (mail send)")
	cc   := flag.String("cc", "", "CC address(es), comma-separated (mail send)")
	bcc  := flag.String("bcc", "", "BCC address(es), comma-separated (mail send)")
	body := flag.String("body", "", "Message body text (mail send, mail reply)")

	// ── Categorize flag ───────────────────────────────────────────────────────
	set := flag.String("set", "", "Comma-separated category names to apply; empty string clears all (mail categorize)")

	// ── Calendar create flags ─────────────────────────────────────────────────
	title     := flag.String("title", "", "Event title (calendar create)")
	start     := flag.String("start", "", "Start date/time: \"2006-01-02 15:04\" (calendar create)")
	end       := flag.String("end", "", "End date/time: \"2006-01-02 15:04\" (calendar create)")
	location  := flag.String("location", "", "Location string (calendar create)")
	attendees := flag.String("attendees", "", "Comma-separated attendee emails (calendar create)")

	flag.Usage = printUsage
	flag.Parse()

	if *action == "" {
		printUsage()
		return nil
	}

	fmt.Fprintln(os.Stderr, "Authenticating with Microsoft...")
	client, err := auth.NewGraphClient(clientID, tenantID)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	ctx := context.Background()

	switch *group {
	case "mail":
		return handleMail(ctx, client, *action, *ref, *query, *jsonOut, *count, *page,
			*since, *before, *from, *unread, *folder, *subject,
			*to, *cc, *bcc, *body, *set)

	case "calendar":
		return handleCalendar(ctx, client, *action, *jsonOut, *count,
			*since, *before,
			*title, *start, *end, *location, *attendees)

	default:
		return fmt.Errorf("unknown group %q — valid groups: mail, calendar", *group)
	}
}

// ── env loading ──────────────────────────────────────────────────────────────

// loadEnv tries to load credentials from several locations so the binary works
// regardless of where it is invoked from (Forge agent, terminal, CI, etc.).
// godotenv never overwrites vars already set in the environment, so if
// CLIENT_ID / TENANT_ID are exported in .zshrc they take precedence.
func loadEnv() {
	// 1. .env sitting next to the binary (e.g. ~/.forge/tools/outlook-assistant/.env)
	if exe, err := os.Executable(); err == nil {
		_ = godotenv.Load(filepath.Join(filepath.Dir(exe), ".env"))
	}
	// 2. ~/.outlook-assistant.env — dedicated home-directory credentials file
	if home, err := os.UserHomeDir(); err == nil {
		_ = godotenv.Load(filepath.Join(home, ".outlook-assistant.env"))
	}
	// 3. .env in current working directory (repo dev workflow)
	_ = godotenv.Load()
}

// ── mail ──────────────────────────────────────────────────────────────────────

func handleMail(
	ctx context.Context,
	client *msgraphsdkgo.GraphServiceClient,
	action, ref, query string,
	jsonOut bool,
	count, page int,
	since, before, from string,
	unread bool,
	folder, subject string,
	to, cc, bcc, body, set string,
) error {
	switch action {
	case "list":
		opts := mail.ListOptions{
			Since:      since,
			Before:     before,
			From:       from,
			UnreadOnly: unread,
			Folder:     folder,
			Subject:    subject,
		}
		return mail.List(ctx, client, int32(count), page, opts, jsonOut)

	case "read":
		if ref == "" {
			return fmt.Errorf("--ref is required for mail read")
		}
		return mail.Read(ctx, client, ref, jsonOut)

	case "send":
		if to == "" || subject == "" {
			return fmt.Errorf("--to and --subject are required for mail send")
		}
		return mail.Send(ctx, client, to, cc, bcc, subject, body)

	case "reply":
		if ref == "" {
			return fmt.Errorf("--ref is required for mail reply")
		}
		if body == "" {
			return fmt.Errorf("--body is required for mail reply")
		}
		return mail.Reply(ctx, client, ref, body)

	case "forward":
		if ref == "" {
			return fmt.Errorf("--ref is required for mail forward")
		}
		if to == "" {
			return fmt.Errorf("--to is required for mail forward")
		}
		return mail.Forward(ctx, client, ref, to, cc, bcc, body)

	case "search":
		if query == "" {
			return fmt.Errorf("--query is required for mail search")
		}
		opts := mail.SearchOptions{Since: since, Before: before}
		return mail.Search(ctx, client, query, int32(count), opts, jsonOut)

	case "archive":
		if ref == "" {
			return fmt.Errorf("--ref is required for mail archive")
		}
		return mail.Archive(ctx, client, ref)

	case "move":
		if ref == "" || folder == "" {
			return fmt.Errorf("--ref and --folder are required for mail move")
		}
		return mail.Move(ctx, client, ref, folder)

	case "categorize":
		if ref == "" {
			return fmt.Errorf("--ref is required for mail categorize")
		}
		return mail.Categorize(ctx, client, ref, set)

	case "markread":
		if ref == "" {
			return fmt.Errorf("--ref is required for mail markread")
		}
		return mail.MarkRead(ctx, client, ref, !unread)

	case "delete":
		if ref == "" {
			return fmt.Errorf("--ref is required for mail delete")
		}
		return mail.Delete(ctx, client, ref)

	case "folders":
		return mail.Folders(ctx, client, jsonOut)

	default:
		return fmt.Errorf("unknown mail action %q", action)
	}
}

// ── calendar ──────────────────────────────────────────────────────────────────

func handleCalendar(
	ctx context.Context,
	client *msgraphsdkgo.GraphServiceClient,
	action string,
	jsonOut bool,
	count int,
	since, before string,
	title, start, end, location, attendees string,
) error {
	switch action {
	case "list":
		return calendar.List(ctx, client, int32(count), since, before, jsonOut)

	case "create":
		if title == "" || start == "" || end == "" {
			return fmt.Errorf("--title, --start, and --end are required for calendar create")
		}
		return calendar.Create(ctx, client, title, start, end, location, attendees, jsonOut)

	default:
		return fmt.Errorf("unknown calendar action %q", action)
	}
}

// ── usage ─────────────────────────────────────────────────────────────────────

func printUsage() {
	fmt.Fprint(os.Stderr, `
Outlook Assistant — Microsoft Graph mail & calendar CLI.

All flags are named; no positional arguments. Designed for agent and pipeline use.

REQUIRED FLAGS (always)
  --group=<mail|calendar>    Command group
  --action=<action>          Action to perform (see below)

MAIL ACTIONS
  list        List messages
              --folder=inbox --n=20 --page=1 --since=YYYY-MM-DD --before=YYYY-MM-DD
              --from=email --subject=text --unread --json

  read        Read a message body
              --ref=<index|id> --json

  send        Send a new message
              --to=<email,...> --subject=<text> --body=<text>
              --cc=<email,...> --bcc=<email,...>

  reply       Reply to a message
              --ref=<index|id> --body=<text>

  forward     Forward a message to new recipients
              --ref=<index|id> --to=<email,...> [--cc=<email,...>] [--bcc=<email,...>] [--body=<text>]

  search      Search messages
              --query=<text> --n=20 --since=YYYY-MM-DD --before=YYYY-MM-DD --json

  archive     Archive a message         --ref=<index|id>
  move        Move to folder            --ref=<index|id> --folder=<name>
  categorize  Set categories            --ref=<index|id> --set=<cat1,cat2,...>
  markread    Mark read/unread          --ref=<index|id> [--unread]
  delete      Delete a message          --ref=<index|id>
  folders     List all mail folders     --json

CALENDAR ACTIONS
  list        List events in a date range
              --n=20 --since=YYYY-MM-DD --before=YYYY-MM-DD --json
              (default: 30 days ago → 30 days ahead)
  create      Create an event
              --title=<text> --start="2006-01-02 15:04" --end="2006-01-02 15:04"
              --location=<text> --attendees=<email,...> --json

NOTES
  --json outputs structured JSON to stdout; all status messages go to stderr.
  --ref accepts the index number from the last mail list/search, or a raw Graph ID.
  Well-known folder names: inbox, archive, deleteditems, drafts, sentitems, junkemail.
  Credentials: CLIENT_ID and TENANT_ID must be set in environment or .env file.
`)
}

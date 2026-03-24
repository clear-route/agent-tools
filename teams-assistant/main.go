package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"teams-assistant/auth"
	"teams-assistant/graph"
	"teams-assistant/meetings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	loadEnv()

	clientID := os.Getenv("CLIENT_ID")
	tenantID := os.Getenv("TENANT_ID")
	if clientID == "" || tenantID == "" {
		return fmt.Errorf("CLIENT_ID and TENANT_ID must be set in environment or .env file")
	}

	action := flag.String("action", "", "Action: list | search | insights | transcripts | transcript")
	ref := flag.String("ref", "", "Meeting reference: list index (e.g. 3) or raw online meeting ID")
	query := flag.String("query", "", "Search query string (meeting search)")
	jsonOut := flag.Bool("json", false, "Output results as JSON to stdout")
	count := flag.Int("n", 20, "Number of meetings to fetch")
	since := flag.String("since", "", "Only meetings on or after date: YYYY-MM-DD")
	before := flag.String("before", "", "Only meetings on or before date: YYYY-MM-DD")

	flag.Usage = printUsage
	flag.Parse()

	if *action == "" {
		printUsage()
		return nil
	}

	fmt.Fprintln(os.Stderr, "Authenticating with Microsoft...")
	cred, err := auth.NewCredential(clientID, tenantID)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	client := graph.NewClient(cred, auth.Scopes())
	ctx := context.Background()

	switch *action {
	case "list":
		return meetings.List(ctx, client, int32(*count), *since, *before, *jsonOut)

	case "search":
		if *query == "" {
			return fmt.Errorf("--query is required for search")
		}
		return meetings.Search(ctx, client, *query, int32(*count), *since, *before, *jsonOut)

	case "transcript":
		if *ref == "" {
			return fmt.Errorf("--ref is required for transcript")
		}
		return meetings.Transcript(ctx, client, *ref, *jsonOut)

	default:
		return fmt.Errorf("unknown action %q — valid actions: list, search, transcript", *action)
	}
}

func loadEnv() {
	if exe, err := os.Executable(); err == nil {
		_ = godotenv.Load(filepath.Join(filepath.Dir(exe), ".env"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		_ = godotenv.Load(filepath.Join(home, ".teams-assistant.env"))
	}
	_ = godotenv.Load()
}

func printUsage() {
	fmt.Fprint(os.Stderr, `
Teams Assistant — Microsoft Graph meeting transcripts CLI.

All flags are named; no positional arguments. Designed for agent and pipeline use.

ACTIONS
  list          List recent Teams meetings (includes hasTranscript flag)
                --n=20 --since=YYYY-MM-DD --before=YYYY-MM-DD --json

  search        Search meetings by subject (includes hasTranscript flag)
                --query=<text> --n=20 --since=YYYY-MM-DD --before=YYYY-MM-DD --json

  transcript    Get meeting transcript as chunked plain-text files
                --ref=<index|id> --json

NOTES
  --json outputs structured JSON to stdout; all status messages go to stderr.
  --ref accepts the index number from the last list/search, or a raw online meeting ID.
  Default date range for list: 30 days ago to 30 days ahead.
  Default date range for search: 3 months ago to 30 days ahead.
  Credentials: CLIENT_ID and TENANT_ID must be set in environment or .env file.

  The transcript action writes plain-text chunks to temp files and returns
  metadata (duration, speakers, chunk paths). Read chunks sequentially.
`)
}

# outlook-assistant

A Forge custom tool for interacting with Microsoft 365 mail and calendar via the Graph API.

All output is structured JSON (with `--json`) or clean plain text — designed for agent pipelines and terminal use.

---

## Installation

```bash
# From the repo root
cd outlook-assistant
go build -o ~/.forge/tools/outlook-assistant/outlook-assistant .
cp tool.yaml ~/.forge/tools/outlook-assistant/tool.yaml
```

Forge auto-discovers tools in `~/.forge/tools/` — no restart needed.

**First time setup?** See [setup.md](setup.md) to register the Azure app and configure credentials.

---

## Credentials

A `.env.example` is included in the repo with the correct `CLIENT_ID` and `TENANT_ID` for ClearRoute. Copy it into place after building:

```bash
cp .env.example ~/.forge/tools/outlook-assistant/.env
```

No client secret is needed — authentication uses Interactive Browser Flow (see below).

---

## Authentication

On first run, your default browser opens automatically to the Microsoft 365 sign-in page. Sign in with your ClearRoute account and grant consent.

An auth record is cached at `~/.outlook-assistant-auth.json`. Subsequent runs are silent — no browser interaction until the token expires.

---

## Commands

All flags are named. No positional arguments.

### Mail

| Action | Required flags | Optional flags |
|--------|---------------|----------------|
| `list` | — | `--folder` `--n` `--page` `--since` `--before` `--from` `--subject` `--unread` `--json` |
| `read` | `--ref` | `--json` |
| `send` | `--to` `--subject` | `--body` `--cc` `--bcc` |
| `reply` | `--ref` `--body` | — |
| `forward` | `--ref` `--to` | `--body` `--cc` `--bcc` |
| `search` | `--query` | `--n` `--since` `--before` `--json` |
| `archive` | `--ref` | — |
| `move` | `--ref` `--folder` | — |
| `categorize` | `--ref` `--set` | — |
| `markread` | `--ref` | `--unread` (to mark unread instead) |
| `delete` | `--ref` | — |
| `folders` | — | `--json` |

### Calendar

| Action | Required flags | Optional flags |
|--------|---------------|----------------|
| `list` | — | `--n` `--since` `--before` `--json` |
| `create` | `--title` `--start` `--end` | `--location` `--attendees` `--json` |

### Flag reference

| Flag | Description |
|------|-------------|
| `--group` | `mail` or `calendar` (default: `mail`) |
| `--action` | Action name from the tables above |
| `--ref` | Message index from last `list`/`search`, or raw Graph message ID |
| `--n` | Number of results (default: 20) |
| `--page` | Page number, 1-based (default: 1) |
| `--folder` | Mail folder name. Well-known: `inbox` `archive` `sentitems` `drafts` `deleteditems` `junkemail` |
| `--since` / `--before` | Date filter: `YYYY-MM-DD` or `YYYY-MM-DD HH:MM` |
| `--from` | Filter by sender email |
| `--subject` | Filter by subject substring (list) or set subject (send) |
| `--unread` | Filter unread only (list) or mark as unread (markread) |
| `--query` | Search query string |
| `--to` / `--cc` / `--bcc` | Recipient addresses, comma-separated |
| `--body` | Message body text |
| `--set` | Comma-separated category names (empty string clears all) |
| `--title` | Event title |
| `--start` / `--end` | Event date/time: `"2006-01-02 15:04"` |
| `--location` | Event location |
| `--attendees` | Comma-separated attendee emails |
| `--json` | Output structured JSON to stdout; status messages go to stderr |

### Examples

```bash
# List 10 unread emails
outlook-assistant --action=list --unread --n=10 --json

# Read the 3rd email from the last list
outlook-assistant --action=read --ref=3 --json

# Send an email
outlook-assistant --action=send --to=someone@clearroute.io --subject="Hello" --body="Hi there"

# Search for emails about invoices
outlook-assistant --action=search --query="invoice" --json

# List calendar events for the next two weeks
outlook-assistant --action=list --group=calendar --since=2025-01-01 --before=2025-01-15 --json

# Create a calendar event
outlook-assistant --action=create --group=calendar --title="Standup" --start="2025-01-10 09:00" --end="2025-01-10 09:30" --attendees="alice@clearroute.io,bob@clearroute.io"
```

---

## Security

- `.env` and `~/.outlook-assistant-auth.json` must **never** be committed — both are covered by `.gitignore`.
- The tool requests only the minimum Graph permissions: `Mail.ReadWrite`, `Mail.Send`, `Calendars.ReadWrite`, `User.Read`.
- No client secret is stored — authentication delegates entirely to the browser sign-in flow.

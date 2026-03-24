# teams-assistant

A Forge custom tool for fetching Microsoft Teams meeting transcripts and metadata via the Graph API. Use transcripts to generate meeting summaries, action items, and key decisions.

All output is structured JSON (with `--json`) or clean plain text — designed for agent pipelines and terminal use.

---

## Installation

```bash
# From the repo root
cd teams-assistant
go build -o ~/.forge/tools/teams-assistant/teams-assistant .
cp tool.yaml ~/.forge/tools/teams-assistant/tool.yaml
```

Forge auto-discovers tools in `~/.forge/tools/` — no restart needed.

**First time setup?** See [setup.md](setup.md) to add API permissions to the existing Azure app and configure credentials.

---

## Credentials

The same `CLIENT_ID` and `TENANT_ID` from outlook-assistant are used. A `.env.example` is included in the repo with the correct values for ClearRoute. Copy it into place after building:

```bash
cp .env.example ~/.forge/tools/teams-assistant/.env
```

No client secret is needed — authentication uses Interactive Browser Flow (see below).

---

## Authentication

On first run, your default browser opens automatically to the Microsoft 365 sign-in page. Sign in with your ClearRoute account and grant consent for the new meeting permissions.

An auth record is cached at `~/.teams-assistant-auth.json`. Subsequent runs are silent — no browser interaction until the token expires.

---

## Commands

All flags are named. No positional arguments.

### Meetings

| Action | Required flags | Optional flags |
|--------|---------------|----------------|
| `list` | — | `--n` `--since` `--before` `--json` |
| `search` | `--query` | `--n` `--since` `--before` `--json` |
| `transcript` | `--ref` | `--json` |

### Flag reference

| Flag | Description |
|------|-------------|
| `--action` | Action name from the table above |
| `--ref` | Meeting index from last `list`/`search`, or raw online meeting ID |
| `--n` | Number of results (default: 20) |
| `--since` / `--before` | Date filter: `YYYY-MM-DD` |
| `--query` | Search query string (matches meeting subject) |
| `--json` | Output structured JSON to stdout; status messages go to stderr |

### Examples

```bash
# List recent Teams meetings (includes hasTranscript flag)
teams-assistant --action=list --n=10 --json

# List meetings from a specific date range
teams-assistant --action=list --since=2026-03-01 --before=2026-03-15 --json

# Search for meetings by subject
teams-assistant --action=search --query="sprint planning" --json

# Get transcript for a meeting (chunked plain-text files + metadata)
teams-assistant --action=transcript --ref=3 --json
```

---

## Security

- `.env` and `~/.teams-assistant-auth.json` must **never** be committed — both are covered by `.gitignore`.
- The tool requests only the minimum Graph permissions: `User.Read`, `Calendars.Read`, `OnlineMeetings.Read`, `OnlineMeetingTranscript.Read.All`.
- No client secret is stored — authentication delegates entirely to the browser sign-in flow.

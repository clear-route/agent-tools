# Outlook Assistant — Feature Backlog

## 🔴 High Priority (biggest gaps for agent use)

| # | Feature | Command | Notes |
|---|---------|---------|-------|
| 1 | Folder-aware mail listing | `mail list --folder=<name>` | List any folder (Sent, Archive, Drafts, custom), not just inbox |
| 2 | Calendar date filtering | `calendar list --since=YYYY-MM-DD --before=YYYY-MM-DD` | Date range filtering; currently fetches next N from now with no window control |
| 3 | Unicode/zero-width char stripping | (body rendering) | Strip invisible Unicode (e.g. `\u200c`, `\u034f`) from body previews — HTML stripper only handles entities |
| 4 | Multi-recipient send | `mail send --cc=<email> --bcc=<email>` | Currently only supports a single `--to` recipient |

## 🟠 Medium Priority (completeness)

| # | Feature | Command | Notes |
|---|---------|---------|-------|
| 5 | Search date filtering | `mail search --since / --before` | Graph won't allow `$search` + `$filter` together — post-filter client-side |
| 6 | Mark read/unread | `mail mark-read <index\|id>` / `mail mark-unread <index\|id>` | Toggle isRead flag via PATCH |
| 7 | Delete mail | `mail delete <index\|id>` | Move to deletedItems folder |

## 🟡 Lower Priority (polish)

| # | Feature | Command | Notes |
|---|---------|---------|-------|
| 8 | Multiple --to recipients | `mail send --to=a@b.com,c@d.com` | Comma-separated, consistent with `--attendees` on calendar |
| 9 | Subject filter | `mail list --subject=<text>` | Client-side substring match (Graph doesn't support subject $filter directly) |
| 10 | Calendar create JSON output | `calendar create ... --json` | Return created event ID and webLink in JSON for agent chaining |

## ✅ Done

- `mail list` with `--since`, `--before`, `--from`, `--unread`, `--page`, `-n`, `--json`
- `mail read <index|id>` with `--json`
- `mail send --to --subject --body`
- `mail reply <index|id> --body`
- `mail search <query> -n --json`
- `mail archive <index|id>`
- `mail move <index|id> --folder`
- `mail categorize <index|id> --set`
- `mail folders --json`
- `calendar list -n --json`
- `calendar create` with flags (`--title`, `--start`, `--end`, `--location`, `--attendees`)
- Persistent token cache (OS keychain, `~/.outlook-assistant-auth.json`)
- ID cache accumulates across pages (`~/.outlook-assistant-mail-cache.json`)
- HTML entity decoding in body rendering
- All status to stderr, data to stdout

# Outlook Assistant

A terminal-based tool for interacting with your Microsoft 365 / Outlook inbox and calendar using the [Microsoft Graph API](https://learn.microsoft.com/en-us/graph/overview).

Built in Go. No browser. No web server. Just a binary you run from your terminal.

---

## Features

- 📬 **List emails** — view your inbox at a glance with sender, subject, and timestamp
- 📖 **Read emails** — open any message by its list number; HTML is stripped to readable plain text
- ✉️  **Send emails** — compose and send interactively from the terminal
- 📅 **List calendar events** — see upcoming events for the next 30 days
- 🗓️  **Create calendar events** — schedule meetings without leaving the terminal

---

## Quick Start

**First time?** Follow [setup.md](setup.md) to register the Azure app and build the binary.

Once the binary is built and `.env` is configured:

```bash
# List your 20 most recent emails
./outlook-assistant mail list

# Read email #3 from the last list
./outlook-assistant mail read 3

# Compose and send an email
./outlook-assistant mail send

# List upcoming calendar events (default: next 30 days, up to 20)
./outlook-assistant calendar list

# Create a new calendar event
./outlook-assistant calendar create
```

---

## Authentication

Uses **Device Code Flow** — no local web server or redirect URI needed at runtime.

On first run, the tool prints a URL and a short code:

```
To sign in, use a web browser to open the page https://microsoft.com/devicelogin
and enter the code XXXXXXXX to authenticate.
```

Open the URL in any browser, enter the code, and sign in with your Microsoft 365 account.
A `token_cache.json` file is written locally; subsequent runs reuse the cached token silently.

---

## Commands

| Command | Description |
|---------|-------------|
| `mail list [n]` | List most recent `n` emails from inbox (default: 20) |
| `mail read <n>` | Read email number `n` from the last `mail list` output |
| `mail send` | Compose and send an email interactively |
| `calendar list [n]` | List next `n` upcoming events (default: 20) |
| `calendar create` | Create a new calendar event interactively |

---

## Project Structure

```
outlook-assistant/
├── auth/
│   └── auth.go          # Azure Device Code authentication → Graph client
├── mail/
│   └── mail.go          # List, Read, Send
├── calendar/
│   └── calendar.go      # List, Create
├── main.go              # CLI dispatcher
├── .env                 # Credentials (never committed)
├── setup.md             # One-time Azure setup instructions
└── README.md            # This file
```

---

## Security Notes

- `.env` and `token_cache.json` are listed in `.gitignore` and must **never** be committed.
- The tool requests only the minimum Microsoft Graph permissions needed:
  `Mail.ReadWrite`, `Mail.Send`, `Calendars.ReadWrite`, `User.Read`, `offline_access`
- If your `CLIENT_SECRET` was ever shared in plain text, rotate it in the Azure Portal
  (**App Registrations → Certificates & secrets**) and update `.env`.

---

## Requirements

- Go 1.21+
- Microsoft 365 account (personal or organisational)
- Azure App Registration with admin consent granted (see [setup.md](setup.md))

---

## Building from Source

```bash
git clone <your-repo>
cd outlook-assistant
cp .env.example .env   # fill in CLIENT_ID, CLIENT_SECRET, TENANT_ID
go build -o outlook-assistant .
./outlook-assistant
```

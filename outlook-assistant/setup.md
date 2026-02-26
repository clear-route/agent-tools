# Setup Guide

One-time setup for ClearRoute staff. Estimated time: 10 minutes.

---

## Prerequisites

- Go 1.21 or later (`go version` to check)
- A ClearRoute Microsoft 365 account (`@clearroute.io`)
- Access to [Azure Portal](https://portal.azure.com) with permission to register apps (or ask IT to do steps 1–3)

---

## 1. Register an Azure App

1. Go to [portal.azure.com](https://portal.azure.com)
2. Search for **App registrations** → **New registration**
3. Fill in:
   - **Name**: `outlook-assistant`
   - **Supported account types**: *Accounts in this organizational directory only (Single tenant)*
   - **Redirect URI**: Type = **Web**, Value = `http://localhost:4321`
4. Click **Register**

From the **Overview** page, copy:

| Value | Where to find it |
|-------|-----------------|
| `CLIENT_ID` | Application (client) ID |
| `TENANT_ID` | Directory (tenant) ID |

---

## 2. Add API Permissions

1. Go to **API permissions** → **Add a permission** → **Microsoft Graph** → **Delegated permissions**
2. Add all of the following:
   - `Mail.ReadWrite`
   - `Mail.Send`
   - `Calendars.ReadWrite`
   - `User.Read`
3. Click **Grant admin consent for ClearRoute** → **Yes**

Each permission should show a green ✅ in the status column.

---

## 3. Configure Credentials

A `.env.example` is included in the repo with the correct values. Copy it next to the installed binary:

```bash
cp outlook-assistant/.env.example ~/.forge/tools/outlook-assistant/.env
```

No client secret is required. Authentication uses the browser-based interactive flow.

---

## 4. Build and Install

```bash
git clone https://github.com/clear-route/agent-tools.git
cd agent-tools/outlook-assistant
go build -o ~/.forge/tools/outlook-assistant/outlook-assistant .
cp tool.yaml ~/.forge/tools/outlook-assistant/tool.yaml
```

The first build downloads dependencies (~60–120 seconds). Subsequent builds take a few seconds.

---

## 5. First Run

```bash
outlook-assistant --action=list
```

On first run, your default browser opens to the Microsoft 365 sign-in page. Sign in with your `@clearroute.io` account and grant consent when prompted.

An auth record is cached at `~/.outlook-assistant-auth.json`. Subsequent runs are silent — no browser interaction until the token expires.

---

## Files Written at Runtime

| File | Purpose |
|------|---------|
| `~/.forge/tools/outlook-assistant/.env` | Your credentials — never commit |
| `~/.outlook-assistant-auth.json` | OAuth auth record — never commit |
| `~/.outlook-assistant-mail-cache.json` | Message ID cache for `--ref` index lookups |

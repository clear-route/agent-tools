# Setup Guide

One-time instructions for configuring the Azure App Registration and building the binary.

---

## Prerequisites

- Go 1.21 or later
- An active Microsoft 365 / Outlook account (e.g. `@clearroute.io`)
- Access to [Azure Portal](https://portal.azure.com) with permissions to register apps

---

## 1. Register an Azure App

1. Go to [portal.azure.com](https://portal.azure.com)
2. Search for **"App registrations"** → **New registration**
3. Fill in:
   - **Name**: `outlook-assistant` (or any name you like)
   - **Supported account types**: *Accounts in this organizational directory only (Single tenant)*
   - **Redirect URI**: `http://localhost:4321` (type: Web)
4. Click **Register**

After registration, copy the following from the **Overview** page:

| Field | Where to find it |
|-------|-----------------|
| `CLIENT_ID` | Application (client) ID |
| `TENANT_ID` | Directory (tenant) ID |

---

## 2. Create a Client Secret

1. In the app registration, go to **Certificates & secrets** → **New client secret**
2. Set a description (e.g. "outlook-assistant") and an expiry (12 or 24 months)
3. Click **Add**
4. **Copy the secret value immediately** — it won't be shown again

This becomes `CLIENT_SECRET` in your `.env`.

> ⚠️ If the secret was ever shared in plain text (e.g. in a chat), rotate it before production use:
> delete the old secret and create a new one, then update `.env`.

---

## 3. Add API Permissions

1. Go to **API permissions** → **Add a permission** → **Microsoft Graph** → **Delegated permissions**
2. Add all of the following:
   - `Mail.ReadWrite`
   - `Mail.Send`
   - `Calendars.ReadWrite`
   - `User.Read`
   - `offline_access`
3. Click **Grant admin consent for [your organisation]** → **Yes**

The status column for each permission should show a green ✅ tick.

---

## 4. Configure the `.env` File

Copy the example below and save it as `.env` in the project root:

```
CLIENT_ID=<your-application-client-id>
CLIENT_SECRET=<your-client-secret-value>
TENANT_ID=<your-directory-tenant-id>
```

The `.gitignore` already blocks `.env` from being committed.

---

## 5. Build the Binary

```bash
go build -o outlook-assistant .
```

The first build downloads dependencies and takes ~60–120 seconds.
Subsequent builds are much faster (4–10 seconds).

---

## 6. First Run & Authentication

The tool uses the **Interactive Browser Flow** — a browser window opens automatically for sign-in.

```bash
./outlook-assistant mail list
```

On first run your default browser will open to the Microsoft sign-in page. Sign in with
your Microsoft 365 account and grant consent. After authentication succeeds, a
`token_cache.json` file is written locally so you won't be prompted again until the token
expires.

---

## File Reference

| File | Purpose |
|------|---------|
| `.env` | Credentials — **never commit** |
| `token_cache.json` | OAuth token cache — **never commit** |
| `.mail_id_cache.json` | Cached message IDs for `mail read N` |
| `outlook-assistant` | Compiled binary |

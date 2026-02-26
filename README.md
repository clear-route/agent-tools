# agent-tools

A monorepo of custom [Forge](https://github.com/clear-route/forge) tools for ClearRoute staff.

Each tool lives in its own subfolder and can be installed into Forge on any local machine.

---

## Tools

| Tool | Description |
|------|-------------|
| [outlook-assistant](./outlook-assistant) | Interact with Outlook mail and calendar via Microsoft Graph API |

---

## Installing a Tool into Forge

Each tool must be compiled and placed in `~/.forge/tools/<tool-name>/` alongside its `tool.yaml`.

### Quick install (example: outlook-assistant)

```bash
cd outlook-assistant
go build -o ~/.forge/tools/outlook-assistant/outlook-assistant .
cp tool.yaml ~/.forge/tools/outlook-assistant/tool.yaml
```

Forge auto-discovers tools in `~/.forge/tools/` — no restart required.

### Credentials

Each tool documents its own credential requirements in its `README.md` and `setup.md`.
Never commit `.env` files or secrets — use environment variables or a local `.env` file outside the repo.

---

## Adding a New Tool

1. Create a subfolder: `my-tool/`
2. Add your source, a `tool.yaml` descriptor, and a `README.md`
3. Ensure `.gitignore` excludes compiled binaries, `.env`, and any token/cache files
4. Open a PR — tools are reviewed before merging to `main`

---

## Repository Structure

```
agent-tools/
├── README.md                  ← you are here
├── outlook-assistant/
│   ├── tool.yaml              ← Forge tool descriptor
│   ├── main.go
│   ├── auth/
│   ├── mail/
│   ├── calendar/
│   ├── go.mod
│   ├── README.md
│   └── setup.md
└── (future tools as additional subfolders)
```

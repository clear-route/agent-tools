---
id: mem_fdc05400-f5de-4bed-82ef-48dddb2bb9bc
created_at: 2026-03-02T23:45:14.296327Z
updated_at: 2026-03-02T23:45:14.296327Z
version: 1
scope: repo
category: architectural-decisions
session_id: f4a4905dc25e630f
trigger: cadence
---
The excalidraw-assistant is modelled on the `coleam00/excalidraw-diagram-skill` Claude Code skill. Key architecture: natural language → Excalidraw JSON (per `json-schema.md` + `element-templates.md`) → `render_excalidraw.py` (Playwright/Chromium) → PNG → visual validation loop. Reference files in that repo: `SKILL.md` (23.9 KB, main methodology), `references/json-schema.md`, `references/element-templates.md`, `references/element-templates.md`, `references/color-palette.md`, `references/render_excalidraw.py`, `references/render_template.html`, `references/pyproject.toml` (deps: playwright). Render setup: `uv sync` + `uv run playwright install chromium`.
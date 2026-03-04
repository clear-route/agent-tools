---
id: mem_bd39fec0-99a1-45f6-ba0a-d95f19ca703f
created_at: 2026-03-03T00:04:50.153077Z
updated_at: 2026-03-03T00:04:50.153077Z
version: 1
scope: repo
category: architectural-decisions
related:
    - id: mem_86644ceb-198a-4866-834c-a8c82f697d85
      relationship: refines
    - id: mem_fdc05400-f5de-4bed-82ef-48dddb2bb9bc
      relationship: refines
    - id: mem_f1452c23-1fbd-430e-a108-4a4613d90bc5
      relationship: refines
session_id: f4a4905dc25e630f
trigger: compaction
---
The `excalidraw-assistant` tool in `clear-route/agent-tools` is fully implemented with the following architecture:
- Go CLI at `/Users/justin/.forge/tools/excalidraw-assistant/excalidraw-assistant.go` with flags: `--description`, `--output`, `--render`, `--scale`, `--width`
- `--render` flag invokes Python renderer via `uv run python render_excalidraw.py` as a subprocess
- Reference files copied from `coleam00/excalidraw-diagram-skill` into `references/` subfolder: `color-palette.md`, `element-templates.md`, `json-schema.md`, `render_excalidraw.py`, `render_template.html`, `pyproject.toml`
- `getToolDir()` resolves tool directory via `os.Executable()` + `filepath.Dir()`
- `renderDiagram()` sets `cmd.Dir` to `<toolDir>/references` before invoking the renderer
- `tool.yaml` entrypoint is `excalidraw-assistant.go` (source file, not binary)
- Renderer setup: `cd ~/.forge/tools/excalidraw-assistant/references && uv sync && uv run playwright install chromium`
- A `README.md` in the tool folder serves as agent-facing documentation (methodology, design patterns, render-validate-fix loop)
---
id: mem_cfb9febd-ff82-4de3-8856-905ab4bd2b97
created_at: 2026-03-03T00:43:46.668666Z
updated_at: 2026-03-03T00:43:46.668666Z
version: 1
scope: repo
category: project-conventions
related:
    - id: mem_cc03edf4-dfca-43bd-980e-40ba4e334e76
      relationship: refines
    - id: mem_a05c74f1-296f-4f1e-866f-f33123c23f57
      relationship: refines
session_id: f4a4905dc25e630f
trigger: cadence
---
The deployed `~/.forge/tools/excalidraw-assistant/tool.yaml` must have `entrypoint: excalidraw-assistant` (the compiled binary name). This was discovered when `run_custom_tool` failed — the file had `entrypoint: excalidraw-assistant.go` (the source file name) but the deployed location needs the binary. The repo source `tool.yaml` keeps `entrypoint: excalidraw-assistant.go`; only the deployed copy uses the binary name.
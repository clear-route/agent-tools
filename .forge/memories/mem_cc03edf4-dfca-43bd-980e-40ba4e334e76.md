---
id: mem_cc03edf4-dfca-43bd-980e-40ba4e334e76
created_at: 2026-03-03T00:27:47.006247Z
updated_at: 2026-03-03T00:27:47.006247Z
version: 1
scope: repo
category: project-conventions
related:
    - id: mem_043bdd8a-c339-42e1-bcc0-84518d8e1963
      relationship: refines
    - id: mem_c88cfe82-6fef-4e4f-9c84-7311ca73f8fc
      relationship: refines
session_id: f4a4905dc25e630f
trigger: cadence
---
The `excalidraw-assistant` tool's `tool.yaml` entrypoint in the **deployed** Forge location (`~/.forge/tools/excalidraw-assistant/tool.yaml`) must be set to `excalidraw-assistant` (the compiled binary name), NOT `excalidraw-assistant.go`. The repo source `tool.yaml` uses the `.go` source file; only the deployed copy uses the binary name. This was discovered when `run_custom_tool` failed until the entrypoint was corrected.
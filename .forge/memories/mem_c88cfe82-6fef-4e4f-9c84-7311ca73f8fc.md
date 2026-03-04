---
id: mem_c88cfe82-6fef-4e4f-9c84-7311ca73f8fc
created_at: 2026-03-03T00:04:50.153077Z
updated_at: 2026-03-03T00:04:50.153077Z
version: 1
scope: repo
category: project-conventions
session_id: f4a4905dc25e630f
trigger: compaction
---
For Forge custom tools in `clear-route/agent-tools`, the `tool.yaml` entrypoint should reference the `.go` source file (e.g. `excalidraw-assistant.go`) rather than a compiled binary name — Forge appears to compile/run `.go` entrypoints directly.
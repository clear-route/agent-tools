---
id: mem_9c2d264e-8c8f-4fd4-8ff2-cc2f4ab88b5e
created_at: 2026-02-26T06:23:00.709721Z
updated_at: 2026-02-26T06:23:00.709721Z
version: 1
scope: repo
category: project-conventions
related:
    - id: mem_7248ef9f-1da8-44db-9ca8-4e559fc9939d
      relationship: refines
session_id: 5e8b645a0da65b2d
trigger: compaction
---
The outlook-assistant `Forward` body combiner uses `ExtractBodyContent` (from `mail/formatting.go`) to strip outer `<html>`/`<body>` tags from the original message before splicing in the new body HTML with an `<hr>` separator. The old `strings.TrimSuffix(…, "</body></html>")` approach was fragile and has been replaced.
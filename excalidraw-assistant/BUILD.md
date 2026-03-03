# Build & Installation Instructions

## Prerequisites

- Go 1.21+ (for compiling the tool binary)
- Python 3.8+ (for the PNG renderer)
- `uv` (Python package manager: https://github.com/astral-sh/uv)

## Build Steps

1. **Clone the repository** (if not already cloned):
   ```bash
   git clone https://github.com/clear-route/agent-tools.git
   cd agent-tools/excalidraw-assistant
   ```

2. **Build the Go binary**:
   ```bash
   go build -o excalidraw-assistant excalidraw-assistant.go
   ```

3. **Set up the Python renderer** (one-time setup):
   ```bash
   cd references
   uv sync
   uv run playwright install chromium
   cd ..
   ```

## Installation to Forge

Copy the built tool to your Forge tools directory:

```bash
# Create the tool directory
mkdir -p ~/.forge/tools/excalidraw-assistant

# Copy all necessary files
cp -r excalidraw-assistant ~/.forge/tools/excalidraw-assistant/
cp -r references ~/.forge/tools/excalidraw-assistant/
cp tool.yaml ~/.forge/tools/excalidraw-assistant/
cp README.md ~/.forge/tools/excalidraw-assistant/
```

Forge will automatically discover the tool from `~/.forge/tools/`.

## Verification

Test the tool installation:

```bash
# Test diagram generation (no rendering)
~/.forge/tools/excalidraw-assistant/excalidraw-assistant \
  --description "Simple test diagram with a rectangle" \
  --output /tmp/test.excalidraw

# Test with PNG rendering
~/.forge/tools/excalidraw-assistant/excalidraw-assistant \
  --description "Simple test diagram with a rectangle" \
  --output /tmp/test.excalidraw \
  --render

# Verify output files exist
ls -lh /tmp/test.excalidraw /tmp/test.png
```

## Troubleshooting

**"No such file or directory" when running the tool:**
- Ensure you ran `go build` in the `excalidraw-assistant/` directory
- Verify the binary exists: `ls -lh excalidraw-assistant`

**Renderer fails with "playwright not found":**
- Run the renderer setup: `cd references && uv sync && uv run playwright install chromium`
- Verify Chromium installation: `uv run playwright show-browsers`

**"references/render_excalidraw.py not found":**
- Ensure you copied the entire `references/` directory during installation
- The tool expects `references/` to be a sibling directory to the binary

## Architecture Notes

The tool consists of two components:

1. **Go binary** (`excalidraw-assistant`): Parses CLI flags, generates Excalidraw JSON from descriptions
2. **Python renderer** (`references/render_excalidraw.py`): Converts `.excalidraw` JSON to PNG using Playwright/Chromium

When `--render` is specified, the Go binary invokes the Python renderer as a subprocess.

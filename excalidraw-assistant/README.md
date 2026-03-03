# Excalidraw Assistant

Generate beautiful Excalidraw diagrams from natural language descriptions using visual argument methodology.

## Using This Tool

This tool is available as a custom tool. Invoke it using the `run_custom_tool` with `tool_name: excalidraw-assistant` and provide arguments as parameters:

- `description` (required): Natural language description of the diagram to create
- `output` (required): Output file path for .excalidraw JSON
- `render` (optional): Set to true to also generate PNG output
- `scale` (optional): Render scale factor (default: 2)
- `width` (optional): Render viewport width (default: 1920)

**IMPORTANT:** Before generating any diagram, you must read the reference files located at `~/.forge/tools/excalidraw-assistant/references/`:
- `color-palette.md` - All color definitions (read this FIRST, never hardcode colors)
- `element-templates.md` - JSON templates for each element type
- `json-schema.md` - Complete Excalidraw JSON format specification

These reference files contain the exact specifications needed to generate valid Excalidraw JSON.

## Core Philosophy

Use this guide as your source of truth for generating beautiful excalidraw diagrams for your user.

**Diagrams should ARGUE, not DISPLAY.**

A diagram isn't formatted text. It's a visual argument that shows relationships, causality, and flow that words alone can't express. The shape should BE the meaning.

**The Isomorphism Test**: If you removed all text, would the structure alone communicate the concept? If not, redesign.

**The Education Test**: Could someone learn something concrete from this diagram, or does it just label boxes? A good diagram teaches—it shows actual formats, real event names, concrete examples.



## Design Methodology

### 1. Depth Assessment

Before designing, determine what level of detail the diagram needs:

**Simple/Conceptual Diagrams** - Use abstract shapes when:
- Explaining a mental model or philosophy
- The audience doesn't need technical specifics
- The concept IS the abstraction (e.g., "separation of concerns")

**Comprehensive/Technical Diagrams** - Use concrete examples when:
- Diagramming a real system, protocol, or architecture
- The diagram will be used to teach or explain
- The audience needs to understand what things actually look like
- You're showing how multiple technologies integrate

**For technical diagrams, you MUST include evidence artifacts** (see below).

### 2. Research Mandate (For Technical Diagrams)

Before drawing anything technical, research the actual specifications:

1. Look up the actual JSON/data formats
2. Find the real event names, method names, or API endpoints
3. Understand how the pieces actually connect
4. Use real terminology, not generic placeholders

Bad: "Protocol" → "Frontend"  
Good: "AG-UI streams events (RUN_STARTED, STATE_DELTA, A2UI_UPDATE)" → "CopilotKit renders via createA2UIMessageRenderer()"

### 3. Evidence Artifacts

Evidence artifacts are concrete examples that prove your diagram is accurate and help viewers learn.

**Types of evidence artifacts:**

| Artifact Type | When to Use | How to Render |
|---------------|-------------|---------------|
| **Code snippets** | APIs, integrations, implementation details | Dark rectangle + syntax-colored text |
| **Data/JSON examples** | Data formats, schemas, payloads | Dark rectangle + colored text |
| **Event/step sequences** | Protocols, workflows, lifecycles | Timeline pattern (line + dots + labels) |

Use colors from `~/.forge/tools/excalidraw-assistant/references/color-palette.md` for evidence artifact styling.

### 4. Layout Principles

**Spatial meaning:**
- **Vertical flow** = time/sequence (top → bottom)
- **Horizontal flow** = parallel processes or choices
- **Proximity** = related concepts
- **Separation** = distinct concerns

**Alignment and spacing:**
- Align related elements to the same grid
- Use consistent spacing (100px for loose grouping, 50px for tight grouping)
- Leave breathing room—don't cram

**Visual hierarchy:**
- Larger shapes = more important concepts
- Darker colors = focal points
- Text size: 20px for titles, 16px for labels, 14px for details

### 5. Visual Pattern Library

**Shape meanings:**

| Shape | Meaning | When to Use |
|-------|---------|-------------|
| Rectangle | Process, component, action | Default for most elements |
| Rounded rectangle | External system, API, service | Things outside your control |
| Ellipse | Entry/exit point, user, start/end | Beginnings and endings |
| Diamond | Decision point, conditional | When flow branches |
| Dark rectangle | Evidence artifact | Code snippets, data examples |

**Arrow patterns:**

- **Solid arrow** = main flow, primary relationship
- **Dashed arrow** = optional, conditional, or secondary flow
- **Bidirectional** = two-way communication
- **No arrow (just line)** = structural grouping, not flow

### 6. Color as Meaning

**All colors come from `~/.forge/tools/excalidraw-assistant/references/color-palette.md`.**

You MUST read the color palette file using read_file before generating any diagram. Never hardcode color values. Use semantic colors:
- Primary fill/stroke for main elements
- Secondary fill/stroke for supporting elements
- Warning/success colors for status indicators
- Evidence artifact background for code/data blocks
- Text colors based on contrast (light/dark backgrounds)

### 7. Multi-Section Workflow (For Large Diagrams)

For comprehensive diagrams, generate JSON section by section:

1. **Section 1**: Core flow (main process)
2. **Section 2**: Supporting details (side processes)
3. **Section 3**: Evidence artifacts (code/data examples)
4. **Section 4**: Annotations and refinements

This prevents token limit issues and allows iterative refinement.

### 8. Render-Validate-Fix Loop

After generating JSON:

1. Render to PNG using the Python renderer
2. Visually inspect the result
3. Identify issues (alignment, spacing, color, clarity)
4. Apply fixes to the JSON
5. Re-render and validate
6. Repeat until diagram is publication-ready

## Reference Files Location

All reference files are located at `~/.forge/tools/excalidraw-assistant/references/`:

- `color-palette.md` - Brand colors and semantic color usage (READ THIS FIRST)
- `element-templates.md` - Copy-paste JSON templates for all element types
- `json-schema.md` - Excalidraw JSON format reference
- `render_excalidraw.py` - Python + Playwright PNG renderer (invoked automatically with render=true)

**You must read these files using read_file before generating diagrams.** The color palette especially is mandatory reading.

## Best Practices

1. **Read the reference files first** - Use read_file to load `~/.forge/tools/excalidraw-assistant/references/color-palette.md`, `element-templates.md`, and `json-schema.md` before generating any diagram
2. **Never hardcode colors** - All color values must come from the color palette file
3. **Use evidence artifacts** - For technical diagrams, include real code/data examples
4. **Think spatially** - Use position and proximity to convey meaning
5. **Iterate visually** - Use render=true to generate PNG, then refine the JSON based on visual inspection
6. **Keep it focused** - One clear argument per diagram

## Common Patterns

**Timeline/sequence:**
- Vertical line down the left
- Small dots marking each step
- Labels to the right of dots
- Arrows showing flow between steps

**System architecture:**
- Rounded rectangles for external services
- Regular rectangles for internal components
- Arrows showing data flow
- Dark rectangles for API contracts/data formats

**Decision tree:**
- Diamond shapes for decision points
- Arrows labeled with conditions
- Rectangles for resulting actions
- Clear visual branching

**Protocol flow:**
- Two columns (client/server or sender/receiver)
- Horizontal arrows between columns
- Labels showing actual message names
- Evidence artifacts showing payload formats

# Excalidraw Assistant

Generate beautiful Excalidraw diagrams from natural language descriptions using visual argument methodology.

## Core Philosophy

**Diagrams should ARGUE, not DISPLAY.**

A diagram isn't formatted text. It's a visual argument that shows relationships, causality, and flow that words alone can't express. The shape should BE the meaning.

**The Isomorphism Test**: If you removed all text, would the structure alone communicate the concept? If not, redesign.

**The Education Test**: Could someone learn something concrete from this diagram, or does it just label boxes? A good diagram teaches—it shows actual formats, real event names, concrete examples.

## Usage

The tool accepts a natural language description and generates Excalidraw JSON (and optionally renders to PNG):

```bash
excalidraw-assistant --description "Create a diagram showing OAuth 2.0 authorization code flow" --output auth-flow.excalidraw [--render]
```

### Parameters

- `--description` (required): Natural language description of the diagram to create
- `--output` (required): Output file path for .excalidraw JSON
- `--render`: Also generate PNG output using the renderer
- `--scale`: Render scale factor (default: 2)
- `--width`: Render viewport width (default: 1920)

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

Use colors from `references/color-palette.md` for evidence artifact styling.

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

**All colors come from `references/color-palette.md`.**

Read the color palette file before generating any diagram. Use semantic colors:
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

## Reference Files

- `references/color-palette.md` - Brand colors and semantic color usage
- `references/element-templates.md` - Copy-paste JSON templates for all element types
- `references/json-schema.md` - Excalidraw JSON format reference
- `references/render_excalidraw.py` - Python + Playwright PNG renderer

## Renderer Setup

The PNG renderer requires Python and Playwright:

```bash
cd ~/.forge/tools/excalidraw-assistant/references
uv sync
uv run playwright install chromium
```

## Examples

**Simple conceptual diagram:**
```bash
excalidraw-assistant --description "Three-layer architecture: presentation, business logic, data access" --output architecture.excalidraw
```

**Technical diagram with evidence:**
```bash
excalidraw-assistant --description "OAuth 2.0 authorization code flow with actual HTTP requests and JSON response formats" --output oauth-flow.excalidraw --render
```

## Best Practices

1. **Read the color palette first** - Always check `references/color-palette.md` before generating
2. **Use evidence artifacts** - For technical diagrams, include real code/data examples
3. **Think spatially** - Use position and proximity to convey meaning
4. **Iterate visually** - Render and refine until the diagram teaches on its own
5. **Keep it focused** - One clear argument per diagram

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

# Excalidraw Assistant

A validation and rendering engine for Excalidraw diagrams.

## Overview

This tool is a **rendering and validation engine** for Excalidraw diagrams. You (the agent) are responsible for generating the Excalidraw JSON using your LLM capabilities. This tool validates that JSON against the Excalidraw schema and renders it to PNG, providing feedback for you to iterate on.

## How It Works

**Your Role:**
- Read the reference files in the `references/` directory located inside the tool's installation folder to understand colors, element templates, and JSON structure
- Generate valid Excalidraw JSON based on user requirements and the design methodology below
- Write the JSON to a `.excalidraw` file
- Pass that file to this tool for validation and rendering

**Tool's Role:**
- Validates JSON structure against Excalidraw schema
- Renders the diagram to PNG
- Provides error feedback for iteration

This creates a **render-validate-fix loop** where you generate diagrams, the tool validates and renders them, and you iterate based on feedback.

## Reference Files

Before creating diagrams, read these files from the `references/` directory inside the tool's installation folder:

- `color-palette.md` - Semantic color palette with fill/stroke pairs
- `element-templates.md` - JSON templates for each Excalidraw element type
- `json-schema.md` - Complete Excalidraw JSON schema reference

These files are your source of truth for generating valid diagrams.

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

**Default to comprehensive.** If in doubt, err on the side of showing real examples and concrete details.

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

#### Information Architecture
- **Containers communicate scope**: Outer boxes = broader context. Inner boxes = implementations.
- **Shape conveys nature**: Rectangles for systems, rounded rectangles for processes, diamonds for decisions, ellipses for start/end
- **Color indicates category**: Use the semantic color palette from `references/color-palette.md` consistently

#### Spatial Relationships
- **Vertical flow = time/sequence**: Top to bottom shows what happens first, second, third
- **Horizontal grouping = categorization**: Things at the same level are peers or alternatives
- **Nesting = containment**: One box inside another shows ownership or implementation detail
- **Proximity = coupling**: Related elements should be visually close

#### Visual Pattern Library

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

#### Data Flow & Interactions
- **Arrows show causality**: Not just connection, but what causes what
- **Arrow labels should be concrete**: Show real method names, HTTP verbs, actual event types
- **Bidirectional flows need both arrows**: Don't use a single line for request/response

### 5. Text Strategy

**Every text element should add information, not just label.**

Good text:
- `POST /api/token` (shows the actual endpoint)
- `{access_token, refresh_token}` (shows what's actually returned)
- `userId: "user123"` (shows real data structure)

Bad text:
- "Authentication Service" (just labels a box)
- "Database" (doesn't show what's stored)
- "Request" (doesn't show what kind)

**Text hierarchy:**
1. **Headers** (20-24px): Section names, major component labels
2. **Body** (16-18px): Primary content, data examples, code snippets
3. **Annotations** (12-14px): Notes, clarifications, edge case details

All text should use `fontFamily: 3` (monospace) for technical diagrams to maintain the code-like aesthetic.

### 6. Color Semantics

**READ `references/color-palette.md` for all color values.** Never hardcode colors.

The palette provides semantic categories:
- **Primary**: Main system components, default elements
- **Start/Trigger**: Entry points, user actions, initiating events
- **End/Success**: Terminal states, success outcomes, completed flows
- **AI/LLM**: AI components, LLM calls, intelligent processing
- **Process**: Active processing, transformations, computations
- **Data**: Storage, databases, data structures
- **Warning**: Caution points, rate limits, potential issues
- **Validation**: Checks, guards, validation steps
- **Text hierarchy**: Headers, body, annotations

Each category has a fill color and stroke color pair designed to work together.

### 7. Design Principles

**Start with the main path, then add variations.**

1. Draw the happy path / normal flow first (Core flow)
2. Add error cases and edge cases as branches
3. Use color to distinguish normal vs exceptional flows
4. Group related elements with visual proximity

**Use whitespace intentionally:**
- Dense spacing for tightly coupled elements
- Wide spacing to show separation of concerns
- Consistent spacing within groups for visual rhythm

**Alignment matters:**
- Align related elements to show grouping
- Stagger elements to show sequence
- Center important nodes to show focus

#### Multi-Section Workflow (For Large Diagrams)

For comprehensive diagrams, generate JSON section by section:

1. **Section 1**: Core flow (main process)
2. **Section 2**: Supporting details (side processes)
3. **Section 3**: Evidence artifacts (code/data examples)
4. **Section 4**: Annotations and refinements

This prevents token limit issues and allows iterative refinement.

### 8. Common Patterns

#### Flow Diagrams
- Start with a trigger (start/trigger color)
- Show the sequence top-to-bottom
- Use arrows with concrete labels
- End with outcomes (success/error colors)
- Include actual data formats in transit

#### System Architecture
- Outer container for system boundary
- Inner boxes for components
- Show actual protocols on connections
- Include concrete examples of requests/responses
- Use color to distinguish layers (UI, logic, data)

#### State Machines
- States as rounded rectangles
- Transitions as labeled arrows
- Show actual event names on transitions
- Highlight start state and end states with semantic colors
- Include example data that triggers transitions

### 9. Rendering & Validation

After generating your JSON:
1. Write it to a `.excalidraw` file
2. Pass it to this tool for validation
3. Review the rendered PNG
4. Look for these common issues:
   - Text elements without `containerId` (orphaned labels)
   - Overlapping elements (adjust x/y coordinates)
   - Inconsistent spacing (use multiples of 20 for grid alignment)
   - Wrong color palette (verify against `color-palette.md`)
   - Missing arrow labels (add concrete method/event names)
5. Repeat the generate → render → observe → refine loop until the diagram is clear, accurate, and visually appealing.

**The render loop is your feedback mechanism.** Don't try to get it perfect in one shot. Generate → render → observe → refine.

### 10. Technical Requirements

**All elements must include:**
- `id` (unique string)
- `type` (rectangle, ellipse, diamond, arrow, text, line)
- `x`, `y` (position coordinates)
- `width`, `height` (dimensions)
- `strokeColor`, `backgroundColor` (from color palette)
- `fillStyle` (solid, hachure, cross-hatch)
- `strokeWidth`, `strokeStyle`, `roughness`, `opacity`
- `seed`, `version`, `versionNonce` (for rendering consistency)
- `isDeleted`, `groupIds`, `boundElements`, `link`, `locked`

**Text elements must:**
- Set `containerId` to parent shape ID (for labels inside boxes)
- Parent shape must have `boundElements` array listing the text element
- Use `fontFamily: 3` (monospace) for technical content
- Set `textAlign` and `verticalAlign` appropriately

**Arrow elements must:**
- Define `points` array for arrow path
- Set `startBinding` and `endBinding` to connect to shapes
- Include `endArrowhead: "arrow"` for directional arrows
- Use `text` property for arrow labels

**Rounded rectangles must:**
- Include `roundness: {"type": 3}` for corner rounding

See `references/element-templates.md` for complete JSON templates and `references/json-schema.md` for the full schema specification.

## Examples

The tool comes with reference examples in `.tmp/excalidraw-diagram-skill/examples/` showing:
- OAuth flows with actual endpoint calls
- Multi-agent systems with real message formats
- Database schemas with concrete field names
- State machines with example transitions

Study these examples to understand the level of detail and visual argument principles in practice.

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

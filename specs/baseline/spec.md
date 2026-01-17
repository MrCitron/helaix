# HelAIx Core Application Specification

HelAIx is an AI-powered preset engineer for the Line 6 Helix family. It bridges the gap between musical intent and technical implementation by using LLMs to generate hardware-compatible presets.

## Visual Requirements

### Layout & Navigation
- **Collapsible Sidebar**: 
    - Full view (labels + icons) vs. compact view (icons only).
    - Chat history with deletion capabilities and confirmation modals.
    - Responsive auto-collapse at widths below 1024px.
- **Main Chat Area**:
    - Centralized feed with assistant identity (helAIx assistant).
    - Contextual message bubbles for user and AI.
    - Sticky chat input at the bottom with high z-index to prevent clipping.

### Signal Chain Visualization
- **Dual Representation**:
    - **Real Chain**: Concept-level design proposed by the assistant.
    - **Helix Chain**: Technical implementation of the preset.
- **Iconography**:
    - Authentic Line6-inspired SVG icons for all categories (Amp, Cab, Distortion, etc.).
    - Category-based coloring (e.g., Green for Delay, Purple for Wah).
- **Interactivity**:
    - Selectable blocks with inline parameter expansion (vertical mode) or docked bottom view (horizontal mode).
    - Visual indicators (progress bars) for numeric parameters.

## Functional Requirements

### AI & Preset Engineering
- **Multi-Stage Workflow**:
    - Sound design stage (conceptualizing gear).
    - Preset engineering stage (technical mapping to Helix parameters).
- **Hardware-Aware DSP Management**:
    - Constrain preset complexity based on target hardware (Helix Floor, LT, Stomp, etc.).
    - Automatic distribution of blocks across Path 1 and Path 2 for dual-DSP units.
    - Offline model catalog with precise DSP weights (v3.80).

### Data & Configuration
- **Model Support**: Google Gemini API (Pro and Flash versions).
- **Preset Export**:
    - Generation of standard `.hlx` compatible JSON files.
    - **Incremental Versioning**: Option to add suffixes (e.g., `_1.hlx`) to prevent accidental overwrites.
- **Settings Management**:
    - Persistence for API keys, provider selection, and default output paths.
    - Internationalization (i18n) support for English and French.

## Edge Cases

- **DSP Overload**:
    - If the AI generates a rig exceeding 100% on a single path, the system must either prompt for simplification or attempt auto-rebalancing across paths if hardware supports it.
- **Invalid API Key/Provider**:
    - Graceful degradation of chat functionality with clear prompts leading to the Settings page.
- **Large/Complex Rigs**:
    - Vertical stacking of the signal chain visualizer when horizontal space is insufficient (Deep Responsiveness).
- **File System Collisions**:
    - Handling existing files based on user-selected strategy (Overwrite vs. Increment).
- **Model Incompatibility**:
    - Handling models in the catalog that might not exist in older firmware versions (pinned to v3.80 base).

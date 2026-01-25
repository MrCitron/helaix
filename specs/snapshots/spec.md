# Feature: Snapshot Support & Variax Engine 2.0

## 1. Overview
HelAIx currently generates static presets with a single signal chain configuration. Users need the ability to have multiple sound variations within a single preset (Snapshots) to handle different song parts (e.g., Intro/Clean, Rhythm/Overdrive, Solo/Lead).

This feature enables the AI to propose up to 4 snapshots per preset, defining which effects are enabled/disabled and managing physical Variax hardware properties.

## 2. Functional Requirements

### 2.1 AI Design (Sound Engineer)
- **Snapshot Generation**: The Sound Engineer (AI) must define 1 to 4 snapshots for song-based requests.
- **Naming**: Snapshots must be named descriptively based on song parts or characteristics.
- **State Management**: For each snapshot, the AI specifies the `enabled` state of components.
- **Iconic Branding**: The AI must use real-world guitar names (e.g., "Fender Jaguar") for the design view while mapping them to technical Variax banks internally.
- **Variax Branding**: The Variax component in the signal chain must be labeled **"Line6 Variax"**.

### 2.2 Preset Implementation (Preset Engineer)
- **Snapshot Logic**: Translates AI descriptions into `.hlx` format (`data.tone.snapshot0` to `snapshot7`).
- **Dual-DSP Initialization**: Ensures snapshot bypass maps are initialized for both `dsp0` and `dsp1` to prevent discrepancies on Path 2 blocks.
- **Snapshot Controllers**: Parameters that change between snapshots are automatically assigned to **Snapshot Controller (ID 9)**.
- **Hardware-Aware Variax Mapping**:
    - Uses `variax_models.json` as the single source of truth.
    - Standardized JTV Bank Order: T-Model (1), Spank (2), Lester (3), etc.
    - Character-Perfect Resolution: Implements hardware-specific variant logic (e.g., JTV inverted pickup positions).

### 2.3 UI Visualization
- **Snapshot Selector**: Tabs/Pills above the signal chain to preview different states.
- **Smart Snapshot Markers**: A blue dot appears on blocks that have any parameter assigned to a snapshot controller.
- **Visual Parity**: Both "Real Chain" and "Helix Chain" must show synchronized block bypass states.

## 3. Visual & Design Standards
- **Perfect Circuit Alignment**: The signal interconnect line must precisely bisect the center-mass of the stompbox icons.
- **Vertical Offset**: Anchored at `68px` from the container top.
- **Baseline Grouping**: Blocks use `items-start` layout so icons stay locked to the line while labels grow downwards.
- **State Indication**: Disabled blocks in a snapshot use `opacity-30` and `grayscale`.

## 4. Data Model

### 4.1 RigDescription Update
```json
{
  "guitar_model": "Fender Jaguar (Pickup Pos 4)",
  "tuning": "Drop D",
  "snapshots": [
    {
      "name": "Verse (Chorus)", 
      "active_blocks": ["Line6 Variax", "Boss CE-2"],
      "guitar_model": "Fender Jaguar",
      "params": { "Line6 Variax": { "Model": "T-Model 4" } }
    }
  ]
}
```

## 5. Edge Cases
- **Path 2 Sync**: Blocks on the second DSP must be explicitly included in the snapshot bypass map to avoid "always on" bugs.
- **Variax Hardware Types**: Models and labels switch dynamically between **JTV** and **Shuriken** based on user hardware selection.

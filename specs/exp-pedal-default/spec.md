# Feature Spec: Default Expression Pedal Target

## Goal
Allow the user to define which expression pedal (Exp 1, Exp 2, Exp 3, or None) should be automatically assigned to variable parameters like "Pedal" in Wah, Volume, and Pitch blocks when generating a preset.

## User Experience
- A new setting "Default Expression Pedal" will be added to the Settings screen under the "Hardware" or "Logic" section.
- Options:
  - **None / Nothing**: No controller assignment is made.
  - **Exp 1**: Assigned to controller 1.
  - **Exp 2**: Assigned to controller 2.
  - **Exp 3**: Assigned to controller 3.

## Technical Implementation

### 1. Configuration
- **File**: `pkg/config/manager.go`
- **Struct**: `AppConfig`
- **Field**: `DefaultExpPedal int` (0 = None, 1 = Exp 1, 2 = Exp 2, 3 = Exp 3)

### 2. Frontend UI
- **File**: `frontend/src/components/Settings.jsx`
- Add a new `Select` or `Radio` group.
- **Labels**: Localized via `i18n.js`.

### 3. Backend Logic (Preset Generation)
- **File**: `pkg/gemini/preset_engineer.go`
- **Function**: `ChatPresetEngineer`
- **Logic**:
  - Iterate through blocks in the generated preset.
  - Identify blocks that typically use expression pedals:
    - Wah (e.g., `HD2_Wah...`)
    - Volume (e.g., `HD2_Vol...`)
    - Pitch Wham (e.g., `HD2_PitchWham...`)
  - If a block is identified and `DefaultExpPedal > 0`:
    - Add it to the `tone -> controller -> dsp0` map.
    - Specifically, map the `Pedal` parameter (or `Position` if appropriate) to `@controller: DefaultExpPedal`.

## Edge Cases
- **Hardware Limitations**: Some devices (like HX Stomp) only have Exp 1 & 2. If Exp 3 is selected but not supported by the hardware, the behavior should ideally be transparent (Helix will just ignore or show it as Unassigned).
- **Multiple Blocks**: If multiple Wah/Volume blocks exist, they will all be assigned to the same pedal by default.
- **Overwriting AI Intent**: This logic should override any default controller assignment specified in the `template.json` or `catalog.json` to ensure user preference is respected.

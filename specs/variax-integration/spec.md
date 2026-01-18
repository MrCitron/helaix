# Feature Spec: Variax Integration

## Goal
As a user with a Line 6 Variax guitar, I want HelAIx to automatically configure my instrument's model and tuning based on the sound I am requesting, respecting the capabilities of my specific Variax hardware.

## User Experience
1.  **Settings**: A new section "Variax Control" in Settings allows selecting:
    *   **Variax Hardware Model**: Options: JTV, Variax Standard, Shuriken. This allows the AI to know which simulations are available.
    *   **Automatic Variax Control**: Toggle to enable/disable.
2.  **Sound Engineer Agent**:
    *   Contextually aware of the user's Variax model.
    *   Suggests a **Real Guitar Model** (e.g., "1959 Stratocaster") and tuning (e.g., "Drop D") if explicitly requested **or contextually appropriate** (e.g., for specific songs or styles).
3.  **Visual Confirmation**: 
    *   **Real Chain**: Displays the icon + the real guitar name (e.g., "Stratocaster").
    *   **Helix Chain**: Displays the Variax bank/simulation name (e.g., "Spank").
4.  **Hardware Sync**: The `.hlx` file is populated with the correct `@variax_model` ID and tuning offsets for the user's specific performance environment.

## Functional Requirements
### Sound Engineer
*   Take `VariaxHardwareModel` as context.
*   Update `RigDescription` to include:
    *   `guitar_model`: string (Real Guitar Name, e.g., "Stratocaster", "Les Paul")
    *   `tuning`: string (e.g., "Drop D")

### Configuration
*   `VariaxHardwareModel` (string): "JTV", "Standard", "Shuriken".

*   Map real `guitar_model` name (e.g., "Les Paul") to Variax `@variax_model` ID (e.g., Lester) based on the hardware profile.
*   Display Variax bank name in the technical visualizer.
*   Apply 6-string tuning offsets if a tuning is suggested by the AI (either explicitly requested or contextually inferred). If no tuning is suggested, leave Variax tuning parameters unspecified (don't force).

## UI Design
- **Icon**: Replace "GUITAR" text with a custom SVG guitar icon in the `DesignVisualizer` and `PresetVisualizer`.
- **Category Color**: Purple or a custom "Instrument" color (e.g., Cyan).

## Technical Implementation Notes

### Variax Hardware Context
| Model | Simulation Range | Specific Features |
| :--- | :--- | :--- |
| **JTV** | HD Set (0-49) | 25-29 models |
| **Standard** | HD Set (0-49) | 24-28 models |
| **Shuriken** | HD Set + Shuriken Models (0-59+) | 60+ models |

### Common Simulation IDs (@variax_model)
| Bank | Variations (Pos 1-5) |
| :--- | :--- |
| **T-Model** | 0-4 |
| **Spank** | 5-9 |
| **Lester** | 10-14 |
| **Special** | 15-19 |
| **R-Billy** | 20-24 |
| **Chime** | 25-29 |
| **Semi** | 30-34 |
| **Jazzbox** | 35-39 |
| **Acoustic** | 40-44 |
| **Reso** | 45-49 |
| **Shuriken** | 50-54 (on Shuriken HW only) |

### Tuning Logic (Hardware Dependent)
Apply if requested or contextually inferred. Below are common HD tunings.

| Name | Offsets (Low to High) | Hardware Compatibility |
| :--- | :--- | :--- |
| **Drop D** | `[-2, 0, 0, 0, 0, 0]` | All |
| **Eb Standard** | `[-1, -1, -1, -1, -1, -1]` | All |
| **D Standard** | `[-2, -2, -2, -2, -2, -2]` | All |
| **Drop C** | `[-4, -2, -2, -2, -2, -2]` | JTV-89, Shuriken |
| **Baritone** | `[-5, -5, -5, -5, -5, -5]` | All (Commonly B-E-A-D-F#-B) |
| **Open G** | `[-2, -2, 0, 0, 0, -2]` | All (D-G-D-G-B-D) |
| **DADGAD** | `[-2, 0, 0, 0, 0, -2]` | All |

> [!IMPORTANT]
> If tuning is not suggested by the AI, the preset should NOT contain any tuning offsets (leave at "Don't Force").

# Support Bass Presets Specification

This feature adds comprehensive support for generating Bass presets within HelAIx. The AI assistant will be able to design and engineer presets specifically tailored for bass guitar, utilizing the correct bass amp models, cabs, and effects available in the Line 6 Helix ecosystem.

## Visual Requirements

- **Default Instrument Settings**:
  - Add a toggle or dropdown in the application settings to select between "Guitar" and "Bass" as the default instrument context.
- **Iconography Updates**:
  - Ensure bass-specific categories or tags in the visual signal chain correctly display bass-related icons if distinct from guitar.

## Functional Requirements

- **AI Context & System Prompts**:
  - Update the system prompts sent to the Gemini API to explicitly instruct it to prioritize bass amp models (e.g., SVT, Ampeg, GK, etc.) and bass cabs when the resolved instrument is Bass.
- **Instrument Resolution**:
  - The user selects a default instrument (`Guitar` or `Bass`) in application settings.
  - An explicit instrument mention in the prompt overrides that default. If the prompt does not specify an instrument, the configured default remains authoritative.
  - The Sound Engineer must return the resolved `instrument` (`Guitar` or `Bass`) in the `RigDescription`.
  - The Preset Engineer must use that resolved value and must not infer a different instrument from artist, genre, model names, or effect choices.
- **Catalog & Model Validation**:
  - Ensure the internal catalog correctly tags and lists bass-specific amps, cabs, and effects.
  - The preset engineer must properly account for DSP limits of bass models.
- **Bass-Specific Routing**:
  - Keep the routing simple. The AI should avoid splitting into dual parallel chains (e.g., dual amps) unless explicitly necessary for the tone, preferring a single chain for simplicity and DSP efficiency.

## Edge Cases

- **Ambiguous User Prompts**:
  - If a user asks for a "heavy preset" without specifying the instrument, the AI should rely on the default instrument setting in the settings.
- **Variax Pitch Simulation Constraint**:
  - Variax is available only when the resolved instrument is Guitar and Automatic Variax Control is enabled in settings.
  - When the resolved instrument is Bass, Variax is always disabled, regardless of the setting and regardless of the user's physical instrument.
  - The Sound Engineer must return `use_variax` and add a `Line6 Variax` chain component only when `use_variax` is true.
  - The Preset Engineer must use `use_variax` as the sole Variax decision and must not derive it from `guitar_model`.
- **DSP Overload with Parallel Paths**:
  - Complex dual-amp bass routings might hit DSP limits on devices like the HX Stomp. The AI must handle DSP overload by falling back to simpler single-amp solutions or prompting the user.
- **Mixed Gear Requests**:
  - If a user explicitly requests a guitar amp on a bass preset, the AI should allow it but perhaps warn about typical frequency response issues, acting as a true "preset engineer".

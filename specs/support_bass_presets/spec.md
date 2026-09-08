# Support Bass Presets Specification

This feature adds comprehensive support for generating Bass presets within HelAIx. The AI assistant will be able to design and engineer presets specifically tailored for bass guitar, utilizing the correct bass amp models, cabs, and effects available in the Line 6 Helix ecosystem.

## Visual Requirements

- **Default Instrument Settings**:
  - Add a toggle or dropdown in the application settings to select between "Guitar" and "Bass" as the default instrument context.
- **Iconography Updates**:
  - Ensure bass-specific categories or tags in the visual signal chain correctly display bass-related icons if distinct from guitar.

## Functional Requirements

- **AI Context & System Prompts**:
  - Update the system prompts sent to the Gemini API to explicitly instruct it to prioritize bass amp models (e.g., SVT, Ampeg, GK, etc.) and bass cabs when the user's intent is bass-oriented or when the default bass context is selected.
- **Dynamic Instrument Inference**:
  - The sound engineer (AI) should rely on the user's prompt to determine if they are asking for a different instrument than the default one set in settings.
- **Catalog & Model Validation**:
  - Ensure the internal catalog correctly tags and lists bass-specific amps, cabs, and effects.
  - The preset engineer must properly account for DSP limits of bass models.
- **Bass-Specific Routing**:
  - Keep the routing simple. The AI should avoid splitting into dual parallel chains (e.g., dual amps) unless explicitly necessary for the tone, preferring a single chain for simplicity and DSP efficiency.

## Edge Cases

- **Ambiguous User Prompts**:
  - If a user asks for a "heavy preset" without specifying the instrument, the AI should rely on the default instrument setting in the settings.
- **Variax Pitch Simulation Constraint**:
  - In the case of a user using a Variax as a guitar but asking for a bass preset, the sound engineer should *not* use the Variax to simulate a bass (e.g., via pitch shifting or alternate tunings on the Variax block). The preset must use standard bass amp modeling for a physical bass instrument instead.
- **DSP Overload with Parallel Paths**:
  - Complex dual-amp bass routings might hit DSP limits on devices like the HX Stomp. The AI must handle DSP overload by falling back to simpler single-amp solutions or prompting the user.
- **Mixed Gear Requests**:
  - If a user explicitly requests a guitar amp on a bass preset, the AI should allow it but perhaps warn about typical frequency response issues, acting as a true "preset engineer".

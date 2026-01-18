# Feature: Delay Feedback Safety Limit

## 1. Overview
The "Feedback" parameter in Delay effects can cause uncontrollable self-oscillation and extreme volume spikes if set too high (near 100%). To ensure a safe and usable user experience, we will enforce a hard limit of **75% (0.75)** on the Feedback parameter for all Delay blocks generated or managed by the application. This mirrors the `0.7` safety limit previously implemented for Reverb Decay.

## 2. Visual Requirements (Vision)
*N/A - Backend Logic Constraint*
- The generated presets will simply have lower feedback values.
- If the UI displays these values, they will reflect the clamped limit (e.g., showing "75%" instead of "100%").

## 3. Functional Requirements
- [ ] **Identify Delay Blocks**: The system must identify blocks that fall under the "Delay" category (typically those with internal IDs starting with `HD2_Delay` or `VIC_Delay`).
- [ ] **Clamp Feedback**: For any parameter identified as "Feedback" (or common variations like "Fdbk"), if the requested value exceeds `0.75`, it must be clamped to `0.75`.
- [ ] **Apply to Generation**: This logic must apply during the `preset_engineer` generation phase.
- [ ] **Apply to Updates**: (Optional but recommended) If we have logic updating existing blocks, this limit should apply there too.

## 4. Edge Cases
- **Parameter Naming**: Ensure we catch "Feedback", "Fdbk", "Bk", or whatever typical keys Helix uses for delays.
- **Specific Models**: Some specific/weird delays (e.g., "Cosmos Echo") might have unique parameter names.
- **User Intent**: If a user explicitly asks for "Self Oscillation", the AI might try to request 100%, but our safety layer should override this for safety, or we strictly enforce the limit regardless of intent for this version.

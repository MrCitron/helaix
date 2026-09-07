# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.11.0] - 2026-09-06

### Added
- **Codex Parameter Control**: Codex can now set validated numeric, boolean, and string block parameters to dial in a preset rather than relying on catalog defaults alone.
- **Live Schema Verification**: An opt-in ChatGPT Codex test verifies that dynamic block parameters work with the real CLI session.

### Changed
- **Build Hygiene**: CI rejects generated frontend hash files if they are accidentally added to version control.

## [0.10.0] - 2026-09-06

### Added
- **DSP-Safe Candidate Planning**: Tone-ranked Helix model combinations now stay within the per-path DSP budget before preset generation.
- **Validation-Guided Recovery**: Gemini and Codex receive deterministic correction instructions and retry one rejected block mapping with the exact validation context.

### Changed
- **Stricter Preset Engineering Prompts**: Providers now preserve the validated model/path combination, chain order, and dual-DSP path progression unless the user explicitly requests a change.

### Fixed
- **Invalid Mapping Retries**: Codex retries now include the rejected JSON response so the provider can return a complete corrected mapping.

## [0.9.0] - 2026-02-03

### Added
- **Linux Build Support**: Comprehensive guide for building on Linux (Ubuntu/Zorin OS).

### Fixed
- **Instant Connection Testing**: "Test Connection" and model listing in Settings now work immediately after entering an API key, without requiring a save.

## [0.8.0] - 2026-01-27

### Added
- **Multi-Snapshot Support**: AI now generates song-specific snapshots (e.g., Intro, Verse, Chorus) with independent bypass states and parameter shifts.
- **Snapshot Visualization**: New UI selector for snapshots and real-time visualization of dimmed/active blocks.
- **Character-Perfect Variax Engine 2.0**:
    - **Physical Calibration**: Standardized JTV bank order (T-Model, Spank, Lester) to match physical hardware MIDI mapping.
    - **Hardware-Aware UI**: Dynamic bank labels that adapt to your Variax model (JTV vs Shuriken).
    - **Real-World Branding**: AI now uses iconic guitar names (e.g., "Fender Jaguar") in metadata for a professional design view.
- **"Perfect Circuit" UI Alignment**: Standardized signal path lines to precisely bisect stompbox icons at a fixed 68px offset.

### Changed
- **Baseline Alignment**: Switched signal chain layout to `items-start` for better icon stability across varying label lengths.

## [0.7.0] - 2026-01-18

### Added
- **Delay Feedback Safety Limit**: Lowered the maximum internal feedback value to **0.75* (75% in UI) to limit self-oscillation.

### Changed
- **Application Refactoring**: Moved application code to 'app' subdirectory and updated build config.

## [0.6.0] - 2026-01-18

### Added
- **Deep Variax Integration**: Full support for Variax modeling in preset generation and visualization.
- **Hardware Selection**: Added settings to specify Variax hardware type (Standard, JTV, Shuriken) for accurate model mapping.
- **Real Guitar Names**: The AI now understands and maps iconic guitar names (e.g., "Les Paul", "Stratocaster") to the correct Variax models.
- **Context-Aware Tuning**: Automated tuning suggestions (e.g., "Drop D" for relevant styles) based on Variax capabilities.
- **Real Chain Visualizer**: New "Guitar" header display in the chain visualizer showing the simulated instrument and tuning.

## [0.5.0] - 2026-01-17

### Added
- **Default Expression Pedal Selection**: Users can now define a default target controller (Exp 1, Exp 2, Exp 3, or Nothing) for effects like Wah, Volume, and Pitch Wham in the Settings.
- **Improved Internationalization**: Added French and English translations for the new expression pedal settings.
- **Hardware-Aware AI Prompt**: Refined the AI system prompt to better handle hardware path distribution and parameter constraints.

### Changed
- **Reverb Decay Safety Limit**: Lowered the maximum internal decay value to **0.7** (7.0 in UI) for all Reverb blocks to prevent noise and feedback loops.
- **Refined Signal Chain Logic**: Improved internal handling of DSP paths and controller assignments in the `preset_engineer`.

### Fixed
- **Anti-Noise Logic**: Specific sanitization added to Heliosphere and other Delay models with reverb tails to respect the new decay safety cap.

## [0.4.0] - 2026-01-09

### Added
- **Incremental Save Strategy**: Option to add numbered suffixes (e.g., `_1`, `_2`) to exported presets instead of overwriting.
- **Helix Model Constraints**: Ability to select specific hardware (Floor, LT, Stomp) to constrain AI-generated preset complexity.
- **Dual-Path Support**: Automatic distribution of blocks across Path 1 & 2 for dual-DSP units.
- **Branding**: Official HelAIx icon and blue theme implemented across the application and assets.

### Changed
- **README Overhaul**: Comprehensive documentation on features, setup, and hardware safety.
- **CI/CD Pipeline**: GitHub Actions for automated Windows (.exe) and macOS (.app) releases.

### Fixed
- **UI Clipping**: Resolved z-index conflicts and overflow issues in the sidebar and chat visualizers.
- **Sidebar Responsiveness**: Automatic collapse on smaller screens.

## [0.3.0] - 2026-01-07
- Initial Beta release with Core Rig Design and Preset Engineering logic.

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

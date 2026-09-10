# HelAIx Evolution Roadmap

This roadmap captures the next evolution areas for HelAIx. It is intentionally
prioritized by expected business value relative to implementation effort, not
by technical interest alone.

## Scoring Model

- **Business value (BV):** expected impact on user experience, adoption,
  reliability, maintainability, or delivery speed, scored from 1 (low) to 10
  (very high).
- **Effort:** relative implementation and validation effort, scored from 1
  (small) to 10 (very large). The score includes design, implementation,
  testing, documentation, and cross-platform validation.
- **BV/Effort:** BV divided by effort. Higher values should generally be
  scheduled first, subject to dependencies and risk.

These are initial estimates for one engineer familiar with the codebase. They
should be recalibrated after discovery, especially for OS integration and
provider API work.

## Prioritized Backlog

| Priority | Initiative | Scope and key outcomes | Effort | BV | BV/Effort | Dependencies / risks |
| ---: | --- | --- | ---: | ---: | ---: | --- |
| 1 | **Persistent and sensible window sizing** | Define a usable default size and minimum size; restore the last position and dimensions on startup; save state safely on close; handle invalid, off-screen, or multi-monitor values on Windows, macOS, and Linux. | 3 | 8 | **2.67** | Requires a small native/Wails integration and cross-platform verification. |
| 2 | **README refresh and new screenshots** | Update setup, supported providers, current capabilities, limitations, workflow, and troubleshooting. Replace or supplement screenshots so they reflect the current UI and export flow. Validate image paths and captions. | 2 | 5 | **2.50** | Best done after the highest-impact UI and provider changes land. |
| 3 | **Update existing specifications** | Audit every current spec against the application as it exists today; document implemented behavior, remove stale assumptions, add missing acceptance criteria, and identify gaps for the roadmap items. Keep specs, code, and README terminology consistent. | 4 | 8 | **2.00** | Should start early and continue as feature behavior changes. High ratio, but lower direct user visibility than product features. |
| 4 | **Additional OpenCode delivery commands** | Add project-local commands for preparing/creating a pull request and for release execution or release preparation. Reuse existing release and pre-PR documentation, enforce checks, avoid secrets, and make destructive/publishing steps explicit and confirmable. | 3 | 6 | **2.00** | Must align with repository permissions, GitHub workflow behavior, and the existing `scripts/release.mjs`. |
| 5 | **Provider abstraction and OpenCode Go provider** | Make provider configuration explicit and extensible; add OpenCode Go as an LLM provider with API-key configuration, secure local storage, model selection, clear errors, and a connectivity/test path. Keep Gemini working and avoid exposing keys to logs or generated frontend artifacts. | 7 | 10 | **1.43** | Confirm the OpenCode Go API contract, authentication scheme, supported models, limits, and data-handling expectations before implementation. |
| 6 | **Full codebase review with a current, stronger LLM** | Review Go, React, Wails integration, prompts, `.hlx` generation, tests, security, error handling, and cross-platform behavior with a more capable current LLM than the original review. Convert findings into actionable issues, then fix and verify high-severity findings rather than treating the review as a one-off report. | 6 | 8 | **1.33** | Needs a defined review rubric, current build/test baseline, and human validation of suggested changes. Do not delegate security or product decisions blindly. |
| 7 | **User-controlled iterative signal-chain workflow** | Let users correct the sound engineer output and the preset engineer output in context, for example changing an amp, removing reverb, or rejecting the rig and requesting a new attempt. Preserve the current chain and user corrections between steps, show what changed, and validate the resulting `.hlx` preset. | 8 | 10 | **1.25** | Requires prompt/state redesign, regression coverage, and careful handling of invalid or contradictory requests. |
| 8 | **Dependency and toolchain refresh** | Audit and update Go, Wails, React, Vite, Tailwind, icons, and other direct/transitive dependencies; remove obsolete workarounds; update lockfiles and build documentation; run tests and builds on supported OSes. Prefer small, separable upgrade batches. | 6 | 7 | **1.17** | May expose breaking changes or platform-specific regressions. Avoid changing core build configuration without a tested need. |
| 9 | **Export experience and host-application integration** | Improve export feedback and recovery; offer “open file” using OS-native file handling; investigate drag-and-drop to HX Edit or Helix Native; clearly document what is and is not possible. Treat prior failed attempts as constraints, and design platform-specific adapters only where necessary. | 8 | 9 | **1.13** | HX Edit/Helix Native may not expose a stable external drop/open protocol. Must validate on Windows, macOS, and Linux best effort without shell-specific assumptions. |

## Recommended Delivery Sequence

The ratio ranking is a prioritization aid, not a strict implementation order.
The following sequence reduces rework and risk:

1. Establish a baseline: run tests/builds, audit the existing specs, and run
   the LLM code review. Record findings separately from fixes.
2. Implement persistent window sizing and the OpenCode delivery commands. These
   are bounded improvements with good value-to-effort and help the workflow.
3. Refresh dependencies in small batches, resolving build and platform issues
   before adding larger product behavior.
4. Define the provider abstraction and integrate OpenCode Go. Add provider
   contract tests and keep credentials out of source, logs, and frontend
   bundles.
5. Implement iterative sound-engineer and preset-engineer corrections. Use the
   updated specs as the acceptance contract and add regression fixtures for
   representative chains.
6. Rework export and investigate native application handoff. Time-box drag and
   drop research and ship a reliable direct-open/export experience even if
   third-party drag and drop is not technically supported.
7. Refresh the README and capture screenshots from the resulting application.
8. Re-run the code review, update all specs, and perform the final supported-OS
   verification before release.

## Definition of Done

- Each completed initiative has an updated spec or an explicit rationale for
  why no spec is needed.
- Go tests, frontend build, and relevant platform builds pass, or known limits
  are documented.
- API keys and generated presets are handled without leaking secrets.
- User-visible behavior is documented in the README and screenshots where
  appropriate.
- Cross-platform behavior is tested on available Windows, macOS, and Linux
  environments, with gaps recorded rather than assumed away.
- The backlog is re-scored after discovery of the OpenCode Go API and the
  third-party HX Edit/Helix Native integration constraints.

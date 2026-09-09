---
description: Verify a feature's logic, UI requirements, and release readiness.
agent: build
---

Verify the feature named `$ARGUMENTS`.

1. Read `specs/[feature_name]/spec.md`, `specs/[feature_name]/plan.md` if present, `AGENTS.md`, and `memory/constitution.md`.
2. Run the relevant backend tests, frontend checks, and build commands available in this repository. Do not invent platform-specific commands; keep checks portable across Windows, macOS, and Linux.
3. Exercise the happy path and inspect console or build errors.
4. For UI work, use an available browser or screenshot-capable tool if configured. If OpenCode has no browser tool in the current session, explicitly report that visual verification could not be automated and give the manual verification steps instead of claiming success.
5. Compare results with the specification and report each requirement as `Match`, `Discrepancy`, or `Not verified`.
6. Write the release-readiness result to `specs/[feature_name]/verification.md` only when this file is requested by the user or already exists; otherwise report it in the response.

---
description: Implement the unchecked steps in a feature plan.
agent: build
---

Implement the feature named `$ARGUMENTS`.

1. Read `AGENTS.md`, `docs/STYLEGUIDE.md`, `memory/constitution.md`, and `specs/[feature_name]/plan.md` if it exists. If no plan exists, inspect the specification and proceed with a minimal implementation plan in the response before editing.
2. Execute only unchecked items in the plan's Step-by-Step section, preserving the existing architecture and cross-platform support.
3. Use Wails bindings for internal frontend/backend communication; do not add raw HTTP or WebSocket connections.
4. After each logical group of changes, run the most relevant tests, type checks, or build checks.
5. Update completed plan items from `- [ ]` to `- [x]` only after verification.
6. Do not modify unrelated user changes. Report changed files, checks run, and any remaining unchecked items.

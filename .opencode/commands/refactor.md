---
description: Refactor a target without changing its business behavior.
agent: build
---

Refactor `$ARGUMENTS`.

1. Audit the target directory and read `AGENTS.md`, `docs/STYLEGUIDE.md`, and `memory/constitution.md`.
2. Identify legacy or unnecessarily complex patterns and propose the smallest safe cleanup.
3. Apply the refactor without changing business logic, public behavior, `.hlx` output, or platform support.
4. Run the existing relevant tests and checks.
5. If a check fails, attempt one focused correction, then report the remaining failure clearly.

---
description: Create a unified implementation plan from a feature specification.
agent: plan
---

Create or update the implementation plan for `$ARGUMENTS`.

1. Read `specs/[feature_name]/spec.md`, `AGENTS.md`, `docs/STYLEGUIDE.md`, and the relevant source files.
2. Create or update exactly one file: `specs/[feature_name]/plan.md`.
3. Include an Architecture section listing changed files, interfaces, and data models.
4. Include a Step-by-Step section with actionable checkbox items.
5. Label each implementation item `[Frontend]`, `[Backend]`, `[Integration]`, or `[Tests]`.
6. Do not create `tasks.md` or modify application code during this command.
7. Keep tasks atomic and mark no item complete until it has actually been implemented and verified.

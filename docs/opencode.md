# OpenCode Workflow

This repository keeps the original Antigravity workflows in `.agent/workflows/` and provides equivalent OpenCode commands in `.opencode/commands/`. The project artifacts remain the same: `memory/constitution.md`, `specs/`, and `templates/`.

## Commands

Run OpenCode from the repository root and use these slash commands:

- `/define <feature-name>` creates or updates `specs/<feature-name>/spec.md`.
- `/architect <feature-name>` creates or updates `specs/<feature-name>/plan.md`.
- `/execute <feature-name>` implements unchecked plan items.
- `/verify <feature-name>` runs logic and build verification and reports visual verification limits.
- `/refactor <target>` cleans up a target without changing behavior.

The command files are project-local and are discovered automatically by OpenCode. No global installation, npm dependency, API key, or `opencode.json` file is required.

## Differences From Antigravity

- `.agent/workflows/` is retained for Antigravity users; it is not replaced.
- `$ARGUMENTS` is supplied by OpenCode when a slash command is invoked.
- The Antigravity Integrated Browser is not assumed. `/verify` only reports a visual match when a browser or screenshot-capable tool is available in the current OpenCode session.
- The constitution's artifact rule is preserved: planning commands write Markdown artifacts rather than relying on chat history.

## Typical Flow

```text
/define reverb-decay-limit
/architect reverb-decay-limit
/execute reverb-decay-limit
/verify reverb-decay-limit
```

Keep feature names aligned with the existing directory naming convention under `specs/`.

# Pre-PR Checklist

Run this checklist before opening a pull request. The goal is to catch the recurring CodeRabbit findings before review.

## Scope And Diff

- Confirm the branch is based on the current `main`.
- Review `git status`, `git diff --stat`, and the complete diff.
- Remove generated files, local configuration, and unrelated changes.
- Confirm the PR title uses a Conventional Commit prefix and describes the main change.
- Link the relevant issue when one exists.

## Release Metadata

- Verify that `CHANGELOG.md` is unchanged in the PR diff.
- Do not add a version heading or release notes to a feature/fix PR.
- Versioning and changelog updates are handled by the version tag/release process in [`docs/RELEASING.md`](RELEASING.md), not by a PR branch.
- Apply this check explicitly to branches coming from forks, where release files must not be changed as part of the contribution.
- If a changelog change is genuinely required, document the exception in the PR and confirm that it is not generated release output.

## Go Backend

- Run `gofmt -w` on changed Go files, then verify `gofmt -l` returns no changed files.
- Run `go test ./...` from `app`.
- Check every returned error and every fallback path explicitly.
- For catalog lookups, verify that the final entry is retained after ID or alias resolution.
- Confirm defaults, parameter limits, controller metadata, and type normalization use the resolved catalog entry.
- Add a docstring to each new or modified non-trivial function. Explain assumptions and reasons, especially around Helix format and hardware behavior.
- Check that Wails context and generated bindings are used for frontend communication.

## Frontend And Build

- Run `npm run build` from `app/frontend`.
- Use generated Wails bindings rather than raw HTTP or WebSocket communication.
- Run a Wails build when backend, frontend embedding, bindings, or build configuration changed.
- Consider the Windows, macOS, and Linux behavior of every platform-sensitive change.
- Do not add a dependency without checking its binary size and cross-platform impact.

## Presets And Hardware

- Add or update regression tests for `.hlx` generation and Gemini response parsing.
- Inspect representative generated `.hlx` JSON, including parameter types, limits, paths, snapshots, and controllers.
- Verify generated presets with HX Edit or physical Helix hardware when available.
- Document the exact Helix model and firmware used for manual testing.
- If hardware was not tested, say so explicitly in the PR.
- For changes that can affect hardware recovery, document backup requirements and prefer non-destructive recovery before a full factory restore.

## Final Validation

- Run `git diff --check`.
- Re-read the PR description and include exact commands, environments, limitations, and manual test results.
- State whether AI assistance was used and confirm all generated changes were reviewed.
- Ensure no API keys, credentials, personal data, or machine-specific paths are present.

# Contributor Guide

## Scope

- Keep changes focused and avoid modifying generated frontend artifacts unless a Wails generation step requires them.
- Do not version generated frontend hash files (`*.md5`, `*.sha1`, `*.sha256`, `*.sha512`).
- Preserve catalog-backed Helix model names and parameter contracts; do not invent models or controls.

## Validation

Run these checks from the repository before opening or updating a pull request:

```bash
cd app
go test ./...
cd frontend
npm run build
```

For a macOS package build, use the Wails CLI version pinned in CI:

```bash
cd app
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails build -platform darwin/arm64 -skipbindings
```

## AI-Assisted Changes

- Disclose AI assistance in the pull request description.
- Review generated code, preserve repository conventions, and add regression tests for behavior changes.
- Treat model-generated presets as untrusted input: keep validation for catalog models, parameter ranges, signal-chain order, and DSP limits.

## Platform Coverage

- Record the operating systems, Go version, Wails version, Codex CLI version, and authentication method used for testing.
- Codex CLI changes must retain Windows-compatible executable discovery and environment-variable handling.
- Test generated presets in HX Edit or on Helix hardware before claiming tonal or export fidelity.

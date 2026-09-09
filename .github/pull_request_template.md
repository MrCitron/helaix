## Summary
<!-- Explain the problem, the motivation, and the user-visible result. Link an issue when one exists. -->

## Scope
<!-- List the important implementation changes and explicitly mention anything out of scope. -->

## Release Metadata
- [ ] `CHANGELOG.md` is unchanged for this PR
- [ ] Version and release notes are intentionally deferred to the version tag/release process
<!-- If CHANGELOG.md was changed intentionally, explain why and confirm it is not release automation output. -->

## Change Type
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Refactoring
- [ ] Documentation or configuration

## Validation
<!-- Include the exact commands run and their results. Do not only state that tests pass. -->
- [ ] `gofmt -l` returns no changed Go files
- [ ] `go test ./...`
- [ ] `npm run build` (from `app/frontend`)
- [ ] `git diff --check`
- [ ] Wails build, when backend, frontend embedding, or build configuration changed
- [ ] Manual testing, described below when applicable

### Manual Testing
<!-- Describe the flows tested, generated files inspected, and any hardware validation. -->

## Environment
- **Operating system:**
- **Go version:**
- **Node.js version:**
- **Wails version:**
- **Helix hardware:** <!-- Include none if no hardware was available. -->
- **Helix firmware:** <!-- Include not tested if unavailable. -->

## Limitations
<!-- State what was not tested, especially unsupported OSes, physical hardware, HX Edit, or tonal fidelity. -->

## Documentation And Safety
- [ ] Documentation and specs were updated when behavior changed
- [ ] New or modified non-trivial functions have useful docstrings/comments
- [ ] Hardware recovery or data-loss risks are documented when relevant
- [ ] No credentials, API keys, or sensitive data were added

## AI Assistance
<!-- State whether AI tools were used. If used, confirm generated changes were reviewed and validated. -->

## Author Checklist
- [ ] The title uses a clear Conventional Commit prefix
- [ ] The diff contains only in-scope changes
- [ ] I verified that `CHANGELOG.md` was not modified, including when working from a fork
- [ ] I performed a self-review of the complete diff
- [ ] I followed `docs/PRE_PR_CHECKLIST.md`

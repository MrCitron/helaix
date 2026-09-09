# Releasing HelAIx

Releases are created through the `Release` GitHub Actions workflow. Do not update `CHANGELOG.md` or package versions in a feature or fix PR.

## Procedure

1. Merge the desired changes into `main`.
2. Open **Actions > Release > Run workflow** on the `main` branch.
3. Enter the next semantic version, with or without the `v` prefix, for example `v0.13.0`.
4. Review the generated release commit and the created Git tag.
5. Review the GitHub release and its Windows and macOS artifacts.

The workflow performs these operations in order:

- validates that the version is a new semantic version greater than the current frontend package version;
- generates a `CHANGELOG.md` entry from non-merge Conventional Commit subjects since the latest reachable version tag;
- updates `app/frontend/package.json` and both root version fields in `package-lock.json`;
- regenerates `app/frontend/package.json.md5`;
- commits the release metadata to `main`;
- creates and pushes an annotated `vMAJOR.MINOR.PATCH` tag;
- builds the tagged source for Windows and macOS;
- creates the GitHub release using the generated changelog entry as its body.

The workflow requires repository write permission and is intentionally restricted to the default branch. A fork can contribute through pull requests, but cannot create a release for the upstream repository.

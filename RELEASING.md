# Releasing

pixelize follows [Semantic Versioning](https://semver.org/) and keeps a
human-readable [CHANGELOG.md](CHANGELOG.md) in the
[Keep a Changelog](https://keepachangelog.com/) format.

## Version numbers

Given a version `MAJOR.MINOR.PATCH`:

- **MAJOR** — incompatible CLI/behavior changes.
- **MINOR** — new, backward-compatible functionality (new flag, new preset
  mechanism, etc.).
- **PATCH** — backward-compatible bug fixes.

While the project is pre-1.0 (`0.x`), the CLI surface may still change between
minor versions.

## Cutting a release

1. Move the entries under `## [Unreleased]` in `CHANGELOG.md` into a new
   `## [X.Y.Z] - YYYY-MM-DD` section, and add the matching link references at
   the bottom of the file.
2. Commit the changelog: `git commit -am "release: vX.Y.Z"`.
3. Tag it (annotated):

   ```bash
   make release VERSION=vX.Y.Z
   # or manually:
   git tag -a vX.Y.Z -m "vX.Y.Z"
   ```

4. Push the commit and tag: `git push origin main --follow-tags`.

The build injects the version from `git describe --tags` via `-ldflags`, so a
tagged commit produces a clean `vX.Y.Z` and later commits produce
`vX.Y.Z-<n>-g<sha>`.

# Packaging graduation: GoReleaser

**Status:** Tracked for adoption · **Trigger:** first package-manager publication, or when the CI release bash outgrows itself

## What

Replace the hand-rolled cross-compile/release bash in `.github/workflows/hooks.yml` with [GoReleaser](https://goreleaser.com/) — the de-facto standard for Go release packaging.

## Why

The current release job is ~20 lines of bash looping GOOS/GOARCH pairs, producing per-platform binaries + `SHA256SUMS`, uploaded by an action. It works, is dependency-free, and is tested by use. GoReleaser earns its place when we need what it does declaratively and the bash does not:

- **Package-manager manifests** — scoop (Windows), winget, Homebrew taps, generated and published per release. This is the planned install-story graduation from `plans/claude-port-plan.md` §3a′ ("publish to scoop/winget/brew so the per-OS install string is a one-liner and PATH is handled for us").
- Declarative build matrix (`.goreleaser.yaml`) instead of a bash loop; archives, checksums, changelog, and release upload in one tool.
- Signing/attestation when we want it.

GoReleaser is a **CI-only dependency** — nothing lands on user machines, so it does not fight the install problem (§3a′).

## When

Adopt in one PR at the point either condition holds:

1. We publish the hook binaries to a package manager (scoop first — Windows is the primary dev box), or
2. The bash release job needs its third nontrivial edit.

Until then the bash stays: it is small, visible, and has no supply-chain surface beyond the upload action.

## Sketch of the adoption PR

- `.goreleaser.yaml`: builds for each `tools/cmd/*` (windows/darwin/linux × amd64/arm64), binary naming `<name>_{goos}_{goarch}`, checksum file `SHA256SUMS` (same contract `sc-doctor -fix` already consumes), release on `{plugin}--v*` tags.
- `hooks.yml` release job becomes: checkout → setup-go → `goreleaser release --clean`.
- Later increment: scoop bucket + winget manifest publication from the same config.
- No change needed in `sc-doctor` — the asset-name and `SHA256SUMS` contract is preserved by construction.

## Origin

Raised in review of PR #11 (`hooks.yml` — "do we need a build system?"). Conclusion there: `go build` is the build system; the bash is release packaging; GoReleaser is the graduation. Tracked here so it isn't lost.

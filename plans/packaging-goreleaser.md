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

---

## Trigger state, measured 2026-08-15

**2 of 3. The trigger has NOT fired, and the bash stays.**

The two nontrivial edits to the release job since this was written:

| Commit | Date | Change |
|---|---|---|
| `4baf282` | 2026-07-18 | Added *"Resolve the plugin from the tag"*. The job had been hardcoded to prosthetic-conscience through the workflow-level `working-directory`, so a FEOV tag would have built PC's binaries and published them under FEOV's release — silently, because the build succeeds. |
| `6eebad6` | 2026-07-29 | Replaced the third-party upload action with `gh release create` (zizmor backlog). |

### The trigger is not well formed — read the count above with that in mind

**"2 of 3" is softer than it looks, and the softness is the point.** The counting criterion
had to be invented at measurement time (a grep for `GOOS|SHA256SUMS|gh release|dist/`), because
"nontrivial" was never defined. Is `6eebad6` — a one-line publish-mechanism swap inside a
security sweep — nontrivial? Counted here as yes. A reasonable reader says no and it is 1 of 3;
another counts `4baf282`'s `working-directory` restructure separately and it is 3 of 3, fired.

A trigger whose value depends on who greps is prose, not a field. That is the defect this repo
has been removing everywhere else, sitting in the document that governs a decision.

**And condition 2 proxies the wrong signal.** Edit COUNT does not measure danger. `4baf282` was
not an edit; it was a defect in the publish path that would have shipped one plugin's binaries
under another plugin's release, silently. By the question condition 2 is reaching for — has this
bash shown it is hard to change safely? — the answer is already yes, once, in the worst
available way.

**Condition 1 cannot fire from the inside.** "Adopt when we publish to a package manager" is
blocked by the thing adoption would fix: hand-rolling scoop/winget manifests is what stops us
publishing. A trigger that defers itself explains why this sat for a month.

### The claim that should have been challenged first

> It works, is dependency-free, and is **tested by use**.

It is not tested. The job emits a contract — assets named `<name>_<goos>_<arch>[.exe]` plus
`SHA256SUMS` — that is parsed by a DIFFERENT program on a USER's machine: `sc-doctor -fix`
(`verifySHA256`, `strings.TrimPrefix(fields[1], "*") == asset`). Two programs, one string
contract, and no test joining them. `doctor`'s own tests check its parser against synthetic
sums; nothing checks that what the release job emits is what `doctor` looks for.

"Tested by use" means tested in production, once per tag, with no way to fail safely — and this
is a fact composed into a filename at one end and recovered by string match at the other.

### So the real gap is not the packaging tool

Swapping bash for GoReleaser changes WHAT is untested, not WHETHER. The release path still
cannot be exercised before a tag, and the silent-mis-publication class survives the swap.

**A release dry-run gate is worth more than either option**: on every pull request, run the
cross-compile, the naming and the checksum into a temp directory, and assert the emitted asset
names are the ones `sc-doctor` actually looks for. No new dependency, it closes the class, and
it is what would make a later GoReleaser swap SAFE — today that swap would be an untested
rewrite of an untested job, verifiable only by tagging a release and watching.

### One argument this document did not anticipate

`6eebad6` moved in the opposite direction from adoption: it **removed** a third-party dependency from the publish path. Adopting GoReleaser adds one back — into the one job that has write access to releases and that no local check can exercise.

That does not overturn the case (the package-manager manifests are still the thing the bash cannot do). It sharpens condition 1: **adopt when we actually publish to a package manager**, where GoReleaser earns its supply-chain cost by replacing manifest-generation nobody wants to hand-roll. Condition 2 — the third nontrivial edit — is the weaker trigger, and on this evidence it should be read as *"the bash has become hard to change safely"*, not as a counter to reach.

### Why this document is now in the tree

It spent a month as an unmerged tracking pull request, headed *"Tracked here so it isn't lost."* It was lost: an issue search found nothing, an in-tree grep found nothing, and it was recovered only because a human remembered the pull request number.

**A tracking pull request does not track.** The decision now lives in the tree, where a grep finds it, and the trigger's state is recorded here as a measurement with its evidence rather than as a thing someone will remember to re-check.

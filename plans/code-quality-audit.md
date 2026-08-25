# Code-quality audit — `plugins/prosthetic-conscience/tools`

> 42 non-test files, 53 test files. Audited 2026-08-25 against Go practice and this repo's own
> stated principles. Every claim below carries the command that produced it, so a reader can
> re-run rather than trust — the standard §III applies to censuses applies here too.
>
> **Ranked by whether the class has already produced a defect**, not by how much code it touches.
> Two of the six have; one has twice.

---

## 1. The hook preamble is copy-pasted eleven times, and it carries the binary name TWICE

```bash
grep -rn "func run(" --include=*.go internal/ | grep -v _test    # 13
grep -rc "showVersion\|buildid.Line" --include=*.go internal/*/main.go | grep -v ":0"
```

Eleven packages open with the same ten lines, differing only in a string:

```go
fs := flag.NewFlagSet("sc-postcompact-observe", flag.ContinueOnError)
fs.SetOutput(stderr)
showVersion := fs.Bool("version", false, "print version and exit")
if err := fs.Parse(args); err != nil {
    return 0
}
if *showVersion {
    fmt.Fprintln(stdout, buildid.Line("sc-postcompact-observe"))
    return 0
}
```

**This class has already produced a shipped defect.** The name appears twice per package, and when
`checkpointseal` began serving three shims from one `run()`, the literal could not be right for all
three: `sc-precompact`, `sc-sessionend` and `sc-subagentstop` all answered `-version` with
`sc-checkpoint-seal`, the name `#201 step 3` retired. `sc-doctor` lists hook binaries by name and
prints that line, so its table carried three rows with one name — and "which of the three is stale"
is the only question the line exists to answer. Fixed in `edc4822`, by hand, in one package.

The boilerplate is not the cosmetic problem; **the duplicated name is a defect generator**, and it
will generate the same defect again the next time a binary is renamed or split.

**Extraction.** A `hookmain.Run(name string, fn func(...) int)` that owns the flag set, the version
line and the exit convention, taking the name ONCE. The eleven call sites lose ten lines each and
gain a single place where "what is this binary called" is answered.

**Cost, stated honestly:** it puts a layer between `main()` and the logic, and this suite's hooks are
deliberately shallow. The justification is not line count — it is that the name stops being repeated.

**DONE — `65a4679`.** `internal/hookmain`, 100% covered. The name is supplied once and **as a
function**, because `checkpointseal` cannot know its name until it has parsed `-event`; an API
serving only the easy nine would have left the tenth writing the code that broke. All eleven
migrated.

---

## 2. State-file IO is implemented twice, and the two have ALREADY diverged

```bash
grep -rln "os.CreateTemp" --include=*.go . | grep -v _test
#   internal/freshness/gauge.go     writeState / readState
#   internal/stopnudge/stopnudge.go save       / load
```

Both write a small JSON state file with temp-file-plus-rename and read it back. Written hours apart,
by the same author, for the same reason.

**They are no longer the same.** `freshness.readState` retries three times for the transient Windows
sharing violation that CI measured (21 of 800 reads under concurrent writers) and distinguishes
ABSENT from UNREADABLE. `stopnudge.load` does neither: it has no retry, so a transient failure
silently suppresses a nudge rather than being absorbed.

That difference is not a decision anybody made. It exists because CI exercised one copy and not the
other, and the fix went where the red was.

**Extraction.** `internal/statefile` with `Read[T]`/`Write[T]` over a path, owning the temp+rename,
the retry, and the absent/unreadable tri-state. Both callers keep their own struct and their own
policy about what a failed read MEANS — `freshness` refuses to stamp, `stopnudge` fails closed —
because those genuinely differ.

**This is the highest-value item in the audit.** It is the only one where the duplication has already
cost correctness rather than tidiness.

---

## 3. Three JSONL appenders

```bash
grep -rln "O_APPEND" --include=*.go . | grep -v _test
#   internal/hooklog/hooklog.go
#   internal/checkpointseal/sealrow.go
#   internal/postcompactobserve/main.go
```

Three implementations of "marshal a row, append a line, never fail the hook". They have not
diverged in behaviour yet, but they have diverged in *care*: `hooklog` is the oldest and most
defensive; `sealrow` reports every failure to stderr; `postcompactobserve` reports some.

**Extraction is worth less here than in §2** — appending a line is genuinely simple, and the shared
helper would be four lines of body plus a signature. The argument for doing it is consistency of the
failure posture, not reuse. **Recommend: extract only if §2's `statefile` lands**, since the same
package is the natural home and the marginal cost is then near zero.

---

## 4. `exeSuffix` defined twice

```bash
grep -rn "func exeSuffix" --include=*.go .
#   internal/doctor/main.go:45
#   internal/hookinvocation/invocation_test.go:133
```

Trivial, and the second one is mine. The shipped hook command already encodes the same knowledge a
third time (`if [ -x "$B" ] || [ -x "$B.exe" ]`). Three statements of one platform fact.

**Recommend:** export the `doctor` one, or move it beside `buildid`. Low value, low risk, do it while
touching the area for §1.

**DONE — `57750e8`.** `buildid.ExeName`. There was a **fourth** statement the census missed: an
inline `if runtime.GOOS == "windows"` in `buildid`'s own test. `AssetName` deliberately does not
call it — that is a CROSS-compile name whose extension must follow the argument `goos`, never the
host — and a fifth site was hand-assembling `<name>_<os>_<arch><ext>` in the message telling a
human which file to place, duplicating the function whose own comment calls it "THE CONTRACT …
in one function". That one now calls it.

---

## 5. Error wrapping is inconsistent

```bash
grep -rn "fmt.Errorf" --include=*.go internal/ | grep -v _test | grep -c "%w"    # 4
grep -rn "fmt.Errorf" --include=*.go internal/ | grep -v _test | grep -vc "%w"   # 7
```

Seven of eleven `fmt.Errorf` calls flatten their cause. In a codebase whose hooks are best-effort and
mostly discard errors, this matters less than it would elsewhere — but where an error IS returned it
is usually to a caller deciding whether to continue, and `errors.Is` cannot see through a flattened
one. `freshness.readState`'s `errors.Is(err, os.ErrNotExist)` is exactly the pattern that breaks.

**CORRECTION — the number above is wrong, and the way it is wrong is the point.** `7` is a grep
count of `fmt.Errorf` lines lacking `%w`. Re-reading each: **five have no cause in scope at all.**
They are original errors (`gh not on PATH`, `checksum mismatch for %s`, the two lock refusals,
`cannot pin a release tag`) and flatten nothing. A census that counts a SHAPE and reports it as a
DEFECT overstates by however many instances are innocent — and nobody re-derives a number once it
is written down.

Two were real, and are fixed in `57750e8`:
- `fetchRelease` discarded `gh`'s exit status, keeping only its output text — so a caller could
  not distinguish "gh refused" from "gh was killed".
- `readRearm`'s "present but unparsable" covers TWO failures and only one carries a cause: invalid
  JSON, and valid JSON that is not this record. Wrapping unconditionally would have rendered the
  second as `%!w(<nil>)` — a decoding error reported where there was none. The second case now has
  its own sentence.

**Recommend:** `%w` where a cause EXISTS and the error is returned rather than logged. Not
mechanical: at one of the two real sites the cause is sometimes nil, and the naive sweep would
have printed a lie.

---

## 6. `stopnudge` deviates from the house shape — and it is mine

`run()` in every other package takes `args []string` and owns its flag parsing; `stopnudge.run` does
not, and its `-version` handling lives in `Main()` instead, parsed by hand:

```go
if len(os.Args) > 1 && os.Args[1] == "-version" {
```

That is a worse parser than the one it skipped — it only matches the flag in first position — and it
means the binary answers `--version` differently from its ten siblings. I wrote it that way because
`run` had already grown a thresholds parameter and adding `args` felt like clutter. That is not a
reason, it is a preference expressed as one.

**Recommend:** conform, as part of §1.

**DONE — `65a4679`, and it was hiding a live defect.** Conforming the flag parsing surfaced a
second deviation in the same function: `Main` passed `os.Getwd()` where its ten siblings pass
`CLAUDE_PROJECT_DIR`. `hookenv.ProjectDir` prefers its first argument and a working directory is
never empty, so the payload-`cwd` fallback beneath it was **unreachable**, and the refusal
`hookenv.Explain` exists to perform — *"doing nothing rather than guessing from the working
directory"*, in a message this package itself prints — could not fire.

Not latent. Driven from any directory that is not the project root, `sc-stop` read a
`CHECKPOINT.md` belonging to somebody else and wrote a `freshness.json` **baseline** beside it —
stamping a Phase 1 reading against the wrong note, into the corpus whose only purpose is to set
thresholds. Observed directly: the bait tree gained `.claude/checkpoints/freshness.json`.

The guard is `internal/hookinvocation/projectroot_test.go`: every binary that calls `Explain`, run
as a **process**, against a working directory baited to look exactly like a project root. It
asserts a filesystem observation rather than searching stderr for the refusal's wording — a
substring that stops matching after a reword reads exactly like a pass ([[facts-are-fields]]
clause 3). Confirmed RED against the reverted `Main`, green after.

**This is the audit's own lesson.** §6 was ranked last, described as style, and flagged only
because it was mine. The defect was in the same six lines.

---

## NOT a finding: the nine `hookInput` structs should NOT be unified

```bash
grep -rln "type hookInput struct" --include=*.go .   # 9
```

This is the obvious extraction and it would be **wrong**, so the refusal is recorded here rather than
left for someone to attempt.

Each struct documents what its event was MEASURED to carry — `hook-surface-spike.md` §9–§13 — and the
differences are the point:

- `SessionStart` carries no `transcript_path`, which is why the restore digest cannot measure turns
  or growth at all (§9e correction 3).
- `PreCompact` and `SessionEnd` carry no `background_tasks`, so a handle count there is manufactured
  (§12).
- `Stop` carries no `agent_*` fields, which is why the debounce is keyed on `session_id` alone.

A shared payload type would present every field to every hook, and the compiler would stop
distinguishing "this event does not carry that" from "I forgot to read it". Three of this design's
defects were exactly that confusion, caught by measurement rather than by types. **Nine structs that
disagree are the safety property**, not the duplication.

The same argument does NOT protect §2: an atomic write is one idea, and the two copies are not
disagreeing about anything except how carefully they were written.

---

## Order of work

**Status: §1, §2, §4, §5 and §6 are done (`7a76cf0`, `65a4679`, `57750e8`). §3 remains.**

1. **§2 `statefile`** — the only class that has already cost correctness, and it fixes `stopnudge`'s
   missing retry as a side effect rather than as a separate patch.
2. **§1 `hookmain`** — the class that has already shipped a defect, and the one most likely to ship
   another.
3. **§6, §4, §5** — mechanical, do them while in the area.
4. **§3** — only if §2 lands, and only into the same package.

Tests must **improve** across all of it: every extraction takes the strictest behaviour of its
copies, not the average. §2 in particular means `stopnudge` inherits the retry and the
absent/unreadable distinction it does not currently have, and that inheritance needs its own test
rather than being assumed from the shared implementation.

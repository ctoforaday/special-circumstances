# The tool is the contract — retiring the file paths before the next run

Written 2026-07-19. Goal: the next run (same prompt, haiku, as a controlled repeat) is the
one where a seat never reads or writes a storage path, and the hand-written markdown stops
being authoritative.

## I. Summary & Goals

A seat could already WRITE every act through `feov-record` while still READING the run by
opening markdown at paths it learned from a prompt. That asymmetry is the whole problem:
it is how a seat comes to trust a hand-written artifact over the event log, and how the
two came to disagree by 9 open / 9 closed against a hand-written 3 open / 15 closed.

**Measured:** of ~34,300 bytes of seat-prompt text in `debate.js`, ~9,500 (28%) are
file-path instructions and ~5,500 (16%) are procedural ceremony the tool could enforce.
**~44% of the prompt corpus is tool-absorbable.** The rest is genuine judgement and stays.

## II. What already shipped (0.12.0 – 0.14.0)

- `show --view ledger|archive|debate|changelog|citation-ledger|lines-of-inquiry` — reads
  through the tool, byte-identical to the projection files.
- `show --view board` — the board as **structured JSON**, the merge seat's default. Gaps
  with grades, closures with their anchors as FIELDS, observations with `disposed` stated,
  counts, and replay anomalies.
- Anomalies surfaced. The 2026-07-18 records produce 12, including 8 dropped judicial
  closures.
- One vocabulary, enforced by a test that walks the real command tree both ways.

## III. The correction that changes the plan

**The 34,086-vs-7,527 archive parity gap is contaminated and must be re-measured.**

A parity analysis concluded the board misclassifies six gaps because `BoardState` honours
only red's `close` and ignores bench closures. That is **false** — a cross-seat test proves
bench closures close gaps. Re-rendering the real run with today's binary still gives 9
open, and the reason is different:

> **Zero of 271 events carry `ts`.** The run predates the timestamp schema. Ordering falls
> back to `(SeatID, Seq)`; `"judge-r2"` sorts before `"red-merge-r1"`; every judge opinion
> replays before the mint it references and is dropped as an unknown gap.

That is the defect the timestamp work already fixed, for runs recorded after it. So part of
the measured byte gap is legacy-ordering damage, not missing verbs. **Do not build verbs
against this measurement.** Re-measure on the next run, which will carry timestamps.

## IV. What must land before the flip

### A. Views that retire a path (no new verbs; extend `views` in `seat/verbs.go`)

| View | Retires | Note |
|---|---|---|
| `report` | `${runDir}/blue/report.md` in **5 prompts** | highest value |
| `lens-passes --round N` | the `cat …lens-*.md` ceremony (513 B) and its stray-file hazard | |
| `gap-patterns [--class C]` | `${runDir}/inputs/red-gap-patterns.md` | |
| `law` | `${runDir}/inputs/law/` | |
| `frontier`, `candidates [--lane N]` | blue's draft paths | needs `hypothesis`/`draft` verbs |
| `debate --round N --section RED\|BLUE\|LEAD` | "read the latest ### RED section" | flags, not a verb |
| `archive --id <gap>` | the DEMANDED-READS targeted read | also enables B6 below |

**`report.md` is the one genuine file** — it is the deliverable, not a projection. It stays
a file. Make it the ONLY path any prompt names.

### B. Ceremony the tool should enforce instead of asking for

Ordered by prompt text retired. Each becomes an integration test.

1. **Round record is derived, not attested** — delete `round_record_appended`. Query the
   record for a `position` + `revision` for round N. (~430 B, one envelope field, two
   script throws.)
2. **Board telemetry is computed** — the merge hand-computes ~1,100 B of schema that the
   script then independently recomputes. Two computations of one fact.
3. **Near-match rule** — `mint` returns near-matches and requires `--supersedes` or an
   explicit `--not-a-reopen`. Currently pure good faith.
4. **Pattern duty is a flag** — `manifest-row --pattern-checked`. This is the clause with
   MEASURED non-compliance: run 5's lanes read the pattern file and committed the warned
   defects anyway. A required flag is the only thing that fixes reading-is-not-binding.
5. **`register` is a precondition** — every verb refuses before it, rather than the prompt
   asking for it as "FIRST ACTION".
6. **Demanded reads become impossible to skip** — `opinion`/`close` on a gap with lineage
   require that the ancestor's record was read through the tool this session.
7. **`verdict` refuses while duties are outstanding** — an unanswered dispute, or a missing
   spot-check when the archive entered non-empty.
8. **`found_by` is a foreign key** — `mint --from-finding L5-F3` derives the lens from the
   finding event instead of trusting a self-report.

### C. Dead text, deletable today with no new code

- `frictionClause`'s "ALSO append each entry to `${runDir}/friction.md`" (261 B) — the
  `friction` verb already does exactly this. The dual-write is a pre-tool workaround.
- "You MUST NOT write to `${runDir}/debate.md`" — already enforced structurally: the lens
  role has no `position` verb.
- The lineage paragraph at 596 — enforced twice already, in the tool and in the script.
- `speedClause`'s sanctioned `bash grep/ls` carve-out (~380 B) — it exists only to serve
  run-directory reads. Once no seat reads paths, it has no referent.

## V. Verification plan

1. `go test -race -count=1 ./...` in `plugins/frank-exchange-of-views/tools`.
2. **`TestSeatNeverNeedsAPath`** — the migration's own validation loop. Drive a full round
   for each of the six seat kinds using ONLY `feov-record` invocations; assert no
   filesystem path appears outside `--run`. Then assert the `debate.js` prompt strings
   contain no `${runDir}/` outside the `--run` argument. **Write it first and let it fail** —
   it is the gate that keeps the migration from regressing.
3. Re-measure record parity on the FIRST post-timestamp run, not on 2026-07-18.
4. The smoke: rerun the same topic in haiku and compare against the captured baseline in
   `research/2026-07-18_gray-area-telemetry/`. The controlled repeat is the point — same
   prompt, same model tier, so a difference is attributable.

## VI. Honest status

Shipped: the read surface, the structured board, the vocabulary lock, cross-seat tests.
**Not shipped: the path retirement itself.** No prompt has been rewritten yet, `report`,
`lens-passes`, `gap-patterns` and `law` views do not exist, and none of the §IV.B gates are
built. The next run is NOT yet path-free, and saying otherwise would be the kind of claim
this document exists to make checkable.

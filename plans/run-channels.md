# Run channels carry what they are for — the report is research, the log is typed

> STATUS 2026-09-06: not started — design proposed for review. Supersedes the
> `report-voice-separation.md` draft (PR #755), which was **corrected by measurement**: it
> proposed a mechanism (#737, an addressable friction id the report could cite) to preserve
> content that turned out to be 65% ceremony. **#737 is withdrawn.** Nothing built.

A run writes to two human-facing channels. The **report** is addressed to a reader of the
subject. The **operator channel** is addressed to whoever can retool the seat. Today both carry
things they are not for: the report narrates its own construction, and the operator channel is
half compliance ritual. This plan makes each carry what it is for, and names each for what it
carries.

The previous draft of this plan got the diagnosis right and the remedy wrong. It assumed the
operational content leaking into the report was worth preserving somewhere, and designed a
citable id to preserve it. An audit of the actual content says otherwise, and the remedy is
mostly **deletion and naming**, not new mechanism.

## The measurement that corrected this plan

All 35 friction entries from `research/2026-09-02_quadratic-formula` — **142,891 characters** —
classified, plus the two seats that wrote the largest entries resumed on their own conversation
IDs and asked their intent.

| bucket | chars | % |
|---|---|---|
| **A** actionable defect (reproducer + proposed fix) | 65,994 | 46.2% |
| **B** surface survey ("verbs I read and did not use") | 68,169 | **47.7%** |
| **C** epistemic limit (belongs in the report) | 2,576 | 1.8% |
| **D** narration / complaint | 6,152 | 4.3% |

- The explicitly-headed survey sections alone are **64,960 chars (45.5%)**, and a mechanical
  search of all of them for any fix proposal (`WHAT THE WORK WANTED`, `WHAT I WOULD HAVE DONE`)
  returns **zero hits**.
- A further **~13%** is duplicate re-reports of a defect already on the record.
- **~65% is deletable with no information loss.** The waste is not complaint (4.3%) — it is a
  mandated ritual.
- `motion petition file` was invoked **zero times in the entire run**, yet 23 of 35 entries spend
  **9,584 chars (6.7%)** explaining why they did not file one; 12 recite the same stock formula.
- Bucket B is not merely useless, it is **unreliable**: `red-lens-r4-L2`'s survey states it
  *rejected* `reproduce` and `finding`, while `events(seat_id, round, type)` for that seat and
  round records `register, reproduce, verify, finding, anchor, friction`. The one place the
  attestation is mechanically checkable, it fails.

**Root cause, verified** — `debate.js:94`, applied to **8 seat prompts**:

> `AND THE REASON OWES THE SURVEY: name the verbs you read in the tree and did NOT use… and a
> seat that cannot name a single rejected option did not weigh any.`

It mandates the survey, **pre-accuses** the seat (which bought the performative length), and
fuses a deliberation log into a defect channel. **A gate DOES pin this text** — an assertion at
`tests/simulator/debate.test.mjs:258-259` across five seat classes, plus ten prompt goldens; an
earlier draft of this plan claimed otherwise on a wrong-scoped grep. The pin, its rationale, and
the answer to it are in PR-1's census below.

**The seats, asked directly:**
- red-lens-r2-L5: *"Duty, not judgement… I wrote it at that length to be visibly compliant with a
  prompt that pre-accuses the seat. That is writing to prove I had weighed, not writing to inform,
  and I knew the difference while doing it."* · *"The record already knows which verbs my seat ran;
  a projection can derive used/unused for free. The prompt made me hand-write a fact the record
  holds — the same defect the suite's own rule names, one level up."*
- judge-r3: *"The page-by-page inventory is a receipt, not information… close to half the entry."*
  · *"I filed a bug report as prose into a channel nothing forces anyone to read — the exact
  failure this repository's facts-are-fields rule names."* · *"Ask me only where non-use was a
  judgement rather than an absence of occasion."*

Both, independently, named the same destination for the valuable half — a tracker — and the same
severity signal — **recurrence**.

## Why the remedy is naming, not mechanism

The consumer is a human (or an agent acting for one) reading the channel after a run and deciding
what to do — and **the entries are not always actionable or advisable**. That makes triage
judgement irreducible, which rules out both an auto-filer and a ticket-shaped schema: a
`component | wanted | repro` form would push a seat to shape every entry as a filable ticket, which
is a claim the seat is not entitled to make.

What the triager actually needs is the one thing the writer knows and a reader must currently
infer from prose: **what kind of thing is this, and who said it.** That is a field.

And the channel is misnamed. `friction` names the *least actionable* of the things it carries, so
the container is named after one of its own members. `Friction{kind: friction}` cannot be written
coherently; `Log{type: friction}` can. The rename is what makes the taxonomy legible — it is not
cosmetic.

---

## I. Summary & Goals

**Objective.** (1) Stop the operator channel demanding ceremony, and give its entries a type and a
source so triage is a filter rather than a reading. (2) Keep the report to research prose. (3) Make
the report's biggest caveat — is a source unread, or merely unreachable from here — a field.

**Success criteria**, measured by re-running the two censuses in §V against a *new* run:

| Measure | Baseline (2026-09-02) | Target |
|---|---|---|
| operator-channel chars | 142,891 | **≤ 75,000** — the arithmetic, not an aspiration: PR-1 removes bucket B (68,169); A+C+D = 74,722 remain |
| headed "verbs I read and did not use" sections | 64,960 (45.5%) | **0** |
| entries with an explicit type + source | 0 of 35 | **all** |
| report: process-voice tells (`this run/round/report`, `the debate`) | 161 | ≤ 5, provenance footer only |
| report: inline lane-attribution tags | 24 | 0 |
| report: inlined access limits | 13 | 0 — re-voiced as limits on the conclusion |

**Two things are deliberately NOT targeted, and saying so is the point.** Bucket D (narration,
4.3%) and the ~13% duplicate re-reports have no change proposed against them: D is small enough
that a mechanism would cost more than it saves, and the duplicates follow from seats being unable
to read the channel — which §II records as a deliberate design property, not a defect. §V measures
both so a later run can show whether that judgement held; neither carries a target it could be
scored against, because nothing here is trying to move them.

**Non-goals.** No auto-filing to a tracker (credentials in a research container, tracker volume,
and seat-error quality all argue against it, and triage judgement is irreducible). No addressable
id for cross-channel citation (#737, withdrawn). No `component/wanted/repro` schema. No migration, no archaeology, no backwards compatibility: archived records are not readable by a post-rename binary and are not rewritten. No cross-run
recurrence counter — measure again first.

## II. Technical Context

- **Engine:** `skills/research-protocol/scripts/debate.js` (JavaScript, goja under fuzz).
- **Tools:** Go; protobuf-defined record (`internal/record/recordpb/record.proto`) over append-only
  SQLite; projections in `internal/report`.
- **Verified constraints:**
  - Nothing imports stdlib `log` (0 files), so a `Log` type introduces no collision; `frictionLog`
    is one function in one file.
  - `FRICTION_KIND_TOOL_ERROR` has **no functional readers** (two comments only), so it can
    collapse. `FRICTION_KIND_ESTOPPEL` **is** read (`estoppel.go:173`) and must survive.
  - The friction-parity audit joins on **seat** (`capture.go:283`, `wroteToRecord[fr.SeatID]` ←
    `SeatOfAgent`), not on text or id.
  - Seats **cannot read** the channel — it is the operator's read by design (`friction.go`,
    *"yours, not theirs"*). This is why the same defect is re-filed 4-5×; it is left alone
    deliberately, because letting seats read it pulls the channel back into the debate.
  - Agent-facing prose is gated: `promptverbs` (no spelled flag/verb), `archaeology` (no
    obituary), `debate.test.mjs`, `naming`, `rulesweep` trailers. Prompt edits clear all five.

## III. Proposed Changes

### PR-1 — Delete the survey mandate `[MODIFY]`

`debate.js:94` `frictionClause`. Remove *"THE REASON OWES THE SURVEY…"* and the pre-accusing
sentence. Keep the duty to close the channel. Add one narrow ask, which is the only part both
seats defended: **where you declined an act as a judgement rather than for lack of occasion, one
sentence.** (judge-r3's case: *"an absent `outcome` stamp is ambiguous between 'the bench declined
to stamp' and 'the bench never reached it'"* — a real plausible-zero about a consequential act.)

Also retire *"what you reached for and found"* from the `--none` help (`seat/verbs.go:114`), which
re-invites the inventory.

**A gate DOES pin this text, and the earlier draft of this plan said otherwise.** That claim came
from a grep scoped to `skills/ tools/` while the pin lives under `tests/`; the zero was a wrong
search reported as an absence — the plausible-zero this repository is written against, committed
in a plan about it. Recorded here because the correction is the evidence.

**Consumer census — `frictionClause` (prompt contract), run 2026-09-06 from the repo root:**
```
$ P=plugins/frank-exchange-of-views
$ grep -rn "frictionClause" $P/skills $P/tools --include=*.js --include=*.go
$ grep -rl "OWES THE SURVEY" $P
```
| Consumer | Changes? |
|---|---|
| `$P/skills/research-protocol/scripts/debate.js:94` (definition) | **YES** — the edit |
| `debate.js:688,738,856,882,1099,1131,1241,1274` (8 call sites) | no — call shape unchanged |
| `$P/tests/simulator/debate.test.mjs:258-259` | **YES** — asserts `/THE REASON OWES THE SURVEY/ && /did NOT use/` across five seat classes (`:257` pins the ENVELOPE field and is a PR-2 consumer, not PR-1); the assertion and its rationale comment (`:252-256`) are rewritten to pin the new narrow ask |
| `$P/tests/simulator/testdata/prompt-*.golden` (**10 files**) | **YES** — each carries the string; refreshed |
| `$P/tools/internal/cli/seat/verbs.go:114` (`--none` help) | **YES** — retire "what you reached for and found" |
| `$P/tools/integration/surface/promptverbs_test.go:259`, `$P/tools/releasegate/fuzz/promptverbs_test.go:259` | no — renders clauses for the flag scan; pins no wording |

**The rationale the pin defends, answered rather than deleted.** `debate.test.mjs:254-256` argues
the survey "is the instrument the traversal is measured with, and no help page can ask a seat about
the pages it chose not to act on." The second half is true — the record's event types are all
*acts that write events* (`register, finding, verify, …`); there is no event for reading a help
page, so traversal genuinely is not otherwise recorded. But the instrument is **unfalsifiable
self-attestation**, and in the one case where it can be checked mechanically it is **false**:
`red-lens-r4-L2`'s survey states it rejected `reproduce` and `finding` while that seat's own events
record both. A check that cannot fail, and does not hold where it can be tested, is what the bench
docked blue for at R3-5 in this same run. Deleting it forfeits the *appearance* of a measurement,
not a measurement.

**Stated loss:** after PR-1 there is no signal at all about surface traversal. A real instrument is
possible — help invocations are observable at the hook/trajectory layer, where tool calls are seen
— but that is new mechanism this plan deliberately does not build. It is named here so the choice
is on the record rather than discovered later.

This PR alone is expected to remove ~48% of the channel's volume.

### PR-2 — `friction` → `log`, with `source` and `type` `[MODIFY]`

Rename the channel to what it is, and split the one overloaded enum into the two axes it is
actually carrying.

- `[MODIFY]` `record.proto`: `message Friction` → `message Log`; `message FrictionNone` →
  `message LogNone`; `enum FrictionKind` → **two** enums:
  - `LogSource ∈ { SEAT, TOOL }` — who recorded it. Today this is inferable only by knowing which
    enum values are tool-emitted; making it explicit means a new tool-emitted type never breaks a
    "seat reports only" filter.
  - `LogType ∈ { DEFECT, REQUEST, FRICTION, ESTOPPEL }` — what is being claimed.
    - `DEFECT` — something is broken. Absorbs today's `TOOL_ERROR` as `(TOOL, DEFECT)`; verified
      to have no functional readers.
    - `REQUEST` — a capability that does not exist. This is the category that made "bug" wrong as
      a name: *"`reproduce` needs a `does_not_run` outcome"* is a request, and it is among the
      best content in the corpus.
    - `FRICTION` — impeded the work; noted, **not necessarily actionable or advisable**. The
      honest home for content that today has to pose as a defect report.
    - `ESTOPPEL` — the tool refused a mint; `(TOOL, ESTOPPEL)`. Retained: it is read at
      `estoppel.go:173`.
  - `UNSPECIFIED` stops carrying meaning. Today `UNSPECIFIED = 0` doubles as "unset" *and* "a
    seat's capability gap" — a default value with semantics, which is the plausible-zero shape.
- `[MODIFY]` the seat's write to take the type (source is set by the write path, not asked of the
  seat); the operator read and `## Friction` → `## Log` projection to render and filter on both.

**Three renames this PR must decide, because each has mechanical readers — decided here:**
the **CLI command** (`cli/friction.go:28` `Use: "friction"`), the **seat verb**
(`seat/verbs.go:95` `New("friction", …)`, help keyed at `seat/help/friction.md`), and the
**envelope field** (`debate.js` declares `friction` in five envelope schemas at `:297,403,500,513,1263`,
aggregates at `:583-584`, emits at `:1300`; recovered by `capture.go:119` as `r["friction"]` and
`dashboard/render.go:147` as `j["friction"]`; tagged `json:"friction"` at `viewjson.go:1000`).
**All three rename**, in ONE atomic PR — the envelope is a cross-language contract between
`debate.js` and Go, and a half-applied rename silently breaks the capture join.

**Consumer census A — Go/proto, run 2026-09-06 from the repo root:**
```
$ P=plugins/frank-exchange-of-views
$ grep -rhno "Friction[A-Za-z]*\|frictionLog\|friction-parity" $P --include=*.go --include=*.proto \
    | sed 's/.*://' | sort | uniq -c | sort -rn          # ~25 symbols, ~300 occurrences
$ grep -rl "Friction\|frictionLog\|friction-parity" $P --include=*.go --include=*.proto  # 46 files
```
Counts: `Friction` 131, `FrictionNone` 41, `FrictionKind` 35, `FrictionEntryJSON` 11,
`FrictionJSONOf` 10, `FrictionFooter` 9, `FrictionAudit` 9, `frictionLog` 5, `FrictionJSON` 4,
`FrictionJSONBytes` 3, `friction-parity` 3, plus ~14 single-occurrence test names. 46 files; the
load-bearing ones, each changing:

| Carrier | Role |
|---|---|
| `record/recordpb/record.proto:11-14` (the retired permanence rule) and `:1163-1200` (incl. `FrictionNone` at 1198-1200) | the messages + enums (see No compatibility) |
| `record/record.go:921,925,931,935` | body dispatch for both messages |
| `record/viewjson.go:999-1053` | the JSON view — gains `source`/`type` |
| `record/estoppel.go:158,170,172` | the one live `kind` filter |
| `record/recordsql/schema.go:117` | **derives table names from message names** — see No compatibility |
| `report/assemble.go:1258,1279`, `docs.go:127,154` | the `## Log` markdown projection |
| `capture/capture.go:119,283,1719,1733` | envelope recovery + `FrictionAudit`→`LogAudit` (joins on seat; unaffected by the new fields) |
| `dashboard/model.go:29-31,96,270-282,348`, `render.go:147,340,424-434,615` | **operator-facing tile**, heading `<h2>Friction — logged pain points</h2>`, reads `FrictionJSONOf` |
| `seatprobe/seatprobe.go:206-248,386`, `boards.go:450,556,614,707,720`, `production.go:71` | probe expectations |
| `cli/seat/verbs.go:95,103,108,114`; `cli/blue/cite.go:70`, `blue/prove.go:66`; `cli/merge/mint.go:195-196` | the five write sites + the verb |
| `cli/friction.go:26,28,69` | the operator read + command name |
| `cli/root.go:218,233`; `cli/{blue,merge,lens,bench}/command.go` | root wiring + four `seat.Friction()` registrations |
| `cli/seat/seat.go` (`FrictionFooter`) | the footer closing every help page |
| `integration/surface/retiredsurfaces_test.go:18` | pins `show …friction` as a RETIRED surface — the retired name must stay pinned under its old spelling |

**Consumer census B — the string half (`*.js,*.mjs,*.md,*.json,*.golden`), same date. Scoped to
`plugins/`, NOT `$P`:** the envelope field is named in a SIBLING plugin, and a `$P`-scoped grep
cannot see it — the same wrong-scope class this plan already had to correct once, one directory
level up.
```
$ grep -rl "friction" plugins/ --include=*.js --include=*.mjs --include=*.md --include=*.json --include=*.golden
$ grep -rl "friction" docs/
```
48 files. **The cross-plugin carrier:** `plugins/prosthetic-conscience/skills/semantic-consent/SKILL.md:13`
— *"the escalation channel is your return envelope: report the gap as `friction`"* — an ALWAYS-ON
rule imported by this repository's `CLAUDE.md`, instructing every subagent to use the envelope
field this PR renames. Left unchanged it is exactly the outcome `complete-the-concept` names: a
constitution still teaching the old model after the code has moved. It changes with PR-2.

Prose surfaces, each changing: `agents/{lead-judge,blue-synthesizer,red-auditor,blue-researcher}.md`;
`skills/research-protocol/SKILL.md`; `skills/research-protocol/references/report_template.md`;
`commands/research.md`; `tools/internal/cli/seat/help/{friction,ingest,mint}.md`;
`docs/{seat-surface-naming,seat-duty-channel,why-a-seat-stops,seat-command-triggers,duty-docket-preregistration}.md`;
plus repo-root `docs/setup-script.md`. Machine surfaces: `hooks/hooks.json`;
`tests/simulator/{debate.test.mjs,harness.mjs}` + 13 `testdata/prompt-*.golden`;
`dashboard/testdata/render-{live,terminal}.golden`; 12 `difftest/testdata/*.golden`.
`#541`'s issue text names the channel and is updated with it.

### PR-2 — No compatibility, and the one refusal that makes that safe

**Directive: no migration, no archaeology, no backwards compatibility.** This is what the format
already says of itself (`record/schema_gen.go:10-13`: *"Nothing here promises backwards
compatibility, so there is no window to describe and no delta list to maintain: equal runs, unequal
refuses."*). So: renumber freely, no `reserved`, no delta list, no compatibility window, no
rewriting of archived tarballs, no code that reads a pre-rename record.

- **`record.proto:11-14` is retired by this directive and changes with it** — it is listed as a
  carrier above. It currently states *"FIELD NUMBERS ARE PERMANENT … the one piece of backwards
  compatibility the migration does NOT delete."* Left standing it is a comment teaching the
  opposite of the code, which is the half-state `complete-the-concept` names.
- **One thing is retained, and it is not compatibility — it is the precondition for having none.**
  The permanence rule existed to protect *readers of old records*. Renumbering is safe exactly when
  there are no such readers, which requires the binary to actually REFUSE them rather than quietly
  mis-decode. `recordsql/schema.go:117` derives table names from message names, so `friction`→`log`
  renames the SQL table and a new binary pointed at an archived run finds no `log` table and returns
  **zero rows — byte-identical to a clean board.** That is not backwards compatibility; it is a
  wrong answer, and `run-archive/` is re-read by every audit (`CLAUDE.md`), so the path is
  exercised rather than theoretical.
- **Declare the break:** bump `eventSchema` 2→3 in `requirements.json:3` and regenerate
  `record/schema_gen.go:13` via `scripts/schemagen`. This is the existing incompatibility
  DECLARATION, not a shim; it fires at `internal/setup/setup.go:584-594` (the `got != expectSchema`
  branch; `:509-517` is the manifest read supplying the expected value). Note `Event.schema_version`
  is not this gate — `recordpb/schemaversion.go:8-14` says that stamp is "read back by nothing."
- **Enforce it on read:** `record/store.go:82-93` refuses only when `dbName` is ABSENT with
  `legacyShards` (the JSONL era) present. An archived pre-rename run has a valid
  `records/record.db`, so `openRun`/`openRunForRead` proceed to the silent zero. Extend it to refuse
  a pre-rename SQLite schema (a `friction` table with no `log` table) in the same voice.
- **#501 is not a prerequisite.** Its body is measured on the JSONL era ("226 of 226 events have
  `schema_version` ABSENT") and is stale — `record.go:414` stamps the field on every write, so what
  survives is *read back by nothing*, not *emitted by nothing*. It also defers to the unresolved
  store-authority question, which would make it an unbounded blocker. Both mechanisms PR-2 needs
  exist today.

### PR-3 — Access-state as a field (#736) `[MODIFY]`

Retained from the prior draft **on its own evidence**, not as a destination for #710: the report's
single most load-bearing caveat is a prose substring, and the run miscounted it twice by grep
(`report.md:66` counts a sentence that counts itself). Add `Cite.source_text_read ∈ {LEAF |
SUMMARY_ONLY | UNREAD}` (`record.proto:905`, field 10; 8 reserved) and
`fetchcache.Entry.{http_status, refusal_class, access_state}` (`fetchcache.go:92`) — the status is
already seen and discarded at `httpfetcher.go:119`. Consumer censuses as pasted in the prior draft
(7 `Cite` consumers, 4 changing) and re-run at implementation.

**Re-examined for minimalism:** the prose alternative — "just write the caveat correctly" — is what
the run tried, and it produced two miscounts and a third hand-drawn access state. The field is the
smaller fix.

### PR-4 — Report voice (#710) `[MODIFY]`

Simpler than the prior draft, because the report now cites nothing: the operational half goes to
the log and the report does not point at it; the epistemic half stays, re-voiced, and stands alone.

- `[NEW]` One generated tell-set (`internal/reportvoice`, `Tells()`), modelled on `flags.All()` +
  `TestNoRenderedPromptSpellsAFlag` (`promptverbs_test.go:489`, pinned 0).
- A **non-blocking advisory** at `blue edit` (`edit.go:170`/`:142`) — names the tell and its
  destination, appends unchanged. Never returns an error. (Owner's call: flag for red, don't block.)
- A red **voice lens L7**: append to `RED_LENSES` (`debate.js:551-555`), push `{role: 7}` at
  `:853`, update **both** role-map strings (`:845` and `:856`). `findinglabel.go:12` `roleRe` is
  generic over `-L\d+`, so no label code changes. The lens carries **disclosure ≠ discharge**: a
  self-narrating sentence is a leak even where the report admits it.
- Research-voice clauses extending the existing AUDIENCE split (`lead-judge.md:117`,
  `SKILL.md:41`) to blue's report prose. `recordClause` (`debate.js:240`) is correct for
  act-reasoning and is left alone.

### WITHDRAWN — #737 (addressable friction id + report→log citation)

The measurement says the content it would have preserved is 65% ceremony, and the channel's own
doctrine says a capability gap is *"a report addressed to the human who can retool the seat, not
material for the debate"* (`friction.go`). A report→log citation re-imports what the design
deliberately routes away. Both seats, asked, wanted the opposite pointer (gap → tracker) and
neither wanted the report to cite the channel. **Close #737 with this reasoning.**

---

## IV. Risk & Mitigation

| # | Risk | L×I×C | Mitigation |
|---|---|---|---|
| R1 | Deleting the mandate also loses the small defensible subset (judgement-based declines). | med × med × low | PR-1 replaces the mandate with the narrow ask both seats defended; §V counts that those sentences still appear. |
| R2 | `REQUEST` invites wishlist inflation — seats filing features they'd like rather than gaps they hit. | med × med × low | The clause stays anchored to what the seat *walked into*; §V watches the request:defect ratio and the per-entry length. |
| R3 | Seats mis-type entries. | high × low × low | Mis-typing is cheap — the triager re-reads; a wrong type is strictly better than today's zero types. `source` is set by the write path, not asked. |
| R4 | The rename sweeps the word into a namespace where it means something else (facts-are-fields cl.4). | low × med × low | `friction` survives as a `LogType` VALUE with its meaning intact; the sweep is name-only. Verified: no stdlib `log` import, `TOOL_ERROR` has no readers, `ESTOPPEL` has exactly one. |
| R5 | Prompt-gate rejection on the PR-1/PR-4 wording (cost a round of reverts in #709). | high × med × low | §V runs all five gates; clauses phrased as acts, no verb/flag spelled, no obituary. |
| R6 | Measuring success on one future run over-fits to it. | med × low × low | The targets are ratios and absences, not tuned constants; the census commands are recorded so any later run re-measures identically. |
| R7 | **The rename silently blinds every archived run** — the SQL table name is derived from the message name (`recordsql/schema.go:117`), so a new binary reading an old record returns 0 rows, identical to a clean board. | high × high × med | No-compatibility section: renumber freely (no `reserved`), archived records unreadable BY DESIGN, and a LOUD refusal so "unreadable" is not "returns a clean board". The two changes that deliver it: (a) bump `eventSchema` 2→3 in `requirements.json:3` and regenerate `record/schema_gen.go:13` via `scripts/schemagen` — the epoch `setup.go:584-594` already refuses on; (b) extend `record/store.go:82-93`, which today fires only when `dbName` is ABSENT with `legacyShards` present and so lets an archived pre-rename `record.db` through to the zero. §V check 2 asserts the error rather than the absence. **Not blocked on #501** — both mechanisms exist today; see §III for why that issue is stale and would have been an unbounded blocker. |
| R8 | The envelope rename half-lands, breaking the `debate.js` → `capture.go` join across languages. | med × high × low | All three renames (CLI command, seat verb, envelope field) ship in ONE atomic PR; census B enumerates both sides; the friction-parity audit's own test covers the join. |
| R9 | Deleting the survey forfeits the only surface-traversal signal (the pin's stated rationale). | high × low × low | Accepted and recorded in PR-1 rather than argued away: the instrument is unfalsifiable and is false where checkable. A real instrument (hook-observed help invocations) is named, and deliberately not built here. |

## V. Verification Plan

All commands are rooted at the repository root; `P=plugins/frank-exchange-of-views`.

**Automated:**
- `(cd $P/tools && go test ./internal/record/... ./internal/cli/... ./internal/report/... ./internal/capture/... ./internal/dashboard/... ./internal/seatprobe/...)`
  — `Log`/`LogNone` round-trip; `source`/`type` persist on new fields 4/5; `(TOOL, DEFECT)` covers
  the retired `TOOL_ERROR`; `estoppel.go`'s filter still selects `(TOOL, ESTOPPEL)`; `LogAudit`
  still joins on seat; the dashboard tile renders the new heading.
- `(cd $P/tools && go test ./internal/reportvoice/...)` — `Tells()` pin/staleness gate.
- blue-edit advisory: a tell in `--new` **appends the event** and emits the advisory (proves
  flag-not-block).
- L7: a `red-lens-r?-L7` seat id yields `L7-F1` from the generic `roleRe`, no label-code change.
- **Schema refusal (the no-compatibility directive's own test):** open an archived pre-rename `records/record.db`
  with the new binary and assert it **errors naming the event-schema epoch**. Asserting the error is
  the point — a test that merely checks "no rows" would pass on the bug.
- Gates, each at its real path: `(cd $P/tools && go test ./integration/surface/... ./internal/cli/... ./cmd/seatprobe/...)`
  covers `promptverbs`, `constitutiondirective`, `retiredsurfaces`, and the naming tests
  (`internal/seatprobe/naming.go`, `cmd/seatprobe/naming_report_test.go`,
  `internal/cli/viewnaming_test.go`); `node $P/tests/simulator/debate.test.mjs`;
  `scripts/archaeology`; `(cd scripts && go run ./golden -update)`; `rulesweep` trailers on the
  protocol-surface commit.
- Full suite + `(cd $P/tools && deadcode ./...)` clean.

**Driveable checks on REAL data (all three required — fixtures cannot surface any of them):**
1. **Channel census, on a NEW run** (post-change; the query below cannot be run against the
   archived baseline, whose table is still `friction` — that is the point, not a
   loophole):
   `select e.seat_id, e.round, l.type, l.source, length(l.text) from log l join events e on e.id=l.event_id`.
   Report total chars (target ≤75,000), headed-survey chars (target 0), type/source coverage
   (target 100%), and — untargeted but measured — bucket-D share and the duplicate rate.
2. **Archived-run refusal.** Point the new binary at
   `run-archive/2026-09-02_quadratic-formula.tar.gz`'s `records/record.db` and confirm a **loud
   error naming the event-schema epoch** (not the schema-version stamp, which is not the epoch),
   never an empty projection. This is the check that would have caught the defect the first draft
   shipped.
3. **Report census.** Re-run the five #710 greps that produced 161 / 24 / 13 / 9 / 2 on the
   2026-09-02 report against the new run's `report.md`, and confirm every surviving access limit
   reads as a limit on the conclusion rather than an inlined hostname or status code.

**Auditor gate:** `/plan-audit` on this plan; then per-PR on each PR's own scope.

**Sequencing.** PR-1 first and alone — it is pure deletion, ships immediately, and its effect is
measurable on the next run independently of everything else. PR-2 next. PR-3 and PR-4 are
independent of both and of each other.

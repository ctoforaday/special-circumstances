# Report-voice separation — the research report speaks about the subject, not about itself

> STATUS 2026-09-05: not started — design proposed for review. Covers #710 (voice), #736
> (access-state destination), #737 (friction-id destination) as ONE concept delivered across
> three sequenced pull requests. Nothing built.

The assembled research report is written in DEBATE voice: it narrates its own construction —
lane tags in the prose, "this run / this round / the debate," the run's 403s imported as
epistemic caveats, its own draft history, its own verification apparatus. The microsite split
(#663-era) fixed the CONTAINER — `report.md` is its own file for a human audience — but not the
VOICE, which is still commentary-on-itself. Same shape as the report-lockdown before #709: the
separation exists structurally and is enforced by nobody.

This is ONE concept — *the report is about its subject* — but it cannot be delivered as a single
voice edit, because #710's own prescription ("move the operational half OUT, re-voice the
epistemic half") has **nowhere to move the operational half to**. #736 and #737 are the two
missing destinations; #710 is the enforcement. The concept is therefore three PRs, sequenced
**destinations first** (§III), and each PR is a complete sub-concept in its own right.

The governing principle throughout, from the issues and from `facts-are-fields`: **SEPARATION,
not deletion.** The residue is frequently load-bearing and must be kept, re-voiced — "Savage 1989
is known only through the interested party's summary" (a limit on the CONCLUSION) stays; "after
every full-text route failed across four hosts" (a fact about the RUN) moves to friction. Deleting
the second loses the first's warrant; inlining the second poisons the voice. A field and an
addressable id are what let a seat keep one and move the other.

---

## I. Summary & Goals

**Objective.** Make the assembled `report.md` read as research prose addressed to a human reader
of the *subject*, by (a) giving the two channels that bleed into it — process voice and the run's
operational friction — refusable destinations to live in, and (b) enforcing the voice at the
cheapest layer that can, escalating to red's judgment, never a silent hard block.

**Success criteria** — measured by re-running the #710 grep census (§V) on the report of a *new*
research run, against the 2026-09-02 quadratic-formula baseline (measured this session):

| Tell class | Baseline (report.md) | Target |
|---|---|---|
| process-voice: `this (run\|report\|round)` / `the debate` in report BODY | 161 | ≤ 5, and only inside a single designated provenance footer |
| inline lane-attribution tags (`[minority:…]`, `[lane-N…]`) in research prose | 24 | 0 |
| access-limit tells inlined (`403 to this container`, `still unread`, `could not be read at the leaf`) | 13 | 0 inlined — each relocated to an access-state field or a cited friction id |
| draft-history tells (`an earlier version of this sentence`, `corrected here`) | 9 | 0 |
| apparatus tells (`the checking program`, `measurement apparatus`) | 2 | 0 |

Plus two qualitative criteria a grep cannot score, checked by the red voice-lens on a real run:
1. Every access/operational limit that REMAINS in the report is re-voiced as a limit on the
   *conclusion* and **cites a friction id or reads from an access-state field** — never inlines
   hostnames or HTTP codes.
2. A voice leak that the report *discloses about itself* ("this is this round's correction to the
   report's account of itself") is still counted a leak — **disclosure is not discharge** (Fork B's
   red-protocol lever: a disclosed-but-load-bearing hole is a finding, not a reason to pass).

**Non-goals.** Not the perf investigation (#684). Not the report-as-record projection mechanics
(#709, merged). Not the broader red-triviality / source-trust concern (#247, #418, #549) beyond
the one disclosure≠discharge principle named above. Not deleting prose written for the human
reader — `facts-are-fields` clause 5: prose for a human audience is not the violation.

---

## II. Technical Context

- **Engine:** `skills/research-protocol/scripts/debate.js` — JavaScript, run under goja in the fuzz
  and by Claude Code in production. Seat prompts are string-interpolated and dispatched via
  `agent(...)`; the red lenses fan out through `parallel(...)`.
- **Tools:** `tools/` — Go (module `feov-record`), cobra verbs; the record is protobuf-defined
  (`internal/record/recordpb/record.proto`) over an append-only SQLite store; the report and the
  process docs are SQL projections (`internal/report/assemble.go`).
- **The fetch cache is NOT protobuf** — it is a JSON index (`internal/fetchcache`, one line per
  fetch), separate from the event record.
- **Constraints that already cost a reverted attempt (in #709 / PR #733) and bind this work:**
  - The report-access change must stay **INVISIBLE to a seat** — a seat reads "the report," the
    tool serves it. Over-editorializing the round-seat prompts is what forced the #709 reverts.
  - Agent-facing prose is gated: `promptverbs` (a prompt may not spell a verb/flag —
    `TestNoRenderedPromptSpellsAFlag`, pinned 0), `archaeology` (no obituary — a prompt may not
    narrate a gone capability), `debate.test.mjs` (the lens prompt must keep "read it whole in
    consecutive windows"; blue-synthesize must not say "through the tool"), `naming_test`
    (constitutions name no verb), `rulesweep` (a protocol-surface change needs `Rule-Class:` +
    `Sibling-Sweep:` trailers). Any prompt/constitution edit here passes all of these.
  - Proto schema additions bump the schema epoch under the existing version-gate discipline
    (`record.proto` version constant; `flags/names.go:19`).

---

## III. Proposed Changes (the spec)

Three PRs. **A and B ship before C** — C's enforcement redirects residue INTO A's and B's
destinations, so building C first leaves the seat with nowhere to move the operational half and
reproduces exactly today's leak. This is the complete-the-concept sequencing: each PR is a whole
sub-concept, and the thread is carried explicitly across the three (this file's STATUS line + the
three issues).

### PR-A — Access-state destinations (#736)

The source-side and fetch-side fields that let a seat separate "unread" from "absent" and
"container-egress refusal" from "origin refusal," as refusable fields validated at the write.

- `[MODIFY]` **`record.proto` `message Cite`** (`recordpb/record.proto:905`) — add
  `source_text_read` at field 10 (8 is reserved), an enum `{LEAF | SUMMARY_ONLY | UNREAD}`. This
  is the blue-side twin of the existing red-side `SourceOutcome`/`Confidence` (`:317-332`), which
  today only red's `Verify` carries. A citation to a source whose text was never read is a
  different object from one read at the leaf, and only the citing seat knows which.
- `[MODIFY]` **`fetchcache.Entry`** (`internal/fetchcache/fetchcache.go:92`) + the
  `<run>/cache/index` line format — add `http_status`, `refusal_class ∈ {container-egress |
  origin | unknown}`, `access_state ∈ {leaf-read | located-blocked | exhausted | absent}`. The
  status is already SEEN and dropped at `internal/fetchcache/httpfetcher.go:119` (403/405/407
  handling) — thread it from `Fetcher.Response` into `Entry` at `Store(...)` (`fetchcache.go:161`)
  and `fetch(...)` (`~:283`).
- `[MODIFY]` **`blue cite`** (`internal/cli/blue/cite.go:100`) — populate `source_text_read`;
  default `UNREAD` so the honest state is the one you get for free and `LEAF` must be asserted.
  `refusal_class`/`access_state` are set where the fetch is recorded, not by the citing seat.
- `[NEW]` a projection column / render so the report and the process docs can show access-state
  from the field instead of the prose phrase.

**Consumer census — `message Cite` (a schema contract).** Command and results (run 2026-09-05):
```
$ grep -rn "recordpb.Cite\|GetCite\|\.Cite{" tools/ --include=*.go | grep -v _test
```
| Consumer | Role | Changes? |
|---|---|---|
| `internal/cli/blue/cite.go:100` | writer (`&recordpb.Cite{}`) | YES — populate `source_text_read` (default `UNREAD`) |
| `internal/record/available.go:253` | type-switch on `*recordpb.Cite` | no — ignores new field |
| `internal/record/evidenceview.go:295` | evidence projection | YES — render access-state in the evidence view |
| `internal/record/citationid.go:67` | citation-id assignment | no |
| `internal/record/consistency.go:163` | consistency walk | no |
| `internal/report/{docs.go,proofs.go,assemble.go}` | report + bibliography projection | YES — render access-state instead of the prose phrase |
| `internal/report/assemble_cite_test.go`, fuzz/golden fixtures | tests/goldens | YES — assert the new field; `golden -update` |

Additive field with an honest default (`UNREAD`): readers that ignore it are unaffected; the four
YES rows are the render/assert sites. Re-running the grep surfaces nothing this table omits.

### PR-A friction-writer note

`source_text_read` is set by the *citing* seat; `refusal_class`/`access_state`/`http_status` are
set where the fetch is recorded (`fetchcache.Store`), not by any seat — so a citation and its
fetch record can disagree honestly (cited-from-summary over a leaf-readable fetch, or vice versa),
and that disagreement is itself a signal red can read.

### PR-B — Friction as an addressable, citable destination (#737)

A docketed gap gets an id the report cites (`R2-8`, `R3-4`); a pure-capability friction gap gets
none, so the report re-inlines its whole operational story. Give friction an id and a render, and
a report→friction citation path.

- `[MODIFY]` **`record.proto` `message Friction`** (`:1184`) — the friction event gains a
  run-unique **citable id** (e.g. `F12`), minted by the mint pattern mirroring `MintGapID` /
  `NextFindingLabel`. `FrictionKind` (`:1163`) unchanged. **Naming note (facts-are-fields cl.4):**
  the friction domain already uses "address" for the write-vs-read command surface
  (`frictionaddress_test.go`); this new field is a *citable id*, a distinct concept — the spec and
  code call it `friction id` / `citable id`, never "friction address," so a future sweep does not
  conflate the two.
- `[MODIFY]` **mint the id in ONE place, not per writer.** Friction bodies are appended from **five
  sites** — `internal/cli/seat/verbs.go:108` (the `friction` seat verb), `internal/cli/blue/cite.go:70`
  and `internal/cli/blue/prove.go:66` and `internal/cli/merge/mint.go:196` (tool-emitted), plus the
  estoppel path. A `MintFrictionID` helper called at the single `record.Append` path for a
  `*recordpb.Friction` body (where `citationid`/`findinglabel` are already assigned) gives every
  writer the id for free; each of the five then returns it so a seat can cite it.
- `[MODIFY]` **the seat-facing friction projection** — `internal/record/viewjson.go`:
  `FrictionEntryJSON` (`:1014`, fields `SeatID/Round/Text` at `:1033`), `FrictionJSONOf` (`:1023`),
  `FrictionJSONBytes` (`:1046`). This is the view a seat READS friction through (`friction.go:69`,
  `dashboard/model.go:285`); if the id is not added here, no seat can discover it to cite. **This is
  the load-bearing carrier the concept turns on.**
- `[MODIFY]` **`frictionLog` markdown projection** (`internal/report/assemble.go:1258`, header
  `:1279`, wired at `docs.go:127`) — render each friction row *with its id*, so `## Friction` is a
  citable index.
- `[NEW]` a **report→friction citation form** the report body can carry — a seat writes "Savage's
  contents are the interested party's summary `[friction F12]`" instead of inlining five hostnames
  and four status codes. The render resolves the pointer; the operational detail lives in
  `## Friction`, once.

**Consumer census — `message Friction` + `FrictionEntryJSON` (schema contracts).** Command and
results (run 2026-09-05):
```
$ grep -rn "recordpb.Friction\|GetFriction\|FrictionJSON\|FrictionEntryJSON\|frictionLog" tools/ --include=*.go | grep -v _test
```
| Consumer | Role | Changes? |
|---|---|---|
| `cli/seat/verbs.go:108`, `cli/blue/cite.go:70`, `cli/blue/prove.go:66`, `cli/merge/mint.go:196` | Friction writers (5 sites) | YES — return the minted id |
| `record/viewjson.go:1014,1023,1033,1046` (`FrictionEntryJSON`/`FrictionJSONOf`/`Bytes`) | seat-facing JSON view | YES — carry the id |
| `report/assemble.go:1258` (`frictionLog`), `docs.go:127` | markdown `## Friction` | YES — render the id |
| `capture/capture.go:283` (`FrictionAudit`, verdicts `:308/:314`); onRecord built `:1719,:1733`; envelope half `debate.js:619` | friction-parity audit | **no re-key** — the join keys on SEAT (`wroteToRecord[fr.SeatID]` ← `SeatOfAgent(e.AgentID)`, `:285-292`); the envelope carries no minted id, so the id can NEVER be the join key. (The 90-char slice at `:291` is display-only for the failure detail; the `:250` comment's "60-char substring" describes a superseded fallback, not the live join — do not trust it.) |
| `record/estoppel.go:172`, `seatprobe/seatprobe.go:247`, `dashboard/model.go:285` | kind-filter / probe / dashboard | no — ignore the id |

The id is additive; the four YES rows are the write/render sites, the "no re-key" row is the one the
prior census gestured at without locating. Re-running the grep surfaces nothing this table omits.

### PR-C — Voice enforcement (#710)

Layered cheapest-first. Your call, recorded: **flag for red, don't block.** The mechanical layer
is *advisory* — it never returns an error that refuses a `blue edit` — and the actual gate is
red's voice lens. Both read ONE generated tell-set (facts-are-fields: no second hand-kept copy).

- `[NEW]` **the voice tell-set, generated from one named source.** Model it on the flag-word
  precedent: `internal/flags/names.go` declares one constant per word, `flags.All()`
  (`names.go:241`) enumerates them, and `TestNoRenderedPromptSpellsAFlag`
  (`integration/surface/promptverbs_test.go:489`, pinned 0) is the single gate over every
  agent-facing file. Add `internal/reportvoice` with a `Tells()` enumerator (process-voice
  substrings/patterns: `this run|this round|this report|the debate`, lane tags `\[minority:…\]` /
  `\[lane-\d…\]`, apparatus `the checking program|measurement apparatus`, draft-history `an earlier
  version of this (sentence|bullet)|corrected here`). It is the ONE source both the blue-edit
  warner and the red voice-lens read. A **staleness gate** pins it (mirroring the flag gate); where
  a tell is a phrase that cannot be exhaustively generated, the gate is the documented fallback
  (`constitutiondirective_test.go:20-23` is the precedent for "generate where you can, guard where
  you can't, and say why").
- `[MODIFY]` **`blue edit` — a non-blocking flag-and-redirect.** Hook at `validateEdit` /
  `planEdit` (`internal/cli/blue/edit.go:170` / `:142`), which already receive `new`. On a tell
  match, **do not return an error** (that would block; the user's call is flag-not-block). Instead
  emit a stderr advisory naming the tell and its destination ("this reads as process voice; move
  the operational half to `blue friction` and cite its id; keep the epistemic half re-voiced"),
  and append the edit unchanged. The `ingest.go:53-55` refuse-and-redirect is the *shape* to copy
  but softened to advise-and-proceed. This is a courtesy redirect in-the-moment, not a gate.
- `[NEW]` **red voice lens (L7).** In `debate.js`: append a 4th dimension to `RED_LENSES`
  (`:551-555`), push `{ role: 7, lens: RED_LENSES[3] + <clause> }` at `:853`, and update **both**
  hardcoded role-map strings that enumerate the lens set as exactly L1-L6 and will otherwise speak
  the old six-lens model after L7 lands: the ROLE-STABLE dispatch comment at `debate.js:845`
  ("Citation slices are L1-L4, logic/completeness is ALWAYS L5, dark-side/risk is ALWAYS L6") and
  the lens-prompt role map at `debate.js:856`. The Go finding-label machinery is **generic over
  `-L\d+`** (`record/findinglabel.go:12` `roleRe`), so no label code changes — only the debate.js
  dispatch and the descriptive comment at `findinglabel.go:11`. The L7 lens reads the same
  `reportvoice.Tells()` set, mints findings on leaks the advisory layer let through, and carries the
  **disclosure≠discharge** clause: a self-narrating sentence is a leak even when the report admits it.
- `[MODIFY]` **constitutions + the two authoring prompts.** Give the existing AUDIENCE-split
  primitive teeth for blue's report prose. It already exists for the bench (`lead-judge.md:117`
  "the split is by AUDIENCE… the bench's own voice, never wearing the debate's authority") and in
  `SKILL.md:41` ("its audience is human. Write for the reader"). Extend it to blue: a research-voice
  clause in `agents/blue-synthesizer.md:33` (the report's author) and, minimally, in the
  blue-synthesize/blue-respond prompts (`debate.js:729`, `:1080`) — phrased as ACTS not verbs, no
  obituaries, keeping the report-access invisible to the seat. The red-auditor constitution
  (`agents/red-auditor.md:59`) names voice as an audit dimension. The global debate-voice clause
  `recordClause` (`debate.js:240`) is CORRECT for act-reasoning on the record and is **left alone**
  — the fix distinguishes the report SURFACES (human audience) from the reasoning channel, it does
  not de-voice the record.

**Consumer census — the L1-L6 role numbering (a prompt contract).** Command and results
(run 2026-09-05):
```
$ grep -rn "RED_LENSES\|ALWAYS L5\|ALWAYS L6\|L1-L4\|role: [567]" skills/research-protocol/scripts/debate.js
$ grep -rn "L1-L4\|L5 logic\|L6 dark-side" tools/ agents/ --include=*.go --include=*.md
```
| Carrier | What it says | Changes for L7? |
|---|---|---|
| `debate.js:551-555` (`RED_LENSES`) | the 3 dimension strings | YES — append the voice dimension |
| `debate.js:853` (`lensPasses.push`) | fixes roles L5/L6 | YES — push `{ role: 7, … }` |
| `debate.js:845` (ROLE-STABLE dispatch comment) | "L1-L4 … ALWAYS L5 … ALWAYS L6" | YES — enumerates the old six-lens set |
| `debate.js:856` (lens-prompt role map) | "L1-L4 citation, L5 logic, L6 dark-side" | YES — the seat-facing role map |
| `record/findinglabel.go:12` (`roleRe = -(L\d+)`) | extracts the role generically | no — regex is generic over `L<n>` |
| `record/findinglabel.go:11` (comment) | "L1-L4 … L5 logic, L6 dark-side" | doc-hygiene — update the comment, no logic |

The role NUMBER is consumed generically (the regex), so no label code changes; the four YES rows
are the debate.js dispatch + the two role-map strings (`:845` and `:856`) that would otherwise
speak the old six-lens model. The prior claim that the enumeration lived "ONLY" at `:551-555` was
wrong — `:845` is a third site. Re-running the greps surfaces nothing this table omits.

### Proposed structure (new/changed files)

```
plugins/frank-exchange-of-views/
  skills/research-protocol/scripts/debate.js        [MODIFY] RED_LENSES +L7, role-map string,
                                                             blue research-voice clause (acts-not-verbs)
  agents/blue-synthesizer.md                         [MODIFY] research-voice authoring clause
  agents/red-auditor.md                              [MODIFY] voice as an audit dimension
  tools/internal/reportvoice/tells.go                [NEW]    the ONE generated tell-set + Tells()
  tools/internal/reportvoice/tells_test.go           [NEW]    staleness/pin gate
  tools/internal/cli/blue/edit.go                    [MODIFY] advisory flag-and-redirect on new_text
  tools/internal/cli/friction.go                     [MODIFY] mint addressable friction id
  tools/internal/record/recordpb/record.proto        [MODIFY] Cite.source_text_read; Friction id
  tools/internal/fetchcache/fetchcache.go            [MODIFY] Entry.{http_status,refusal_class,access_state}
  tools/internal/fetchcache/httpfetcher.go           [MODIFY] thread status into Entry
  tools/internal/cli/blue/cite.go                    [MODIFY] populate source_text_read
  tools/internal/report/assemble.go                  [MODIFY] frictionLog renders ids; access-state render
```

---

## IV. Risk & Mitigation

| # | Risk | L×I×C | Mitigation (step) |
|---|---|---|---|
| R1 | **Over-redaction**: enforcement strips load-bearing residue (the Savage caveat) instead of relocating it → the conclusion loses its warrant. | med × high × med | SEPARATION-not-deletion is the stated principle; the mechanical layer is advisory (never deletes); PR-A/PR-B ship the destinations FIRST so there is somewhere to move the residue. §V criterion 1 checks residue is relocated, not lost. |
| R2 | **Prompt-gate rejection** — a voice clause spells a verb, narrates a gone capability, or drops the lens prompt's required phrase. Cost a full round of reverts in #709. | high × med × low | §II names all five gates; every prompt/constitution edit runs them (`promptverbs`, `archaeology`, `debate.test`, `naming_test`, `rulesweep`) in §V; clauses phrased as acts, report-access kept invisible to the seat. |
| R3 | **The tell-set becomes a hand-kept second copy** — reproducing facts-are-fields one level up. | med × med × low | ONE named source (`reportvoice.Tells()`), read by both the warner and the lens; a pin/staleness gate; the documented fallback doctrine where a phrase can't be generated. |
| R4 | **Voice lens false-positives on legitimate subject prose** (a quoted source says "this round"), or consumes red's construction budget. | med × low × low | Flag-not-block: the advisory layer never refuses; the L7 lens mints findings under red JUDGMENT (a false positive is cheap — it is argued, not enforced); L7 is a lens dimension, not a new lane, so it rides existing red dispatch. |
| R5 | **Silent scope truncation** — shipping C (voice) without A/B (destinations), which reads as done while the residue still has nowhere to go. | med × high × low | The concept is enumerated as all three PRs here; sequencing A/B before C is a stated gate; the STATUS line + three issues carry the thread (complete-the-concept). |
| R6 | **Schema-epoch / fixture churn** from the proto + cache-index additions. | low × med × low | Additive fields with honest defaults (`UNREAD`); version-gate bump per existing discipline; golden/fuzz fixtures updated in the same PR as the field. |

---

## V. Verification Plan (the checklist)

**Automated (exact commands):**
- `reportvoice` unit + pin gate: `go test ./internal/reportvoice/...` — `Tells()` enumerates the
  measured tell classes; the staleness gate fails if a tell is added in one reader and not the
  source.
- blue-edit advisory: a test that `blue edit` with a tell in `--new` **appends the event** (does
  not error) AND emits the advisory — proves flag-not-block.
- L7 dispatch + label: `go test ./internal/record/... ./internal/difftest/...` — a `red-lens-r?-L7`
  seat id yields `L7-F1` from the generic `roleRe` with no label-code change; the fuzz exercises a
  run that dispatches L7.
- proto/cache round-trip: `go test ./internal/record/... ./internal/fetchcache/...` — `Cite`
  carries `source_text_read`; `Entry` carries the three access fields; defaults are the honest
  ones.
- friction id: `go test ./internal/cli/... ./internal/report/...` — a friction event gets a
  run-unique id; `frictionLog` renders it; a report→friction citation resolves.
- prompt/constitution gates (the #709 lesson): `go test ./tools/integration/surface/...`
  (`promptverbs`, `archaeology`, `constitutiondirective`, `naming`) and
  `node skills/research-protocol/scripts/debate.test.mjs`; golden refresh
  `(cd scripts && go run ./golden -update)`; `rulesweep` trailers on the protocol-surface commit.
- full suite + `deadcode ./...` clean.

**Driveable check on REAL data (required, not synthetic):** re-run the #710 grep census (the exact
five greps in §I, which produced 161 / 24 / 13 / 9 / 2 on the 2026-09-02 baseline) against the
`report.md` of a **fresh** research run after C ships, and confirm the targets in §I are met AND
that each surviving access limit cites a friction id or reads an access-state field. Fixtures prove
the label logic and the schema round-trip; only a real run's report surfaces whether the *voice*
actually separated — the thing a grep over a synthetic fixture cannot stage.

**Auditor gate:** `/plan-audit` on this spec (Alignment, Completeness, Safety) before it is treated
as approved; then per-PR the plan-auditor on each PR's own scope.

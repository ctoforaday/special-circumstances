# One marker, one id — everything pinned to the report attaches the same way

> STATUS 2026-10-04: **APPROVED** by gblock on this revision, without a further audit (R-26).
> Plan-audit round 6 FAILed (2026-10-04) on one local gap, L1; revised 2026-10-04 against base
> `8798488d` (origin/main) for gblock's seventh rulings (R-24 the typed pair, R-25 the block
> openers, R-26 the approval). Every audit note still owed to implementation is a required test
> of its part (§V.4) and travels into that part's PR body; each part's PR gets an independent
> review before merge (§V.3).
> Eight parts in ship order: **Part 1 → 2 → 3 → 4 → 5 → 6**, Part 7 (fork test) in parallel,
> Part 8 after Parts 6 and 7. No fork open.

Paths are relative to `plugins/frank-exchange-of-views/tools/` unless they begin with `skills/`,
`docs/`, `tests/` or `agents/` (all under `plugins/frank-exchange-of-views/`). Line numbers are at
`8798488d`. Sweep commands are written to run at the tip from the repository root; the counts were
taken with `git grep … 8798488d -- …` (one more `[^:]+:` in a comment filter). A parenthesised
audit id names the gap or note a clause closes, in `~/.claude/scratch/markers-plan/audit-r<n>.md`:
round 2's gaps (written "audit G<n>", never the Goals of §I) and notes N1–N15, round 3's H1–H3,
round 4's J1–J2, round 6's L1. Census 2 and Census 3 are `census-ids.md` and `census-markers.md`
in the same directory.

## I. Summary & Goals

### The problem

The report has four things attached to places in it: findings, citations, proofs and gaps. It
attaches them in **two ways**, names them in **three id schemes**, and implements the first way
in many places.

- **Findings, citations and proofs** go through an in-text marker (`<!--fx:f-1a2b3c4d-->`,
  `<!--cite:c-…-->`, `<!--proof:p-…-->`). The tool places the marker, and blue carries it through
  every edit. The edit guard refuses an edit that drops a marker.
- **Gaps** go through a quote. `mint --quote` stores a sentence. Each later reader then *replays
  every edit since the mint* to guess where that sentence went (`internal/record/gapedit.go`,
  194 lines). The locator audit over all 28 known runs (201 gaps;
  `~/.claude/scratch/locator/FINDINGS.md`): quotes are unique at mint (102 of 102); after edits the
  location **widens** (74 of 191 grew more than 2×, median 102 words, maximum 536); 29 gaps in 11
  runs sit in a **blind spot** (an edit overlapping the quote, neither containing the other, is
  ignored); 6 drifted from ground truth, four cut to a bare marker no state names.
- **Ids.** Gaps, avenues and motions are tool-assigned sequences: `G3`, `Q2`, `M1`. Findings carry
  two ids: a random hex `f-…` underneath a sequential `<area>-F<n>` label. Citations (`c-…`) and
  proofs (`p-…`) are random hex.
- **Implementations.**
  - Twelve independent readers and writers of the marker grammar (Census 3).
  - Four write-time placers (finding, cite, prove, verify's corroboration) plus the replay insert;
    the four build the token by hand. Mint places nothing.
  - Four sentence splitters; three break at every `.`, `!` and `?`, so "3.5%", "e.g.," and
    "verify_91.py" end a sentence, and at every line break, so a soft-wrapped sentence reads as
    one fragment per line. Seven id functions. Four `report_op` insert arms.
  - One per-kind lifecycle rule: `retire` keeps a bare finding marker while a crediting gap is
    open or no gap credits it (`record.FindingMarkerHold`, 49 lines, and its caller in
    `retire.go`). Every other marker leaves with its claim.

A marker is **where the information actually lives**. The auto-carry measurement (211 carried
markers, 140 edits, 28 runs): 73% of carried markers sat in a sentence blue rewrote, and there no
tool-side locator recovered the claim better than 59% right with 9% silently wrong. Blue already
carries finding, citation and proof markers through exactly those rewrites. Gaps are the one
attachment that does not use that information.

### The objective

**One attachment mechanism, one marker lifecycle, one placement event, one id scheme, one
implementation.** Every place-pinned thing is a marker the tool places. Every marker, of every
kind, lives the same way: the tool places it at the end of its quote; blue carries it through its
edits; the tool re-places it when its sentence survives an edit verbatim; the edit guard refuses
an edit that drops or introduces one; `retire` is its one exit. A kinds table holds only what a
kind MEANS. Every id has one shape, `<LETTER>-<8 hex>`. The gap-location replay is deleted.

### Success criteria (each a command; today at `8798488d` → target)

| # | Criterion | Command | Today | Target (part) |
|---|---|---|---|---|
| S1 | Marker spelling has one home in code | `git grep -nE '<!--(fx\|cite\|proof):' -- 'plugins/frank-exchange-of-views/tools/*.go' ':!*_test.go' \| grep -v internal/anchor/ \| grep -vE '^[^:]+:[0-9]+:\s*//'` | 16 lines / 12 files (74 / 27 with comments) | 0 (Part 1); with comments 0 (Part 8, when it runs) |
| S2 | One write-time placement function | `git grep -nE 'InsertAnchor\(' -- 'plugins/frank-exchange-of-views/tools/*.go' ':!*_test.go' \| grep -v 'func InsertAnchor'` | 5 (cite.go:171, prove.go:152, finding.go:119, verify.go:349, render.go:124) | 2 after Part 1: inside `anchortext.Attach` and the replay insert in `render.go`; 1 after Part 4: the replay insert — `Attach` places through `LocateOnce` |
| S3 | One sentence splitter | `git grep -nE "case '\.', '!', '\?'\|IndexAny\([^)]*\.!\?\|!= '\.' &&" -- 'plugins/frank-exchange-of-views/tools/*.go' ':!*_test.go' \| grep -v internal/anchor/` | 5 lines / 4 files (claimcount.go:89, render.go:234, bluedoc.go:361,368, assemble.go:739) | 0: three files in Part 1, `assemble.go:739` in Part 2. A renamed splitter still matches; a reader that splits another way fails `TestEverySentenceReaderReadsOneSentence`. |
| S4 | One id minter | `git grep -nE 'rand\.Read\(\|Sprintf\("[GQMCFP]%d\|Sprintf\("%s-F%d' -- 'plugins/frank-exchange-of-views/tools/internal/record/*.go' ':!*_test.go' ':!*/lock.go'` | 9 (7 minters + 2 in `migrate/remap.go`) | 1, inside `record.NewID` (Part 6) |
| S5 | Gap replay gone | `git grep -nwE 'GapEdits\|CurrentLocation\|relocate\|mintedIfMoved\|editsSince\|location_edits\|minted_location' -- 'plugins/frank-exchange-of-views/tools/*.go' ':!*_test.go'` | 26 | 0; `gapedit.go` deleted (Part 4) |
| S6 | Net code reduction | `git diff --numstat 8798488d -- 'plugins/frank-exchange-of-views/tools/*.go' ':!*_test.go' ':!*.pb.go' ':!*_gen.go'`, summed per part | — | ≥ 320 net production lines removed across Parts 1–6 (estimate ≈ 350: Part 1 −180, Part 2 +30, Part 3 +70, Part 4 −170, Part 5 +35, Part 6 −135). Each PR reports its measured figure and explains a deviation over 30% from its estimate. |
| S7 | Gap location precision on real runs | locator-audit harness over the 28 migrated runs (§V.2) | widening 74/191; blind spot 29; drift 6; no stated location for 89 | every quote gap carries one of three `location_state`s; every `marked` location is the one sentence holding its marker (widening 0, blind spot 0), and the six fragment gaps in renderable runs hold their token, at the mint, in the whole sentence §III.2 names, as do b3's three gaps; drift against the finding-token ground truth 0 except the F-c fallback placements, listed by name (Part 4) |
| S8 | Auto-place is right where it acts | auto-carry harness (§V.2) through Part 3's `bluedoc.AutoPlace` | — | On every carry whose sentence survives verbatim in `new` and which `LocateOnce` finds there once (≈18%): the tool's placement equals blue's. The ≈3% where blue moved a marker off a surviving sentence is the stated residue, counted and listed. Every other drop is refused with its sentence named. |
| S9 | A finding id never prints without its area | `TestEveryFindingIdPrintsItsArea` (Part 6) | — | every `F-<8 hex>` the tool prints — outside a marker token and outside a seat-argument field's own text — is followed by ` (<area>)` |
| S10 | Id and marker shapes are read from one table | `git grep -nE 'MustCompile\(.*(G\\d\|Q\\d\|M\\d\|F\\d\|\[fcp\]\|[fcp]-\[\|fx:\|cite:\|proof:\|<!--)\|HasPrefix\(id, "[fcp]-"\)\|\[a-z\]\{1,2\}-\[0-9a-f\]' -- 'plugins/frank-exchange-of-views/tools/*.go' ':!*_test.go' \| grep -vE 'internal/anchor/\|internal/record/migrate/'` | 25 | 0 (Parts 1, 4, 6); `migrate` keeps its readers of the OLD shapes |
| S11 | Agent-facing text speaks one id shape | the §III.6 agent-text sweep | 35 lines | 0 (Part 6; spelling-only lines in Part 8) |
| S12 | No per-kind lifecycle branch | `git grep -nE 'FindingMarkerHold\|takeable\(\|HasPrefix\(id, "f-"\)' -- 'plugins/frank-exchange-of-views/tools/*.go' ':!*_test.go'` | 8 lines / 3 files | 0: `anchor.go:94` in Part 1, the rest in Part 4. A renamed hold still fails `TestRetireTakesABareMarkerOfEveryKind`. |
| S13 | No text says a marker outlives its retire | `git grep -niE 'immortal\|marker remains\|neither side deletes\|never removes? a marker\|content is gone' -- plugins/frank-exchange-of-views scripts/agentgen/src ':!*_test.go' ':!*.pb.go' ':!*testdata*' ':!*/agent-memory/*' \| grep -vE '^[^:]+\.go:[0-9]+:\s*//'` | 16 lines (35 with Go comments) | 0 (Part 4). Of the 19 Go comments, 10 are rewritten with their code (`anchor.go:1`, `anchortext.go:1`, `claimcount.go:351,358,390`, `assemble.go:52`, `edit.go:25,270`, `citationid.go:264`, `record.go:788`; Parts 1, 3, 4); 9 use "immortal" for the edit guard and stay (`anchortext.go:288`, `bluedoc.go:5,13`, `cite.go:26`, `finding.go:79,146`, `citationid.go:20,239`, `scorecard.go:501`). |
| S14 | A kind list names every kind | the §III.4 kind-list sweep | 29 lines | 7, each a list of meaning, not of all kinds: the four `Backing` docs (`viewjson.go:103,739,838,1249`, *Backs*), `window.go:94`'s `show evidence` clause, `debate.js:960` and `assemble.go:3` (not anchor kinds). Comments in Part 1, text in Part 4. |

### Value case

Patching the replay (a gone state, the blind spot, a widening cap) adds code to the weaker
mechanism and still guesses in the 73% of carries only the author knows. This plan **removes**
code: it keeps the mechanism blue already operates for three kinds, minus its one per-kind
exception, at the cost of one translation per record change and a spelling experiment. It buys
one grammar and one lifecycle instead of twelve readers and a finding-only hold; gap locations of
one sentence with a named state instead of a paragraph-sized guess; one id shape whose letter
names its kind, minted with no counter to race.

### Goals

- G1 — One marker grammar and lifecycle: a kinds table of meaning (letter, label, assembly,
  claim, backs) that every reader and writer consults; no lifecycle rule asks a marker's kind
  (Part 1; the last one goes in Part 4).
- G2 — One placement function, `anchortext.Attach(doc, id, quote)`: the marker at the quote's end.
  Part 1 keeps today's rule; from Part 4 it places through `LocateOnce`, the one write-time
  matcher (Part 3), at the one occurrence that matcher counted, or refuses.
- G3 — One sentence splitter, `anchor.Sentences` (Part 1), whose rule ends a sentence only where
  one ends — at a terminator before a new sentence or at the end of a markdown block, never at a
  soft wrap — read by every reader of a marker's sentence (Part 2).
- G4 — On `edit`, the tool re-places a dropped marker whose sentence survives verbatim; otherwise it
  refuses and names the sentence (Part 3).
- G5 — Gaps attach by marker; the replay is deleted; every quote gap's location is `marked`,
  `gone` or `unrendered`; "what changed under this gap" is the edits whose `reopened` names it,
  and a cut sentence is an answer (Part 4).
- G6 — One placement event: every placer appends `Anchor`; `report_op` has one insert arm (Part 5).
- G7 — One id scheme `<LETTER>-<8 hex>` minted by `record.NewID`; a finding's area leaves its name
  and one renderer prints it beside the id (Part 6).
- G8 — A legible marker spelling, identical for every seat, chosen by a fork test (Parts 7, 8).

### Non-goals

- **No per-seat view of markers.** Every seat sees the same text (ruled 2026-09-30).
- No change to footnote numbering: citations stay `[^N]`, proofs `[^P<n>]`. Ids are not footnotes.
- No change to seat keys (`--key`, the `*_key` fields). Only their help example changes (§III.6).
- Gaps about an avenue or section (`about_ref`, 10 of 201) stay unplaced and carry no state.
- No span machinery: a gap is one sentence (ruled 2026-10-03).
- **The re-run key stays the script's sha256** (ruled 2026-10-03, R-10). `Reproduce` is untouched:
  the sha names a script, not a place in the report, and proofs sharing a script share a re-run,
  as today. Its readers keep their join: `available.go:406,433,456`, `capture/proofrerun.go:76`,
  `citationid.go:387,390`, `evidenceview.go:352,364,429`.

## II. Technical Context

- **Vocabulary.** "Marker" in this plan is the terms registry's **anchor** (`terms.json`); every
  agent-facing text the plan specifies says "anchor".
- **Language and build.** Go 1.27 (`GOTOOLCHAIN=local`). The FEOV tools module is
  `plugins/frank-exchange-of-views/tools`. The engine is `skills/research-protocol/scripts/debate.js`.
  Agent definitions are generated by `scripts/agentgen`.
- **Record.** SQLite, event-sourced, one protobuf payload per event type (`record.proto`). The
  event-schema epoch was **18** at `8798488d` and is **20** at `595744c6`, where Part 1 starts; the epoch steps below count from 20, and each part re-reads the tip's epoch before it bumps. Old runs are brought
  forward by `migrate`, which replays them through the current write path with `record.Migrating`
  set. Each Migrating exemption sits beside the check it exempts (`record.go:547-556`; `grep -n
  Migrating` is the list); structural checks — references, UNIQUE, `--distinct-from` — stay on.
  `migrate.Registry` is keyed by archived word (`registry.go:32`): one translation per word, straight
  to the current shape. `Translate` returns `[]proto.Message`; each returned body then passes
  `remap.apply` (`replay.go:181`) and `Append` (`:194`). A same-sitting correction's replacement must translate to
  exactly one body (`replay.go:163`). `res.SourceHash` (`replay.go:79`) is sha256 over the source
  event stream.
- **Archive.** `run-archive/` holds 16 run tarballs. `TestEveryArchivedRunMigrates`
  (`migrate/everyarchive_test.go:14`) replays all 16 and counts refusals; nothing renders them.
  7 of the 16 (2026-08-22 … 2026-09-02) hold no `base_ingest`, so no report renders; the other 9
  do. Census over the extracted tarballs: 472 `fx`, 253 `cite`, 141 `proof` tokens and **no
  other HTML comment** (`grep -raoE '<!--[a-z]+:'`); 0 duplicate `Anchor` ids; 3 corrections
  (2 avenue, 1 revision — none of a cite, proof, verify or mint); 1 exact-span edit (b7).
- **Marker lifecycle today.**
  - A placer mints an id, builds a token, and checks `anchortext.InsertAnchor` against the current
    render. `InsertAnchor` puts the token right after the quote's last content character — before
    any trailing punctuation and before any marker already there (`anchortext.go:locate`).
  - `reportproj.RenderFromRecord` (`render.go:274`) folds `report_op` (`recordsql/views.go:706`):
    four insert arms (cite, proof, anchor, verify), a remove arm from `retire_anchors`, and edits.
    **Replay re-runs placement and its guards on every render:** `InsertAnchor` (`render.go:124`)
    for inserts; `LocateUniqueReplacing` (`render.go:100`), and with it `settleAbuttingAnchor`, for
    each splice; `LocateLiteral` (`:97`) for exact-span edits; `RemoveAnchorAt` (`:159`) for
    removes.
  - `bluedoc.AnchorsTransitUnchanged` (`bluedoc.go:196`) refuses an edit that drops a marker;
    `ErrAnchorIntroduced` (`:237`) one that types a marker; `settleAbuttingAnchor` (`:129`) one
    whose quote stops just before a marker run; `ReopenedAnchors` (`:309`) records the anchors whose
    sentence an edit changed into `BlueEdit.reopened`, skipping an anchor the edit left bare.
  - `retire` is the only exit, with one per-kind exception: `takeable` (`retire.go:266-281`)
    branches on `HasPrefix(id, "f-")` and keeps a bare finding marker while
    `record.FindingMarkerHold` (`estoppel.go:276-324`) finds a crediting gap open or none.
  - Assembly strips `fx` (`report/docs.go:222`) and weaves `cite` into `[^N]`, `proof` into
    `[^P<n>]`.
- **Gap location today.** `mint --quote` is checked with `bluedoc.LocateUnique` (`mint.go:127`:
  whole-document unique, may cross a paragraph, may not split a word) and stored raw. `mint --new`
  is checked by `ValidateProposal` and stored raw as `fix_new`; `edit --accept` replays
  `record.Proposal` (`estoppel.go:143`) as the edit's old and new. `viewjson` computes location,
  passage, backing, `minted_location`, `location_edits` and `edited_since` through
  `record.GapEdits`/`relocate`. A render failure leaves every passage empty without saying so
  (`viewjson.go:353-364`).
- **Constraints.** No backwards compatibility; old records are translated by `migrate`, one
  translation per change, each with a test on archived runs. Agent-facing text names acts, not
  verbs, and is present tense; `requiredmarker_test.go` and `deadpath_test.go` gate help and
  refusal text; `TestTheSeatPromptsNameNoVerb` gates prompts. Prompt ceilings: judge 7900,
  red-lens-evidence 12300. A seat never types a marker (`ErrAnchorIntroduced`).

## III. Proposed Changes

### III.0 Forks — all ruled (first set 2026-09-30 … 10-03; second to fourth sets 2026-10-03; fifth to seventh sets 2026-10-04)

| # | Fork | Ruling | Why |
|---|---|---|---|
| F-a | **Id shape** | **`<LETTER>-<8 hex>` from `crypto/rand`, for every kind**: `G3`, `Q2`, `M1` become `G-5e10a3c2` and so on (gblock, 2026-10-02). | One shape, one minter. An id you cannot guess is one you look up (a 2026-07-18 chair "continued the sequence" L6-F8..F16). Minted before the append, as cite, prove and finding already are. A timestamp hash collides for two seats in one tick. |
| F-b | **A closed gap's marker** | **Persists, inert** (gblock, 2026-09-30). | The one lifecycle (R-6): `retire` is the one exit; assembly strips the gap kind. |
| F-c | **Old runs' gaps in migration** | **The stored quote at the mint replay, then the verbatim rule, otherwise the start of the dropping edit's replacement** (gblock, 2026-09-30). | A migration translation only (§III.4), never a read-time fallback. |
| F-e | **Legible spelling** | Decided by the Part 7 fork test under its stated rule. | — |
| F-f | **The finding's area** | **It leaves the name and is printed beside the id** (gblock, 2026-10-02): `F-7cdcd115 (adversary)`, `"area":"adversary"`. | The area is the finding event's `seat_id`; the label copies it. Nothing parses it back out. |
| R-1 | **Order** | **Gaps before ids** (gblock, 2026-10-03), with the third set's split and the fifth set's sentence part: Part 1 → 2 → 3 → 4 → 5 → 6; Part 7 in parallel; Part 8 after Part 7 rules and after Part 6. | Gap attachment is the value; the placement event and the ids follow it. |
| R-2 | **Locate rule in Part 1** | **Part 1 is a pure refactor**, proven by a committed test that every archived run renders byte-identically (gblock, 2026-10-03). | Replay re-runs placement; a rule change in a no-migration part would move history silently. |
| R-3 | **Auto-place** | **Build it**; the ≈3% residue is stated (S8). | — |
| R-4 | **Gap granularity and position** | **One sentence** (gblock, 2026-10-03). The gap's marker sits where every kind's marker sits — **at the quote's end**, by `Attach`'s rule (third set, the author's derivation from R-6, disclosed to gblock). The gap's sentence is the sentence holding the marker, which is the quote's last sentence. The board shows that sentence plus its section passage; mint's help asks for one sentence. | One placement rule; position tested at write, replay and migration (§III.4). |
| R-5 | **Fork-test tier** | **Production models (sonnet tier), n ≥ 5 per seat kind per candidate**. | — |
| R-6 | **Marker behaviour per kind** | **Identical for every kind** (gblock: "I'd love no special rules - everything behaved the same way"). No hold: `record.FindingMarkerHold` and its caller are deleted (Part 4). The kinds table keeps meaning only. | The hold's rationale is answered by the `gone` state (R-7). |
| R-7 | **Location states** | **Three**: `marked` — the sentence holding the gap's marker; `gone` — the marker stands at no prose (cut, retired, or never placed), the minted text shown; `unrendered` — the report cannot render, the minted text shown. About-gaps carry no state. | Each is one a reader acts on differently. |
| R-8 | **Remap scope** | **Inverted** (third set): bare old ids are rewritten only in fields that hold ids and in process prose (logs, closing arguments, reasons). Report text and source data change only by exact marker-token rewrite. `migrate` keeps hand lists of the fields that MAY change (R-19); a test asserts byte-identity of every other string field over every tarball. | A rewritten `Proof.script` breaks its sha; "Q2 earnings" in a title is the subject's. |
| R-9 | **Cut = answer** | **Teach `gone`, no new machinery** (third set). Red's prompt and `WorkGapJSON.EditedSince`'s doc teach `gone` as an answer, not silence, naming both causes — blue cut the sentence, or in a migrated run the quote never placed — and asserting neither; the change history names a cutting edit. Every carrier of "empty `edited_since` means unanswered" is swept (§III.4). | `ReopenedAnchors` skips the cut by design; the state carries the fact. |
| R-10 | **Re-run key** | **Keep the sha** (third set). `Reproduce` is untouched (Non-goals). | The sha names a script, not a report thing. |
| R-11 | **Split the id part** | **Three pieces** (third set): the write-time unique-quote check moves into Part 4 (mint needs it); one placement event ships alone (Part 5); id respelling, the finding-label collapse and the area renderer ship alone (Part 6). Each record change carries its own translation and epoch step. | Smaller migrations; the mint needs the check the moment it places. |
| R-12 | **Prose ids** | **Rewritten in every seat-argument field** (fourth set): bench opinions, outcome prose, manifest rows, docket rulings, reproduce notes, motion relief, observations, `Proof.drift`, positions, certifications, declarations, halts — every field holding a seat's own wording, by sweep (§III.6: 50 fields). | An id in a seat's argument is a reference to the record. |
| R-13 | **`--accept` after a later marker** | **Survives** (fourth set): the proposal served to blue carries the marker run that now follows the quoted text at the gap's location — read-time, in `record.Proposal` (A-3); `applied_verbatim` compares a typed pair with that function's output (R-24); estoppel compares `Visible` text, as today. | A second lens minting on the gap's sentence must not cost red's fix its estoppel. |
| R-14 | **Residues** | **Accepted** (fourth set): (1) a sentence repeated verbatim as its own paragraph cannot carry a marker (R2); (2) a migrated run keeps the gap markers a historical cut left bare, and their husks (measured: R-20). The skeleton digest takes a gap marker out as a retire would, husk included, and is otherwise unchanged (§III.1). | — |
| R-15 | **Process** | Superseded by R-26. | — |
| R-16 | **Sentence unit** | **Fix the splitter, as its own part right after Part 1** (fifth set, gblock, 2026-10-04): Part 2, a bug fix — claim counts read fragments today too. A terminator ends a sentence only when followed by whitespace and then a capital, an opening quote or bracket, or the end of text; closing quotes and brackets after it belong to the sentence (the full rule, §III.2, with R-21 and R-22). Claim-count goldens move and are read line by line; archived renders stay byte-identical because no archived run holds a `retire_anchors` row (measured, §III.2). The seven corpus fragments are named by a test. One splitter, `anchor.Sentences`, serves every reader of a marker's sentence. | "The sentence holding the marker" is only as good as the splitter that finds it. |
| R-17 | **`--accept` carry** | **Through auto-place, or refused** (fifth set): the non-gap anchors in the run at the gap's location are re-placed by `AutoPlace`'s verbatim rule into the accepted text; otherwise the accept is refused as stale and blue edits by hand. No fourth placement path (A-3). | The accept is an edit; it carries as every edit does. |
| R-18 | **F-g** | **Extended to every kind** (fifth set, confirmed by gblock). | — |
| R-19 | **Field lists** | **Three hand-kept lists in `migrate`** — ids, seat-argument, source data — **and a guard test** failing when any `record.proto` string field of a message reachable from an event body is in none of them or in two (fifth set; §III.6). | A field added to the schema must be classed before it migrates. |
| R-20 | **Husk residue** | **Stated with its count** (fifth set; the author's wording, from the measurement): 0 of the 21 renderable corpus runs holds a husk beside a gap marker — no run has a `retire_anchors` row; four historical cuts leave a gap marker bare with no husk (b6 `G5`, b9 `G1` and `G4`, m7 `G2`; they read `gone`). Only migrated runs can keep husks (R-14); live runs retire normally. | — |
| R-21 | **Soft wraps** | **A soft wrap is whitespace** (sixth set, gblock, 2026-10-04): a newline ends a sentence only where it ends a markdown block — a blank line, or a next line that opens a list item, heading, table row, block quote, fence or footnote definition, each by markdown's own rule (R-25). The claim index's per-line segments become per-sentence segments that may span lines, reported by their first line. b3's three gaps and the 19 soft-wrapped anchors read whole sentences, named by a test (§III.2). | A sentence the author wrapped is still one sentence. |
| R-22 | **After a closer** | **Test the boundary again after the marker run** (sixth set): an anchor run flush after a closer belongs to the sentence the closer ends, and whitespace then a capital or an opener after the run ends that sentence ("**Wet.**<c> Next." is two). A `TestSentences` row has a following sentence. | A tool-made shape (`cite` places after a closing `**`) must not join two sentences. |
| R-23 | **Process** | Superseded by R-26. | — |
| R-24 | **The typed pair** | **One function serves both** (seventh set, gblock, 2026-10-04): `ProposalAppliedVerbatim` compares the typed pair with the output of `record.Proposal`, the function that serves the board's pair and `--accept`'s; no new mechanism. A test of b3 `G1`'s shape (L1) and the release fuzz's typed arm (§III.4). | `mint --quote` locates after whitespace collapse, so the stored location can differ from the render in whitespace (2 of 102 corpus quote locations: b3 `G1`, `G3`); a compare against the stored location loses a typed application `--accept` would record. |
| R-25 | **Block openers** | **Markdown's own rules, in the one block reader** (seventh set): `#` opens a heading only when followed by a space or the end of the line; only `1.`/`1)` starts a list that interrupts a paragraph, any number after a blank line or inside a list (as the next item of the ordered list already open, §III.2); `\|` opens a table only when the next line is a delimiter row; a `>` line after a `>` line continues the quote, its `>` stripped and read as whitespace by rule 2, a `>`-only line ends the paragraph inside the quote, and openers inside a quote are read after stripping `>`. `TestSentences` rows for each, the auditor's three `>` rows included (§III.2). | A newline ends a sentence where markdown ends a block, and nowhere else. |
| R-26 | **Approval** | **gblock approves this revision without another audit round** (seventh set). Every audit note still owed to implementation is a required test of its part (§V.4); each part's PR gets an independent review before merge (§V.3). | — |
| A-1 | **Mint placement** | Mint appends an `Anchor` event after `Mint`; `report_op` gains no insert arm. | The `anchor` arm already renders it. |
| A-2 | **Unique-quote check** | Write-time only, through `LocateOnce` (Part 3; `Attach` from Part 4); replay keeps `InsertAnchor`'s first match, which lands where `LocateOnce` counted on every write it admits (§III.4). | Replay reproduces history; the write proves uniqueness. |
| A-3 | **The proposal carries the marker run** (R-13, R-17) | `record.Proposal(run, gap, report)`: where the location, located as an edit locates it, is followed by a marker run holding the gap's anchor, `old` is the location as the render holds it (its bytes through that run, then its own trailing punctuation), and `new` is `fix_new` with each of the run's gap anchors it does not already carry inserted at its content end by `Attach`'s insert step — the position `Attach` gives any quote ending there, with no second locate (the author's reading of R-17: `fix_new` is the gap's new text). So `new` carries each gap anchor of the run exactly once (§III.4). The run's other anchors are not added: `PlanSplice`'s `AutoPlace` re-places each that `new` lacks whose sentence survives verbatim in `fix_new`, and its transit check refuses the rest, which `--accept` reports as stale (`edit.go:137-140`). Otherwise the raw pair. `fix_new` stays red's text. | No fourth placement path: `Attach` and `AutoPlace` only. |
| A-4 | **Stated residues** | Ruled: R-14 and R-20 — every placer refuses a sentence repeated verbatim as its own paragraph as ambiguous (R2). Disclosed: the three field lists are hand-kept in `migrate` under a guard test (R-19; `(prose)` means correction tier, a different concept; a test holds every `(prose)` field to the seat-argument list but `Cite.title`, which is source data); seat-argument text is the seat's wording and prints as written, so S9 does not scan it — the remap rewrites b3's `Retire.reason` "voice-F1" to the bare id like every id (reversible: writing `F-… (area)` there is one branch in the remap and drops S9's exemption); the sentence rule (§III.2, measured on the 21 renders) joins a sentence that opens with a digit to the one before it (37 joins; three more digit boundaries keep "ca. 240", "no. 2", "p. 33" whole), splits after an initial (5: "G. L. Honaker" ×2, "J.-M. De Koninck", `Eric W. "…"` ×2; "Dr. Smith" splits, stated), and splits before an opener after any terminator that does not end a sentence (5: the two `W. "`, "Prime Curios! (The", "30.33... (not" ×2; "Smith et al. (2020)" and "e.g. `verify_91.py`" split, stated), and a terminator inside a table cell splits the row (0 tables in the corpus's renders); a seat-argument field using a mapped id's spelling for something else is rewritten; ids re-carried by corrections (`cite.label`, `proof.proof_id`, `avenue.avenue_id`) carry no UNIQUE; the ≈3% moved-marker carries (S8); the F-c fallback placements (R10); unique-quote refusals of quotes that once placed silently (R2); in a migrated run, a gap whose quote never placed reads `gone` with no cutting edit, listed by name; an archived edit that no longer locates because a kept husk stands in its span is a migration refusal, named by run; a run live across Part 2's upgrade with an anchor on a fence's own line stops rendering, loudly (none archived, §III.2); retiring the bare anchor of an emptied ordered list item leaves its "1.", as today (§III.2). | — |
| A-5 | **`extractQuote`** | **Deleted in Part 4** (CLAUDE.md: `migrate` is the only upgrader, so the live renderer keeps no path that exists only for old records). From Part 4 no write reaches it; Part 4's migration placement step rewrites any archived placement whose stored location resolves only through the fallback into the quote it resolves to, and `TestEveryArchivedRunRenders` plus the universe comparison stay identical. The count of such archived placements is measured in Part 4 and pasted into its PR. | — |
| F-g | **The crash-window check** (`consistency.go:513-517`) | **Extended to every kind** (R-18): mint in Part 4; cite, prove, verify in Part 5. The violation reads "has no anchor". `consistency` runs in `citecrash_test.go` and the release fuzz; on a migrated run it reports each never-placed gap, truthfully. | One lifecycle (R-6). |

Round 6's gap (L1) and notes, each with its resolution and required test: §V.4.

### III.1 Part 1 — one grammar, one splitter, one placement function (pure refactor)

**Value.** Estimated −180 production lines; **measured −48** (+359 −407): the kinds table and its readers cost `anchor.go` +126 that the estimate did not price, so the S6 total now rests more on Parts 4 and 6. Nothing a seat, a record or a render sees changes. Every
later part becomes a table row or column instead of a twelve-site sweep.

**The kinds table** (`internal/anchor`) holds what a kind MEANS, never how a marker of it lives:
no column decides placing, carrying, protecting, auto-placing or retiring (R-6). Part 1 has three
rows; Part 4 adds `gap`; Part 6 respells the ids; Part 8 respells the tokens.

| Kind | Id (Part 1 → Part 6) | Token (Part 1) | Label | Assembly | Claim | Backs |
|---|---|---|---|---|---|---|
| finding | `f-<hex>` → `F-<8 hex>` | `<!--fx:ID-->` | finding anchor | strip | no | no |
| citation | `c-<hex>` → `C-<8 hex>` | `<!--cite:ID-->` | citation anchor | weave `[^N]` | yes | yes |
| proof | `p-<hex>` → `P-<8 hex>` | `<!--proof:ID-->` | proof anchor | weave `[^P<n>]` | no | yes |
| gap (Part 4) | `G<n>` → `G-<8 hex>` | `<!--gap:ID-->` | gap anchor | strip | no | no |

- *Id*, *Token* — spelling: `Token`, `idAt`, `Kind`, `tokenLenAt`, `IDPattern`. An id no row
  claims reads as a finding, as today; Part 6 deletes that default (every id then has a letter).
- *Label* — the noun every message uses. Part 1 keeps the three texts byte-identical; Part 4
  collapses them (§III.4). The per-kind labels compose with the registered "anchor"
  (`terms.json:566`) and are not registered; "finding anchor" is registered only because it
  redirects the GATED "finding-marker" (`:603-605`), which `Label` prints today and stops printing
  in Part 4.
- *Assembly* — `anchor.StripAssembled` (`report/docs.go:223`) strips every `strip` row.
- *Claim* — a cited sentence is the report's unit of claim: `Scan`/`Count`/`Index`/
  `segmentLabels` (`claimcount.go:327`), `mintbudget.go:42`, `scorecard.go:475`.
- *Backs* — the marker is evidence standing behind its sentence, as `GapJSON.Backing`'s doc says
  ("the citation and proof anchors in its location"). Part 1: `gapBacking`'s `about_ref` test
  (`viewjson.go:1033`, `Kind(ref) != "finding"`) becomes `Backs(ref)` — identical, since an
  unclaimed id reads as a finding. Part 4: the location loop reads it too (§III.4).

Every lifecycle reader walks every row alike: `ProtectedAnchorIDs` (edit guard, `droppedMarker`,
`ReopenedAnchors`, `RemoveAnchorAt`'s husk rule, retire), `BareAnchorIDs`,
`AnchorsTransitUnchanged`, `settleAbuttingAnchor`, Part 3's auto-place. `ProtectedAnchorIDs` walks
in row order, so its output order is unchanged.

**Per-kind census** (`git grep -nE 'anchor\.Kind\(|anchor\.Kind[A-Z]|HasPrefix\([^,]+, *"[fcp]-"\)|"[fcp]-"|case "(fx|cite|proof)"|Kind(Finding|Citation|Proof)\b' 8798488d -- 'plugins/frank-exchange-of-views/tools/*.go' ':!*_test.go' ':!*.pb.go'` → 19 lines):
`anchor.go:29,31,77,79,132` spelling → table (Part 1); `anchor.go:90,92,94` `Label` → table, texts
kept (Part 1), collapsed (Part 4); `retire.go:268` the one lifecycle branch, DELETED (Part 4);
`claimcount.go:327`, `scorecard.go:475` *Claim*; `viewjson.go:1017,1046` *Label*, `:1033` *Backs*;
`citationid.go:25,26,33,271`, `findingid.go:45` minters → `NewID` (Part 6). Outside the 19:
`shapes.go:167-178` `CitationAnchor` (only a citation is a source) stays; `consistency.go:208-216`
→ F-g.

Changes:
- [MODIFY] `anchor/anchor.go`: `Token`, `idAt`, `Kind`, `Label`, `tokenLenAt` read the table.
  [NEW] `IDPattern()`, the one id matcher, and `Backs(id)`. [NEW] `Sentences(text) [][2]int`, the
  one splitter: `.`/`!`/`?` and newline are boundaries, runs collapse, any `<!--…-->` comment is
  skipped — exactly what `splitClaims` and `sentenceBoundaries` do today (Part 2 changes the rule).
  `RemoveAnchorAt` reads the boundaries with the token still in place, then removes it — the same
  answer under today's rule; Part 2's needs it ("One. <c>. Two." keeps its boundary, and
  "**One.** <c>. Two." its closer).
- [NEW] `anchortext.Attach(doc, id, quote) (string, error)` = `InsertAnchor(doc, quote,
  anchor.Token(id))` with today's rule. It returns `anchortext`'s sentinels (`ErrMisQuote`,
  `ErrInFence`) and **maps nothing**: each verb keeps its own refusal text, so no message moves
  (N6). It lives in `anchortext` because `anchortext` imports `anchor`.
- [MODIFY] The four placers call `Attach` and stop building tokens: `finding.go:104,119`,
  `cite.go:166,171`, `prove.go:147,152`, `verify.go:341,349`. `render.go:124` keeps
  `InsertAnchor`: replay reproduces the write's choice.
- [MODIFY] `claimcount.go`: the four token regexes (`:213,:338,:345,:355`) and
  `Finding/Citation/ProofAnchorIDs` (`:361-373`, one caller, `ProtectedAnchorIDs:396`) become one
  table walk; `StripAnchors`, `HasProse`, `segmentLabels` read the table.
- [MODIFY] `flags/shapes.go:44` `anchorShape` and `:133` `anchorToken` read `IDPattern()`/`Token`
  — same bytes accepted in Part 1; from Part 4 they accept a gap anchor with no further edit (audit G2).
- [DELETE] Duplicate readers: `report/assemble.go:46,55`, `report/proofs.go:16`,
  `lens/verify.go:388`, `bluedoc.go:391 anyAnchorToken`. Splitters: `render.go:224`,
  `bluedoc.go:340 sentenceAround` (and its strip-then-scan workaround), `claimcount.go:79`.
- [KEEP] `annotationLen:63` skipping any HTML comment (Part 8 adds the table's token);
  `InsertAnchor`'s first-match `StopAtParagraph` rule and its `extractQuote` fallback (R-2).
- [MODIFY] Help strings that spell a token build it from the table (`shapes.go:163,178`,
  `verify.go:96`, `reproduce.go:63`, `seat/verbs.go:274,281`); bytes unchanged, no golden moves.
- [MODIFY] Comments that list the anchor kinds (§III.4 kind-list sweep: `anchor.go:6,43,120`,
  `claimcount.go:173,184,212,244,376`, `bluedoc.go:383`, `edit.go:27`, `citationid.go:26,264`,
  `evidenceview.go:13`) say "every kind in the table"; `claimcount.go:351,358,390,393`,
  `assemble.go:52` and `edit.go:25` go or follow with their code (S13).
- `settleAbuttingAnchor` and the reopened computation narrow from `<!--[a-z]+:…-->` to the
  table. `settleAbuttingAnchor` also runs on every replayed splice (`render.go:100`); what keeps
  renders unchanged is the archive census — no HTML comment but the three marker kinds — and
  `TestEveryArchivedRunRenders` proves it.

**Test that fails if Part 1 is wrong** — [NEW] `migrate/everyrender_test.go`
`TestEveryArchivedRunRenders`: for each tarball, migrate, then `RenderFromRecord`; a render error
fails with the run named. Golden `migrate/testdata/archived_renders.golden`, one line per run:
`<run> <rendered|no-base> <raw> <normalized> <skeleton> <markers per kind>` — *raw* the sha256 of
the render; *normalized* with each marker token replaced by `⟨kind#n⟩` (n its first-appearance
ordinal within its kind); *skeleton* the sha256 of the render with each gap token taken out as a
retire takes it out (`RemoveAnchorAt`, husk and stray terminator included; there is none before
Part 4), then normalized, every line holding no prose (`claimcount.HasProse`) dropped and
whitespace runs folded to one space. Test and golden land in Part 1's **first commit, against unrefactored code**; every
later Part 1 commit leaves the golden byte-identical. The 12 universe runs are compared by hand
(§V.2).

**Consumer census (Part 1).**
- S1 → 16 code lines (`claimcount.go:338,345,355`; `cite.go:166`; `prove.go:147`;
  `finding.go:104`; `reproduce.go:63`; `verify.go:96,341`; `seat/verbs.go:274,281`;
  `shapes.go:163,178`; `assemble.go:46,55`; `proofs.go:16`): all CHANGE as above.
- S10 → 25 lines: Part 1 CHANGES the marker readers (`bluedoc.go:391`,
  `claimcount.go:213,327,338,345,355`, `verify.go:388`, `shapes.go:44,133`, `assemble.go:46,55`,
  `proofs.go:16`, `scorecard.go:475`); Part 4 deletes `retire.go:268`; Part 6 the id readers.
- `git grep -nE "HasPrefix\((line|s)\[i:\], \"<!--\"\)" -- '…/tools/*.go' ':!*_test.go'` → 3:
  `anchortext.go:64` KEEP (R-2); `claimcount.go:82`, `render.go:227` DELETE into `Sentences`.
- Kind-agnostic callers (`git grep -noE '(anchor|claimcount|anchortext)\.[A-Z][A-Za-z]+\(' …`): all
  KEEP their call; the behaviour behind it is the table's.
- Tests: `claimcount/anchors_test.go:22-58` (the three per-kind functions) becomes one table case
  over `ProtectedAnchorIDs`; `cli/citelockdown_test.go:80` filters `ProtectedAnchorIDs` by
  `Kind == "citation"` (N12). `TestSkipRunAgreesWithTheProtectionSweep` DELETE: one reader, nothing
  to agree.
- gray-area: not a consumer (Census 3).

### III.2 Part 2 — one sentence rule (bug fix; no record change)

**Value.** Every reader of the sentence holding a marker reads a whole sentence. Today "3.5%",
"e.g.," and "verify_91.py" end one, and so does every line break, so a soft-wrapped sentence is one
fragment per line: the claim unit, `reopened` and the husk rule read fragments, and auto-place and
the gap location would too (R-16, R-21). About +30 production lines (+80 rule and block reader;
−25 `assemble.go`'s `sentenceEnd`, −15 `insideFence`'s loop, −10 `claimcount`'s per-line loop,
`lead` and `listMarkerRe`). **Measured +98** (+347 −249): the block reader carries every R-25 opener
condition, each pinned by its own row. After Parts 1 and 2 the running S6 total is +50, so the ≥ 320
target needs about −370 from Parts 3–6.

**The rule** (`anchor.Sentences`; R-16, R-21, R-22, R-25; its anchor clauses read off the bare
shapes `claimcount/bare_test.go` pins — the author's reading, disclosed):
1. **Blocks**, by markdown's own rules (CommonMark, and GFM for tables). A newline ends a sentence
   only where it ends a block: the next line is blank, or it opens a block —
   - a **list item**: `-`, `*` or `+`, or `<n>.` or `<n>)`, then whitespace or the end of the line.
     Only `1.` or `1)` starts a list that interrupts a paragraph; any number opens after a blank
     line or as the next item of an ordered list already open (same delimiter). Otherwise the line
     is a soft wrap ("The steps are⏎2. compute it." is one paragraph; "- one⏎2. two", one item);
   - a **heading**: one to six `#` followed by a space or the end of the line ("#552" is text);
   - a **table row**: a `|` line whose next line is a delimiter row (cells of `-`, each optionally
     `:`-aligned, between `|`), then every following line until a blank line or another opener; a
     `|` line with no delimiter row under it is text;
   - a **block quote**: `>`. A `>` line after a `>` line continues the quote: its `>` (and one space
     after it) is stripped and reads as whitespace to rule 2, so the newline is a soft wrap ("> One.⏎>
     Two." is one block of two sentences). A `>`-only line ends the paragraph inside the quote, as a
     blank line does. Openers inside a quote are read on the line after its `>` ("> - y" opens a list
     item, "> # h" a heading). A line with no `>` after a quoted paragraph line continues that
     paragraph (markdown's lazy continuation);
   - a **fence** (```` ``` ```` or `~~~`) or a **footnote definition** (`[^…]:`).

   A heading, a table row and a fence's closing line end their own block. Every other newline is a
   soft wrap, and a soft wrap is whitespace. A block's first sentence starts after its `>` markers
   and its list marker. A fence, opener to closer, is one unit, split nowhere: no marker stands in
   one (`Attach` refuses `ErrInFence`) and `Scan` skips it, as today. An HTML comment — every anchor
   token — is opaque.
2. **Terminators.** A terminator run (`.` `!` `?`) with the closers right after it — `"` `'` `”`
   `’` `»` `)` `]` `}` and markdown's `*` `_` `` ` `` — ends a sentence when followed by the end of
   its block; by whitespace (a soft wrap included) and then an uppercase letter, an opener (`"` `'`
   `“` `‘` `«` `(` `[` `{` `*` `_` `` ` ``) or an anchor token; or, with no closer, by an anchor
   token flush.
3. **Ownership.** The run and its closers belong to the sentence they end. An anchor run flush
   after a closer belongs to it too, and rule 2 is applied again after the run (R-22): "**Wet.**<c>
   Next." → "**Wet.**<c>" + "Next."; "**Wet.**<c> and more." stays one; "Seen (p. 3.)<c>" ends
   attached. An anchor after whitespace, or flush after a bare terminator, opens the next sentence
   ("One. <c>", "One.<c> Two.": the cut shapes, bare). All 18 bare cases and every other
   `claimcount` test pass unchanged (measured, §V.2).
4. No abbreviation list (a hand-kept table stops covering its case): a digit after a terminator
   joins, an initial splits, an opener after any terminator splits — counted in A-4.

**Block openers, swept** over the 21 renders: 425 paragraphs, 322 list items, 248 headings, 1
fence; no table (the template's risk matrix is one; two pre-#709 runs' edits hold 29 rows), block
quote or footnote definition. A list item interrupts a paragraph 32 times, never one numbered other
than 1. The shapes R-25 narrows occur 0 times: a line opening with `#` not followed by a space, a
line opening with `|`, a line opening with `>`, and an ordered item other than 1 after a paragraph
line (15 such items follow an item of their own ordered list, and open). Not openers: `---` (6,
each between blank lines, no prose), an indented line (23, all list continuations), a line opening
with `<` (3, anchor tokens). The harness's block reader restated to rule 1
(`~/.claude/scratch/markers-plan/r7/cc`, §V.2) reproduces every number below byte for byte — claim
`Count` 143, `Index`, the bare set, the boundary table, b3's three gaps, the seven fragments
(`r7/before` and `r7/after` compare equal) — and gives the `>`, `#`, `|` and ordered-item rows of
`TestSentences` the answers written there.

**Boundaries by cause** — every newline ending a line of a prose block and every terminator run in
one, over the 21 renders (§V.2 `cls6`). "Today" is `splitClaims`: every newline and every run ends.

| Class | n | Rule |
|---|---|---|
| newline ending a block: blank line next 511, end of report 21, a list item next 214, a fence 1; a heading, table row, block quote or footnote definition next 0 | 747 | ends |
| soft wrap after a terminator, then a capital (36) or an anchor (1) | 37 | ends |
| soft wrap after a terminator, then a digit (m8 ×2: "computation).⏎2^90 ≡ 64") | 2 | joins |
| soft wrap, no terminator (b3 174 of the 214 wraps; m8 10; b6, m6, m12, m13 7 each; m14 2) | 175 | joins |
| terminator: end of block | 574 | ends |
| terminator: whitespace, then a capital (3 after an initial, A-4) | 960 | ends |
| terminator: whitespace, then an opener (`[` 17, `(` 14, `"` 6, `*` 6; 5 mid-sentence, A-4) | 43 | ends |
| terminator: an anchor flush after a bare terminator (the cut shape) | 4 | ends |
| terminator: closer, anchor run, then a capital (R-22) | 0 | ends |
| terminator: no whitespace after (decimals, file names, domains, "e.g.,") | 167 | joins |
| terminator: whitespace, then a digit (35 join two sentences; "ca. 240", "no. 2", "p. 33") | 38 | joins |
| terminator: whitespace, then lowercase ("Prime Curios! entry", "90! mod", "“is 91 prime?” is") | 16 | joins |
| terminator: whitespace, then a symbol (`≡` 3, `✓` 2, `—` 2, `×` 1) | 8 | joins |
| an ordered-list marker `1.` (today a terminator, the "1" a segment of its own) | 96 | outside the block's body |

**Tests.**
- [NEW] `anchor/sentences_test.go` `TestSentences`: "Prices rose 3.5% in 2024. Then fell." → two;
  "Some bases (e.g., 3) fool it.", "Run scripts/verify_91.py to check.", "Sources include
  numbers.education and OEIS.", "√91 ≈ 9.54, so we test 2–9.", "see e.g.⏎foo bar." → one each;
  "He said “Stop.” Then he left." → "He said “Stop.”" + "Then he left."; "Dr. Smith agrees." and
  "Smith et al. (2020) found X." split after "Dr." and "al." (stated); "methods⏎with zero
  disagreement among them. The" → the first sentence spans the wrap; "First.⏎Second." → two; "A
  list:⏎- one⏎- two" → three, items without markers; "Intro.⏎```⏎A. B.⏎```⏎After." → the fence
  one unit; rule 3: "**Wet.**<c> Next sentence." → "**Wet.**<c>" + "Next sentence.", "**Wet.**<c>
  and more." → one, "Seen (p. 3.)<c> Next." → two, "One. <c>. Two." and "One.<c> Two." → the cut
  shapes. Rule 1's openers (R-25), one row per condition: "Text⏎# Heading⏎More" and "Text⏎#⏎More" →
  three; "Fixed in⏎#552 after review. Next." → "Fixed in #552 after review." + "Next."; "Values
  differ⏎| sharply here. Then more." → two, the first spanning the wrap; "| a | b |⏎| --- | --- |⏎|
  c | d |" → three table rows, no prose in the delimiter row; "The steps are⏎2. compute it. Done." →
  "The steps are 2. compute it." + "Done."; "Steps:⏎1. compute it.⏎2. check it." → three, items
  without markers; "- one⏎2. two" → one item; "Intro.⏎⏎3. starts a list." → two, the item without
  its marker; "> a long⏎> sentence." → one sentence; the auditor's three — "> One.⏎> Two." → two
  sentences (the continuation `>` read as whitespace), "> a⏎>⏎> b" → two blocks,
  "> x:⏎> - y" → "x:" + the list item "y"; and "> # h⏎> body text." → a heading and one sentence,
  "Text before.⏎> quoted. Here.⏎lazy line." → "Text before." + "quoted." + "Here. lazy line." (a
  quote interrupts a paragraph; a lazy line continues the quoted one). Deleting any one opener
  condition fails its row.
- [NEW] `TestSentencesKeepsTheCorpusSentencesWhole`: each row names its run and gap or anchor,
  copies from the record the block its quote or anchor stood in, newlines kept, and holds the whole
  sentence as a literal, so the test shares no code with the splitter. Rows: the seven terminator
  fragments — quadratic `R2-2` ("…50.8% of v1's prose…"), universe2 `G3`, `G14` ("√91 ≈ 9.54,
  so…"), m12 `G3` ("etc.) reinforces"), b5 `G6` ("e.g.,"), b6 `G5` ("verify_91.py"), b7 `G3`
  ("numbers.education"); b3's three gaps at their mint — `G1` its whole list item
  ("**Deterministic Miller–Rabin**, with the witness set {2, 3, 5, 7} — … result: composite.";
  today "proof rather than a probabilistic result: composite"), `G2` "This is settled with
  certainty — … with zero disagreement among them." (today "with zero disagreement among them"),
  `G3` "Not distinguished here." under both rules (its quote's last sentence ends at a `?`, R-4);
  the 19 anchors in a soft-wrapped sentence — b3's `f-8be4bb71`, `p-6844d749`, `p-bd21ef70`,
  `p-e2c14426`, `p-d6107488` with `f-811a6ab9`, `p-a897c923`, `p-8f09d8a4`, `p-c5213636`,
  `c-2e118c62` with `c-ecf4fa0a`, `p-a30013b5`, `c-5eabefc8`, `c-8cde8cc2`, `c-38bea55a`,
  `c-f733bc64`, `c-91b91a86`, `f-950a7234`, and m14's `c-1b559e10`; one residue row, m8's
  `f-7af4335b`, whose sentence the digit clause joins to the one before it: "Base 2: Compute 2^90
  mod 91 (verified by reproducible computation).⏎  2^90 ≡ 64 (mod 91) ≢ 1 (mod 91)<fx>." — the
  literal copied from the record, its list marker outside and its continuation indent kept (A-4).
  Deleting the block clause fails the b3 and m14 rows; the terminator clause, the seven.
- [NEW] `TestEverySentenceReaderReadsOneSentence` (`cli`): fragment and wrap cases driven through
  every reader; each row fails on a reader that splits at an internal period or a soft wrap. Part
  2: on "Prices rose 3.5% in 2024<!--cite:c-1-->. Then fell.", and on the same sentence wrapped
  after "3.5%", an edit "3.5%" → "1.5%" (its `old` excludes the anchor) records `c-1` in
  `reopened`; "91 = 7 × 13. ✓<!--cite:c-2-->" counts as a claim and `retire` does not offer `c-2`
  as bare (the m13 shape); retiring the bare anchor in "**One.** <!--cite:c-3-->. Two." leaves
  "**One.** Two." (`RemoveAnchorAt`; today "**One. Two.", the closing `**` lost — a live
  corruption this part fixes); retiring the bare anchor of an emptied ordered item "1.
  <!--cite:c-4-->" leaves "1." (the whole-line test reads the line, list marker included — unchanged
  from base); the risk-matrix cell of a problem reading "The committee chair said
  “Stop.” Then everyone left the room." is "The committee chair said “Stop.”" (`report.concise`;
  today the whole field). Part 3 adds: "Prices rose 3.5% in 2024" → "Prices rose 1.5% in 2024",
  the cite left out of `new`, is refused naming that sentence (the fragment "5% in 2024" occurs in
  `new`, and `AutoPlace` would re-place the cite on the rewritten claim). Part 4 adds: a gap minted
  on the sentence reads `marked` on the whole of it, lists the "3.5%" → "1.5%" edit in the work
  list's `edited_since`, and lists `c-1` as backing.
- [NEW] `anchortext` `TestInsideFenceReadsTheBlockReader`: a quote whose end falls on a fence's
  opener or closer line is refused `ErrInFence` by `InsertAnchor` (outside at base), and one ending
  on the line after the closer places; deleting `insideFence`'s call to `anchor.Blocks` fails it.

**Readers** (J1's sweep: every reader of a marker's sentence), each on `anchor.Sentences`:

| Reader | Today | From |
|---|---|---|
| `claimcount.Scan` → `Count`, `Index`, `BareAnchorIDs`: the claim unit, retire's bare set, `ReopenedAnchors`' bare skip, Part 4's `gone` | `splitClaims`, per line | Part 1 |
| `RemoveAnchorAt`'s in-line husk (replayed `retire_anchors`, live retire): the token's sentence, emptied when it holds no prose or other anchor; the cut stays on the token's line, and a line left with neither goes whole, as today — `HasProse` over the whole line, list marker included, so an emptied ordered item ("1.") stays and an emptied bullet goes. Unchanged behaviour; reading the block's body there instead is optional and outside this plan | `sentenceBoundaries`, per line | Part 1 |
| `ReopenedAnchors` → `BlueEdit.reopened` → `edited_since`, `show evidence`'s `reopened` | `sentenceAround`, line-bounded | Part 1 |
| `report.concise` (risk-matrix cell: first sentence ending past `minSentence`, within the limit) | `sentenceEnd` | Part 2 |
| `bluedoc.AutoPlace` — `PlanSplice`, the accept's carry (A-3), the F-c verbatim rule | — | Part 3 |
| gap `location`, `gapBacking`'s loop, the F-c fallback, the `marked` check of `TestEveryArchivedGapHasANamedLocation` | — | Part 4 |

`PassageAround` takes the marker's offset and returns its section passage; it reads no sentence.
Census: `git grep -nE 'claimcount\.(Scan|Count|Index|BareAnchorIDs)\b|RemoveAnchorAt\(|ReopenedAnchors\(|concise\(' -- '…/tools/*.go' ':!*_test.go' | grep -v 'func '`
→ 9: `mintbudget.go:42`, `countclaims.go:47`, `claimindex.go:59`, `retire.go:220`,
`bluedoc.go:311`, `edit.go:158`, `render.go:149`, `assemble.go:672,895` — all KEEP their call.

**One block reader.** `anchor.Blocks(text)` — each block's kind, span and body start — is rule 1:
`Sentences` splits inside it, `Scan` reads its kinds for its exclusions, and `Paragraphs` (the mint
budget's blank-line unit) reads its line classifier, `claimcount`'s `fenceLine` and `footnoteDef`
moving into it, its count unchanged. `anchortext.insideFence` (`anchortext.go:330`, behind
`ErrInFence`) reads it too, and that changes replay: `InsertAnchor` runs on every render and
Part 2 has no epoch step, so an offset on a fence's own opener or closer line, which reads as
outside today, reads as inside, and a recorded anchor standing on a fence line stops rendering
after the upgrade. No archived or corpus run holds one (one fence in the 21 renders, no anchor on
its lines). Gates: `TestEveryArchivedRunRenders`, loud for archived runs, and the §V.2 render
identity of the 12 universe runs. Residue, stated: a run outside both, live across the upgrade,
with an anchor on a fence line fails to render, loudly, at every verb that renders. Fence readers
that stay: the assembly's section splitter (`assemble.go:351,457`) reads blue's markdown during
assembly for `## ` headings outside fences; `md.go:28` renders the assembled markdown; and
`capture.go:566-568` strips code before counting footnotes in the run's documents. None reads a
sentence.

**The claim index** (R-21). `Scan` walks blocks, not lines: a segment is one sentence and may span
lines, and `Segment.Line` — `Index`'s `occurrences[].line` — is the line its sentence starts on.
Fence, heading and footnote-definition blocks stay out, and a heading still sets the context.
`afterProse` reads the block, not the line; `attached` reads a soft wrap as whitespace
("One.**⏎<c>" stays gutted); `lead` and `listMarkerRe` go, the list marker being outside the body.
Readers (`git grep -nE 'claimcount\.(Scan|Index|LabelOccurrences)|Segment\{|\.Line\b' 8798488d -- '…/tools/*.go'
':!*_test.go' ':!*.pb.go'` → 16, each read for what it means): `claimcount.go:152,305` build the
segment and its occurrence (`Scan`'s callers are `Count`, `Index`, `BareAnchorIDs`); `claimindex.go:58-59`
serve `claim-index`, whose Long (`:31-35`) gains "— line is where the claim's sentence starts"
(regenerated: `agents/blue-researcher.md:506`, `agents/blue-synthesizer.md:429`, the `manual`
golden `:1151`); the other 12 are other concepts — an avenue's line (`blue/avenue.go:111,221`,
`record/avenue.go:171,173,395`, `assemble.go:850`, `scorecard.go:536`, `seatprobe/build.go:214`,
`view.go:792`), a voice tell's (`ingest.go:150`), seatprobe's report (`devcmd/seatprobe/main.go:463,473`).
No test pins a claim-index line (`git grep -nE '"line": ?[0-9]|Line: *[0-9]' 8798488d --
'…/tools'` → 0); `claimcount_test.go`'s "a line break bounds the unit without punctuation" still
counts 2 and is renamed for what it shows. The package doc (`claimcount.go:43-47`: "bounded by … OR
a line break, so a cited list emits one claim per line"), `Segment` (`:113-122`) and `Occurrence`
(`:272-276`) are rewritten; `splitClaims`' (`:72-78`) goes with it.

**Measured** (§V.2), on the 21 renderable corpus reports, today → terminator clause alone → final
rule: `Count` 143 → 143 → 143, equal in every run; segments carrying prose 2,278 → 2,052 → 1,779;
cited claims whose sentence changes — → 40 → 48; `Index` identical except in b3, where seven
occurrences report their sentence's first line (`c-2e118c62` 177→174, `c-ecf4fa0a` 178→174,
`c-5eabefc8` 198→196, `c-8cde8cc2` 204→200, `c-38bea55a` 211→209, `c-f733bc64` 216→214,
`c-91b91a86` 227→223), labels, headings and order unchanged; `BareAnchorIDs` moves only in m13, by
the terminator clause (two live anchors after "= 7 × 13. ✓" read bare today, offering a live claim
to `retire`). Anchors in a soft-wrapped sentence: 20 — R-21's 19 and m8's residue row. Patched into
a scratch copy of the base module (the terminator clause at all four sites, the block clause in
`claimcount` and `bluedoc`), `claimcount`, `bluedoc`, `reportproj`, `report`, `record/...` and
`cli/...` fail no test the unpatched copy passes (10 fail on both, each needing the repository
root), with R-25's block reader too (`~/.claude/scratch/markers-plan/r7/tree-blk`: the same 10); the
module's 57 repository-root tests are unmeasured there, so §V.1 item 1 runs Part 2 in the
repository. Goldens that move (`count-claims`, `claim-index`, the assembled risk matrix) are read
line by line.

**Renders.** The rule reaches replay only through a `retire_anchors` row. `SELECT (SELECT count(*)
FROM retire), (SELECT count(*) FROM retire_anchors), (SELECT count(*) FROM base_ingest)` over the
28 migrated corpus records → `retire_anchors` 0 in every run (5 retires: quadratic, b3, b9 ×2,
m7; 21 runs with a base). `TestEveryArchivedRunRenders` and the universe comparison stay
byte-identical; a future run with a row moves the raw digest loudly.

### III.3 Part 3 — the tool re-places a marker whose sentence survives (behaviour)

**Value.** It removes the carry burden where carrying is mechanical (≈18% of carries) and turns
the other refusals from "you dropped `<!--fx:f-…-->`" into "the sentence that held it was *'…'*;
put it where that claim now is". About +70 lines. It ships before Part 4 because gap markers add
markers to carry. **Measured +101** (+240 −139): the token goes inside the occurrence `LocateOnce`
counted at the place it held in its sentence (walked in step, after the anchors that preceded it in
its run), because at the occurrence's end 11 of 32 corpus placements left the place blue kept;
`AutoPlace` returns the text alone, and the transit check runs it and is the one refusal; the refusal
names the replacement's nearest sentence by near-match's `Tokenize`/`Jaccard`, moved to
`anchortext`. S8 over the 211 carries (3 not carried in their old-epoch record): 29 placed
byte-for-byte as blue placed them, 0 wrong, 172 refused naming the sentence, 4 bare refused as
bare, every placement in a run with a base (21) replayed byte for byte from the recorded
replacement, and 3 residue where blue moved the anchor off its surviving text (quadratic
`f-33643203`, universe2 `f-36e5440b`, m6 `c-4a145c1b`, the last refused as changing nothing). After
Parts 1–3 the running S6 total is +151, so the ≥ 320 target needs about −471 from Parts 4–6.

- [NEW] `anchortext.LocateOnce(doc, quote, scope) (start, end int, err error)` — **the one
  write-time matcher**. It takes `locate`'s first `CrossParagraphs` match and refuses, in this
  order, `ErrMisQuote` (none), `ErrAmbiguous` (a second match after it — whether or not the first
  crosses a blank line, so a quote that also occurs inside one paragraph hears "occurs more than
  once", never "may not cross"), under `StopAtParagraph` `ErrCrossesParagraph` (the one match
  crosses a blank line) and `ErrSplitsWord` (`spanBoundaryOK`, moved here from `bluedoc`).
  `bluedoc.LocateUnique` becomes `LocateOnce(…, CrossParagraphs)` plus its texts — the answers it
  gives today, so replay's splices (`render.go:100`) do not move. Every write-time locate of a quote
  calls it: AutoPlace here, whose placements have no history; the placers through `Attach` from
  Part 4 (R-11).
- [NEW] `bluedoc.AutoPlace(span, new string) (string, []string)`, in `bluedoc` so `migrate` can
  call it (Part 4): for each marker in `span` absent from `new`, take its sentence with
  `anchor.Sentences`; where the marker-stripped sentence occurs verbatim in `new`'s marker-stripped
  text and `LocateOnce(new, sentence, StopAtParagraph)` accepts it, the token goes in at the `end`
  that call returned — the one occurrence it counted, inside text this edit writes. So `new` =
  "Costs rose sharply. Costs rose." holds "Costs rose" twice and is refused, never placed on the
  first (H2). It returns the new text and the ids it could not place. It reads the marker's
  sentence within `span`, the text the edit replaces, not the document's: a fragment edit
  "fell sharply in Q1<c>" → "rose, then fell sharply in Q1" re-places the cite, and the document's
  sentence has changed ("Revenue rose, then fell sharply in Q1"), so `ReopenedAnchors`, which reads
  the document's sentence, records the cite in `reopened` and red re-reads it — as for any
  fragment edit that leaves a marker in place.
- [MODIFY] `reportproj.PlanSplice` (`plan.go:59`): after `LocateUniqueReplacing` fixes the span
  (`settleAbuttingAnchor` included), `AutoPlace(report[start:end], new)` runs, then the transit
  check (`:64`); the literal branch does the same (`:72-73`). PlanSplice returns the replacement it
  applied, and `blue edit` records it as `BlueEdit.new`, so replay needs no new logic. An unplaced
  id refuses the edit with the marker, its sentence and the replacement's nearest sentence; the
  text (`bluedoc.go:200-237`) resolves under `deadpath_test.go`. `droppedMarker` (`edit.go:236`)
  reads PlanSplice's result. `ValidateProposal`'s transit check (`bluedoc.go:282`) reads
  `AutoPlace`'s output too, so red is refused exactly where blue's edit would be.
- [MODIFY] `droppedMarker`'s comment (`edit.go:270`, "(finding OR citation)") becomes kind-free.
- Auto-place is not a kind property: the `gap` kind gets it in Part 4 with no edit here.

**Test.** `planedit_splice_test.go`, through `blue edit`: verbatim sentence once in `new` → placed,
byte-equal to the hand-carried form; twice → refused; rewritten → refused naming the sentence;
"Costs rose sharply. Costs rose." dropping the marker of "Costs rose." → refused, nothing recorded;
the fragment edit above → placed, with the cite in `BlueEdit.reopened`.
`TestEveryArchivedRunRenders` stays byte-identical (`LocateUnique` re-expressed). S8 (§V.2) drives
`AutoPlace` with blue's 211 real carries.

**Consumer census.** `git grep -nE 'AnchorsTransitUnchanged|droppedMarker|settleAbuttingAnchor|ErrAnchorIntroduced|ReopenedAnchors' -- '…/tools/*.go' ':!*_test.go'` →
`edit.go:158,222,273`, `bluedoc.go:129,196,237,282,309`, `citationid.go:533`, `retire.go:143-266`.
retire KEEPs `anchorsExiting`/`leftByTheCut`. `PlanSplice` callers: `edit.go:232`, `mint.go:163`.
The `error_catalogue` golden moves.

### III.4 Part 4 — gaps attach by marker; the replay retires (record change, epoch 20 → 21)

**Value.** About −170 production lines (−267 replay, −85 retire hold, +20 location state, +120
placement step, +10 `Attach` and its texts, +10 mint placement, +12 proposal run and the compare
reading it, +5 retries) and ≈ −245 test lines. S5, S7, S12, S13. Depends on Parts 1, 2 and 3.
Ids stay `G<n>` until Part 6.

**Placement.**
- [MODIFY] kinds table: the `gap` row (`G<n>`, `<!--gap:ID-->`, "gap anchor", strip, no claim, does
  not back). Its markers live the one lifecycle (R-6).
- [MODIFY] `Attach` (A-2): `LocateOnce(doc, quote, StopAtParagraph)`, then `ErrInFence` at its
  `end`, then the token at that `end`. It stops calling `InsertAnchor` and never falls back to
  `extractQuote` (H2). The occurrence `LocateOnce` counted is paragraph-bounded, so it is also
  `InsertAnchor`'s first match; the whole quote matched, so `InsertAnchor`'s fallback is not
  reached; replay therefore puts the token at the same offset. From Part 4, **`Attach` refuses
  everything `LocateUnique` refuses**, plus a quote that crosses a blank line or ends in a fence.
  `extractQuote` is deleted (A-5): the migration step rewrites any archived location that needed it.
- [NEW] **Placement refusal texts**, one per new sentinel, shared by finding, cite, prove, verify
  and mint, each prefixed with its verb: *ambiguous* — "the quote occurs more than once in the
  report, and an anchor needs one place: quote the sentence with the text before it in its
  paragraph, so the quote occurs once — a quote may not cross a blank line. A sentence that stands
  alone as its paragraph and repeats verbatim elsewhere cannot be anchored" (mint adds "name its
  section with --about-kind section"); *crosses a paragraph* — "the quote's one match in the
  report runs across a blank line, and an anchor sits in one passage: quote text inside one
  paragraph" (`LocateOnce` reports a second match first, so this text reaches only a quote with
  one match); *splits a word* — `LocateUnique`'s words. None advises one
  edit per site (R2). Each verb keeps its own not-found and fence texts; mint takes finding's.
- [MODIFY] `lens/mint.go:116-127`: `--quote` is placed with `Attach` against the current render
  (replacing `LocateUnique`). After every check passes, the mint appends `Mint` and then
  `Anchor{id: gap id, location: quote}` — the existing `anchor` arm renders it (A-1).
- [MODIFY] **Retries anchor the stored location** (H2): a retry under the same key that finds its
  `Anchor` missing (`AnchorEventExists`) appends `Anchor{id, the location its claim stored}` —
  never the retry's own `--quote` — once `Attach` accepts that location against the current
  render; otherwise it returns the refusal and the claim stays unplaced. This is finding's path
  (`finding.go:81-84` appends the retry's unchecked `--quote` today) and mint's from this part;
  cite, prove and verify's from Part 5.
- [MODIFY] **The proposal carries the marker run** (R-13, A-3). `record.Proposal(run, gapID,
  report)` (`estoppel.go:143`) takes the render its caller validates against — `edit.go` renders
  once, before reading the proposal. It locates the location with `LocateOnce(…, CrossParagraphs)`,
  the matcher `edit` uses, and reads the marker run abutting the located end — after any trailing
  punctuation, where `settleAbuttingAnchor` reads it (`anchor.SkipRun`). When that run holds the
  gap's anchor: `old` is the location as the render holds it — its bytes from the located start
  through the run, then the location's own trailing punctuation, if any ("S." against the render's
  "S<G1><c1>." gives "S<G1><c1>."). The edit locates that to the span through the run
  (`normalizeQuote` drops the trailing "." and `settleAbuttingAnchor` extends over the run), the
  "." stays after the span as for any whole-sentence edit, and the seam tidy takes the doubled
  terminator `fix_new` brings. `old` holds the render's bytes, which can differ from the stored
  location in whitespace (`mint --quote` locates after whitespace collapse; b3 `G1` stored "proven
  to⏎correctly" where the render holds "proven to⏎  correctly"). `new` is `fix_new` with each of
  the run's gap anchors it does not already carry inserted at its content end by `Attach`'s insert
  step — after its last content character, no second locate (A-3). So `new` carries each gap
  anchor of the run exactly once: a second lens whose `--quote` abutted G1 had to quote
  `<!--gap:G1-->` and carry it into its `--new` (`settleAbuttingAnchor`), and G1 is not added
  again. The run's other anchors are not added: the accept is an edit, so `PlanSplice`'s
  `AutoPlace` re-places each that `new` lacks and whose sentence survives verbatim in `fix_new`,
  and its transit check refuses the rest, naming the anchor and its sentence, inside
  `edit.go:137-140`'s stale text — blue then edits by hand (R-17). Otherwise
  the raw pair: it applies as today where no marker abuts the location, and is refused as stale
  where one does. With G2 minted on the same sentence (`S<!--gap:G2--><!--gap:G1-->.`),
  `--accept --answers G1` carries both gap anchors to `fix_new`'s end. The board's
  `fix_old`/`fix_new` (`viewjson.go:447`) come from the same function, so a seat typing the
  board's pair types what `--accept` sends to `PlanSplice`.
- [MODIFY] **The typed pair is compared with `record.Proposal`'s output** (R-24, L1).
  `ProposalAppliedVerbatim(run, gapID, report, old, new)` (`estoppel.go:153`) calls
  `Proposal(run, gapID, report)` and compares the typed `old` and `new` with its pair, byte for
  byte; its own query of the mint row goes. `edit.go:184` passes the render it validated against
  and the pair the seat typed, never `PlanSplice`'s applied replacement. One function serves the
  board's pair, `--accept`'s pair and this comparison, so over one render a seat typing the
  board's pair records `applied_verbatim` wherever `--accept` would. `EstoppelConflict` compares
  `Visible` text and `appliedVerbatim` reads the recorded flag, so estoppel needs no change.
  `fix_new` is stored as red wrote it; assembly's `fixProposal` (`assemble.go:1185`) reads it
  unchanged. `mint --new` is still validated by `ValidateProposal` against the pre-mint render.
- **Typed-pair sweep (L1)** — every comparison of a seat-typed pair against the record: `git grep
  -nE 'ProposalAppliedVerbatim\(|record\.Proposal\(|func Proposal\(|proposalFor\(|someProposal\(|aVerbatimGap\(\)|GetFixNew\(\)|FixOld\b'
  -- '…/tools/*.go' ':!*.pb.go' | grep -vE '\.go:[0-9]+:\s*//'` → 22 at `8798488d`. CHANGE (11):
  `estoppel.go:143,153`, `edit.go:78,184`, `viewjson.go:447` (above); `queries_parity_test.go:145,148`
  pass a render; the release fuzz's typed arm — `proposalFor` (`fuzz_test.go:4655,4666`) reads
  `record.Proposal` over the render its drive (`:4512`) now takes first, and `mintEstopped`
  (`:1227`) finds the applied text in the render (`:1256`) on `Visible` text, both sides, as
  `EstoppelConflict` does, since that text now carries the gap's anchor and any anchor `AutoPlace`
  re-placed. DELETE (2): `someProposal` (`:4669,4676`), no caller. KEEP (9), red's text as
  written, compared with no typed pair: `estoppel.go:62` (estoppel's prescription, read through
  `Visible`), `mint.go:249`, `redvoice.go:57`, `assemble.go:1194`, `changes.go:149`,
  `viewjson.go:161` (the field), `blueedit_test.go:365`, `redvoice_test.go:94`, `fuzz_test.go:447`.
- [MODIFY] The mint page's `--quote` help asks for one sentence and says the board shows that
  sentence and its section (R-4); `flags.DescQuote` is unchanged for other verbs.
- [MODIFY] `record.go:786-796`: the comment says an `Anchor` places one marker of any kind at
  its quote's end, keyed on `id` so a retry writes one event; `:791` → "record: anchor requires
  id (the id of the act whose marker this places)"; `:795` → "record: anchor requires location
  (the quote the marker sits at the end of)". The `Anchor` message comment (`record.proto:998`)
  and `EVENT_TYPE_ANCHOR`'s `means` (`:240`): "a marker placed in the report".

**Location — one of three named states per quote gap** (R-7).
- [NEW] `location_state` on `GapJSON`, `WorkGapJSON` and `EstoppedJSON`:

  | State | When | `location` | `passage` | `backing` |
  |---|---|---|---|---|
  | `marked` | the gap's token stands at prose in the render | the sentence holding it (`anchor.Sentences`), marker-stripped | its section passage | the *Backs* markers in that sentence |
  | `gone` | the token stands at no prose: bare (`claimcount.BareAnchorIDs` — blue cut its sentence), absent (retired), or never placed (a migrated mint whose quote did not place) | the minted text | empty | empty |
  | `unrendered` | the report does not render: no base, or a render error | the minted text | empty | empty |

  An `about` gap carries no state; its backing is its `about_ref` when `Backs(about_ref)`.
  `edited_since` is read from the record in every state.
- Readers: `show board`, the work view and the estoppel list carry the state; in `unrendered`,
  `BoardJSON.anomalies` names the cause. `show report --anchor <gap id>`: `marked` and bare `gone`
  read the window at the marker; a retired one names the retire (`AnchorRetiredAt`,
  `seat/verbs.go:760-772`); for an id with no marker in the report, `anchor/window.go:93-94` says
  so and names where each kind resolves — `show board` for a gap, whose minted text it shows when
  the quote never placed (N8). The ledger, `changes`, the docket and assembly read the minted text,
  labelled as minted (unchanged).
- [MODIFY] `viewjson.go`: location (`:441`, `:983-988`), passage (`:442`, `:988`; `PassageAround`
  takes the marker's offset) and backing (`:453`, `:1032`) come from the render. `gapBacking`'s
  location loop keeps only ids with `Backs`: a gap's own marker, other gaps' markers and finding
  markers are not evidence and leave the list (audit G2). The render failure at `:353-364` becomes
  `unrendered` plus an anomaly.
- [MODIFY] `edited_since` (`:761`): the BlueEdits since the reader's epoch whose recorded `reopened`
  names the gap. `ReopenedAnchors` computes it at the write over the marker's whole sentence
  (`anchor.Sentences`, Part 2), so a fragment edit inside the sentence is listed. Its bare skip
  applies to every kind, so the edit that cuts the sentence is not listed; the gap reads `gone`, and
  **that is an answer** (R-9). The element keeps its shape; the `GapEdit` type moves into
  `viewjson.go`.
- [DELETE] the retire hold (R-6): `record/estoppel.go:276-325` (50 lines); `cli/blue/retire.go:254-282`
  `takeable` (29), the call at `:119-121` → `body.Anchors = exiting`, `kept` at `:83`, `:129`,
  `retireResult.Kept` (`:292`) and its `Human()` loop (`:301-303`) — 85 lines; the hold sentence in
  `Retire.anchors`' comment (`record.proto:1414-1415`). Every bare marker an edit left exits with
  its retired claim, whatever its kind.
- [DELETE] `record/gapedit.go` (194), `currentLoc`/`mintedIfMoved`/`orEmptyEdits`/`editsSince`
  (37), `minted_location` and `location_edits` (~22), the `GapEdits` loads (~14).

**Migration placement step (F-c)** — epoch 20 → 21. One step in `migrate.Replay`, after
`remap.apply` and before `Append` (`replay.go:181,194`). It reads bodies already in the
destination's spelling, so from Part 6 on it compares respelled `old`/`new` with a respelled render
(**respell before locating**), and the `Anchor` it emits carries the remapped gap id with no edit
to `remap`. It adds only what the source lacks: an `Anchor` only for an id the source stream never
anchored (a pre-scan, as the correction pairing at `replay.go:115` is) and only the first per id;
a gap token only where `old` or `new` lacks it. A run written by this part's binary therefore
migrates with its own anchors and nothing added. It inserts only gap tokens into archived report
text; Append's checks stay on; an exemption that turns out to be needed sits beside its check.
- *Mint.* At each replayed mint with a quote: render the destination record and `Attach` the
  stored quote (the live rule, uniqueness included). Placed → `Mint`, then `Anchor`; a
  correction's replacement re-carries its gap id and adds none (the archive holds none, §II). Not
  placed → `Mint` alone (`gone`).
- *BlueEdit.* At each replayed edit, render, locate as replay does, then rewrite so it renders and
  carries every placed gap token (N2):
  - **(a) abutting** — the run of markers abutting the located span's end holds a gap token `old`
    lacks (the commonest shape: an edit quoting the gap's own sentence). `old`'s marker run after
    its last content character (often empty) is replaced by the render's run, so
    `settleAbuttingAnchor` extends the span over it.
  - **(b) literal** — an exact-span edit whose literal `old` no longer occurs because gap tokens now
    stand inside it: locate `old` with gap tokens skipped and set `old` to the render's bytes over
    that range.
  - **carry** — every gap token in the (rewritten) span and absent from `new` goes into `new`:
    `bluedoc.AutoPlace` where it places it; else, where `new` holds prose, right after the last
    content character of `new`'s first sentence by `anchor.Sentences` (F-c); where `new` holds none
    — empty, or markers only (the cut idiom: an edit down to the bare anchor, then `retire`) — the
    token joins `new`'s marker run (bare → `gone`).
  - Then the gap ids `ReopenedAnchors(before, after)` reports join `BlueEdit.reopened`, as a set.
  Every rewrite inserts gap tokens only: with gap tokens removed, `old` and `new` are the archived
  bytes.
- Runs with no base place nothing: their quote gaps read `unrendered`.
- An archived edit that no longer locates because a kept husk stands in its span is a refusal,
  named by run (A-4).
- Cost: one render per mint and per edit; the PR reports the time on the largest archived run.

**Tests that fail if Part 4 is wrong.**
- `TestEveryArchivedRunRenders` (R-14): a gap token adds a placeholder, keeps any husk a historical
  retire left beside it (none in the archive: no `retire_anchors` row, R-20; shape (e) exercises
  the clause), and is stepped over at a seam — so raw and normalized digests move. The
  **skeleton**, which takes each gap token out as a retire would (§III.1), and the fx/cite/proof
  counts do not: each renderable run's skeleton equals Part 1's. The golden gains a gap-token count
  and a bare-gap count per run (a historical cut never retired its gap's marker, so those gaps read
  `gone`), read in review.
- [NEW] `TestGapTranslationRewritesEveryShape` (`migrate`): one synthetic source stream per shape —
  (a) an edit quoting the gap's sentence without its token, (a) with a cite already in the run,
  (b) an exact-span edit across the token, (c) an in-span drop, (d) a cut to empty, (e) the cut
  idiom — a markers-only replacement, then its retire: the gap token stands bare beside the husk
  the retire kept, and the skeleton equals that of the same stream without the mint. Each renders; each token lands where the rule says;
  marker-stripped `old`/`new` are the source bytes. (f) A run written by the Part 4 binary (mint,
  `--accept`) migrates with one `Anchor` per gap and a byte-identical render; deleting the
  pre-scan fails it.
  Over the archive, the golden lists rewrites per run and shape, and the test fails if shape (a)
  counts zero across all tarballs (it is the commonest; zero means the rule never ran).
- [NEW] `TestGapAnchorSitsAtTheQuoteEnd` (audit G3): `mint --quote` of a fragment ending mid-sentence and
  of a whole sentence → the token right after the quote's last content character, before its
  punctuation; `location` is the sentence holding it. Re-rendering from the record gives the same
  bytes (replay). The b9 `G4` migration places it at the offset `Attach` gives on the destination
  render.
- [NEW] `TestMintRefusesWhatLocateUniqueRefused` (audit G2, H2): through `lens mint` and `lens finding`,
  a quote absent, present twice in two paragraphs, splitting a word → refused for the cause
  `LocateUnique` gives, in the placement text; crossing a blank line, inside a fence → refused; the
  #552 shape — a quote crossing a blank line that holds a quoted phrase occurring earlier — →
  refused; a quote whose first match crosses a blank line and which also occurs inside one
  paragraph → refused as ambiguous, not with the crossing text. Nothing appended for any.
- [NEW] `TestAttachPlacesWhereReplayDoes` (`anchortext`, fuzz): over generated documents and
  quotes, whenever `Attach` succeeds, `InsertAnchor` on the same document returns the same bytes.
  Deleting `LocateOnce`'s paragraph check fails it (the #552 shape again).
- [NEW] `TestRetryAnchorsTheStoredLocation` (`cli`): a finding and a mint whose `Anchor` append was
  lost, retried with a different `--quote`, append the stored location; with the report changed so
  the stored location occurs twice, the retry refuses and appends nothing.
- [NEW] `TestAcceptCarriesTheGapAnchor` (`cli`, N3, R-17): mint with `--quote S --new N`, then
  `edit --accept` applies; the render holds the gap token once, at N's end; `applied_verbatim` is
  recorded; the gap reads `marked` on N. With a cite abutting S: where N keeps S verbatim and adds a
  sentence, `AutoPlace` puts the cite back at S's end inside N and the accept applies with
  `applied_verbatim`; where N rewrites S, the accept is refused as stale naming the cite and S,
  nothing is recorded, and blue's hand edit (`fix_new` plus the cite, `--answers G1`) applies.
  Deleting `PlanSplice`'s `AutoPlace` call fails the first; a cite inserted at N's end — a fourth
  path — fails the second. A third mints G2 on S after G1 (H3): `edit --accept --answers G1` applies
  with `applied_verbatim`, both anchors stand at N's end, and G2 reads `marked` with the accept in
  its `edited_since`; reading the location alone fails it at `settleAbuttingAnchor`. A fourth types
  the board's `fix_old`/`fix_new` as a plain edit with `--answers G1` and records `applied_verbatim`;
  a mint quoting N afterwards is refused by estoppel. A fifth mints G1 with `--new`, then G2 with
  `--quote "S<!--gap:G1-->" --new "N2<!--gap:G1-->"`: `--accept --answers G2` applies with each
  anchor once, and typing the board's pair for G2 applies too; adding every run anchor regardless
  fails it (G1 twice). A sixth is b3 `G1`'s shape (L1, R-24): the report holds a list item wrapped
  with a continuation indent ("… proven to⏎  correctly decide …"), `mint --quote` quotes it without
  the indent ("proven to⏎correctly decide") with `--new`, and typing the board's pair as a plain
  edit with `--answers G1` records `applied_verbatim`, as `--accept` on the same report does;
  comparing with the stored location, stripped of markers or not, fails it. Every row runs twice,
  with the location stored without and with its terminator ("S", "S."). The
  `--accept` cases in `cli/blueedit_test.go` and `cli/editnoop_test.go` (8 uses) stay green; the
  typed seeds (`seedProposalApplied`, `blueedit_test.go:439`, and the counter-edit at `:465`)
  quote the gap's anchor as `show report` prints it, the seed typing the board's pair. The release
  fuzz's ZERO gates on verified proposals (`fuzz_test.go:4046`), verbatim applications (`:4052`)
  and estoppel refusals (`:4072`) fail if accept or proposal stop applying; a ZERO gate on typed
  verbatim applications joins them — `applied_verbatim` without `accepted`, tallied beside
  `:3317` — so a typed arm that stops recording fails even while `--accept` records.
- `queries_parity_test.go:145-149`: `ProposalAppliedVerbatim` with a render — the exact pair
  matches, the near-application (a trailing space) does not.
- [NEW] `TestEveryArchivedGapHasANamedLocation`: every quote gap has a state, every `about` gap
  none and a non-empty `about_ref`; in the 9 renderable runs every `marked` gap's token occurs
  once with prose before it in its sentence (by `anchor.Sentences`, whose rule the literal tests of
  Part 2 pin); the 7 no-base runs read `unrendered`; `gone` and
  fallback-placed gaps are listed by name.
- [NEW] `TestEditedSinceSeesTheSentence` (`cli`): a fragment edit inside the gap's sentence whose
  `old` excludes the marker is listed; an edit to an identical sentence elsewhere is not; a cut is
  not listed and the gap reads `gone`.
- [NEW] `TestRetireTakesABareMarkerOfEveryKind` (`cli/retireanchor_test.go`, replacing
  `TestRetireTakesARedMarkerOnlyOnceItsGapIsClosed`, `:120-195`): finding, citation, proof and gap
  markers, each cut to bare and retired by `--anchor <id>` (so `anchorShape` takes `G3`) — a
  crediting gap open for the finding row, the gap itself open for the gap row. Each retire names
  the marker; `show report --anchor` names the retire; the gap reads `gone`. A hold fails a row.
- [NEW] `gapBacking`: a gap whose sentence holds a cite, a proof, a finding marker and another
  gap's marker lists the cite and the proof; an about-gap whose `about_ref` is a gap lists nothing.
- [NEW] `TestWorkTeachesEveryLocationState` (`difftest`): the rendered `work` and `board` help
  name each `location_state` value, read from its Go constants, and the engaged prompt golden
  (`tests/simulator/testdata/prompt-red-lens-evidence-engaged.golden`) carries the constant's
  `gone` sentence verbatim — the wording, not only the word.
- Migrated: b5 `G5` (blind spot, 7 edits) has a non-empty `edited_since`; b9 `G4` (wrong
  occurrence) is `marked` on its minted sentence (`TestB8ArchiveMigratesWithNoRefusal`,
  `archive_test.go:270`, is the template).
- Assembly holds no `<!--gap:`; `claimcount.Count` is unchanged by a mint; `spanVoiceTells` over a
  replacement carrying `<!--gap:G3-->` reports nothing.

**Consumer census (Part 4).**
- Replay: `git grep -nwE 'GapEdits|CurrentLocation|relocate|PassageAround|mintedIfMoved|editsSince|location_edits|minted_location|edited_since' -- '…/tools/*.go' ':!*_test.go'`
  → 33. Callers only `viewjson.go` 348, 441-443, 941, 986-988, 1071; definitions in `gapedit.go`,
  `passage.go:40`, `viewjson.go:1702-1729`.
- Readers of a gap's location: `git grep -nE 'm\."location"|"location", "fix_new"|"location", "about_kind"|Mint\.GetLocation\(\)|\b(m|mint)\.GetLocation\(\)|\bg\.Location\b|gj\.Location|Location: g\.|currentLoc\(|EstoppelConflict\(' -- '…/tools/*.go' ':!*_test.go'` → 26:
  `viewjson.go:418,441,442,453,955,983,987,988,1032,1068,1093` CHANGE; `viewjson.go:1702`,
  `gapedit.go:63,65` DELETE; `recordsql/views.go:380` KEEP (minted text); `view/view.go:348`,
  `view/changes.go:149`, `report/assemble.go:997,1194`, `estoppel.go:73`,
  `duplicatescreen.go:67`, `nearmatch.go:99`, `mint.go:204` KEEP (minted text, labelled);
  `seatprobe/build.go:150` KEEP (fixture); `estoppel.go:146` CHANGE (`Proposal`, above) and `:160`
  DELETE (`ProposalAppliedVerbatim` reads `Proposal`, R-24), and with them `viewjson.go:447` (the
  board's `fix_old`/`fix_new`) and `edit.go:78,119,184` (render first; the typed pair compared).
- **Write-time placement (H2)**: `git grep -nE '\b(InsertAnchor|LocateUnique|LocateUniqueReplacing|LocateSpanUniqueScoped|LocateSpanScoped|LocateSpan|LocateLiteral|locateEnd|locate|ValidateProposal|PlanSplice)\(' -- '…/tools/*.go' ':!*_test.go' | grep -vE '\.go:[0-9]+:\s*//' | grep -vE ':\s*func '`
  → 25: 7 inside `anchortext` (the matcher; `LocateOnce` joins them); 5 placers (`cite.go:171`,
  `prove.go:152`, `finding.go:119`, `verify.go:349`, `mint.go:127`) → `Attach`; 8 replacement
  validators (`bluedoc.go:49,70,278`, `edit.go:232`, `mint.go:155,163`, `plan.go:60,72`) →
  `LocateUnique` (now `LocateOnce`) or `LocateLiteral`, answers unchanged; 3 replay
  (`render.go:94,100,124`) KEEP; 2 other readers (`passage.go:44`, which takes the marker's offset
  from this part; `cite.go:271`, the source cache). `Anchor` appends: `git grep -nE
  'recordpb\.Anchor\{' -- '…/tools/*.go' ':!*_test.go'` → 2 (`finding.go:84` the retry, `:151`):
  the retry takes the stored-location rule. The proposal's gap anchors go in by `Attach`'s insert step
  at `fix_new`'s content end; the F-c fallback is a migration translation only.
- `report_op` readers: `queries.go:231`, `render.go:274`, `evidenceview.go:320`. Anchor arm
  unchanged.
- `Anchor` readers: `git grep -nE 'FROM "anchor"|\*recordpb\.Anchor\b|EVENT_TYPE_ANCHOR' -- '…/tools/*.go' ':!*_test.go' ':!*.pb.go'` → 6:
  `findinglabel.go:80` (mint retry), `record.go:427` (key from `id`, KEEP), `record.go:786` (text),
  `consistency.go:218` (F-g: a quote mint joins the finding set), `seatprobe/seatprobe.go:154`
  (`verbOfEvent` maps `ANCHOR` to `finding`: DELETE; every `Anchor` follows a body event that
  already names its verb; the seatprobe goldens counting `finding` from `Anchor` events move and
  are read — N15), `views.go:720`.
- Kind-agnostic readers: `ProtectedAnchorIDs`/`BareAnchorIDs` include it; `RemoveAnchorAt` keeps a
  line holding it; `ReopenedAnchors` reports it; `anchorsExiting`/`leftByTheCut` let it exit;
  `Scan`/`Count`/`Index` ignore it (*Claim*); `gapBacking` excludes it (*Backs*);
  `anchor.StripAssembled` strips it; `StripAnchors`/`HasProse`/`selector.visibleText`/`Visible`
  strip it; `SkipRun`/`tidySeam` step over it; `ReadAround` serves `show report --anchor G3`;
  `mintbudget` counts it as neither claim nor paragraph. **`show evidence`'s `reopened`**
  (`citationid.go:525`) lists every id any edit reopened, finding ids included today; gap ids join
  them unchanged (N7).
- Id-shape readers the gap kind reaches before Part 6 — the boundary walk is §III.9.
- JSON: `location_state` added, `minted_location`/`location_edits` removed — viewjson tests, the
  generated `agents/*.md` OUTPUT lines (11 agents), the `manual` golden.
- **Agent-facing text: one lifecycle for every kind** (R-6). Each surface teaches it once — the
  tool places an anchor at the end of its quote; an edit carries it; the tool re-places one whose
  sentence the edit kept verbatim; an edit that drops or types one is refused; retiring its claim
  is the one way it leaves — and names the gap kind where it lists kinds.
  - `skills/research-protocol/SKILL.md:24-29`: "every anchor, of every kind" plus the lifecycle
    sentence; `:44-45` gains "`gap:` where red holds a gap open"; `:17` loses "IMMORTAL".
  - `skills/adversarial-audit/SKILL.md:42` ("location MUST name the section heading and quote the
    challenged sentence") → "a gap is minted on ONE sentence, quoted exactly; the tool anchors it
    there and the board shows it with its section"; `seat/selector.go:162`'s comment follows.
  - `anchor.Label` (`anchor.go:88-98`): "<label> <id>" plus one lifecycle sentence for every kind —
    "(the tool places it and every edit carries it; its claim leaves by `edit` down to the bare
    anchor, then `retire`, which takes the anchor out with it — never by a raw edit)".
  - `ErrAnchorIntroduced` (`bluedoc.go:237`) → "anchors are placed by the tool, never typed into a
    replacement"; the comment at `bluedoc.go:193` likewise.
  - `retire` help (`seat/help/retire.md:9`) → "takes every bare anchor out, whatever its kind";
    the finding-hold sentences go. `retire.go:137` → "an anchor id of any kind"; `retire.go:302`
    goes with `Kept`. Regenerated: `agents/blue-researcher.md:936,950`,
    `agents/blue-synthesizer.md:859,873`, the `manual` golden (`:1581`).
  - `edit` help: `seat/help/edit.md:19`, `edit.go:202,298` → kind-free. Regenerated:
    `agents/blue-researcher.md:569,596`, `agents/blue-synthesizer.md:492,519`.
  - `scripts/agentgen/src/blue-synthesizer.md:16` → "cannot drop an anchor"; `:18` KEEPs.
  - `cite.go:308` ("an invisible immortal anchor") → "an invisible anchor"; `consistency.go:515`
    → "has no anchor event" (F-g).
  - **Docs (audit G6)**: `docs/finding-markers.md:1,5,21,55-62,70-73,78-96,100,138,179` state the one
    lifecycle for every kind — no "immortal", no "the marker remains, the content is gone"; blue
    removes an anchor of any kind only by `retire`, and red's assent is the gap's closure, which
    reads the gap's own anchor. `docs/citations-and-bibliography.md:111` "The anchor is immortal"
    → "The anchor leaves only by `retire`". `docs/propagation-and-anchoring.md` gains the gap kind
    where it lists kinds; `docs/record-flow.md:90` by the kind-list sweep below.
  - **Cut = answer (R-9)** — carriers of "an empty `edited_since` means unanswered" (`git grep -nE
    'edited_since|EditedSince|empty list is the answer|WHETHER BLUE MOVED|edits that touched' --
    plugins/frank-exchange-of-views scripts/agentgen/src ':!*_test.go' ':!*.pb.go' ':!*/agents/*'
    ':!*/agent-memory/*' ':!*manual_runs_every_help_page_live.golden'` → 10 lines): `debate.js:956`
    (the engaged clause) becomes "YOUR WORK LIST SAYS WHETHER BLUE MOVED: each gap carries its
    location state and the edits that changed its sentence since you last sat. A `marked` gap with
    no edits is unanswered … `gone`: the gap's anchor is not in the report — blue cut its sentence,
    or, in a migrated run, its quote never placed; that is not silence — judge the report as it
    stands." (acts, no verb; under the 12300 ceiling). The `gone` sentence is the one Go constant's
    (below), word for word, and asserts no single cause. Its golden
    (`prompt-red-lens-evidence-engaged.golden:19`) regenerates; `viewjson.go:752-760`
    (`EditedSince`'s doc) carries the same sentence with `changes` named for a cut's edit;
    `viewjson.go:121-126` and `gapedit.go:49,136` go with their code; `viewjson.go:707` (a usage
    count) KEEPs. [NEW] one Go constant naming the three states and what each asks of a reader,
    rendered on the `work` and `board` help pages (`seat/verbs.go:275,277`). `blueanswered.go:61-72`
    reads `BlueEdit.answers`; a cut that names its gap is ANSWERED there. [MODIFY] its engaged arm
    (`:68-69`) reads the gap's location state, a field `WorkGapState` carries from Part 4 on: a
    `gone` gap gets "gap G… is open and its anchor is no longer in the report — judge whether the
    report as it stands meets your acceptance check, then close it or say what still fails", never
    "NO edit answers it". Test: `TestBlueAnsweredReadsAGoneGapAsMoved` (a cut by an `edit` with no
    `--answers`; deleting the state branch fails it).
  - Comments that describe the replay or an immortal marker (`git grep -nE
    'gap_edit|CurrentLocation|LOCATION-DRIFT' -- '…/tools/*.go' ':!*_test.go' ':!*/gapedit.go'` → 6:
    4 in `viewjson.go` go with their code): `anchortext.go:84` (the `gap_edit` view,
    CurrentLocation), `blueanswered.go:14` (`gap_edit` LOCATION-DRIFT), and the package docs
    `anchortext.go:1` and `anchor.go:1` ("immortal anchors") describe the one marker model.
  - Kind lists that gain the gap: `seat/verbs.go:274,281,559`; `shapes.go:163`;
    `anchor/window.go:94`; `terms.json:567` ("a citation, a proof, a finding or a gap"; then
    `vocabdoc`). No registry entry is added (N14, §III.1 *Label*).
  - **Kind-list sweep (J2, S14)**: `git grep -niE '\b(finding|citation|cite|proof)( anchor| marker)?s?(,? or |,? and |, )(a |an |the )?(finding|citation|cite|proof|gap)s?( anchor| marker)|\b(of|in) (either|both)( of the)?( two| three)? (anchor |marker )?(class|kind)(es)?\b|\b(either|both|two|three) ([a-z-]+ )?(anchor|marker) (class|kind)(es|s)?\b|\b(two|three) classes\b|\bneither (a |an |the )?(finding|citation|cite|proof|gap)s?( anchor| marker)?\b' -- plugins/frank-exchange-of-views scripts/agentgen/src ':!*_test.go' ':!*.pb.go' ':!*testdata*' ':!*/agent-memory/*' ':!plugins/frank-exchange-of-views/agents/*'`
    → 29 at `8798488d` (6 more in generated `agents/*.md`). "neither …" matches the list's first
    half, so a list wrapped after its "nor" is still found (`claimcount.go:393`); one interposed word
    ("BOTH immortal anchor classes") is allowed. Property: a list of anchor kinds names
    all four or says "every kind". Text (5): `docs/record-flow.md:90` ("of either class" → "of any
    kind"), `window.go:94`, `edit.go:202`, `seat/help/edit.md:19`, `seat/help/retire.md:9` (above).
    Comments (16): Part 1's (§III.1). With the hold (2): `retire.go:257`, `estoppel.go:279`.
    KEEP (6), lists of meaning: `viewjson.go:103,739,838,1249` (*Backs*), `debate.js:960` and
    `assemble.go:3` (not anchor kinds); `window.go:94`'s `show evidence` clause stays a 7th.
  - Sweeps: the lifecycle sweep `git grep -nE "red's (finding )?(anchor|marker)|(finding|red's) (anchor|marker) is red|marker is red's|which were kept|every gap crediting|citation or proof anchor out|placed by .finding. and .cite.|finding anchor or (a )?citation anchor|finding anchors preserved|red's finding anchors|anchor kept|c-…, p-… or f-…" -- plugins/frank-exchange-of-views scripts/agentgen/src ':!*_test.go' ':!*.pb.go' ':!*testdata*'`
    → 32 lines / 14 files; target the 7 provenance lines (`finding-markers.md:113`, `edit.go:227`,
    `lens/verify.go:319`, `citationid.go:122`, `reportproj/splice.go:33`,
    `agentgen/src/blue-synthesizer.md:18`, `agents/blue-synthesizer.md:32`). S13 → 0.
- **Harness readers** (N10): `releasegate/fuzz/fuzz_test.go:4311 fuzzAnchorRe` and `:4316
  anchorsLeftFor` read `anchor.IDPattern()`/`Token`, so `retire --anchor` is driven with gap
  anchors. The edit drive (`:1820-1845`) and proposal drive (`:1150-1170`) quote their span with
  the marker run abutting it, as `show report` prints it, and carry it into `--new`: with a gap
  marker after "rising over time", both are otherwise refused on every run after a proposal
  mint. The typed arm reads the pair the board serves: its drive (`:4512`) renders first and
  `proposalFor` reads `record.Proposal` over that render, so it types what `--accept` sends and
  `ProposalAppliedVerbatim` compares against (R-24); `someProposal` goes and `mintEstopped` reads
  `Visible` text (the typed-pair sweep above). The ZERO gates above, the typed arm's included,
  catch a drive this leaves silent.
- Tests deleted: `cli/gaplocation_test.go` (89); `record/locatorclass_test.go` 121-169, 228-281 and
  the ~81-86 case; `passage_test.go` to offsets; `cli/board_test.go:258-290`.

### III.5 Part 5 — one placement event (record change, epoch 21 → 22)

**Value.** `report_op` drops from four insert arms to one; every marker is placed by an `Anchor`
event, so "where is X placed" has one table. About +35 lines (−14 arms, +15 appends with retry,
+4 consistency, +30 translation).

- [MODIFY] `cite`, `prove` and verify's corroboration append `Anchor{id, location}` after their
  event, as `finding` and `mint` do; a retry under the same key completes a missing `Anchor` by the
  stored-location rule (§III.4). A same-sitting correction re-carries the id and appends no
  `Anchor`.
- [MODIFY] `recordsql/views.go:706` `report_op`: one insert arm from `anchor`, one remove arm; its
  comment follows. `recordsql/testdata/schema.sql` regenerates.
- [MODIFY] `record.proto`: `Anchor` keeps `{id, location}` and reserves `finding_id`,
  `finding_key`, `label`, `text` (no archived payload carries them). **UNIQUE on `anchor.id`.**
- [MODIFY] `consistency.go:205-220` (F-g): one set — every `Finding`, `Mint`, `Cite`, `Proof` and
  labelled `Verify` with a quote expects an `Anchor`.
- **Translation 21 → 22**, in Part 4's placement step: after each archived `Cite`/`Proof` with
  location and id, and each labelled `Verify` with a claim, it emits `Anchor{id, location}` — one
  per placement event, for an id the source stream never anchored; a correction's replacement
  re-carries its id and adds none. Every insert arm is one `insertMut` (`render.go:116-128`), which
  skips while the token stands in the text, so the `anchor` arm renders each emitted `Anchor`
  exactly as the old arm rendered its event. Report-op order is event order, so every placement
  keeps its stream position. A second placement event for one id (a crash retry, or an id placed
  again after a retire) would emit a second `Anchor`, which UNIQUE refuses loudly; the 28 corpus
  records hold none (0 repeated `cite.label`, `proof.proof_id`, `verify.label` or `anchor.id`;
  N11, §II), and `TestEveryArchivedRunMigrates` reports one if a run brings it.

**Tests.** `TestEveryArchivedRunRenders`: the **raw** digest of every run is unchanged — a
placement missing its `Anchor`, or at another stream position, moves it.
`TestRetireTakesABareMarkerOfEveryKind` and the crash-retry tests (`citecrash_test.go`) pass with
the `Anchor` pair; `TestRetryAnchorsTheStoredLocation` gains a cite row. Shape (f) gains a cite: a
run written by the Part 5 binary migrates with one `Anchor` per id (UNIQUE refuses a second).
`consistency` reports a cite with no `Anchor` (deleting the widened set fails it).

**Consumer census.** `report_op` readers as Part 4. `Anchor` readers: the six above;
`AnchorEventExists` now serves four retries. `seatprobe`: nothing further (Part 4 stopped counting
a verb for `Anchor`).

### III.6 Part 6 — one id (record change, epoch 22 → 23)

**Value.** Seven id functions become one (S4); the finding's double identity goes (its one
id-to-label join left with `FindingMarkerHold` in Part 4). About −135 production lines: −≈210
deleted (`findingid.go` 46, `NextFindingLabel` + doc ≈22, `NewCitationID`/`NewProofID` ≈31,
`record.go:617-629` 9, `shapes.go` ≈34, `FindingLabelAlt` 13, the three count queries ≈32 with
docs, the `Kind` default), +≈15 `NewID` and `FindingRef`, +≈60 translation.

**Ids.**
- [NEW] `record.NewID(kind)`: the kind's letter, `-`, 8 hex from `crypto/rand`. Every minting verb
  calls it before `Append`; the assignment at `record.go:628` goes.
- UNIQUE where an id is minted: `finding.id` (new), beside `anchor.id` (Part 5), `mint.gap_id` and
  `motion.motion_id`. Re-carried columns stay plain (`cite.label`, `proof.proof_id`,
  `avenue.avenue_id`, every reference).
- [MODIFY] `record.proto`: `Finding.finding_id` and `label` collapse to `id` (old numbers reserved).
- [NEW] `record.FindingRef(id, area)` prints `F-7cdcd115 (adversary)`; the area is the finding
  event's `seat_id` without `red-lens-`. JSON carries `area`.

**Readers that take the one shape** (from `anchor.IDPattern()`/`Kind`): `flags/shapes.go:40,52,167,186`
(G, M, citation, Q — table-backed) and `:49,55` with `FindingLabel()`/`SHA()` DELETE;
`scorecard.go:475`; `capture.go:1124`; `report/md.go:57 idToken` (`P\d+` stays: it is the proof
footnote number); `reportvoice/tells.go:80,85,87,89`; `seatprobe/build.go:399`; `refs.go:72,109`;
`verify/verify.go:224`. A reader that would match **nothing** on a new id gets a unit test feeding
one — the test fails on the old pattern: tells, `idToken`, `capture.anchorID`, `scorecard`,
`seatprobe.mintedID`.

**Finding-label readers** (`git grep -nE 'GetFindingId\(\)|FindingId\b|f\."label"|f\."finding_id"|GetLabel\(\)' -- '…/tools/*.go' ':!*_test.go' ':!*.pb.go'` → 33):
read `id` and, where they print, call `FindingRef`: `verify/verify.go:228,528`; `view/view.go:300,308`;
`findinglabel.go:65`; `record.go:624-628` (DELETE), `:1081-1082`; `consistency.go:205,208`;
`viewjson.go:499,1321,1322,1336`; `report/assemble.go:964,965,1277,1283`; `lens/finding.go:132`.
Citation and verify sites KEEP: `citationid.go:85,101,166`, `evidenceview.go:298,324,326,393`,
`blue/cite.go:110`, `lens/verify.go:378`, `available.go:388`, `redvoice.go:72`.

**Translation 22 → 23** — `migrate/remap.go` changes from a field-name switch to **three
hand-kept lists** covering every string field of every event body at this part's schema (R-8,
R-19). Id map: `f-`/`c-`/`p-` keep their hex and take the capital; `G<n>`/`Q<n>`/`M<n>` take the
letter, `-` and the first 8 hex of `sha256(SourceHash ‖ old id)`, so two migrations of one source
agree wherever they are written; `<area>-F<n>` maps to its finding's id. The old id is the
archive's own spelling: an `R<r>-<n>` gap id hashes straight to its `G-` id and an `L<k>-F<n>`
label maps straight to its finding's `F-` id, so `remap.go:125,131` (the `G%d` and `%s-F%d`
Sprintfs S4 counts) go, and every id is formatted by `record.NewID`'s formatter.
- **Every string field**: an exact marker token is respelled where it stands.
- **Source-data list** (68) — nothing else changes: report text (`BaseIngest.text`,
  `BlueEdit.old/new`, `Mint.location`, `Mint.fix_new`, every `location`, `Verify.claim`,
  `Retire.claim`), subject sources (`Cite.url/title/ocr_quote`, `Verify.url/title`,
  `Proof.script`, `Reproduce.recorded_output/observed_output`, `Retire.superseded_by`), keys,
  hashes, dates, enumerated values and identities.
- **Id list** (34) — a whole value or list element equal to a mapped id is replaced:
  `Mint.gap_id/about_ref/supersedes/found_by/distinct_from`, `Close.gap_id/successor`,
  `Closing.gap_id`, `Regrade.gap_id`, `SpotCheck.ids`, `Finding.id/about_ref`,
  `Observe.label`, `Anchor.id`, `Cite.label`, `Verify.anchor/label`,
  `Proof.proof_id/answers/cites`, `Avenue.avenue_id`, `BlueEdit.answers/reopened`,
  `Retire.anchors`, `ManifestRow.gap_id`, `Log.estopped_by`, `Motion.motion_id`,
  `GradeMotion.gap_id`, `DocketMotion.gap_id`, `AvenueMotion.avenue_id`, `MotionRule.motion_id`,
  `MotionAppeal.motion_id`, `Dispatch.gap_ids`, `Gate.migration_admitted_gap_ids`.
- **Seat-argument list** (50; R-8, R-12) — a mapped id at word boundaries is replaced, as the bare
  id (A-4). Every
  field holding a seat's own wording: `Mint.problem/required_fix/acceptance_check/definition/
  neighbor/distinguisher/mint_reason`, `ClassNew.definition/neighbor/distinguisher`,
  `Close.prose/anchor_tool/anchor_target`, `Closing.text`, `Regrade.basis`, `SpotCheck.reason`,
  `Finding.text`, `Observe.text/observation`, `Cite.text`, `Verify.text`, `Proof.text/drift`,
  `Reproduce.note`, `Avenue.line/hypothesis/method/reason`, `AvenueReview.reason`,
  `BlueEdit.text`, `Revision.text`, `Retire.reason`, `ManifestRow.row`, `Log.text`,
  `Motion.basis/relief`, `DocketRuling.principle/tension/review_flag/settled/reopens_on`,
  `MotionRule.opinion`, `MotionAppeal.reason`, `Outcome.prose/verdict_why`, `Position.text`,
  `Halt.opinion`, `Certify.statement`, `Declare.holding`, `Correction.why`.
- **Census (H1)** at `8798488d`, lists in `~/.claude/scratch/markers-plan/r4/{ids,arg,rest}.txt`:
  - string fields: `awk '/^message /{m=$2} /^extend /{m=""} m!="" && /^ *(optional|repeated) string [a-z_0-9]+ = /{print m"."$3}' plugins/frank-exchange-of-views/tools/internal/record/recordpb/record.proto`
    → 168; 11 are not event bodies (`Sql`, `SqlCheck`, `Event`, `TelemetryLine`); the 157 are 35
    id + 50 seat-argument + 72 unchanged (11 report text, 9 subject sources, 4 `Anchor` fields no
    archive carries — `finding_id`, `finding_key`, `label`, `text` — and 48 keys, hashes, dates,
    enumerated values and identities). At this part's schema Part 5 has reserved those four and
    `Finding.finding_id`/`label` are one `id`: 34 + 50 + 68 = 152.
  - `(prose) = true`: `awk '/^message /{m=$2} /^extend /{m=""} m!="" && /^ *(optional|repeated) string [a-z_0-9]+ = /{c=m"."$3; b=""} c!=""{b=b $0; if ($0 ~ /;[ \t]*(\/\/.*)?$/) { if (b ~ /prose\) = true/) print c; c="" }}' <same file>`
    → 29: 28 in the seat-argument list; `Cite.title` is source data (R-8) and is not (A-4).
- Append's structural checks stay on; a missed reference is a refusal
  `TestEveryArchivedRunMigrates` reports.

**Tests that fail if Part 6 is wrong.**
- `TestEveryArchivedRunRenders`: raw digests move (golden regenerated, every line read);
  normalized, skeleton and per-kind counts do not.
- [NEW] `TestEveryStringFieldIsInOneList` (R-19): walks every string field of every message
  reachable from an event body in the `recordpb` registry — nested messages and oneof arms
  included, so `Motion`'s `GradeMotion`, `DocketMotion`, `AvenueMotion` and `DocketRuling` (8
  fields) are walked — and fails on a field in none of the three lists or in two —
  so a field added to `record.proto` is classed before it migrates; deleting a list entry fails it.
- [NEW] `TestIdRemapChangesOnlyIdAndArgumentFields` (audit G1): over every tarball, every string field of
  every translated body in the source-data list equals its source value with marker-token ids
  respelled and nothing else; every migrated proof's `proof_sha` equals the sha256 of its cached
  script. A unit case maps `Q1`, `G2`, `Q2` and holds `"quartile Q1"` in `Proof.script`, `"G2
  summit"` in `Cite.title` and `"Q2 earnings"` in `Retire.superseded_by` — all unchanged, while
  `Log.text`'s "closes G2", `Dispatch.gap_ids` and `G2` in `Mint.problem`, `DocketRuling.principle`
  and `Proof.drift` are rewritten. Moving a source-data field into the seat-argument list fails it,
  and so does a field the schema marks `(prose)` missing from that list (`Cite.title` excepted).
- [NEW] `TestNoOldIdSurvivesMigration`: over every tarball, no string field but a `*_key` field
  holds a whole value matching `^([GQM]\d+|[fcp]-[0-9a-f]+|[a-z-]+-F\d+|L\d+-F\d+|R\d+-\d+)$` (a
  missed id field fails here, listed or not). `*_key` is excluded by name: seat keys are a non-goal
  and hold such values (`mint_key` `G1`–`G7` in quadratic and `G1` in b5; `finding_key`
  `L1-F1`…`L5-F3` in quadratic, `logic-F1`…`F3` in b5 and b6, `evidence-F1` in b8, `voice-F1` in
  b9). No seat-argument field holds a mapped id at a word boundary — the fields read from the list
  itself, so the test covers exactly the rewritten fields.
  Deleting the list branch fails it on `Dispatch.gap_ids`.
- [NEW] determinism: one tarball migrated into two directories gives identical ids.
- `TestB8ArchiveMigratesWithNoRefusal` (`archive_test.go:298`) and the b5/b9 tests resolve old ids
  through `sha256(SourceHash ‖ "G4")` from the manifest's `source_hash`.
- `record/idnamespace_test.go`, `flags/idnamespace_test.go`: letter disjointness.
- `TestEveryFindingIdPrintsItsArea` (S9): every text view of every migrated archived run, with
  the text the view copies from a seat-argument field taken out first — a seat's wording prints as
  written, live or migrated (A-4; b3's `Retire.reason` holds the remapped bare id); each printer is
  also checked by deleting its `FindingRef` call.

**Agent-facing text** — sweep (S11):
`git grep -nE '\b[GQM][0-9]+\b|[GQMFCP]<n>|\b[a-z-]+-F[0-9]+\b|<area>-F|-F\{N\}|\b[fcp]-(…|<hex>|[0-9a-f]{4,})|\(C1, F2' -- '…/tools/*.go' '…/tools/*.md' '…/tools/*.proto' '…/skills' '…/docs' ':!*_test.go' ':!*testdata*' ':!*.pb.go' ':!*/agent-memory/*' | grep -vE '^[^:]+:[0-9]+:\s*(//|--)' | grep -vE 'MustCompile|internal/record/migrate/|internal/anchor/|internal/seatprobe/boards.go'`
→ 35 lines, all CHANGE: `shapes.go:157,163,178,191,197,203`; `names.go:494` (DescKey's "(C1, F2,
P3 …)" becomes a plain word, so a key never looks like an id); `motion/verbs.go:486,488`;
`blue/avenue.go:195`; `blue/edit.go:204`; `blue/prove.go:188`; `blue/cite.go:234`;
`blue/retire.go:137`; `lens/finding.go:160,169`; `lens/mint.go:266,277`; `lens/reproduce.go:63`
(the token's `p-…` respells; `--id` keeps the sha256, R-10); `lens/verify.go:96`;
`seat/verbs.go:274,281,559`; `seat/help/propose.md:11`; `record.go:1082`; `record.proto:968`;
`claimcount.go:119`, `citationid.go:296` (trailing comments); `debate.js:871`; `SKILL.md:17`;
`docs/record-flow.md:36,69`; `docs/finding-markers.md:21,34`;
`docs/citations-and-bibliography.md:96`. Also `debate.js:602` (comment) and `seatprobe/boards.go`
(7 fixture ids, mapped by the builder as `build.go:392-399` does for avenues; `build.go:158,170,252`
and `production.go:43` stop predicting `G<n>`/`M<n>`). The regenerated `agents/*.md` and the
simulator's 16 prompt sha256s move with the text.

**Harness readers**: `fuzz_test.go:363 avenueIDPat` reads the table, and the avenue-move drive gains
the zero-across-N error `:4034` gives cites; `:3300` and `:3938` read the table.
`difftest/harness_test.go:180,197,213` normalize the new shapes (a miss fails loudly as an
unstable golden).

**Tests pinning old shapes** (Census 2, plus `8798488d`): `[fcp]-hex` 204 lines in 43 test files,
198 short ids in 45; finding labels 154 in 35 test files and 65 in 11 goldens;
`duplicatescreen_test.go` (10), `docketruling_test.go` (14), `mintbudget_test.go` (3),
`readerror_test.go` (3, pins `MintAvenueID`), `requiredmarker_test.go` (2),
`migrate/archive_test.go:298`.

### III.7 Part 7 — the spelling fork test (experiment; no code in the tree; parallel)

**Value.** It decides F-e on evidence before a migration pays for a spelling.

- **Candidates**, each with a real F-a id: `<!--C-3fa29b01-->`, `⟦C-3fa29b01⟧`, `[[C-3fa29b01]]`,
  `{C-3fa29b01}` (bare `[C-…]` collides with `[1]` and `[sic]`; `[^x]` is the footnote grammar).
- **Method** (`gray-area:elicitation-testing`): forks of real m15 seat sessions on the production
  tier (R-5), installed constitutions, no tools. The report each seat saw is rewritten into the
  candidate spelling; blue-respond, red-lens-evidence and red-lens-voice forks get the same edit
  task, return the replacement and quote the sentence they change, then answer: what is
  `⟦C-3fa29b01⟧`; who wrote it; would you quote it; does a reader see it.
- **Measures**, n ≥ 5 per seat kind per candidate (60 in all): (a) markers carried correctly;
  (b) markers invented or altered; (c) quotes that include or exclude markers correctly;
  (d) markers mistaken for a citation, footnote or prose; (e) the voice lens flagging the marker.
- **Decision rule.** Among candidates with (b) = 0 and (d) = 0, the highest (a); ties go to the
  most visually distinct. **If none qualifies, the HTML-comment spelling stays and Part 8 is
  empty.** At n = 15 a zero count bounds the true rate below ≈ 20%: the test screens out a
  spelling that fails often; the edit guard catches a rare failure loudly.
- **Artifacts.** `~/.claude/scratch/markers-plan/forktest/`; the result is pasted here before Part 8.

### III.8 Part 8 — the legible spelling (record change, epoch 23 → 24; after Parts 6 and 7)

- [MODIFY] The kinds table's token prefix and suffix, and the two readers that know the invisible
  layer by its comment shape rather than by the table: `annotationLen` (`anchortext.go:63`, kept by
  R-2; behind `locate`, `Visible` and `normalizeQuote`) and `anchor.Sentences`' comment skip. Both
  also skip the table's token and keep skipping HTML comments and footnote references — live
  behaviour, not old-record support: an edit can carry a plain HTML comment, and it stays opaque
  to every reader; archived tokens are respelled by `migrate`; `Sentences`'
  anchor clauses (§III.2 rules 2–3) read the table's token. Census:
  `git grep -nF '"<!--' -- '…/tools/*.go' ':!*_test.go' | grep -vE '\.go:[0-9]+:\s*//'` → 14 at
  `8798488d`: 9 become the table in Part 1 (`anchor.go:30,32,34,66,125`, `cite.go:166`,
  `prove.go:147`, `finding.go:104`, `verify.go:341`), `claimcount.go:82` and `render.go:227` become
  `Sentences`, `anchortext.go:64` is `annotationLen`, and `dashboard/render.go:257` and
  `setup/setup.go:156` write files that are not the report.
- [MODIFY] Text that describes markers: `SKILL.md:17,25`; `report_template.md:82`; `debate.js:939`
  ("invisible anchors"; 4 simulator prompt goldens move); `seat/help/cite.md:7`,
  `seat/help/edit.md:19`; `bluedoc.go:36`, `edit.go:92`, `cite.go:308` ("invisible");
  `terms.json:566-611`, then `vocabdoc`; `docs/finding-markers.md`,
  `docs/propagation-and-anchoring.md`, `docs/record-flow.md:89`, `docs/seat-command-triggers.md:86`;
  the 58 Go comment lines S1 counts; the regenerated agents. The text states the spelling once:
  a marker is how seats point at a place; the tool places it; carry it; never type one.
- [MODIFY] `migrate`: one translation respelling marker tokens in every string field — Part 6's
  token pass with another table; the field lists and byte-identity tests unchanged.
- Test: `TestEveryArchivedRunRenders` normalized and skeleton digests unchanged. A token in the
  chosen spelling is skipped by `LocateSpan`, `Visible` and `Sentences`, and `TestSentences`' anchor
  shapes pass in it; deleting either reader's table check fails it.

### III.9 Order, and every boundary a later part's reader crosses

**Part 1 → 2 → 3 → 4 → 5 → 6**; Part 7 from the start; Part 8 after Parts 6 and 7. Parts 1, 2
and 3 change no record. Epochs: Part 4 20 → 21, Part 5 21 → 22, Part 6 22 → 23, Part 8 23 → 24; each
carries its migration change and archived-run tests. `migrate` translates each archived word once,
straight to the current shape (`registry.go:32`); `remap.apply` then respells ids and tokens; Part
4's placement step (widened in Part 5) runs last, before `Append`, on bodies already in the
destination's spelling. Each step adds only what its source lacks, so a run written at any epoch
migrates.

**Boundary walk.** For each reader or check a later part changes, the earlier part that reaches it
first and its treatment there (method: the S10 lines, the Part 4/5/6 censuses and the fuzz readers,
each read against the gap row, the `Anchor` pair and the id shape):

| Reader (changed in) | Reached first by | Treatment |
|---|---|---|
| `mint.go` `LocateUnique` (unique check) | Part 4 (mint places) | `LocateOnce` lands in Part 3; `Attach` uses it from Part 4; `TestMintRefusesWhatLocateUniqueRefused` |
| `viewjson.go:1033` `Kind(about_ref)` (Part 6) | Part 4 (`Kind("G3")` = gap) | Part 1 reads `Backs`; Part 4's location loop too; `gapBacking` test |
| `shapes.go:44` `anchorShape`, `:133` `anchorToken` (Part 6) | Part 4 (`retire --anchor G3`, `show report --anchor G3`, a pasted gap token) | read the table from Part 1; the retire test passes `--anchor G3` |
| `shapes.go:40` `^G\d+$` | Part 4 | unchanged: gap ids are `G<n>` until Part 6 |
| `fuzzAnchorRe`, `anchorsLeftFor`, fuzz edit/proposal drives (Part 6) | Part 4 | moved into Part 4 (N10) |
| `capture.go:1124 anchorID` (Part 6) | Part 4 (`G3` in a closure's `anchor_tool`) | KEEP in Parts 4–5: a gap id was never in its set and reconciles by prose as today; Part 6 reads the table |
| `reportvoice/tells.go:80-89` (Part 6) | Part 4 (gap tokens in a replacement) | KEEP: `gap:G3` matches no tell (colon, no space); `spanVoiceTells` test |
| `report/md.go:57 idToken` (Part 6) | Part 4 | KEEP: it reads assembled text, gap tokens already stripped |
| `difftest/harness_test.go:180-213` (Part 6) | Part 4 (`<!--gap:G1-->` in goldens) | KEEP: `G<n>` is stable across runs |
| Part 4's placement step | Part 6 (respelled ids and tokens), Part 8 (respelled tokens) | runs after `remap.apply`: respelled `old`/`new` against a respelled render; its `Anchor` carries the remapped gap id |
| `annotationLen`, `Sentences`' comment skip (R-2 KEEP) | Part 8 (a token that is not a comment) | read the table in Part 8 |
| `consistency.go:205-220` (F-g) | Part 4 (mint), Part 5 (cite/prove/verify) | split by part |
| `report_op` (Part 5) | Part 4 (gap `Anchor`s) | the existing `anchor` arm renders them |
| UNIQUE `anchor.id` (Part 5) | Part 4 (one gap `Anchor` per mint, retry checks) | lands in Part 5; the placement step's pre-scan keeps a re-migrated run to one `Anchor` per id |
| `Anchor.id`, `Cite.label`, `Verify.label` (Part 6 id list) | Part 5 (`Anchor` carries `c-`/`p-` ids) | all in the id list; Part 6's raw-digest-only change proves it |
| `bluedoc.AutoPlace`, `LocateOnce` (Part 3) | Part 4's placement step | in `bluedoc`/`anchortext`, callable from `migrate` |
| `anchor.Sentences`' rule (Part 2) | Part 3 (`AutoPlace`), Part 4 (location, backing, F-c fallback, `edited_since` through `reopened`) | each reads `Sentences` and nothing else; `TestEverySentenceReaderReadsOneSentence` gains each part's rows |
| `RemoveAnchorAt`'s token-in-place read (Part 1) | Part 2's anchor clauses | same answer under Part 1's rule; Part 2's "One. <c>. Two." case needs it |
| Part 6's token pass | Part 8 | reused with another table |

## IV. Risk & Mitigation (likelihood × impact × cost-to-mitigate)

| # | Risk | L × I × C | Mitigation |
|---|---|---|---|
| R1 | Part 1's refactor moves a historical render | M × H × L | `TestEveryArchivedRunRenders` lands first, against unrefactored code; the golden must not move. Universe runs compared by hand (§V.2). |
| R2 | The unique-quote check (Part 4) refuses quotes that placed silently, and seats stall | M × M × L | A write-time refusal that says how to make the quote occur once within its paragraph; replay untouched. Ruled residue (R-14): a sentence repeated verbatim as its own paragraph cannot be anchored — the refusal says so, and mint's offers `--about-kind section`. The `error_catalogue` golden shows every new refusal; universe smoke after Part 4. |
| R3 | Two random ids collide | L × M × L | Minted columns are UNIQUE and refuse loudly (≈1 in 8,600 for 1,000 ids). Residue: re-carried ids (≈3×10⁻⁷ for 50 avenues). Release fuzz under `-race`, `FUZZ_C=6`. |
| R4 | The migration rewrites subject prose or source data | L × H × L | Three field lists under a guard (R-8, R-19); `TestIdRemapChangesOnlyIdAndArgumentFields`; `proof_sha` re-checked. Residue: a seat-argument field using a mapped id's spelling for something else. |
| R5 | Gap markers raise blue's carry burden and refusals | M × M × M | Part 3 ships first; `--accept` carries the gap's anchor (A-3). Edit refusals per sitting on the smoke before and after Part 4. |
| R6 | Closed gaps' inert markers clutter reading (F-b) | M × L × L | Part 7's (d) measures it; assembly strips them. A remedy removing a closed gap's anchor is a per-kind rule and returns as a fork against R-6. |
| R7 | Legible markers get quoted into `mint --quote` or `edit --old` | M × L × L | `Visible`/`LocateSpan` strip markers through one definition (Part 1). |
| R8 | A finding id prints without its area | M × L × L | One renderer; S9's test. |
| R9 | The migration misses an id-bearing field | L × M × L | `TestNoOldIdSurvivesMigration` checks every string field for a whole old-shape value, listed or not (`*_key` excepted, a non-goal). |
| R10 | Migrated gaps land where blue would not have put them (F-c fallback) | H × L × L | Counted and listed by name in the golden and in S7. Historical runs are read for audits, not re-debated. |
| R11 | A gap reads as located when it is not | M × H × L | `location_state` names every case; `TestEveryArchivedGapHasANamedLocation`. |
| R12 | The Part 4 placement step is slow on large runs | L × L × L | One render per mint and edit; timed on the largest archived run. |
| R13 | With no hold, blue's retire takes a finding or gap anchor while its gap is open | M × M × L | The gap reads `gone` with its minted text; `show report --anchor` names the retire; the retire still needs the cut first. |
| R14 | A migrated gap whose quote never placed reads `gone` | L × L × L | The teaching names both causes and asks for a judgment of the report as it stands, not of a cut; those gaps are listed by name in the golden (A-4). |
| R15 | `blueanswered.go` says "NO edit answers it" beside a `gone` gap | — | Closed: its engaged arm reads the gap's location state (a field, not a join) and `TestBlueAnsweredReadsAGoneGapAsMoved` pins it. |
| R16 | The sentence rule joins a sentence opening with a digit to the one before it, splits after an initial or before an opener after a non-final terminator, or misreads a block | M × L × L | Measured on the 21 renders (§III.2): 37 two-sentence joins, 5 initial splits, 5 opener splits, 0 fragments among the seven or in b3; a joined pair widens a location to two sentences, never cuts one. Tables, which split per cell, occur in no rendered corpus report. Block openers follow markdown's own rules (R-25), one `TestSentences` row per condition. `TestSentences` pins "Dr. Smith" and "et al. (2020)", and the corpus test pins m8's join, so changing the rule is a decision. |

## V. Verification Plan

### V.1 Per part (the validation loop; `unset $(env | grep -o '^FEOV_[A-Z_]*')` first, every call)

1. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./...` — count FAILed packages and grep
   `build failed`; includes `requiredmarker_test.go`, `deadpath_test.go`, the naming gates and
   Part 2's three sentence tests. Run in the repository, never a scratch copy: Part 2 is measured
   here with the 57 repository-root tests (prompt and vocabulary gates among them) the scratch
   measurement could not run — pass: no FAIL that `8798488d` does not show. · re-armed by: any Go
   change.
2. `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRun|TestEveryArchivedGap|TestGapTranslation|TestNoOldIdSurvivesMigration|TestIdRemapChangesOnlyIdAndArgumentFields|TestEveryStringFieldIsInOneList|TestB8Archive' ./internal/record/migrate/`
   → all pass; `archived_renders.golden` raw unchanged (Parts 1, 2, 3, 5), skeleton unchanged
   (Parts 4, 6, 8). · re-armed by: `anchor`, `anchortext`, `bluedoc`, `claimcount`, `reportproj`,
   `migrate`, `record.proto`.
3. `go -C scripts test -count=1 ./...`
4. `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` (Parts 4, 6, 8) ·
   re-armed by: `debate.js`.
5. `go -C scripts run ./agentgen -check` (Parts 4, 6, 8) · re-armed by: help text, JSON shapes.
6. `go -C scripts run ./schemagen -check` and `go -C scripts run ./protogen` (Parts 4, 5, 6, 8) ·
   re-armed by: `record.proto`, `requirements.json`.
7. `go -C scripts run ./vocabdoc` (Parts 4, 8) · re-armed by: `terms.json`.
8. `UPDATE_GOLDENS=1 go test -count=1` over difftest, recordsql, dashboard, cli, seatprobe and
   migrate; read every changed line.
9. `go -C scripts run ./check` on the committed tree; read `N gate(s): X passed, Y failed`.
10. `FEOV_RELEASE_GATE=1 go test -count=1 -timeout 40m ./releasegate/fuzz/` and the `-race`
    `FUZZ_C=6` variant (Parts 4, 5, 6, 8); read every ZERO gate. · re-armed by: record, flag or
    enum changes, the fuzz drives.
11. S1–S5, S10–S14 commands return their targets; the S6 numstat goes in the PR body.

### V.2 Real data

- **Render identity, universe runs.** The 12 universe runs in `~/.claude/scratch/locator/mig`,
  rendered with the base binary and the part's binary: raw equal (Parts 1, 2, 3, 5), skeleton
  equal (Parts 4, 6, 8).
- **Sentence rule (Part 2).** Harness `~/.claude/scratch/markers-plan/r7/cc` (`claimcount` at
  `8798488d`, three rules side by side, its block reader rule 1 as R-25 states it) over
  `~/.claude/scratch/markers-plan/r5/renders`: `go run` `./cmd/m6` (Count, `Index`, bare set),
  `./cmd/wrap`, `./cmd/cc6`, `./cmd/cls6` (the table), `./cmd/op` (residues), `./cmd/b3`
  (`r6/b3`), `./cmd/frag6` (`r6/bases`); `RULE=r6 go test ./claimcount/` (`-run TestRows -v` prints
  the `TestSentences` shapes); the `retire_anchors` query over
  `~/.claude/scratch/locator/mig/*/records/record.db`. Outputs before and after the R-25 reader:
  `r7/before`, `r7/after`, equal. Pass: the §III.2 numbers, re-run on the part's own
  `anchor.Sentences`.
- **Locator audit (S7, Part 4).** The 28 runs re-migrated with the Part 4 binary; the harness
  (`locatoraudit_scratch_test.go`, out of tree) reads `location_state` and the marker offset. Pass:
  every quote gap has a state; every `marked` location is one sentence holding its marker; the six
  fragment gaps in renderable runs (all but quadratic `R2-2`) and b3's three hold their token, in
  the render right after their mint, in the whole sentence `TestSentencesKeepsTheCorpusSentencesWhole` names,
  compared as strings; 0 drift
  against the 82 finding-token pairs outside the listed fallback placements; b9 `G4` and b5 `G5`
  resolved by name.
- **Auto-place replay (S8, Part 3).** `autocarry.py`'s 211 carries through `bluedoc.AutoPlace` with
  blue's marker removed from `new`. Pass per S8; the moved-marker residue listed.
- **Universe smoke** after Parts 1, 2, 3, 4, 5 and 6 (`scripts/universe.sh build`, grep the installed
  copy for a phrase from the change, then `run "Is 91 prime?"`, haiku, one lane). Pass: a verdict
  with 0 marker refusals the seat cannot recover from. Watch: edit refusals per sitting against
  m15, `location_state` on the board, `--accept` applications.

### V.3 Gate

Approved by gblock on this revision without a further audit (R-26); `plan-audit` is not run
again. Each part ships as its own PR that carries its part's required tests, the §V.4 notes of
that part included, in its body, runs §V.1 to green with the fail counts read, and gets an
independent review before merge.

### V.4 Notes owed to implementation

Every audit note still owed to implementation, with the required test of its part that proves it
(the one wording note moves no code and has none). Each PR body carries its part's rows.

| Note (audit) | Part | Resolution | Required test |
|---|---|---|---|
| L1 — the typed pair compared with the stored location (r6) | 4 | `ProposalAppliedVerbatim` compares with `record.Proposal`'s output (R-24); the typed-pair sweep (§III.4) | `TestAcceptCarriesTheGapAnchor`'s sixth row (b3 `G1`'s shape); the fuzz's typed-arm ZERO gate; `queries_parity_test.go:145-149` |
| `>` continuation and the three `>` rows (r6) | 2 | Rule 1's block quote (R-25) | `TestSentences`' `>` rows |
| Openers markdown does not read as a block end — `#` with no space, `\|` with no delimiter row, `<n>.` other than 1 inside a paragraph (r6) | 2 | Rule 1's openers (R-25); 0 in the corpus, every corpus number unchanged (§III.2) | `TestSentences`' opener rows |
| `insideFence` reading the block reader changes replay (r6) | 2 | Stated, with its gates and residue (§III.2 *One block reader*, A-4) | `TestInsideFenceReadsTheBlockReader`; `TestEveryArchivedRunRenders`; §V.2 render identity |
| `RemoveAnchorAt` keeps an emptied ordered item's "1." (r6) | 2 | Unchanged behaviour, stated; the body-reading fix is optional and outside the plan | the "1. <c>" row of `TestEverySentenceReaderReadsOneSentence` |
| The assembly's fence readers (r6) | 2 | Wording: they split blue's markdown into `## ` sections; KEEP (§III.2) | none: no code moves |
| m8's residue literal (r6) | 2 | The sentence starts "Base 2: Compute 2^90 …" | `TestSentencesKeepsTheCorpusSentencesWhole`'s m8 row, literal copied from the record |
| The 57 repository-root tests the scratch measurement could not run (r5) | 2 | Run in the repository (§V.1 item 1) | §V.1 item 1: no FAIL that `8798488d` does not show |

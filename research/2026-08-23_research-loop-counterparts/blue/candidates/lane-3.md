# Lane 3 — local-repo critical-stance: the loop as it is BUILT, not as it is described

**Method lens:** direct audit of the subject artifacts and codebase at pinned commit `2ce929f`
(the run's PINNED base). This lane trusts nothing secondhand: where lanes 1–2 read the two design
documents and the external literature, this lane reads the Go that implements the FEOV record and
greps the tree for what the memory/dream loop has actually shipped. Its job is to say which of the
positioning claims survive contact with the code — and to add the one axis only a code audit can
supply: **built-and-refusable versus proposed-and-unmediated.**

Order of work: hypothesis 3 (the mediated-record discipline) first, then breadth across hypotheses
1, 4, 5 as they are carried by the artifacts. Lanes 1–2 already saturated the external corpus
(AI-Scientist, Darwin-Gödel-Machine, AlphaEvolve, FunSearch, Reflexion, Self-Refine, the OKF
spec, the multi-agent-debate literature); this lane is additive to theirs and does not re-walk them.

Source convention for this lane: repository artifacts are cited by path + line at pin `2ce929f`,
which is the leaf a skeptic re-reads (`git show 2ce929f:<path>`). External sources are noted with
URL/title/access date only where this lane introduces one; it introduces almost none by design.

---

## Part A — The through-line this lane adds: built-and-refusable vs proposed-and-unmediated

Lanes 1–2 correctly frame the three in-repo loops as **designs** and compare their intended
machinery. A code audit sharpens that frame into a fact that changes what "compare and contrast"
can honestly claim:

**One of the three loops is implemented and enforced in a compiled tool; the other two do not exist
as code at all at pin `2ce929f`.**

- **frank-exchange-of-views / the FEOV record** is a substantial Go program: over 330 `.go`
  source and test files under `plugins/frank-exchange-of-views/tools/` (glob of
  `plugins/frank-exchange-of-views/tools/**/*.go` at `2ce929f` reports 336 matches). It is the
  running substrate this very research run writes through.
- **sleeper-service** (the `/self-improve` loop's home) ships exactly three files at `2ce929f`:
  `.claude-plugin/plugin.json`, `README.md`, and `bin/.gitkeep` (an empty binary directory). Its
  own README states: *"Status: scaffold only (Phase 0)"*
  (`plugins/sleeper-service/README.md:9`).
- **The dream loop and the memory architecture** have **zero implementation** at `2ce929f`. A tree
  search for the primitives the design names — `dream.js`, `memory-consolidator`, `memory-curator`,
  `knowledge-ingest`, `.knowledge.toml`, any `okf`-named profile file, a `.claude/knowledge/`
  store — returns **nothing** (find over the worktree, excluding `.git`/`node_modules`, at
  `2ce929f`). `plans/memory-architecture.md:3` labels itself *"Status: Proposal (for discussion)."*

This is not a pedantic distinction. It sets the altitude of every comparison in the report:
**FEOV is compared as a working mechanism; dream and memory are compared as intentions.** A
convergence claim ("all three share a promotion ladder") is, on the FEOV side, a property you can
run and, on the other two, a property someone still has to build. The report must not let the
symmetry of the *prose* ("three scheduled human-gated git-native loops") hide the asymmetry of the
*state* (one enforced in Go, two on paper). Where lane-1 wrote "these are DESIGNS… intended
machinery not shipped behavior," this lane supplies the leaf proof of exactly how much is shipped:
one loop's worth of enforcement, and none of the other two's.

---

## Part B — Hypothesis 3, verified in the code: the load-bearing novelty is a REFUSABLE record

Lane-2 pursued hypothesis 3 by reading the OKF spec and concluding the memory store's persistence
layer is "structurally weaker than the FEOV record it parallels." That conclusion is correct, and a
code audit makes it **stronger and more concrete** than a spec comparison can: the weakness is not
"OKF is permissive" in the abstract — it is that **FEOV's record refuses malformed writes in
compiled code, and the memory store, as designed, refuses nothing.**

### B.1 — The FEOV record is a validated, append-only, single-write-path log

Read at the leaf (`plugins/frank-exchange-of-views/tools/internal/record/record.go` at `2ce929f`):

- **One write path.** `Append` is documented and implemented as *"the ONLY write path. It
  validates, derives the idempotency key, and assigns the per-shard sequence number"*
  (`record.go:354`, function at `:445`). Every event a seat emits — mint, close, proof, cite,
  finding, verdict, motion, friction, line-of-inquiry — passes through it.
- **Refusal at the write.** Before appending, `Append` calls `validate(runDir, seatID, typ, p)`
  (`record.go:471`). `validate` spans roughly `record.go:528–1059` (~530 lines) and is a
  per-event-type gauntlet of refusals, each with a verbatim error a seat reads. A sample of what it
  will not let onto the record:
  - a `mint` (raising a gap) missing `--check`, `--class`, `--check-kind`, `--likelihood`,
    `--impact`, or `--problem` is refused (`record.go:553–598`);
  - a grade outside the canonical enum is refused (`record.go:529–538`);
  - a `verdict` of `PASS` is refused if any gap is still open — *"A PASS is a claim that nothing is
    left open. Enforce it here, at the one write path, so no verdict route can record a PASS over an
    unadjudicated board (the 2026-07-20 rubber-stamp: PASS with 9 open gaps)"* (`record.go:879–892`,
    `requirePassClosesAllGaps`);
  - a `close` without the verification triple (`--verified-by/--verified-with/--verified-against`)
    or a genuine `--carried-from` is refused as *"an unverified closure is unauditable"*
    (`record.go:653–667`); a `--carried-from` that names no real prior closure is refused as a
    laundering path (`record.go:684–692`);
  - a `retire` (removing a claim) with no `--reason` is refused (`record.go:870–878`);
  - dangling references — a `supersedes`, `answers`, or `successor` naming a gap the board does not
    hold — are refused *here* rather than "accepted and dropped at replay" (`record.go:606–613`,
    `:619–621`, `:693–701`).

  This is precisely the `facts-are-fields` ideal named in the repo's own rule corpus: *"a field on
  a record a writer can refuse — validated at the write, where a wrong value is an error the author
  sees"* (`plugins/prosthetic-conscience/skills/facts-are-fields/SKILL.md`). FEOV does not merely
  aspire to it; it implements it as the only way to write.

- **Append-only physics.** The package header describes *"an append-only per-process-shard event
  log with structural idempotency keys, render-time deterministic merge"* (`record.go:1–8`). Events
  carry a monotonic logical clock guaranteed in code, not hoped for from the hardware
  (`nextStamp`, `record.go:399–443`); a torn line from a crash is healed rather than overwritten
  (`appendLine`/`durableAppend`, `record.go:489–523`); a re-dispatched seat rotates its shard nonce
  so a retry loses nothing (`RegisterSeat`, `record.go:206–306`).

### B.2 — "Proof, not prose" is enforced, not merely offered

Lane-2 and the blue constitution assert that a computation-gap "cannot be closed on prose." A code
audit confirms this is a hard refusal, not a convention:

`plugins/frank-exchange-of-views/tools/internal/cli/merge/close.go:39–57` — if a gap was minted
`--check-kind computation` and `record.ProofAnswers` finds no proof event naming it, `merge close`
returns an error and the closure does not happen: *"closing it on prose would accept the one kind
of evidence you declared insufficient."* And the demand cannot be dodged upstream: `mint` **requires**
`--check-kind` (`record.go:568–570`), a requirement added because *"The 2026-08-05 smoke produced
ZERO proofs across a full run… because NOTHING ASKED: all ten of red's acceptance checks were
document probes"* (`record.go:560–570`). So the loop does not just permit rerunnable verification —
it forces red to declare, per check, whether prose could ever settle it, and wires a compiled
refusal to the answer.

This is the mechanized form of the operator's standing requirement ("prototyping and testing …
proof, not prose"). It is also the code-level version of lane-2's finding that FEOV imports
execution-verification into open-ended research: the import is real and enforced, not a stated
intention.

### B.3 — The repo's OWN history is the proof that this discipline is load-bearing

The strongest evidence for hypothesis 3 is not in the record package — it is in the auditor that
runs after every research run. `plugins/frank-exchange-of-views/tools/internal/capture/capture.go`
(1811 lines at `2ce929f`) is a **graveyard of the exact failure the memory design's persistence
layer would reintroduce.** At least six of its audits carry a documented history of a prior version
that read a fact out of a markdown file or a regex over prose, and *returned a plausible zero that
read like a clean board* — the signature `facts-are-fields` failure:

- **"AUDIT 2" was deleted outright.** It compared line counts in `red/ledger.md` and
  `red/archive.md` against self-reported envelope counts; when those files became rendered
  projections and stopped being written, it *"returned SKIP, 'no ledger/archive (pre-sharding
  run)' — a benign, plausible reading, and wrong… those were the NEWEST runs. Every 2026-08 capture
  reports it, and the audit had been measuring nothing for months"* (`capture.go:215–237`).
- **`FrictionAudit`** *"COMPARED PROSE TO PROSE, and reported 5 failures out of 5 on a run where
  every one of them was on the record"* — because it recovered the seat identity by splitting a
  string instead of joining on a field (`capture.go:249–309`). Fixed by joining on the `seat_id`
  the record holds as a field.
- **`AssemblyScreen`** regex-scanned a rendered prose column for `REFUTED|ABSENT`; when the grade
  became a closed enum with no field for a failed verification, the file became a 46-byte stub and
  the screen *"reported PASS — 0 REFUTED-row token(s)… on every record-mode run"* (`capture.go:410–439`).
  Fixed by joining on record fields (`refutes`/`absent` outcomes + `--anchor`).
- **`AttestationAudit`** read `red/archive.md`, splitting a pipe-delimited prose line back into the
  `anchor_seat`/`anchor_tool`/`anchor_target` fields the record already held — *"Fields to string
  to regex to fields, with a markdown file in the middle purely as a courier"* — and *"reconciled
  nothing, in silence, for as long as the record tier has existed"* (`capture.go:785–795`).
- **`RecordParityAudit`** counted round records from headings in a hand-authored `blue/CHANGELOG.md`;
  the `2026-08-05` run *"carried a 6,847-byte CHANGELOG and exactly ONE revision event"* — the regex
  read the plausible number and passed while the round records were never filed (`capture.go:612–620`).
  Fixed by counting `revision` events on the record.
- **`BackfillAudit`** replaced a `RecordJoinAudit` that scraped Bash command strings out of agent
  transcripts and *"had FIVE independent ways to be wrong,"* the fatal one silent (`capture.go:644–680`).
- **`HarvestPrecedents`** once read rulings from envelope free-text arrays: *"A bench that ruled six
  gaps and listed four promoted four, and nothing noticed… capture reported 'precedent harvest: 0
  ruling(s)' — the same bytes as an honest run where the bench genuinely ruled nothing"*
  (`capture.go:1141–1160`). Fixed by reading ruling events from the record.

Every one of these was the same defect: a fact that lived in a markdown file or a prose substring,
recovered later by a reader, failing silently by returning a zero indistinguishable from health.
**The FEOV team paid for this lesson six-plus times, in code, and the fix was always the same —
move the fact onto a refusable field on the record and read it in-process.**

### B.4 — The memory/dream store, as designed, is the shape those audits were deleted for

Now read the memory design at the leaf against that hard-won standard
(`plans/memory-architecture.md` at `2ce929f`):

- Its persistence layer is *"a directory of markdown files with YAML frontmatter"* (OKF); *"the
  only required frontmatter field is `type`"* (`§3`, `:49`). The lifecycle fields that drive the
  entire loop — `status`, `confidence`, `last_seen`, `review_count`, `supersedes`, `provenance` —
  are **all optional** and *"recovered by READING"* (`§3.1`, `:91–94`). A file missing them *"is
  still a valid OKF concept — it is simply treated as a low-confidence candidate."* That is the
  plausible-zero shape exactly: an absent `review_count` and a genuine `review_count: 0` are the
  same bytes, and a consolidator that reads them cannot tell "never corroborated" from "field never
  written."
- Its consolidation step (`/dream`, `§7.5`) is a **merge agent** (`memory-consolidator`) editing
  markdown by *"overlap detection"* — the identical operation FEOV's `close`/`supersedes`/`mint`
  route through `validate()` refusals (dangling `supersedes` refused, `repaired_with_regression`
  requires a successor, lineage never drops). The memory design's version has no refusable writer.
- Its **mitigation for silent knowledge loss is human review of a git diff**: *"Bad merges silently
  lose knowledge (the OpenClaw diary's 'details unavailable' failure). Mitigation: every merge is a
  git diff a human can review"* (`§9`, item 4, `:254`). This is precisely the mitigation the FEOV
  audits abandoned as insufficient — a human skimming a diff is not a validator that refuses a
  malformed write; it is the OpenClaw failure mode the design itself cites as the cautionary tale
  (`§2`, `:37`).

**So hypothesis 3 survives at the leaf, and in a stronger form than the spec comparison yielded:**
the load-bearing novelty of the FEOV loop is not multi-agent debate; it is that every fact another
party acts on is written through a compiled refusable field, and the same repository has already
deleted, six or more times, exactly the unmediated-markdown persistence the dream/memory loop is
currently designed to have.

### B.5 — The disconfirmers, run honestly (this lane's disconfirming budget)

The protocol requires hunting evidence *against* the current position. Three probes, and what each
returned:

1. **"Is git + PR review not itself a refusable writer?"** Partially yes, and this narrows the
   claim honestly. A commit *can* be refused — by branch protection or a schema-validating CI gate.
   The memory design gets the *projection* layer right on exactly this basis: `projections/active.md`
   is **generated, never hand-edited** (`§4`, `:127`; `§5`), which is the `facts-are-fields`
   preferred remedy ("prefer GENERATING the derived carrier over guarding two hand-written copies").
   **But the design specifies no schema-validating gate on the concept layer** — nothing in
   `plans/memory-architecture.md` proposes a CI check that refuses a malformed OKF concept at the
   commit; the only stated gate is human diff review (`§9.4`) and a secret-scrub before push
   (`§9.5`). So the disconfirmer fires against the *projection* layer (correctly mediated by
   generation) and **fails against the concept layer** (unmediated). The precise, defensible claim
   is therefore narrower than "the memory store is unmediated": it is *the memory store's
   lifecycle-field / concept layer is unmediated, while its projection layer is correctly generated.*

2. **"Is FEOV's record actually refusable in practice, or do seats route around it into markdown?"**
   This run itself shows the mediation is defense-in-depth, not airtight: the `register` result and
   the work board both report *"this run's PreToolUse hook did not fire"* (the identity-binding hook
   is absent this sitting), and the blue/red constitutions repeatedly warn seats not to hand-write
   artifacts — which they warn precisely because a seat *can*. The honest qualification: FEOV's
   refusability is real for anything written **through** `Append`, but the tool cannot force a seat
   to use it rather than narrating into a markdown file. That is exactly why capture.go ships
   `BackfillAudit`, `StrayRecordsAudit`, and `RecordParityAudit` — a run that routes around the
   record is caught after the fact, not prevented. So the claim is not "FEOV cannot be bypassed"; it
   is "FEOV's load-bearing facts are refusable at the write **and** a bypass is detected by a
   post-hoc audit that reads the record, not prose." The memory design has neither half built.

3. **"Is the memory store's weakness even relevant if the store does not exist yet?"** This cuts
   toward the recommendation, not against the finding. Because nothing is built (Part A), the
   weakness is a **design-time correction that is free to make now** and expensive to retrofit
   later — which is the most actionable possible state for a positioning report to catch it in.

---

## Part C — Breadth, verified against the artifacts (hypotheses 1, 4, 5)

### C.1 — Hypothesis 1 (red owns the gate): confirmed, and TWO-SIDED, in code

Lanes 1–2 established that FEOV externalizes the adversary. The artifact audit confirms it and adds
a nuance the design-doc reading understates:

- Red is a **separate seat with its own constitution and `memory: project`**
  (`plugins/frank-exchange-of-views/agents/red-auditor.md:1–9`), and it *owns the binary verdict*:
  *"return exactly the envelope the invoker specifies, including its binary verdict field — PASS
  only when every remaining gap is closed, rebutted with evidence you accept, or explicitly
  defect_accepted"* (`red-auditor.md:64`). The `PASS`-over-open-gaps refusal in `record.go:879–892`
  is the compiled backstop to that duty.
- **The nuance a code/constitution audit surfaces:** the gate is calibrated in *both* directions.
  Red's TELOS is *"a CERTIFIED report: you win either by finding real defects or by issuing a PASS
  that survives scrutiny. The gate opening is not red losing — an unearned FAIL is red losing,
  exactly as an unearned PASS is… Never-soft-pass has a twin: never-hard-fail"*
  (`red-auditor.md:32–40`), grounded in *"13/13 FAIL across three runs"* under an earlier
  endpoint-less constitution. So the FEOV differentiator is not merely "an adversary can withhold
  PASS" (which, as lane-2 found, is emerging elsewhere). It is a **two-sided calibrated gate**:
  refusing certification to converged material is scored against red as grade inflation, exactly as
  a rubber-stamp is. This is a sharper positioning point than "has an external adversary," and it is
  visible only by reading the seat's own constitution.

### C.2 — Hypothesis 5 (execution verification): confirmed and mechanized

Covered in B.2. The artifact layer confirms lane-2/Q5 in its strong form: not only is
execution-based verification available, it is a **required declaration at mint** and a **compiled
refusal at close**. The seam lane-2 identified (AI-Scientist executes experiments but gates paper
quality with an LLM reviewer) is exactly the seam FEOV closes by refusing to let a `computation`
check be settled by review prose.

### C.3 — Hypothesis 4 (gate placement): the in-repo divergence a code audit reveals

Lanes stated the three in-repo loops converge on gate position (propose freely; human ratifies
behavior change). The artifact audit confirms the *intent* and adds a **divergence in enforcement
altitude** the design-doc reading cannot see:

- FEOV's gates are **mechanized**: the `verdict`/`PASS` refusal, the computation-close refusal, and
  the post-hoc `capture.go` audits (exit 2 on any FAIL, per its package header) are compiled Go.
- The sleeper-service / dream / memory **promotion gate is currently only PROSE**. The guardrail
  "the loop writes only `research/` and `ideas/`; promotion into rules/skills requires the human"
  lives in `plugins/sleeper-service/README.md:7` and `plans/claude-port-plan.md:367` as a design
  statement. There is **no implemented write-scope enforcement** at `2ce929f` (the plugin is
  scaffold-only; the relevant `sc-secrets-gate` hook is marked Phase 2 in the port plan's hook
  inventory, `claude-port-plan.md:227`). So the three loops **converge on gate *position* but
  diverge on gate *enforcement*:** one loop's gate is refusable in code; the other two's gate is an
  intention in a README.

This is a concrete, filable finding: **to make the three loops genuinely parallel, the
sleeper/dream/memory promotion gate wants the same mechanization FEOV's record already has** — a
compiled PreToolUse write-scope guard that refuses a write outside `research/`+`ideas/`, rather than
a prose guardrail a headless `claude -p` loop can silently overrun.

### C.4 — A convergence the code already demonstrates: friction is mediated; knowledge is not

One integration point only the code shows. The port plan's Phase 4 says `/self-improve` *"consumes
the friction records"* as a first-class input (`claude-port-plan.md:410`). Friction is **already a
refusable event on the FEOV record**: `feov-record friction` requires a `--reason` even in its empty
form (`record.go:775–785`), and `capture.go`'s `FrictionAudit` reconciles every envelope-reported
friction against the record. So the self-improve loop's *substrate* (friction) is already mediated,
while the memory loop's *substrate* (knowledge concepts) is designed to be unmediated markdown. The
two sleeper-service loops therefore inherit **different substrate disciplines from birth** — and the
cheap, correct move is to give the knowledge substrate the same refusable-record treatment friction
already has, rather than inventing a second, weaker persistence model beside the one that works.

---

## Part D — What only this lane changes about the report's recommendations

Additive to lanes 1–2; nothing here subtracts their findings. This lane's distinct, leaf-grounded
deliverables:

1. **Reframe every convergence claim by build-state (Part A).** The report's compare-and-contrast
   must state, once and plainly, that FEOV is compared as a running mechanism while dream and memory
   are compared as unbuilt proposals (zero implementation at `2ce929f`). Filable as a framing note,
   not an issue.

2. **The record-discipline recommendation is corroborated from the code side, and sharpened
   (Part B).** Lane-2's "adopt FEOV's record discipline for the memory consolidate step" is right;
   the code shows *why it is urgent and how*: the same repo has deleted the unmediated-markdown
   persistence pattern six-plus times as silent-zero failures (`capture.go`), and the fix was always
   "refusable field, read in-process." Filable issue: **give the OKF memory profile a
   schema-validating write gate (a compiled `validate`-style refusal, or a CI check that refuses a
   malformed concept at commit) before the consolidator is built** — the concept layer, not the
   already-correct generated projection layer (per the B.5 disconfirmer).

3. **Mechanize the promotion gate (Part C.3).** Filable issue: **the sleeper/dream/memory
   write-scope-and-promotion guardrail should ship as a compiled hook that refuses out-of-scope
   writes, not as README prose**, matching FEOV's enforced gates — especially because these loops
   are designed to run headless (`claude -p`), where a prose guardrail binds nothing.

4. **Unify the two sleeper-service substrates (Part C.4).** Filable note: friction records are
   already a refusable event on the FEOV record; the knowledge substrate should reuse that
   discipline rather than introduce a parallel, weaker markdown store.

5. **Positioning nuance for the external comparison (Part C.1).** FEOV's differentiator against the
   external corpus is best stated as a **two-sided calibrated gate** (unearned FAIL penalized as
   much as unearned PASS), not merely "an external adversary," which the 2025–26 literature is
   beginning to match.

---

## Lines of inquiry this lane considered (recorded on the record, summarized here)

- **Pursued (taken):** audit the FEOV record's Go for genuine refusability, and grep the tree for
  the memory/dream loop's implementation state — hypothesis 3 first, then breadth. Paid off: the
  refusability is real and compiled, the memory loop is unbuilt, and capture.go is a documented
  graveyard of the exact failure the memory design would reintroduce.
- **Considered and left for a later run (deferred):** build the schema-validating write gate for the
  OKF memory profile as a runnable prototype (a Go `validate`-style refusal over a sample concept
  bundle), to settle by execution that the FEOV discipline transfers to markdown+YAML without making
  the store worse to author. Out of scope for a positioning lane; it is the operator's
  prototyping step for the implementation run.
- **Weighed and rejected:** re-deriving the external-system census (AI-Scientist, DGM, AlphaEvolve,
  FunSearch, the MAD literature) from the code side. Rejected — those systems are external and lanes
  1–2 saturated them at their own leaves; a local-repo lens adds nothing there and would duplicate,
  not broaden. Recorded so a later run does not re-walk it under a code-audit banner.
- **Weighed and rejected:** treating the memory design's `git commit per pass` as equivalent
  mediation to FEOV's record. Rejected on the B.5 disconfirmer: a commit is refusable only if a gate
  refuses it, and the design specifies no schema gate on the concept layer — only human diff review,
  which is the OpenClaw failure mode it warns against.

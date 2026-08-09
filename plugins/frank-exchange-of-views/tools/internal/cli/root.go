// Package cli assembles feov-record: the root command, the flags every seat
// shares, and the four role trees.
//
// The seat-side record runtime. ONE binary whose role subcommands carry bespoke
// verb sets, because the verb set IS the role boundary — a lens has no mint verb
// to call, blue has no board verbs at all, and the bench rules without
// originating. Seat identity is bound to its namespace, so the boundary is
// enforced rather than merely described.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/bench"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/blue"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/merge"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Version is stamped on register events and answered by --version. The setup
// preflight compares it against the plugin manifest BEFORE the run-live marker
// is written, so a skewed binary fails at setup rather than mid-round.
//
// It tracks the binary's OBSERVABLE CONTRACT, not the plugin release: bump it when a run
// would behave differently against a stale binary — the events on disk, the projections
// render writes, or the verb/flag surface a prompt may call. It is deliberately decoupled
// from the plugin version, which says what shipped.
//
//	0.2.0  the 2026-07-19 schema work — events gained `ts` and replay orders by it,
//	       findings gained a tool-assigned `finding_id`, four flags were renamed with
//	       aliases deleted, cross-references enforced at write time.
//	0.3.0  render computes `realized_open` in the board telemetry (Phase B1) — a stale
//	       binary renders telemetry a downstream reader now expects.
//	0.4.0  the global `--json` flag: mutating verbs emit a structured result and errors,
//	       so a machine consumer a prompt drives needs the binary that speaks it.
//	0.5.0  the --json envelope carries `role` and, on failure, a `code` (feov.Error) a
//	       consumer branches on — a wire-shape change a stale binary would not produce.
//	0.6.0  `bench assemble` writes report.md in-tool (verbatim sections + board risk matrix);
//	       a prompt that drives it needs the binary that has the verb.
//	0.7.0  `bench outcome` records the terminal verdict as an event, and `bench assemble`
//	       drops --inputs entirely: the report is composed from the record (event log +
//	       board) and blue's audited report, so a driving prompt needs the inputs-free verb
//	       and the outcome verb it now depends on.
//	0.8.0  `merge verdict --as PASS` is REFUSED while any gap is still open — the gate is
//	       enforced at the tool, not trusted to the seat. A prompt that drives it sees a new
//	       failure it must handle (close the gaps or issue FAIL).
//	0.9.0  the prose vocabulary collapses to --reason (+ --reason-file); --file/--text/
//	       --comment/--basis are retired, and every claim/judgment act (mint, close,
//	       closing, dispute, dispute-respond, opinion, retire, halt, certify) now requires
//	       its prose. A prompt driving the old flags sees unknown-flag and new refusals.
//	0.10.0 `bench assemble` composes a materially different report.md: a top "## Read this
//	       first" orientation (open gaps ranked from the board + the bench's certify/halt
//	       voice promoted), the blue-report embed carries ONLY blue's non-composed remainder
//	       (its lifted and tool-owned sections are dropped, killing the lift-AND-embed
//	       duplication and the stale-verdict contradiction), and lens findings the merge did
//	       not mint are surfaced. A stale binary silently produces the duplicated, contradicted
//	       report the preflight must now refuse.
//	0.11.0 the root `verify` command — a read-only cross-check of a run's record against its
//	       invariants (gaps disposed, refs resolve, PASS closed everything, seats registered
//	       first) plus the authoritative tally. Operator/CI-facing, not seat-driven, but the
//	       run-verification workflow depends on the binary having it, so a stale binary fails
//	       the check step.
//	0.12.0 the root `graph` command (render a run's actual behaviour from the record as
//	       mermaid/dot); and `bench assemble` now renders the dispute / dispute-respond /
//	       petition-rule prose it silently dropped on a payload-key mismatch (A1–A3), and
//	       surfaces the friction log. A stale binary produces the graph-less, prose-dropping
//	       report the record now expects.
//	0.14.0 #62 Stage 2 — grade disputes onto the record: the `debate` view now renders the
//	       dispute / dispute-respond thread (evidence + rationale), which since Stage 2 lives
//	       ONLY on the record (the debate.js envelope carries routing refs — gap_id/dimension/
//	       proposed/response — not the prose). A stale binary renders a debate view missing the
//	       dispute argument the bench now rules on from the record.
//	0.13.0 confidence wired end-to-end as a NON-AUTHORITATIVE signal: `bench assemble` renders a
//	       "Blue's confidence self-assessment" section and the `debate` view carries a per-round
//	       "BLUE CONFIDENCE" block, both composed from the (previously write-only) confidence
//	       event. It sets no grade and never enters the risk matrix. A stale binary drops blue's
//	       calibration the record now carries.
//	0.15.0 the root `count-claims` command computes claim_count deterministically from
//	       blue/report.md (#70), and `revision --claim-count` is DROPPED. The number that sizes
//	       red's dispatch and arms the retire-vs-drop detector was hand-counted by the blue LLM
//	       (two honest merges diverged 2x); a prompt now runs the tool and relays it. A stale
//	       binary lacks the command the blue prompt calls and still accepts the removed flag.
//	0.16.0 #62(1) citations onto the record (#71): lenses record each verified claim as a `cite`
//	       event instead of hand-writing citation-ledger.md (which the renderer already projected
//	       AND clobbered). The board JSON and `verify` now tally `citations`, and red sets
//	       citations_checked from the board's count instead of self-reporting it (fabricated on
//	       haiku). A stale binary lacks the citations tally the merge prompt now reads.
//	0.17.0 #62(1) findings onto the record (#71.2): lenses record findings as `finding` events
//	       with a TOOL-assigned run-unique label (L{role}-F{N}); `--label` is DROPPED and `--key`
//	       added (a local retry handle, mint-parity idempotency). The new `findings` view serves
//	       the findings as structured JSON — the merge coalesces from it and scorecards attribute
//	       from it, retiring the red/candidates/*.md channel; a gap's found_by names labels. A
//	       stale binary lacks the findings view and still accepts the removed --label.
//	0.18.0 the root `setup` command (#121 slice 1) — the run-setup mechanics ported from
//	       setup-research-run.mjs: skeleton, pin validation, memory/law/scorecard mirrors, the
//	       run-config + run-live marker, and the record-binary preflight. /research step 2 now
//	       calls `feov-record setup` instead of `node …setup-research-run.mjs`, so a stale binary
//	       lacks the command the setup step depends on (same rationale as 0.11.0's `verify`).
//	0.19.0 the root `cost` command (#121 slice 2) — the cost audit ported from cost-audit.mjs:
//	       per seat-round token/dollar table, tier check (#111), and the board-telemetry join.
//	       /research's capture step now generates cost.md via `feov-record cost` when it has the
//	       binary, so a stale binary lacks the command that path depends on.
//	0.20.0 the root `scorecard` command (#121 slice 3) — a chair's in-run self-read ported from
//	       scorecards.mjs (compute + renderChair), reading the record in-process. debate.js's
//	       in-run self-read prompt now runs `feov-record scorecard --run --chair` when binDir is
//	       set, so a stale binary lacks the command that self-read depends on.
//	0.21.0 the root `dashboard` command (#121 slice 4) — the live run dashboard ported from
//	       render-run-dashboard.mjs (buildModel + renderHtml), reading the record in-process.
//	       /research's monitoring step now runs `feov-record dashboard --watch`, so a stale
//	       binary lacks the command that step depends on.
//	0.22.0 the root `capture` command (#121 slice 5) — the post-hoc capture auditor ported from
//	       capture-research-run.mjs; the final .mjs port, debate.js now the only engine script.
//	       /research's capture step now runs `feov-record capture <run> <transcript>` (no --bin —
//	       the command IS the tool), so a stale binary lacks the command that step depends on.
//	0.42.0 a REPLY NAMES WHAT IT ANSWERS. merge dispute-respond recorded gap_id and the
//	       response and NOT the dimension — while blue disputes a single grade and may contest
//	       more than one on the same gap. Demonstrated live: severity and impact both disputed,
//	       one "rejected" naming neither. The ORCHESTRATOR always had it (debate.js matches on
//	       (gap_id, dimension) and reads the grade at that dimension) — off the ENVELOPE, which
//	       evaporates with the session, while the durable record kept the lossier half. The
//	       ephemeral channel was more precise than the permanent one. --dimension is now
//	       REQUIRED, and requirePriorDispute matches the PAIR rather than the gap, so answering
//	       a grade nobody contested is refused by the same rule that already refused answering
//	       a gap nobody disputed. A stale binary records unaddressed answers.
//	0.41.0 an avenue ruling has a RECORDED consequence (#246 remainder). Red's ruling and the
//	       avenue's fate were both on the record and joined NOWHERE, so blue pursuing a line
//	       red called out-of-scope looked exactly like pursuing one red endorsed. The design
//	       says a ruling is an ARGUMENT blue may contest "the way we would any other dispute" —
//	       and there is no such channel: blue dispute --id A1 is refused outright, because
//	       disputes are gap-shaped and take grade dimensions. So a ruling could only be
//	       silently obeyed or silently ignored. The MOVE is the contest, and it now records
//	       contests_ruling when blue pursues a line red ruled out-of-scope or too-thin, so the
//	       disagreement is a fact rather than something a reader reconstructs from two lists.
//	       A stale binary records the pursuit and loses the argument.
//	0.40.0 a RETIREMENT is evidenced, not asserted. blue retire recorded whatever it was told:
//	       nothing confirmed the claim had ever been in the report, and nothing confirmed it
//	       had left — so "substance leaves only through the retire verb", the rule its own doc
//	       calls "strictly stronger than the prose rule it replaces", rested on the seat's word.
//	       A phantom retire is WORSE than uninformative: the scorecard computes
//	       unrecorded_claim_loss as the drop in claim_count MINUS the retire events, so retiring
//	       a claim that was never there subtracts from the accounted side and CANCELS REAL
//	       LOSS, blinding the one detector built to catch silent deletion. The verb now REFUSES
//	       a claim still present (retiring EXPLAINS a removal, it does not perform one — edit it
//	       out first) and records removal_basis: verified | asserted, where verified means the
//	       claim appears in the old span of a recorded edit so the record can SHOW it leaving.
//	       A stale binary records unevidenced retirements.
//	0.39.0 the terminal VERDICT is DERIVED, not claimed (#308). It was the last big
//	       derived-not-asserted violation and it sat on the most consequential value the
//	       engine emits: debate.js computed it from red's ENVELOPE JSON field, stated it in the
//	       assembler's prompt, and the seat typed it back through `bench outcome --as`. Two
//	       independent channels carried the same fact — the envelope red returns and the
//	       `merge verdict` event on the record — and NOTHING reconciled them. The tool now
//	       computes the verdict from the record (a halt outranks a pass; a PASS event is
//	       VERIFIED; rounds against the recorded ceiling are CEILING) and REFUSES an --as that
//	       disagrees, naming both. The event carries verdict_basis: derived | asserted, on the
//	       same footing as fix_basis and proof_basis. Exactly one case stays asserted — a judged
//	       deadlock, whose determination lives only in the bench envelope and leaves no trace
//	       (#289) — and debate.js cannot currently produce one, so every verdict today is
//	       derived. --max-rounds is now REQUIRED at setup for the same reason the model tiers
//	       are: the ceiling is the bound CEILING is derived against, and a run that does not
//	       record it cannot decide its own outcome. A stale binary records an unchecked verdict.
//	0.38.0 RED CAN ASK FOR A COMPUTATION (#277). `merge mint --check-kind document |
//	       computation | source` is REQUIRED — what would SETTLE the acceptance check. The
//	       measured cause of zero proofs across six runs was never blue's reluctance: in the
//	       2026-08-05 smoke all TEN of red's checks were document probes (R1-1 was literally
//	       "execute the assembly step"), so red could only ask whether the report SAYS
//	       something, and a full round went on R1-2/R2-2 refusing an asserted test that three
//	       lines of trial division settle. An OPTIONAL field would have been answered the way
//	       avenue --hypothesis was before it was required, which is not at all. A `computation`
//	       gap now CANNOT be closed until a proof answers it: `blue prove --answers <gap>` is
//	       the join, checked against the board at write time like every other reference. A
//	       stale binary rejects --check-kind outright, and its mints are refused for the
//	       missing required field.
//	0.37.0 the estoppel-rejection count is keyed on a FIELD, not on prose (#283). It read
//	       `strings.Contains(frictionText, "estoppel <emdash>")`, which returns ZERO both when the
//	       guard never fires and when anyone rewords the refusal — and zero is printed to an
//	       operator as "red behaved". It also COUNTED UP on a seat quoting the refusal in its
//	       own complaint, so the measurement could be wrong in both directions at once. The
//	       refusal now carries kind=estoppel and estopped_by=<gap> on the friction event, and
//	       the count reads the field. A stale binary writes friction without the field, so its
//	       rejections read 0 — the honest answer for events that never recorded the fact.
//	0.36.0 the run directory is INJECTED, not typed (#281). The first live run logged 55
//	       tool-call errors in 534 executions and TEN were one flag; `InferRunDir` was added
//	       for that and never fires, because the prompt hands the seat an absolute --run at
//	       every call site and an explicit flag always wins. Worse than absence is a WRONG
//	       value: the 2026-08-05 smoke's blue-respond-r1 typed `special circumstances` for
//	       `special-circumstances`, believed the tool's "no such gap" answer, filed friction
//	       against an innocent rule and abandoned five manifest receipts. Now the PreToolUse
//	       hook rewrites a Bash command invoking feov-record in COMMAND POSITION (never a
//	       mention — a grep, a commit message or a heredoc is left alone, and a command
//	       containing a heredoc is not rewritten at all), prefixing `export FEOV_RUN=...;`,
//	       and a seat verb resolves FEOV_RUN -> --run -> inference. A --run DISAGREEING with
//	       the injected value is REFUSED naming both, which is the whole point: the typo now
//	       fails loudly at the one place that can see both values. Deny still wins
//	       structurally. A stale binary ignores FEOV_RUN and silently obeys the typo.
//	0.35.0 the STOPPING SIGNAL (#284): "is this close enough" is the bench's job and had no
//	       organ — red is not incentivised to stop finding things, blue is not incentivised to
//	       stop being found out, and maxRounds is a cost ceiling the protocol already says is
//	       never the terminator of record. New `--view telemetry` gives the bench the per-round
//	       SERIES through the tool for the first time (it existed only in the dashboard and
//	       cost.md), and the line now carries new_mint.by_class plus class_repeat_rate against
//	       the previous round. Severity answers "how bad" and can sit flat for rounds while the
//	       KIND of defect moves from "wrong about the world" to "this artifact disagrees with
//	       itself" — the phase change measured across two audit loops (the #251 spec, five
//	       FAILs; plans/feov-run-injection.md, six) where the findings never stopped, they
//	       changed character. Class was already a required, registry-validated mint field, so
//	       this is an aggregation and never a new assertion. A stale binary answers "unknown
//	       view" to the read the bench's constitution now tells it to make.
//	0.34.0 PROOF — the engine may compute an answer, not only cite one (#277). Every seat has
//	       had Write/Edit/Bash since it shipped and across six runs NO SEAT EVER WROTE A
//	       PROGRAM to settle a question: never tied, never invited. Measured cost — the smoke's
//	       R1-2 asked for the protocol to be tested on a false claim, blue answered in PROSE
//	       that it had been, and R2-2 refused it for showing no evidence: a round, for three
//	       lines of trial division. New `blue prove --location --script [--cites]` runs the
//	       script TWICE, caches script+output under <run>/proofs/<sha256>, and splices an
//	       invisible `<!--proof:p-…-->` anchor — a third class on the one immortal-anchor
//	       mechanism, protected by the same lockdown. The basis is DERIVED: identical output is
//	       `reproducible`, moving output is `observed` (still evidence, of a system in motion,
//	       and the drift is named). A non-zero exit is a RESULT, not an error — disproving is
//	       worth as much as confirming. New `lens reproduce --id <sha256>` re-EXECUTES and
//	       compares: the one audit that does not end in believing bytes someone else chose.
//	       It is the citation model with the last mile walked — cite the METHOD, compute the
//	       INSTANCE. A stale binary lacks both verbs.
//	0.33.0 the avenue LIFECYCLE (#246): an avenue was a one-shot append — measured over 86
//	       events in six runs, ZERO lines were recorded twice, ZERO statuses ever changed, and
//	       83 of 86 landed in round 0. So "pursued" meant "I intend to", nothing could falsify
//	       it, and a direction that died mid-run had no way to say so; the rejection rate that
//	       produced measured only what blue could rule out BEFORE STARTING. Now: the tool
//	       assigns an id (A1, A2 …), `--hypothesis` states what would be true if the line pays
//	       off, and `blue avenue --id A1 --status <new> --reason` MOVES it — the move is the
//	       evidence of choosing the old shape could not record. New `merge rule-avenue --id
//	       --ruling endorsed|out-of-scope|too-thin --reason` is red's fate for a proposed
//	       direction (red rejected NONE of 86 across six runs because it had no verb to); a
//	       ruling is an argument blue may contest through `blue dispute`, never a command, and
//	       red still never PROPOSES — that is a gap's required_fix. lines-of-inquiry renders
//	       the path and the ruling, and surfaces avenues still owing a decision. A stale binary
//	       lacks rule-avenue and the move form.
//	0.32.0 ESTOPPEL (#267 stage 4, the last): red is bound by the fix it prescribed. `blue edit`
//	       records applied_verbatim when its old/new are byte-identical to the answered gap's
//	       fix_old/fix_new — the tool comparing bytes, never a seat claiming it. `merge mint`
//	       then REFUSES a fresh gap whose --location is that text, and LOGS THE REFUSAL AS
//	       FRICTION: a guard that silently blocks is invisible, and this makes "how often red
//	       relitigates its own prescriptions" a per-run number. Red is not silenced — naming
//	       the estopping gap in --supersedes lets the amendment through, which is what the guard
//	       asks for. It binds ONLY verbatim text: a counter-edit is blue's authorship and red
//	       audits it normally. `--view changes` now prints the two falsifying numbers, offered/
//	       applied/declined and the estoppel-rejection count, and says so when blue declined
//	       nothing. A stale binary neither records the fact nor enforces the guard.
//	0.31.0 concrete fix proposals + derived fix_basis (#267 stage 3): `merge mint` takes an
//	       OPTIONAL `--fix-old`/`--fix-new` pair for TEXTUAL defects. The tool validates it against
//	       the live report — present, unique, no word split, anchors transit unchanged — so red
//	       cannot state one without reading the text it prescribes against; that forced re-read is
//	       the check whose absence produced 3 of 3 round-2 gaps in the 2026-08-04 smoke. A
//	       replacement more than 120 characters longer than its span is REFUSED as authoring
//	       (measured: every prescribed ADDITION in that smoke was +285 or more, textual repairs
//	       +60 or less). `fix_basis` is DERIVED from the pair validating — there is no flag to
//	       claim it. The checks moved to a new internal/bluedoc so `blue edit` and `merge mint`
//	       ask one implementation, not two. A stale binary rejects --fix-old, failing the mint
//	       red's prompt now instructs.
//	0.30.0 the `changes` view (#268): the `blue_edit` diff stack has been recorded since 0.27.0
//	       and NO projection rendered it — written by one seat, read by nobody, while `changelog`
//	       rendered faithfully from a `revision` event no seat emits. `show --view changes` lists
//	       every recorded edit with the gap it answers; `--view changes --id <gap>` prints red's
//	       required_fix and acceptance_check ABOVE the exact old->new spans blue recorded against
//	       them — the comparison that replaces red inferring whether its prescription was answered.
//	       --id is a hard error on any other view. Red's prompts now point at it: the lens's
//	       navigation hint moves off the hand-written blue/CHANGELOG.md onto the derived view, and
//	       the merge is told to compare before closing. A stale binary rejects `--view changes`,
//	       failing the read both prompts now instruct.
//	0.29.0 edit provenance (#267): `blue edit` takes `--answers <gap-id>` — the gap the edit
//	       responds to, validated against the board like every other reference. It replaces a
//	       CONVENTION: 19 of 26 edits in the 2026-08-04 smoke opened --reason with the gap id and
//	       7 did not, so the link that every #267 measurement joins on existed 73% of the time.
//	       Prose naming a real gap while --answers is empty is now REFUSED, so the key cannot
//	       decay back into prose. A stale binary rejects --answers outright, which fails every
//	       edit the seat prompt now instructs — the hard break this version gate exists for.
//	0.28.0 the bibliography core (#256): a citation axis symmetric to the finding markers. A new root
//	       `fetch --run --url` is a cached, hash-verified web read that REPLACES WebFetch — fetch once,
//	       cache to `<run>/cache/<sha256>`, serve every seat the same bytes (a second fetch is a cache
//	       hit; red re-reads exactly what blue cited). A new `blue cite --location --url --title` fetches
//	       through the cache and splices an INVISIBLE, IMMORTAL `<!--cite:c-…-->` anchor at the quoted
//	       sentence (reusing the `lens.InsertAnchor` machinery); an unreachable source is rejected + auto-
//	       friction. The lockdown's `blue edit` guard now protects citation anchors too (class-swept to the
//	       fx∪cite union), so the set of cite events is a strict bijection with the doc's anchors. claimcount
//	       counts the cite anchor as the claim unit (the `[^label]` footnote is retired); assembly weaves
//	       the anchors into visible `[^N]` footnotes + a composed `## Bibliography`; a scorecard
//	       `unbacked_citations` detector flags a broken bijection. A stale binary lacks `fetch`/`blue cite`,
//	       so citations are unmanaged and the weave/detector silently do not engage.
//	0.27.0 the blue-report lockdown + append-only diff-stack (#250/#251): blue/report.md is READ-ONLY to
//	       every seat except the allowlisted author `agent_type` (frank-exchange-of-views:blue-synthesizer).
//	       A new `blue edit --old --new --reason` is the ONE write path for a response seat — an
//	       exact-on-stripped span replace (extends the finding-marker matcher with `lens.LocateSpan`) that
//	       REJECTS an edit dropping/spanning a `<!--fx:…-->` marker and appends a `blue_edit` diff-stack op
//	       (event-first, crash-reconciled). A new `hook pretooluse`/`hook posttooluse` command group backs
//	       the plugin's first hooks.json: PreToolUse denies a non-author raw write to report.md (Edit/Write
//	       by file_path, common Bash by target), PostToolUse blocks + forces repair on a dropped marker via
//	       any mechanism (reusing the shared claimcount.MissingAnchorIDs check). A stale binary lacks
//	       `blue edit` and the hook backends, so the lockdown silently does not engage.
//	0.26.0 finding-markers (slice 1b): a `lens finding` now REQUIRES --location + --reason, VALIDATES
//	       that the quoted --location text is present in blue/report.md (a mis-quote is rejected), and
//	       anchors the finding by inserting an invisible immortal footnote marker `[^<finding_id>]` at
//	       that sentence — written under a lock via the new `record.MutateBlueReport`, with a new
//	       `anchor` event as the record of it. claimcount routes the `f-` namespace out of the claim
//	       count; assembly strips every `[^f-…]` so none ships; a scorecard detector flags an anchored
//	       marker that vanished from the report (blue dropping red's audit point). A stale binary would
//	       record findings without anchoring them and miss the tampering detector.
//	0.25.0 blue's read-once + the existence write-path (Envelope-refs Phase 3): a new read-only
//	       `blue claim-index` enumerates, per footnote label, every site a claim appears
//	       ({label, occurrences:[{heading, line, sentence_hash}]}) so blue propagates a correction to
//	       all sites from one query instead of re-reading the whole report; it reuses count-claims'
//	       exclusion rule (one scanner, two readings). And `merge mint` gains `--existence
//	       verified|suspected`, the write path for the leaf-check axis the board already reads. A stale
//	       binary lacks claim-index and drops the existence axis on the floor.
//	0.24.0 the merge gets its shrinking working set: a new `worklist` view (OPEN gaps only, lean —
//	       grades + class + location + a problem synopsis + found_by — plus a prose-free closed_index)
//	       becomes `merge show`'s default, and a new `merge near-match --candidate "<text>"` screens a
//	       candidate against the board (open + closed) by lexical overlap so a near-duplicate surfaces
//	       as a reopen rather than a fresh gap. The screen is read-only and moves off the seat: the
//	       tool ranks, the seat decides. (Envelope-refs Phase 2.)
//
//	0.23.0 the materialized-projection cache is eliminated: the standalone `render` verb is REMOVED
//	       and projections are no longer written to disk. Every markdown `show` view and the telemetry derive
//	       just-in-time from the record via the shared internal/view library; `merge verdict`
//	       checkpoints the event log only. A prompt that called `<seat> render` or `merge render`
//	       gets an unknown-verb refusal, and debate.js no longer issues those calls, so a stale
//	       binary still exposes a verb the engine has stopped driving.
//
//	0.43.0 `lens observe` and `merge dispose` are REMOVED, with their events (#327). A finding's
//	       fate is COALESCENCE and nothing else — its label credited in a gap's found_by — so the
//	       Observation replay no longer carries a Disposition and `undisposed_observations`
//	       becomes `uncredited_findings`, a number that can still be non-zero. A stale binary
//	       still ACCEPTS both verbs and writes events this replay drops, which is why the
//	       contract version moves: setup must refuse it.
//
// versionsync_test.go asserts this equals recordToolVersion in the plugin manifest, which
// is what setup preflights against. Without that test the two drift and the preflight
// compares a stale number to itself.
const Version = "0.43.0"

func init() { record.ToolVersion = Version }

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "feov-record",
		Short: "the seat-side record runtime (the verb set IS the role boundary)",
		Long: `feov-record — the seat-side record runtime (the verb set IS the role boundary)

A lens structurally cannot mint or close a board gap: no such verb exists in its
namespace. Blue has no board verbs at all. The bench rules and never originates.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// The two flags EVERY verb needs, declared once and inherited. Persistent
	// flags are the mechanism the first cut of this CLI re-implemented by
	// re-declaring --run and --seat-id on all sixteen verbs.
	root.PersistentFlags().String(flags.Run, "", "the run directory — OPTIONAL: the engine injects it, and a value that disagrees with the run you were dispatched into is refused rather than obeyed")
	root.PersistentFlags().String(flags.SeatID, "", "your seat id, assigned by the engine and bound to this role's namespace")
	// --json makes every mutating verb emit a structured result and every failure a
	// structured error, so a machine consumer parses fields instead of prose. On READS the
	// format is primarily view-selected: board/findings/friction are JSON by name, the rest
	// are markdown. --json opts a markdown view into its structured form where one exists
	// (today only `show --view debate --json`); it is an error on a JSON-by-name view or a
	// markdown view with no JSON form, so there is exactly one way to reach each form.
	root.PersistentFlags().Bool(flags.JSON, false, "emit a structured JSON result (and structured errors) instead of human text")

	root.AddCommand(
		lens.NewCommand(),
		merge.NewCommand(),
		blue.NewCommand(),
		bench.NewCommand(),
		newVerify(),      // operator cross-check, not a seat role — read-only over the record
		newGraph(),       // operator: render a run's actual behaviour from the record
		newCountClaims(), // operator/blue: deterministic claim_count over blue/report.md
		newFetch(),       // operator: cached, hash-verified web read (replaces WebFetch) — feeds the run source cache
		newSetup(),       // operator: build a research run's blackboard (ported from setup-research-run.mjs)
		newScorecard(),   // operator: a chair's in-run self-read scorecard (ported from scorecards.mjs)
		newDashboard(),   // operator: the live run dashboard.html (ported from render-run-dashboard.mjs)
		newCapture(),     // operator: the post-hoc capture auditor (ported from capture-research-run.mjs)
		newHook(),        // hook backend: the blue-report lockdown PreToolUse/PostToolUse gates (invoked by hooks.json)
	)
	return root
}

// Execute runs the CLI. Abort-safety is armed first: a seat killed mid-command
// must lose an event, never leave a torn one or a stuck lock.
func Execute() {
	defer record.InstallSignalGuard()()

	if err := newRoot().Execute(); err != nil {
		// Errors print bare, without cobra's usage dump: a validation refusal here
		// is a TEACHING message a seat reads and acts on, and burying it under a
		// flag listing is how it stops being read.
		fmt.Fprintf(os.Stderr, "feov-record: %v\n", err)
		os.Exit(2)
	}
}

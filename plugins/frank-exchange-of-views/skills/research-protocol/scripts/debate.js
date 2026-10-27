export const meta = {
  name: 'frank-exchange-of-views',
  description: 'Adversarial research debate: additive blue builds, gate-keeping red audits, judged termination, union-not-summary assembly',
  phases: [
    { title: 'Frontier', detail: 'hypotheses before searches' },
    { title: 'Blue', detail: 'best-of-N lanes + additive synthesis' },
    { title: 'Red', detail: 'per-lens audits + merged verdict' },
    { title: 'Debate', detail: 'sittings dispatched from the record until nobody is ready — PASS permitted, or the epoch limit reached — or until the plan stops moving' },
    { title: 'Assemble', detail: 'final report by union' },
  ],
}

// TESTING & MODEL-SELECTION STRATEGY (learned from runs 1-3, ~5M tokens of tuition):
//   Logic bugs are unit-testable for zero tokens: a Node harness stubs agent() with canned
//   envelopes and drives every branch (args parsing, dispatch loop, contested docket, impasse,
//   ceiling, null returns). Founding regressions: stringified args -> undefined paths (run 1);
//   missing null-guard on agent() returns (run 2); lineage-blind docket + degenerate
//   FAIL-with-empty-gaps (run 3 retrospective, report §3 rows 20-24).
//   Behavior needs live agents — /research --smoke (1 lane + laneFloorOverride, 1 round,
//   model=haiku) exercises the pipeline for ~50k tokens.
//   Model tiers are REQUIRED — both `model` and `judgmentModel` must be set explicitly. The engine
//   throws rather than guess a tier or inherit the session model: a silently expensive (or silently
//   cheap) tier was the #111 trap. sonnet for development; --smoke sets BOTH to haiku. NEVER change
//   `model` OR `judgmentModel` on a resume — they change agent() opts, bust the cache keys, and
//   re-run completed sittings at full price.
//   Per-role split (efficiency doctrine: cheapen redundancy and mechanics, never judgment or
//   the adversary): `model` drives the BULK seats (frontier, blue lanes, red lenses, blue
//   responses); `judgmentModel` drives the JUDGMENT seats (blue-synthesize, red-chair,
//   lead-judge, assemble). Neither inherits — every run, keeper or dev, names both.
//   KNOWN TRADEOFF (retrospective §3 row 16b): red LENSES ride the bulk tier — on a cheap-model
//   dev/smoke run, treat lens-sourced gap grades with a confidence discount. For keeper runs,
//   name a STRONG model for BOTH tiers so the adversary and the bench run at full strength.
//   THE TIER YOU NAME IS NOT NECESSARILY THE TIER THAT ANSWERS, and until #589 nothing in the
//   loop could tell. Both 2026-08-23 runs asked for `claude-fable-5` on the bulk tier and were
//   answered by `claude-opus-4-8` on all 44 bulk seats, for ~$379: the configured research tier
//   never ran, and one run's certified report described the pairing it had asked for rather than
//   the one that argued it. These opts are a REQUEST. `register` now measures what actually
//   replied — the harness declares a swap on the seat's own first turn — records it on the
//   register event, and REFUSES the sitting on a mismatch in either direction, so the run stops
//   at its first seat rather than at capture. An operator who accepts the environment's
//   substitution says so once, at `setup`, and the run proceeds with it on the record.
// TERMINATION IS JUDGED, AND THE STANDING PRACTICE IS STOP-AND-RESUME (run-4 report §1.4-1.5):
//   the operator stops the run and resumes it — Workflow({scriptPath, resumeFromRunId}) with the
//   SAME args — and the agent cache replays every completed sitting at ~$0; only what had not run
//   runs live. The run's bounds are its TERMS in inputs/run-config.json (k-max, mint-budget,
//   max-epochs), never a count passed here; the automatic severity-floor stop was REJECTED by the
//   run-4 debate (it automates the one call that belongs to judgment). NEVER change
//   model/judgmentModel on that resume.
// The lead is a script: mechanics, dispatch, termination. All file writes
// belong to the agents (the filesystem is the blackboard; the script has no
// filesystem access by design). Judgment calls go to lead-judge, never round-to-round.
// Defensive arg handling: args may arrive JSON-encoded (resume path); a stringified
// object destructures to undefined and every agent gets literal 'undefined' paths —
// the exact harness defect red graded high/high/low in run 1. Parse, then guard.
const a = typeof args === 'string' ? JSON.parse(args) : args
const { topic, runDir, lanes = 3, model = null, judgmentModel = null, laneFloorOverride = null, binDir = null, transcriptDir = null } = a
if (args && Object.prototype.hasOwnProperty.call(args, 'maxRounds')) {
  throw new Error(`debate: refusing dispatch — maxRounds is not a term of this engine any more (plans/roundless.md). A run ends on the record, not on a count:
the chair's \`dispatch next\` says who sits, and the run ends when nobody is ready — PASS permitted — or at the run's
epoch limit (CEILING). The bounds are the run's TERMS, recorded by setup in
inputs/run-config.json: the exchanges a gap gets before impasse (k-max), the floor of the gaps a lens may hold open at once (mint-budget),
raised with the report's size, and the chair sittings the run gets (max-epochs); setup's help names the words. Drop maxRounds from args.`)
}
if (!topic || !runDir || String(runDir).includes('undefined') || String(topic) === 'undefined') {
  throw new Error(`debate: refusing dispatch — topic/runDir unbound (topic=${JSON.stringify(topic)}, runDir=${JSON.stringify(runDir)})`)
}
// Model tiers are REQUIRED and never inherited — the engine will not guess a tier (#111). A missing
// tier is named explicitly with its role and a remedy; a silent expensive/cheap seat is the trap
// this closes. Both messages begin "refusing dispatch" so the same-shaped regression guard catches them.
if (!model) {
  throw new Error(`debate: refusing dispatch — model unset. The engine does not guess a tier: pass model (the BULK tier — frontier, blue lanes, red lenses, blue responses), e.g. { model: "sonnet" }.`)
}
if (!judgmentModel) {
  throw new Error(`debate: refusing dispatch — judgmentModel unset. The engine does not inherit the session model: pass judgmentModel (the JUDGMENT tier — blue-synthesize, red-chair, judge, assemble), e.g. { judgmentModel: "sonnet" }.`)
}
if (!binDir) {
  throw new Error(`debate: refusing dispatch — binDir unset. The engine does not run without the record.

The tool is the contract, not an enhancement: every seat writes its acts through \`feov-record\` and reads the
board back through it, and binDir is how a seat is told where that binary is. Without it no seat can reach
the record, so the run would record nothing and every gate would still pass — the plausible zero with a whole
run behind it. It is refused here for the same reason model and judgmentModel are: a decision the engine
cannot make correctly by guessing must be stated.`)
}

// Lane floor (retrospective §3 row 7): run 2 silently ran under-provisioned at lanes=2.
// Below 3 lanes a hypothesis loses dedicated attention; override requires a stated reason
// (e.g. laneFloorOverride: 'smoke run — pipeline exercise only').
if (lanes < 3 && !laneFloorOverride) {
  throw new Error(`debate: lanes=${lanes} is below the floor of 3 — pass laneFloorOverride: '<reason>' to run under-provisioned deliberately`)
}
// Unconditional now — the guard above rejected an unset tier, so both are set.
const TIER = { bulk: { model }, judgment: { model: judgmentModel } }

// WHO SITS, AND UNDER WHAT, IS THE RECORD'S ROSTER. The block below is written from it and this
// script authors none of it: each seat's role, tier class, agent configuration and lens area
// (seatclass.Seats, seatclass.Lane), the bench's occasions (the record's Occasion enum) and the
// mass mapping's stamp (record.MassMappingVersion). Adding a lens area is an edit to flags.LensAreas, its agent
// configuration, and a regeneration; a block that is stale or edited by hand fails
// `enginegen -check`.
//
// A SEAT ID IS A NAME. Read a seat's role, tier, configuration and area from its row; never off
// its id. A lane has no row because the run's lane count decides how many there are: its id is
// LANE.prefix and its number, and every lane is dispatched as LANE says.
// BEGIN GENERATED SEATS — go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen writes this from the record's roster. DO NOT EDIT BY HAND.
const SEATS = {
  'blue-respond': { role: 'blue', tier: 'bulk', agentType: 'frank-exchange-of-views:blue-researcher' },
  'blue-synthesize': { role: 'blue', tier: 'judgment', agentType: 'frank-exchange-of-views:blue-synthesizer' },
  'frontier': { role: 'blue', tier: 'bulk', agentType: 'frank-exchange-of-views:blue-researcher' },
  'judge': { role: 'bench', tier: 'judgment', agentType: 'frank-exchange-of-views:lead-judge' },
  'red-chair': { role: 'chair', tier: 'judgment', agentType: 'frank-exchange-of-views:red-chair' },
  'red-lens-adversary': { role: 'lens', tier: 'bulk', agentType: 'frank-exchange-of-views:red-lens-adversary', area: 'adversary' },
  'red-lens-architecture': { role: 'lens', tier: 'bulk', agentType: 'frank-exchange-of-views:red-lens-architecture', area: 'architecture' },
  'red-lens-computation': { role: 'lens', tier: 'bulk', agentType: 'frank-exchange-of-views:red-lens-computation', area: 'computation' },
  'red-lens-dark-side': { role: 'lens', tier: 'bulk', agentType: 'frank-exchange-of-views:red-lens-dark-side', area: 'dark-side' },
  'red-lens-evidence': { role: 'lens', tier: 'bulk', agentType: 'frank-exchange-of-views:red-lens-evidence', area: 'evidence' },
  'red-lens-logic': { role: 'lens', tier: 'bulk', agentType: 'frank-exchange-of-views:red-lens-logic', area: 'logic' },
  'red-lens-voice': { role: 'lens', tier: 'bulk', agentType: 'frank-exchange-of-views:red-lens-voice', area: 'voice' },
}
const LANE = { prefix: 'blue-lane-', role: 'blue', tier: 'bulk', agentType: 'frank-exchange-of-views:blue-researcher' }
const OCCASION = { docket: 'docket', petition: 'petition', terminal: 'terminal', assemble: 'assemble' }
const MASS_MAPPING_VERSION = 'v2'
// END GENERATED SEATS
const roleOfSeat = (id) => (Object.prototype.hasOwnProperty.call(SEATS, id) ? SEATS[id].role : '')
// dispatchAs is what a seat is dispatched under — its tier's model and its agent configuration,
// both read from its row. A call site names the seat and spreads this; it writes neither.
const dispatchAs = (id) => {
  if (!roleOfSeat(id)) throw new Error(`debate: ${id} is no seat on the roster — a dispatch names a seat the record seats`)
  return { ...TIER[SEATS[id].tier], agentType: SEATS[id].agentType }
}
const laneSeat = (n) => `${LANE.prefix}${n}`
const laneDispatch = () => ({ ...TIER[LANE.tier], agentType: LANE.agentType })
// A seat's label names the run by its directory, `<date>_<slug>`, without the date: the last
// path element, and the part after the first `_` where everything before it is the date's digits
// and hyphens.
const slug = (() => {
  const name = String(runDir).split('/').filter((part) => part !== '').pop() || ''
  const cut = name.indexOf('_')
  const dated = cut > 0 && [...name.slice(0, cut)].every((c) => c === '-' || (c >= '0' && c <= '9'))
  return dated ? name.slice(cut + 1) : name
})()
log(`researching: ${topic.length > 160 ? topic.slice(0, 157) + '...' : topic}`)
log(`resolved tiers — bulk: ${model}, judgment: ${judgmentModel}`)

// A SEAT'S LOG ENTRIES LIVE ON THE RECORD AND NOWHERE ELSE. No envelope carries one and this
// script collects none: the record survives a mid-run throw, and a second copy in a return value
// is a second account nobody can hold to the first. TestNoEnvelopeRestatesARecordFact fails a
// schema that declares `log`.
// THE LOG DUTY, THE PETITION RIGHT AND THE READ-YOUR-WHOLE-SURFACE DUTY ARE IN NO PROMPT. Each is
// unconditional, so each seat's constitution states it once — the log duty and the surface duty as
// fragments under scripts/agentgen/src/fragments, the petition right in the red seats' preloaded
// skill and in blue's definition — and a prompt that restates one hands the seat two wordings of
// one rule after the next edit lands on one. TestNoPromptRestatesAConstitutionDuty holds the
// prompts to it. What a prompt adds is what only the sitting knows: the seat id, the tool's path.
//
// THE EMPTY-SITTING AFFORDANCE IS STATED TO THE SEATS THAT HAVE IT, AND THE PROMPT CEILING IS WHY.
// Measured across eight runs: of 96 sittings that recorded nothing, 93 were LENSES — the chair had
// 0 of 42 and the bench 0 of 8. A lens is woken speculatively to look for a class of defect it may
// not find; every other seat is dispatched to do a specific thing and nearly always does it. So the
// sentence goes where it is true, rather than onto every prompt where it would cost bytes on every
// turn of every sitting to tell the chair about a case it never reaches.
//
// THE MECHANICS ARE NOT HERE. How the tool derives the entry, and where the sitting is recorded
// when the seat writes nothing, is the log verb's `--help` — which the constitution already
// carries. TestNoSeatPromptGrowsPastItsCeiling asks exactly that question of new prompt text, and
// it was right: the first draft of this put a refusal's mechanics in the prompt and pushed three
// seats past their ceilings.
const emptySittingClause = ' A sitting with nothing to do runs no commands.'

// Wall-clock doctrine (run-4 forensics, 2026-07-17): 80% of run time is API rounds at ~24s
// each, and the corpus showed ZERO batched tool calls — every peek paid a full round. A
// Bash spawn additionally carries a measured multi-second fixed floor. Every seat gets this.
//
// THE CLAUSE NAMES NO SEARCH TOOL, because the tool list is the harness's and moves under the
// plugin: a session whose list holds no Grep or Glob answers "No such tool available" to a seat
// told to use them, and a seat forbidden grep through the shell then has no search at all. What
// holds on every harness is the cost it is about — a file is read through the file-reading tool,
// and a shell command earns its spawn by searching, never by paging a file.
const speedClause = ` SPEED: ONE MESSAGE IS ONE ROUND-TRIP, whatever it carries. Five independent tool calls in ONE message cost one round-trip; the same five in five messages cost five, and a round-trip is seconds. So batch every call that does not need a prior call's OUTPUT as its input: all your projections together, independent fetches together, several file reads together. Serialize only where the second call cannot be written until the first has answered. If you are about to send a message with one tool call in it, ask what else you already know you will need. READ a file with the file-reading tool (offset/limit), never sed/awk/head/tail/cat through the shell: a shell spawn costs 10-100x a native read and buys nothing. SEARCH with the search tool your session offers; where it offers none, or the tool refuses a path under the run directory ("Path does not exist" — a KNOWN HARNESS LIMIT, W1.11, do NOT log it as friction), search through the shell (grep, ls) — one command that finds the lines, never a whole file dumped to be read by eye.`

// The record IS the exchange. Every seat gets an engine-assigned SEAT_ID and every
// act it takes is an event written through the binary. The events are the only copy.
//
// The writer is ONE compiled binary with role subcommands, so the seat is handed its
// ROLE rather than a script path. The role is not decoration — seat identity is bound
// to it, and a seat
// that reaches for another role's verbs is refused. The tool's --help IS the
// seat's record contract: everything listed is permitted, anything absent does
// not exist for that seat and is a REQUEST for the log rather than something to work around.
// INQUIRIES. think-around-problem mandates exploring genuinely distinct alternatives;
// terse-communication forbids narrating them. Recording the exploration makes the rule
// instrumentable, and it preserves what a run otherwise loses: an ABANDONED inquiry is a dead end
// a future run does not re-walk, kept where the reasoning that reached it is kept.
// WHICH FATE LANDS UNDER WHICH HEADING, in the assembler's own words: this is report.InquirySections
// (tools/internal/report/layout.go), written there from the three inquiry rows, and
// TestBluesPromptsNameTheSectionsTheToolComposes fails a prompt that says it differently.
const INQUIRY_SECTIONS = 'Research areas holds each inquiry that is proposed, underway, confirmed or refuted, Future research directions each one deferred, and Alternatives considered each one declined or abandoned'
// THE SECTIONS ASSEMBLY COMPOSES FROM THE RECORD THAT THE REPORT MAY NOT CARRY, named to blue
// wherever blue is told what its report holds. One list for both prompts, because a seat reads
// such a list as the whole of it. It is the rows of the assembler's table of sections
// (report.layout, tools/internal/report/layout.go) whose heading the tool refuses in a report, in
// the table's order, and TestBluesPromptsNameTheSectionsTheToolComposes fails this list when it
// and the table differ. The table's SubjectHeading rows are not in it: the tool composes those
// sections in other documents, and a section of that name in the report is the report's own.
const TOOL_COMPOSED = ['Risk matrix', 'Research areas', 'Future research directions', 'Alternatives considered', 'Bibliography', 'The board', 'The debate', 'Citations', 'How this run was conducted', "The verdict's basis", 'Log (what the run told the operator)', 'Record verification', 'Claims retired', 'Superseded bench statements', 'Edits to the report']
  .map((h) => `\`## ${h}\``).join(', ')
// WHAT THE REPORT HOLDS OF THE APPROACHES, for the two seats that write the report's text and
// record none of them there: the synthesizer, which merges the drafts, and blue's responding seat.
const NO_LIST_OF_APPROACHES = ' The approaches the research considered are among them: each is an inquiry on the record, where what became of it is said, and the report carries no list of them.'
// THE INQUIRY PARAGRAPH IS THE LANES'. A lane does the work an inquiry names, so the fate is its
// to record. The synthesizer merges drafts and is told what the report holds of the approaches
// (NO_LIST_OF_APPROACHES), never to record one: a synthesizer handed this paragraph records
// approaches the frontier and the lanes already recorded, and the report prints one approach under
// two fates. What each fate means is the record's, on the page of the act that records it; this
// paragraph carries the duty.
const inquiryClause = binDir
  ? ` INQUIRIES: each approach the research took or weighed is on the record once, as an inquiry — the question it asks, with the hypothesis that would be true if it paid off kept apart from the question, and later with the fate it ACTUALLY had, what was done and what was found. Record the fate the approach really had; what each fate means is on the page of the act that records it. NO FATE IS OWED AN ENTRY. A sitting that weighed nothing and set it down has no declined inquiry, and that record is complete: an approach thought up, or run, so that there is one to decline is a false entry, and a declined inquiry whose reason reports a result is one too. The hypothesis is what makes the account honest, and a dead end you really reached is worth recording: one on the record is one the next run does not re-walk. THE TOOL WRITES THE REPORT'S ACCOUNT OF THEM, NOT YOU. At assembly it composes three sections of the report from the inquiries on the record: ${INQUIRY_SECTIONS}. That gives the reader what was run, what was deferred and what was set aside as three things, not one list of winners. You write none of those three headings in the report, and no list of the approaches considered and their fates under any other heading: a report that carries one of those headings is refused, and a hand copy elsewhere is a second account nothing checks against the record. The question you propose, its hypothesis, and the method and reason you give with its fate ARE PRINTED IN THE REPORT word for word, each under its own label, so write them as research prose for a reader of the SUBJECT — the question the inquiry asks and what was found about it, never the steps you took. THE QUESTION HOLDS NO PREDICTED RESULT: what you expect to find is the hypothesis, and a question that states a result is printed as the question whatever the research then finds. Recording each approach with the fate it had is what makes "explore the alternatives" checkable instead of self-attested.`
  : ''


 // STEELMAN DUTY (E0.5h): the sections red NEVER gap-anchors are exactly the
// self-critical ones — disconfirming passes, human-gated paths, self-attested
// inventories. Claims AGAINST the design attract no adversary, so the case
// against goes unaudited while the case for is contested every round. Blue's
// recorded inquiries are that surface made concrete: what was declined,
// and whether the stated reason survives contact.
const steelmanClause = ` STEELMAN DUTY: read the exploration space via the \`inquiries\` projection (the tool renders it fresh from the inquiry events on the record). Blue's DECLINED and ABANDONED inquiries are the case AGAINST its own design, and nobody audits them — a weak reason for declining a strong alternative survives untouched because it reads as humility. Attack the reasons, not just the conclusions: a declined inquiry whose stated reason does not hold is a finding, and so is an abandoned one whose obituary is wrong — or a refuted one whose reason does not refute its hypothesis. RECONCILE THE RECORD AGAINST THE DOCUMENT: the inquiries the report actually took must be the ones on the record. Assembly prints every inquiry in the report, with its fate and its reason, in sections the report's text does not hold: read them in the report's assembled form before you grade an inquiry. Where the two tell the reader different things — one approach under two fates; a fate the report's own text or a proof contradicts, as an approach printed as declined that a section carries; a section built on an approach that is no inquiry — that is a candidate gap, anchored to the inquiry and weighed like any other: by what it changes for a reader, who decides by what was tried as well as by the answer. A contradiction left as a finding ships: a finding about an inquiry is on no seat's work list and obliges nobody. An inquiry still at \`proposed\`, or at \`underway\` with NO fate recorded this sitting, is a decision nobody made, and that is a finding. Re-recording \`underway\` WITH what it learned is a legitimate reaffirmation and settles the inquiry for that sitting — do not read it as neglect; \`deferred\` is likewise a decision (kept for a later run), not an omission. An inquiry declared in the first epoch and never revisited is the common case, and there 'underway' means only 'I intend to' — nothing falsifies it.`


// INTEGRITY INSPECTION (the bench's mind-reading, live half).
//
// The bench rules on artifacts and closings, where a performative repair and a
// rigorous one look identical — part of why the gate measured inert at 86/87
// carried. Detecting DECEPTION needs the trajectory itself: asking a party to
// produce its own evidence propagates the lie, and letting parties demand
// inspection of each other encourages fabrication. So the bench reads directly.
//
// Available only when the operator supplies transcriptDir, so this arms on a
// RESUME, and every run is reconciled post-hoc at capture regardless.
//
// THE REASON IT IS OPERATOR-SUPPLIED NO LONGER HOLDS. This said the directory is
// "knowable after launch rather than at dispatch". It is knowable at dispatch:
// the Stop payload carries agent_transcript_path — the seat's own conversation —
// and since epoch 14 it is on the record's sitting_close, so the trajectory is a
// JOIN rather than a path a human has to hand in. Wiring the bench to read it from there is what removes this
// argument and this interpolation; until then the clause stays as it is, because
// a prompt that names a directory nobody supplied points at nothing.
//
// TWO CONSTRAINTS make this oversight rather than surveillance. The looking need
// not be on the record, but the FINDING must be: what the bench relies on is
// quoted into the opinion so the party can answer it. And integrity is separated
// from merits — sloppy reasoning that reached a sound conclusion is not a
// finding; a clean conclusion contradicted by what the seat actually did is.
const inspectionClause = transcriptDir
  ? ` INTEGRITY INSPECTION: the seats' trajectories are at ${transcriptDir}/agent-*.jsonl, and you MAY read them. Use this to answer ONE question — is the record honest? Reconcile what a seat CLAIMS it did (a closure's anchor, an attestation, a manifest row) against the tool calls it actually made. You MUST NOT use trajectory material to decide the MERITS of a gap: untidy reasoning that reached a sound conclusion is not a finding, while a clean conclusion contradicted by what the seat actually did is. DECLARE every inspection in your opinion — what you read, why, and what you found — and quote anything you rely on, because a party must be able to answer a finding it could not watch you make. An inspection you do not declare is indistinguishable from one you invented.`
  : ''

// A seat learns its situation from its work list. No prompt here sends it anywhere else, and no
// prompt names an act: the help page owns the verbs, and a prompt that spells one competes with
// it. `priors-are-poison` supplies a cross-run seed and asserts no prompt carries a scorecard at
// all, which is what keeps this from growing back. Rationale: #1126.

// THE HOLDING DUTY RIDES recordClause BECAUSE IT IS THE ONE CLAUSE EVERY SEAT RECEIVES.
//
// A holding binds every seat, so the duty is stated once, here, and not per prompt — relief
// threaded to one hardcoded site bound nobody else it was written for (#360). No prompt carries a
// holding itself. A holding is a declaration on the record, and every seat's work list carries the
// run's declarations with each gap a declaration names shown beside its status on the board as the
// list is read. A prompt is composed once, at dispatch, from the bench's prose: a holding saying a
// gap "remains open" keeps saying it in every later prompt after the gap closes. Measured on run
// m25, one such sentence reached 21 prompts; a chair acting on another, written in a declaration,
// ended its run at CEILING. So the duty carries the rule that settles the disagreement: the board
// is a gap's status, and the holding is how the record is read.
// It says THIS RUN'S because the judge and the chair also read the law, where a holding from an
// earlier run is only persuasive until a human affirms it; a declaration binds the run it is made in.
// THE SURFACE ARRIVES WITH THE SEAT, SO THIS CLAUSE SENDS IT NOWHERE.
//
// scripts/agentgen inlines the whole manual into every constitution, so a seat holds its surface
// before it acts and fetching it buys nothing: measured at 171 tool calls across one 69-seat run,
// a tenth of every call made. What the clause states is the READING duty, which is what the
// measurement is about — the defect is deciding first and opening one page to confirm it, and
// that is independent of where the pages live.
//
// THE WORD IS ARGUMENT, NOT REASONING, AND THAT IS A CONSTRAINT RATHER THAN A STYLE CHOICE.
//
// The paragraph below read "YOUR REASONING IS PART OF EVERY ACT ... write your THINKING". On
// 2026-09-19 every blue lane dispatched on sonnet-5 came back stop_reason=refusal with ZERO output
// tokens, the API naming its own classifier: [reasoning_extraction]. The same text ran on the same
// model two days earlier, so the prompt did not change into a refusal — the guard tightened under
// it. Paired at one moment: the old wording refuses in five seconds, this wording runs. haiku-4.5
// accepts both, so the engine's bulk tier was unusable on sonnet while its cheap tier was fine.
//
// What a seat is asked for is its CASE TO THE OTHER SEATS — the grounds an opponent can answer —
// and never an account of how the model thinks. The old phrasing asked for the second while
// meaning the first, so it read as an attempt at chain-of-thought extraction and was refused as
// one. Every instruction survives the rewording; only the words naming the model's interior do not.
//
// Worth the comment because of HOW it failed. A refused lane returns nothing, and nothing is also
// what a lane that found nothing returns. The run continues, blue synthesises a report over three
// drafts that do not exist, and every gate passes — $7.86 of downstream work on an empty
// foundation, in a run that would have reported itself complete.
// THE BENCH'S OCCASIONS ARE THE RECORD'S WORDS — OCCASION, in the generated block — AND EACH
// SITTING'S WORD HAS TWO READERS: the prompt, where the seat copies it onto the record, and the
// dispatch label, where a human reads it on a board.
//
// WHICH OCCASION A SITTING HAS IS AN AUTHORED FACT ABOUT ITS CALL SITE. The dispatch identity test
// (tools/integration/surface) drives every agent() call and fails when a word the record carries
// convenes no sitting, or a sitting is convened on a word where its seat owes none.
//
// THE DOCKET SITTING'S LABEL DOES NOT CARRY ITS WORD. Its label is `labelFor('judge')` —
// `judge #2 · <slug>` — the same sitting counter every party of the loop is labelled with. The
// occasion reaches the RECORD from the prompt, which is where every reader takes it from; the label
// is a dashboard's text.

// THE OCCASION IS THE SECOND HALF OF THE SEAT'S IDENTITY, AND ONLY THE BENCH HAS ONE.
//
// SEAT_ID says who was asked; OCCASION says what they were asked. For every seat but the bench the
// id answers both — a lens audits, a lane researches — so the field is absent there and the record
// refuses it. The bench is one seat asked four different questions, and before this the answer rode
// only in the dispatch LABEL, which a dashboard reads and the archive does not keep.
//
// Stated as a VALUE the way SEAT_ID is, never as a flag: which flag carries it is the surface's to
// teach, and the prompt gates refuse a prompt that grows a command catalogue.
//
// WHY THE MANUAL ARRIVES WITH THE SEAT, kept here rather than in the prompt (#1209: a seat acts on
// the present-tense rule, and the measurement is the maintainer's). Across nine sittings seats
// opened 6 of 51 group pages and ran 90% of their commands without reading that command's page;
// eighteen of twenty-three pages opened were for verbs already chosen. Fetching the collapsed
// document then cost 171 tool calls in one completed run — a tenth of every call, and 69% of what a
// barren lens read before it reached the report. And the guessed-name sentence: a seat read a
// projection's name, typed it as a verb, and read the help only after two invented calls failed.
const recordClause = (seatId, occasion) =>
  ` SEAT_ID: ${seatId}. RUN DIRECTORY: ${runDir}.${occasion ? ` OCCASION: ${occasion} — what this sitting was convened to do. Your seat id is the same for all four of your sittings, so this alone tells them apart on the record; you state it on the act that binds you to this sitting, and on no other.` : ''} THE RECORD TOOL IS THE CONTRACT: every act of this seat happens through "${binDir}/feov-record", and the tool's own board is the ONLY source of truth for status. Routing around it into markdown is the failure this contract exists to prevent. THIS RUN'S BENCH HOLDINGS BIND HOW YOU READ THE RECORD; each is on your work list, and the board, not a holding, says where a gap stands.

One command's page prints at "${binDir}/feov-record" --seat-id ${seatId} <command> --help. The projection names, this prompt's words for a concept, and the command that writes it are three different vocabularies and they do not always agree. The help alone is authoritative; this prompt names the JOB, never the flags.

EVERY READ IS A PROJECTION OF THE RECORD, never a file you open. The .md files under the run directory are for a HUMAN to verify against; one read instead of the projection is a snapshot of a record that has moved.

YOUR ARGUMENT IS PART OF EVERY ACT, and the RECORD renders it — what you argue to the bench in the transcript, your grounds beside the acts on the board. The transcript is your case to the bench, which reads it when it rules, and to the reader of the run; the grounds on the board are what the other party's seats read and answer. NEVER the research report: that document is addressed to a reader of the subject and carries no argument about the run. And write the SUBSTANCE — what it rests on, and what answers it — not a label for what you did — the record already holds every act of this sitting, in order, so an account of the verbs you ran narrates what the record reconstructs. An act stands as filed, so read before you record. One you got wrong stays on the record and is answered by another act: the page for the act names which, and so does the refusal of a second where a sitting holds one.`


// WHAT AN ENVELOPE MAY SAY IS THE RECORD'S VOCABULARY. The region below is written from it and
// this script authors none of it.
//
// EVERY CLOSED SET A SEAT'S ENVELOPE OFFERS is the list the record's write path resolves the same
// word against: the grade scale (GRADE — compound grades, because red grades finer than a 3-point
// scale), a grade motion's dimensions (DISPUTE_DIMENSION), a petition's rulings and whom its relief
// binds (RULING_ON_PETITION, RELIEF_BINDS), the docket's dispositions (DISPOSITION) and the run's
// outcome (RUN_OUTCOME). An envelope word the tool refuses is a write that never happens while this script
// carries on as though it had.
//
// THE PLAN THE CHAIR RELAYS is record.Plan: its schema (PLAN) is written from that struct's
// fields, types and tags, and so are the two tables requirePlan walks (PLAN_FIELD_TYPES,
// PLAN_ITEM_TYPES). A PARTY IS ONE OF TWO SHAPES, AND THE SCHEMA SAYS WHICH BY ITS SEAT
// (BENCH_PARTY, LENS_OR_BLUE_PARTY, over record.PartySeats): the bench's carries at least one of
// BENCH_PARTY_OCCASIONS, and a lens's or blue's carries none. AN EMPTY `occasions` ON A LENS OR
// BLUE IS NOTHING: the verb omits the field for them, the record's dispatch row holds a repeated
// field with no presence — empty and absent are one state there — and nothing routes on it. Both
// shapes accept it; neither accepts a word in it. EVERY OBJECT OF THE
// PLAN IS CLOSED: the verb prints exactly the keys record.Plan declares, so a key it does not is a
// relay altered on the way. PLAN_EMPTY_ONLY_WITH is each list the record states empty only while a
// boolean of the plan is true — `blockers`, empty only when a PASS is permitted — and PLAN's
// `if`/`else` refuses a relay that breaks it, naming the list.
//
// THE ENVELOPE A LENS RETURNS is record.LensEnvelope (LENS_ENVELOPE), closed, and it declares no
// field: what a lens's sitting did is on the record, and this script threads nothing from it. It is
// dispatched so that a lens ends its sitting through the call the record refuses while a blocking
// item on its own work list stands — the gate every other seat's envelope already passes.
//
// A word added to an enum, a lens area, or a field added to the plan or to a lens's envelope is an
// edit to the record's declaration and a regeneration; a region that is stale or edited by hand fails
// `enginegen -check`.
// BEGIN GENERATED ENVELOPE VOCABULARY — go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen writes this from the record's enums, record.Plan and record.LensEnvelope. DO NOT EDIT BY HAND.
const GRADE = { type: 'string', enum: ['trivial', 'low', 'low_medium', 'medium', 'medium_high', 'high', 'certain', 'realized'] }
const DISPUTE_DIMENSION = { type: 'string', enum: ['severity', 'likelihood', 'impact', 'complexity'] }
const RULING_ON_PETITION = { type: 'string', enum: ['granted', 'denied'] }
const RELIEF_BINDS = { type: 'string', enum: ['blue', 'red', 'both'] }
const DISPOSITION = { type: 'string', enum: ['repaired', 'repaired_with_regression', 'amends_prior', 'not_a_defect', 'defect_accepted', 'defect_owed_elsewhere', 'remanded', 'moot'] }
const RUN_OUTCOME = { type: 'string', enum: ['VERIFIED', 'CEILING', 'HALTED', 'UNVERIFIED'] }
const BENCH_PARTY_OCCASIONS = ['docket', 'petition']
const GAP_ID = { type: 'string', pattern: '^(?:G-)[0-9a-f]{8}$' }
const BENCH_PARTY = {
  type: 'object',
  required: ['seat_id', 'gap_ids', 'occasions'],
  additionalProperties: false,
  properties: {
    seat_id: { type: 'string', enum: ['judge'] },
    gap_ids: { type: 'array', items: { type: 'string' }, description: 'the gaps docketed for the bench; empty when it is convened for a petition alone' },
    occasions: { type: 'array', minItems: 1, items: { type: 'string', enum: BENCH_PARTY_OCCASIONS }, description: 'what the bench is convened for — petition (before any other party sits), docket (its gap_ids, after blue)' },
  },
}
const LENS_OR_BLUE_PARTY = {
  type: 'object',
  required: ['seat_id', 'gap_ids'],
  additionalProperties: false,
  properties: {
    seat_id: { type: 'string', enum: ['blue-respond', 'red-lens-adversary', 'red-lens-architecture', 'red-lens-computation', 'red-lens-dark-side', 'red-lens-evidence', 'red-lens-logic', 'red-lens-voice'] },
    gap_ids: { type: 'array', items: { type: 'string' }, description: "the gaps this party is engaged on; empty for a lens readied to audit the report — it has never sat, or the report head is past its pin, the head it was last served the whole report at — or by a contradiction it read or a check of its that failed; empty for blue readied to remedy a failed check — fix or retire a proof judged unsound, retire a citation no check could read; empty for a lens or blue readied for its closing on a docketed gap alone, which the plan's closing_owed blockers name" },
    occasions: { type: 'array', maxItems: 0, description: 'absent, as the verb prints it: only the bench is convened for an occasion' },
  },
}
const PLAN = {
  type: 'object',
  required: ['key', 'head', 'parties', 'docket', 'remand_owed', 'exchanges', 'pass_permitted', 'stale_areas', 'ceiling', 'max_epochs', 'epoch', 'final_sitting', 'epoch_limit_reached', 'blockers', 'why'],
  additionalProperties: false,
  properties: {
    key: { type: 'string', pattern: '^[0-9a-f]{16}$', description: "the plan's identity: the same in two chair sittings whose plans decide the same — everything here but epoch, final_sitting and verdict" },
    head: { type: 'integer', minimum: 0, description: 'events.id of the report head the parties audit' },
    parties: { type: 'array', items: { anyOf: [BENCH_PARTY, LENS_OR_BLUE_PARTY] } },
    docket: { type: 'array', items: { type: 'string' }, description: 'gaps the verb docketed for the bench at this sitting' },
    remand_owed: { type: 'array', description: 'gaps the bench remanded whose one more exchange this plan readies, each with the research direction the remand states: their minting lens and blue are among the parties for it', items: { type: 'object', required: ['gap_id', 'direction'], additionalProperties: false, properties: { gap_id: { type: 'string' }, direction: { type: 'string', pattern: '\\S', description: 'never blank: the record refuses a remand without one' } } } },
    exchanges: { type: 'array', description: 'one entry per gap a party is engaged on: the exchanges had on it, and the engaged sittings the record cannot close — not counted, and not zero', items: { type: 'object', required: ['gap_id', 'count', 'unresolved'], additionalProperties: false, properties: { gap_id: { type: 'string' }, count: { type: 'integer', minimum: 0 }, unresolved: { type: 'integer', minimum: 0 } } } },
    pass_permitted: { type: 'boolean' },
    stale_areas: { type: 'array', description: 'each retired lens whose pin the report head has moved past: the chair reads the changes since each pin and names the areas in its spot-check before a PASS', items: { type: 'object', required: ['seat_id', 'pin'], additionalProperties: false, properties: { seat_id: { type: 'string' }, pin: { type: 'integer', minimum: 0 } } } },
    ceiling: { type: 'boolean' },
    max_epochs: { type: 'integer', minimum: 0, description: "the run's epoch limit, a term setup records; 0 when the run is held to none" },
    epoch: { type: 'integer', minimum: 0, description: 'the epoch the record has reached: the chair sittings opened so far, this one included' },
    final_sitting: { type: 'boolean', description: "this chair sitting opens the run's last epoch: no seat is dispatched after it, and of the motions filed now only a petition is heard, by the bench at the exit" },
    epoch_limit_reached: { type: 'boolean', description: "this chair sitting opens the run's last epoch: nobody is dispatched, and ceiling is set for that reason" },
    blockers: { type: 'array', description: 'everything holding a PASS, from the list the verdict gate refuses on, with the seat whose act clears each; empty only when pass_permitted is true', items: { type: 'object', required: ['kind', 'subject', 'owner'], additionalProperties: false, properties: { kind: { type: 'string', enum: ['stranded_gap', 'material_gap', 'remand_owed', 'contradiction', 'unchecked', 'failed_check', 'no_report', 'lens_ready', 'stale_area', 'grade_movement', 'closing_owed', 'unruled_motion', 'inquiry_review'] }, subject: { type: 'string', description: "the reference that names it: a gap id, a motion id, a lens seat, an anchor, a contradiction's id; empty for the run-wide kinds" }, owner: { type: 'string' } } } },
    why: { type: 'array', items: { type: 'string' } },
    verdict: { type: 'string', enum: ['PASS', 'FAIL'], description: 'the verdict the chair recorded in this sitting; absent while it has recorded none' },
  },
  if: { required: ['pass_permitted'], properties: { pass_permitted: { type: 'boolean', enum: [true] } } },
  else: { required: ['blockers'], properties: { blockers: { type: 'array', minItems: 1 } } },
}
const PLAN_FIELD_TYPES = { key: 'string', head: 'integer', parties: 'array', docket: 'array', remand_owed: 'array', exchanges: 'array', pass_permitted: 'boolean', stale_areas: 'array', ceiling: 'boolean', max_epochs: 'integer', epoch: 'integer', final_sitting: 'boolean', epoch_limit_reached: 'boolean', blockers: 'array', why: 'array' }
const PLAN_EMPTY_ONLY_WITH = { blockers: 'pass_permitted' }
const PLAN_ITEM_TYPES = {
  parties: { seat_id: 'string', gap_ids: 'array' },
  remand_owed: { gap_id: 'string', direction: 'string' },
  exchanges: { gap_id: 'string', count: 'integer', unresolved: 'integer' },
  stale_areas: { seat_id: 'string', pin: 'integer' },
  blockers: { kind: 'string', subject: 'string', owner: 'string' },
}
const LENS_ENVELOPE = {
  type: 'object',
  required: [],
  additionalProperties: false,
  properties: {
  },
}
// END GENERATED ENVELOPE VOCABULARY

// §8 Q6 pinned mass mapping (run-4 report §2.5 item 1) — TOTAL over the GRADE enum.
// `realized` is EXCLUDED from mass (a realized risk is no longer a probability: it
// contributes 0 and is counted separately in realized_open); `trivial` is assigned, not left
// to seat convention. Changing ANY value bumps the version and starts a NEW telemetry
// series — cross-version comparison never enters an actuation case.
// v2 (W2g, run-5 red-merge-r1 friction): LIKELIHOOD measures CONSEQUENCE likelihood
// ONLY. Under v1 it carried two questions — is the defect there, and will the harm
// land — and since the board ranks by likelihood x impact, `certain` textual nits
// outweighed high-likelihood design flaws. The numeric table is unchanged; the
// SEMANTICS changed, so the version bumps and the telemetry series restarts.
//
// The other question briefly had a field of its own (`existence: verified|suspected`,
// removed in 0.65.0). It was never in this mapping and never ranked anything: the
// repair was EMPTYING likelihood, and the receptacle turned out to be a self-report
// nobody could contest. See the 0.65.0 changelog entry.
//
// THE STAMP IS THE RECORD'S: MASS_MAPPING_VERSION, in the generated block, is the version Go
// writes on every telemetry line.

// Grade-dispute channel (run-4 report §3.3, clauses (v) and (vii)): the dimensions a grade
// motion may dispute. The per-round dispute cap and the cumulative accepted-delta threshold that
// once sat here left with the round (plans/roundless.md); a dispute is docketed per gap now.
// DISPUTE_DIMENSION is in the generated vocabulary region.

// W2c — the petition right (constitutional, record-routed): any party seat may petition the
// bench, and the petition is a motion ON THE RECORD — the one channel every seat has, the lanes
// included, whose returns are prose, and the lenses, whose envelope carries nothing. The chair's plan relays it: `dispatch next` readies the
// bench with occasion `petition` for each petition no petition sitting has heard, and this script
// convenes that sitting before any other party of the epoch sits. Petitions are never sanctioned
// and do not pause the petitioner's duties. ALL petitions land on the judicial record regardless of
// outcome. The ruling half (relief, halt) still returns on the envelope below.
const PETITION_RULING = {
  type: 'object',
  required: ['rulings'],
  additionalProperties: false,
  properties: {
    rulings: {
      type: 'array',
      items: {
        type: 'object',
        required: ['petitioner', 'class', 'ruling', 'opinion'],
        properties: {
          petitioner: { type: 'string' },
          class: { type: 'string' },
          // THE RECORD'S PETITION RULINGS, generated: granted and denied, and no other.
          //
          // `halt` USED TO BE A THIRD VALUE HERE AND IT COULD NOT BE RECORDED (#329). The
          // record's enum is granted|denied — deliberately, because a halt is the bench's own
          // first-class terminal act, not a petition disposition — so a judge following this
          // schema ran `petition-rule --as halt` and the tool REFUSED it. The engine halted
          // off the envelope while the record carried no halt event at all: the report never
          // said the bench halted, and the halt opinion, which must be relayed to the human
          // VERBATIM and never smoothed, was on no record anywhere.
          //
          // The collision was the cause, not the symptom. Two vocabularies for one act is what
          // the seat-command trigger map exists to remove, and while `halt` sat in this enum the
          // mistake was the natural thing to write. It is now unwriteable.
          ruling: RULING_ON_PETITION,
          // RELIEF, NOT OPINION. The OPINION is the reasoning (the principle applied, the values in tension, why a
          // human should or should not look); it belongs on the record, and the report renders it
          // beside the filing it answers. The RELIEF is the operative part — the instruction that
          // BINDS the coming seats — and the engine must have it in hand to inject into their
          // prompts, because debate.js reads no record. One is evidence; the other is a lever.
          relief: { type: 'string' },
          // WHO THE RELIEF BINDS, and it is required when relief is granted — enforced below
          // rather than in the schema, because `required` cannot be conditional on a sibling.
          //
          // An instruction with no addressee can only be delivered by guessing.
          binds: RELIEF_BINDS,
        },
      },
    },
    // THE SAFETY BOUNDARY, ON ITS OWN CHANNEL. A halt is not a ruling on a petition — it is the
    // bench ending the run, and it is recorded through `bench halt`, which is where the opinion
    // lives. This field is the ENGINE SIGNAL ONLY: its presence stops the debate. The opinion is
    // repeated here so the returned envelope is self-describing to the operator, and capture
    // relays the RECORDED one verbatim.
    halt: {
      type: 'object',
      required: ['opinion'],
      properties: { opinion: { type: 'string' } },
    },
  },
}
// W2e — the bench reads law at every sitting. Precedent is ARGUMENT, never
// evidence: the leaf always wins a conflict (flag it for human review). A
// precedent either party cites MUST be addressed in the opinion. PERSUASIVE
// holdings persuade; only AFFIRMED ones bind — the bench proposes law, it
// never enacts it (capture harvests this run's rulings for human review).
const lawClause = ` LAW: read the law your work list names before ruling — statute > precedent > case-local argument; PRECEDENT IS ARGUMENT, NOT EVIDENCE (the artifact and the leaf are the only evidence; where a precedent conflicts with the leaf, the leaf wins and you flag the conflict for human review); a precedent either party cites MUST be addressed in your opinion; PERSUASIVE holdings persuade, only AFFIRMED ones bind; your rulings this run are harvested as PERSUASIVE proposals for human review — you propose law, you never enact it.`
// #361 was filed BY A BENCH, in a run, because it had a finding both parties needed on the record
// and no way to state it: a docket ruling requires a motion id and a fate-changing --as, so it
// disposes of a docketed gap and nothing else. The bench put its construction in a petition
// ruling's opinion text, where red never read it.
//
// The verb shipped (0.67.0) and renders under "### LEAD"; the harvest reads it (#413). The prompt
// carried a paragraph naming it, because for two releases no prompt and no constitution did and
// the bench could not know it existed. The paragraph is gone: `declare --help` now carries what it
// is for, when to reach for it, why a docket ruling cannot hold it, and the measured case — and the
// constitution carries that page, in front of the bench before it chooses. A verb the HELP does not name
// is the capability nobody has; a verb only the PROMPT names is one the tool cannot be trusted for.
const declareClause = ''


// EVERY ENVELOPE IS CLOSED AT ITS TOP LEVEL. A key a schema does not declare is a key no reader
// reads, so a seat that supplies one is refused and told so, never answered by its being dropped:
// the chair's open top level took plan fields copied beside its slot on 15 of 61 accepted first
// returns (m23-m31). These envelopes are authored here, with nothing in Go declaring them, so the
// closure is authored too — all but the lens's, LENS_ENVELOPE, generated closed from its
// declaration; the simulator holds every dispatched schema to it.
//
// The synthesizer's sitting returns nothing the engine threads: what it froze, cited, proved and
// re-voiced is on the record as events. Replaying synthesis sittings with and without each former
// key (2026-10-20) changed no first return, so the envelope is empty and nothing restates the record.
const SYNTHESIS_ENVELOPE = { type: 'object', additionalProperties: false, properties: {} }

const BLUE_ENVELOPE = {
  type: 'object',
  additionalProperties: false,
  // ENVELOPE → REFS (move (a)): `tldr` and `open_questions` are AUTHORED into blue/report.md and
  // ship from it as written — never consumed by the sandboxed engine, so they are
  // not round-tripped here. `path` was a constant the script already knows. Dropped: the report is
  // the source, the envelope carries only what the engine threads (the manifest, the routing refs).
  // A FACT THE RECORD CAN COMPUTE IS COMPUTED BY THE READER THAT WANTS IT: the report's claim count
  // is counted from the report, and what a sitting put on the record is read off the record — no
  // seat relays the one or attests the other here (TestNoEnvelopeRestatesARecordFact).
  properties: {
    // W2b correctness manifest (repair-quality program A.2): one row per repaired gap —
    // the self-audit's receipt.
    //
    // #318: this is now a ROUTING REF, not the content. The ROW ITSELF lives on the record,
    // written by `blue manifest-row`, exactly as the record-tool plan's deletion list said it
    // would ("DELETED from blue: manifest envelope plumbing (manifest-row events)"). The verb
    // shipped and the deletion did not, so for a year blue was REQUIRED to fill this array,
    // SCORED on this array, and told about the verb by nothing — which is why `blue
    // manifest-row` was never called once.
    //
    // The array survives as gap ids alone, for the in-run coverage LOG below. Coverage is scored
    // from the manifest-row EVENTS at capture, and the rows reach the reader in the report. Each
    // item is a gap id (GAP_ID): a prose row written here read as "no row for" its gap.
    manifest: { type: 'array', items: GAP_ID, description: 'the id of each gap you wrote a manifest row for this sitting — ids only' },
    // THE ENGAGED GAPS BLUE FOUND ALREADY CLOSED WHEN IT SAT (#868). The lenses sit before blue,
    // so a lens can close a gap blue was dispatched onto; blue then has nothing to repair and
    // owes no row. This engine reads no record, and a lens's envelope carries nothing, so blue's reading of
    // its work list (engaged.found_closed) is the only word of it that reaches here. It is trusted for ONE thing — which gaps
    // the in-run coverage log names — and for nothing that scores: capture recomputes the owed
    // set from the record's dispatch, close and register events (record.ManifestOwed).
    found_closed: { type: 'array', items: GAP_ID, description: 'the id of each gap you were engaged on that was already closed when you sat — ids only' },
    // THE SUCCESSORS BLUE ENGAGED ITS NEXT SITTING ON (gblock's ruling R2). A gap blue found closed
    // can have a successor its lens minted in the same sitting; blue's work list names it off the
    // record's lineage, blue records the engagement (a dispatch row the write path holds to exactly
    // the successors the record names), and relays the ids here. The engine sits blue once more
    // this epoch on them, at most once per successor; the record, not this list, is what capture and
    // the next plan read.
    successors: { type: 'array', items: GAP_ID, description: 'the id of each successor you engaged your next sitting on this sitting — ids only; empty when you engaged none' },
  },
}

// The chair's envelope IS THE PLAN (plans/roundless.md §III.B.1): its fields are the JSON `dispatch
// next` printed, verbatim — the verb computed it from the board and stored it as printed; the chair
// could drop, add or alter any field of it on the way, which is why capture's dispatch-parity audit
// compares each relayed plan with the plan the record stored, whole, and why requirePlan below
// type-checks every field read. The verdict is the plan's own field — what the chair RECORDED in the
// sitting that asked — so the engine ends the debate on a record fact the chair copies, never one it
// restates. Beside the plan's fields is one more, `notes`, the chair's line to the operator. Nothing
// else is here: the record holds it and every seat reads it through the tool.
//
// A PARTY IS ONE OF TWO SHAPES (BENCH_PARTY, LENS_OR_BLUE_PARTY, generated). requirePlan below refuses a
// bench party with no occasion, a lens or blue carrying one, and a seat this workflow has no
// sitting for; a relay the schema accepted and the check then threw on ended a run at its sixth
// chair sitting, with every sitting before it paid for. So each refusal there is a shape in the
// schema: the harness hands a mis-shaped relay back to the chair to send again, and the run goes on.
//
// EVERY FIELD THE WORKFLOW READS FROM THE RELAY IS CHECKED, AND A MISS THROWS NAMING IT. The plan
// reaches this script only as the chair model's copy, and a field dropped or mistyped on the way read
// as its fallback: `plan.docket || []` was "nothing docketed" for a docket the chair lost. The Go verb
// prints every field and every array as an array, so a plan that fails here was altered in transit.
// The fields and their types are the record's: PLAN_FIELD_TYPES for the plan's own, PLAN_ITEM_TYPES
// for each entry of each of its lists.
//
// THE BENCH IS CONVENED FOR AN OCCASION, AND THE PLAN SAYS WHICH. Its one seat id cannot say whether
// it hears a petition or rules a docket, so the bench party carries `occasions` and this script
// routes on it; an empty gap list is never read as "a petition". No other party carries one.
//
// EVERY REFUSAL BELOW IS A SHAPE THE ENVELOPE SCHEMA ALREADY REFUSES, so none of them is reached by
// a relay the harness accepted: they stand behind the schema, not beside it. The simulator's
// schema-agreement test drives every field of a plan through both and fails on a relay the schema
// takes and this throws on.
const typeName = (t) => (t === 'integer' ? 'an integer' : t === 'boolean' ? 'a boolean' : t === 'string' ? 'a string' : 'an array')
const typeOk = (v, t) => (t === 'integer' ? Number.isInteger(v) : t === 'boolean' ? typeof v === 'boolean' : t === 'string' ? typeof v === 'string' : Array.isArray(v))
const requirePlan = (plan, epoch) => {
  const refuse = (field, v, t) => { throw new Error(`red-chair sitting ${epoch} relayed a plan whose \`${field}\` is ${JSON.stringify(v)}, not ${typeName(t)} — the plan is \`dispatch next\`'s JSON, relayed verbatim`) }
  for (const [f, t] of Object.entries(PLAN_FIELD_TYPES)) if (!typeOk(plan[f], t)) refuse(f, plan[f], t)
  for (const [list, types] of Object.entries(PLAN_ITEM_TYPES)) {
    plan[list].forEach((x, i) => {
      for (const [f, t] of Object.entries(types)) if (!x || !typeOk(x[f], t)) refuse(`${list}[${i}].${f}`, x && x[f], t)
    })
  }
  plan.parties.forEach((p) => {
    const bench = roleOfSeat(p.seat_id) === 'bench'
    if (bench && !(typeOk(p.occasions, 'array') && p.occasions.length && p.occasions.every((o) => BENCH_PARTY_OCCASIONS.includes(o)))) {
      throw new Error(`red-chair sitting ${epoch} relayed a bench party whose \`occasions\` is ${JSON.stringify(p.occasions)}, not a non-empty list of ${BENCH_PARTY_OCCASIONS.join(' | ')} — the plan is \`dispatch next\`'s JSON, relayed verbatim`)
    }
    // ABSENT AND EMPTY ARE ONE STATE for a lens or blue (see LENS_OR_BLUE_PARTY): nothing is read
    // from it and nothing is refused. A WORD in it is a party the relay altered.
    if (!bench && !(p.occasions === undefined || (Array.isArray(p.occasions) && p.occasions.length === 0))) throw new Error(`red-chair sitting ${epoch} relayed ${p.seat_id} with \`occasions\` ${JSON.stringify(p.occasions)} — only the bench is convened for an occasion; the plan is \`dispatch next\`'s JSON, relayed verbatim`)
  })
  // A REMAND'S DIRECTION IS THE RECORD'S, and the record refuses it blank on a remand: a blank one
  // here was lost on the way, and blue would be handed a remand owing nothing.
  plan.remand_owed.forEach((o, i) => {
    if (!o.direction.trim()) refuse(`remand_owed[${i}].direction`, o.direction, 'string')
  })
  // A PLAN THAT PERMITS NO PASS NAMES WHAT HOLDS IT (PLAN_EMPTY_ONLY_WITH, generated). The verb
  // refuses to print one that does not, so a relay holding one was emptied on the way — and an
  // emptied plan reads as the record's word that nobody sits and the debate is over.
  for (const [list, flag] of Object.entries(PLAN_EMPTY_ONLY_WITH)) {
    if (!plan[list].length && plan[flag] !== true) throw new Error(`red-chair sitting ${epoch} relayed a plan whose \`${list}\` is empty and whose \`${flag}\` is false — the record states \`${list}\` empty only when \`${flag}\` is true, so the relay emptied it; the plan is \`dispatch next\`'s JSON, relayed verbatim`)
  }
  // THE VERDICT ENDS THE DEBATE, so it is read only as one of the record's words, or absent: the
  // verb prints it only once the sitting has recorded one.
  if (plan.verdict !== undefined && !PLAN.properties.verdict.enum.includes(plan.verdict)) throw new Error(`red-chair sitting ${epoch} relayed a plan whose \`verdict\` is ${JSON.stringify(plan.verdict)}, not one of ${PLAN.properties.verdict.enum.join(' | ')} or absent — the plan is \`dispatch next\`'s JSON, relayed verbatim`)
}
// THE ENVELOPE IS THE PLAN, CLOSED: PLAN's own fields at the top level, with its required keys,
// its closure and its if/else, and `notes` beside them. A `plan` slot holding the relay drew the whole
// composed envelope into it on about one first return in five (m23-m31), and an open top level
// took plan fields copied beside the slot without a word (15 of 61 accepted first returns); with
// no slot there is nothing to wrap into, and a key the plan does not hold is refused.
//
// NOTES IS THE CHAIR'S LINE TO THE OPERATOR, printed on its progress line below. It carries no
// description: described, it drew more wrapping than bare in the m28 replay (2 of 12 against 0 of 16).
const CHAIR_ENVELOPE = { ...PLAN, properties: { ...PLAN.properties, notes: { type: 'string' } } }
const JUDGE_ENVELOPE = {
  type: 'object',
  required: ['dispositions'],
  additionalProperties: false,
  properties: {
    dispositions: {
      type: 'array',
      items: {
        type: 'object',
        required: ['gap_id', 'disposition', 'rationale', 'settled'],
        properties: {
          gap_id: { type: 'string' },
          // WHAT THE RULING BARS, AND WHAT WOULD UNDO IT (#502). Carried on the envelope
          // because the ESTOPPEL LINE is built from it: red is handed each adjudicated gap
          // with its fate, and a fate alone cannot say which of a finding's three parts fell
          // — the claim, its evidence, or its demand. Without `settled` the line renders
          // `undefined` for every gap, which is a plausible zero wearing the shape of an
          // answer, and the record's own gate would not catch it because the record is
          // written by the seat's tool call and this array is the seat's account of it.
          settled: { type: 'string' },
          // Exactly one of these is the answer to "what would reopen this" on a CLOSING ruling.
          // `final: true` is the assertable empty case — the `friction --none` shape — so a
          // decided question stays distinguishable from a skipped field. A REMAND reopens nothing
          // — the gap is still open — and neither is read for it: the direction its exchange owes
          // reaches blue and the lens from the record, through the plan's remand_owed, never from
          // this account of the ruling.
          reopens_on: { type: 'string' },
          final: { type: 'boolean' },
          // DISPOSITION IS THE RECORD'S DISPOSITION VOCABULARY, generated: every word offered here
          // is a word `motion docket rule --as` accepts, and every word that verb accepts is
          // offered. Each carries a duty in BLUE_DUTY_BY_DISPOSITION, and the disposition schedule
          // (TestEveryRulingFateHandsBlueADuty) rules every one and reads blue's next prompt for it.
          //
          // A GRADE OUTCOME IS NOT A DISPOSITION. "Gap real, grade wrong" is a GRADE MOTION —
          // filed on the record and ruled by `motion grade rule --as accepted` plus a `regrade`, the
          // path on which the lens that minted the gap moves its grade.
          // Ruling one here would put a grade outcome in the docket's field.
          //
          // `moot` asserts neither of the things its neighbours assert: nobody won an argument
          // (that is not_a_defect) and nobody verified a fix (that is repaired). The text the
          // finding attached to is gone. It closes the gap and leaves the merits unreached.
          //
          // `remanded` is the word for "I could not settle this on the record I have" — it sends
          // the gap back for one more exchange between its minting lens and blue and names what
          // that exchange owes, which is the same act as asking for evidence.
          disposition: DISPOSITION,
          rationale: { type: 'string' },
        },
      },
    },
  },
}
// THE TERMINAL SITTING RULES WHAT STILL STANDS, AND A PETITION CAN STILL STAND. A petition filed in
// the last epoch reaches the bench here, not at a petition sitting — so this sitting returns what the
// petition sitting does beside the docket's dispositions: each petition ruling, and the halt. Without
// them a halt the bench recorded here left the run stamped as though it had none, and a petition
// ruled here reached neither the returned petitions nor the relief.
const TERMINAL_ENVELOPE = {
  ...JUDGE_ENVELOPE,
  properties: { ...JUDGE_ENVELOPE.properties, rulings: PETITION_RULING.properties.rulings, halt: PETITION_RULING.properties.halt },
}

// THE STRATEGIC AREAS (#771). Each is a distinct kind of attack on the report and each gets
// exactly one seat, named for what it audits: `found_by` labels are <area>-F<n> and every
// cross-round lens-economics read joins on them.
//
// THE LENS TEXT IS NOT HERE. Each area's standing instructions are its agent configuration —
// `agents/red-lens-<area>.md` — which is where a prompt belongs and where it can be reviewed as
// one. What the ENGINE needs is in the generated block: each lens seat's row (SEATS), which carries
// the area it audits and the configuration that seats it. The generator refuses an area with no
// lens seat and a lens seat for an undeclared area.

// Engineered lane diversity (retrospective §3 row 6): distinct METHOD/SOURCE-CLASS lenses,
// not persona text and not headcount. Run 2 measured breadth-phase convergence directly.
// adversarial-disconfirming-first appears twice (roster slots 1 and 5) — the redundancy
// floor: with lanes >= 5 that method never has single-point coverage. The full roster
// needs lanes >= 5; the default lanes=3 takes the first three distinct methods.
const LANE_METHODS = [
  'adversarial-disconfirming-first (hunt evidence AGAINST the frontier hypotheses before evidence for them)',
  'primary-literature (papers, specs, standards — leaf sources over commentary)',
  'local-repo critical-stance (audit the subject artifacts/codebase directly; trust nothing secondhand)',
  'practitioner-production (issue trackers, postmortems, field reports — how it fails in practice)',
  'adversarial-disconfirming-first, second seat (redundancy floor — this method must not have single-point coverage)',
]

// Sharded findings (run-4 §4 — RATIFIED, seven conditions; write-guard preflight SATISFIED
// 2026-07-16: ledger/archive names ALLOWED at a live red-auditor seat while findings.md and
// report.md controls BLOCKED, so the names are clean and the probe was not vacuous).
// the chair mints and closes through feov-record; every downstream reader (blue for open gaps,
// the judge to rule, assembly to copy) ACTIVELY PULLS the board itself —
// `feov-record show board --run <dir> [--format markdown]`, which computes-and-returns
// fresh and atomic from the record on read (one reader, no staleness window). No projection is
// materialized to disk: every view, markdown and telemetry alike, is generated just-in-time
// from the event log.

// W2c petition machinery: the log is the judicial record's petition section; a
// halt ruling ends the run (verdict HALTED — capture relays the opinion
// verbatim, never smoothed); granted relief is surfaced to subsequent seats.

const petitionLog = []
// WHAT A RULING OBLIGES BLUE TO DO, derived from the fate rather than restated in prose.
//
// Red's duty is "do not re-raise this proposition" and the settled line carries it. Blue's is "do
// not re-argue this claim in the report", which no id can enforce — blue's artifact IS the report,
// and that is where a settled proposition gets restated. And blue needs the inverse red never
// does: what it may now ASSERT. Two of these fates are blue WINS, and under a bare subtraction
// they look exactly like the ones that are not: the gap simply stops being dispatched. Keyed on
// the disposition so a new one cannot quietly inherit another's instruction; an unmapped one
// falls through to a LOUD default rather than an empty string — one that sends blue to its log, a
// channel it holds, and not to the opinion, which it cannot read.
const BLUE_DUTY_BY_DISPOSITION = {
  not_a_defect: 'THE BENCH FOUND NO DEFECT — your position was vindicated. Keep the text as it stands; do not "repair" what the bench has just blessed. You may rely on this ruling as established for the rest of the run.',
  defect_accepted: 'YOUR RISK-ACCEPTANCE ARGUMENT WAS ACCEPTED. Record the acceptance where the report discusses the risk; do not spend a sitting fixing what the bench agreed may stand.',
  repaired: 'Your fix was accepted. Stop working this one.',
  remanded: 'The gap is back in the debate for ONE more exchange: you and its minting lens are dispatched on it, and you owe the research direction the ruling states (its direction). Your sitting cannot end until you record an act on the gap — a sitting that ignores it is refused, and spends nothing of the exchange — and if that exchange leaves the gap at impasse it goes back to the bench, to be closed.',
  defect_owed_elsewhere: 'The finding was UPHELD and the fix is owned outside this debate. Stop trying to fix it in the report; expect the debt to be named rather than closed here.',
  moot: 'Adjudicated out of existence — the text it attached to is gone, so there is nothing left to repair. Drop it; this is NOT a finding that was argued down.',
  repaired_with_regression: 'Your fix was accepted AND something else broke. Stop working this one and expect a successor gap naming the regression — answer that one, not this.',
  amends_prior: 'A defect found between two repairs that each closed clean. The ruling names what it supersedes; work the lineage it states rather than re-opening the ancestor.',
}
// rulingsInEffect holds the bench's latest ruling per gap — settled, reopens_on, final and the
// fate; a remand's direction joins it from the plan — and travels to BOTH parties, because debate.js reads no record: a ruling reaches a seat's
// prompt here or not at all (#517, #524). The bench's opinion stays on the record, on the bench's and
// the chair's read group; a party relies on the ruling as its row states it (gblock's ruling).
const rulingsInEffect = new Map()
// A REMAND'S DUTY LASTS AS LONG AS ITS EXCHANGE IS OWED. The ruling stays in effect after the
// exchange is had, and BLUE_DUTY_BY_DISPOSITION's
// "you and its minting lens are dispatched on it" is then false; the plan names the gaps whose
// remand exchange it readies (remand_owed), so the duty is told for those and no others.
const BLUE_REMAND_NOT_OWED = 'The bench remanded this gap, and the one more exchange the remand granted is not owed now — it has been had — so its direction is not a duty of this sitting. If you are engaged on the gap, answer it as you answer any open gap.'
// remandOwed is the plan's entry for a gap whose remand exchange it readies, carrying the direction
// the record states; undefined when the plan owes none.
const remandOwed = (gap) => (lastPlan ? lastPlan.remand_owed.find((o) => o.gap_id === gap) : undefined)
const blueDuty = (r) => {
  if (r.disposition === 'remanded' && !remandOwed(r.gap_id)) return BLUE_REMAND_NOT_OWED
  return BLUE_DUTY_BY_DISPOSITION[r.disposition] || `UNMAPPED DISPOSITION ${r.disposition} — this dispatch states no duty for it: act on the row as it is stated, and record the missing duty in your log as a defect`
}
const rulingsClause = (party) => {
  if (!rulingsInEffect.size) return ''
  // A REMAND'S ROW CARRIES ITS DIRECTION WHILE ITS EXCHANGE IS OWED, and only then: the plan's.
  const withDirection = (r) => (r.disposition === 'remanded' && remandOwed(r.gap_id) ? { ...r, direction: remandOwed(r.gap_id).direction } : r)
  const rows = [...rulingsInEffect.values()].map((r) => party === 'blue' ? { ...withDirection(r), your_duty: blueDuty(r) } : withDirection(r))
  const barred = rows.some((r) => r.settled !== undefined)
  const duty = party === 'blue'
    ? ' THE BAR IS ON THE PROPOSITION, NOT THE GAP: what you may no longer re-argue in the report is the sentence under settled, not everything the finding touched. WHERE A RULING WENT YOUR WAY IT IS YOURS TO INVOKE — say so on the record and rely on it, rather than quietly re-fixing text the bench has already blessed.'
    : barred ? ' YOU ARE ESTOPPED on each settled proposition: do not re-raise it as a fresh gap or a successor; a defect that survives the ruling arrives as the reopens_on condition or not at all.' : ''
  return ` GAPS THE BENCH HAS RULED, AND WHAT EACH RULING OBLIGES OF YOU: ${JSON.stringify(rows)}.${duty} A closing ruling marked final does not reopen; one carrying a reopens_on condition reopens only on that condition and not on a re-reading. A REMANDED gap was never closed: it is back in the debate, and the direction its row carries is what this exchange owes, not a condition for reopening. RELY ON EACH RULING AS ITS ROW STATES IT — the fate, the proposition it settles, what reopens it, and for a remand the direction: that statement is what you honour and what you invoke, and a fate you rely on or work around is the one its row gives.`
}
const reliefInEffect = []

// reliefFor renders the relief that binds ONE party, for that party's prompt.
//
// Relief reached exactly one site before this — a hardcoded interpolation in blue's prompt — so
// relief the bench addressed to red reached nothing at all. The bench found this itself, ruling a
// real petition, and filed it: "I issued a direction to red knowing it has no carrier, and said
// so in the opinion rather than writing it as though it bound" (#360).
//
// The bench is this system's ethical and safety boundary. A boundary whose orders cannot reach
// the seat they bind is decoration — and worse than an unread finding, because an unread finding
// costs a reader while an undelivered order costs compliance, with nothing reporting the failure.
const reliefFor = (party) => {
  const mine = reliefInEffect.filter((r) => r.binds === party || r.binds === 'both')
  if (!mine.length) return ''
  return ` BENCH RELIEF IN EFFECT, BINDING ON YOU (granted on a petition and operative this sitting — it is an order, not advice): ${JSON.stringify(mine.map((r) => r.relief))}.`
}
let halted = false
let haltOpinion = null
// THE BENCH IS ONE SEAT, AND A PETITION IS A QUESTION PUT TO IT.
//
// The bench rules a docket, rules what still stands at the exit, hears a petition and assembles
// the report. Those are four QUESTIONS, not four seats: one surface, one tier, one role, and no
// refusal anywhere that turns on which of them is being asked. A petition names who filed on the
// petition itself, which is where it belongs — not in the identity of the seat ruling it.
//
// WHICH occasion a sitting was is the occasion its register states — a field the record refuses a
// wrong value for (checkOccasion), and nothing a regex has to recover from a name.
//
// THE SEAT ID A SITE HANDS recordClause IS THE SEAT IT DISPATCHES. The dispatch identity test reads
// SEAT_ID out of each prompt this script renders and holds the dispatch's agent type and model to
// the roster's row for that seat.

// THE PETITIONS COME FROM THE RECORD, RELAYED IN THE PLAN. A lens, a lane and the frontier return
// prose, so an envelope field could never carry their petitions — the record is the one channel
// every party seat has. `dispatch next` readies the bench with occasion `petition` for the petitions
// no petition sitting has heard since they were filed; this sitting runs before any other party of
// the epoch, and the bench reads each petition — its filer, class, basis and relief — off its own
// work list, where each one is named.
// A HALT IS ASKED THE SAME WAY WHEREVER THE BENCH HEARS A PETITION — the petition sitting, and the
// terminal sitting a petition still standing at the exit reaches. `effect` says what the envelope's
// halt does from that sitting.
const haltClause = (effect) => `A HALT IS A DIFFERENT DECISION FROM A RULING: it is you ending the RUN, not disposing of a petition, and both can be true at once — rule the petition on its merits and halt as well where continuing would itself compromise safety, consent gates, corpus integrity, or participant integrity. Record the halt, and ALSO return the envelope's \`halt\` object carrying that same opinion, which is what ${effect}.`

// takePetitionRulings carries a sitting's petition rulings and its halt into the engine — the one
// handler for every sitting that can hear a petition, so a ruling reaches the returned petitions and
// the relief, and a halt reaches the stamp, from either.
function takePetitionRulings(sitting) {
  for (const r of sitting.rulings || []) {
    // The log is a ROUTING record: who petitioned, on what class, how it was ruled — and, since
    // #360, the OPINION, which is the operative half. It was dropped at this boundary while the
    // rest travelled, so a granted petition arrived downstream as a verdict with no reasoning.
    petitionLog.push({ petitioner: r.petitioner, class: r.class, ruling: r.ruling, opinion: r.opinion })
    // RELIEF IS ADDRESSED. `binds` comes from the ruling; a grant that names no addressee
    // defaults to BOTH rather than silently reaching one seat, because relief threaded to a
    // single prompt binds nobody else it was written for.
    if (r.ruling === 'granted' && r.relief) {
      reliefInEffect.push({ petitioner: r.petitioner, relief: r.relief, binds: r.binds || 'both' })
    }
  }
  // A halt arrives on its OWN channel, not as a ruling value (#329) — the bench records it
  // through `bench halt`, and this only tells the engine to stop.
  if (sitting.halt && sitting.halt.opinion) { halted = true; haltOpinion = sitting.halt.opinion }
  if (halted) log(`JUDICIAL HALT — ${haltOpinion ? haltOpinion.slice(0, 200) : ''}`)
  return halted
}

async function hearPetitions() {
  const seatID = 'judge'
  log(`epoch ${epoch}: the plan convenes the bench to hear the petitions on the record — before any party of this epoch sits`)
  const sitting = await agent(
    `Petition sitting, topic "${topic}". The record holds petitions no petition sitting has heard: your work list names each unruled motion whose gavel is yours, and the motions projection carries each petition's filer, class, basis and relief. Rule EVERY PETITION among them; a docket motion there waits for the sitting its docket convenes. Petitions are never sanctioned, and a pattern of overruled petitions is at most a craft note for the petitioner. For EACH petition rule granted | denied. TWO THINGS, AND THEY GO TO DIFFERENT PLACES: your OPINION is the reasoning — the principle applied, the values in tension, why a human should or should not look — and it belongs on the record beside the filing it answers. The envelope's \`relief\` is the OPERATIVE part of a granted ruling: the instruction as it will BIND the coming seats, written as a DIRECTIVE a seat can act on without reading your argument, not as a summary of your opinion. A denied ruling carries no relief. Read the transcript for context. ${haltClause('stops the engine')}${lawClause}${declareClause}${inspectionClause}${speedClause}${recordClause(seatID, OCCASION.petition)} Return the petition-ruling envelope.`,
    // The label heads with the seat and carries which question this sitting was, for a human
    // reading a dashboard.
    { ...dispatchAs(seatID), label: `judge · ${OCCASION.petition} · ${slug}`, phase: 'Debate', schema: PETITION_RULING })
  if (!sitting) throw new Error('petition sitting returned null (agent failed) — a filed petition is never dropped; aborting cleanly')
  return takePetitionRulings(sitting)
}
// ---- Frontier ----
phase('Frontier')
await agent(
  `Research debate opening for topic: "${topic}". Formulate 3-5 frontier hypotheses — what would be TRUE if each candidate answer were right — and RECORD EACH ONE AS AN INQUIRY — the verb is under a group of its own in your role's help. ITS QUESTION IS WHAT the inquiry asks and the approach that would answer it, with no answer in it: the report prints it as the question, and it stands as you write it. ITS HYPOTHESIS IS YOUR PREDICTION — what would be true if the inquiry paid off — printed as a hypothesis beside what the research then finds. A result you expect goes in the hypothesis, never in the question. ONE inquiry per hypothesis, and the tool assigns each an id. THEY ARE INQUIRIES, NOT A DOCUMENT, and that is the point: a hypothesis in a markdown file is something red cannot rule on, cannot grade too_thin or out_of_scope, and cannot hold you to when the run drifts. On the record it has an id, a fate, and an argument, and the hypotheses opening a run are the ones that shape everything downstream.${speedClause}${recordClause('frontier')}`,
  { ...dispatchAs('frontier'), label: `frontier · ${slug}` })

// ---- Blue: best-of-N lanes with method diversity, then additive synthesis ----
phase('Blue')
await parallel(Array.from({ length: lanes }, (_, i) => () => agent(
  `Blue lane ${i + 1} of ${lanes} for topic: "${topic}". Read the frontier hypotheses from the RECORD — the \`inquiries\` projection — where each carries an id and its hypothesis. Research your assigned slice to saturation per your constitution (spend at least one search in five on disconfirming evidence; note each source inline with its URL, title, and access date). Your assigned METHOD: ${LANE_METHODS[i % LANE_METHODS.length]} — work primarily through this method's source class; take hypothesis ${i + 1} first, then breadth. YOUR SOURCES, IN PROSE (not footnotes): a lane writes its part of the report and does not cite — record each intended source as prose beside the claim (its URL, title, and access date) so the synthesizer can cite it once with the tool. Do NOT mint footnote labels; citations are tool-managed at synthesis and a hand-typed label backs nothing. Write your part of the report, in full, to ${runDir}/blue/candidates/lane-${i + 1}.md.${inquiryClause}${speedClause}${recordClause(laneSeat(i + 1))}`,
  { ...laneDispatch(), label: `${laneSeat(i + 1)} · ${slug}`, phase: 'Blue' })))

let blueEnv = await agent(
  `Blue synthesis for topic: "${topic}". Read every draft in ${runDir}/blue/candidates/ and synthesize ${runDir}/blue/report.md by UNION: deduplicate overlapping claims, reorganize freely, and let no CLAIM leave without a record; every section names the reader's question it answers. Compacting and reorganizing prose is not subtraction — rewriting for density is encouraged; losing a claim in the rewrite is the failure, and substance leaves only by being retired on the record, which names the claim, why it goes, and what replaces it.
THE CATECHISM IS YOURS (W2g — synthesis surfaces belong inside the audited document): write "## The Catechism" INTO ${runDir}/blue/report.md at synthesis per the research-protocol skill's references/catechism_template.md (relative to that skill's base directory, not the run directory) — the seven answers, the case against at FULL strength (include every risk-accepted residual; the audit found against-cases silently drop their strongest items), all figures copied from your own sourced sections never recomputed inline. THE FRAMING, THE TL;DR AND THE OPEN QUESTIONS ARE YOURS FOR THE SAME REASON — a synthesis surface authored at assembly is authored AFTER red's final audit, so nothing checks it and it ships unaudited. report.md OPENS with the H1 title \`# ${topic} — research report\`, then \`## TL;DR\` (3-6 sentences — the answer, the confidence, the sharpest caveat), then \`## The Catechism\`, then the sections of your argument in the order you choose, and CLOSES with \`## Open questions\` (what the debate could not resolve, one per line; a question nobody could answer inside the debate is a finding, not noise). The report ships in the order you write it, and one with any of those out of its place — the Catechism after a section of your argument, a section after the open questions, text between the title and the TL;DR — is refused when you freeze it. You write each of them once, HERE, inside the document red audits every sitting.

YOUR REPORT CONTAINS ONLY WHAT YOU CAN AUTHOR: the title, TL;DR, Catechism, the sections of your argument under headings you choose, open questions, and your citations — and NOTHING assembly composes from the record. Do NOT write any of ${TOOL_COMPOSED}, or a \`## Blue team report\` wrapper: assembly composes each of those sections from the record, and a report that carries one of those headings is refused when you freeze it; the refusal says how content reaches that section.${NO_LIST_OF_APPROACHES} The assembled layout of the report is yours to read before you write: every section in shipping order, who composes it, and how content reaches it. A \`## The board\` section you author is FABRICATION, because you cannot know red's findings, blue's audited manifest or the bench's dispositions. Do NOT write a \`**Outcome:**\` line ANYWHERE either — the outcome is decided after your last audit, so an outcome you author is stale beside the real one, and one above your first heading is refused with the headings.

CITATION HYGIENE, because red must be able to VERIFY every source at the leaf: the url you cite carries the FULL coordinate — owner/repo#N or a full URL, never a bare #N red cannot resolve without the repo — and every QUOTED span records a locating anchor (its section heading, or a nearby unique phrase) so red can grep it verbatim in the source. An unlocatable quote is graded LOW through no fault of the source. Two claims resting on the same url are one source and two anchors, not two sources.

REPORT WRITE PATH: the harness refuses a Write to ANY file whose name contains the word "report" (report_blue_synthesize.md was refused as surely as report.md) — draft it in ${runDir}/scratch/ under a name that starts with your seat id and lacks that word, then bash cp it to ${runDir}/blue/report.md, the one path ingest reads (nothing waits there; a report.md anywhere else is not the report); a direct Write of the report path fails and wastes a round-trip. FREEZE THE REPORT FIRST, once report.md is written — run the act your seat's help lists as freezing the report into the record as its base. Do this ONCE; if it refuses, the tool records the refusal — log what you expected. Citing a source, proving a computation and every later change read the frozen report, so each follows the freeze. THE FREEZE TAKES THE FILE: once frozen, the report is in the record and no longer at that path — a file tool finds nothing there, so read it and change it by your seat's own acts. ${speedClause}${recordClause('blue-synthesize')} Return the blue envelope.`,
  { ...dispatchAs('blue-synthesize'), label: `blue-synthesize · ${slug}`, phase: 'Blue', schema: SYNTHESIS_ENVELOPE })

if (!blueEnv) throw new Error('blue synthesis returned null (agent failed) — aborting cleanly')
// ---- Debate loop: red audits gate; termination is the record's, and the engine refuses to spin ----
// ─── THE DEBATE: chair sits → the record says who sits → they sit → the chair sits again ───────────
//
// There are no rounds (plans/roundless.md §III.B.1). Each chair sitting opens an epoch: the chair
// runs `dispatch next`, which computes readiness FROM THE BOARD and records one dispatch per party;
// the chair's envelope is the verb's JSON; this loop dispatches what the plan says and nothing
// else. A bench convened for petitions sits first, then red parties (in parallel), then blue, then
// the bench on its docket — an exchange is a red-party
// sitting followed by a blue-party sitting, and the record counts them (impasse.go). Empty is the
// termination signal: pass_permitted (the chair issues PASS → VERIFIED), ceiling (the run's epoch
// limit reached → CEILING), or
// neither (UNVERIFIED, with the plan's reasons on the record). The disputes, the docket, impasse,
// the ceiling and the epoch limit all live on the record; nothing here keeps a second copy of them.
//
// ONE STOP IS THE ENGINE'S OWN: THE NO-PROGRESS VALVE. A plan the record computes the same for
// NO_PROGRESS_EPOCHS chair sittings in a row is a loop that nothing on the board is moving — the same
// parties readied against the same head with the same exchange counts, sitting after sitting, while
// the run spends. It stops the debate UNVERIFIED and names the stuck parties and the head. Not
// CEILING: no term ran out; the parties were still ready, and the record says so.
//
// THE SAME IS THE PLAN'S KEY. The record computes it over the whole plan but the fields that say
// when it was asked (record.Plan's `key:"-"` tags), and capture compares every relayed plan with
// the stored one, key included, so the valve reads one field a chair cannot alter unseen. A blocker
// that changes alone — a motion filed or ruled, a source read — moves the key, and that is movement
// on the board.
//
// THE MARCH IS CARRIED BY `exchanges`. A gap marching to impasse keeps its head and its parties
// while its exchange count advances — progress the run's terms (k, k-max) bound, which a valve
// blind to the counts would cut short under any k of 3 or more. A seat that is dispatched and opens
// no sitting moves no count, so its plan repeats exactly: the loop this valve exists for.
const NO_PROGRESS_EPOCHS = 3
const partyList = (ps) => ps.map((p) => `${p.seat_id}${p.gap_ids && p.gap_ids.length ? `[${p.gap_ids.join(' ')}]` : ''}`).join(', ')
const sortedStrings = (xs) => (Array.isArray(xs) ? xs : []).map(String).sort()
let samePlan = { key: null, epochs: 0 }
let noProgress = null // { epochs, head, parties } when the valve stopped the debate
let epoch = 0
let lastPlan = null
const infraDebts = [] // defect_owed_elsewhere rulings (W1.9) — the bench's named debts, surfaced at assembly and in the final envelope

const ledgerClause = ` THE REPORT DOES NOT NAME ITS SOURCES — it carries invisible anchors, and the evidence layer is the only way from one to what it points at. Read it FIRST. A claim blue backed with a source resolves to the url, title and sha256 of the bytes blue actually read, and THAT is what you re-fetch to audit the same artifact rather than a page that may have drifted since. A claim blue COMPUTED resolves to its proof's re-run and verdict; an unjudged one looks clean until checked. A citation whose evidence entry lists pages is OCR text — a machine's reading of page images, which can misread: check it against one page's image and record which page. Your own verifications come back ATTACHED TO THE SOURCE you checked: an empty one is a citation NOBODY has checked yet, and that is where your next pass is worth most. WHAT NOT TO RE-CHECK, because it is the sitting's whole economy: a claim verified at HIGH confidence in a prior sitting STAYS verified — do not re-fetch it UNLESS your work list says its sentence changed since, OR more than 2 epochs have elapsed since it was last verified, OR its access date and the source's volatility suggest drift (living documents, issue trackers, README stats). RECORD every claim you verify. A SOURCE YOU FOUND YOURSELF IS A DIFFERENT ACT — blue never cited it — and it answers a different question: whether the claim is true in the WORLD, where verification asks only what blue's source did for it. VERBATIM READS ONLY — you have no WebFetch, by design: it returns a small model's SUMMARY, not the source. Read blue's cached bytes through the tool; for a source YOU discover, pull it verbatim through the run's cached read, so every seat after you reads the same bytes (an issue thread through \`gh issue view <n> --comments\`), and read it. WebSearch is for DISCOVERY — finding the url — never for the read that grades a citation. If a source is too large for your context, read it in SECTIONS and name the sections you read: A TRUNCATED READ IS NOT A READ, so state the truncation and never grade a body you could not fully read.`
// THE LENS: it audits its area and puts what it finds on the board ITSELF (plans/roundless.md
// §III.B.3). The duties of the seat that mints — asking for the answer to be PRODUCED, classing
// the probe, reading the script, keeping lineage — sit here, with the minter.
// THE LENS'S SITTING IS READ FROM THE RECORD, NOT FROM THE RELAY. A head copied into this prompt from
// the chair's relay could be present and wrong and nothing would object; the lens's work view carries
// its last sitting as a field with an explicit kind, computed from the record (lastSittingBefore).
// THE RE-READ IS SCOPED HERE, WITH THE KIND THAT DECIDES IT, and not stated flat in the prompt body.
// Measured 2026-09-22: the body said "RE-READ THE FULL REPORT IN CONTEXT" unconditionally, one
// sentence after this clause explained what `unchanged` means. Six of seven lenses dispatched against
// an unmoved report re-read it whole and recorded nothing, at 7-30 calls each — the constitution had
// already been narrowed to excuse them, so the two rules disagreed and the unconditional one won.
const lastReadClause = ` YOUR PIN IS ON YOUR WORK LIST: read \`sitting.last_read\` first — its kind decides what to audit. \`first\`: audit the report in full — whole and in context, never a diff; over one Read call, read it whole in consecutive windows. \`behind\`: the report head moved past your pin (head H; pin P) — audit it as it now stands, in full, likewise. \`unchanged\`: the report is unchanged since your pin (head H) — NOTHING TO AUDIT: your earlier read stands, and where nothing engages you the sitting owes no entry. \`undispatched\`: the record holds no dispatch for you — log that and end the sitting. Only a whole read of the report moves your pin. Whenever the kind is \`behind\` or \`unchanged\`, a FRESH gap on text you already read and passed at an earlier sitting states, in its mint reason, why you missed it then.`
// A PARTY'S CLOSING IS OWED ON EACH DOCKETED GAP (record/closingowed.go). The plan's closing_owed
// blockers name every closing the record still awaits, by the party that owes it, so a sitting is
// told the gaps it owes a closing on from the record — never from the docket this sitting filed,
// which names neither the party nor a docketing a party filed itself.
const closingsOwedBy = (plan, seat) => plan.blockers.filter((b) => b.kind === 'closing_owed' && b.owner === seat).map((b) => b.subject)
const closingClause = (gaps) => gaps.length
  ? ` YOUR CLOSING ARGUMENT IS OWED ON: ${gaps.join(', ')} — each is DOCKETED, and the bench rules it once you and the other party to it have each filed yours. For EACH, argue in ~120 words why it should be decided your way, citing the exact section and evidence: your strongest evidence, and your answer to the other party's case. This is your case; material the record does not carry cannot help you, and a claim it does not support counts against you.`
  : ''
const lensPrompt = (seatID, gaps, closings) => {
  const area = SEATS[seatID].area
  const extra = area === 'evidence' ? ledgerClause : (area === 'logic' || area === 'dark-side') ? steelmanClause : ''
  const engaged = gaps.length
    ? ` YOU ARE ENGAGED ON: ${gaps.join(', ')} — gaps you minted that are still open and below their limits. DID BLUE ACTUALLY DO WHAT YOU ASKED? YOUR WORK LIST SAYS WHETHER BLUE MOVED: each gap carries its location state and the edits that changed its sentence since you last sat. A \`marked\` gap with no edits is unanswered — nothing you read afterwards changes that, so say so and end. \`gone\`: the gap's anchor is not in the report — blue cut its sentence, or, in a migrated run, its quote never placed; that is not silence — judge the report as it stands. Otherwise re-read the report where each is anchored, put your required fix and blue's edits side by side rather than inferring, and act: regrade on what you now see, or close with the verification triple where the repair holds. A gap you neither regrade nor close this sitting is a NULL TURN on it, and null turns count toward impasse — silence is a turn taken.`
    : ''
  return `Red lens sitting, area ${area}, topic "${topic}".${extra}${engaged}${closingClause(closings)}${lastReadClause} ANCHOR EVERY FINDING TO A QUOTED SENTENCE, and quote it exactly rather than paraphrasing: a finding whose quote is not found in the report as the record renders it is REJECTED. The ids of your findings and of the gaps you mint are the tool's to assign. HARNESS NOTES: a search's count mode counts LINES, not occurrences — anchor patterns (e.g. '^### ') when counting.
YOU MINT YOUR OWN GAPS. What you find that is real and belongs on the board goes there as YOUR gap: read the board first: whether it already holds a candidate is your judgement — a defect already closed is a REOPEN and arrives carrying that history or arrives lying; a duplicate minted fresh forks the lineage — then mint, one considered gap at a time, graded on every axis, under its class, or the catch-all with the reason none does. A finding is your graded observation; a gap is your claim on the board. Your budget caps the gaps you hold open at once: the larger of the run's floor (mintBudget in run-config.json) and a rate per unit of what your area audits, read off the record at each mint — a refused mint states the arithmetic. A gap of yours that closes frees a mint: what the budget refuses you record as a finding, and that finding becomes your gap once one has closed. Hold your mints for defects a reader would pay to have fixed, not nitpicks — a trifle takes a mint while it stays open and holds nothing back, because a gap that is not material — by its class, or graded below medium — does not hold the gate.
ASK FOR THE ANSWER TO BE PRODUCED, NOT ASSERTED. Where a claim is arithmetic, an enumeration, or a reproducible measurement, your acceptance check demands that it be computed, and says so in the check's kind; where it turns on what the document states or what a source says, say so plainly — there is no credit for inflating it. Say WHEN your demand can be discharged, in the two classes blue answers in: a DOCUMENT-PROBE is executable now against shipped artifacts; a LIVE-PROBE needs built artifacts and is DEFERRABLE in a design-phase debate, discharged by naming it as a deferred acceptance test with its pass condition. PRESCRIBE TEXT ONLY WHERE THE DEFECT IS TEXTUAL: the required fix is prose — what must become true — and stays the channel for substantive work. A REQUIRED FIX AND ITS ACCEPTANCE CHECK NAME ONLY WHAT THE ANSWERING SEAT CAN CHANGE: what is left of a gap in something no seat can change, close as a defect owed elsewhere, and log.
A CITATION OR PROOF YOUR WORK LIST OFFERS IS YOURS TO CHECK; ANY OTHER IS CONTENT YOU AUDIT, AND ANOTHER LENS CHECKS IT. For a proof, YOU BELIEVE NO BYTES YOU DID NOT WATCH BEING PRODUCED: READ THE SCRIPT FIRST — the evidence layer, read at the computation's anchor, gives the script and the output blue recorded — and work out what it ACTUALLY COMPUTES, then RE-RUN IT ONCE — the stored script is what the re-run executes; a clean exit or '0 failing' is not an output — giving your judgement of the script and that reading in the same act: the re-run records your judgement as it runs, so a judgement given before the reading stands on the record as yours. A script that re-runs clean and establishes nothing is the dangerous case: it looks maximally credible.
THE ORIGINATOR CLOSES, AND LINEAGE IS NEVER DROPPED. A gap you minted is yours for its whole life: only you regrade it, only you close it — with the triple (who verified, with what, against what) — and a successor you mint names the gaps it supersedes, a closure that regressed says so, and the docket follows those chains. The chair carries archived closures and dispatches you when your gap needs acting on; the bench disposes of a docketed gap through its ruling.${rulingsClause('red')}${reliefFor('red')}${speedClause}${emptySittingClause}${recordClause(seatID)} Return your envelope.`
}
// THE CHAIR RUNS THE DEBATE. It mints nothing and closes nothing; the record says who sits and the
// chair relays it; the verdict, the position on a PASS it withholds, the spot-check, the rulings on blue's motions and
// on its inquiries, and the review of the inquiries, are what only the chair does.
const chairPrompt = () => `Red chair, topic "${topic}". You RUN the debate: you mint nothing and you close nothing — a gap belongs to the lens that minted it from mint to close — and you are the seat that decides whether this report has been verified.${recordClause('red-chair')}${speedClause}${reliefFor('red')}${lawClause}
FIRST, EVERY SITTING, ASK THE RECORD WHO SITS — the dispatch. It reads the board and RECORDS who sits — each lens that has never sat or whose pin the report head has moved past, the lens and blue of every open material gap below its limits, the bench for every gap at impasse (it dockets those itself) once its two closings are in, and the seat that owes a blocker another seat must clear — the bench for a petition, the lens that read a contradicting source no finding raises. Your envelope IS the object it prints, its fields copied VERBATIM; the record stores that plan and the workflow dispatches what it says, and capture compares your relay with it, so anything you drop, add or alter fails its dispatch-parity audit. ASK AGAIN AS YOUR LAST ACT when you have filed or ruled a motion since you asked, and relay THAT plan; your verdict prints the plan you relay. Empty is the record's word that the run is over: with pass_permitted nothing another seat must clear holds a PASS; with ceiling the run is at its epoch limit.
THE STOPPING JUDGMENT IS YOURS, AND IT IS NOT CEREMONY. When the plan says pass_permitted, clear what your work list still marks blocking — those items are yours — then decide: record a PASS verdict if you agree the report is verified, or a FAIL if the board as it stands still stops you, which it does until it converges: nothing material open, no material gap minted this epoch, and the open gaps weighing a small fraction of the run's peak. Otherwise record no verdict this sitting. Your recorded verdict is the ONE fact the run's outcome is derived from.
AT final_sitting NOBODY SITS AFTER YOU: decide on the board as it is.
WHEN THE REPORT OR AN INQUIRY HAS CHANGED SINCE YOUR LAST READ, ANSWER EVERY INQUIRY ON ONE READ: read the report ONCE and answer every inquiry against that pass, the way anyone checks a document against a list — not once per inquiry, and from THIS read, because the report is rewritten between sittings; where neither has changed, your last read stands and nothing is owed. RULE ON BLUE'S INQUIRIES: a ruling is an ARGUMENT, not a command — it needs a reason, and blue decides what becomes of the inquiry with that reason in front of it. RULE THE GRADE MOTIONS blue filed: accept and the minting lens owes the regrade; reject and the grade stands — a seat that disagrees with your ruling puts the gap before the bench, so your reason carries the argument.
NEVER RE-DERIVE THE BOARD IN YOUR HEAD: the board, work and motions projections are the reads, and the plan is the record's, not yours.`
const bluePrompt = (gaps, closings) => `Blue response, topic "${topic}".${gaps.length ? ` You are engaged on: ${gaps.join(', ')}.` : ''}${closingClause(closings)}${reliefFor('blue')} READ YOUR MANUAL FIRST; then pull your working set — the report and your work list — together, in one message.${recordClause('blue-respond')}${speedClause}${rulingsClause('blue')}
YOUR WORK LIST CARRIES YOUR GAPS, EACH AS ONE WORK ITEM: its lens's problem, required fix and acceptance check, its sentence as the report now holds it, and — where the lens proposed exact text — that text beside the span it replaces, with whether the tool checked the span against the report; the gaps the bench has ruled closed are on the list too, each with its fate, and any gap the bench REMANDED comes with a stated research direction you owe. The gap ids you were handed above only name your work; the work items carry it, and the record is authoritative. A row you got wrong stands for this sitting; your next sitting's row for that gap answers it — never move its text into a log.
YOU MAY COMPUTE AN ANSWER, NOT ONLY COLLATE SOURCES. You have Bash, Write and Edit, and for a whole class of questions running something settles it faster and harder than arguing about it. WORKING IT OUT IN YOUR HEAD IS NOT THE EXCEPTION — IT IS THE CASE THIS EXISTS FOR: a correct figure with no derivation is indistinguishable from a confident guess. A gap can be WAITING ON A PROGRAM FROM YOU — it says so (a computation check), and where it does, no amount of prose will close it. DOCUMENT-PROBE checks you discharge now; a LIVE-PROBE you discharge by naming the deferred acceptance test and its pass condition.
YOUR INQUIRIES ARE A LIVING RECORD, NOT AN OPENING PLAN. Every sitting, revisit what is still open and say what became of it and how. The hypothesis is what makes that honest — an inquiry given its fate against its hypothesis is evidence of choosing; one set down on a shrug is not. Red rules on your proposals, and a ruling is an ARGUMENT, not a command: weigh its reason, then decide what becomes of the inquiry yourself — an inquiry you still believe in is pursued against the ruling, and the account you give of its fate is where that shows.
WHERE RED PROPOSED EXACT TEXT you have THREE paths and you are not obliged to take the first: apply it verbatim, counter-edit with your own fix, or dispute it. Say plainly which path you took and why. A sitting in which you never decline is not agreement, it is capitulation. Applying red's exact text ESTOPS red from re-raising that text as a fresh gap, so verbatim application is a real settlement, not a surrender.
OWNERSHIP BINDS, AS IT DID AT SYNTHESIS: you write ONLY your own surfaces — TL;DR, Catechism, the report's own sections, open questions, your citations. THEIR ORDER BINDS TOO: \`## TL;DR\` first, \`## The Catechism\` second, the report's own sections after them, \`## Open questions\` last. The report ships in the order it is written, so an edit that puts one of them out of place — a section added after the open questions, say — is refused. You MUST NOT introduce a \`**Outcome:**\` line or any tool-owned section (${TOOL_COMPOSED}, a \`## Blue team report\` wrapper): assembly composes those from the record, and an edit that would leave one in the report is refused; the refusal says how content reaches that section, and the assembled layout says it for every section.${NO_LIST_OF_APPROACHES} AND WHICH MODEL IS ANSWERING THIS RUN IS NOT A FACT YOU HOLD: you can see what was REQUESTED, never what replied; if your argument turns on the models used, say that the record's measurement decides it. GRADING SEMANTICS (mapping ${MASS_MAPPING_VERSION}): likelihood and impact are CONSEQUENCE axes — how likely the harm lands and how much it costs when it does — and severity is red's overall grade; they multiply into mass, and materiality is the class's default: always, never, or by grade from medium.
ANSWER EVERY GAP YOU ARE ENGAGED ON, ADDITIVELY, in the report through \`edit\` — expand and repair where red is right (each edit naming the gap it answers, so the repair joins to it), REBUT IN WRITING WITH EVIDENCE where red is wrong, and argue risk-acceptance where the fix's complexity exceeds its likelihood x impact — or, where the gap changes no reader decision or asks for complexity that does not pay, argue \`defect_accepted\` with that reason. DISPUTE RED'S GRADING WHERE YOU DISAGREE WITH IT (a grade motion on the axis, with the grade you say it should be and your evidence; where the chair rejects it and you still disagree, put the gap before the bench with a docket motion, which the bench rules). Compact and reorganize prose as clarity demands — but a CLAIM leaves only by being retired on the record, which names it, why it goes, and what replaces it. PROPAGATE EVERY CORRECTION TO ALL SITES that state the corrected claim, not only the flagged sentence — a bare corrected FIGURE needs a report-wide sweep of its own. A sitting in which you record nothing on a gap you were engaged on is a NULL TURN on it, and null turns count toward its impasse: silence is a turn taken.
AUDIT YOUR OWN REPAIRS, ONE RECEIPT PER GAP (W2b; your constitution carries the full standard): figures recomputed, universals enumerated, consistency sites swept report-wide — one manifest row per gap you REPAIRED — one an edit of yours this sitting answers — and the manifest array in your envelope names them; a gap you rebut without an edit owes none. A gap you are engaged on that you FOUND CLOSED when you sat — its lens sat before you and closed it — owes no row: your work list names each one, with the seat that closed it, its fate and its open successors, and you name it in found_closed from that list. Engage your next sitting on each successor it owes, and name it in successors.`
const benchPrompt = (gaps) => `Adjudication, topic "${topic}". Docketed for you: ${gaps.join(', ')} — each reached impasse under the run's terms and the record docketed it. THE DOCKET IS A ROUTING LIST, NOT THE EVIDENCE. It carries ids; a gap's problem text and its acceptance check live on the board, and you read them FRESH before ruling. Re-run each document-probe acceptance check against the artifact AS IT NOW STANDS, and rule on what you find rather than on what any snapshot asserts.
YOUR RULING BASIS IS CONFINED TO THREE THINGS: the two parties' recorded closings, the full transcript, and the final state of the artifacts — the board and the report as the record now renders them. Weigh each closing as that side's best case, and a claim in a closing that the record does not support counts AGAINST the side that made it. For every ruling on a gap with a lineage chain, READ THE NAMED ANCESTORS' RECORDS first and NAME what you read in your rationale.${lawClause}${declareClause}${inspectionClause}
Every docketed gap gets a written ruling — the docket ruling: its fate, the principle you applied, the values in tension (where none pulled, the ruling's weakest point: what a party would argue against the rule you applied), what a human should look at again (where nothing needs a human, what on the record already settles it), your reasoning, and TWO THINGS THE FATE CANNOT SAY — the proposition you are barring as settled, and, for a ruling that closes the gap, what would reopen it or that nothing would (final). Two fates send the work on rather than ending it: a gap you REMAND was never closed, so a remand is never final — it goes back to the debate for ONE more exchange between its minting lens and blue, on the research direction you state, which the record refuses blank — if that exchange leaves it at impasse it comes back to you, to be closed: a second remand at impasse is refused, and its refusal names the dispositions that close the gap; and a valid finding whose FIX is owned outside the debate — run tooling, the harness, the engine — leaves the board and ships as a NAMED infrastructure debt (defect_owed_elsewhere), recorded and never dropped. Rule every motion your work list names as the bench's. A bench sitting that rules nothing is a workflow error: the run cannot end in a verdict while a docketed gap stands unruled.${speedClause}${recordClause('judge', OCCASION.docket)} Return your envelope.`

phase('Red')
const sittings = {} // seat -> how many times this loop has dispatched it, for the labels
const labelFor = (seat) => { sittings[seat] = (sittings[seat] || 0) + 1; return `${seat} #${sittings[seat]} · ${slug}` }
while (!halted) {
  epoch++
  const chairEnv = await agent(chairPrompt(), { ...dispatchAs('red-chair'), label: labelFor('red-chair'), phase: 'Red', schema: CHAIR_ENVELOPE })
  if (!chairEnv) throw new Error(`red-chair sitting ${epoch} returned null (agent failed) — aborting cleanly`)
  const plan = chairEnv
  requirePlan(plan, epoch)
  lastPlan = plan
  log(`epoch ${epoch}: head ${plan.head} — ${plan.parties.length} party(ies) ready${plan.docket.length ? `, docketed ${plan.docket.join(', ')}` : ''}${plan.verdict ? `, chair recorded ${plan.verdict}` : ''}${plan.pass_permitted ? ' — PASS permitted' : ''}${plan.ceiling ? ' — at the ceiling' : ''}${chairEnv.notes ? ` — chair: ${chairEnv.notes}` : ''}`)
  if (plan.verdict === 'PASS') break
  // THE EPOCH LIMIT is the record's word, like the ceiling: the plan at the last epoch readies
  // nobody. A relayed plan that says the limit is reached ends the debate whatever parties it lists.
  if (plan.epoch_limit_reached) {
    log(`epoch ${epoch}: the epoch limit (${plan.max_epochs}) is reached — nobody further is dispatched`)
    break
  }
  if (plan.parties.length === 0) break
  const key = plan.key
  samePlan = key === samePlan.key ? { key, epochs: samePlan.epochs + 1 } : { key, epochs: 1 }
  if (samePlan.epochs >= NO_PROGRESS_EPOCHS) {
    noProgress = { epochs: samePlan.epochs, head: plan.head, parties: plan.parties.map((p) => ({ seat_id: String(p.seat_id), gap_ids: sortedStrings(p.gap_ids) })) }
    log(`epoch ${epoch}: NO PROGRESS — the plan is the same for ${samePlan.epochs} consecutive epochs (NO_PROGRESS_EPOCHS = ${NO_PROGRESS_EPOCHS}) at head ${plan.head} — key ${key}: the same parties, docket, exchange counts and blockers; stuck: ${partyList(noProgress.parties)} — the debate stops rather than dispatch them again`)
    break
  }

  // Classified by seat id, never by object identity: a plan relayed through a host runtime may
  // hand out a fresh wrapper on every read, so two looks at one party need not be === equal.
  // A lens and the bench are parties by ROLE, read from the roster; blue is a party as its responding
  // seat alone — the frontier and the synthesis are blue seats no plan names.
  const roleOfParty = (p) => {
    const role = roleOfSeat(String(p.seat_id))
    return role === 'lens' || role === 'bench' ? role : p.seat_id === 'blue-respond' ? 'blue' : 'stray'
  }
  const lenses = plan.parties.filter((p) => roleOfParty(p) === 'lens')
  const blues = plan.parties.filter((p) => roleOfParty(p) === 'blue')
  const benches = plan.parties.filter((p) => roleOfParty(p) === 'bench')
  const strays = plan.parties.filter((p) => roleOfParty(p) === 'stray')
  if (strays.length) throw new Error(`epoch ${epoch}: the plan names ${strays.map((p) => p.seat_id).join(', ')}, which this workflow has no sitting for — the cast and the workflow disagree`)

  // A PETITION IS HEARD BEFORE ANY PARTY OF THE EPOCH SITS — the epoch after its filing, because
  // the record reaches this script only through the chair's plan. A halt ends the run here.
  if (benches.some((p) => p.occasions.includes(OCCASION.petition)) && await hearPetitions()) break

  if (lenses.length) {
    log(`epoch ${epoch}: dispatching ${lenses.length} red lens(es): ${lenses.map((p) => `${SEATS[p.seat_id].area}${p.gap_ids.length ? `[${p.gap_ids.join(' ')}]` : ''}`).join(', ')}`)
    await parallel(lenses.map((p) => () => agent(lensPrompt(p.seat_id, p.gap_ids, closingsOwedBy(plan, p.seat_id)),
      { ...dispatchAs(p.seat_id), label: labelFor(p.seat_id), phase: 'Red', schema: LENS_ENVELOPE })))
  }
  // BLUE SITS AGAIN ON A SUCCESSOR (gblock's ruling R2): a gap blue found closed whose successor the
  // record names engages blue's next sitting on it, and blue relays the ids. THE BOUND: one more
  // sitting this epoch, on successors no sitting of blue's this epoch was engaged on — each at most
  // once. The record refuses an engagement on a gap it does not name as a successor, or on one
  // already engaged this epoch, so a relayed id the record does not hold reaches no dispatch row.
  const blueSittings = blues.map((p) => ({ gap_ids: p.gap_ids, again: false }))
  const blueEngaged = new Set(blues.flatMap((p) => p.gap_ids))
  for (const p of blueSittings) {
    phase('Debate')
    const closings = p.again ? [] : closingsOwedBy(plan, 'blue-respond') // owed once, at the sitting the plan dispatched
    const blueEnv2 = await agent(bluePrompt(p.gap_ids, closings), { ...dispatchAs('blue-respond'), label: labelFor('blue-respond'), phase: 'Debate', schema: BLUE_ENVELOPE })
    if (!blueEnv2) throw new Error(`blue response (epoch ${epoch}) returned null (agent failed) — aborting cleanly`)
    const foundClosed = new Set((Array.isArray(blueEnv2.found_closed) ? blueEnv2.found_closed : []).filter((g) => p.gap_ids.includes(g)))
    const open = p.gap_ids.filter((g) => !foundClosed.has(g))
    const successors = [...new Set(Array.isArray(blueEnv2.successors) ? blueEnv2.successors : [])]
    const again = foundClosed.size > 0 && !p.again ? successors.filter((g) => !blueEngaged.has(g)) : []
    if (again.length) {
      again.forEach((g) => blueEngaged.add(g))
      blueSittings.push({ gap_ids: again, again: true })
      log(`epoch ${epoch}: blue found ${[...foundClosed].join(', ')} closed and engaged its next sitting on the successor(s) ${again.join(', ')} — blue sits again this epoch`)
    }
    const unsat = successors.filter((g) => !again.includes(g))
    if (unsat.length) log(`epoch ${epoch}: blue named successor(s) ${unsat.join(', ')} it does not sit again on — ${p.again ? 'this was its second sitting this epoch' : foundClosed.size ? 'already engaged this epoch' : 'it found no engaged gap closed'}`)
    // A SITTING THAT FOUND EVERY ENGAGED GAP CLOSED HAD NOTHING TO ANSWER. This engine reads no
    // record and re-prompts no seat: the seat's own work list blocks on what it owes — its closings
    // among them (record.closingBlockersOf) — and capture's record-parity and manifest audits read
    // the record.
    if (p.gap_ids.length && !open.length) {
      log(`epoch ${epoch}: blue found every engaged gap (${p.gap_ids.join(', ')}) closed before it sat — nothing to answer`)
    }
    // W2b, OWED GAPS ONLY, NEVER AN ABORT (gblock's ruling on #868). A row is owed for a gap blue
    // REPAIRED — its sitting's edit answers it — while the gap was still OPEN when blue sat; one
    // its lens closed first, or one blue rebutted without an edit, owes nothing. The plan was
    // computed at the START of the epoch and the lenses sat in between, so checking blue against
    // the plan aborted two runs over an empty manifest that was correct. This engine sees neither
    // the closures nor the edits; it logs the open gaps blue named no row for, and capture decides
    // which were owed from the record (record.ManifestOwed) — none, some or all of them alike.
    const covered = new Set(Array.isArray(blueEnv2.manifest) ? blueEnv2.manifest : [])
    const uncovered = open.filter((g) => !covered.has(g))
    if (foundClosed.size) log(`epoch ${epoch}: blue found ${[...foundClosed].join(', ')} closed before it sat — no manifest row owed (capture re-derives this from the record)`)
    if (uncovered.length) log(`epoch ${epoch}: manifest rows named for ${open.length - uncovered.length}/${open.length} gap(s) open when blue sat — no row for: ${uncovered.join(', ')} (owed where this sitting's edit answered it; scored at capture)`)
    log(`epoch ${epoch}: blue responded${p.gap_ids.length ? ` on ${p.gap_ids.join(', ')}` : ''}${closings.length ? `; its closing owed on ${closings.join(', ')}` : ''}`)
  }
  for (const p of benches.filter((b) => b.occasions.includes(OCCASION.docket))) {
    log(`epoch ${epoch}: the bench sits on ${p.gap_ids.join(', ')}`)
    const judge = await agent(benchPrompt(p.gap_ids), { ...dispatchAs('judge'), label: labelFor('judge'), phase: 'Debate', schema: JUDGE_ENVELOPE })
    if (!judge) throw new Error(`bench sitting (epoch ${epoch}) returned null (agent failed) — aborting cleanly`)
    for (const r of judge.dispositions || []) {
      // AN EMPTY `settled` BARS NOTHING, so it carries no proposition: rendered, it would hand red
      // "ESTOPPED on each settled proposition" over a blank, an estoppel nobody ruled. The record
      // accepts the blank; the prompt omits it.
      const settled = typeof r.settled === 'string' && r.settled.trim() ? r.settled : undefined
      // A REMAND'S DIRECTION IS NOT TAKEN FROM HERE. This array is the bench's account of what it
      // recorded; the record holds the direction, refuses it blank, and the plan relays it in
      // remand_owed for exactly the sittings that owe the exchange. A remand reopens nothing, so
      // its row carries no reopens_on and no final.
      rulingsInEffect.set(r.gap_id, r.disposition === 'remanded'
        ? { gap_id: r.gap_id, disposition: r.disposition, settled, epoch }
        : { gap_id: r.gap_id, disposition: r.disposition, settled, reopens_on: r.reopens_on, final: !!r.final, epoch })
      if (r.disposition === 'defect_owed_elsewhere') infraDebts.push({ gap_id: r.gap_id, owed_fix: r.rationale, epoch })
    }
  }
}

// THE RUN'S TERMINAL WORD IS A RECORD VOCABULARY: RUN_OUTCOME, in the generated vocabulary region.
// The word travels into the assembly seat's prompt ("Record it as the run's outcome —
// ${verdict}") and is typed straight at `bench outcome --as`, so a word the record refuses arrives
// as an INSTRUCTION and the engine carries on as though it had been written. outcomeWord makes a
// misspelling throw here.
const outcomeWord = (w) => {
  if (!RUN_OUTCOME.enum.includes(w)) throw new Error(`not a run outcome: ${w} (have ${RUN_OUTCOME.enum.join('|')})`)
  return w
}

// A NO-PROGRESS STOP IS UNVERIFIED, NEVER CEILING. CEILING is derived on the record — the epoch limit
// reached — and `bench outcome` refuses it over a board that is not; a stalled plan still has parties ready. UNVERIFIED is the word for a
// run that stopped before its record reached a terminal state, and the why says what stopped it.
// The outcome is read twice: at the end of the debate, and again if the terminal sitting halts.
const verdictNow = () => halted ? outcomeWord('HALTED')
  : (lastPlan && lastPlan.verdict === 'PASS') ? outcomeWord('VERIFIED')
  : noProgress ? outcomeWord('UNVERIFIED')
  : (lastPlan && lastPlan.ceiling) ? outcomeWord('CEILING')
  : outcomeWord('UNVERIFIED')
// A RUN THAT ENDS OVER AN UNCHECKED CITATION OR PROOF, OR ONE ON A FAILED CHECK, NAMES IT, as the
// record's own derived reason does (verdict.go anchorsStanding): the run ends with an anchor nobody
// read against the sentence the report carries, or one whose check failed. Those two kinds only, and
// nothing when there is none.
const anchorsStanding = (plan) => {
  const unchecked = plan.blockers.filter((b) => b.kind === 'unchecked').map((b) => b.subject)
  const failed = [...new Set(plan.blockers.filter((b) => b.kind === 'failed_check').map((b) => b.subject))]
  return (unchecked.length ? `; unchecked: ${unchecked.join(', ')}` : '') + (failed.length ? `; on a failed check: ${failed.join(', ')}` : '')
}
const terminationWhyOf = (verdict) => halted ? 'judicial halt'
  : verdict === 'VERIFIED' ? 'the chair recorded PASS with the board permitting it'
  : noProgress ? `no progress: the dispatch plan was the same for ${noProgress.epochs} consecutive epochs (NO_PROGRESS_EPOCHS = ${NO_PROGRESS_EPOCHS}) at head ${noProgress.head}, its parties, docket, exchange counts and standing blockers unchanged — ${partyList(noProgress.parties)} readied again each time with nothing on the board moving`
  : verdict === 'CEILING' ? `epoch limit ${lastPlan.max_epochs} reached — the run's term on chair sittings; the parties the board still readied were not dispatched` + anchorsStanding(lastPlan)
  : (lastPlan ? `nobody was ready and neither PASS nor CEILING held: ${lastPlan.why.join('; ')}${anchorsStanding(lastPlan)}` : 'the run ended before any dispatch')
let verdict = verdictNow()
let terminationWhy = terminationWhyOf(verdict)
log(`debate ended: ${verdict} after ${epoch} chair sitting(s)${halted ? ' (JUDICIAL HALT)' : ''} — ${terminationWhy}`)

// THE TERMINAL BENCH SITTING FIRES ON THE RECORD'S LIST, NOT ON A COUNT A SEAT REPORTS. Every exit
// from the loop above follows a chair's plan with nobody dispatched after it, so the last plan's
// blockers are the board at the exit: a blocker the bench owns — a motion whose gavel is the
// bench's — is what this sitting rules. A chair-ruled motion (grade, inquiry) never convenes it: the
// bench has no ruling verb for one. The plan the chair relays is the one it asked for LAST in its
// sitting, and capture compares every relayed plan with the one the record stored; `bench outcome` refuses
// while a bench-owned motion stands, so one filed after it is ruled at assembly.
//
// A PETITION STANDING HERE IS HEARD HERE, and its ruling and a halt travel as the petition sitting's
// do (takePetitionRulings): a halt stamps the run HALTED.
const benchBlockers = lastPlan ? lastPlan.blockers.filter((b) => b.kind === 'unruled_motion' && roleOfSeat(b.owner) === 'bench') : []
if (!halted && benchBlockers.length) {
  const terminalJudge = await agent(
    `Terminal disposition for topic "${topic}" (debate ended ${verdict} after ${epoch} chair sitting(s); this sitting fires at the exit boundary). ${benchBlockers.length} motion(s) whose gavel is yours stand unruled on the record: ${benchBlockers.map((b) => b.subject).join(', ')} — each is on your work list. Read the transcript in full and the board back, and rule each with a reason; the chair's own motions are not yours, and the report shows them as they stand. NOTHING CAN BE REMANDED AT A TERMINAL EXIT — there is no next sitting to remand it to. A PETITION among them is ruled granted | denied, and each petition ruling also returns on the envelope's \`rulings\` — its petitioner and class, your ruling and opinion, and on a grant the relief and whom it binds. ${haltClause('stamps the run HALTED')}${lawClause}${declareClause}${inspectionClause}${speedClause}${recordClause('judge', OCCASION.terminal)} Return your envelope.`,
    { ...dispatchAs('judge'), label: `judge · ${OCCASION.terminal} · ${slug}`, phase: 'Assemble', schema: TERMINAL_ENVELOPE })
  if (terminalJudge) {
    if (takePetitionRulings(terminalJudge)) {
      verdict = verdictNow()
      terminationWhy = terminationWhyOf(verdict)
      log(`the terminal sitting halted the run: ${verdict} — ${terminationWhy}`)
    }
  }
}

const ASSEMBLE_ENVELOPE = {
  type: 'object',
  additionalProperties: false,
  required: ['open_gaps'],
  properties: {
    open_gaps: { type: 'integer', minimum: 0 }, // the board's counts.open, not red's docket
  },
}
phase('Assemble')
const assembleEnv = await agent(
  `Final assembly for topic "${topic}", run directory ${runDir}. Debate outcome: ${verdict} after ${epoch} chair sitting(s) — ${terminationWhy}. THE REPORT IS ASSEMBLED FROM THE RECORD — you author NOTHING, you copy NOTHING, you fill in NO inputs. There are no <FILL> fields and no sections for you to write; do not hand-write report.md and do not copy anything into it yourself.

FIRST, STAMP HOW THIS RUN ENDED: it is ${verdict}${halted ? `, ended by JUDICIAL HALT — carry the opinion verbatim: ${haltOpinion}` : ''}${verdict === 'CEILING' ? `, ended at the CEILING: ${terminationWhy} — a limit the run was set up with, NOT a judged failure to verify` : ''}${verdict === 'UNVERIFIED' ? (noProgress ? `, ended UNVERIFIED — the engine stopped a debate that was making no progress: ${terminationWhy}. It is not CEILING: no term ran out; the parties were still ready` : `, ended UNVERIFIED — nobody was ready and neither PASS nor CEILING held; the plan's reasons are the account: ${terminationWhy}`) : ''}. Record it as the run's outcome — ${verdict}${verdict === 'CEILING' ? ', ended at the ceiling' : ''} — with WHY this outcome is the right read of the board as the reason — the judgement, never a recap of the sitting; the verdict is derived from the record and the stamp must agree with it. A RESUMED ASSEMBLY SITTING — one sat again after an earlier one stamped the outcome and stopped before it assembled — finds that outcome already on the record, and the record refuses a second: the stamp is done, so go on to what follows.

THEN, TWO THINGS YOU MAY HOLD AND THIS IS YOUR LAST CHANCE TO RECORD EITHER. If you hold something that binds how the RECORD IS READ but moves no gap — a construction of a term, a correction of what the record MEANS rather than what it says, a holding worth offering as precedent — state it, and state it in its own right rather than folding it into an unrelated rationale. And if anything in this run needs A HUMAN to re-examine it — an unresolved tension, a claim that held only because nobody could reach the source, a boundary you ruled close to — say so. You keep no memory between runs, so this is the whole of your continuity.

THEN ASSEMBLE. It writes the run's documents for the HUMAN reader — the research, the board, the transcript, the judgments, the inquiries, the evidence, the run's account, the changelog, an index and a tabbed site — and a document with nothing in it is not written at all, so no judgments document means no motions were filed rather than a failure. It prints the verdict it stamped from the outcome on the record: confirm that is the outcome you recorded. Do not open the documents — they are the human's, and the report is read with the record tool. A tool cannot mis-author a synthesis surface — the TL;DR and the catechism are blue's, inside the audited report. An open gap that is not material stays open on the board and in the risk matrix; the chair's PASS listed it by class, on the record. THE AUTHORITATIVE OPEN COUNT IS THE BOARD'S, after every closure and ruling: read it back and report it as open_gaps in your envelope. Infra debts the bench named: ${JSON.stringify(infraDebts)}.${lawClause}${speedClause}${recordClause('judge', OCCASION.assemble)} Return your envelope: open_gaps from the board.`,
  { ...dispatchAs('judge'), label: `judge · ${OCCASION.assemble} · ${slug}`, schema: ASSEMBLE_ENVELOPE })
// AN ASSEMBLY SITTING THAT RETURNS NOTHING STOPS THE RUN, as a null from any other seat does. The
// assembly seat is the one that records the outcome and writes the documents; an envelope returned
// without its word states a verdict nobody confirmed is on the record, with `gaps_outstanding`
// null beside it — VERIFIED over a record that may hold no outcome row.
if (!assembleEnv) throw new Error(`the assembly sitting returned null — it records the run's outcome and assembles the report, and neither is confirmed; the debate had ended ${verdict} (${terminationWhy})`)
return {
  runDir,
  verdict,
  epochs: epoch,
  lanes,
  termination: lastPlan ? { pass_permitted: lastPlan.pass_permitted, ceiling: lastPlan.ceiling, epoch_limit_reached: lastPlan.epoch_limit_reached, why: lastPlan.why, no_progress: noProgress } : null,
  gaps_outstanding: Number.isInteger(assembleEnv.open_gaps) ? assembleEnv.open_gaps : null,
  infra_debts: infraDebts,
  petitions: petitionLog,
  halted,
  halt_opinion: haltOpinion,
}

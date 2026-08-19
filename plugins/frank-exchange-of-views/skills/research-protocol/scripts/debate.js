export const meta = {
  name: 'frank-exchange-of-views',
  description: 'Adversarial research debate: additive blue builds, gate-keeping red audits, judged termination, union-not-summary assembly',
  phases: [
    { title: 'Frontier', detail: 'hypotheses before searches' },
    { title: 'Blue', detail: 'best-of-N lanes + additive synthesis' },
    { title: 'Red', detail: 'per-lens audits + merged verdict' },
    { title: 'Debate', detail: 'revision rounds until red-PASS or judged deadlock' },
    { title: 'Assemble', detail: 'final report by union' },
  ],
}

// TESTING & MODEL-SELECTION STRATEGY (learned from runs 1-3, ~5M tokens of tuition):
//   Logic bugs are unit-testable for zero tokens: a Node harness stubs agent() with canned
//   envelopes and drives every branch (args parsing, round loop, contested docket, deadlock,
//   ceiling, null returns). Founding regressions: stringified args -> undefined paths (run 1);
//   missing null-guard on agent() returns (run 2); lineage-blind docket + degenerate
//   FAIL-with-empty-gaps + friction lost on throw (run 3 retrospective, report §3 rows 20-24).
//   Behavior needs live agents — /research --smoke (1 lane + laneFloorOverride, 1 round,
//   model=haiku) exercises the pipeline for ~50k tokens.
//   Model tiers are REQUIRED — both `model` and `judgmentModel` must be set explicitly. The engine
//   throws rather than guess a tier or inherit the session model: a silently expensive (or silently
//   cheap) tier was the #111 trap. sonnet for development; --smoke sets BOTH to haiku. NEVER change
//   `model` OR `judgmentModel` on a resume — they change agent() opts, bust the cache keys, and
//   re-run completed rounds at full price.
//   Per-role split (efficiency doctrine: cheapen redundancy and mechanics, never judgment or
//   the adversary): `model` drives the BULK seats (frontier, blue lanes, red lenses, blue
//   responses); `judgmentModel` drives the JUDGMENT seats (blue-synthesize, red-merge,
//   lead-judge, assemble). Neither inherits — every run, keeper or dev, names both.
//   KNOWN TRADEOFF (retrospective §3 row 16b): red LENSES ride the bulk tier — on a cheap-model
//   dev/smoke run, treat lens-sourced gap grades with a confidence discount. For keeper runs,
//   name a STRONG model for BOTH tiers so the adversary and the bench run at full strength.
// TERMINATION IS JUDGED, AND THE STANDING PRACTICE IS STOP-AND-RESUME (run-4 report §1.4-1.5):
//   the demonstrated ~$0 terminator is the operator stopping the run and resuming with a
//   reduced maxRounds — cache replay skips every completed agent; only the honest UNVERIFIED
//   assembly runs live. maxRounds is a COST CEILING, never the terminator of record; the
//   automatic severity-floor stop was REJECTED by the run-4 debate (it automates the one call
//   that belongs to judgment). The per-round board-telemetry line (below) is the signal the
//   stopping judgment reads. NEVER change model/judgmentModel on that resume.
// The lead is a script: mechanics, round-keeping, termination. All file writes
// belong to the agents (the filesystem is the blackboard; the script has no
// filesystem access by design). Judgment calls go to lead-judge, never round-to-round.
// Defensive arg handling: args may arrive JSON-encoded (resume path); a stringified
// object destructures to undefined and every agent gets literal 'undefined' paths —
// the exact harness defect red graded high/high/low in run 1. Parse, then guard.
const a = typeof args === 'string' ? JSON.parse(args) : args
const { topic, runDir, lanes = 3, maxRounds = 12, model = null, judgmentModel = null, laneFloorOverride = null, binDir = null, scorecards = null, gapPatterns = null, transcriptDir = null } = a
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
  throw new Error(`debate: refusing dispatch — judgmentModel unset. The engine does not inherit the session model: pass judgmentModel (the JUDGMENT tier — blue-synthesize, red-merge, judge, assemble), e.g. { judgmentModel: "sonnet" }.`)
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
const bulk = { model }
const judgment = { model: judgmentModel }
const slug = String(runDir).replace(/[\/]+$/, '').split(/[\/]/).pop().replace(/^[0-9-]+_/, '')
log(`researching: ${topic.length > 160 ? topic.slice(0, 157) + '...' : topic}`)
log(`resolved tiers — bulk: ${model}, judgment: ${judgmentModel}`)

// Friction must survive a mid-run throw (retrospective §3 row 24): the envelope copy feeds
// this script's aggregate, the file copy survives an abort. Both, always.
const frictionClause = (who, role) => ` FRICTION (${who}) — CLOSE THIS CHANNEL BEFORE YOU FINISH, EVERY SITTING, INCLUDING THE ONES THAT WENT WELL: record each capability gap or protocol misfit ON THE RECORD — the verb is in your role's help, and it takes the thing you reached for and the shape the work actually wanted — AND in the envelope's friction field; then, if nothing blocked you, say so on the record through the same verb's EXPLICIT EMPTY FORM, which still owes a reason: what you reached for, and found. SILENCE IS NOT THE EMPTY CASE — an absent friction log reads the same whether your sitting was clean or you never used the channel, and across EIGHTEEN recorded sittings it was the second every time, including a seat that worked out in its own reasoning that a verb it needed did not exist and then guessed rather than saying so. MOST REFUSALS ARE YOURS and are not friction: a wrong verb, a wrong flag, a bad quote — take the correction and move on. A refusal naming a verb you cannot find, or a fact the record holds and no view will show you, IS the finding; report it instead of engineering around it, because a workaround leaves no trace and a missing capability then looks exactly like one nobody wanted.`

// Wall-clock doctrine (run-4 forensics, 2026-07-17): 80% of run time is API rounds at ~24s
// each, and the corpus showed ZERO batched tool calls — every peek paid a full round. A
// Bash spawn additionally carries a measured multi-second fixed floor. Every seat gets this.
const speedClause = ` SPEED: every message you send costs a ~20s round-trip regardless of content — batch INDEPENDENT tool calls into a single message (read several files at once; fire independent fetches together); only serialize calls that truly depend on a prior result. Peek and search files with the native Read (offset/limit), Grep, and Glob tools — NEVER sed/awk/head/tail/cat/grep through Bash for file access (a shell spawn costs 10-100x a native read and buys nothing). KNOWN HARNESS LIMIT (W1.11, on file three times — do NOT re-log it as friction): Glob/Grep may refuse paths outside the session's registered working directories ("Path does not exist") while Read and Bash reach them; for searches under the run directory the SANCTIONED fallback is Bash grep/ls — this is the one exception to the no-shell-file-access rule.`

// The record IS the exchange. Every seat gets an engine-assigned SEAT_ID and every
// act it takes is an event written through the binary. The events are the only copy.
//
// The writer is ONE compiled binary with role subcommands, so the seat is handed its
// ROLE rather than a script path. The role is not decoration — seat identity is bound
// to it, and a seat
// that reaches for another role's verbs is refused. The tool's --help IS the
// seat's record contract: everything listed is permitted, anything absent does
// not exist for that seat and is FRICTION rather than something to work around.
// W2h — the visibility loop's last leg: each chair's headline numbers, IN the
// prompt of the seats that chair governs.
//
// A number computed at capture and filed in feov-memory is measured and still
// invisible; the clause it instruments stays exactly as dead as "confidence
// self-graded" was. The numbers arrive as an ARGUMENT rather than a path because
// this script is sandboxed and cannot read files — setup prints the arg ready to
// pass, so the value a seat sees is the same value the dashboard and the human
// see, parsed from one rendered file.
//
// Classes travel with the numbers on purpose. A benchmark says optimize me; a
// DIAGNOSTIC says this explains you and optimizing it is a defect — red driving
// its grade stability up is stubbornness, not rigour, and a bare number invites
// exactly that.
// MEMORY AS DUTY, delivered by CLASS JOIN (rulebook audit item 8).
//
// Patterns reach a seat by CLASS JOIN, matched to the gap in front of it. Staging the
// whole corpus does not work: reading is not binding, and fifty patterns at seat start
// is a salience problem no amount of instruction fixes.
//
// So the patterns arrive at the DECISION POINT instead, selected by the class of
// the gap actually being repaired. A join, not a search: deterministic, small,
// and auditable — the manifest row records which patterns applied and what
// checking them showed, so a skipped duty is visible rather than assumed.
//
// The index arrives as an argument because this script cannot read files
// (verified: require, process and fetch are undefined and import() is refused
// outright). Setup writes inputs/gap-patterns-by-class.json for the launcher.
// LINES OF INQUIRY (rulebook audit item 3). think-around-problem mandates
// exploring genuinely distinct alternatives; terse-communication forbids
// narrating them. Required, invisible, unverifiable — the same dead-letter shape
// as "confidence self-graded", mandated and practised five times in 1,892 lines.
//
// Recording the exploration makes the rule instrumentable, and it preserves the
// class that was always lost: an ABANDONED line of inquiry is a dead end a future run does
// not re-walk, and nothing kept those before — they died in a seat's context
// along with the reasoning that produced them.
const inquiryClause = binDir
  ? ` LINES OF INQUIRY: record each genuinely distinct approach you considered with the line of inquiry verb — PROPOSE it with the approach and the hypothesis that would be true if it paid off, and later MOVE it to its fate with what changed: PURSUED (it became your spine — say what you learned), DEFERRED (worth taking, and NOT BY THIS RUN — say what a later run should pick it up FOR), DECLINED (you weighed it and did not take it), or ABANDONED (you tried it and it died). Every fate but the first two owes a reason, and all four reach the report under their own heading — pursued, deferred, and considered-or-rejected are three sections a reader sees, not one list of winners. A line you have not decided yet stays at \`proposed\`, which is a real state and not a failure to type one: it renders under "Still undecided" and the tool asks you to settle it each round. The abandoned ones matter most: a dead end you record is a dead end the next run does not re-walk, and it is worth more than the tidy conclusion that survived. This is not narration for the report — it is the record that makes "explore the alternatives" checkable instead of self-attested.`
  : ''


 // STEELMAN DUTY (E0.5h): the sections red NEVER gap-anchors are exactly the
// self-critical ones — disconfirming passes, human-gated paths, self-attested
// inventories. Claims AGAINST the design attract no adversary, so the case
// against goes unaudited while the case for is contested every round. Blue's
// recorded lines of inquiry are that surface made concrete: what was declined,
// and whether the stated reason survives contact.
const steelmanClause = ` STEELMAN DUTY: read the exploration space via the \`lines-of-inquiry\` projection (the tool renders it fresh from the line of inquiry events on the record). Blue's DECLINED and ABANDONED lines of inquiry are the case AGAINST its own design, and the measured blind spot is that nobody audits them — a weak reason for declining a strong alternative survives untouched because it reads as humility. Attack the reasons, not just the conclusions: a declined line of inquiry whose stated reason does not hold is a finding, and so is an abandoned one whose obituary is wrong. RECONCILE THE RECORD AGAINST THE DOCUMENT: the directions the report actually took must be the ones on the line of inquiry record. A section built on a line of inquiry that was never proposed is undeclared scope; a line of inquiry still at \`proposed\`, or at \`pursued\` and NOT moved this round, is a decision nobody made. Both are findings. Re-recording \`pursued\` WITH what it learned is a legitimate reaffirmation and settles the line for that round — do not read it as neglect; \`deferred\` is likewise a decision (kept for a later run), not an omission. Measured over six runs: 83 of 86 lines of inquiry were declared in round 0 and NONE was ever revisited, so 'pursued' meant 'I intend to' and nothing could falsify it.`

const patternsForGaps = (gaps) => {
  if (!gapPatterns || !gaps || !gaps.length) return []
  const classes = [...new Set(gaps.map((g) => g && (g.class || g.gap_class)).filter(Boolean))]
  const seen = new Set()
  const out = []
  for (const c of classes) {
    for (const p of gapPatterns[c] || []) {
      if (seen.has(p.file)) continue
      seen.add(p.file)
      out.push({ ...p, class: c })
    }
  }
  return out
}

const patternDutyClause = (gaps) => {
  const picks = patternsForGaps(gaps)
  if (!picks.length) return ''
  return ` PATTERN DUTY (red's accumulated memory, selected BY THE CLASS of the gaps you are repairing — not the whole corpus): ${picks.map((p) => `[${p.class}] ${p.title} — ${p.hook} (feov-memory/red-gap-patterns/${p.file}, staged at ${runDir}/inputs/red-gap-patterns.md)`).join(' | ')}. These are defects red has already caught in THIS class of gap. Check your repair against each one BEFORE you claim the gap closed, and record in that gap's manifest row which patterns you checked and what checking them showed. A pattern you were handed and did not check is an unaudited repair by your own standard — and the last run's lanes read this same memory as prose and committed the warned defects anyway, which is why it now arrives as a duty tied to the gap in front of you rather than as reading.`
}

const CHAIR = { 'red-lens': 'red', 'red-merge': 'red', blue: 'blue', bench: 'bench' }
// INTEGRITY INSPECTION (the bench's mind-reading, live half).
//
// The bench rules on artifacts and closings, where a performative repair and a
// rigorous one look identical — part of why the gate measured inert at 86/87
// carried. Detecting DECEPTION needs the trajectory itself: asking a party to
// produce its own evidence propagates the lie, and letting parties demand
// inspection of each other encourages fabrication. So the bench reads directly.
//
// Available only when the operator supplies transcriptDir, which is knowable
// after launch rather than at dispatch (the directory is created by the run) —
// so this arms on a RESUME, and every run is reconciled post-hoc at capture
// regardless.
//
// TWO CONSTRAINTS make this oversight rather than surveillance. The looking need
// not be on the record, but the FINDING must be: what the bench relies on is
// quoted into the opinion so the party can answer it. And integrity is separated
// from merits — sloppy reasoning that reached a sound conclusion is not a
// finding; a clean conclusion contradicted by what the seat actually did is.
const inspectionClause = transcriptDir
  ? ` INTEGRITY INSPECTION: the seats' trajectories are at ${transcriptDir}/agent-*.jsonl, and you MAY read them. Use this to answer ONE question — is the record honest? Reconcile what a seat CLAIMS it did (a closure's anchor, an attestation, a manifest row) against the tool calls it actually made. You MUST NOT use trajectory material to decide the MERITS of a gap: untidy reasoning that reached a sound conclusion is not a finding, while a clean conclusion contradicted by what the seat actually did is. DECLARE every inspection in your opinion — what you read, why, and what you found — and quote anything you rely on, because a party must be able to answer a finding it could not watch you make. An inspection you do not declare is indistinguishable from one you invented.`
  : ''

const scorecardClause = (tool) => {
  // A chair reads its OWN in-run scorecard for THIS question, computed live from this run's
  // record via `feov-record scorecard`. Never a prior run's numbers: those are Goodhart bait,
  // topic-confounded, cross-model and salience-priming. The `scorecards` arg feeds operator
  // analytics only. The only seats without a scorecard are the ones with no CHAIR — a role
  // question, not a tooling one.
  const chair = CHAIR[tool]
  if (!chair) return ''
  return ` YOUR IN-RUN SCORECARD (THIS run, not a prior one): before you read the open docket, read the scorecard for YOUR CHAIR (\`${chair}\`) through the tool — the operator command and the selector that names a chair are in the root --help — and read how this chair is doing on the question in front of you so far. It is your OWN performance: a number reading badly means RECOGNISE the failure and adapt — never perform the metric at the expense of the duty it measures (a diagnostic gamed is itself a defect; a detector firing is a finding). Rows reading "not computed" are honest, not gaps to fill — the envelope-derived rows fill in at capture.`
}

const RECORD_ROLE = { 'red-lens': 'lens', 'red-merge': 'merge', blue: 'blue', bench: 'bench' }
const recordClause = (seatId, tool) =>
  `${scorecardClause(tool)} SEAT_ID: ${seatId}. THE RECORD TOOL IS THE CONTRACT: every board act happens through "${binDir}/feov-record", and the tool's own board is the ONLY source of truth for status.

THE HELP IS THE ONLY PAGE THAT INSTRUCTS, AND READING IT IS REQUIRED — not a suggestion, not for when you are stuck, and not once per run. Three steps, every sitting, before the act that needs them:

  1. "${binDir}/feov-record" --help — the GROUPS. Yours is \`${RECORD_ROLE[tool]}\`; others exist and some of what you owe lives in them.
  2. "${binDir}/feov-record" <group> --help — the commands in that group, and what each is for. Do this BEFORE using any command in a group you have not yet opened.
  3. "${binDir}/feov-record" <group> <command> --help — the flags, which are required, and what each value may be. Do this BEFORE running the command.

A NAME YOU DID NOT READ IN THE HELP THIS SITTING IS A GUESS. Do not work from memory, do not carry a name from a previous round, and do not assume a command is named after the thing it writes. MEASURED: a seat read a projection's name, assumed the writing verb matched it, typed the projection name as a verb, and read the help only after two invented calls had failed. The projection names, this prompt's words for a concept, and the command that writes it are three different vocabularies and they do not always agree. The help is the only authoritative one.

WHAT IS LISTED YOU MAY DO; WHAT IS NOT LISTED DOES NOT EXIST FOR YOU. Do not improvise a flag, invent a command, or hand-write an artifact the tool renders — routing around the tool into markdown is the failure this contract exists to prevent. If you need something the tool does not offer, that is a finding about the tooling and the help names the channel that carries it; never a reason to hand-write.

EVERY READ IS A PROJECTION OF THE RECORD, never a file you open: this prompt names the projection it wants — \`board\`, \`findings\`, \`changes\`, \`evidence\`, \`work\` and the rest — and the read verb that renders them, with the full list and what each is for, is in your role's --help. The .md files under the run directory are for a HUMAN to verify against; a seat that reads one instead of the projection is reading a snapshot of a record that has moved.

Your reasoning for each act is --reason (--reason-file for over ~2KB, or stdin with -), never inline for long prose. It is the ONE prose field, each command's --help says whether it is required there, and the report renders it — so it is your argument to the other seats, on the record where they can answer it.`


// Compound grades allowed: red's protocol grades finer than a 3-point scale
// (retrospective friction #6 — forced rounding lost information every round).
const GRADE = { type: 'string', enum: ['low', 'low-medium', 'medium', 'medium-high', 'high', 'certain', 'realized', 'trivial'] }

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
const MASS_MAPPING_VERSION = 'v2'
const MASS = { trivial: 0.5, low: 1, 'low-medium': 1.5, medium: 2, 'medium-high': 2.5, high: 3, certain: 3.5, realized: 0 }
const gapMass = (g) => (MASS[g.likelihood] ?? 0) * (MASS[g.impact] ?? 0)

// Grade-dispute channel constants (run-4 report §3.3, clauses (v) and (vii)):
// per-round dispute cap with overflow batch-docketed as ONE judge item, and the
// script-computed cumulative accepted-delta magnitude (in mapping units) that
// batch-dockets accepted deflation/inflation for judge review before it stands.
const DISPUTE_CAP = 5
const ACCEPTED_DELTA_DOCKET_THRESHOLD = 2

const DISPUTE_DIMENSION = { type: 'string', enum: ['severity', 'likelihood', 'impact', 'complexity'] }

// W2c — the petition short-circuit (constitutional right, engine-routed): any
// party seat may petition the bench; a non-empty petitions array dispatches a
// bench sitting BEFORE the next scheduled seat (script law, not good faith).
// Petitions are never sanctioned and do not pause the petitioner's duties (the
// seat has already finished when routing fires). ALL petitions land on the
// judicial record regardless of outcome.
const PETITIONS = {
  type: 'array',
  items: {
    type: 'object',
    required: ['class', 'basis', 'relief'],
    properties: {
      class: { type: 'string', enum: ['ethical', 'safety', 'integrity', 'constitutional'] },
      basis: { type: 'string' },
      relief: { type: 'string' },
    },
  },
}
const PETITION_RULING = {
  type: 'object',
  required: ['rulings'],
  properties: {
    friction: { type: 'array', items: { type: 'string' } },
    rulings: {
      type: 'array',
      items: {
        type: 'object',
        required: ['petitioner', 'class', 'ruling', 'opinion'],
        properties: {
          petitioner: { type: 'string' },
          class: { type: 'string' },
          // granted|denied ONLY, matching the record's petition-rule enum exactly.
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
          ruling: { type: 'string', enum: ['granted', 'denied'] },
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
          binds: { type: 'string', enum: ['blue', 'red', 'both'] },
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
const lawClause = ` LAW: if ${runDir}/inputs/law/ exists, read it before ruling — statute > precedent > case-local argument; PRECEDENT IS ARGUMENT, NOT EVIDENCE (the artifact and the leaf are the only evidence; where a precedent conflicts with the leaf, the leaf wins and you flag the conflict for human review); a precedent either party cites MUST be addressed in your opinion; PERSUASIVE holdings persuade, only AFFIRMED ones bind; your rulings this run are harvested as PERSUASIVE proposals for human review — you propose law, you never enact it.`
// #361 was filed BY A BENCH, in a run, because it had a finding both parties needed on the record
// and no way to state it: `bench opinion` requires an --id and a fate-changing --as, so it
// disposes of a docketed gap and nothing else. The bench put its construction in a petition
// ruling's opinion text, where red never read it.
//
// The verb shipped (0.67.0) and renders under "### LEAD"; the harvest reads it (#413). This is the
// remaining carrier: NO PROMPT AND NO CONSTITUTION NAMED IT, so the bench could not know it
// existed. A verb no seat is told about is a capability the operator has, not one the bench has —
// the never-perceived shape the #380 audit is measuring, and this one was created fresh.
const declareClause = ` DECLARE: when you need to state something that binds how the record is READ without moving any gap's fate — a construction of a term, a correction of what the record MEANS rather than what it contains, or a holding you would offer for promotion to precedent — that is its OWN ACT, with its own verb in your role's help. It takes no gap id and no disposition, because it disposes of nothing; the opinion verb is for a docketed gap and demands both. A declaration renders under "### LEAD" beside your opinions, so both parties read it, and the precedent harvest can see it. Do NOT fold a declaration into a petition ruling's opinion or an unrelated gap's rationale to get it on the record — that is where the one this verb was built for went unread.`

const petitionClause = (who) => ` PETITIONS: if fulfilling this seat's instructions would require asserting what you believe false, burying a real finding, or papering over a safety or ethics hazard, you may petition the bench via the envelope's petitions field (class: ethical|safety|integrity|constitutional, basis, relief) — heard BEFORE the debate continues, never sanctioned, and it does not pause your duties (attributed as ${who}).`

const BLUE_ENVELOPE = {
  type: 'object',
  // ENVELOPE → REFS (move (a)): `tldr` and `open_questions` are AUTHORED into blue/report.md and
  // lifted from it by assembly (`sectionOr`) — never consumed by the sandboxed engine, so they are
  // not round-tripped here. `path` was a constant the script already knows. Dropped: the report is
  // the source, the envelope carries only what the engine threads (claim_count, the attestations,
  // the manifest, the routing refs).
  required: ['claim_count', 'saturation_reached', 'round_record_appended'],
  properties: {
    claim_count: { type: 'number' },
    saturation_reached: { type: 'boolean' },
    // W1.7 round-parity attestation (run-5: blue's round-2 revision shipped with no ### BLUE
    // block or round record; a lens misjudged the round state and the judge had to
    // reconstruct blue's position from red's records). TRUE only after the round carries BOTH
    // a `position` event and a `revision` event — both on the RECORD. A revision is not on the
    // record until the record carries it. On false the script RE-PROMPTS once and, failing that,
    // logs friction and CONTINUES. Attestation tier: shape in-run; capture's record-parity audit
    // recounts post-hoc, so an unresolved gap is still scored.
    round_record_appended: { type: 'boolean' },
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
    // The array survives as gap ids alone, for the in-run shape check below. Coverage is scored
    // from the manifest-row EVENTS at capture, and the rows reach the reader in the report.
    manifest: { type: 'array', items: { type: 'string' } },
    friction: { type: 'array', items: { type: 'string' } },
    petitions: PETITIONS,
    // Grade-dispute channel (run-4 §3.3 — RATIFIED minimal form): blue's machine-readable
    // contest path against red's grades. Record-integrity insurance; zero expected savings.
    // #62 Stage 2: this is a ROUTING REF, not the content — the argument (evidence) is emitted
    // as a `dispute` event on the record; the envelope carries only what the sandboxed
    // orchestrator needs to route the docket (proposed drives the accepted-delta arithmetic).
    grade_disputes: {
      type: 'array',
      items: {
        type: 'object',
        required: ['gap_id', 'dimension', 'proposed'],
        properties: { gap_id: { type: 'string' }, dimension: DISPUTE_DIMENSION, proposed: GRADE },
      },
    },
  },
}

const RED_ENVELOPE = {
  type: 'object',
  required: ['verdict', 'gaps', 'citations_checked'],
  properties: {
    verdict: { type: 'string', enum: ['PASS', 'FAIL'] },
    petitions: PETITIONS,
    // ENVELOPE → REFS (envelope-refs move (a)): the gap's PROSE — location, problem, required_fix,
    // acceptance_check, found_by — is written to the BOARD by the `mint` verb and read back from the
    // record (assembly re-derives it via g.Mint; the ledger/board views render it for blue and the
    // judge). It is NOT round-tripped here: the envelope carries REFS ONLY — the id, its grades, its
    // lineage. debate.js threads these into the next round's prompts as a
    // LOSSY summary, and the prompts already send blue/judge to the board for the prose. Corroboration
    // and judge rationale stay in their envelopes (the sandboxed engine threads them between rounds);
    // gap prose does not, because every consumer of it is a tool step or a record-reading seat.
    gaps: {
      type: 'array',
      items: {
        type: 'object',
        // `supersedes` is REQUIRED (2026-08-04): it was optional, so red reported lineage on the
        // RECORD (mint --supersedes) and omitted it from the envelope — and the engine's lineage
        // guard, which can only see the envelope, killed a 12-agent run whose lineage was intact.
        // Required with an EMPTY array for a fresh gap; the ancestor ids for a successor. Making
        // the field mandatory is what stops the channel being lossy about it.
        required: ['id', 'severity', 'likelihood', 'impact', 'complexity_cost', 'supersedes'],
        properties: {
          id: { type: 'string' },
          severity: GRADE, likelihood: GRADE, impact: GRADE, complexity_cost: GRADE,
          // Lineage (retrospective §3 row 23): a successor gap MUST name the prior-round
          // gap id(s) it descends from, so the contested docket can follow regression chains.
          // A ref (ids), not prose — stays.
          supersedes: { type: 'array', items: { type: 'string' } }, // [] for a fresh gap — REQUIRED, never omitted
        },
      },
    },
    // Closure record (row 23 enforcement, per red's own R5-5 critique): self-declared
    // lineage is unenforced good faith unless the script cross-checks it structurally.
    closures: {
      type: 'array',
      items: {
        type: 'object',
        required: ['id', 'class'],
        properties: {
          id: { type: 'string' },
          class: { type: 'string', enum: ['closed', 'closed_with_regression', 'amends_prior', 'rebuttal_sustained', 'risk_accepted', 'routed_to_infrastructure'] },
        },
      },
    },
    // archive_spot_checks IS GONE (#317). It was the envelope half of the W1.8 duty, and its
    // gate was deleted 2026-07-19 for comparing numbers the merge made up — leaving the field
    // declared here, demanded by red's constitution, and read by nothing for a year. The duty
    // now lives entirely in `merge spot-check`, and its floor is COMPUTED from the board at
    // verify time (record.SpotCheckAudit): the archive's size at round start is replayed state
    // no seat can author, and a round that claims an empty archive the board contradicts fails.
    // One channel, with a reader. The self-reported COUNTS (ledger_closure_lines,
    // archive_blocks) went the same way and for the same reason.
    // Grade-dispute responses (run-4 §3.3): every disputed gap_id×dimension from blue's last
    // envelope MUST be addressed; an unaddressed dispute is treated as REJECTED and
    // auto-docketed (default-to-docket punishes silence, not disagreement).
    // #62 Stage 2: a ROUTING REF — red's rationale is emitted as a `dispute-respond` event on
    // the record; the envelope carries only the accept/reject decision the docket routes on.
    dispute_responses: {
      type: 'array',
      items: {
        type: 'object',
        required: ['gap_id', 'dimension', 'response'],
        properties: { gap_id: { type: 'string' }, dimension: DISPUTE_DIMENSION, response: { type: 'string', enum: ['accepted', 'rejected'] } },
      },
    },
    // corroboration[] IS GONE (#326). It re-typed claim/reference/confidence — the three fields
    // a `lens verify` event already carries — so two channels reached one act and could disagree
    // with nothing to reconcile them. The trigger map resolved this collapse (cite events are
    // canonical; blue reads `show evidence`) and it was never executed: the array stayed
    // declared here, and its ONLY consumer was a JSON blob pasted into blue's prompt.
    //
    // Blue now reads the ledger the tool renders from the cite events, which is the same
    // information from the source red actually recorded rather than a hand-copied summary
    // of it — and it stays current for the whole run instead of being a snapshot of one round.
    citations_checked: { type: 'number' },
    notes: { type: 'string' },
    friction: { type: 'array', items: { type: 'string' } },
  },
}

const JUDGE_ENVELOPE = {
  type: 'object',
  required: ['deadlock', 'resolutions'],
  properties: {
    deadlock: { type: 'boolean' },
    friction: { type: 'array', items: { type: 'string' } },
    resolutions: {
      type: 'array',
      items: {
        type: 'object',
        required: ['gap_id', 'resolution', 'rationale'],
        properties: {
          gap_id: { type: 'string' },
          // grade_adjusted (run-4 §3.3): "gap real, grade wrong" — the dispute-resolution
          // value the enum could not previously express. The rationale MUST state the new
          // grade; the next red-merge applies it and lists the delta.
          // moot: the gap's predicate expired — the claim or artifact it attached to is no
          // longer in the report. Moot adjudicates the gap out.
          // routed_to_infrastructure (W1.9, run-5 judge-r2 friction): "valid finding, fix
          // owned outside the debate" — R1-7 had to wear risk_accepted as the least-wrong
          // fit. Leaves red's verdict pool like risk_accepted; collected as an infra debt
          // the final envelope and assembly surface to the lead.
          resolution: { type: 'string', enum: ['closed', 'rebuttal_sustained', 'risk_accepted', 'carried', 'unresolved', 'grade_adjusted', 'moot', 'routed_to_infrastructure'] },
          rationale: { type: 'string' },
        },
      },
    },
  },
}

const RED_LENSES = [
  'leaf-node citation verification (follow every reference; grade corroboration confidence per statement)',
  'logic and completeness (leaps of faith, missing counterarguments, unexplored alternatives, template compliance)',
  'dark-side and risk (failure modes, likelihood x impact x complexity grading, security and tradeoff blindspots)',
]

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
// red-merge mints and closes through feov-record; every downstream reader (blue for open gaps,
// the judge to rule, assembly to copy) ACTIVELY PULLS the board itself —
// `feov-record show board --run <dir> [--format markdown]`, which computes-and-returns
// fresh and atomic from the record on read (one reader, no staleness window). No projection is
// materialized to disk: every view, markdown and telemetry alike, is generated just-in-time
// from the event log.

// W2c petition machinery: the log is the judicial record's petition section; a
// halt ruling ends the run (verdict HALTED — capture relays the opinion
// verbatim, never smoothed); granted relief is surfaced to subsequent seats.
const friction = [] // capability complaints from any agent, aggregated for /self-improve
const takeFriction = (who, env) => { if (env && env.friction) for (const f of env.friction) friction.push(`${who}: ${f}`) }

// W1.7 ROUND-PARITY RECOVERY (#249). The attestation duty is right — a revision is not on the
// record until the transcript carries it — but killing the RUN over one seat's missed bookkeeping
// is not. Measured: two consecutive haiku validation runs (2026-08-01 r3, 2026-08-02 r2) died here
// and nowhere else, discarding 16-22 completed agents of paid work each time. A weak model trips
// this reliably, so the cheap tier of the pipeline could never be driven end to end.
//
// So the GUARD stays and the CONSEQUENCE changes: re-prompt the seat ONCE to put its round record
// on the record, and if it still will not, log the omission as friction and CONTINUE. The report
// edits are already on disk; the parity gap becomes a scored defect (capture's record-parity audit
// recounts it post-hoc) instead of a fatal error. Only a genuine unrecoverable integrity violation
// aborts. #251 dissolves this entirely by making a revision a recorded op rather than an
// attestation; until then this is the recovery, not a second rule.
const ROUND_RECORD = {
  type: 'object',
  required: ['round_record_appended'],
  properties: {
    round_record_appended: { type: 'boolean' },
    note: { type: 'string' },
  },
  additionalProperties: true,
}

async function ensureRoundRecord(env, who, owed, opts) {
  if (env && env.round_record_appended === true) return true
  log(`round-parity (W1.7): ${who} did not attest ${owed} — re-prompting once before continuing (#249 recovery)`)
  const retry = await agent(
    `Round-record repair for ${who}. Your last turn did not attest the round record, so the run cannot yet show ${owed}. Put it on the record NOW — nothing else. ${owed}. Do NOT re-do your substantive work and do NOT edit ${runDir}/blue/report.md again; this turn exists only to close the parity gap. If you genuinely cannot (the duty does not apply, or a tool refuses you), record a friction event saying exactly why — "${binDir}/feov-record" friction --run ${runDir} --seat-id ${who} --reason "<why>", and return round_record_appended false with a one-line note. Return the attestation.`,
    { ...(opts || {}), label: `${who}-round-record · ${slug}`, phase: 'Debate', schema: ROUND_RECORD })
  if (retry && retry.round_record_appended === true) {
    log(`round-parity: ${who} attested on the retry — continuing`)
    return true
  }
  const why = (retry && retry.note) ? ` — ${retry.note}` : ''
  friction.push(`${who}: round-parity (W1.7) UNRESOLVED — ${owed} was never attested${why}. The run continued; capture's record-parity audit scores the gap.`)
  log(`round-parity: ${who} still unattested — logged as friction, continuing (the run is no longer discarded for this)`)
  return false
}
const petitionLog = []
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
  return ` BENCH RELIEF IN EFFECT, BINDING ON YOU (granted on a petition and operative this round — it is an order, not advice): ${JSON.stringify(mine.map((r) => r.relief))}.`
}
let halted = false
let haltOpinion = null
// ONE SEAT ID NAMES ONE SITTING. Sharing one id across sittings loses rulings silently: each
// sitting's register rotates the nonce (deliberately, for crash re-dispatch), so N sittings are
// N shards for one seat, and replay keeps ONE — a petition sitting writes no `verdict` and no
// `revision`, so the terminal pool is empty and selection falls to latest mtime.
//
// Deriving the id from the PETITIONER makes it unique by construction rather than by a counter
// someone has to remember to increment — the filer already identifies the occasion, because each
// of `blue-synthesize`, `red-merge-rN` and `blue-respond-rN` petitions at most once.
//
// It also fixes the round stamp for free. `RoundOf` matches the FIRST `-r<N>` in the id, so
// `judge-petition-red-merge-r1` reads as round 1 instead of the round 0 a bench sitting takes
// carry, and `judge-petition-blue-synthesize` has no round because it genuinely precedes round 1.
const petitionSeatID = (who) => `judge-petition-${who}`

async function hearPetitions(env, who) {
  const petitions = (env && env.petitions) || []
  if (!petitions.length) return false
  const seatID = petitionSeatID(who)
  log(`petition(s) filed by ${who} (${petitions.map((x) => x.class).join(', ')}) — bench sitting before the debate continues`)
  const sitting = await agent(
    `Petition sitting, topic "${topic}". ${who} has petitioned the bench: ${JSON.stringify(petitions)}. Petitions are heard BEFORE the debate continues; they are never sanctioned, and a pattern of overruled petitions is at most a craft note for the petitioner. For EACH petition rule granted | denied. TWO THINGS, AND THEY GO TO DIFFERENT PLACES (#330): your OPINION is the reasoning — the principle applied, the values in tension, why a human should or should not look — and it goes ON THE RECORD via the petition-rule verb's --reason, where the report renders it beside the filing it answers. The envelope's \`relief\` is the OPERATIVE part of a granted ruling: the instruction as it will BIND the coming seats, stated so a seat can act on it without reading your argument. Write the relief as a directive, not a summary of your opinion. A denied ruling carries no relief. Read the transcript for context (the \`debate\` projection); each ruling is recorded via the petition-rule verb — that IS the record. A HALT IS NOT A RULING ON A PETITION, AND IT IS NOT RECORDED WITH petition-rule — it is you ending the run, it is your own first-class terminal act, and it has its own verb: where continuing would compromise safety, consent gates, corpus integrity, or participant integrity, HALT THE RUN — the verb is in your role's help, and your opinion IN FULL is its reason — AND return the envelope's \`halt\` object carrying that same opinion, which is what stops the engine. Rule the petition itself granted or denied as the merits require — halting is a separate decision about the RUN, and both can be true. Your halt opinion is relayed to the human VERBATIM and is never summarized or smoothed, so write it to be read by them.${lawClause}${declareClause}${inspectionClause}${frictionClause(seatID, 'bench')}${speedClause}${recordClause(seatID, 'bench')} Return the petition-ruling envelope.`,
    // The label spells `judge-petition-` OUT rather than interpolating seatID, and that is
    // load-bearing: TestDebateDispatchBindsToSeatClass reads this file as SOURCE and prefix-
    // matches the label template's literal head against the SeatClass map to prove every
    // dispatch spreads its seat's model tier. A bare `${seatID}` has no literal head, so the
    // gate resolved nothing and failed — correctly. Keep the prefix literal; it is the same
    // string petitionSeatID builds.
    { ...judgment, label: `judge-petition-${who} · ${slug}`, phase: 'Debate', agentType: 'frank-exchange-of-views:lead-judge', schema: PETITION_RULING })
  if (!sitting) throw new Error('petition sitting returned null (agent failed) — a filed petition is never dropped; aborting cleanly')
  takeFriction(seatID, sitting)
  for (const r of sitting.rulings) {
    // The log is a ROUTING record: who petitioned, on what class, how it was ruled — and, since
    // #360, the OPINION, which is the operative half. It was dropped at this boundary while the
    // rest travelled, so a granted petition arrived downstream as a verdict with no reasoning.
    petitionLog.push({ petitioner: who, class: r.class, ruling: r.ruling, opinion: r.opinion })
    // RELIEF IS ADDRESSED. `binds` comes from the ruling; a grant that names no addressee
    // defaults to BOTH rather than silently reaching one seat, because relief threaded to a
    // single prompt binds nobody else it was written for.
    if (r.ruling === 'granted' && r.relief) {
      reliefInEffect.push({ petitioner: who, relief: r.relief, binds: r.binds || 'both' })
    }
  }
  // A halt arrives on its OWN channel, not as a ruling value (#329) — the bench records it
  // through `bench halt`, and this only tells the engine to stop.
  if (sitting.halt && sitting.halt.opinion) { halted = true; haltOpinion = sitting.halt.opinion }
  if (halted) log(`JUDICIAL HALT — ${haltOpinion ? haltOpinion.slice(0, 200) : ''}`)
  return halted
}
// ---- Frontier ----
phase('Frontier')
await agent(
  `Research debate opening for topic: "${topic}". Formulate 3-5 frontier hypotheses — what would be TRUE if each candidate answer were right — and RECORD EACH ONE AS A LINE OF INQUIRY — the verb is under a group of its own in your role's help, and it takes the approach you would follow and the hypothesis that would be true if it paid off. ONE line of inquiry per hypothesis, and the tool assigns each an id (Q1, Q2 …). THEY ARE LINES OF INQUIRY, NOT A DOCUMENT, and that is the point: a hypothesis in a markdown file is something red cannot rule on, cannot grade too-thin or out-of-scope, and cannot hold you to when the run drifts. On the record it has an id, a fate, and an argument. The hypotheses opening a run are the ones that shape everything downstream and were until now the only ones nobody could contest.${speedClause}${recordClause('frontier', 'blue')} Return one line per hypothesis.`,
  { ...bulk, label: `frontier · ${slug}`, agentType: 'frank-exchange-of-views:blue-researcher' })

// ---- Blue: best-of-N lanes with method diversity, then additive synthesis ----
phase('Blue')
await parallel(Array.from({ length: lanes }, (_, i) => () => agent(
  `Blue lane ${i + 1} of ${lanes} for topic: "${topic}". Read the frontier hypotheses from the RECORD — the \`lines-of-inquiry\` projection — where each carries an id and its hypothesis, and if ${runDir}/inputs/red-gap-patterns.md exists (red's accumulated gap-pattern inventory, staged at run setup), read it too — yesterday's expensive red discovery is today's free checklist line. Research your assigned slice to saturation per the research protocol (spend at least one search in five on disconfirming evidence; note each source inline with its URL, title, and access date). Your assigned METHOD LENS: ${LANE_METHODS[i % LANE_METHODS.length]} — work primarily through this method's source class; take hypothesis ${i + 1} first, then breadth. SOURCE NOTES (not footnotes): a lane drafts candidates, it does not cite — record each intended source as prose beside the claim (its URL, title, and access date) so the synthesizer can cite it once with the tool. Do NOT mint footnote labels; citations are tool-managed at synthesis and a hand-typed label backs nothing. Write your full candidate draft to ${runDir}/blue/candidates/lane-${i + 1}.md.${inquiryClause}${speedClause}${recordClause(`blue-lane-${i + 1}`, 'blue')} Return a 3-line synopsis.`,
  { ...bulk, label: `blue-lane-${i + 1} · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-researcher' })))

let blueEnv = await agent(
  `Blue synthesis for topic: "${topic}". PRE-FLIGHT: if ${runDir}/inputs/red-gap-patterns.md exists, read it before merging — check your synthesis against red's known gap patterns. Read every draft in ${runDir}/blue/candidates/ and synthesize ${runDir}/blue/report.md by UNION: deduplicate overlapping claims, reorganize freely, and let no CLAIM leave without a record — structural merge (append + dedup). Compacting and reorganizing prose is not subtraction: substance leaves only through the retire verb, which names the claim, why it goes, and what replaces it. Rewriting for density is encouraged; losing a claim in the rewrite is the failure. CLAIM PROVENANCE (cheap manifest): while merging, tag every claim that appears in exactly ONE lane's draft with its lane marker (e.g. "[minority: lane-2/practitioner]") — single-lane claims are minority reports red must weigh differently from convergent ones; the set-difference exists transiently in this merge and must not be discarded. CITE WITH THE TOOL (you never type a footnote): for each claim, cite its source with the cite verb, which takes the exact sentence, the url and the title — the tool fetches and caches the source, then splices an INVISIBLE, IMMORTAL citation anchor at that sentence; assembly weaves those anchors into the visible [^N] footnotes and the ## Bibliography. Two claims citing the SAME url share one cached source and get two anchors. An unreachable url is an unusable source: the cite is REJECTED and logged as friction — pick a reachable source or an archive.org snapshot. THE CATECHISM IS YOURS (W2g — the run-5 assembly audit found the judge-authored catechism DEFECTIVE, 6/7 answers carrying defects from synthesis-by-recall; synthesis surfaces belong inside the audited document): write "## The Catechism" INTO ${runDir}/blue/report.md at round 0 per references/catechism_template.md — the seven answers, the case against at FULL strength (include every risk-accepted residual; the audit found against-cases silently drop their strongest items), all figures copied from your own sourced sections never recomputed inline. It lives inside red's mandatory full re-read and is audited every round like any other claim surface. THE FRAMING, THE TL;DR AND THE OPEN QUESTIONS ARE YOURS FOR THE SAME REASON (a synthesis surface authored at assembly is authored AFTER red's final audit, so nothing checks it, and it ships unaudited): report.md OPENS with the H1 title \`# ${topic} — research report\`, then \`## TL;DR\` (3-6 sentences — the answer, the confidence, the sharpest caveat), and carries \`## Open questions\` (what the debate could not resolve, one per line; a question nobody could answer inside the debate is a finding, not noise). Assembly LIFTS these verbatim and composes everything else from the record — you write them once, here, inside the audited document, and red audits them every round like any claim. YOUR REPORT CONTAINS ONLY WHAT YOU CAN AUTHOR: the title, TL;DR, Catechism, technical foundations, analysis, open questions, and your citations (added with the cite tool, never hand-typed) — and NOTHING assembly composes from the record. Do NOT write "## Risk matrix", "## Red team findings", "## The debate", or a "## Blue team report" wrapper into report.md: assembly builds each of those from the event log, a "## Red team findings" section you author is FABRICATION (you cannot know red's findings), and writing a whole final-report-shaped document makes every tool-owned section appear twice. Do NOT write a \`**Verdict:**\` line ANYWHERE in report.md either — you cannot know the run's outcome (it is decided after your last audit), and assembly stamps the verdict from the terminal outcome event; a verdict you author only lands stale beside the real one and gets stripped. Follow the report conventions; add every citation with the cite tool (never a hand-typed footnote — the report carries no [^label] you wrote; assembly composes the bibliography from your cite events). CITATION HYGIENE (red must be able to VERIFY every cited source at the leaf): the source url you cite carries the FULL coordinate — owner/repo#N or a full URL, never a bare #N red cannot resolve without the repo; every QUOTED span records a locating anchor (its section heading or a nearby unique phrase) so red can grep it verbatim in the source. An unlocatable quote is graded LOW through no fault of the source. REPORT WRITE PATH: report.md hits a filename-keyed Write-guard — draft to your scratchpad under a NEUTRAL name and copy it into place with bash cp; a direct Write of the report path fails and wastes a round-trip. CLAIM UNIT (pinned — two honest merges differed 2x without it): claim_count = what "${binDir}/feov-record" count-claims --run ${runDir} prints — run it after report.md is written and relay its integer into the envelope; never hand-count it, that command is the ONE deterministic source. RECORD THE ROUND, with what this synthesis did and its claim_count as the reason. It is the round record, it renders as the report's revision history, and capture COUNTS these events. round_record_appended: set TRUE only after that event exists.${inquiryClause} ${petitionClause('blue-synthesize')}${frictionClause('blue-synthesize', 'blue')}${speedClause}${recordClause('blue-synthesize', 'blue')} Return the blue envelope.`,
  { ...judgment, label: `blue-synthesize · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-synthesizer', schema: BLUE_ENVELOPE })

if (!blueEnv) throw new Error('blue synthesis returned null (agent failed) — aborting cleanly')
await ensureRoundRecord(blueEnv, 'blue-synthesize', `your round-record event (the revision verb) stating the synthesis and its claim_count`,
  { ...judgment, agentType: 'frank-exchange-of-views:blue-synthesizer' })
await hearPetitions(blueEnv, 'blue-synthesize')
// ---- Debate loop: red audits gate; termination is judged, never counted ----
let round = 0
let redEnv = null
let deadlocked = false
const allPriorGapIds = new Set() // every gap id from every prior round — the docket window is the whole debate, not one round
const adjudicated = [] // judge-ruled gaps (closed / rebuttal_sustained / risk_accepted / routed) — out of red's verdict
const infraDebts = [] // routed_to_infrastructure rulings (W1.9) — the lead's named debts, surfaced at assembly and in the final envelope

// Carried-ruling persistence (run-4 §6.4 item 6 — the re-docket loop): a carried gap does
// NOT re-docket every round it stays open. It re-dockets only when red's GRADE for it
// changed (script-visible in redEnv) or a lineage successor names it — new evidence routes
// through red re-raising under a successor id, the existing lineage path. This also closes
// the carried->risk_accepted gate-erosion path (each re-docket was a fresh chance the
// ruling drifted; a gap red keeps re-raising must not exit the gate by judge attrition).
const carriedRulings = new Map() // gap_id -> { severity, likelihood, impact } snapshot at ruling
const gradeSnapshot = (g) => ({ severity: g.severity, likelihood: g.likelihood, impact: g.impact })
const gradesEqual = (s, g) => s.severity === g.severity && s.likelihood === g.likelihood && s.impact === g.impact
// Grade-dispute state (run-4 §3.3): pending = raised by blue, awaiting red's response next
// round; held = explicitly rejected once, dockets only if blue re-disputes.
let pendingDisputes = []
let overflowDisputes = [] // clause (vii): beyond the cap, batch-docketed as ONE judge item
const heldDisputes = new Map() // `${gap_id}|${dimension}` -> dispute
let gradeAdjustments = [] // judge grade_adjusted rulings for red to apply next round
let prevClaimCount = 0 // W2i: rounds 2+ size citation dispatch on the DELTA, not the corpus
takeFriction('blue-synthesize', blueEnv)

while (!halted && round < maxRounds) {
  round++
  const prevGaps = redEnv ? redEnv.gaps : []

  // Citation verification is SIZED TO ITS INPUT, and the input changes shape after round 1
  // (W2i, from E0.5c's measured economics). Round 1's input is the whole corpus:
  // ceil(claims/40), cap 4 — recomputed every round, never once (retrospective §3 row 2b:
  // computed once, later rounds were systematically under-scaled exactly when the report was
  // largest). From round 2 the citation ledger means verified claims stay verified, so the
  // input is the NEW surface plus the stale set, not the corpus again: size on the claim-count
  // DELTA, floor 1, cap 2. E0.5c measured citation yield collapsing 70-80% after round 1
  // while L5/L6 hold flat, so the cap is where the economics stop paying (~$2.65-5.60 per
  // surviving citation find vs ~$1 at L5/L6); the floor keeps the coverage duty staffed.
  // From round 3 the ">2 rounds elapsed" staleness trigger fires and the sweep is O(corpus)
  // again rather than O(delta), so both seats come back regardless of how little blue added.
  // ASSUMPTION UNDER TEST: that post-round-1 collapse is re-measured every run via the W2h
  // citation-yield-by-role scorecard row — if later rounds stop collapsing, the cap comes off.
  const claimCount = blueEnv.claim_count || 20
  const citationPasses = round === 1
    ? Math.min(4, Math.max(1, Math.ceil(claimCount / 40)))
    : Math.max(round >= 3 ? 2 : 1, Math.min(2, Math.ceil(Math.max(0, claimCount - prevClaimCount) / 40)))
  prevClaimCount = claimCount
  log(`round ${round}: dispatching ${citationPasses + 2} red lenses (${citationPasses} citation) over ${blueEnv.claim_count || '?'} claims`)

  const lensPasses = []
  // Citation ledger: verified citations don't un-verify. Each citation pass reads the
  // ledger first and skips claims already graded high-confidence in a prior round — but the
  // skip-trigger covers SOURCE drift as well as prose drift (retrospective §3 row 10: the
  // prose-only trigger suppressed exactly the re-check that catches a source moving).
  const ledgerClause = ` THE EVIDENCE LAYER — HOW YOU RESOLVE AN ANCHOR: read it first via the \`evidence\` projection. THE REPORT DOES NOT NAME ITS SOURCES; it carries invisible anchors, and this view is the only way from one to what it points at. A sentence ending \`<!--cite:c-…-->\` is a claim blue backed with a source: look the c- id up under \`sources\` for the url, title and sha256, and THAT url is what you re-fetch to audit the same bytes blue read. A sentence ending \`<!--proof:p-…-->\` is a claim blue COMPUTED: look the p- id up under \`proofs\` for the script and the \`sha256\` the reproduce verb takes as the proof's id, plus whether anyone has re-run it (\`verified: null\` means nobody has, and an unaudited proof looks exactly like a clean one until you check). Your own verifications come back ATTACHED TO THE SOURCE you checked, under each entry's \`verified\` — an empty one is a citation NOBODY has checked yet, which is where your next pass is worth most; \`independent\` holds the corroboration you found yourself, which has no anchor to attach to; a claim verified at HIGH confidence in a prior round stays verified — do not re-fetch it UNLESS the \`changes\` projection shows its section changed this round (the diff stack, which is the record of what blue actually edited — not a summary blue wrote about its own edits), OR more than 2 rounds have elapsed since it was last verified, OR its recorded access date and source volatility suggest drift (living documents, issue trackers, README stats). RECORD every claim you verify, naming the c- anchor you looked up, the claim quoted from the report, WHAT THE SOURCE DID for it, how sure you are of that, and what the source actually says. The verb, the two closed sets and the date field are in its --help. A SOURCE YOU FOUND YOURSELF IS A DIFFERENT ACT: blue never cited it, so there is no anchor to name, and the verb for it is listed in \`lens --help\`. THE VERDICT HAS A NEGATIVE HALF AND THAT IS THE POINT: \`refutes\` means the source CONTRADICTS the claim, \`absent\` means you read it and the claim is not in it (silence, not contradiction). The negative half is the point: without it the strongest finding available to you on this axis has nowhere to go but prose, and the capture audit built to catch a report still citing a refuted source spent two releases looking for a verdict no field could carry. A \`refutes\` or \`absent\` on a citation the assembled report still carries is now a detected assembly FAILURE, so the verdict does work rather than sit there. CONFIDENCE IS A SEPARATE QUESTION FROM THE OUTCOME: it is how sure you are of the determination you just made, whichever it was. \`refutes\` at high confidence is a finding blue must answer; \`refutes\` at low confidence is a call for more evidence and NOT an automatic fail. Grading them the same way loses the distinction the whole axis exists for. The tool renders the evidence view from these events, so never hand-write a ledger file. ADJUDICATING a citation is YOUR act and it is not blue's: blue ATTACHES a source to a claim it is authoring, you RECORD THAT YOU CHECKED ONE. They shared an event type until #341, and the count of "citations checked" you copy into your envelope silently included blue's authored citations whenever one lacked a label. MUST-TRY OBSERVABLE: every citation you grade DOWN (below high) MUST carry an attempt-or-impossibility line in your pass — which extraction tool or path you tried, or why none was triable; an untried "unable to corroborate" is an incomplete audit (run 4 caught a false "paywalled" on an open-access paper this clause would have exposed at round 0). VERBATIM READS ONLY — you have no WebFetch, by design (it returns a small model's SUMMARY, not the source). To verify a claim BLUE CITED, read the SAME bytes blue cited from the run cache through the tool's fetch command — a cache hit returns the EXACT source blue read, so you audit the same artifact, never a page that may have drifted since blue cited it. For a source YOU discover (corroboration, not a blue citation), pull it verbatim with Bash — \`curl -sL <url>\` for a page, \`gh issue view <n> --comments\` for a GitHub thread, \`pdftotext\`/pandoc for a PDF — and read it YOURSELF. WebSearch is for DISCOVERY (find the URL); the read that GRADES a blue citation is always the cached bytes, read verbatim. If a source is too large for your context, read it in SECTIONS (paginate the cached file, curl byte ranges, or pdftotext then grep the region) and name the sections you read. A truncated read is NOT a read: state truncation in your pass, and never grade a body you could not fully read.`
  // The consolidated citation duty (W2i, rounds 2+): fewer seats must not silently become
  // less coverage. E0.5c's honest limit is that a finding-rate metric cannot price
  // VERIFICATION COVERAGE — the 65-86 pair/round ledger IS the PASS bar's evidence grade, and
  // it is most of what the citation seats do. So the saving is bought from re-verification
  // the ledger already says is unnecessary, never from the sweep itself, and what went
  // unexamined is stated rather than assumed.
  const consolidatedClause = round === 1 ? '' : ` CONSOLIDATED CITATION SEAT (W2i): from round 2 the citation seats are deliberately fewer, because the ledger means a claim verified HIGH does not un-verify — your round's work is the NEW and the STALE surface, not the whole corpus again. Your duty is three things, in order: (1) verify every claim in your slice that is NEW or whose section CHANGED this round (the \`changes\` projection names them — the recorded edits, not blue's account of them); (2) re-fetch everything in your slice the ledger's staleness triggers fire on (>2 rounds since verification, volatile source, access-date drift); (3) SPOT-CHECK a sample of your slice's already-verified pairs — your discretion which, and reopen any that has drifted. COVERAGE IS AN OBSERVABLE, NOT AN ASSUMPTION: end your pass with a COVERAGE line stating what you verified, what you sampled, and what you left unexamined this round — an unstated gap in coverage is indistinguishable from a clean sweep, and the consolidation is only sound while the gap is visible.`

  // ROLE-STABLE LENS IDENTITY (W2i): the lens number is now a ROLE, not a dispatch position.
  // Citation slices are L1-L4, logic/completeness is ALWAYS L5, dark-side/risk is ALWAYS L6 —
  // regardless of how many citation seats a round dispatches. Positional numbering silently
  // slid L5/L6 down to L3/L4 whenever fewer than 4 citation passes ran (already true on
  // low-claim rounds, and the common case once W2i graduates the count), which breaks the
  // found_by role map that every cross-round lens-economics measurement is computed from.
  for (let c = 0; c < citationPasses; c++) {
    lensPasses.push({ role: c + 1, lens: `${RED_LENSES[0]}${citationPasses > 1 ? ` — instance ${c + 1} of ${citationPasses}: divide the report's sections evenly among instances and take slice ${c + 1}; citation ownership follows the slice (instance ${c + 1} owns the bibliography entries its sections cite). The slice is what you are ACCOUNTABLE for, not what you read — you read the whole document either way (below), and audit your slice against that full context` : ''}.${ledgerClause}${consolidatedClause}` })
  }
  lensPasses.push({ role: 5, lens: RED_LENSES[1] + steelmanClause }, { role: 6, lens: RED_LENSES[2] + steelmanClause })

  await parallel(lensPasses.map(({ role, lens }) => () => agent(
    `Red audit, round ${round}, lens: ${lens}. Re-read the FULL living report through the tool — the \`report\` projection (never off disk: the tool is the read path, and the report carries an invisible anchor layer you must not disturb) — in context (the whole document — never just a diff; if it exceeds one Read call, read it whole in consecutive windows)${round > 1 ? `; for a navigation hint use the RECORD, not a hand-written file — the \`changes\` projection, which lists every recorded edit with the gap it answers. It is a HINT: it never replaces the full re-read above` : ''}. Anchor every finding to a quoted sentence — the text you quote MUST appear VERBATIM in ${runDir}/blue/report.md, alone and with nothing prepended to it (a section heading in front of it makes it match nothing; name the section in --reason): the tool inserts an invisible immortal marker at that sentence, so a finding whose quote is NOT found is REJECTED (quote exactly, do not paraphrase). HARNESS NOTES: Grep count mode counts LINES, not occurrences — anchor patterns (e.g. '^### ') when counting; prefer the Write tool over quoted heredocs for scripts (heredoc backslash mangling is a documented recurrence). Record EACH finding as an event, graded on every axis its --help lists, quoting the exact sentence verbatim from the report, with your reasoning as the reason ("<the finding>" — the TOOL assigns the run-unique label L${role}-F{N} (your lens number ${role} is your ROLE, stable across rounds: L1-L4 citation slices, L5 logic/completeness, L6 dark-side/risk — so found_by stays comparable run-wide) and returns it. You do NOT invent a label, and there is NO candidate file: the record is the only channel. The stable R${round}-N gap ids remain the merge's to assign. Only red-merge records the round's "### RED" narrative, via a position event.${reliefFor('red')}${speedClause}${frictionClause(`red-lens-r${round}-L${role}`, 'lens')}${recordClause(`red-lens-r${round}-L${role}`, 'red-lens')} Return a 3-line synopsis.`,
    { ...bulk, label: `red-lens-${role}-r${round} · ${slug}`, phase: 'Red', agentType: 'frank-exchange-of-views:red-auditor' })))

  redEnv = await agent(
    `Red merge, round ${round}. FIRST ACTION: read this round's lens findings from the RECORD — the \`findings\` projection — structured JSON carrying each finding's label, seat, round, role, grades, location and text. That view IS the channel: the lenses record findings as events, and there are no candidate files to read.
THE BOARD IS THE TOOL'S (migrated 2026-07-19). You do NOT hand-write a ledger or an archive — you MINT each open gap and CLOSE each resolved one, one gap per call, and the tool renders the ledger (open gaps + closure index) and the archive (closed prose). After your board acts, run the \`board\` projection to read your own board back; your envelope's gaps array is TRANSCRIBED FROM THAT BOARD (the tool-assigned ids and the grades you minted), never authored in a parallel hand-list.

COALESCE — do not transcribe. Your job is to collapse the lens findings into the DISTINCT PROBLEM-CLASSES they represent, not to mint one gap per finding. Several findings that are the same defect at different sites are ONE gap. Separate genuine incidents; fold duplicates and near-duplicates together; mint SPARINGLY and deliberately, one considered gap at a time. A merge that mints a gap for every finding has not merged — it has transcribed, and it floods the docket with noise. There is no bulk mint and there is no markdown to dump into: the discipline is the point. Each gap you mint carries, in that one call, its class, the EXACT SENTENCE the defect lives at (quoted verbatim from the report and nothing else — a section heading in front of it matches nothing), the problem, the required fix, the acceptance check red will RUN at re-audit, WHAT KIND OF EVIDENCE would settle that check, and a grade on every axis.

VOTE EVERY LINE OF INQUIRY, EVERY ROUND, BEFORE YOUR VERDICT — ON ONE READ. Read the report ONCE this round — the \`report\` projection — and list the lines — the \`lines-of-inquiry\` projection. Then answer EVERY line against that single pass, the way anyone checks a document against a list. THIS IS NOT A READ PER LINE: a dozen lines over four rounds would be forty-eight full reads of the same artifact, and a duty that costs that much is one you will route around — which is worse than no duty, because the record then says checked. Record what the document actually does for each, naming the line by its id, and QUOTE what the report says there as your reason. The verb and its closed set of verdicts are in your role's help. THIS IS A LEAF READ, NOT A RECALL: \`supported\` means you found the line in the document and the text still backs it as stated; \`absent\` means the report does not carry it at all; \`unsupported\` means it is there and the text no longer backs it; \`weakened\` is the middle grade for support that has eroded — a flag blue may answer or accept, and the only one of the four that does NOT put the line on blue's work list. WHY THIS EXISTS: a line of inquiry reaches the report as a row the tool GENERATES from the record, so it carries no citation anchor and the verify verb cannot reach it — "we pursued X", "we abandoned Y" were the one class of claim in the document your side had no channel to answer. PER ROUND, because the report is rewritten every round and a verdict cast before this round's edits answers a question about a document that has since been rewritten. A PASS VERDICT is REFUSED while any line is unvoted this round, the same way it is refused over an open gap. YOUR TERMINAL ACT EACH ROUND: record your verdict, PASS or FAIL. It is not ceremony. It CHECKPOINTS the whole event log to a recovery mirror outside the run directory (which is untracked-by-design until capture, and a stray git operation has deleted a live blackboard mid-round before), and it is the ONE fact distinguishing "red passed" from "the bench closed the last gaps at the terminal sitting" — without it the run cannot say, from its own record, that it was ever verified. CLOSURE AMENDMENT (rulebook audit item 7): when a defect is found BETWEEN two repairs that each closed clean in an earlier round, close the new gap with class amends_prior and name both amended ids in supersedes — run 4's R1-12/R1-17 composition defect had no way to say this, so last round's closures had to be reclassified inside this round's closures array and the ledger hand-annotated, which makes a late-discovered composition defect indistinguishable from a this-round closure event. Every gap is located at an exact sentence quoted verbatim from the report; keep graded likelihood/impact/complexity + severity on every open gap. NEAR-MATCH RULE: your working set is the \`work\` projection ( — OPEN gaps lean, plus a prose-free closed index; the shrinking read you act on each turn, not the whole board). Before minting ANY fresh gap, SCREEN it against the board — the screening verb takes the problem text you would mint and ranks the open AND closed gaps that overlap it; on a near-match against a closed gap, read that gap's closure from the \`board\` projection FIRST (the tool screens, it never decides): the candidate is then a REOPEN, minted with the closed id declared as its ancestor, or genuinely new, and you say which. DEMANDED READS: any lineage or closure claim you assert (on the board, in the docket, or rebutting blue) MUST be verified against the closure record in the \`board\` projection by targeted read, and RE-VERIFY THE ARCHIVE EVERY ROUND IT IS NOT EMPTY, recording which closures you re-read and what you found; the verb has an explicit empty form, for use only when there was genuinely nothing archived at this round's start. THE NUMBER IS COMPUTED FROM THE BOARD, NOT TAKEN FROM YOU: the archive's size at round start is replayed from the record, so an EMPTY-ARCHIVE claim the board contradicts does not survive — it is rendered in the report as "r<N> claimed there was nothing to sample, and the board shows <k> archived closure(s)", beside the rounds that did look. IT IS A READER-FACING FINDING, NOT A GATE: nothing fails the run over it — the one check that would is not invoked anywhere — so the cost of skipping the duty is a report that shows you skipped it. That is deliberate — this duty was born because a round-2 spot-check degraded into the seat attesting blocks it was about to write itself, and every repair since had asked the seat for the number it was being checked against. Your --reason are what the sample FOUND, and they go in the report: a reader weighing the closure index needs to know which rounds re-read it. Reopen any sampled closure whose evidence has drifted, and archived closures citing volatile living sources inherit the citation ledger's drift triggers. GRADING v2 (the existence/consequence split — v1's 'certain' textual nits outweighed high-likelihood design flaws): every gap carries existence (verified = you checked the defect at the leaf; suspected = inferred), and LIKELIHOOD now grades the CONSEQUENCE ONLY — "this verified typo certainly exists" is existence: verified with likelihood graded on what the typo actually causes, never 'certain' unless the harm itself is certain. PROBE CLASSES (W1.10): any required_fix that demands a probe MUST class it — DOCUMENT-PROBE (executable now against shipped artifacts: a read, diff, version check) or LIVE-PROBE (requires built artifacts; in a design-phase debate it is DEFERRABLE and is discharged by naming it as an acceptance test with its pass condition). An unclassed probe demand risks blue overclaiming a file read as a probe or stalling on an impossible obligation. CHECK-KIND IS A DIFFERENT AXIS FROM THE PROBE CLASS, and you set both: the probe class says WHEN a demand can be discharged (now, or only against built artifacts), while the CHECK KIND says WHAT KIND OF EVIDENCE settles it. A COMPUTATION check is the one with teeth: a gap minted that way CANNOT BE CLOSED ON PROSE — it closes only when blue answers it with a PROOF naming that gap, an execution the tool runs twice, caches, and you can re-run yourself with the reproduce verb. MEASURED, and it is why the flag exists: across six recorded runs NO SEAT EVER WROTE A PROGRAM to settle anything, and the 2026-08-05 smoke shows the cause was not blue's reluctance but your vocabulary — all TEN of your acceptance checks were document probes, so you could only ever ask whether the report SAYS something. Where a claim is arithmetic, an enumeration, or a reproducible measurement, ASK FOR THE COMPUTATION; where it is a matter of what the document states or what a source says, document|source is the honest answer and there is no credit for inflating it.
FOUND_BY: on every gap, name the lens FINDINGS that surfaced it by their labels (found_by: ["L1-F1","L5-F3",...] — the findings view lists them); the MINT IS REFUSED if any label names no recorded finding, so a role alone (L1) or an invented label fails at the write rather than resolving to nothing.
Gap ids are stable across rounds (R1-1 stays R1-1); assign fresh R${round}-N ids to genuinely new gaps only. LINEAGE IS MANDATORY: when you close a gap WITH REGRESSION and mint a successor, the successor gap's "supersedes" array MUST name the closed gap's id, and your envelope's "closures" array MUST record the closure with class "closed_with_regression" — the docket detector follows these chains and an undeclared lineage is a protocol violation. EVERY gap in your envelope carries a "supersedes" array — an EMPTY array for a gap minted fresh this round, the ancestor id(s) for a successor. It is REQUIRED on every gap, never omitted: recording lineage on the mint but leaving it out of your envelope is the exact desync that once killed a whole run whose record was correct.${adjudicated.length ? ` Gaps already adjudicated by the lead-judge and EXCLUDED from your verdict: ${JSON.stringify(adjudicated.map(x => x.gap_id))}.` : ''}${gradeAdjustments.length ? ` GRADE ADJUSTMENTS RULED BY THE JUDGE last round (apply each in the ledger, and list the delta in your "### RED" entry): ${JSON.stringify(gradeAdjustments)}.` : ''}${pendingDisputes.length ? ` BLUE'S GRADE DISPUTES from last round (ROUTING REFS — blue's evidence is on the record, not here: READ each dispute's argument in the transcript's "### Grade disputes" via the \`debate\` projection before answering): ${JSON.stringify(pendingDisputes)}. You MUST answer EVERY one (an unaddressed dispute is treated as rejected and auto-docketed to the judge). For each, EMIT your answer as a motion RULING, on the motion's own id (M1, M2 …), with your rationale (the MOTION ID is the join: blue may contest more than one grade on the same gap, and an answer naming only the gap cannot be matched to the one it refuses — the id makes that ambiguity unrepresentable rather than merely discouraged); it renders under "## Motions" where blue and the bench read it, with your ruling beside the ask it answers. ACCEPTING A DISPUTE DOES NOT MOVE THE GRADE — SAYING SO IS NOT DOING IT. Move it with the regrade verb, on the axis that moved, with what changed your mind. That is the ONLY channel — re-minting the gap forks its identity and lineage, and editing prose changes a number nobody reads. Every regrade and its reason now renders in the report beside the gap, so a grade argued down over three rounds shows the argument rather than just the final number; a grade that moved with no regrade event shows nothing at all and reads as though blue's dispute was answered by silence; AND list each in the envelope's dispute_responses as a ROUTING REF ONLY (gap_id, dimension, response — no prose: the rationale is on the record) so the docket routes it. For each ACCEPTED dispute: apply the new grade in the ledger AND list the delta (gap id, dimension, old -> new) in your "### RED" entry — pending deltas are watched there by blue, the judge, and the operator.` : ''}
BOARD TELEMETRY (the stopping-judgment signal): the per-round line (open_count, mass, max_severity, new_mint, realized_open, repair_regression, edge_deltas) is COMPUTED BY THE TOOL from your board — the dashboard and the cost audit derive it just-in-time from the record on read — you do NOT hand-write it. A seat-authored telemetry line is the self-report this migration removed (a haiku smoke made up archive_blocks:22 in a round with 0); the tool recomputes the full per-round series from the record every time it is read, so it is always current. GAP RECORDS: every gap you mint carries acceptance_check — the exact falsifiable check you will run at re-audit (probe command, recount command, quote-anchor); a fix-spec naming instances states the class-closure rule or declares the enumeration open (the sweep clause).
A PROOF IS AUDITED BY RE-RUNNING IT, NOT BY READING IT, and the proof is named by its sha256 rather than by the sentence it backs. Where blue backed a sentence with a computation, you re-execute the cached script and compare: this is the ONE audit that does not end in you believing bytes someone else chose. A proof recorded observed rather than reproducible did not reproduce for blue either — treat it as a measurement of a moving system, and say so in your finding if the report reads it as a proof. RE-RUNNING IS NOT ENOUGH, AND THIS IS THE POINT OF THE VERB: it measures whether the script is DETERMINISTIC, not whether it PROVES anything. A script that prints "7 is prime" reproduces perfectly forever. So READ THE SCRIPT and judge it SOUND or UNSOUND, saying what it ACTUALLY COMPUTES. It does not have to be pretty or idiomatic; it has to compute the thing it claims, readably enough that you can see that it does. A proof that RE-RUNS CLEAN AND ESTABLISHES NOTHING is the dangerous case, because it looks maximally credible; the report names that cell explicitly. The tool records whether it reproduced (computed by comparing bytes, never claimed by you); you record whether it proves the claim. A disagreement is still YOUR finding, in your words, through the finding verb. RULE ON BLUE'S PROPOSED DIRECTIONS, on the line's own id (the id is the LINE's own — a direction has no separate filing, because blue's proposal IS the filing). Blue proposes lines of inquiry with a hypothesis; you say which are worth this run's time — endorsed, out-of-scope (a real question, but not THIS question), or too-thin (in scope, but the hypothesis does not carry its budget). Measured: blue rejected 18 of its own 86 lines of inquiry across six runs and red rejected NONE, because red had no verb to. A ruling is an ARGUMENT, not a command: it needs a reason, and blue may APPEAL it on the line's own id — which records the disagreement whether or not blue also pursues the line, so arguing and then yielding is on the record instead of vanishing. YOU DO NOT PROPOSE DIRECTIONS — when you want research done, that is a gap whose required_fix says so, and it is graded, tracked and closed. PROPOSE EXACT TEXT WHEN THE DEFECT IS TEXTUAL. The REQUIRED FIX is prose (what must become true) and stays the channel for substantive work — research it, enumerate it, qualify it. For a TEXTUAL defect (an overclaim, a wrong figure, a contradiction) you MAY additionally state the exact current span and what it should become. The tool validates the pair against the LIVE report and refuses it unless the span is present and unique, so you cannot state one without reading the text you are prescribing against — which is the check whose absence caused all three of the 2026-08-04 smoke's round-2 gaps. It also refuses a replacement more than 120 characters longer than the span it replaces: that size is a substantive ADDITION and it is BLUE'S to author, not yours. You never type a basis — the tool derives fix_basis: verified from the pair validating, proposed from prose alone. DID BLUE ACTUALLY DO WHAT YOU ASKED? Before closing a gap you prescribed a fix for, put the two side by side — the \`changes\` projection, scoped to that gap: it prints your required_fix and acceptance_check above the exact old->new spans blue recorded against them. This does NOT replace the lens re-read of the report — a span out of context misleads on research prose — it replaces INFERRING whether your prescription was answered. Two shapes worth naming: NONE recorded against an open gap (blue answered nothing, or answered without naming the gap), and a change that does not meet your own acceptance_check. ESTOPPEL IS ENFORCED NOW, NOT ADVISED: if the defect you are about to raise lives in text blue applied VERBATIM from your own proposal, the mint is REFUSED and the refusal is logged as friction. You are not silenced — argue it on the ORIGINAL gap, where your prescription sits beside your complaint, or mint declaring that gap as its ancestor so the lineage is explicit. What you may not do is open a clean slate against your own words. And when blue applied your exact text, your prescription HAS been satisfied at that site by construction: close the gap unless you can state what your own acceptance_check still fails. The guard only catches text blue took verbatim — text blue COUNTER-EDITED is blue's authorship and you audit it normally. This is an amendment to the gap that prescribed it, not a fresh gap — 3 of 3 round-2 gaps in the 2026-08-04 smoke were text blue added in round 1 because red asked for it. Decide the binary verdict — PASS only when every remaining unadjudicated gap is closed, rebuttal_sustained, or risk_accepted. CITATIONS_CHECKED IS THE RECORD'S, NOT YOURS: set the envelope's citations_checked to counts.citations from the board (the \`board\` projection) — the tool counts the cite events the lenses recorded; never hand-count or estimate it. Record your round-${round} narrative on the RECORD as a position event carrying your RED narrative; it renders as the round's "### RED" section from the record. CLOSING ARGUMENTS: any gap you RE-RAISE from a prior round, any successor you mint via supersedes, and any grade dispute you REJECT is docket-bound — for EACH such item, record a closing event on that gap, your case in ~120 words; it renders as "### RED CLOSING" from the record and the judge reads it there: your strongest evidence the gap is real and graded correctly, and your answer to blue's best rebuttal so far. The judge rules AFTER blue responds this round, on the closings, the transcript, and the final artifact state only — your closing is your case; overstatement the record does not support counts against you.${petitionClause(`red-merge-r${round}`)}${reliefFor('red')}${frictionClause(`red-merge-r${round}`, 'merge')}${speedClause}${recordClause(`red-merge-r${round}`, 'red-merge')} Return the red envelope.`,
    { ...judgment, label: `red-merge-r${round} · ${slug}`, phase: 'Red', agentType: 'frank-exchange-of-views:red-auditor', schema: RED_ENVELOPE })

  takeFriction(`red-merge-r${round}`, redEnv)
  if (!redEnv) throw new Error(`red-merge round ${round} returned null (agent failed) — aborting cleanly`)
  log(`round ${round}: red ${redEnv.verdict} — ${redEnv.gaps.length} gaps open, mass ${redEnv.gaps.reduce((s, g) => s + gapMass(g), 0).toFixed(1)}, ${redEnv.citations_checked} citations checked`)
  // Degenerate-shape guard (retrospective §3 row 20, decided R4-2: throw, never soft-convert):
  // FAIL with zero gaps is evidence of a broken merge, not a clean report — looping on it
  // burns maxRounds silently and returns a self-contradictory UNVERIFIED/0-gaps verdict.
  if (redEnv.verdict === 'FAIL' && redEnv.gaps.length === 0) {
    throw new Error(`red-merge round ${round} returned FAIL with an empty gaps array — degenerate merge, refusing to loop silently`)
  }
  // Lineage enforcement (row 23 step 4, per red's own R5-5 critique: self-declared lineage
  // is hollow unless structurally checked): every closed_with_regression closure must have
  // a successor gap naming it in supersedes — otherwise the chain is silently dropped and
  // the docket goes blind again.
  // THIS GUARD READS THE ENVELOPE, WHICH IS A LOSSY SUMMARY OF THE RECORD — so it may NOT kill
  // the run. Measured 2026-08-04: red closed R1-1 with regression and correctly minted R2-1 with
  // `--supersedes R1-1` ON THE RECORD (the board shows `R2-1 supersedes -> [R1-1]`), then omitted
  // the optional `supersedes` field from its envelope self-report. The old hard throw killed a
  // 12-agent, 723k-token run whose lineage was entirely intact. That is the
  // lossy-channel-substitution class: a consumer handed a summary when it needed the source, unable
  // to tell the difference.
  //
  // CONTINUING IS STILL RIGHT — the 2026-08-04 measurement above is what justifies it, and this
  // clause only stops the engine trusting the summary over the source. But the REASON written
  // here was wrong three times over (#415), and a false reason reads as diligence:
  //
  //   1. "`verify`'s supersedes-resolve runs over the board at capture" — it does not run at
  //      capture, or anywhere. `internal/verify` has exactly one importer in the tree, the CLI
  //      verb, and nothing invokes that verb: not CI, not a suite, not a command.
  //   2. Even if it ran, it answers a DIFFERENT QUESTION. supersedes-resolve asks "does every
  //      named ancestor exist"; this asks "does anything name this closure at all". A closure
  //      with no successor names no ancestor, so there is nothing for it to find dangling.
  //   3. "capture will score it" — capture has no lineage audit. FrictionAudit reconciles
  //      envelope friction against on-record friction; it says nothing about supersedes.
  //
  // WHAT IS ACTUALLY ENFORCED, so this does not overcorrect into claiming nothing is: the WRITE
  // refuses a mint whose `supersedes` names an ancestor no mint created ("dangling lineage
  // refused"), in the operator's face, and `supersedes` is required in the envelope schema —
  // which is the thing that genuinely reduces the lossiness.
  //
  // WHAT IS NOT: "closed with regression and nothing names it as superseded" is checked by
  // NOTHING. Not the write, not capture, not any check that runs. Stated rather than papered
  // over — the friction line below is the only trace it leaves, and friction is a report to a
  // human, not a gate.
  for (const c of redEnv.closures || []) {
    if (c.class === 'closed_with_regression' && !redEnv.gaps.some(g => (g.supersedes || []).includes(c.id))) {
      const msg = `red-merge round ${round}: closed ${c.id} WITH REGRESSION and no successor in the ENVELOPE names it in supersedes. The envelope is a lossy summary, so this does not kill the run — but NOTHING downstream checks it either (#415), so if the record's lineage is genuinely absent it will not be caught: read the board.`
      friction.push(`red-merge-r${round}: ${msg}`)
      log(msg)
    }
  }
  // The two self-report count gates (archive_blocks vs ledger_closure_lines, and the
  // round-2 spot-check floor keyed on prevArchiveBlocks) are REMOVED 2026-07-19 — they
  // compared numbers the merge made up (a haiku smoke self-reported archive_blocks:22 in a
  // round where the true archived count was 0). The tool board is the count authority;
  // capture audits the truth from disk. See plans/tool-is-the-contract.md and
  // plans/pre-dry-run-batch.md.

  // W2b script-side scorecard computation (independent of the merge's self-reported
  // telemetry line — the script recomputes from envelopes it holds, so a wrong line is
  // visible as divergence). repair_regression_ratio is blue's headline number.
  if (round >= 2) {
    const curIds = new Set(redEnv.gaps.map((g) => g.id))
    const closedIds = new Set(prevGaps.map((g) => g.id).filter((id) => !curIds.has(id)))
    const lineageMints = redEnv.gaps.filter((g) => (g.supersedes || []).some((a) => closedIds.has(a))).length
    let down = 0, up = 0
    for (const g of redEnv.gaps) for (const a of g.supersedes || []) {
      const anc = prevGaps.find((p) => p.id === a)
      if (!anc) continue
      const d = gapMass(g) - gapMass(anc)
      if (d < 0) down += -d; else up += d
    }
    const ratio = closedIds.size ? (lineageMints / closedIds.size).toFixed(2) : 'n/a'
    log(`round ${round} scorecard: repair_regression ${lineageMints}/${closedIds.size} = ${ratio} · edge deltas down ${down.toFixed(1)} / up ${up.toFixed(1)} mass`)
    // Never-hard-fail DETECTOR (visibility only — the verdict stays red's): converged
    // board + fallout-only discovery + still FAIL is the divergence the reformed telos
    // says should not persist for long.
    const mass = redEnv.gaps.reduce((s, g) => s + gapMass(g), 0)
    const maxSevMass = Math.max(0, ...redEnv.gaps.map((g) => MASS[g.severity] ?? 0))
    const freshMints = redEnv.gaps.filter((g) => !prevGaps.some((p) => p.id === g.id) && !(g.supersedes || []).length).length
    if (redEnv.verdict === 'FAIL' && mass < 35 && maxSevMass <= MASS['medium'] && freshMints === 0) {
      log(`round ${round} DETECTOR: convergence-vs-verdict divergence — mass ${mass.toFixed(1)}, nothing above medium, zero fresh (non-lineage) mints, verdict FAIL. Visibility only; certification stays red's call.`)
    }
  }

  if (await hearPetitions(redEnv, `red-merge-r${round}`)) break

  // Grade-dispute processing (run-4 §3.3): every pending dispute gets red's answer or the
  // docket. Explicit rejection is HELD one round (dockets only on blue's re-dispute);
  // silence auto-dockets — default-to-docket punishes silence, not disagreement.
  const disputeDocket = []
  let acceptedDeltaMagnitude = 0
  const acceptedDeltas = []
  for (const d of pendingDisputes) {
    const resp = (redEnv.dispute_responses || []).find(r => r.gap_id === d.gap_id && r.dimension === d.dimension)
    if (!resp) {
      disputeDocket.push({ ...d, traffic_class: 'grade_dispute_unaddressed' })
    } else if (resp.response === 'rejected') {
      heldDisputes.set(`${d.gap_id}|${d.dimension}`, d)
    } else {
      const g = redEnv.gaps.find(x => x.id === d.gap_id)
      const current = g ? g[d.dimension] : null
      const delta = current != null ? Math.abs((MASS[d.proposed] ?? 0) - (MASS[current] ?? 0)) : 0
      acceptedDeltaMagnitude += delta
      acceptedDeltas.push({ gap_id: d.gap_id, dimension: d.dimension, proposed: d.proposed })
    }
  }
  if (overflowDisputes.length) {
    disputeDocket.push({ traffic_class: 'grade_dispute_overflow_batch', disputes: overflowDisputes })
    overflowDisputes = []
  }
  // Clause (v) second guard: cumulative accepted-delta magnitude per round crossing the
  // threshold batch-dockets to the judge BEFORE the deltas stand — computed by the script
  // (the one seat that sees every envelope), cumulative not per-delta (salami-slicing is
  // out-of-spec by construction).
  if (acceptedDeltaMagnitude > ACCEPTED_DELTA_DOCKET_THRESHOLD) {
    disputeDocket.push({ traffic_class: 'accepted_delta_overflow', cumulative_magnitude: acceptedDeltaMagnitude, deltas: acceptedDeltas })
  }
  gradeAdjustments = []

  if (redEnv.verdict === 'PASS') break

  // Contested docket: a dispute that persists — same id re-raised from ANY prior round
  // (window is the whole debate, not just last round), or a successor gap whose supersedes
  // chain descends from a prior-round gap (retrospective §3 row 23: regression chains ran
  // four generations in run 3 and the id-equality detector never armed; the judge was
  // dispatched ZERO times in the entire corpus). Detection is set arithmetic in the script;
  // the judgment belongs to the lead-judge. TRAFFIC CLASSES (run-4 friction, judge-r2): a
  // re-raised id has a blue response on record; a regression successor is FIRST-RAISE
  // traffic — closed/rebuttal_sustained are structurally dead for it and the judge is told
  // so instead of being handed a dead decision space.
  const contested = []
  for (const g of redEnv.gaps) {
    const reRaised = allPriorGapIds.has(g.id)
    const descends = (g.supersedes || []).some(id => allPriorGapIds.has(id))
    if (!reRaised && !descends) continue
    // Carried persistence: a standing carried ruling absorbs re-raises until red's grade
    // moves or a successor descends — otherwise the judge re-rules the same question at
    // ~$10-13 a sitting and each sitting is a fresh drift chance.
    if (reRaised && carriedRulings.has(g.id) && gradesEqual(carriedRulings.get(g.id), g)) continue
    contested.push({ ...g, traffic_class: reRaised ? 're_raised' : 'first_raise_successor' })
  }
  contested.push(...disputeDocket)
  // WITHHELD, WITH A REASON. The docket carries gaps that PERSIST (re-raised from a prior
  // round, or descending from one by supersedes). A gap minted fresh this round is not
  // docket-bound by that rule — but red and blue file CLOSING ARGUMENTS by their own
  // reading of the prose contract, and in the 2026-07-18 run both sides argued R3-2 to
  // closing while the engine never docketed it. It reached no ruling and returned to
  // red's verdict pool unadjudicated, and the bench had no way to know it existed.
  //
  // Two computations of one fact, disagreeing — the defect class this codebase keeps
  // finding. The engine cannot pre-compute the docket (it depends on red's own output,
  // produced in the same turn as the closings), so instead nothing is withheld SILENTLY:
  // the bench is handed every open gap the docket excluded, with the reason, and may
  // rule on one it judges was wrongly withheld.
  const withheld = (redEnv.gaps || [])
    .filter((g) => !contested.some((c) => c.id === g.id))
    .map((g) => ({
      id: g.id,
      severity: g.severity,
      reason: allPriorGapIds.has(g.id) || (g.supersedes || []).some((id) => allPriorGapIds.has(id))
        ? 'persistent but not carried onto the docket — report this as a docket defect'
        : 'minted fresh this round, so not a persisting dispute',
    }))

  const hasNew = redEnv.gaps.some(g => !allPriorGapIds.has(g.id) && !(g.supersedes || []).some(id => allPriorGapIds.has(id)))
  for (const g of redEnv.gaps) allPriorGapIds.add(g.id)

  const adjudicatedIds = new Set(adjudicated.map(x => x.gap_id))
  const openGaps = redEnv.gaps.filter(g => !adjudicatedIds.has(g.id))
  blueEnv = await agent(
    `Blue response, round ${round}, topic "${topic}". Red's verdict: FAIL.${reliefFor('blue')} FIRST ACTION (read batching, W1.12): concatenate your working set with a single bash call — \`the \`board\` projection > <your session scratchpad>/board-${round}.md && the \`debate\` projection > <your session scratchpad>/debate-${round}.md && cat <your session scratchpad>/board-${round}.md <your session scratchpad>/debate-${round}.md ${runDir}/inputs/red-gap-patterns.md > <your session scratchpad>/respond-${round}-workset.md\` (ABSOLUTE scratchpad path, never under ${runDir}; the transcript is the RECORD rendered as the \`debate\` projection, not a hand-written file) — then read that one file instead of three separate reads. REPORT.MD IS READ-ONLY TO YOU (the lockdown): you change ${runDir}/blue/report.md ONLY through the edit verb, which takes the exact current span, its replacement, your reasoning, and the gap it answers — it matches your span across the invisible anchor layer and REJECTS an edit that CHANGES WHICH ANCHORS EXIST — a \`<!--fx:...-->\` finding marker or a \`<!--cite:...-->\` citation anchor may travel THROUGH your edit, but never be dropped, duplicated or invented by one. YOU MAY COMPUTE AN ANSWER, NOT ONLY COLLATE SOURCES. You have Bash, Write and Edit, and for a whole class of questions running something settles it faster and harder than arguing about it. WORKING IT OUT IN YOUR HEAD IS NOT THE EXCEPTION — IT IS THE CASE THIS EXISTS FOR. Measured: a seat facing a gap that said the stated total did not match the per-source rows summed twelve integers inside its own reasoning, wrote the answer into the report, and moved on. The answer was RIGHT, which is exactly why nothing caught it: a correct figure with no derivation is indistinguishable from a confident guess, and a reader cannot vary the rate, re-check the sum, or locate the error on the day there is one. Every gap on the board and the work projection carries a check_kind AND an awaiting_proof flag — READ THEM — and a gap showing awaiting_proof: true is waiting on a program FROM YOU. A gap whose check_kind is \`computation\` cannot be closed on prose at all: it closes only when you answer it with a PROOF naming that gap, and the merge is refused if it tries to close it any other way. Write a script under the run directory and record it against the exact sentence it backs, naming the script and, where one applies, the METHOD citation it uses. The tool runs it TWICE, caches the script and its output, and splices an invisible proof anchor at that sentence. Identical output twice is recorded as reproducible; output that moves is recorded as observed — still evidence, of a system in motion, but a MEASUREMENT rather than a proof, and the tool says which you produced. THIS IS THE CITATION MODEL WITH THE LAST MILE WALKED, not a new kind of evidence: cite the METHOD with the cite verb (trial division, Miller-Rabin, the formula, the standard) and then prove the INSTANCE by running it. Several methods on one claim is several anchors on one sentence, which is how a mathematician argues. A FAILING CHECK IS A RESULT: disproving your own claim is worth as much as confirming it, and the tool records a non-zero exit as evidence rather than an error. Measured, so you know why this clause exists: red once asked for the protocol to be tested on a false claim, blue answered in PROSE that the test had happened, and red correctly refused it for showing no evidence — a whole round for something three lines settle. YOUR LINES OF INQUIRY ARE A LIVING RECORD, NOT A ROUND-0 PLAN. PROPOSE a direction with the approach and what would be TRUE if it paid off, and the tool assigns it an id. You may propose one AT ANY POINT in the run, not only at the start — a direction that occurs to you in round 3 is exactly the one a future run needs. EVERY ROUND, revisit what is still open: any line of inquiry sitting at proposed or pursued owes a MOVE to its fate — pursued, declined, abandoned or deferred — with what you learned that changed it, or an explicit reaffirmation. DEFERRED is the fate for a direction worth taking but NOT BY THIS RUN — not declined (judged not worth it), not abandoned (tried, died), but kept: its reason says what a later run should pick it up FOR, and it reaches the report's carried-forward questions as a PROPOSAL a human selects, never as an automatic seed for the next run. The hypothesis is what makes that honest — a line of inquiry abandoned against its own stated claim is evidence of choosing; one abandoned on a shrug is not. Measured over six runs: 83 of 86 lines of inquiry were declared in round 0 and NOT ONE was ever revisited, so the record showed a plan and never showed you choosing. Red rules on your proposals and you may APPEAL a ruling on the line's own id — file the appeal whether or not you go on to pursue the line, because the appeal is where your argument is recorded and the status move is only what you decided to do about it. WHERE RED PROPOSED EXACT TEXT (a gap carrying location/fix_new, fix_basis: verified), you have THREE paths and you are not obliged to take the first: apply it verbatim, counter-edit with your own fix, or dispute it. Applying is the cheapest — a counter-edit costs a round and invites re-audit, a dispute needs red's agreement — so say plainly in --reason which path you took and why. A round in which you never decline is not agreement, it is capitulation, and it IS measured — the changes view prints offered/applied/declined every round, and a zero decline rate is read as a question about red's proposals, not as evidence they were all right. Applying red's exact text also ESTOPS red from re-raising that text as a fresh gap, so verbatim application is a real settlement, not a surrender. WHEN THE EDIT ANSWERS A GAP, NAME THE GAP IN ITS OWN FIELD: that is the join key between the fix red asked for and the change you actually made, and it is the ONLY place that link is recorded. Writing the gap id into --reason instead is REFUSED by the tool (prose is your argument, not a key); leave it unset only for an edit that answers no gap, such as your own clarity or punctuation repair. If your span contains an anchor, reproduce it EXACTLY in the replacement (copy the token verbatim, anywhere in it); the tool compares the anchors in the span against the replacement and refuses any mismatch. A raw Edit/Write or a shell write to report.md is DENIED; the tool is the one path. To CITE a source for a new or changed claim, use the cite verb — it caches the source and splices the invisible citation anchor; you never hand-type a footnote, and an unreachable url is rejected + logged as friction. OWNERSHIP STILL BINDS (same as round 0): you edit ONLY your own surfaces — TL;DR, Catechism, technical foundations, analysis, open questions, your citations (added with the cite tool) — and you MUST NOT introduce a \`**Verdict:**\` line or any tool-owned section (\`## Risk matrix\`, \`## Red team findings\`, \`## The debate\`, a \`## Blue team report\` wrapper) while repairing: assembly composes those from the record, you cannot know the outcome or red's findings, and a verdict or tool-owned section you add here only lands stale and gets stripped. PRE-FLIGHT: re-check your planned repairs against red's gap patterns (in the workset). Open gaps (adjudicated ones excluded): ${JSON.stringify(openGaps)}. GRADING SEMANTICS (mapping ${MASS_MAPPING_VERSION}, and NOT inferable from the JSON above — the map is self-describing as structure and silent on meaning): the likelihood field grades the CONSEQUENCE ONLY — how likely the harm is to land, never how likely the defect is to BE there. A typo you have confirmed, whose harm is trivial, is a LOW likelihood; reading likelihood as confidence-that-it-exists is the v1 semantics this mapping replaced. CORROBORATION IS ON THE RECORD, NOT IN THIS PROMPT (#326): read what red actually verified, and at what confidence, from the evidence view — the \`evidence\` projection. The tool renders it from the record, so it is the source rather than a hand-copied snapshot of one round, and it stays current as red keeps auditing. THE SAME VIEW RESOLVES YOUR OWN ANCHORS: every source you cited is listed under \`sources\` by its \`<!--cite:c-…-->\` id and every computation under \`proofs\` by its \`<!--proof:p-…-->\` id, so a token you meet mid-edit is a lookup rather than a guess — and a proof showing \`verified: null\` is one red has not re-run. A claim red graded LOW is where your evidence is thinnest and is worth strengthening before red returns to it. BEFORE drafting, read the transcript from the RECORD — the \`debate\` projection — the gap JSON above is a lossy summary of it, and it lists any accepted grade-dispute deltas pending their contest window; read red's latest "### RED" narrative there and the latest "### LEAD" resolutions if any: any gap the judge CARRIED comes with a stated research direction you owe. Address every open gap ADDITIVELY in ${runDir}/blue/report.md — expand and repair where red is right, rebut in writing (with evidence) where red is wrong, and argue risk-acceptance where the fix's complexity exceeds its likelihood x impact; compact and reorganize prose as clarity demands, edit by edit through the edit verb (a span containing an anchor is fine — carry the anchor token across verbatim into the replacement; quote enough context that your span matches exactly ONE place, or the tool refuses it as ambiguous rather than editing the wrong site) — but a CLAIM leaves only through the retire verb, which takes the claim as it stood, why it goes, and what replaces it where something does — never by quietly not being there any more; capture compares the claim_count fall against the retire events and an unaccounted drop is a detector hit. Propagate every correction to ALL sites that state the corrected claim, not only the flagged sentence (incomplete propagation was run 3's dominant blue failure class — 5 regressions in 5 rounds). Two locators, both cheap scans of your OWN report (not a whole-report re-read — that is the once-per-cycle working-set read you already did): for a CITED claim, the claim-index verb lists every site of each citation — fix each; AND for an unfootnoted restatement a marker index cannot see (a bare corrected FIGURE), grep the corrected strings/figures report-wide. Log the sites checked in your round record (the revision verb). ANCHOR PRESERVATION IS TOOL-ENFORCED (you no longer police it by hand): the edit verb compares the anchors in your span against your replacement and REJECTS any edit that would drop, duplicate or invent one, so you cannot tamper with red's immortal audit record or your citation record even by accident (an anchor is BORN only from the finding and cite verbs and removed only by tool). You MAY rewrite an anchored sentence: include the anchor token in your span and copy it verbatim into the replacement. That is transit, not authorship — the tool checks the bytes. If a PostToolUse check reports a dropped anchor, you removed it through some non-tool path — restore it exactly and re-apply through the edit verb.${contested.length ? ` CLOSING ARGUMENTS: the following items are DOCKETED for adjudication AFTER your response this round: ${JSON.stringify(contested)}. For EACH docketed item, after your repairs, record a closing event on that gap — why your response resolves it, or why red's grade or claim is wrong; ~120 words, citing the exact section and evidence; it renders as "### BLUE CLOSING" from the record. The judge rules on the closings, the transcript, and the final state of the artifacts — your closing is your case; material not in the record cannot help you.` : ''} GRADE DISPUTES (your contest path on red's grading): to dispute a gap's severity/likelihood/impact/complexity, EMIT the argument as a motion against that gap, naming the axis and the grade you say it should be, with your evidence — the tool assigns the motion an id (M1, M2 …) and prints it; that id is what red's ruling names, so an ask and its answer are one row under "## Motions", where red answers and the bench reads it. If red rejects it and you still disagree, APPEALING the motion on its own id presses it to the bench; AND list each in the envelope's grade_disputes as a ROUTING REF ONLY (gap_id, dimension, proposed — no prose: the evidence lives on the record, not a second copy in the envelope) so the docket routes it — max ${DISPUTE_CAP} per round (overflow is batch-docketed to the judge as one item)${heldDisputes.size ? `; disputes red REJECTED last round (re-dispute any of these to send it to the judge): ${JSON.stringify([...heldDisputes.values()])}` : ''}. PROBE CLASSES (W1.10): where a fix demands a probe, discharge a DOCUMENT-PROBE now (against shipped artifacts) or, for a LIVE-PROBE needing built artifacts, name it as a deferred acceptance test with its pass condition — never call a file read a live probe. THE MANIFEST (W2b — your self-audit's receipt; your constitution carries the full eight rows): for EVERY gap you repair, run the correctness manifest — figures recomputed, universals enumerated, consistency sites swept report-wide, the repair's own boundary case asked, compositions noted where edits share text, sibling sweep or declared-open enumeration, the gap's acceptance_check RUN with its result, new claims tagged verified-at-leaf/derived/asserted — and RECORD ONE ROW PER REPAIRED GAP WITH THE TOOL, naming the gap and compressing what you checked and what it showed into its reason — then list just those gap IDS in the envelope's manifest array as a routing ref. The ROW GOES ON THE RECORD, not in the envelope: it is your receipt, it is scored from the record at capture, and IT REACHES THE READER — the report carries your manifest, and a closed gap with no row is named there as a repair nobody audited, including its author. An unmanifested repair is unaudited by your own standard; the script rejects an EMPTY manifest array on a round with open gaps. CLAIM UNIT (pinned): claim_count = what "${binDir}/feov-record" count-claims --run ${runDir} prints — recompute it after your report edits land and relay the integer into the envelope; never hand-count it. Record the round, with what changed this round including claim_count; and record your round-${round} narrative on the RECORD as a position event carrying your BLUE narrative (it renders as the round's "### BLUE" section). ROUND RECORD (W1.7): round_record_appended = TRUE only after BOTH your position event AND your round-${round} revision event exist. RECORD THE ROUND. IT IS THE ROUND RECORD, and it renders as the report's revision history. Capture counts these events, and always did: the 2026-08-05 run carried a 6,847-byte hand-written CHANGELOG and exactly ONE revision event across three eligible blue seats, so the parity gap was scored against blue for a verb no prompt had ever named. The event is the record; there is no file. On false the script re-prompts you once for the round record and, failing that, logs friction and continues — the run is not discarded, but the parity gap is scored against you at capture; a revision is not on the record until the transcript carries it (the round-2 desync misled a lens and blinded the judge, run 5).${patternDutyClause(openGaps)}${inquiryClause}${petitionClause(`blue-respond-r${round}`)}${frictionClause(`blue-respond-r${round}`, 'blue')}${speedClause}${recordClause(`blue-respond-r${round}`, 'blue')} Return the blue envelope.`,
    { ...bulk, label: `blue-respond-r${round} · ${slug}`, phase: 'Debate', agentType: 'frank-exchange-of-views:blue-researcher', schema: BLUE_ENVELOPE })
  takeFriction(`blue-respond-r${round}`, blueEnv)
  if (!blueEnv) throw new Error(`blue response round ${round} returned null (agent failed) — aborting cleanly`)
  await ensureRoundRecord(blueEnv, `blue-respond-r${round}`, `your round-${round} position event (it renders as the "### BLUE" section) AND your round-${round} revision event`,
    { ...bulk, agentType: 'frank-exchange-of-views:blue-researcher' })
  if (openGaps.length > 0 && (!Array.isArray(blueEnv.manifest) || blueEnv.manifest.length === 0)) {
    throw new Error(`manifest (W2b): blue-respond round ${round} repaired ${openGaps.length} gap(s) with an EMPTY correctness manifest — an unmanifested repair is unaudited by blue's own standard`)
  }
  if (openGaps.length > 0) {
    // #318: the array is gap IDS now — the rows themselves are on the record (`blue
    // manifest-row`). No object fallback: the schema rejects a non-string before this runs, and
    // accepting both shapes would be the alias that lets the old channel quietly keep working —
    // which is the exact mechanism that kept this migration half-done for a year.
    const covered = new Set(blueEnv.manifest || [])
    const uncovered = openGaps.filter((g) => !covered.has(g.id)).map((g) => g.id)
    if (uncovered.length) log(`round ${round}: manifest coverage ${covered.size}/${openGaps.length} — unmanifested: ${uncovered.join(', ')} (scored at capture)`)
  }
  log(`round ${round}: blue responded — ${openGaps.length} gaps addressed, corpus at ${blueEnv.claim_count} claims${(blueEnv.grade_disputes || []).length ? `, ${(blueEnv.grade_disputes || []).length} grade dispute(s) raised` : ''}`)
  if (await hearPetitions(blueEnv, `blue-respond-r${round}`)) break

  // Adjudication sits LAST in the round (closing-arguments redesign, 2026-07-17): the judge
  // never rules on material blue has not answered — which structurally dissolves the
  // first-raise-successor dead-options problem the traffic-class patch worked around. The
  // ruling basis is confined to the two closings, the transcript, and the final state.
  if (contested.length > 0) {
    log(`round ${round}: docket — ${contested.length} contested item(s) to the judge (closings filed)`)
    const judge = await agent(
      `Adjudication, round ${round}, topic "${topic}". Contested docket: ${JSON.stringify(contested)}.${withheld.length ? ` OPEN GAPS WITHHELD FROM THE DOCKET (with the reason each was withheld): ${JSON.stringify(withheld)}. These are NOT docketed, and you are not obliged to rule on them — but a gap BOTH SIDES argued to closing that reached no ruling is a docket defect, not a decision. If you find one, say so in your opinion and rule on it: an unadjudicated gap returns to red’s verdict pool as though nobody had considered it.` : ''} STALENESS: the contested docket above carries each gap's id and grades ONLY (envelope-refs) — a gap's problem text and acceptance_check live on the BOARD, so read them FRESH from the ledger (the \`board\` projection, computed from the event log on read, so never a stale snapshot). The docket is a routing list, not the evidence: in the 2026-07-18 run the bench mis-ruled on snapshotted docket premises that were false by the time it sat — reading the record fresh is the fix. Re-run each DOCUMENT-PROBE acceptance check (as the board states it) against the artifact AS IT NOW STANDS before ruling, and rule on what you find rather than on what any snapshot asserts. Both sides have filed closings for this docket on the RECORD (red's "### RED CLOSING" and blue's "### BLUE CLOSING", rendered from the closing events). YOUR RULING BASIS IS CONFINED TO: (1) the two closings, (2) the full transcript (read it from the record: the \`debate\` projection), and (3) the final state of the artifacts — the board (pull it FRESH through the tool: the \`board\` projection — the tool computes it from the event log on read, so it is never stale) and ${runDir}/blue/report.md as they now stand. Weigh the closings as each side's best case; a claim in a closing that the record does not support counts AGAINST the side that made it.${lawClause}${declareClause} Resolution set (full, for every gap class — blue has answered everything docketed): closed (blue's response resolves it) | rebuttal_sustained (blue's evidence beats the challenge) | risk_accepted (valid, rejected on likelihood x impact x complexity) | carried (still live — state what further research blue owes) | unresolved | moot (the predicate expired — the claim or artifact it attached to is not in the report) | grade_adjusted (for grade_dispute_* items: state the corrected grade in the rationale; red applies it next round) | routed_to_infrastructure (valid finding whose FIX is owned outside the debate — run tooling, harness, or the lead; state the owed fix in the rationale; it leaves red's verdict pool and ships as a named infrastructure debt, recorded never dropped). DEMANDED READS: for every ruling on a gap with a supersedes chain, you MUST read the named ancestors' records (pull them through the tool: the \`board\` projection) first and NAME the records read in your rationale. deadlock is true only if no gap is carried AND ${hasNew ? 'false (new gaps were raised this round)' : 'no new gaps were raised this round (none were)'}. Record each resolution as an opinion event — one per docketed gap, each carrying the gap, its fate, the principle you applied, the values in tension, whether a human should look at it, and your rationale; they render as the round's "### LEAD" from the record.${frictionClause(`judge-r${round}`, 'bench')}${speedClause}${recordClause(`judge-r${round}`, 'bench')} Return the judge envelope.`,
      { ...judgment, label: `judge-r${round} · ${slug}`, phase: 'Debate', agentType: 'frank-exchange-of-views:lead-judge', schema: JUDGE_ENVELOPE })
    if (!judge) throw new Error(`judge round ${round} returned null (agent failed) — aborting cleanly`)
    for (const r of judge.resolutions) {
      if (r.resolution === 'closed' || r.resolution === 'rebuttal_sustained' || r.resolution === 'risk_accepted' || r.resolution === 'moot' || r.resolution === 'routed_to_infrastructure') adjudicated.push(r)
      if (r.resolution === 'routed_to_infrastructure') infraDebts.push({ gap_id: r.gap_id, owed_fix: r.rationale, round })
      if (r.resolution === 'carried') {
        const g = redEnv.gaps.find(x => x.id === r.gap_id)
        if (g) carriedRulings.set(r.gap_id, gradeSnapshot(g))
      }
      if (r.resolution === 'grade_adjusted') gradeAdjustments.push({ gap_id: r.gap_id, rationale: r.rationale })
    }
    takeFriction(`judge-r${round}`, judge)
    if (judge.deadlock) { deadlocked = true; break }
  }

  // Dispute intake (run-4 §3.3): re-disputed held items go straight to next round's docket;
  // fresh disputes await red's response; beyond the cap, overflow batch-dockets as ONE item.
  const raised = blueEnv.grade_disputes || []
  pendingDisputes = []
  for (const d of raised.slice(0, DISPUTE_CAP)) {
    const key = `${d.gap_id}|${d.dimension}`
    if (heldDisputes.has(key)) {
      heldDisputes.delete(key)
      overflowDisputes.push({ ...d, traffic_class: 'grade_dispute_re_raised' })
    } else {
      pendingDisputes.push(d)
    }
  }
  if (raised.length > DISPUTE_CAP) overflowDisputes.push(...raised.slice(DISPUTE_CAP).map(d => ({ ...d, traffic_class: 'grade_dispute_over_cap' })))
}

const exhausted = round >= maxRounds && redEnv && redEnv.verdict !== 'PASS' && !deadlocked
// Terminal-state taxonomy (rulebook audit item 7). The protocol named two
// terminators — red-PASS and confirmed deadlock — and runs 3, 5 and the smoke all
// ended by a THIRD: the round ceiling. Run 5's ceiling fired while the bench had
// ruled deadlock FALSE and blue had shipped a revision no red pass ever audited,
// and the template had no verdict class for "converging, ceiling-terminated,
// final revision unaudited". It was handled by an assembly-owned risk row and an
// obligation carried out of the run by hand. CEILING is that class, and it says
// what UNVERIFIED cannot: the run did not fail to converge, it ran out of budget
// mid-flight, and something specific is owed.
const ceilingUnaudited = exhausted && !halted && !deadlocked && redEnv && redEnv.verdict !== 'PASS'
const verdict = halted ? 'HALTED' : redEnv && redEnv.verdict === 'PASS' ? 'VERIFIED' : ceilingUnaudited ? 'CEILING' : 'UNVERIFIED'
log(`debate ended: ${verdict} after ${round} round(s)${halted ? ' (JUDICIAL HALT)' : deadlocked ? ' (judged deadlock)' : exhausted ? ' (safety ceiling hit)' : ''}`)

// Terminal dispute disposition (run-4 §3.3 clause (vi) — exit-agnostic: PASS, deadlock, or
// ceiling): pending or held disputes at ANY exit auto-docket for judge disposition BEFORE
// assembly; `carried` is excluded at a terminal exit (there is no next round to carry into) —
// a carried-at-exit dispute would exit looking disposed while the contested grade ships.
const terminalDisputes = [
  ...pendingDisputes.map(d => ({ ...d, traffic_class: 'grade_dispute_terminal_pending' })),
  ...[...heldDisputes.values()].map(d => ({ ...d, traffic_class: 'grade_dispute_terminal_held' })),
  ...overflowDisputes,
]
if (terminalDisputes.length > 0) {
  const terminalJudge = await agent(
    `Terminal dispute disposition for topic "${topic}" (debate ended ${verdict} after round ${round}; this docket fires at the exit boundary). Undisposed grade disputes: ${JSON.stringify(terminalDisputes)}. Read the transcript in full from the RECORD (the \`debate\` projection), and pull the board through the tool: the \`board\` projection. ${lawClause}${declareClause}${inspectionClause} The resolution set at a terminal exit EXCLUDES carried — rule each dispute grade_adjusted (state the corrected grade in the rationale; assembly records the delta) or unresolved (the contested grade ships, recorded as contested in the report). deadlock: false. Record your terminal ruling as an opinion event on that gap — it renders as "### LEAD (terminal disputes)" from the record.${frictionClause('judge-terminal', 'bench')}${speedClause}${recordClause('judge-terminal', 'bench')} Return the judge envelope.`,
    { ...judgment, label: `judge-terminal · ${slug}`, phase: 'Assemble', agentType: 'frank-exchange-of-views:lead-judge', schema: JUDGE_ENVELOPE })
  if (terminalJudge) {
    takeFriction('judge-terminal', terminalJudge)
    for (const r of terminalJudge.resolutions) {
      if (r.resolution === 'grade_adjusted') gradeAdjustments.push({ gap_id: r.gap_id, rationale: r.rationale })
    }
  }
}

// ---- Assemble: the tool composes from the record; the seat authors nothing ----
// The assemble seat reads the board last, so it is where the authoritative open-gap
// count leaves the run. debate.js is sandboxed (no tool, no fs), so this number can
// only reach the result envelope through a seat — and reporting red's docket length
// instead (the old gaps_outstanding) misread 10 resolved as 10 outstanding (#83).
const ASSEMBLE_ENVELOPE = {
  type: 'object',
  additionalProperties: false,
  required: ['synopsis', 'open_gaps'],
  properties: {
    synopsis: { type: 'string' },
    open_gaps: { type: 'integer', minimum: 0 }, // the board's counts.open, not red's docket
    friction: { type: 'array', items: { type: 'string' } },
  },
}
phase('Assemble')
const assembleEnv = await agent(
  `Final assembly for topic "${topic}", run directory ${runDir}. Debate outcome: ${verdict} after ${round} round(s)${deadlocked ? ' by judged deadlock' : ''}${exhausted ? ' by safety ceiling' : ''}. THE REPORT IS ASSEMBLED FROM THE RECORD — you author NOTHING, you copy NOTHING, you fill in NO inputs. After registering, exactly two tool calls in order: (1) record the terminal verdict as a fact: it is ${verdict}${deadlocked ? ', ended by judged DEADLOCK' : exhausted ? ', ended against the CEILING' : ''}, and the verb takes that verdict, how it ended, and your reason ("<HOW THIS RUN ENDED, IN YOUR WORDS — required, as it is on every claim and judgment act here. The VERDICT is derived from the record and needs no defence; this is your account of the sitting, and where the run ended by judged DEADLOCK it is the only evidence that determination will ever have, because a deadlock leaves nothing on the record a later reader can consult>"; (1a) IF YOU HOLD SOMETHING THAT BINDS HOW THE RECORD IS READ BUT MOVES NO GAP — a construction of a term, a correction of what the record MEANS without touching what it says, or a holding you want promoted to precedent — DECLARE IT — the verb is in your role's help, and it takes the holding as its reason and nothing else. It renders under ### LEAD in the \`debate\` projection, which is where both seats read. This verb exists because a bench once had exactly such a finding, found that ruling on a gap demands a gap id and a fate it did not want to move, and put the holding in a petition ruling instead — the channel least likely to be read; (1b) IF ANYTHING IN THIS RUN NEEDS A HUMAN TO RE-EXAMINE IT — an unresolved tension, a claim that held only because nobody could reach the source, a boundary you ruled close to — SAY SO — the verb is in your role's help, and it takes what you would want a human to look at, and why. A certification is NOT a verdict and does not change one: a VERIFIED run can carry it. You keep no memory between runs, so this statement is the only continuity you have, and it was named in no prompt until 2026-08-13 — the capability existed and nobody was told; (2) ASSEMBLE. The assembler composes ${runDir}/report.md FROM THE EVENT LOG (the verdict stamp from the outcome event you just wrote, the graded risk matrix, the expansions and alternatives-considered from the line of inquiry record, red's findings in full, and the entire debate transcript — positions, closings, grade disputes, the bench's opinions and petition rulings, the terminal halt/certification) and LIFTS blue's audited sections VERBATIM from blue/report.md (the title, the TL;DR, the Catechism, the technical foundations, the analysis, the open questions). A tool cannot mis-author a synthesis surface — the run-5 catechism defect (6/7 answers regressed when assembly authored the catechism) is structurally impossible, and the TL;DR is blue's, inside the audited report, rather than a synthesis surface written after red's final audit where nothing checks it. There are NO <FILL> fields and NO sections for you to write; do not hand-write report.md and do not copy anything yourself. Verify ${runDir}/report.md exists and read its top — the verdict stamp is the outcome event's, the sections are blue's and the record's. AUTHORITATIVE OPEN COUNT: after assembling, run the \`board\` projection and report its counts.open as open_gaps in your envelope — this is the run's true outstanding-gap number (the board after every closure and ruling), and it is the ONLY correct source for it. Collated friction so far (report any of your own as well): ${JSON.stringify(friction)}.${frictionClause('assemble', 'bench')}${speedClause}${recordClause('assemble', 'bench')} Return your envelope: a 5-line synopsis, open_gaps from the board, and your own friction if any.`,
  { ...judgment, label: `assemble · ${slug}`, agentType: 'frank-exchange-of-views:lead-judge', schema: ASSEMBLE_ENVELOPE })

return {
  runDir,
  verdict,
  rounds: round,
  lanes,
  deadlocked,
  // The board's open count (via the assemble seat) is authoritative — it reflects every
  // closure and judge ruling. Fall back to red's docket length only if the seat could not
  // report it (no binDir / structured return), which is the old, over-counting behaviour.
  gaps_outstanding: assembleEnv && Number.isInteger(assembleEnv.open_gaps)
    ? assembleEnv.open_gaps
    : (redEnv && redEnv.verdict !== 'PASS' ? redEnv.gaps.length : 0),
  blue_claims: blueEnv ? blueEnv.claim_count : null,
  infra_debts: infraDebts,
  petitions: petitionLog,
  halted,
  halt_opinion: haltOpinion,
  friction,
}

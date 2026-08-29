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
const frictionClause = (who, role) => ` FRICTION (${who}) — CLOSE THIS CHANNEL BEFORE YOU FINISH: on the record, AND in the envelope's friction field. AND THE REASON OWES THE SURVEY: name the verbs you read in the tree and did NOT use, and why each was wrong for this sitting. A reason that lists only what you did is a receipt for the path you took and says nothing about the ones you rejected — and a seat that cannot name a single rejected option did not weigh any.`

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
  ? ` LINES OF INQUIRY: record every genuinely distinct approach you CONSIDERED, not only the one you took — with the hypothesis that would be true if it paid off, and later with what became of it and what changed your mind. The hypothesis is what makes the account honest, and the dead ends matter most: one you record is one the next run does not re-walk, and it is worth more than the tidy conclusion that survived. Alternatives reach the reader under their own headings — what you pursued, what you left for a later run, and what you weighed and rejected are three things a reader needs, not one list of winners. This is not narration for the report: it is what makes "explore the alternatives" checkable instead of self-attested.`
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

const scorecardClause = () => (
  // A chair reads its OWN in-run scorecard for THIS question, computed live from this run's
  // record. Never a prior run's numbers: those are Goodhart bait, topic-confounded, cross-model
  // and salience-priming. The `scorecards` arg feeds operator analytics only.
  //
  // THIS CLAUSE NAMES THE ACT, AND THE HELP NAMES THE VERB. promptverbs' catalogue gate pins
  // debate.js at zero named commands, on the standing rule that the help page is the only page
  // that instructs — so a prompt that spells an invocation is a prompt teaching a surface it does
  // not own. It also names no chair: the tool resolves that from the seat's own registration,
  // which is why the read needs no selector and why there is no way to ask for another party's
  // numbers. See #513 for the friction that established both.
  ` YOUR IN-RUN SCORECARD (THIS run, not a prior one): before you read the open docket, read how YOUR CHAIR is doing on the question in front of you so far. It is a projection of this run's record on your own surface, it needs no selector because your chair is the seat you registered as, and its own page says what the rows mean.`
)

// HOLDINGS RIDE recordClause BECAUSE IT IS THE ONE CLAUSE EVERY SEAT RECEIVES.
//
// A holding binds every seat, so threading it per-prompt would be a dozen insertion points and
// a dozen chances to miss one — which is exactly how relief reached a single hardcoded site and
// bound nobody else it was written for (#360). One carrier, no seat exempt.
//
// It renders to the empty string while no holding exists, which is every run until a bench
// declares one, so the ordinary prompt is unchanged.
const recordClause = (seatId) =>
  `${holdingsClause()}${scorecardClause()} SEAT_ID: ${seatId}. THE RECORD TOOL IS THE CONTRACT: every act of this seat happens through "${binDir}/feov-record", and the tool's own board is the ONLY source of truth for status. Routing around it into markdown is the failure this contract exists to prevent.

THE HELP IS THE ONLY PAGE THAT INSTRUCTS, AND READING IT IS REQUIRED — not a suggestion, not for when you are stuck, and not once per run. READ THE WHOLE TREE BEFORE YOU CHOOSE, NOT THE ONE PAGE FOR THE VERB YOU ALREADY PICKED. Steps 1 and 2 happen ONCE, TOGETHER, IMMEDIATELY AFTER \`register\` — before you have decided what to do, because deciding first is what makes the reading pointless. Step 3 happens per act.

  1. "${binDir}/feov-record" --seat-id ${seatId} --help — your whole surface. The page marks which entries hold commands it does not list.
  2. "${binDir}/feov-record" --seat-id ${seatId} <group> --help — for EVERY group that page listed, one after another, including the groups nested inside those. Not the group holding the verb you want: ALL of them, before you want anything. It is five or six calls and it is the only way you ever see your whole surface.
  3. "${binDir}/feov-record" --seat-id ${seatId} <group> <command> --help — BEFORE you run it.

MEASURED, WHICH IS WHY STEP 2 IS SHAPED THAT WAY. Across nine sittings seats opened 6 of 51 group pages — twelve per cent — and 90% of the commands they ran were run without ever reading that command's own page. The rule they were following said to open a group "before using any command in it", so a seat that obeyed it perfectly still only ever opened the page for a verb it had already chosen: eighteen of the twenty-three pages opened were for verbs the seat went on to run. That is not surveying a surface, it is confirming a decision — and a decision made before reading is made from memory and from this prompt's vocabulary, which is not authoritative.

A NAME YOU DID NOT READ IN THE HELP THIS SITTING IS A GUESS. Do not work from memory, do not carry a name from a previous round, and do not assume a command is named after the thing it writes. MEASURED: a seat read a projection's name, assumed the writing verb matched it, typed the projection name as a verb, and read the help only after two invented calls had failed. The projection names, this prompt's words for a concept, and the command that writes it are three different vocabularies and they do not always agree. The help is the only authoritative one, and this prompt names the JOB, never the flags.

EVERY READ IS A PROJECTION OF THE RECORD, never a file you open. The .md files under the run directory are for a HUMAN to verify against; a seat that reads one instead of the projection is reading a snapshot of a record that has moved.

YOUR REASONING IS PART OF EVERY ACT, and the report renders it: it is your argument to the other seats, on the record where they can answer it. Write it for them, not as a label for what you did.`


// Compound grades allowed: red's protocol grades finer than a 3-point scale
// (retrospective friction #6 — forced rounding lost information every round).
const GRADE = { type: 'string', enum: ['low', 'low_medium', 'medium', 'medium_high', 'high', 'certain', 'realized', 'trivial'] }

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
const MASS = { trivial: 0.5, low: 1, 'low_medium': 1.5, medium: 2, 'medium_high': 2.5, high: 3, certain: 3.5, realized: 0 }
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
    // A PETITION SITTING CAN LAY DOWN A HOLDING TOO, and #361 was filed from exactly there: the
    // bench had a construction both parties needed and put it in a petition ruling's opinion
    // text, where red never read it. Routing it needs it on this envelope as well (#503).
    holdings: { type: 'array', items: { type: 'string' } },
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
// The verb shipped (0.67.0) and renders under "### LEAD"; the harvest reads it (#413). The prompt
// carried a paragraph naming it, because for two releases no prompt and no constitution did and
// the bench could not know it existed. The paragraph is gone: `declare --help` now carries what it
// is for, when to reach for it, why `opinion` cannot hold it, and the measured case — and the
// tree walk puts that page in front of the bench before it chooses. A verb the HELP does not name
// is the capability nobody has; a verb only the PROMPT names is one the tool cannot be trusted for.
const declareClause = ''

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
    verdict: { type: 'string', enum: ['PASS', 'FAIL'], description: 'the verdict you RECORDED this round, restated — the record is the original' },
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
          id: { type: 'string', description: 'the id the tool assigned at the mint' },
          severity: GRADE, likelihood: GRADE, impact: GRADE, complexity_cost: GRADE,
          // Lineage (retrospective §3 row 23): a successor gap MUST name the prior-round
          // gap id(s) it descends from, so the contested docket can follow regression chains.
          // A ref (ids), not prose — stays.
          supersedes: { type: 'array', items: { type: 'string' }, description: 'the ancestor ids this gap replaces — an EMPTY array for a gap minted fresh this round. Required on every gap, never omitted: recording lineage on the record and leaving it out here is the desync that once killed a whole run whose record was correct' }, // [] for a fresh gap — REQUIRED, never omitted
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
          class: { type: 'string', enum: ['repaired', 'repaired_with_regression', 'amends_prior', 'not_a_defect', 'defect_accepted', 'defect_owed_elsewhere'] },
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
    citations_checked: { type: 'number', description: "the board's count of cite events, read back from the record — never hand-counted or estimated" },
    notes: { type: 'string' },
    friction: { type: 'array', items: { type: 'string' } },
  },
}

const JUDGE_ENVELOPE = {
  type: 'object',
  required: ['deadlock', 'resolutions'],
  properties: {
    deadlock: { type: 'boolean' },
    // HOLDINGS THE BENCH LAID DOWN THIS SITTING, carried so the engine can route them. debate.js
    // reads no record, so a holding recorded through `bench declare` reaches the other seats only
    // if it travels here — the same reason `relief` is on the petition envelope (#503).
    holdings: { type: 'array', items: { type: 'string' } },
    friction: { type: 'array', items: { type: 'string' } },
    resolutions: {
      type: 'array',
      items: {
        type: 'object',
        required: ['gap_id', 'resolution', 'rationale', 'settled'],
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
          // Exactly one of these is the answer to "what would reopen this". `final: true` is
          // the assertable empty case — the `friction --none` shape — so a decided question
          // stays distinguishable from a skipped field.
          reopens_on: { type: 'string' },
          final: { type: 'boolean' },
          // grade_adjusted (run-4 §3.3): "gap real, grade wrong" — the dispute-resolution
          // value the enum could not previously express. The rationale MUST state the new
          // grade; the next red-merge applies it and lists the delta.
          // moot: the gap's predicate expired — the claim or artifact it attached to is no
          // longer in the report. Moot adjudicates the gap out.
          // defect_owed_elsewhere (W1.9, run-5 judge-r2 friction): "valid finding, fix
          // owned outside the debate" — R1-7 had to wear defect_accepted as the least-wrong
          // fit. Leaves red's verdict pool like defect_accepted; collected as an infra debt
          // the final envelope and assembly surface to the lead.
          resolution: { type: 'string', enum: ['repaired', 'not_a_defect', 'defect_accepted', 'carried', 'unresolved', 'grade_adjusted', 'moot', 'defect_owed_elsewhere'] },
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
// HOLDINGS BIND EVERY SEAT AND NEVER EXPIRE, which is what makes them different from relief.
//
// `bench declare` exists because the bench sometimes states a CONSTRUCTION rather than disposing
// of a gap — how a term is to be read, for the rest of the run, by both parties. The verb
// shipped, the render shipped, the law harvest shipped, and the DELIVERY never did: a holding
// laid down in round 2 changed how the record should be read in rounds 3 and 4 and neither red
// nor blue was ever told (#503).
//
// That is the same defect the verb was created to fix, moved one carrier along. #361 was "the
// bench had no way to say it"; this was "the bench can say it and nobody hears it" — and relief
// had the identical failure before `reliefFor` (#360), so this is the same repair.
//
// ACCUMULATING, NOT ROUND-SCOPED. Relief is granted on a petition and operative for a sitting;
// a construction of a term holds until the run ends, so this list only grows.
const holdingsInEffect = []
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
// holdingsClause renders every holding in effect, for EVERY seat, as routing refs.
//
// The text is the bench's own words because a construction is operative — a seat cannot act on a
// pointer to a construction it has not read — but the REASONING stays on the record, which is
// where the seat is sent for it.
const holdingsClause = () => {
  if (!holdingsInEffect.length) return ''
  return ` BENCH HOLDINGS IN EFFECT, BINDING ON EVERY SEAT FOR THE REST OF THE RUN: ${JSON.stringify(holdingsInEffect)}. A holding construes how the record is READ; it disposes of nothing and it does not expire with the round. Read the bench's reasoning for any holding you rely on or work around.`
}

const reliefFor = (party) => {
  const mine = reliefInEffect.filter((r) => r.binds === party || r.binds === 'both')
  if (!mine.length) return ''
  return ` BENCH RELIEF IN EFFECT, BINDING ON YOU (granted on a petition and operative this round — it is an order, not advice): ${JSON.stringify(mine.map((r) => r.relief))}.`
}
let halted = false
let haltOpinion = null
// ONE SEAT ID NAMES ONE SITTING.
//
// A seat id is what every act is attributed to, so two sittings sharing one leaves the record
// unable to say which occasion an act belongs to.
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
    `Petition sitting, topic "${topic}". ${who} has petitioned the bench: ${JSON.stringify(petitions)}. Petitions are heard BEFORE the debate continues; they are never sanctioned, and a pattern of overruled petitions is at most a craft note for the petitioner. For EACH petition rule granted | denied. TWO THINGS, AND THEY GO TO DIFFERENT PLACES (#330): your OPINION is the reasoning — the principle applied, the values in tension, why a human should or should not look — and it belongs on the record beside the filing it answers. The envelope's \`relief\` is the OPERATIVE part of a granted ruling: the instruction as it will BIND the coming seats, written as a DIRECTIVE a seat can act on without reading your argument, not as a summary of your opinion. A denied ruling carries no relief. Read the transcript for context. A HALT IS A DIFFERENT DECISION FROM A RULING: it is you ending the RUN, not disposing of a petition, and both can be true at once — rule the petition on its merits and halt as well where continuing would itself compromise safety, consent gates, corpus integrity, or participant integrity. Record the halt, and ALSO return the envelope's \`halt\` object carrying that same opinion, which is what stops the engine.${lawClause}${declareClause}${inspectionClause}${frictionClause(seatID, 'bench')}${speedClause}${recordClause(seatID)} Return the petition-ruling envelope.`,
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
  for (const h of sitting.holdings || []) holdingsInEffect.push(h)
  // A halt arrives on its OWN channel, not as a ruling value (#329) — the bench records it
  // through `bench halt`, and this only tells the engine to stop.
  if (sitting.halt && sitting.halt.opinion) { halted = true; haltOpinion = sitting.halt.opinion }
  if (halted) log(`JUDICIAL HALT — ${haltOpinion ? haltOpinion.slice(0, 200) : ''}`)
  return halted
}
// ---- Frontier ----
phase('Frontier')
await agent(
  `Research debate opening for topic: "${topic}". Formulate 3-5 frontier hypotheses — what would be TRUE if each candidate answer were right — and RECORD EACH ONE AS A LINE OF INQUIRY — the verb is under a group of its own in your role's help, and it takes the approach you would follow and the hypothesis that would be true if it paid off. ONE line of inquiry per hypothesis, and the tool assigns each an id (Q1, Q2 …). THEY ARE LINES OF INQUIRY, NOT A DOCUMENT, and that is the point: a hypothesis in a markdown file is something red cannot rule on, cannot grade too_thin or out_of_scope, and cannot hold you to when the run drifts. On the record it has an id, a fate, and an argument. The hypotheses opening a run are the ones that shape everything downstream and were until now the only ones nobody could contest.${speedClause}${recordClause('frontier')} Return one line per hypothesis.`,
  { ...bulk, label: `frontier · ${slug}`, agentType: 'frank-exchange-of-views:blue-researcher' })

// ---- Blue: best-of-N lanes with method diversity, then additive synthesis ----
phase('Blue')
await parallel(Array.from({ length: lanes }, (_, i) => () => agent(
  `Blue lane ${i + 1} of ${lanes} for topic: "${topic}". Read the frontier hypotheses from the RECORD — the \`lines-of-inquiry\` projection — where each carries an id and its hypothesis, and if ${runDir}/inputs/red-gap-patterns.md exists (red's accumulated gap-pattern inventory, staged at run setup), read it too — yesterday's expensive red discovery is today's free checklist line. Research your assigned slice to saturation per the research protocol (spend at least one search in five on disconfirming evidence; note each source inline with its URL, title, and access date). Your assigned METHOD LENS: ${LANE_METHODS[i % LANE_METHODS.length]} — work primarily through this method's source class; take hypothesis ${i + 1} first, then breadth. SOURCE NOTES (not footnotes): a lane drafts candidates, it does not cite — record each intended source as prose beside the claim (its URL, title, and access date) so the synthesizer can cite it once with the tool. Do NOT mint footnote labels; citations are tool-managed at synthesis and a hand-typed label backs nothing. Write your full candidate draft to ${runDir}/blue/candidates/lane-${i + 1}.md.${inquiryClause}${speedClause}${recordClause(`blue-lane-${i + 1}`)} Return a 3-line synopsis.`,
  { ...bulk, label: `blue-lane-${i + 1} · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-researcher' })))

let blueEnv = await agent(
  `Blue synthesis for topic: "${topic}". PRE-FLIGHT: if ${runDir}/inputs/red-gap-patterns.md exists, read it before merging — check your synthesis against red's known gap patterns. Read every draft in ${runDir}/blue/candidates/ and synthesize ${runDir}/blue/report.md by UNION: deduplicate overlapping claims, reorganize freely, and let no CLAIM leave without a record. Compacting and reorganizing prose is not subtraction — rewriting for density is encouraged; losing a claim in the rewrite is the failure, and substance leaves only by being retired on the record, which names the claim, why it goes, and what replaces it. CLAIM PROVENANCE: while merging, tag every claim appearing in exactly ONE lane's draft with its lane marker (e.g. "[minority: lane-2/practitioner]") — single-lane claims are minority reports red must weigh differently from convergent ones, and that set-difference exists transiently in this merge and must not be discarded.

THE CATECHISM IS YOURS (W2g — the run-5 assembly audit found the judge-authored catechism DEFECTIVE, 6/7 answers carrying defects from synthesis-by-recall; synthesis surfaces belong inside the audited document): write "## The Catechism" INTO ${runDir}/blue/report.md at round 0 per references/catechism_template.md — the seven answers, the case against at FULL strength (include every risk-accepted residual; the audit found against-cases silently drop their strongest items), all figures copied from your own sourced sections never recomputed inline. THE FRAMING, THE TL;DR AND THE OPEN QUESTIONS ARE YOURS FOR THE SAME REASON — a synthesis surface authored at assembly is authored AFTER red's final audit, so nothing checks it and it ships unaudited. report.md OPENS with the H1 title \`# ${topic} — research report\`, then \`## TL;DR\` (3-6 sentences — the answer, the confidence, the sharpest caveat), and carries \`## Open questions\` (what the debate could not resolve, one per line; a question nobody could answer inside the debate is a finding, not noise). Assembly LIFTS these verbatim, so you write them once, HERE, inside the document red audits every round.

YOUR REPORT CONTAINS ONLY WHAT YOU CAN AUTHOR: the title, TL;DR, Catechism, technical foundations, analysis, open questions, and your citations — and NOTHING assembly composes from the record. Do NOT write "## Risk matrix", "## Red team findings", "## The debate", or a "## Blue team report" wrapper: a "## Red team findings" section you author is FABRICATION, because you cannot know red's findings, and a whole final-report-shaped document makes every tool-owned section appear twice. Do NOT write a \`**Verdict:**\` line ANYWHERE either — the outcome is decided after your last audit, so a verdict you author only lands stale beside the real one and gets stripped.

CITATION HYGIENE, because red must be able to VERIFY every source at the leaf: the url you cite carries the FULL coordinate — owner/repo#N or a full URL, never a bare #N red cannot resolve without the repo — and every QUOTED span records a locating anchor (its section heading, or a nearby unique phrase) so red can grep it verbatim in the source. An unlocatable quote is graded LOW through no fault of the source. Two claims resting on the same url are one source and two anchors, not two sources.

REPORT WRITE PATH: report.md hits a filename-keyed Write-guard — draft to your scratchpad under a NEUTRAL name and copy it into place with bash cp; a direct Write of the report path fails and wastes a round-trip. Recompute claim_count with the tool once report.md is written and relay the integer into the envelope; never hand-count it (two honest merges differed 2x before it was pinned). RECORD THE ROUND, with what this synthesis did and its claim_count as the reason; round_record_appended is TRUE only after that event exists.${inquiryClause} ${petitionClause('blue-synthesize')}${frictionClause('blue-synthesize', 'blue')}${speedClause}${recordClause('blue-synthesize')} Return the blue envelope.`,
  { ...judgment, label: `blue-synthesize · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-synthesizer', schema: BLUE_ENVELOPE })

if (!blueEnv) throw new Error('blue synthesis returned null (agent failed) — aborting cleanly')
await ensureRoundRecord(blueEnv, 'blue-synthesize', `your round-record event (the revision verb) stating the synthesis and its claim_count`,
  { ...judgment, agentType: 'frank-exchange-of-views:blue-synthesizer' })
await hearPetitions(blueEnv, 'blue-synthesize')
// ---- Debate loop: red audits gate; termination is judged, never counted ----
let round = 0
let redEnv = null
let deadlocked = false
// One-shot relief for the bench-cleared board (see the deadlock arm below). A cleared docket
// is not a deadlock, but granting red its sitting must not be repeatable, or the bench can
// extend the run indefinitely one clearance at a time.
let benchClearedOnce = false
const allPriorGapIds = new Set() // every gap id from every prior round — the docket window is the whole debate, not one round
const adjudicated = [] // judge-ruled gaps (closed / not_a_defect / defect_accepted / routed) — out of red's verdict
// WHAT A RULING OBLIGES BLUE TO DO, derived from the fate rather than restated in prose.
//
// Red's duty is "do not re-raise this gap" and an id enforces it. Blue's is "do not re-argue this
// claim in the report", which no id can enforce — blue's artifact IS the report, and that is where
// a settled proposition gets restated. And blue needs the inverse red never does: what it may now
// ASSERT. Two of these five fates are blue WINS, and under the bare subtraction blue was handed
// they looked exactly like the three that are not: the gap simply stopped appearing. A run in
// 2026-08-23 ruled twice in blue's favour on one gap, the second time final, and blue-respond saw
// an absence both times.
//
// Keyed on the resolution so a new fate cannot quietly inherit another's instruction; an unmapped
// one falls through to a LOUD default rather than an empty string.
const BLUE_DUTY_BY_RESOLUTION = {
  not_a_defect: 'THE BENCH FOUND NO DEFECT — your position was vindicated. Keep the text as it stands; do not "repair" what the bench has just blessed. You may rely on this ruling as established for the rest of the run.',
  defect_accepted: 'YOUR RISK-ACCEPTANCE ARGUMENT WAS ACCEPTED. Record the acceptance where the report discusses the risk; do not spend a round fixing what the bench agreed may stand.',
  repaired: 'Your fix was accepted. Stop working this one.',
  defect_owed_elsewhere: 'The finding was UPHELD and the fix is owned outside this debate. Stop trying to fix it in the report; expect the debt to be named rather than closed here.',
  moot: 'Adjudicated out of existence. Drop it.',
}
const infraDebts = [] // defect_owed_elsewhere rulings (W1.9) — the lead's named debts, surfaced at assembly and in the final envelope

// Carried-ruling persistence (run-4 §6.4 item 6 — the re-docket loop): a carried gap does
// NOT re-docket every round it stays open. It re-dockets only when red's GRADE for it
// changed (script-visible in redEnv) or a lineage successor names it — new evidence routes
// through red re-raising under a successor id, the existing lineage path. This also closes
// the carried->defect_accepted gate-erosion path (each re-docket was a fresh chance the
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
  const ledgerClause = ` THE REPORT DOES NOT NAME ITS SOURCES — it carries invisible anchors, and the evidence layer is the only way from one to what it points at. Read it FIRST. A claim blue backed with a source resolves to the url, title and sha256 of the bytes blue actually read, and THAT is what you re-fetch to audit the same artifact rather than a page that may have drifted since. A claim blue COMPUTED resolves to its script, and to whether anyone has re-run it — an unaudited proof looks exactly like a clean one until you check. Your own verifications come back ATTACHED TO THE SOURCE you checked: an empty one is a citation NOBODY has checked yet, and that is where your next pass is worth most. WHAT NOT TO RE-CHECK, because it is the round's whole economy: a claim verified at HIGH confidence in a prior round STAYS verified — do not re-fetch it UNLESS its section changed this round (read that off the recorded edits, not off blue's account of them), OR more than 2 rounds have elapsed since it was last verified, OR its access date and the source's volatility suggest drift (living documents, issue trackers, README stats). RECORD every claim you verify, naming the anchor you looked up and quoting the claim from the report. A SOURCE YOU FOUND YOURSELF IS A DIFFERENT ACT — blue never cited it, so there is no anchor to name — and it answers a different question: whether the claim is true in the WORLD, where verification asks only what blue's source did for it. VERBATIM READS ONLY — you have no WebFetch, by design: it returns a small model's SUMMARY, not the source. Read blue's cached bytes through the tool; for a source YOU discover, pull it verbatim yourself (\`curl -sL <url>\`, \`gh issue view <n> --comments\`, \`pdftotext\`/pandoc) and read it. WebSearch is for DISCOVERY — finding the url — never for the read that grades a citation. If a source is too large for your context, read it in SECTIONS and name the sections you read: A TRUNCATED READ IS NOT A READ, so state the truncation and never grade a body you could not fully read.`
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
    `Red audit, round ${round}, lens: ${lens}. RE-READ THE FULL LIVING REPORT IN CONTEXT — the whole document, never just a diff; if it exceeds one Read call, read it whole in consecutive windows${round > 1 ? `. For a navigation HINT use the record of what blue actually edited and the gap each edit answers — never a hand-written file, and never in place of the full re-read above` : ''}. ANCHOR EVERY FINDING TO A QUOTED SENTENCE, and quote it exactly rather than paraphrasing: a finding whose quote is not found in ${runDir}/blue/report.md is REJECTED. Your lens number ${role} is your ROLE and it is stable across rounds — L1-L4 citation slices, L5 logic/completeness, L6 dark-side/risk — so the record stays comparable run-wide; the labels on your findings are the tool's to assign, and the stable R${round}-N gap ids are the merge's. Only red-merge writes the round's RED narrative. HARNESS NOTES: Grep count mode counts LINES, not occurrences — anchor patterns (e.g. '^### ') when counting; prefer the Write tool over quoted heredocs for scripts (heredoc backslash mangling is a documented recurrence).${reliefFor('red')}${speedClause}${frictionClause(`red-lens-r${round}-L${role}`, 'lens')}${recordClause(`red-lens-r${round}-L${role}`)} Return a 3-line synopsis.`,
    { ...bulk, label: `red-lens-${role}-r${round} · ${slug}`, phase: 'Red', agentType: 'frank-exchange-of-views:red-auditor' })))

  redEnv = await agent(
    `Red merge, round ${round}. You are the board's only writer, and the seat that decides whether this report has been verified. The lenses brought findings; nothing is on the board until you put it there.

COALESCE — DO NOT TRANSCRIBE. Collapse this round's findings into the DISTINCT PROBLEM-CLASSES they represent. Several findings that are the same defect at different sites are ONE gap. Separate genuine incidents; fold duplicates and near-duplicates together; raise gaps SPARINGLY and deliberately, one considered gap at a time. A merge that raises a gap per finding has not merged — it has transcribed, and it floods the docket with noise. Screen every candidate against the board first: a defect you already closed is a REOPEN, and it arrives carrying that history or it arrives lying.

ASK FOR THE ANSWER TO BE PRODUCED, NOT ASSERTED. Across six recorded runs no seat ever wrote a program to settle anything, and the 2026-08-05 smoke shows the cause was not blue's reluctance but yours: all TEN of your acceptance checks asked only whether the report SAYS something. Where a claim is arithmetic, an enumeration, or a reproducible measurement, demand that it be computed. Where it turns on what the document states or what a source says, say so plainly — there is no credit for inflating it. And say WHEN your demand can be discharged, in the two classes blue answers in: a DOCUMENT-PROBE is executable now against shipped artifacts, while a LIVE-PROBE needs built artifacts and is DEFERRABLE in a design-phase debate, discharged by naming it as an acceptance test with its pass condition. An unclassed demand invites blue to overclaim a file read as a probe, or to stall on an impossible obligation.

BELIEVE NO BYTES YOU DID NOT WATCH BEING PRODUCED. Where blue backed a sentence with a computation, RE-RUN IT — this is the one audit that does not end in you believing bytes someone else chose. Then READ THE SCRIPT and say what it ACTUALLY COMPUTES. It does not have to be pretty or idiomatic; it has to compute the thing it claims, readably enough that you can see that it does. A script that re-runs clean and establishes nothing is the dangerous case, because it looks maximally credible.

A CLOSURE IS A CLAIM, AND CLAIMS DECAY. Verify any lineage or closure claim you assert against the record BEFORE you assert it, and re-sample the archive every round it is not empty. Reopen any sampled closure whose evidence has drifted; a closure resting on a volatile living source inherits that source's drift triggers. Nothing FAILS the run over a skipped sample — the cost is a report that shows a reader you skipped it, beside the rounds that looked. That is deliberate: this duty exists because a round-2 spot-check once degraded into the seat attesting blocks it was about to write itself.

VOTE EVERY LINE OF INQUIRY THIS ROUND, ON ONE READ. A line reaches the report as a generated row carrying no citation anchor, so "we pursued X", "we abandoned Y" is the one class of claim in the document your side otherwise has no channel to answer. Read the report ONCE and answer every line against that pass, the way anyone checks a document against a list — NOT once per line: a dozen lines over four rounds would be forty-eight reads of the same artifact, and a duty that costs that much is one you will route around, which is worse than no duty because the record then says checked. Answer from THIS READ, not from what you remember voting. PER ROUND, because the report is rewritten every round and a vote cast before this round's edits answers a question about a document that has since changed.

RULE ON BLUE'S DIRECTIONS. Blue proposes lines of inquiry with a hypothesis; you say which are worth this run's time. A ruling is an ARGUMENT, not a command: it needs a reason, and blue may appeal it — which puts the disagreement on the record whether or not blue also pursues the line, so arguing and then yielding is visible instead of vanishing. Measured: blue rejected 18 of its own 86 lines across six runs and red rejected NONE, because red had no way to. You do not propose directions yourself — when you want research done, that is a gap, graded and tracked and closed.

PRESCRIBE TEXT ONLY WHERE THE DEFECT IS TEXTUAL. The required fix is prose — what must become true — and stays the channel for substantive work: research it, enumerate it, qualify it. For an overclaim, a wrong figure, a contradiction, you may also state exactly what the span should become, and you cannot state one without having read the text you are prescribing against — which is the check whose absence caused all three of the 2026-08-04 smoke's round-2 gaps. Anything larger than a correction is a substantive ADDITION and it is BLUE'S to author, not yours.

DID BLUE ACTUALLY DO WHAT YOU ASKED? Put your prescription and blue's edits side by side rather than inferring the answer — and note that this does not replace re-reading the report, because a span out of context misleads on research prose. Two shapes are worth naming: NOTHING recorded against an open gap, and a change that does not meet your own acceptance check. Where blue applied your exact text, your prescription has been satisfied at that site by construction: close the gap unless you can state what your own check still fails. 3 of 3 round-2 gaps in the 2026-08-04 smoke were text blue added in round 1 because red asked for it.

LINEAGE IS NEVER DROPPED. A gap keeps its id across rounds; a successor names its ancestors; a closure that regressed says so, and the docket follows those chains. Where a defect turns up BETWEEN two repairs that each closed clean in an earlier round, it AMENDS both rather than arriving as this round's fresh closure — a late-discovered composition defect and a this-round closure are different events and the record must be able to tell them apart.

THE STOPPING JUDGMENT IS YOURS, AND IT IS NOT CEREMONY. PASS only when every remaining unadjudicated gap is repaired, not_a_defect, or defect_accepted. Your recorded verdict is the ONE fact distinguishing "red passed" from "the bench closed the last gaps at the terminal sitting" — without it the run cannot say, from its own record, that it was ever verified.${adjudicated.length ? ` GAPS THE BENCH HAS ALREADY RULED, WITH THEIR FATES AND WHAT EACH ONE SETTLED — excluded from your verdict, and the exclusion is ESTOPPEL, not amnesia: ${JSON.stringify(adjudicated.map(x => ({ gap_id: x.gap_id, resolution: x.resolution, settled: x.settled, reopens_on: x.reopens_on, final: x.final })))}. THE BARRED PROPOSITION IS NARROWER THAN THE GAP: what you may no longer assert is that sentence, not everything the finding contained. You were previously handed these as bare ids, so the bar was enforced by making them invisible — you could not tell a ruling you should respect from one you had simply lost track of, and you could not tell relitigating from a legitimate successor. The ruling STANDS and you do not re-raise it. If you hold genuinely new evidence the bench did not have, that is a lineage successor: mint it under a new id naming the ruled gap in supersedes, and say what the ruling did not account for. THE REASONING IS ON THE RECORD, NOT IN THIS PROMPT — read the bench's opinion for any fate you are about to rely on or work around, rather than inferring it from the word. A fate you never read is one you cannot honour or contest.` : ''}${gradeAdjustments.length ? ` GRADE ADJUSTMENTS RULED BY THE JUDGE last round — apply each, and list the delta in your round narrative: ${JSON.stringify(gradeAdjustments)}.` : ''}${pendingDisputes.length ? ` BLUE'S GRADE DISPUTES from last round (ROUTING REFS — blue's evidence is on the record, not here: read each dispute's argument before answering): ${JSON.stringify(pendingDisputes)}. You MUST answer EVERY one; an unaddressed dispute is treated as rejected and auto-docketed to the judge. Answer on the MOTION's own id, because blue may contest more than one grade on the same gap and an answer naming only the gap cannot be matched to the one it refuses. ACCEPTING A DISPUTE DOES NOT MOVE THE GRADE — SAYING SO IS NOT DOING IT: move it, on the axis that moved, with what changed your mind. A grade that moves with no recorded reason reads as though blue's dispute was answered by silence. AND list each in the envelope's dispute_responses as a ROUTING REF ONLY (gap_id, dimension, response — no prose: the rationale is on the record) so the docket routes it, and list each accepted delta (gap id, dimension, old -> new) in your round narrative, where blue, the judge and the operator watch for it.` : ''}

YOUR NARRATIVE IS YOUR ARGUMENT and the other side answers it. Where a gap is docket-bound — one you RE-RAISE from a prior round, a successor you mint, a dispute you REJECT — argue it in ~120 words: your strongest evidence the gap is real and graded correctly, and your answer to blue's best rebuttal so far. The judge rules after blue responds, and overstatement the record does not support counts against you.${petitionClause(`red-merge-r${round}`)}${reliefFor('red')}${frictionClause(`red-merge-r${round}`, 'merge')}${speedClause}${recordClause(`red-merge-r${round}`)} Return the red envelope.`,
    { ...judgment, label: `red-merge-r${round} · ${slug}`, phase: 'Red', agentType: 'frank-exchange-of-views:red-auditor', schema: RED_ENVELOPE })

  takeFriction(`red-merge-r${round}`, redEnv)
  if (!redEnv) throw new Error(`red-merge round ${round} returned null (agent failed) — aborting cleanly`)
  log(`round ${round}: red ${redEnv.verdict} — ${redEnv.gaps.length} gaps open, mass ${redEnv.gaps.reduce((s, g) => s + gapMass(g), 0).toFixed(1)}, ${redEnv.citations_checked} citations checked`)
  // Degenerate-shape guard (retrospective §3 row 20, decided R4-2: throw, never soft-convert):
  // FAIL with zero gaps is evidence of a broken merge, not a clean report — looping on it
  // burns maxRounds silently and returns a self-contradictory UNVERIFIED/0-gaps verdict.
  //
  // A PETITION IS THE ONE HONEST WAY TO FAIL WITH A CLEAN BOARD, and this guard predates the
  // tool gate that made it reachable. `verdict --as PASS` is refused while any motion is
  // unruled; a PETITION is the BENCH's to rule, not the merge's, and the merge files one in the
  // same envelope that carries its verdict — so the bench has not sat yet. The seat cannot pass
  // (the tool refuses) and could not fail either (this threw), which is no legal verdict at all.
  // Measured: 4 of 60 fuzz runs died here, and the merge was behaving correctly in every one.
  //
  // Scoped to the envelope's own petitions deliberately. A petition BLUE filed dispatches a
  // bench sitting before the next scheduled seat, so it is already ruled by the time red sits;
  // what this exempts is the case red itself created, which is the case red cannot resolve.
  const petitioned = Array.isArray(redEnv.petitions) && redEnv.petitions.length > 0
  if (redEnv.verdict === 'FAIL' && redEnv.gaps.length === 0 && !petitioned) {
    throw new Error(`red-merge round ${round} returned FAIL with an empty gaps array — degenerate merge, refusing to loop silently`)
  }
  if (redEnv.verdict === 'FAIL' && redEnv.gaps.length === 0) {
    log(`round ${round}: red FAILed on a clean board — ${redEnv.petitions.length} petition(s) outstanding for the bench, which is not a defect on the board`)
  }
  // Lineage enforcement (row 23 step 4, per red's own R5-5 critique: self-declared lineage
  // is hollow unless structurally checked): every repaired_with_regression closure must have
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
    if (c.class === 'repaired_with_regression' && !redEnv.gaps.some(g => (g.supersedes || []).includes(c.id))) {
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
  // traffic — closed/not_a_defect are structurally dead for it and the judge is told
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
  // NOTHING IS WITHHELD FROM BLUE SILENTLY EITHER — the same rule the bench's withheld list
  // follows below, applied to the seat that was still getting the subtraction.
  const blueRulings = adjudicated.map(x => ({
    gap_id: x.gap_id,
    resolution: x.resolution,
    settled: x.settled,
    reopens_on: x.reopens_on,
    final: x.final,
    your_duty: BLUE_DUTY_BY_RESOLUTION[x.resolution]
      || `this fate has no standing instruction in the protocol — read the bench's opinion on the record and report the omission as friction rather than guessing at it`,
  }))
  blueEnv = await agent(
    `Blue response, round ${round}, topic "${topic}". Red's verdict: FAIL.${reliefFor('blue')} YOUR FIRST READ COMES AFTER THE TREE WALK BELOW, NOT BEFORE IT: pull your working set — the board, the transcript, and ${runDir}/inputs/red-gap-patterns.md — in one pass rather than three, concatenating them into a single file under your session scratchpad (an ABSOLUTE path, never under ${runDir}) and reading that.

YOU MAY COMPUTE AN ANSWER, NOT ONLY COLLATE SOURCES. You have Bash, Write and Edit, and for a whole class of questions running something settles it faster and harder than arguing about it. WORKING IT OUT IN YOUR HEAD IS NOT THE EXCEPTION — IT IS THE CASE THIS EXISTS FOR. Measured: a seat facing a gap that said the stated total did not match the per-source rows summed twelve integers inside its own reasoning, wrote the answer into the report, and moved on. The answer was RIGHT, which is exactly why nothing caught it: a correct figure with no derivation is indistinguishable from a confident guess, and a reader cannot vary the rate, re-check the sum, or locate the error on the day there is one. A gap on the board can be WAITING ON A PROGRAM FROM YOU — it says so, and where it does, no amount of prose will close it. THIS IS THE CITATION MODEL WITH THE LAST MILE WALKED, not a new kind of evidence: cite the METHOD (trial division, Miller-Rabin, the formula, the standard) and then prove the INSTANCE by running it. Several methods on one claim is several anchors on one sentence, which is how a mathematician argues. Measured, so you know why this clause exists: red once asked for the protocol to be tested on a false claim, blue answered in PROSE that the test had happened, and red correctly refused it for showing no evidence — a whole round for something three lines settle.

YOUR LINES OF INQUIRY ARE A LIVING RECORD, NOT A ROUND-0 PLAN. Every round, revisit what is still open and say what became of it. The hypothesis is what makes that honest — a line abandoned against its own stated claim is evidence of choosing; one abandoned on a shrug is not. Measured over six runs: 83 of 86 lines were declared in round 0 and NOT ONE was ever revisited, so the record showed a plan and never showed you choosing. Red rules on your proposals and you may APPEAL a ruling — appeal whether or not you go on to pursue the line, because the appeal is where your ARGUMENT is recorded and the fate is only what you decided to do about it.

WHERE RED PROPOSED EXACT TEXT you have THREE paths and you are not obliged to take the first: apply it verbatim, counter-edit with your own fix, or dispute it. Applying is the cheapest — a counter-edit costs a round and invites re-audit, a dispute needs red's agreement — so say plainly which path you took and why. A round in which you never decline is not agreement, it is capitulation, and it IS measured: offered/applied/declined is printed every round, and a zero decline rate is read as a question about red's proposals, not as evidence they were all right. Applying red's exact text also ESTOPS red from re-raising that text as a fresh gap, so verbatim application is a real settlement, not a surrender.

OWNERSHIP BINDS, AS IT DID IN ROUND 0: you write ONLY your own surfaces — TL;DR, Catechism, technical foundations, analysis, open questions, your citations. You MUST NOT introduce a \`**Verdict:**\` line or any tool-owned section (\`## Risk matrix\`, \`## Red team findings\`, \`## The debate\`, \`## How this run was conducted\`, a \`## Blue team report\` wrapper) while repairing: assembly composes those from the record, you cannot know the outcome or red's findings, and anything you add there only lands stale and gets stripped. AND WHICH MODEL IS ANSWERING THIS RUN IS NOT A FACT YOU HOLD: you can see what was REQUESTED, never what replied, and a run configured for one tier is routinely served by another — 44 of 63 seats across two runs, whose certified report reasoned two paragraphs about vendor diversity from the pairing it had asked for. If your argument turns on the models used, say that the record's measurement decides it; assembly composes that section from what actually answered.

PRE-FLIGHT: re-check your planned repairs against red's gap patterns. Open gaps (adjudicated ones excluded): ${JSON.stringify(openGaps)}.${blueRulings.length ? ` GAPS THE BENCH HAS RULED, AND WHAT EACH RULING OBLIGES OF YOU: ${JSON.stringify(blueRulings)}. These are excluded from the open list above, and the exclusion is a RULING, not an oversight. You were previously handed it as a bare SUBTRACTION — a ruled gap simply stopped appearing — and a vindication and an upheld finding are the same absence, so you could not tell a fate you had WON from one you had merely lost sight of. THE BAR IS ON THE PROPOSITION, NOT THE GAP: what you may no longer re-argue in the report is the sentence under settled, not everything the finding touched. WHERE A RULING WENT YOUR WAY IT IS YOURS TO INVOKE — say so on the record and rely on it, rather than quietly re-fixing text the bench has already blessed. A ruling marked final does not reopen; one carrying a reopens_on condition reopens only on that condition and not on a re-reading. THE REASONING IS ON THE RECORD, NOT IN THIS PROMPT — read the bench's opinion for any fate you are about to rely on or work around, because a fate you never read is one you can neither honour nor invoke.` : ''} GRADING SEMANTICS (mapping ${MASS_MAPPING_VERSION}, and NOT inferable from the JSON above — the map is self-describing as structure and silent on meaning): likelihood grades the CONSEQUENCE ONLY — how likely the harm is to land, never how likely the defect is to BE there. A typo you have confirmed, whose harm is trivial, is a LOW likelihood; reading likelihood as confidence-that-it-exists is the v1 semantics this mapping replaced. A claim red graded LOW is where your evidence is thinnest, and it is worth strengthening before red returns to it. Red's corroborations are on the RECORD, not in this prompt: read what red actually verified and at what confidence, rather than a hand-copied snapshot of one round.

BEFORE drafting, read the transcript — the gap JSON above is a lossy summary of it, and it carries any accepted grade-dispute deltas pending their contest window, red's latest narrative, and the bench's latest resolutions. Any gap the judge CARRIED comes with a stated research direction you owe.

ANSWER EVERY OPEN GAP ADDITIVELY in ${runDir}/blue/report.md — expand and repair where red is right, REBUT IN WRITING WITH EVIDENCE where red is wrong, and argue risk-acceptance where the fix's complexity exceeds its likelihood x impact. Compact and reorganize prose as clarity demands — but a CLAIM leaves only by being retired on the record, which names it, why it goes, and what replaces it. Losing one in a rewrite is the failure the additive rule exists to prevent. PROPAGATE EVERY CORRECTION TO ALL SITES that state the corrected claim, not only the flagged sentence — incomplete propagation was run 3's dominant blue failure class, 5 regressions in 5 rounds — and remember that an index of footnoted claims cannot see an unfootnoted restatement of one, so a bare corrected FIGURE needs a report-wide sweep of its own. Log the sites you checked in your round record.${contested.length ? ` CLOSING ARGUMENTS: the following are DOCKETED for adjudication AFTER your response this round: ${JSON.stringify(contested)}. For EACH, after your repairs, argue in ~120 words why your response resolves it, or why red's grade or claim is wrong, citing the exact section and evidence. This is your case; material not in the record cannot help you.` : ''}

DISPUTE RED'S GRADING WHERE YOU DISAGREE WITH IT — on the axis, with the grade you say it should be, and with your evidence. If red rejects it and you still disagree, press it to the bench. List each in the envelope's grade_disputes as a ROUTING REF ONLY (gap_id, dimension, proposed — no prose: the evidence lives on the record, not a second copy in the envelope) so the docket routes it — max ${DISPUTE_CAP} per round, and overflow is batch-docketed to the judge as one item${heldDisputes.size ? `; disputes red REJECTED last round (re-dispute any of these to send it to the judge): ${JSON.stringify([...heldDisputes.values()])}` : ''}. Where a fix demands a probe, discharge a DOCUMENT-PROBE now against shipped artifacts, or for a LIVE-PROBE needing built artifacts name it as a deferred acceptance test with its pass condition — never call a file read a live probe.

AUDIT YOUR OWN REPAIRS, ONE RECEIPT PER GAP (W2b; your constitution carries the full standard): figures recomputed, universals enumerated, consistency sites swept report-wide, the repair's own boundary case asked, compositions noted where edits share text, sibling sweep done or the enumeration declared open, the gap's acceptance check RUN with its result, new claims tagged verified-at-leaf/derived/asserted. Compress what you checked and what it showed into each receipt, and list just those gap IDS in the envelope's manifest array as a routing ref. An unmanifested repair is unaudited by your own standard; the script rejects an EMPTY manifest array on a round with open gaps. Recompute claim_count with the tool after your edits land and relay the integer into the envelope; never hand-count it. ROUND RECORD (W1.7): round_record_appended = TRUE only after BOTH your round narrative AND your round-${round} revision exist. RECORD THE ROUND. On false the script re-prompts you once and, failing that, logs friction and continues — the run is not discarded, but the parity gap is scored against you at capture; a revision is not on the record until the transcript carries it (the round-2 desync misled a lens and blinded the judge, run 5).${patternDutyClause(openGaps)}${inquiryClause}${petitionClause(`blue-respond-r${round}`)}${frictionClause(`blue-respond-r${round}`, 'blue')}${speedClause}${recordClause(`blue-respond-r${round}`)} Return the blue envelope.`,
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
      `Adjudication, round ${round}, topic "${topic}". Contested docket: ${JSON.stringify(contested)}.${withheld.length ? ` OPEN GAPS WITHHELD FROM THE DOCKET (with the reason each was withheld): ${JSON.stringify(withheld)}. These are NOT docketed, and you are not obliged to rule on them — but a gap BOTH SIDES argued to closing that reached no ruling is a docket defect, not a decision. If you find one, say so and rule on it: an unadjudicated gap returns to red's verdict pool as though nobody had considered it.` : ''} THE DOCKET IS A ROUTING LIST, NOT THE EVIDENCE. It carries ids and grades; a gap's problem text and its acceptance check live on the board, and you read them FRESH before ruling. In the 2026-07-18 run the bench mis-ruled on snapshotted premises that were false by the time it sat. Re-run each document-probe acceptance check against the artifact AS IT NOW STANDS, and rule on what you find rather than on what any snapshot asserts.

YOUR RULING BASIS IS CONFINED TO THREE THINGS: the two sides' closings, the full transcript, and the final state of the artifacts — the board and ${runDir}/blue/report.md as they now stand. Weigh each closing as that side's best case, and a claim in a closing that the record does not support counts AGAINST the side that made it. For every ruling on a gap with a lineage chain, READ THE NAMED ANCESTORS' RECORDS first and NAME what you read in your rationale.${lawClause}${declareClause}

Every docketed gap gets a written ruling: its fate, the principle you applied, the values in tension, whether a human should look at it, your reasoning, and TWO THINGS THE FATE CANNOT SAY — the proposition you are barring, and what would reopen it. Only you know either; the verb's help says what each is for. A bare fate teaches the next round nothing. Two fates are worth naming because they route work OUT of the debate rather than ending it: a gap you CARRY stays live and owes blue a stated research direction, and a valid finding whose FIX is owned outside the debate — run tooling, the harness, the lead — leaves red's verdict pool and ships as a NAMED infrastructure debt, recorded and never dropped. deadlock is true only if no gap is carried AND ${hasNew ? 'false (new gaps were raised this round)' : 'no new gaps were raised this round (none were)'}. AND DEADLOCK IS A FACT ABOUT THE PARTIES, NOT ABOUT YOUR OWN PRODUCTIVITY: if disposing this docket leaves NOTHING open they did not deadlock, they converged in your hands, and red — who owns PASS/FAIL — has not yet passed an empty docket. Are the parties stuck, or did you just finish the work? Say deadlock only for the first.${frictionClause(`judge-r${round}`, 'bench')}${speedClause}${recordClause(`judge-r${round}`)} Return the judge envelope.`,
      { ...judgment, label: `judge-r${round} · ${slug}`, phase: 'Debate', agentType: 'frank-exchange-of-views:lead-judge', schema: JUDGE_ENVELOPE })
    if (!judge) throw new Error(`judge round ${round} returned null (agent failed) — aborting cleanly`)
    for (const h of judge.holdings || []) holdingsInEffect.push(h)
    for (const r of judge.resolutions) {
      if (r.resolution === 'repaired' || r.resolution === 'not_a_defect' || r.resolution === 'defect_accepted' || r.resolution === 'moot' || r.resolution === 'defect_owed_elsewhere') adjudicated.push(r)
      if (r.resolution === 'defect_owed_elsewhere') infraDebts.push({ gap_id: r.gap_id, owed_fix: r.rationale, round })
      if (r.resolution === 'carried') {
        const g = redEnv.gaps.find(x => x.id === r.gap_id)
        if (g) carriedRulings.set(r.gap_id, gradeSnapshot(g))
      }
      if (r.resolution === 'grade_adjusted') gradeAdjustments.push({ gap_id: r.gap_id, rationale: r.rationale })
    }
    takeFriction(`judge-r${round}`, judge)

    // DEADLOCK AND A CLEARED BOARD ARE DIFFERENT STATES, and they shared a stamp.
    //
    // MEASURED 2026-08-22, in the run that found it: red's round-2 merge refused PASS with two
    // gaps open ("Two gaps remain open; PASS is refused this round"); the bench then ruled BOTH
    // in the same sitting (one closed, one defect_owed_elsewhere), clearing the board to
    // zero; `break` fired here; the run stamped UNVERIFIED with ZERO gaps outstanding. The bench
    // wrote the defect into its own `certify` rather than banking the stamp: "'red owns
    // PASS/FAIL' is a stated tiebreaker this run's sequencing arguably sidesteps by letting the
    // bench's own docket-clearing act substitute for red's affirmative call."
    //
    // Deadlock means THE TWO SIDES CANNOT CONVERGE. What happened is the opposite — everything
    // converged, in the bench's own hands. UNVERIFIED-with-nothing-open is the same
    // self-contradictory shape the guard above already refuses when red reports FAIL with an
    // empty gap list; it just arrived by the other road.
    //
    // So: ask the board. Gaps still open => the bench found genuine deadlock, terminate. Board
    // cleared => red is owed one sitting to verdict against it, because red owns PASS/FAIL.
    //
    // BOUNDED TWICE, because "one more round" is exactly the shape that runs forever: maxRounds
    // still caps the loop, and the relief is ONE-SHOT. If red refuses PASS again against a board
    // the bench cleared again, that IS irreducible disagreement and deadlock stands. The bench
    // cannot buy unlimited rounds by clearing the docket each time.
    if (judge.deadlock) {
      const ruled = new Set(adjudicated.map(x => x.gap_id))
      const stillOpen = redEnv.gaps.filter(g => !ruled.has(g.id))
      if (stillOpen.length === 0 && !benchClearedOnce && round < maxRounds) {
        benchClearedOnce = true
        log(`round ${round}: the bench cleared the board to zero — NOT deadlock. Red owns PASS/FAIL and has not verdicted against an empty docket; granting one further round so it can. This relief fires once.`)
      } else {
        deadlocked = true
        break
      }
    }
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
    `Terminal dispute disposition for topic "${topic}" (debate ended ${verdict} after round ${round}; this docket fires at the exit boundary). Undisposed grade disputes: ${JSON.stringify(terminalDisputes)}. Read the transcript in full, and read the board back, before you rule. ${lawClause}${declareClause}${inspectionClause} NOTHING CAN BE CARRIED AT A TERMINAL EXIT — there is no next round to carry it into. Each dispute either moves to a corrected grade, which you state, or it ships CONTESTED and the report records it as contested. deadlock: false.${frictionClause('judge-terminal', 'bench')}${speedClause}${recordClause('judge-terminal')} Return the judge envelope.`,
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
  `Final assembly for topic "${topic}", run directory ${runDir}. Debate outcome: ${verdict} after ${round} round(s)${deadlocked ? ' by judged deadlock' : ''}${exhausted ? ' by safety ceiling' : ''}. THE REPORT IS ASSEMBLED FROM THE RECORD — you author NOTHING, you copy NOTHING, you fill in NO inputs. There are no <FILL> fields and no sections for you to write; do not hand-write report.md and do not copy anything into it yourself.

FIRST, STAMP HOW THIS RUN ENDED: it is ${verdict}${deadlocked ? ', ended by judged DEADLOCK' : exhausted ? ', ended against the CEILING' : ''}, and your account of the sitting goes with it — where a run ended by judged deadlock that account is the only evidence the determination will ever have.

THEN, TWO THINGS YOU MAY HOLD AND THIS IS YOUR LAST CHANCE TO RECORD EITHER. If you hold something that binds how the RECORD IS READ but moves no gap — a construction of a term, a correction of what the record MEANS rather than what it says, a holding worth offering as precedent — state it, and state it in its own right rather than folding it into an unrelated rationale. And if anything in this run needs A HUMAN to re-examine it — an unresolved tension, a claim that held only because nobody could reach the source, a boundary you ruled close to — say so. You keep no memory between runs, so this is the whole of your continuity.

THEN ASSEMBLE, and verify ${runDir}/report.md exists and reads correctly at the top: the verdict stamp is the outcome's, the sections are blue's and the record's. A tool cannot mis-author a synthesis surface — the run-5 catechism defect, where 6/7 answers regressed once assembly authored the catechism, is structurally impossible now, and the TL;DR is blue's, inside the audited report, rather than written after red's final audit where nothing checks it. THE AUTHORITATIVE OPEN COUNT IS THE BOARD'S, after every closure and ruling: read it back and report it as open_gaps in your envelope. Collated friction so far (report any of your own as well): ${JSON.stringify(friction)}.${frictionClause('assemble', 'bench')}${speedClause}${recordClause('assemble')} Return your envelope: a 5-line synopsis, open_gaps from the board, and your own friction if any.`,
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

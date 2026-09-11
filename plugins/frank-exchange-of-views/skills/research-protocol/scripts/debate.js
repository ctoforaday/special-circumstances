export const meta = {
  name: 'frank-exchange-of-views',
  description: 'Adversarial research debate: additive blue builds, gate-keeping red audits, judged termination, union-not-summary assembly',
  phases: [
    { title: 'Frontier', detail: 'hypotheses before searches' },
    { title: 'Blue', detail: 'best-of-N lanes + additive synthesis' },
    { title: 'Red', detail: 'per-lens audits + merged verdict' },
    { title: 'Debate', detail: 'sittings dispatched from the record until nobody is ready — PASS permitted, every open gap at its limit, or the epoch limit reached — or until the plan stops moving' },
    { title: 'Assemble', detail: 'final report by union' },
  ],
}

// TESTING & MODEL-SELECTION STRATEGY (learned from runs 1-3, ~5M tokens of tuition):
//   Logic bugs are unit-testable for zero tokens: a Node harness stubs agent() with canned
//   envelopes and drives every branch (args parsing, dispatch loop, contested docket, impasse,
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
const { topic, runDir, lanes = 3, lensAreas = null, model = null, judgmentModel = null, laneFloorOverride = null, binDir = null, scorecards = null, gapPatterns = null, transcriptDir = null } = a
if (args && Object.prototype.hasOwnProperty.call(args, 'maxRounds')) {
  throw new Error(`debate: refusing dispatch — maxRounds is not a term of this engine any more (plans/roundless.md). A run ends on the record, not on a count:
the chair's \`dispatch next\` says who sits, and the run ends when nobody is ready — PASS permitted, or every open material
gap at its limit (CEILING) — or at the run's epoch limit (CEILING). The bounds are the run's TERMS, recorded by setup in
inputs/run-config.json: the exchanges a gap gets before impasse (k-max), the floor of the gaps a lens may mint (mint-budget),
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
const bulk = { model }
const judgment = { model: judgmentModel }
const slug = String(runDir).replace(/[\/]+$/, '').split(/[\/]/).pop().replace(/^[0-9-]+_/, '')
log(`researching: ${topic.length > 160 ? topic.slice(0, 157) + '...' : topic}`)
log(`resolved tiers — bulk: ${model}, judgment: ${judgmentModel}`)

// Friction must survive a mid-run throw (retrospective §3 row 24): the envelope copy feeds
// this script's aggregate, the file copy survives an abort. Both, always.
const frictionClause = (who, role) => ` LOG (${who}) — CLOSE THIS CHANNEL BEFORE YOU FINISH: on the record, AND in the envelope's log field. YOUR AUDIENCE IS THE OPERATOR WHO CAN RETOOL YOU, not the other seats: what you reached for, what it did, and what you wanted instead. Where you set an act aside as a JUDGEMENT rather than for want of occasion, give that one sentence — the record shows what you ran, never what you weighed and declined. Where nothing impeded the work, say that instead.`

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
// W2h — the visibility loop's last leg: each scorecard's headline numbers, IN
// the prompt of the seats it measures.
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
  ? ` LINES OF INQUIRY: record every genuinely distinct approach you CONSIDERED, not only the one you took — with the hypothesis that would be true if it paid off, and later with what became of it and what changed your mind. The hypothesis is what makes the account honest, and the dead ends matter most: one you record is one the next run does not re-walk, and it is worth more than the tidy conclusion that survived. Alternatives reach the reader under their own headings — what you pursued, what you left for a later run, and what you weighed and rejected are three things a reader needs, not one list of winners. The line you propose and the reason you give when you move it ARE PRINTED IN THE REPORT word for word, so write them as research prose for a reader of the SUBJECT — the question the direction asks and what was found about it, never the steps you took. Recording what you considered and declined is what makes "explore the alternatives" checkable instead of self-attested.`
  : ''


 // STEELMAN DUTY (E0.5h): the sections red NEVER gap-anchors are exactly the
// self-critical ones — disconfirming passes, human-gated paths, self-attested
// inventories. Claims AGAINST the design attract no adversary, so the case
// against goes unaudited while the case for is contested every round. Blue's
// recorded lines of inquiry are that surface made concrete: what was declined,
// and whether the stated reason survives contact.
const steelmanClause = ` STEELMAN DUTY: read the exploration space via the \`lines-of-inquiry\` projection (the tool renders it fresh from the line of inquiry events on the record). Blue's DECLINED and ABANDONED lines of inquiry are the case AGAINST its own design, and the measured blind spot is that nobody audits them — a weak reason for declining a strong alternative survives untouched because it reads as humility. Attack the reasons, not just the conclusions: a declined line of inquiry whose stated reason does not hold is a finding, and so is an abandoned one whose obituary is wrong. RECONCILE THE RECORD AGAINST THE DOCUMENT: the directions the report actually took must be the ones on the line of inquiry record. A section built on a line of inquiry that was never proposed is undeclared scope; a line of inquiry still at \`proposed\`, or at \`pursued\` and NOT moved this sitting, is a decision nobody made. Both are findings. Re-recording \`pursued\` WITH what it learned is a legitimate reaffirmation and settles the line for that sitting — do not read it as neglect; \`deferred\` is likewise a decision (kept for a later run), not an omission. Measured over six runs: 83 of 86 lines of inquiry were declared in the first epoch and NONE was ever revisited, so 'pursued' meant 'I intend to' and nothing could falsify it.`

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
  // A seat reads its OWN in-run scorecard for THIS question, computed live from this run's
  // record. Never a prior run's numbers: those are Goodhart bait, topic-confounded, cross-model
  // and salience-priming. The `scorecards` arg feeds operator analytics only.
  //
  // THIS CLAUSE NAMES THE ACT, AND THE HELP NAMES THE VERB. promptverbs' catalogue gate pins
  // debate.js at zero named commands, on the standing rule that the help page is the only page
  // that instructs — so a prompt that spells an invocation is a prompt teaching a surface it does
  // not own. It also names no card: the tool resolves that from the seat's own registration,
  // which is why the read needs no selector and why there is no way to ask for another party's
  // numbers. See #513 for the friction that established both.
  ` YOUR IN-RUN SCORECARD (THIS run, not a prior one): before you read the open docket, read YOUR SCORECARD for the question in front of you so far. It is a projection of this run's record on your own surface, it needs no selector because your scorecard is the one your registered seat is measured on, and its own page says what the rows mean.`
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

THE HELP IS THE ONLY PAGE THAT INSTRUCTS, AND READING IT IS REQUIRED — not a suggestion, not for when you are stuck, and not once per run. READ YOUR WHOLE SURFACE BEFORE YOU CHOOSE, NOT THE ONE PAGE FOR THE VERB YOU ALREADY PICKED. Do it ONCE, IMMEDIATELY AFTER \`register\` — before you have decided what to do, because deciding first is what makes the reading pointless:

  "${binDir}/feov-record" --seat-id ${seatId} manual — every command on your surface, each under a header naming it and followed by that command's own help, run live. It is long: send it to a file under your session scratchpad (an ABSOLUTE path, never under the run directory) and read that file whole — in consecutive windows if Read refuses it at once.

A single command's own page is still there when you want to re-check one: "${binDir}/feov-record" --seat-id ${seatId} <command> --help.

MEASURED, WHICH IS WHY IT IS ONE CALL, FIRST. Across nine sittings seats opened 6 of 51 group pages — twelve per cent — and 90% of the commands they ran were run without ever reading that command's own page. The rule they were following said to open a group "before using any command in it", so a seat that obeyed it perfectly still only ever opened the page for a verb it had already chosen: eighteen of the twenty-three pages opened were for verbs the seat went on to run. That is not surveying a surface, it is confirming a decision — and a decision made before reading is made from memory and from this prompt's vocabulary, which is not authoritative. Walking the surface a page at a time then cost 142 help turns across 26 sittings; the manual is the same pages in one.

A NAME YOU DID NOT READ IN THE HELP THIS SITTING IS A GUESS. Do not work from memory, do not carry a name from a previous sitting, and do not assume a command is named after the thing it writes. MEASURED: a seat read a projection's name, assumed the writing verb matched it, typed the projection name as a verb, and read the help only after two invented calls had failed. The projection names, this prompt's words for a concept, and the command that writes it are three different vocabularies and they do not always agree. The help is the only authoritative one, and this prompt names the JOB, never the flags.

EVERY READ IS A PROJECTION OF THE RECORD, never a file you open. The .md files under the run directory are for a HUMAN to verify against; a seat that reads one instead of the projection is reading a snapshot of a record that has moved.

YOUR REASONING IS PART OF EVERY ACT, and the RECORD renders it — your position and closings in the transcript, your reasons beside the acts on the board. It is your argument to the other seats, where they can answer it. NEVER the research report: that document is addressed to a reader of the subject and carries no argument about the run. And write your THINKING, not a label for what you did — the record already holds every act of this sitting, in order, so an account of the verbs you ran narrates what the record reconstructs. An act you got wrong is corrected by the command that recorded it, before another seat acts.`


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
// BEGIN GENERATED MASS — `cd scripts && go run ./massgen`. Source: record.proto Grade (mass).
// A grade's weight is a property of the WORD, annotated on Grade and carried by
// enum_grade.mass; this table is that annotation, rendered for the engine. Editing it here
// is editing a copy — change record.proto and regenerate.
const MASS = { trivial: 0.5, low: 1, 'low_medium': 1.5, medium: 2, 'medium_high': 2.5, high: 3, certain: 3.5, realized: 0 }
// END GENERATED MASS
const gapMass = (g) => (MASS[g.likelihood] ?? 0) * (MASS[g.impact] ?? 0)

// Grade-dispute channel constants (run-4 report §3.3, clauses (v) and (vii)):
// per-round dispute cap with overflow batch-docketed as ONE judge item, and the
// script-computed cumulative accepted-delta magnitude (in mapping units) that
// batch-dockets accepted deflation/inflation for judge review before it stands.

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
    log: { type: 'array', items: { type: 'string' } },
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
// and no way to state it: a docket ruling requires a motion id and a fate-changing --as, so it
// disposes of a docketed gap and nothing else. The bench put its construction in a petition
// ruling's opinion text, where red never read it.
//
// The verb shipped (0.67.0) and renders under "### LEAD"; the harvest reads it (#413). The prompt
// carried a paragraph naming it, because for two releases no prompt and no constitution did and
// the bench could not know it existed. The paragraph is gone: `declare --help` now carries what it
// is for, when to reach for it, why a docket ruling cannot hold it, and the measured case — and the
// manual puts that page in front of the bench before it chooses. A verb the HELP does not name
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
  required: ['claim_count', 'saturation_reached', 'sitting_record_appended'],
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
    sitting_record_appended: { type: 'boolean' },
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
    // from the manifest-row EVENTS at capture, and the rows reach the reader in the report.
    manifest: { type: 'array', items: { type: 'string' } },
    // THE ENGAGED GAPS BLUE FOUND ALREADY CLOSED WHEN IT SAT (#868). The lenses sit before blue,
    // so a lens can close a gap blue was dispatched onto; blue then has nothing to repair and
    // owes no row. This engine reads no record, and the lenses return prose, so blue's board
    // read is the only word of it that reaches here. It is trusted for ONE thing — which gaps
    // the in-run coverage log names — and for nothing that scores: capture recomputes the owed
    // set from the record's dispatch, close and register events (record.ManifestOwed).
    found_closed: { type: 'array', items: { type: 'string' } },
    log: { type: 'array', items: { type: 'string' } },
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

// The chair's envelope RELAYS the record (plans/roundless.md §III.B.1): `plan` is the JSON `dispatch next`
// printed, verbatim — the verb computed it from the board and recorded it; the chair could drop or add a
// party on the way, which is why capture's dispatch-parity audit checks the relay against the record. The
// verdict is what the chair RECORDED this sitting, if it recorded one. Nothing else the old envelope carried
// (gaps, closures, dispute responses, citation counts) is here: the record holds them and every seat reads
// them through the tool.
const PLAN = {
  type: 'object',
  required: ['head', 'parties', 'pass_permitted', 'ceiling'],
  properties: {
    head: { type: 'integer', minimum: 0, description: 'events.id of the report head the parties audit' },
    parties: {
      type: 'array',
      items: {
        type: 'object',
        required: ['seat_id', 'gap_ids'],
        properties: {
          seat_id: { type: 'string' },
          gap_ids: { type: 'array', items: { type: 'string' }, description: 'the gaps this party is engaged on; empty for a lens whose pin the head moved past' },
        },
      },
    },
    docket: { type: 'array', items: { type: 'string' }, description: 'gaps the verb docketed for the bench at this sitting' },
    pass_permitted: { type: 'boolean' },
    ceiling: { type: 'boolean' },
    max_epochs: { type: 'integer', minimum: 0, description: "the run's epoch limit, a term setup records; 0 when the run is held to none" },
    epoch_limit_reached: { type: 'boolean', description: "this chair sitting opens the run's last epoch: nobody is dispatched, and ceiling is set for that reason" },
    why: { type: 'array', items: { type: 'string' } },
  },
}
const CHAIR_ENVELOPE = {
  type: 'object',
  required: ['plan', 'unruled_motions'],
  properties: {
    plan: PLAN,
    verdict: { type: 'string', enum: ['PASS', 'FAIL'], description: 'the verdict you RECORDED this sitting, restated — the record is the original; absent if you recorded none' },
    unruled_motions: { type: 'integer', minimum: 0, description: 'motions on the record with no ruling, read back from the motions projection of the record — what the terminal bench sitting exists to dispose of' },
    petitions: PETITIONS,
    notes: { type: 'string' },
    log: { type: 'array', items: { type: 'string' } },
  },
}
const JUDGE_ENVELOPE = {
  type: 'object',
  required: ['resolutions'],
  properties: {
    // HOLDINGS THE BENCH LAID DOWN THIS SITTING, carried so the engine can route them. debate.js
    // reads no record, so a holding recorded through `bench declare` reaches the other seats only
    // if it travels here — the same reason `relief` is on the petition envelope (#503).
    holdings: { type: 'array', items: { type: 'string' } },
    log: { type: 'array', items: { type: 'string' } },
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
          // THIS LIST IS THE RECORD'S DISPOSITION VOCABULARY, EXACTLY, and the envelope/record
          // gate holds it there: every word offered here is a word `motion docket rule --as`
          // accepts, and every word that verb accepts appears here with a duty in
          // BLUE_DUTY_BY_RESOLUTION. A value added to one side fails the gate until it is on both.
          //
          // A GRADE OUTCOME IS NOT A DISPOSITION. "Gap real, grade wrong" is a GRADE MOTION —
          // carried by `grade_disputes` on this envelope and by `motion grade rule --as accepted`
          // plus a `regrade` on the record, which is the richer path because it can be appealed.
          // Ruling one here would put a grade outcome in the docket's field.
          //
          // `moot` asserts neither of the things its neighbours assert: nobody won an argument
          // (that is not_a_defect) and nobody verified a fix (that is repaired). The text the
          // finding attached to is gone. It closes the gap and leaves the merits unreached.
          //
          // `carried` is the word for "I could not settle this on the record I have" — it says
          // the gap survives and names what the coming seat owes, which is the same act as
          // asking for evidence.
          resolution: { type: 'string', enum: ['repaired', 'repaired_with_regression', 'amends_prior', 'not_a_defect', 'defect_accepted', 'carried', 'moot', 'defect_owed_elsewhere'] },
          rationale: { type: 'string' },
        },
      },
    },
  },
}

// THE STRATEGIC AREAS (#771). Each is a distinct kind of attack on the report and each gets
// exactly one seat, named for what it audits: `found_by` labels are <area>-F<n> and every
// cross-round lens-economics read joins on them.
//
// THE LENS TEXT IS NOT HERE ANY MORE. Each area's standing instructions are its agent
// configuration — `agents/red-lens-<area>.md` — which is where a prompt belongs and where it can
// be reviewed as one. What stays here is what the ENGINE needs: the key, because it composes the
// seat id and the dispatch label from it. record.LensAreas and the agent files are held against
// this list in every direction by TestTheLensAreasMatchWhatTheEngineDeclares.
const RED_AREAS = ['evidence', 'logic', 'dark-side', 'voice', 'computation', 'adversary', 'architecture']
// EVERY LENS AREA NAMES ITS AGENT TYPE AS A LITERAL, one per line. The record's role attestation
// (agentTypeRoles) is bound to this script by READING these literals; a type composed from the
// seat id at the dispatch site would dispatch correctly and be attested by nothing. The two lists
// are checked against each other at load, below, so an area cannot be declared and undispatchable.
const LENS_DISPATCH = {
  'red-lens-evidence': { agentType: 'frank-exchange-of-views:red-lens-evidence' },
  'red-lens-logic': { agentType: 'frank-exchange-of-views:red-lens-logic' },
  'red-lens-dark-side': { agentType: 'frank-exchange-of-views:red-lens-dark-side' },
  'red-lens-voice': { agentType: 'frank-exchange-of-views:red-lens-voice' },
  'red-lens-computation': { agentType: 'frank-exchange-of-views:red-lens-computation' },
  'red-lens-adversary': { agentType: 'frank-exchange-of-views:red-lens-adversary' },
  'red-lens-architecture': { agentType: 'frank-exchange-of-views:red-lens-architecture' },
}
for (const area of RED_AREAS) if (!LENS_DISPATCH[`red-lens-${area}`]) throw new Error(`debate: lens area ${area} is declared and has no dispatch row — LENS_DISPATCH must name its agent type`)
for (const seat of Object.keys(LENS_DISPATCH)) if (!RED_AREAS.includes(seat.replace(/^red-lens-/, ''))) throw new Error(`debate: LENS_DISPATCH names ${seat}, which RED_AREAS does not declare`)

// The areas a run dispatches. Default is the four that have always sat; the rest are opt-in per
// run, because every area costs a full dispatch every round against a concurrency cap measured at
// about two concurrent agents — seven areas is four waves where four is two.
const DEFAULT_AREAS = ['evidence', 'logic', 'dark-side', 'voice']

// AREA SELECTION IS VALIDATED AT THE DOOR, because a typo'd area name is a lens that silently
// does not sit — a run missing an entire kind of audit, reading exactly like a run that asked for
// fewer. The refusal names what was asked for and what exists.
const selectedAreas = (Array.isArray(lensAreas) && lensAreas.length ? lensAreas : DEFAULT_AREAS).map(String)
{
  const unknown = selectedAreas.filter((k) => !RED_AREAS.includes(k))
  if (unknown.length) {
    throw new Error(`debate: unknown lens area(s) ${JSON.stringify(unknown)} — the areas are ${JSON.stringify(RED_AREAS)}`)
  }
}


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
const friction = [] // capability complaints from any agent, aggregated for /self-improve
const takeFriction = (who, env) => { if (env && env.log) for (const f of env.log) friction.push(`${who}: ${f}`) }

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
const SITTING_RECORD = {
  type: 'object',
  required: ['sitting_record_appended'],
  properties: {
    sitting_record_appended: { type: 'boolean' },
    note: { type: 'string' },
  },
  additionalProperties: true,
}

async function ensureSittingRecord(env, who, owed, opts) {
  if (env && env.sitting_record_appended === true) return true
  log(`sitting-record (W1.7): ${who} did not attest ${owed} — re-prompting once before continuing (#249 recovery)`)
  const retry = await agent(
    `Sitting-record repair for ${who}. Your last sitting did not attest what it put on the record, so the run cannot yet show ${owed}. Put it on the record NOW — nothing else. ${owed}. Do NOT re-do your substantive work and do NOT edit the report again; this turn exists only to close the parity gap. If you genuinely cannot (the duty does not apply, or a tool refuses you), say in the log exactly why, and return sitting_record_appended false with a one-line note. Return the attestation.`,
    { ...(opts || {}), label: `${who}-sitting-record · ${slug}`, phase: 'Debate', schema: SITTING_RECORD })
  if (retry && retry.sitting_record_appended === true) {
    log(`sitting-record: ${who} attested on the retry — continuing`)
    return true
  }
  const why = (retry && retry.note) ? ` — ${retry.note}` : ''
  friction.push(`${who}: sitting-record (W1.7) UNRESOLVED — ${owed} was never attested${why}. The run continued; capture's record-parity audit scores the gap.`)
  log(`sitting-record: ${who} still unattested — logged as friction, continuing (the run is no longer discarded for this)`)
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
// WHAT A RULING OBLIGES BLUE TO DO, derived from the fate rather than restated in prose.
//
// Red's duty is "do not re-raise this proposition" and the settled line carries it. Blue's is "do
// not re-argue this claim in the report", which no id can enforce — blue's artifact IS the report,
// and that is where a settled proposition gets restated. And blue needs the inverse red never
// does: what it may now ASSERT. Two of these fates are blue WINS, and under a bare subtraction
// they look exactly like the ones that are not: the gap simply stops being dispatched. Keyed on
// the resolution so a new fate cannot quietly inherit another's instruction; an unmapped one
// falls through to a LOUD default rather than an empty string.
const BLUE_DUTY_BY_RESOLUTION = {
  not_a_defect: 'THE BENCH FOUND NO DEFECT — your position was vindicated. Keep the text as it stands; do not "repair" what the bench has just blessed. You may rely on this ruling as established for the rest of the run.',
  defect_accepted: 'YOUR RISK-ACCEPTANCE ARGUMENT WAS ACCEPTED. Record the acceptance where the report discusses the risk; do not spend a sitting fixing what the bench agreed may stand.',
  repaired: 'Your fix was accepted. Stop working this one.',
  carried: 'The gap stays OPEN and you owe the research direction the ruling states — it is what CEILING is made of, so a sitting that ignores it is a null turn.',
  defect_owed_elsewhere: 'The finding was UPHELD and the fix is owned outside this debate. Stop trying to fix it in the report; expect the debt to be named rather than closed here.',
  moot: 'Adjudicated out of existence — the text it attached to is gone, so there is nothing left to repair. Drop it; this is NOT a finding that was argued down.',
  repaired_with_regression: 'Your fix was accepted AND something else broke. Stop working this one and expect a successor gap naming the regression — answer that one, not this.',
  amends_prior: 'A defect found between two repairs that each closed clean. The ruling names what it supersedes; work the lineage it states rather than re-opening the ancestor.',
}
// rulingsInEffect holds the bench's latest ruling per gap — settled, reopens_on, final and the
// fate — and travels to BOTH parties, because debate.js reads no record: a ruling reaches a seat's
// prompt here or not at all (#517, #524). The reasoning stays on the record, deliberately.
const rulingsInEffect = new Map()
const rulingsClause = (party) => {
  if (!rulingsInEffect.size) return ''
  const rows = [...rulingsInEffect.values()].map((r) => party === 'blue'
    ? { ...r, your_duty: BLUE_DUTY_BY_RESOLUTION[r.resolution] || `UNMAPPED FATE ${r.resolution} — read the opinion on the record before acting on it` }
    : r)
  const duty = party === 'blue'
    ? ' You were once handed these as a bare SUBTRACTION — a ruled gap simply stopped appearing — and a vindication and an upheld finding were the same absence. THE BAR IS ON THE PROPOSITION, NOT THE GAP: what you may no longer re-argue in the report is the sentence under settled, not everything the finding touched. WHERE A RULING WENT YOUR WAY IT IS YOURS TO INVOKE — say so on the record and rely on it, rather than quietly re-fixing text the bench has already blessed.'
    : ' YOU ARE ESTOPPED on each settled proposition: do not re-raise it as a fresh gap or a successor; a defect that survives the ruling arrives as the reopens_on condition or not at all.'
  return ` GAPS THE BENCH HAS RULED, AND WHAT EACH RULING OBLIGES OF YOU: ${JSON.stringify(rows)}.${duty} A ruling marked final does not reopen; one carrying a reopens_on condition reopens only on that condition and not on a re-reading. THE REASONING IS ON THE RECORD, NOT IN THIS PROMPT — read the bench's opinion for any fate you are about to rely on or work around, because a fate you never read is one you can neither honour nor invoke.`
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
// holdingsClause renders every holding in effect, for EVERY seat, as routing refs.
//
// The text is the bench's own words because a construction is operative — a seat cannot act on a
// pointer to a construction it has not read — but the REASONING stays on the record, which is
// where the seat is sent for it.
const holdingsClause = () => {
  if (!holdingsInEffect.length) return ''
  return ` BENCH HOLDINGS IN EFFECT, BINDING ON EVERY SEAT FOR THE REST OF THE RUN: ${JSON.stringify(holdingsInEffect)}. A holding construes how the record is READ; it disposes of nothing and it does not expire with the sitting. Read the bench's reasoning for any holding you rely on or work around.`
}

const reliefFor = (party) => {
  const mine = reliefInEffect.filter((r) => r.binds === party || r.binds === 'both')
  if (!mine.length) return ''
  return ` BENCH RELIEF IN EFFECT, BINDING ON YOU (granted on a petition and operative this sitting — it is an order, not advice): ${JSON.stringify(mine.map((r) => r.relief))}.`
}
let halted = false
let haltOpinion = null
// A PETITION SITTING IS NAMED FOR ITS PETITIONER.
//
// A seat id is what every act is attributed to. Deriving it from the PETITIONER — `judge-petition-`
// plus the filer's own seat id — means the bench sitting that hears blue-synthesize's petition and
// the one that hears red-chair's are different seats, without a counter anyone has to remember to
// increment. When the same petitioner petitions again, the id is the same and WHICH occasion this
// was is the sitting ordinal the record computes from the registers (plans/roundless.md §III.A.0)
// — the same answer every other seat gets, and nothing a regex has to recover from the name.
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
  `Blue lane ${i + 1} of ${lanes} for topic: "${topic}". Read the frontier hypotheses from the RECORD — the \`lines-of-inquiry\` projection — where each carries an id and its hypothesis, and if ${runDir}/inputs/red-gap-patterns.md exists (red's accumulated gap-pattern inventory, staged at run setup), read it too — yesterday's expensive red discovery is today's free checklist line. Research your assigned slice to saturation per the research protocol (spend at least one search in five on disconfirming evidence; note each source inline with its URL, title, and access date). Your assigned METHOD: ${LANE_METHODS[i % LANE_METHODS.length]} — work primarily through this method's source class; take hypothesis ${i + 1} first, then breadth. YOUR SOURCES, IN PROSE (not footnotes): a lane writes its part of the report and does not cite — record each intended source as prose beside the claim (its URL, title, and access date) so the synthesizer can cite it once with the tool. Do NOT mint footnote labels; citations are tool-managed at synthesis and a hand-typed label backs nothing. Write your part of the report, in full, to ${runDir}/blue/candidates/lane-${i + 1}.md.${inquiryClause}${speedClause}${recordClause(`blue-lane-${i + 1}`)} Return a 3-line synopsis.`,
  { ...bulk, label: `blue-lane-${i + 1} · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-researcher' })))

let blueEnv = await agent(
  `Blue synthesis for topic: "${topic}". PRE-FLIGHT: if ${runDir}/inputs/red-gap-patterns.md exists, read it before merging — check your synthesis against red's known gap patterns. Read every draft in ${runDir}/blue/candidates/ and synthesize ${runDir}/blue/report.md by UNION: deduplicate overlapping claims, reorganize freely, and let no CLAIM leave without a record. Compacting and reorganizing prose is not subtraction — rewriting for density is encouraged; losing a claim in the rewrite is the failure, and substance leaves only by being retired on the record, which names the claim, why it goes, and what replaces it. THE REPORT'S CONTENT RULE — what may be in it at all, and where everything else goes — is on the verbs that write report text: edit, ingest, and line-of-inquiry's propose and move.
THE CATECHISM IS YOURS (W2g — the run-5 assembly audit found the judge-authored catechism DEFECTIVE, 6/7 answers carrying defects from synthesis-by-recall; synthesis surfaces belong inside the audited document): write "## The Catechism" INTO ${runDir}/blue/report.md at synthesis per references/catechism_template.md — the seven answers, the case against at FULL strength (include every risk-accepted residual; the audit found against-cases silently drop their strongest items), all figures copied from your own sourced sections never recomputed inline. THE FRAMING, THE TL;DR AND THE OPEN QUESTIONS ARE YOURS FOR THE SAME REASON — a synthesis surface authored at assembly is authored AFTER red's final audit, so nothing checks it and it ships unaudited. report.md OPENS with the H1 title \`# ${topic} — research report\`, then \`## TL;DR\` (3-6 sentences — the answer, the confidence, the sharpest caveat), and carries \`## Open questions\` (what the debate could not resolve, one per line; a question nobody could answer inside the debate is a finding, not noise). Assembly LIFTS these verbatim, so you write them once, HERE, inside the document red audits every sitting.

YOUR REPORT CONTAINS ONLY WHAT YOU CAN AUTHOR: the title, TL;DR, Catechism, technical foundations, analysis, open questions, and your citations — and NOTHING assembly composes from the record. Do NOT write "## Risk matrix", "## The board", "## The debate", or a "## Blue team report" wrapper: a "## The board" section you author is FABRICATION, because you cannot know red's findings, blue's audited manifest or the bench's dispositions, and a whole final-report-shaped document makes every tool-owned section appear twice. Do NOT write a \`**Outcome:**\` line ANYWHERE either — the outcome is decided after your last audit, so an outcome you author only lands stale beside the real one and gets stripped.

CITATION HYGIENE, because red must be able to VERIFY every source at the leaf: the url you cite carries the FULL coordinate — owner/repo#N or a full URL, never a bare #N red cannot resolve without the repo — and every QUOTED span records a locating anchor (its section heading, or a nearby unique phrase) so red can grep it verbatim in the source. An unlocatable quote is graded LOW through no fault of the source. Two claims resting on the same url are one source and two anchors, not two sources.

REPORT WRITE PATH: report.md hits a filename-keyed Write-guard — draft to your scratchpad under a NEUTRAL name and copy it into place with bash cp; a direct Write of the report path fails and wastes a round-trip. Recompute claim_count with the tool once report.md is written and relay the integer into the envelope; never hand-count it (two honest merges differed 2x before it was pinned). RECORD THE SITTING, with WHY you organised it as you did — what you compacted, what you kept whole, what you retired and against what — and its claim_count as the reason; sitting_record_appended is TRUE only after that event exists. FREEZE THE REPORT, LAST — after report.md is written, the count recorded and the sitting recorded, run the act your seat's help lists as freezing the report into the record as its base. Do this ONCE, at the very end. If it refuses, record friction with the exact refusal and return sitting_record_appended honestly.${inquiryClause} ${petitionClause('blue-synthesize')}${frictionClause('blue-synthesize', 'blue')}${speedClause}${recordClause('blue-synthesize')} Return the blue envelope.`,
  { ...judgment, label: `blue-synthesize · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-synthesizer', schema: BLUE_ENVELOPE })

if (!blueEnv) throw new Error('blue synthesis returned null (agent failed) — aborting cleanly')
takeFriction('blue-synthesize', blueEnv)
await ensureSittingRecord(blueEnv, 'blue-synthesize', `your sitting-record event (the revision verb) stating the synthesis and its claim_count`,
  { ...judgment, agentType: 'frank-exchange-of-views:blue-synthesizer' })
await hearPetitions(blueEnv, 'blue-synthesize')
// ---- Debate loop: red audits gate; termination is the record's, and the engine refuses to spin ----
// ─── THE DEBATE: chair sits → the record says who sits → they sit → the chair sits again ───────────
//
// There are no rounds (plans/roundless.md §III.B.1). Each chair sitting opens an epoch: the chair
// runs `dispatch next`, which computes readiness FROM THE BOARD and records one dispatch per party;
// the chair relays the verb's JSON as `plan`; this loop dispatches what the plan says and nothing
// else. Red parties sit first (in parallel), then blue, then the bench — an exchange is a red-party
// sitting followed by a blue-party sitting, and the record counts them (impasse.go). Empty is the
// termination signal: pass_permitted (the chair issues PASS → VERIFIED), ceiling (every open
// material gap at its limit, ruled and carried, or the run's epoch limit reached → CEILING), or
// neither (UNVERIFIED, with the plan's reasons on the record). The disputes, the docket, impasse,
// the ceiling and the epoch limit all live on the record; nothing here keeps a second copy of them.
//
// ONE STOP IS THE ENGINE'S OWN: THE NO-PROGRESS VALVE. A plan the record computes identically for
// NO_PROGRESS_EPOCHS chair sittings in a row is a loop that nothing on the board is moving — the same
// parties readied against the same head for the same reasons, sitting after sitting, while the run
// spends. It stops the debate UNVERIFIED and names the stuck parties and the head. Not CEILING: no gap
// reached its limit and no term ran out; the parties were still ready, and the record says so.
//
// IDENTICAL MEANS THE WHOLE PLAN, reasons included, not only the head and the parties. A gap marching
// to impasse keeps its head and its parties while the exchange count in its reason advances — that
// march is progress the run's terms (k, k-max) bound, and a valve blind to the reasons would cut it
// short under any k of 3 or more, or a lens regrading without an edit. A seat that sits and never
// registers moves no count, so its plan repeats exactly: the loop this valve exists for.
const NO_PROGRESS_EPOCHS = 3
const partyList = (ps) => ps.map((p) => `${p.seat_id}${p.gap_ids && p.gap_ids.length ? `[${p.gap_ids.join(' ')}]` : ''}`).join(', ')
const sortedStrings = (xs) => (Array.isArray(xs) ? xs : []).map(String).sort()
const planKey = (p) => JSON.stringify({
  head: p.head,
  parties: p.parties.map((x) => ({ seat_id: String(x.seat_id), gap_ids: sortedStrings(x.gap_ids) })).sort((a, b) => (a.seat_id < b.seat_id ? -1 : a.seat_id > b.seat_id ? 1 : 0)),
  docket: sortedStrings(p.docket),
  why: sortedStrings(p.why),
  pass_permitted: !!p.pass_permitted,
  ceiling: !!p.ceiling,
})
let samePlan = { key: null, epochs: 0 }
let noProgress = null // { epochs, head, parties } when the valve stopped the debate
let epoch = 0
let chairEnv = null
let lastPlan = null
let blueEnv2 = null // the latest blue response, for claim_count in the result
const infraDebts = [] // defect_owed_elsewhere rulings (W1.9) — the bench's named debts, surfaced at assembly and in the final envelope

const ledgerClause = ` THE REPORT DOES NOT NAME ITS SOURCES — it carries invisible anchors, and the evidence layer is the only way from one to what it points at. Read it FIRST. A claim blue backed with a source resolves to the url, title and sha256 of the bytes blue actually read, and THAT is what you re-fetch to audit the same artifact rather than a page that may have drifted since. A claim blue COMPUTED resolves to its script, and to whether anyone has re-run it — an unaudited proof looks exactly like a clean one until you check. Your own verifications come back ATTACHED TO THE SOURCE you checked: an empty one is a citation NOBODY has checked yet, and that is where your next pass is worth most. WHAT NOT TO RE-CHECK, because it is the sitting's whole economy: a claim verified at HIGH confidence in a prior sitting STAYS verified — do not re-fetch it UNLESS its section changed since (read that off the recorded edits, not off blue's account of them), OR more than 2 epochs have elapsed since it was last verified, OR its access date and the source's volatility suggest drift (living documents, issue trackers, README stats). RECORD every claim you verify, naming the anchor you looked up and quoting the claim from the report. A SOURCE YOU FOUND YOURSELF IS A DIFFERENT ACT — blue never cited it, so there is no anchor to name — and it answers a different question: whether the claim is true in the WORLD, where verification asks only what blue's source did for it. VERBATIM READS ONLY — you have no WebFetch, by design: it returns a small model's SUMMARY, not the source. Read blue's cached bytes through the tool; for a source YOU discover, pull it verbatim yourself (\`curl -sL <url>\`, \`gh issue view <n> --comments\`, \`pdftotext\`/pandoc) and read it. WebSearch is for DISCOVERY — finding the url — never for the read that grades a citation. If a source is too large for your context, read it in SECTIONS and name the sections you read: A TRUNCATED READ IS NOT A READ, so state the truncation and never grade a body you could not fully read.`
const areaOf = (seatID) => seatID.replace(/^red-lens-/, '')
// THE LENS: it audits its area and puts what it finds on the board ITSELF (plans/roundless.md
// §III.B.3). The duties of the seat that mints — asking for the answer to be PRODUCED, classing
// the probe, reading the script, keeping lineage — sit here, with the minter.
const lensPrompt = (seatID, gaps, plan) => {
  const area = areaOf(seatID)
  const extra = area === 'evidence' ? ledgerClause : (area === 'logic' || area === 'dark-side') ? steelmanClause : ''
  const engaged = gaps.length
    ? ` YOU ARE ENGAGED ON: ${gaps.join(', ')} — gaps you minted that are still open and below their limits. Re-read the report where each is anchored and DID BLUE ACTUALLY DO WHAT YOU ASKED? Put your required fix and blue's edits side by side (the record's changes projection names the recorded edits, not blue's account of them) rather than inferring the answer. Then act: regrade on what you now see, or close with the verification triple where the repair holds. A gap you neither move nor close this sitting is a NULL TURN on it, and null turns count toward its impasse — silence is a turn taken.`
    : ` THE REPORT HEAD MOVED PAST YOUR LAST SITTING (head ${plan.head}): audit the report as it now stands, in full.`
  return `Red lens sitting, area ${area}, topic "${topic}".${extra}${engaged} RE-READ THE FULL REPORT IN CONTEXT — the whole document, never just a diff; if it exceeds one Read call, read it whole in consecutive windows. ANCHOR EVERY FINDING TO A QUOTED SENTENCE, and quote it exactly rather than paraphrasing: a finding whose quote is not found in the report as the record renders it is REJECTED. The labels on your findings are the tool's to assign, and so are the gap ids you mint. HARNESS NOTES: Grep count mode counts LINES, not occurrences — anchor patterns (e.g. '^### ') when counting; prefer the Write tool over quoted heredocs for scripts (heredoc backslash mangling is a documented recurrence).
YOU MINT YOUR OWN GAPS. What you find that is real and belongs on the board goes there as YOUR gap: screen every candidate against the board first for a near match — a defect already closed is a REOPEN and arrives carrying that history or arrives lying; a duplicate minted fresh forks the lineage — then mint, one considered gap at a time, graded on every axis, registering a new class first where the registry lacks one. Nobody coalesces for you and nobody transcribes for you: a finding is your graded observation; a gap is your claim on the board. Your budget scales with the report: the larger of the run's floor (mintBudget in run-config.json) and one mint per so many units of what your area audits, read off the record at each mint — a refused mint states the arithmetic. Spend it on defects a reader would pay to have fixed, not on nitpicks — a trifle costs you a mint and holds nothing open, because a gap below material does not hold the gate.
ASK FOR THE ANSWER TO BE PRODUCED, NOT ASSERTED. Where a claim is arithmetic, an enumeration, or a reproducible measurement, your acceptance check demands that it be computed, and says so in the check's kind; where it turns on what the document states or what a source says, say so plainly — there is no credit for inflating it. Say WHEN your demand can be discharged, in the two classes blue answers in: a DOCUMENT-PROBE is executable now against shipped artifacts; a LIVE-PROBE needs built artifacts and is DEFERRABLE in a design-phase debate, discharged by naming it as a deferred acceptance test with its pass condition. PRESCRIBE TEXT ONLY WHERE THE DEFECT IS TEXTUAL: the required fix is prose — what must become true — and stays the channel for substantive work.
BELIEVE NO BYTES YOU DID NOT WATCH BEING PRODUCED. Where blue backed a sentence with a computation, RE-RUN IT, then READ THE SCRIPT and say what it ACTUALLY COMPUTES: a script that re-runs clean and establishes nothing is the dangerous case, because it looks maximally credible.
THE ORIGINATOR CLOSES, AND LINEAGE IS NEVER DROPPED. A gap you minted is yours for its whole life: only you regrade it, only you close it — with the triple (who verified, with what, against what) — and a successor you mint names the gaps it supersedes, a closure that regressed says so, and the docket follows those chains. The chair carries archived closures and dispatches you when your gap needs acting on; the bench disposes of a docketed gap through its ruling.${rulingsClause('red')}${reliefFor('red')}${speedClause}${frictionClause(seatID, 'lens')}${recordClause(`red-lens-${area}`)}${petitionClause(seatID)} Return a 3-line synopsis.`
}
// THE CHAIR RUNS THE DEBATE. It mints nothing and closes nothing; the record says who sits and the
// chair relays it; the verdict, the closings, the spot-check, the rulings on blue's motions and
// directions and the vote on the lines of inquiry are what only the chair does.
const chairPrompt = () => `Red chair, topic "${topic}". You RUN the debate: you mint nothing and you close nothing — a gap belongs to the lens that minted it from mint to close — and you are the seat that decides whether this report has been verified.${recordClause('red-chair')}${speedClause}${scorecardClause()}${holdingsClause()}${reliefFor('red')}${lawClause}
FIRST, EVERY SITTING, ASK THE RECORD WHO SITS — the dispatch. It reads the board and RECORDS who sits — the lenses whose pin the report head moved past, the lens and blue of every open material gap below its limits, the bench for every gap at impasse (it dockets those itself). Relay its JSON VERBATIM as \`plan\` in your envelope; the workflow dispatches what the record says and capture audits your relay against it, so a party you drop or add is a finding against you. Empty is the record's word that the run is over: with pass_permitted the board permits a PASS; with ceiling every open material gap is at its limit, ruled and carried — or, with epoch_limit_reached, the run has reached its epoch limit and nobody further is dispatched.
THE STOPPING JUDGMENT IS YOURS, AND IT IS NOT CEREMONY. When the plan says pass_permitted, decide: record a PASS verdict if you agree the report is verified — the tool refuses a PASS the board does not permit, so you cannot pass early — or a FAIL with the material defect that stops you, raised as a finding for its lens to mint; a FAIL over a converged board is refused (raise something material, or pass). Otherwise record no verdict this sitting. Your recorded verdict is the ONE fact the run's outcome is derived from.
YOUR POSITION IS YOUR ARGUMENT and the other side answers it: one position per sitting. CLOSING ARGUMENTS on every gap the plan docketed — each is docket-bound and owes ~120 words, your strongest evidence and your answer to blue's — the bench rules on the closings and the artifacts, not on prose in your envelope.
A CLOSURE IS A CLAIM, AND CLAIMS DECAY. Re-sample the archive every sitting it is not empty (the spot-check; its assertable empty form only when the archive was empty when you sat) and put what the sample FOUND in the spot-check's own prose — not in \`log\`, which is the operator's channel; a lens reopens a drifted closure of its own, and a closure resting on a volatile living source inherits that source's drift triggers.
VOTE EVERY LINE OF INQUIRY THIS SITTING, ON ONE READ: read the report ONCE and answer every line against that pass, the way anyone checks a document against a list — not once per line, and from THIS read, because the report is rewritten between sittings. RULE ON BLUE'S DIRECTIONS: a ruling is an ARGUMENT, not a command — it needs a reason, and blue may appeal it. RULE THE GRADE MOTIONS blue filed: accept and the minting lens owes the regrade; reject and blue may re-dispute. Report what still stands unruled as unruled_motions, read from the record's motions projection and never counted by hand.
NEVER RE-DERIVE THE BOARD IN YOUR HEAD: the board, work and motions projections are the reads, and the plan is the record's, not yours.${frictionClause('red-chair', 'merge')}${petitionClause('red-chair')}`
const bluePrompt = (gaps, docket) => `Blue response, topic "${topic}". You are engaged on: ${gaps.join(', ')}.${reliefFor('blue')} YOUR FIRST READ COMES AFTER THE MANUAL BELOW, NOT BEFORE IT: pull your working set — the board and work projections, the transcript, and ${runDir}/inputs/red-gap-patterns.md — in one pass rather than three, concatenating them into a single file under your session scratchpad (an ABSOLUTE path, never under ${runDir}) and reading that.${recordClause('blue-respond')}${speedClause}${holdingsClause()}${rulingsClause('blue')}${lawClause}
READ THE BOARD FOR YOUR GAPS: each carries its lens's problem, required fix and acceptance check, and the transcript's RED section carries red's argument; the bench's latest resolutions are on the record too, and any gap the bench CARRIED comes with a stated research direction you owe. The gap list you were handed is a lossy summary of the record, and the record is authoritative. PRE-FLIGHT: re-check your planned repairs against red's gap patterns — the staged inventory names, per gap class, how this class of repair goes wrong — and say in each manifest row which patterns you checked. A row you got wrong is corrected in this sitting by the same act for that gap — never by a second row, and never by moving its text into a log.
YOU MAY COMPUTE AN ANSWER, NOT ONLY COLLATE SOURCES. You have Bash, Write and Edit, and for a whole class of questions running something settles it faster and harder than arguing about it. WORKING IT OUT IN YOUR HEAD IS NOT THE EXCEPTION — IT IS THE CASE THIS EXISTS FOR: a correct figure with no derivation is indistinguishable from a confident guess. A gap can be WAITING ON A PROGRAM FROM YOU — it says so (a computation check), and where it does, no amount of prose will close it. DOCUMENT-PROBE checks you discharge now; a LIVE-PROBE you discharge by naming the deferred acceptance test and its pass condition.
YOUR LINES OF INQUIRY ARE A LIVING RECORD, NOT AN OPENING PLAN. Every sitting, revisit what is still open and say what became of it. The hypothesis is what makes that honest — a line abandoned against its own stated claim is evidence of choosing; one abandoned on a shrug is not. Red rules on your proposals and you may APPEAL a ruling — appeal whether or not you go on to pursue the line, because the appeal is where your ARGUMENT is recorded.
WHERE RED PROPOSED EXACT TEXT you have THREE paths and you are not obliged to take the first: apply it verbatim, counter-edit with your own fix, or dispute it. Say plainly which path you took and why. A sitting in which you never decline is not agreement, it is capitulation. Applying red's exact text ESTOPS red from re-raising that text as a fresh gap, so verbatim application is a real settlement, not a surrender.
OWNERSHIP BINDS, AS IT DID AT SYNTHESIS: you write ONLY your own surfaces — TL;DR, Catechism, technical foundations, analysis, open questions, your citations. You MUST NOT introduce a \`**Outcome:**\` line or any tool-owned section (\`## Risk matrix\`, \`## The board\`, \`## The debate\`, \`## How this run was conducted\`, a \`## Blue team report\` wrapper): assembly composes those from the record. AND WHICH MODEL IS ANSWERING THIS RUN IS NOT A FACT YOU HOLD: you can see what was REQUESTED, never what replied; if your argument turns on the models used, say that the record's measurement decides it. GRADING SEMANTICS (mapping ${MASS_MAPPING_VERSION}): likelihood and impact are CONSEQUENCE axes — how likely the harm lands and how much it costs when it does — and severity is red's overall grade; they multiply into mass, and material begins at medium.
ANSWER EVERY GAP YOU ARE ENGAGED ON, ADDITIVELY, in the report through \`edit\` — expand and repair where red is right (each edit naming the gap it answers, so the repair joins to it), REBUT IN WRITING WITH EVIDENCE where red is wrong, and argue risk-acceptance where the fix's complexity exceeds its likelihood x impact. DISPUTE RED'S GRADING WHERE YOU DISAGREE WITH IT (a grade motion on the axis, with the grade you say it should be and your evidence). Compact and reorganize prose as clarity demands — but a CLAIM leaves only by being retired on the record, which names it, why it goes, and what replaces it. PROPAGATE EVERY CORRECTION TO ALL SITES that state the corrected claim, not only the flagged sentence — a bare corrected FIGURE needs a report-wide sweep of its own — and list the sites you checked in your revision. A sitting in which you record nothing on a gap you were engaged on is a NULL TURN on it, and null turns count toward its impasse: silence is a turn taken.${docket.length ? ` CLOSING ARGUMENTS: the following are DOCKETED for adjudication AFTER your response this sitting: ${docket.join(', ')}. For EACH, after your repairs, argue in ~120 words why your response resolves it, or why red's grade or claim is wrong, citing the exact section and evidence, as your recorded closing. This is your case; material not in the record cannot help you.` : ''}
AUDIT YOUR OWN REPAIRS, ONE RECEIPT PER GAP (W2b; your constitution carries the full standard): figures recomputed, universals enumerated, consistency sites swept report-wide — one manifest row per gap you REPAIRED — one an edit of yours this sitting answers — and the manifest array in your envelope names them; a gap you rebut without an edit owes none. A gap you are engaged on that the board shows CLOSED when you sit — its lens sat before you and closed it — owes no row: name it in found_closed, from the board you read. Record the sitting's revision with the tool's claim_count; never hand-count it — the tool computes claim_count with the tool's own read.${frictionClause('blue-respond', 'blue')}${petitionClause('blue-respond')}`
const benchPrompt = (gaps) => `Adjudication, topic "${topic}". Docketed for you: ${gaps.join(', ')} — each reached impasse under the run's terms and the record docketed it. THE DOCKET IS A ROUTING LIST, NOT THE EVIDENCE. It carries ids; a gap's problem text and its acceptance check live on the board, and you read them FRESH before ruling. Re-run each document-probe acceptance check against the artifact AS IT NOW STANDS, and rule on what you find rather than on what any snapshot asserts.
YOUR RULING BASIS IS CONFINED TO THREE THINGS: the two sides' recorded closings, the full transcript, and the final state of the artifacts — the board and the report as the record now renders them. Weigh each closing as that side's best case, and a claim in a closing that the record does not support counts AGAINST the side that made it. For every ruling on a gap with a lineage chain, READ THE NAMED ANCESTORS' RECORDS first and NAME what you read in your rationale.${holdingsClause()}${lawClause}${declareClause}${inspectionClause}
Every docketed gap gets a written ruling — the docket ruling: its fate, the principle you applied, the values in tension, whether a human should look at it, your reasoning, and TWO THINGS THE FATE CANNOT SAY — the proposition you are barring as settled, and what would reopen it or that nothing would (final). Two fates route work OUT of the debate rather than ending it: a gap you CARRY stays open, owes blue a stated research direction, and is that gap's deadlock — what CEILING is made of; and a valid finding whose FIX is owned outside the debate — run tooling, the harness, the engine — leaves the board and ships as a NAMED infrastructure debt (defect_owed_elsewhere), recorded and never dropped. Rule the grade motions and directions the parties left for you. A bench sitting that rules nothing is a workflow error: the run cannot end in a verdict while a docketed gap stands unruled.${frictionClause('judge', 'bench')}${speedClause}${recordClause('judge')}${petitionClause('judge')} Return your envelope.`

phase('Red')
const sittings = {} // seat -> how many times this loop has dispatched it, for the labels
const labelFor = (seat) => { sittings[seat] = (sittings[seat] || 0) + 1; return `${seat} #${sittings[seat]} · ${slug}` }
while (!halted) {
  epoch++
  chairEnv = await agent(chairPrompt(), { ...judgment, label: labelFor('red-chair'), phase: 'Red', agentType: 'frank-exchange-of-views:red-chair', schema: CHAIR_ENVELOPE })
  takeFriction('red-chair', chairEnv)
  if (!chairEnv) throw new Error(`red-chair sitting ${epoch} returned null (agent failed) — aborting cleanly`)
  const plan = chairEnv.plan
  if (!plan || !Array.isArray(plan.parties)) throw new Error(`red-chair sitting ${epoch} relayed no plan — the envelope carries \`dispatch next\`'s JSON verbatim, or the workflow has nothing to dispatch`)
  lastPlan = plan
  log(`epoch ${epoch}: head ${plan.head} — ${plan.parties.length} party(ies) ready${plan.docket && plan.docket.length ? `, docketed ${plan.docket.join(', ')}` : ''}${chairEnv.verdict ? `, chair recorded ${chairEnv.verdict}` : ''}${plan.pass_permitted ? ' — PASS permitted' : ''}${plan.ceiling ? ' — at the ceiling' : ''}`)
  if (await hearPetitions(chairEnv, 'red-chair')) break
  if (chairEnv.verdict === 'PASS') break
  // THE EPOCH LIMIT is the record's word, like the ceiling: the plan at the last epoch readies
  // nobody. A relayed plan that says the limit is reached ends the debate whatever parties it lists.
  if (plan.epoch_limit_reached) {
    log(`epoch ${epoch}: the epoch limit (${plan.max_epochs}) is reached — nobody further is dispatched`)
    break
  }
  if (plan.parties.length === 0) break
  const key = planKey(plan)
  samePlan = key === samePlan.key ? { key, epochs: samePlan.epochs + 1 } : { key, epochs: 1 }
  if (samePlan.epochs >= NO_PROGRESS_EPOCHS) {
    noProgress = { epochs: samePlan.epochs, head: plan.head, parties: plan.parties.map((p) => ({ seat_id: String(p.seat_id), gap_ids: sortedStrings(p.gap_ids) })) }
    log(`epoch ${epoch}: NO PROGRESS — the plan is identical for ${samePlan.epochs} consecutive epochs (NO_PROGRESS_EPOCHS = ${NO_PROGRESS_EPOCHS}) at head ${plan.head}; stuck: ${partyList(noProgress.parties)} — the debate stops rather than dispatch them again`)
    break
  }

  // Classified by seat id, never by object identity: a plan relayed through a host runtime may
  // hand out a fresh wrapper on every read, so two looks at one party need not be === equal.
  const roleOfParty = (p) => String(p.seat_id).startsWith('red-lens-') ? 'lens' : p.seat_id === 'blue-respond' ? 'blue' : p.seat_id === 'judge' ? 'bench' : 'stray'
  const lenses = plan.parties.filter((p) => roleOfParty(p) === 'lens')
  const blues = plan.parties.filter((p) => roleOfParty(p) === 'blue')
  const benches = plan.parties.filter((p) => roleOfParty(p) === 'bench')
  const strays = plan.parties.filter((p) => roleOfParty(p) === 'stray')
  if (strays.length) throw new Error(`epoch ${epoch}: the plan names ${strays.map((p) => p.seat_id).join(', ')}, which this workflow has no sitting for — the cast and the workflow disagree`)
  for (const p of lenses) if (!LENS_DISPATCH[p.seat_id]) throw new Error(`epoch ${epoch}: the plan names ${p.seat_id}, a lens area this workflow does not declare — the cast and the workflow disagree`)

  if (lenses.length) {
    log(`epoch ${epoch}: dispatching ${lenses.length} red lens(es): ${lenses.map((p) => `${areaOf(p.seat_id)}${p.gap_ids.length ? `[${p.gap_ids.join(' ')}]` : ''}`).join(', ')}`)
    const envs = await parallel(lenses.map((p) => () => agent(lensPrompt(p.seat_id, p.gap_ids, plan),
      { ...bulk, label: labelFor(p.seat_id), phase: 'Red', ...LENS_DISPATCH[p.seat_id] })))
    for (const [k, env] of envs.entries()) {
      takeFriction(lenses[k].seat_id, env)
      if (env && await hearPetitions(env, lenses[k].seat_id)) break
    }
    if (halted) break
  }
  for (const p of blues) {
    phase('Debate')
    blueEnv2 = await agent(bluePrompt(p.gap_ids, plan.docket || []), { ...bulk, label: labelFor('blue-respond'), phase: 'Debate', agentType: 'frank-exchange-of-views:blue-researcher', schema: BLUE_ENVELOPE })
    takeFriction('blue-respond', blueEnv2)
    if (!blueEnv2) throw new Error(`blue response (epoch ${epoch}) returned null (agent failed) — aborting cleanly`)
    await ensureSittingRecord(blueEnv2, 'blue-respond', `your position event for this sitting (it renders as the "### BLUE" section) AND your revision event`,
      { ...bulk, agentType: 'frank-exchange-of-views:blue-researcher' })
    // W2b, OWED GAPS ONLY, NEVER AN ABORT (gblock's ruling on #868). A row is owed for a gap blue
    // REPAIRED — its sitting's edit answers it — while the gap was still OPEN when blue sat; one
    // its lens closed first, or one blue rebutted without an edit, owes nothing. The plan was
    // computed at the START of the epoch and the lenses sat in between, so checking blue against
    // the plan aborted two runs over an empty manifest that was correct. This engine sees neither
    // the closures nor the edits; it logs the open gaps blue named no row for, and capture decides
    // which were owed from the record (record.ManifestOwed) — none, some or all of them alike.
    const foundClosed = new Set((Array.isArray(blueEnv2.found_closed) ? blueEnv2.found_closed : []).filter((g) => p.gap_ids.includes(g)))
    const open = p.gap_ids.filter((g) => !foundClosed.has(g))
    const covered = new Set(Array.isArray(blueEnv2.manifest) ? blueEnv2.manifest : [])
    const uncovered = open.filter((g) => !covered.has(g))
    if (foundClosed.size) log(`epoch ${epoch}: blue found ${[...foundClosed].join(', ')} closed before it sat — no manifest row owed (capture re-derives this from the record)`)
    if (uncovered.length) log(`epoch ${epoch}: manifest rows named for ${open.length - uncovered.length}/${open.length} gap(s) open when blue sat — no row for: ${uncovered.join(', ')} (owed where this sitting's edit answered it; scored at capture)`)
    log(`epoch ${epoch}: blue responded on ${p.gap_ids.join(', ')} — corpus at ${blueEnv2.claim_count} claims`)
    if (await hearPetitions(blueEnv2, 'blue-respond')) break
  }
  if (halted) break
  for (const p of benches) {
    log(`epoch ${epoch}: the bench sits on ${p.gap_ids.join(', ')}`)
    const judge = await agent(benchPrompt(p.gap_ids), { ...judgment, label: labelFor('judge'), phase: 'Debate', agentType: 'frank-exchange-of-views:lead-judge', schema: JUDGE_ENVELOPE })
    if (!judge) throw new Error(`bench sitting (epoch ${epoch}) returned null (agent failed) — aborting cleanly`)
    for (const h of judge.holdings || []) holdingsInEffect.push(h)
    for (const r of judge.resolutions || []) {
      rulingsInEffect.set(r.gap_id, { gap_id: r.gap_id, resolution: r.resolution, settled: r.settled, reopens_on: r.reopens_on, final: !!r.final, epoch })
      if (r.resolution === 'defect_owed_elsewhere') infraDebts.push({ gap_id: r.gap_id, owed_fix: r.rationale, epoch })
    }
    takeFriction('judge', judge)
    if (await hearPetitions(judge, 'judge')) break
  }
}

// THE RUN'S TERMINAL WORD IS A RECORD VOCABULARY, so it is declared as an enum rather than spelled
// inline. The word travels into the assembly seat's prompt ("Record it as the run's outcome —
// ${verdict}") and is typed straight at `bench outcome --as`, so a word the record refuses arrives
// as an INSTRUCTION and the engine carries on as though it had been written. The declaration is
// what the envelope-enum gate binds — that gate scans `enum: [...]`, and a vocabulary spelled as
// bare literals is out of its reach entirely — and outcomeWord makes a misspelling throw here.
const RUN_OUTCOME = { type: 'string', enum: ['VERIFIED', 'CEILING', 'HALTED', 'UNVERIFIED'] }
const outcomeWord = (w) => {
  if (!RUN_OUTCOME.enum.includes(w)) throw new Error(`not a run outcome: ${w} (have ${RUN_OUTCOME.enum.join('|')})`)
  return w
}

// A NO-PROGRESS STOP IS UNVERIFIED, NEVER CEILING. CEILING is derived on the record — every open
// material gap at its limit and carried, or the epoch limit reached — and `bench outcome` refuses it
// over a board that is neither; a stalled plan still has parties ready. UNVERIFIED is the word for a
// run that stopped before its record reached a terminal state, and the why says what stopped it.
const verdict = halted ? outcomeWord('HALTED')
  : (chairEnv && chairEnv.verdict === 'PASS') ? outcomeWord('VERIFIED')
  : noProgress ? outcomeWord('UNVERIFIED')
  : (lastPlan && lastPlan.ceiling) ? outcomeWord('CEILING')
  : outcomeWord('UNVERIFIED')
const terminationWhy = halted ? 'judicial halt'
  : verdict === 'VERIFIED' ? 'the chair recorded PASS with the board permitting it'
  : noProgress ? `no progress: the dispatch plan was identical for ${noProgress.epochs} consecutive epochs (NO_PROGRESS_EPOCHS = ${NO_PROGRESS_EPOCHS}) at head ${noProgress.head} — ${partyList(noProgress.parties)} readied again each time with nothing on the board moving`
  : verdict === 'CEILING' ? (lastPlan.epoch_limit_reached ? `epoch limit ${lastPlan.max_epochs} reached — the run's term on chair sittings; the parties the board still readied were not dispatched` : 'every open material gap is at its limit, ruled by the bench and carried')
  : (lastPlan ? `nobody was ready and neither PASS nor CEILING held: ${(lastPlan.why || []).join('; ')}` : 'the run ended before any dispatch')
log(`debate ended: ${verdict} after ${epoch} chair sitting(s)${halted ? ' (JUDICIAL HALT)' : ''} — ${terminationWhy}`)

// Terminal bench sitting: whatever the record still holds unruled at the exit boundary (grade
// motions, directions, a petition) is disposed of before assembly — the chair reported the count.
if (!halted && chairEnv && chairEnv.unruled_motions > 0) {
  const terminalJudge = await agent(
    `Terminal disposition for topic "${topic}" (debate ended ${verdict} after ${epoch} chair sitting(s); this sitting fires at the exit boundary). ${chairEnv.unruled_motions} motion(s) stand unruled on the record — read them back from the record's motions projection, read the transcript in full and the board back, and rule each with a reason. NOTHING CAN BE CARRIED AT A TERMINAL EXIT — there is no next sitting to carry it into. Each grade motion either moves to a corrected grade, which you state, or ships CONTESTED and the report records it as contested.${holdingsClause()}${lawClause}${declareClause}${inspectionClause}${frictionClause('judge-terminal', 'bench')}${speedClause}${recordClause('judge-terminal')} Return your envelope.`,
    { ...judgment, label: `judge-terminal · ${slug}`, phase: 'Assemble', agentType: 'frank-exchange-of-views:lead-judge', schema: JUDGE_ENVELOPE })
  if (terminalJudge) takeFriction('judge-terminal', terminalJudge)
}

const ASSEMBLE_ENVELOPE = {
  type: 'object',
  additionalProperties: false,
  required: ['synopsis', 'open_gaps'],
  properties: {
    synopsis: { type: 'string' },
    open_gaps: { type: 'integer', minimum: 0 }, // the board's counts.open, not red's docket
    log: { type: 'array', items: { type: 'string' } },
  },
}
phase('Assemble')
const assembleEnv = await agent(
  `Final assembly for topic "${topic}", run directory ${runDir}. Debate outcome: ${verdict} after ${epoch} chair sitting(s) — ${terminationWhy}. THE REPORT IS ASSEMBLED FROM THE RECORD — you author NOTHING, you copy NOTHING, you fill in NO inputs. There are no <FILL> fields and no sections for you to write; do not hand-write report.md and do not copy anything into it yourself.

FIRST, STAMP HOW THIS RUN ENDED: it is ${verdict}${halted ? `, ended by JUDICIAL HALT — carry the opinion verbatim: ${haltOpinion}` : ''}${verdict === 'CEILING' ? (lastPlan.epoch_limit_reached ? `, ended at the CEILING: ${terminationWhy} — a limit the run was set up with, NOT a judged failure to verify` : ', ended at the CEILING: every open material gap reached its limit, the bench ruled on each and carried it — NOT a judged failure to verify') : ''}${verdict === 'UNVERIFIED' ? (noProgress ? `, ended UNVERIFIED — the engine stopped a debate that was making no progress: ${terminationWhy}. It is not CEILING: no gap reached its limit and no term ran out; the parties were still ready` : `, ended UNVERIFIED — nobody was ready and neither PASS nor CEILING held; the plan's reasons are the account: ${terminationWhy}`) : ''}. Record it as the run's outcome — ${verdict}${verdict === 'CEILING' ? ', ended at the ceiling' : ''} — with WHY this outcome is the right read of the board as the reason — the judgement, never a recap of the sitting; the verdict is derived from the record and the stamp must agree with it.

THEN, TWO THINGS YOU MAY HOLD AND THIS IS YOUR LAST CHANCE TO RECORD EITHER. If you hold something that binds how the RECORD IS READ but moves no gap — a construction of a term, a correction of what the record MEANS rather than what it says, a holding worth offering as precedent — state it, and state it in its own right rather than folding it into an unrelated rationale. And if anything in this run needs A HUMAN to re-examine it — an unresolved tension, a claim that held only because nobody could reach the source, a boundary you ruled close to — say so. You keep no memory between runs, so this is the whole of your continuity.

THEN ASSEMBLE. It writes a SET — ${runDir}/report.md (the research), docket.md, debate.md, judgments.md, lines-of-inquiry.md, evidence.md, run.md, CHANGELOG.md, a README.md index and a tabbed report.html — and a document with nothing in it is not written at all, so a missing judgments.md means no motions were filed rather than a failure. Verify ${runDir}/report.md exists and reads correctly at the top: the verdict stamp is the outcome's, the sections are blue's and the record's. A tool cannot mis-author a synthesis surface — the TL;DR and the catechism are blue's, inside the audited report. Open gaps below material are listed as open, below material, not certified against. THE AUTHORITATIVE OPEN COUNT IS THE BOARD'S, after every closure and ruling: read it back and report it as open_gaps in your envelope. Infra debts the bench named: ${JSON.stringify(infraDebts)}. Collated friction so far (report any of your own as well): ${JSON.stringify(friction)}.${holdingsClause()}${lawClause}${frictionClause('assemble', 'bench')}${speedClause}${recordClause('assemble')} Return your envelope: a 5-line synopsis, open_gaps from the board, and your own friction if any.`,
  { ...judgment, label: `assemble · ${slug}`, agentType: 'frank-exchange-of-views:lead-judge', schema: ASSEMBLE_ENVELOPE })
return {
  runDir,
  verdict,
  epochs: epoch,
  lanes,
  termination: lastPlan ? { pass_permitted: !!lastPlan.pass_permitted, ceiling: !!lastPlan.ceiling, epoch_limit_reached: !!lastPlan.epoch_limit_reached, why: lastPlan.why || [], no_progress: noProgress } : null,
  gaps_outstanding: assembleEnv && Number.isInteger(assembleEnv.open_gaps) ? assembleEnv.open_gaps : null,
  blue_claims: blueEnv2 ? blueEnv2.claim_count : (blueEnv ? blueEnv.claim_count : null),
  infra_debts: infraDebts,
  petitions: petitionLog,
  halted,
  halt_opinion: haltOpinion,
  friction,
}

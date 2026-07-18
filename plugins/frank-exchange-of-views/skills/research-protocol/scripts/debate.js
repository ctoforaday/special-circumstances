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
//   Model ladder: omit `model` for keeper runs (inherits session model); sonnet for development;
//   haiku for smoke. NEVER change `model` OR `judgmentModel` on a resume — they change agent()
//   opts, bust the cache keys, and re-run completed rounds at full price.
//   Per-role split (efficiency doctrine: cheapen redundancy and mechanics, never judgment or
//   the adversary): `model` drives the BULK seats (frontier, blue lanes, red lenses, blue
//   responses); `judgmentModel` drives the JUDGMENT seats (blue-synthesize, red-merge,
//   lead-judge, assemble) and defaults to INHERIT-SESSION, not to `model` — so a dev run with
//   model=sonnet still gets full-strength judgment unless you explicitly cheapen it too.
//   KNOWN TRADEOFF (retrospective §3 row 16b): red LENSES ride the bulk tier — on a cheap-model
//   dev/smoke run, treat lens-sourced gap grades with a confidence discount. For keeper runs,
//   omit `model` entirely so the adversary runs at full strength, per the doctrine.
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
const { topic, runDir, lanes = 3, maxRounds = 12, model = null, judgmentModel = null, laneFloorOverride = null, toolsDir = null } = a
if (!topic || !runDir || String(runDir).includes('undefined') || String(topic) === 'undefined') {
  throw new Error(`debate: refusing dispatch — topic/runDir unbound (topic=${JSON.stringify(topic)}, runDir=${JSON.stringify(runDir)})`)
}
// Lane floor (retrospective §3 row 7): run 2 silently ran under-provisioned at lanes=2.
// Below 3 lanes a hypothesis loses dedicated attention; override requires a stated reason
// (e.g. laneFloorOverride: 'smoke run — pipeline exercise only').
if (lanes < 3 && !laneFloorOverride) {
  throw new Error(`debate: lanes=${lanes} is below the floor of 3 — pass laneFloorOverride: '<reason>' to run under-provisioned deliberately`)
}
const bulk = model ? { model } : {}
const judgment = judgmentModel ? { model: judgmentModel } : {}
const slug = String(runDir).replace(/[\/]+$/, '').split(/[\/]/).pop().replace(/^[0-9-]+_/, '')
log(`researching: ${topic.length > 160 ? topic.slice(0, 157) + '...' : topic}`)

// Friction must survive a mid-run throw (retrospective §3 row 24): the envelope copy feeds
// this script's aggregate, the file copy survives an abort. Both, always.
const frictionClause = (who) => ` FRICTION: if you report any friction, ALSO append each entry to ${runDir}/friction.md as one attributed line ("${who}: ...") in addition to the envelope's friction field, so the complaint survives even if a later phase aborts.`

// Wall-clock doctrine (run-4 forensics, 2026-07-17): 80% of run time is API rounds at ~24s
// each, and the corpus showed ZERO batched tool calls — every peek paid a full round. A
// Bash spawn additionally carries a measured multi-second fixed floor. Every seat gets this.
const speedClause = ` SPEED: every message you send costs a ~20s round-trip regardless of content — batch INDEPENDENT tool calls into a single message (read several files at once; fire independent fetches together); only serialize calls that truly depend on a prior result. Peek and search files with the native Read (offset/limit), Grep, and Glob tools — NEVER sed/awk/head/tail/cat/grep through Bash for file access (a shell spawn costs 10-100x a native read and buys nothing). KNOWN HARNESS LIMIT (W1.11, on file three times — do NOT re-log it as friction): Glob/Grep may refuse paths outside the session's registered working directories ("Path does not exist") while Read and Bash reach them; for searches under the run directory the SANCTIONED fallback is Bash grep/ls — this is the one exception to the no-shell-file-access rule.`

// Record-tool dual-mode (plan §III R2, gated on toolsDir so old runs resume
// untouched): every seat gets an engine-assigned SEAT_ID and records each act
// through its seat CLI IN ADDITION to the file writes. Hand-written artifacts
// stay authoritative until the R2.5 parity gate passes; the events are the
// record under test. The tool's --help is the seat's record contract.
const recordClause = (seatId, tool) => toolsDir
  ? ` SEAT_ID: ${seatId}. RECORD (dual-mode): in ADDITION to the file writes above, record every action through your seat tool. FIRST ACTION: node "${toolsDir}/${tool}.mjs" register --run ${runDir} --seat-id ${seatId} — then the matching verb for each act (see --help, your record contract; findings/cites at lenses, mint/close/dispose/spot-check/position/closing/verdict at the merge, revision/manifest-row/dispute/position/closing at blue, opinion/certify/friction at the bench; prose payloads over 2KB via --file, NEVER inline). The hand-written files remain authoritative this run; the events are the record under test.`
  : ''

// Compound grades allowed: red's protocol grades finer than a 3-point scale
// (retrospective friction #6 — forced rounding lost information every round).
const GRADE = { type: 'string', enum: ['low', 'low-medium', 'medium', 'medium-high', 'high', 'certain', 'realized', 'trivial'] }

// §8 Q6 pinned mass mapping (run-4 report §2.5 item 1) — TOTAL over the GRADE enum.
// `realized` is EXCLUDED from mass (a realized risk is no longer a probability: it
// contributes 0 and is counted separately in realized_open); `trivial` is assigned, not left
// to seat convention. Changing ANY value bumps the version and starts a NEW telemetry
// series — cross-version comparison never enters an actuation case.
const MASS_MAPPING_VERSION = 'v1'
const MASS = { trivial: 0.5, low: 1, 'low-medium': 1.5, medium: 2, 'medium-high': 2.5, high: 3, certain: 3.5, realized: 0 }
const gapMass = (g) => (MASS[g.likelihood] ?? 0) * (MASS[g.impact] ?? 0)

// Grade-dispute channel constants (run-4 report §3.3, clauses (v) and (vii)):
// per-round dispute cap with overflow batch-docketed as ONE judge item, and the
// script-computed cumulative accepted-delta magnitude (in mapping units) that
// batch-dockets accepted deflation/inflation for judge review before it stands.
const DISPUTE_CAP = 5
const ACCEPTED_DELTA_DOCKET_THRESHOLD = 2

const DISPUTE_DIMENSION = { type: 'string', enum: ['severity', 'likelihood', 'impact', 'complexity_cost'] }

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
          // halt = the safety boundary: the run ends HALTED, the opinion is
          // relayed verbatim by capture, never smoothed.
          ruling: { type: 'string', enum: ['granted', 'denied', 'halt'] },
          opinion: { type: 'string' },
        },
      },
    },
  },
}
// W2e — the bench reads law at every sitting. Precedent is ARGUMENT, never
// evidence: the leaf always wins a conflict (flag it for human review). A
// precedent either party cites MUST be addressed in the opinion. PERSUASIVE
// holdings persuade; only AFFIRMED ones bind — the bench proposes law, it
// never enacts it (capture harvests this run's rulings for human review).
const lawClause = ` LAW: if ${runDir}/inputs/law/ exists, read it before ruling — statute > precedent > case-local argument; PRECEDENT IS ARGUMENT, NOT EVIDENCE (the artifact and the leaf are the only evidence; where a precedent conflicts with the leaf, the leaf wins and you flag the conflict for human review); a precedent either party cites MUST be addressed in your opinion; PERSUASIVE holdings persuade, only AFFIRMED ones bind; your rulings this run are harvested as PERSUASIVE proposals for human review — you propose law, you never enact it.`
const petitionClause = (who) => ` PETITIONS: if fulfilling this seat's instructions would require asserting what you believe false, burying a real finding, or papering over a safety or ethics hazard, you may petition the bench via the envelope's petitions field (class: ethical|safety|integrity|constitutional, basis, relief) — heard BEFORE the debate continues, never sanctioned, and it does not pause your duties (attributed as ${who}).`

const BLUE_ENVELOPE = {
  type: 'object',
  required: ['path', 'tldr', 'claim_count', 'saturation_reached', 'open_questions', 'round_record_appended'],
  properties: {
    path: { type: 'string' },
    tldr: { type: 'string' },
    claim_count: { type: 'number' },
    saturation_reached: { type: 'boolean' },
    open_questions: { type: 'array', items: { type: 'string' } },
    // W1.7 round-parity attestation (run-5: blue's round-2 revision shipped with no ### BLUE
    // block or CHANGELOG entry; a lens misjudged the round state and the judge had to
    // reconstruct blue's position from red's records). TRUE only after debate.md carries this
    // round's "### BLUE" section AND CHANGELOG.md its round entry — the script ABORTS on
    // false (a revision is not on the record until the transcript carries it). Attestation
    // tier: shape in-run; capture's record-parity audit recounts post-hoc.
    round_record_appended: { type: 'boolean' },
    // W2b correctness manifest (repair-quality program A.2): one row per repaired gap —
    // the self-audit's receipt. The script requires a NON-EMPTY manifest on any respond
    // round with open gaps and logs coverage (manifest ids vs open-gap ids); completeness
    // is a capture-audited scorecard number, not an in-run throw (a partial manifest is a
    // visible defect, not a dead round).
    manifest: {
      type: 'array',
      items: {
        type: 'object',
        required: ['gap_id', 'row'],
        properties: {
          gap_id: { type: 'string' },
          // The row states what the self-audit DID for this repair: figures recomputed,
          // universals enumerated, consistency sites swept (named), boundary case asked,
          // composition noted, sibling sweep or declared-open enumeration, acceptance
          // check RUN with its result, new claims tagged.
          row: { type: 'string' },
        },
      },
    },
    friction: { type: 'array', items: { type: 'string' } },
    petitions: PETITIONS,
    // Grade-dispute channel (run-4 §3.3 — RATIFIED minimal form): blue's machine-readable
    // contest path against red's grades. Record-integrity insurance; zero expected savings.
    grade_disputes: {
      type: 'array',
      items: {
        type: 'object',
        required: ['gap_id', 'dimension', 'proposed', 'evidence'],
        properties: { gap_id: { type: 'string' }, dimension: DISPUTE_DIMENSION, proposed: GRADE, evidence: { type: 'string' } },
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
    gaps: {
      type: 'array',
      items: {
        type: 'object',
        required: ['id', 'location', 'problem', 'required_fix', 'acceptance_check', 'severity', 'likelihood', 'impact', 'complexity_cost'],
        properties: {
          id: { type: 'string' }, location: { type: 'string' }, problem: { type: 'string' },
          required_fix: { type: 'string' },
          // W2b acceptance-check-at-mint (repair-quality program A.1): the exact falsifiable
          // check red will run at re-audit — a probe command, a recompute, a quote-anchor.
          // Blue runs it BEFORE announcing; red's re-audit becomes a spot-audit of a
          // pre-agreed contract instead of a re-derivation.
          acceptance_check: { type: 'string' },
          severity: GRADE, likelihood: GRADE, impact: GRADE, complexity_cost: GRADE,
          // Lineage (retrospective §3 row 23): a successor gap MUST name the prior-round
          // gap id(s) it descends from, so the contested docket can follow regression chains.
          supersedes: { type: 'array', items: { type: 'string' } },
          // Capture-recapture input (run-4 §2.5 item 2): which lens seats found this gap.
          // Self-reported; auditable against the preserved per-lens candidate files, and any
          // actuation review re-derives a sample independently at a non-red seat.
          found_by: { type: 'array', items: { type: 'string' } },
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
          class: { type: 'string', enum: ['closed', 'closed_with_regression', 'rebuttal_accepted', 'risk_argued'] },
        },
      },
    },
    // Sharding observables (run-4 §4.5 conditions 5 and 7). archive_spot_checks: ids of
    // archived closures re-verified this round — required non-empty from round 2 (script
    // shape-check; the floor is never zero). The two counts ride as merge-reported integers;
    // the script's arithmetic comparison catches a self-inconsistent self-report ONLY — true
    // counts are audited post-hoc over the git-tracked shards (§6.2 attestation ceiling).
    archive_spot_checks: { type: 'array', items: { type: 'string' } },
    ledger_closure_lines: { type: 'number' },
    archive_blocks: { type: 'number' },
    // Grade-dispute responses (run-4 §3.3): every disputed gap_id×dimension from blue's last
    // envelope MUST be addressed; an unaddressed dispute is treated as REJECTED and
    // auto-docketed (default-to-docket punishes silence, not disagreement).
    dispute_responses: {
      type: 'array',
      items: {
        type: 'object',
        required: ['gap_id', 'dimension', 'response', 'rationale'],
        properties: { gap_id: { type: 'string' }, dimension: DISPUTE_DIMENSION, response: { type: 'string', enum: ['accepted', 'rejected'] }, rationale: { type: 'string' } },
      },
    },
    corroboration: {
      type: 'array',
      items: {
        type: 'object',
        required: ['claim', 'reference', 'confidence'],
        properties: {
          claim: { type: 'string' }, reference: { type: 'string' },
          confidence: { type: 'string', enum: ['high', 'medium', 'low'] },
        },
      },
    },
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
          // moot (run-4 coverage audit, GAP-35): the gap's predicate expired — the claim or
          // artifact it attached to no longer exists in the report. Run 4's judge had to
          // misuse `carried` for this live traffic class. Moot adjudicates the gap out.
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
// The ledger is the single source of truth for status; the archive is immutable closed prose.
const LEDGER = `${runDir}/red/ledger.md`
const ARCHIVE = `${runDir}/red/archive.md`
const TELEMETRY = `${runDir}/trajectories/board-telemetry.jsonl`

// W2c petition machinery: the log is the judicial record's petition section; a
// halt ruling ends the run (verdict HALTED — capture relays the opinion
// verbatim, never smoothed); granted relief is surfaced to subsequent seats.
const friction = [] // capability complaints from any agent, aggregated for /self-improve
const takeFriction = (who, env) => { if (env && env.friction) for (const f of env.friction) friction.push(`${who}: ${f}`) }
const petitionLog = []
const reliefInEffect = []
let halted = false
let haltOpinion = null
async function hearPetitions(env, who) {
  const petitions = (env && env.petitions) || []
  if (!petitions.length) return false
  log(`petition(s) filed by ${who} (${petitions.map((x) => x.class).join(', ')}) — bench sitting before the debate continues`)
  const sitting = await agent(
    `Petition sitting, topic "${topic}". ${who} has petitioned the bench: ${JSON.stringify(petitions)}. Petitions are heard BEFORE the debate continues; they are never sanctioned, and a pattern of overruled petitions is at most a craft note for the petitioner. For EACH petition rule granted (state the relief as it will bind the coming seats) | denied (with opinion) | halt (ONLY where continuing would compromise safety, consent gates, corpus integrity, or participant integrity — a halt ends the run and your opinion is relayed to the human verbatim). Every ruling is a written OPINION: the principle applied, the values in tension, and why a human should or should not look. Read ${runDir}/debate.md for context; append your rulings under "### LEAD (petitions)".${lawClause}${frictionClause('judge-petition')}${speedClause}${recordClause('judge-petition', 'bench')} Return the petition-ruling envelope.`,
    { ...judgment, label: `judge-petition · ${slug}`, phase: 'Debate', agentType: 'frank-exchange-of-views:lead-judge', schema: PETITION_RULING })
  if (!sitting) throw new Error('petition sitting returned null (agent failed) — a filed petition is never dropped; aborting cleanly')
  takeFriction('judge-petition', sitting)
  for (const r of sitting.rulings) {
    petitionLog.push({ petitioner: who, class: r.class, ruling: r.ruling, opinion: r.opinion })
    if (r.ruling === 'halt') { halted = true; haltOpinion = r.opinion }
    if (r.ruling === 'granted') reliefInEffect.push({ petitioner: who, opinion: r.opinion })
  }
  if (halted) log(`JUDICIAL HALT — ${haltOpinion ? haltOpinion.slice(0, 200) : ''}`)
  return halted
}
// ---- Frontier ----
phase('Frontier')
await agent(
  `Research debate opening for topic: "${topic}". Formulate 3-5 frontier hypotheses (what would be true if each candidate answer were right) per the research protocol, and write them to ${runDir}/blue/frontier.md.${speedClause}${recordClause('frontier', 'blue')} Return one line per hypothesis.`,
  { ...bulk, label: `frontier · ${slug}`, agentType: 'frank-exchange-of-views:blue-researcher' })

// ---- Blue: best-of-N lanes with method diversity, then additive synthesis ----
phase('Blue')
await parallel(Array.from({ length: lanes }, (_, i) => () => agent(
  `Blue lane ${i + 1} of ${lanes} for topic: "${topic}". Read ${runDir}/blue/frontier.md, and if ${runDir}/inputs/red-gap-patterns.md exists (red's accumulated gap-pattern inventory, staged at run setup), read it too — yesterday's expensive red discovery is today's free checklist line. Research your assigned slice to saturation per the research protocol (spend at least one search in five on disconfirming evidence; semantic footnotes with access dates). Your assigned METHOD LENS: ${LANE_METHODS[i % LANE_METHODS.length]} — work primarily through this method's source class; take hypothesis ${i + 1} first, then breadth. FOOTNOTE NAMESPACE: prefix every footnote label you mint with your lane marker (e.g. [^L${i + 1}CaptureRecapture]) — lanes share no bibliography and unprefixed labels collide at synthesis. Write your full candidate draft to ${runDir}/blue/candidates/lane-${i + 1}.md.${speedClause}${recordClause(`blue-lane-${i + 1}`, 'blue')} Return a 3-line synopsis.`,
  { ...bulk, label: `blue-lane-${i + 1} · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-researcher' })))

let blueEnv = await agent(
  `Blue synthesis for topic: "${topic}". PRE-FLIGHT: if ${runDir}/inputs/red-gap-patterns.md exists, read it before merging — check your synthesis against red's known gap patterns. Read every draft in ${runDir}/blue/candidates/ and synthesize ${runDir}/blue/report.md by UNION: deduplicate overlapping claims, reorganize, and never drop substantive content — structural merge (append + dedup), not a free-form rewrite. CLAIM PROVENANCE (cheap manifest): while merging, tag every claim that appears in exactly ONE lane's draft with its lane marker (e.g. "[minority: lane-2/practitioner]") — single-lane claims are minority reports red must weigh differently from convergent ones; the set-difference exists transiently in this merge and must not be discarded. Lane footnote labels arrive lane-prefixed; when two lanes cite the SAME source, merge to one label and note both lanes. Follow the report conventions (semantic footnotes with access dates). CLAIM UNIT (pinned — two honest merges differed 2x without it): claim_count = the number of FOOTNOTED declarative claims (a sentence carrying at least one footnote marker counts once; unfootnoted prose counts zero; a multi-footnote sentence still counts once). Start ${runDir}/blue/CHANGELOG.md with a Round 0 entry describing the synthesis and stating claim_count (the tracked copy of the envelope figure). round_record_appended: set TRUE only after the CHANGELOG Round 0 entry exists on disk. ${petitionClause('blue-synthesize')}${frictionClause('blue-synthesize')}${speedClause}${recordClause('blue-synthesize', 'blue')} Return the blue envelope.`,
  { ...judgment, label: `blue-synthesize · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-researcher', schema: BLUE_ENVELOPE })

if (!blueEnv) throw new Error('blue synthesis returned null (agent failed) — aborting cleanly')
if (blueEnv.round_record_appended !== true) throw new Error('round-parity (W1.7): blue-synthesize did not attest the CHANGELOG Round 0 entry — the synthesis is not on the record')
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
takeFriction('blue-synthesize', blueEnv)

while (!halted && round < maxRounds) {
  round++
  // W1.8: the spot-check floor keys on round-START archive state, not the round number —
  // run 5's round-2 merge entered with ZERO archive blocks (round 1 closed nothing), so the
  // old from-round-2 floor degraded to same-seat self-attestation of blocks it was itself
  // about to write. prev round's self-reported count is the script-visible proxy.
  const prevArchiveBlocks = redEnv ? (redEnv.archive_blocks || 0) : 0
  const prevGaps = redEnv ? redEnv.gaps : []
  log(`round ${round}: dispatching ${Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count || 20) / 40))) + 2} red lenses over ${blueEnv.claim_count || '?'} claims`)

  // Citation verification scales with the CURRENT report size — recomputed every round
  // (retrospective §3 row 2b: computed once, later rounds were systematically under-scaled
  // exactly when the report was largest).
  const citationPasses = Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count || 20) / 40)))

  const lensPasses = []
  // Citation ledger: verified citations don't un-verify. Each citation pass reads the
  // ledger first and skips claims already graded high-confidence in a prior round — but the
  // skip-trigger covers SOURCE drift as well as prose drift (retrospective §3 row 10: the
  // prose-only trigger suppressed exactly the re-check that catches a source moving).
  const ledgerClause = ` CITATION LEDGER: read ${runDir}/red/citation-ledger.md first if it exists; a claim verified at HIGH confidence in a prior round stays verified — do not re-fetch it UNLESS ${runDir}/blue/CHANGELOG.md shows its section changed this round, OR more than 2 rounds have elapsed since it was last verified, OR its recorded access date and source volatility suggest drift (living documents, issue trackers, README stats). Append every claim you verify to the ledger (one line: claim | reference | confidence | round | access-date). MUST-TRY OBSERVABLE: every citation you grade DOWN (below high) MUST carry an attempt-or-impossibility line in your pass — which extraction tool or path you tried, or why none was triable; an untried "unable to corroborate" is an incomplete audit (run 4 caught a false "paywalled" on an open-access paper this clause would have exposed at round 0). LARGE SOURCES (W1.12): a truncated fetch is NOT a read — for GitHub issues/PRs default to \`gh issue view <n> --comments\` via Bash (WebFetch is lossy on threads, run-5 friction); if a source exceeds one fetch, fetch it in sections and say so; never grade from a truncated body without stating the truncation in your pass.`
  for (let c = 0; c < citationPasses; c++) {
    lensPasses.push(`${RED_LENSES[0]}${citationPasses > 1 ? ` — instance ${c + 1} of ${citationPasses}: divide the report's sections evenly among instances and take slice ${c + 1}; footnote-block ownership follows the slice (instance ${c + 1} owns the footnote definitions its sections reference)` : ''}.${ledgerClause}`)
  }
  lensPasses.push(RED_LENSES[1], RED_LENSES[2])

  await parallel(lensPasses.map((lens, i) => () => agent(
    `Red audit, round ${round}, lens: ${lens}. Re-read the FULL living report ${runDir}/blue/report.md in context (the whole document — never just a diff; if it exceeds one Read call, read it whole in consecutive windows)${round > 1 ? `; blue's change log ${runDir}/blue/CHANGELOG.md is a navigation hint only` : ''}. Anchor every finding to a section heading plus a quoted sentence. HARNESS NOTES: Grep count mode counts LINES, not occurrences — anchor patterns (e.g. '^### ') when counting; prefer the Write tool over quoted heredocs for scripts (heredoc backslash mangling is a documented recurrence). Label your findings with LENS-SCOPED ids (L${i + 1}-F1, L${i + 1}-F2, ...) — stable R${round}-N ids are assigned by the merge, never by a lens. Write your pass to ${runDir}/red/candidates/round-${round}-lens-${i + 1}.md. You MUST NOT write to ${runDir}/debate.md — only red-merge writes the round's "### RED" section.${speedClause}${recordClause(`red-lens-r${round}-L${i + 1}`, 'red-lens')} Return a 3-line synopsis.`,
    { ...bulk, label: `red-lens-${i + 1}-r${round} · ${slug}`, phase: 'Red', agentType: 'frank-exchange-of-views:red-auditor' })))

  redEnv = await agent(
    `Red merge, round ${round}. FIRST ACTION (read batching — run-4 §4.6): concatenate this round's lens passes with a single bash call, \`cat ${runDir}/red/candidates/round-${round}-lens-*.md > <your session scratchpad>/round-${round}-all.md\` — the output path MUST be the ABSOLUTE path of your own session scratchpad directory, never under ${runDir} (a stray file in red/candidates/ corrupts every downstream glob and convergence count) — then read that one file instead of ${lensPasses.length} separate reads.
SHARDED FINDINGS (run-4 §4 — the ledger is the single source of truth for status; the archive is immutable closed prose): maintain ${LEDGER} (OPEN gaps with full grading, plus a compact CLOSURE INDEX — one line per closed gap: id | closure class | one-line summary | supersedes) and ${ARCHIVE} (append-only full prose record of each closed gap: what was found, how verified, closure class). ${round === 1 ? 'ROUND 1: create BOTH files with Write (this seat creates them — the skeleton deliberately does not; the names are write-guard-verified).' : `Update the ledger in place; append closures to the archive; NEVER edit an existing archive block.`} Every gap's location = section heading + quoted sentence; keep graded likelihood/impact/complexity + severity on every open gap. NEAR-MATCH RULE: before minting ANY fresh gap id, scan the closure index — on a near-match, read that gap's full archive record FIRST (targeted read; the index screens, it never decides): the candidate is then a reopen (supersedes the closed id) or genuinely new, and you say which in the gap record. DEMANDED READS: any lineage or closure claim you assert (in the ledger, the docket, or rebutting blue) MUST be verified against the archive record by targeted read, and re-verify at least ${round >= 2 ? 'one' : 'zero'} archived closure(s) this round sampled at your discretion, recording the sampled ids in the envelope's archive_spot_checks (required non-empty whenever the archive ENTERED this round with at least one record — the floor keys on round-START archive state, never the round number; an archive that was empty at round start has nothing to cross-round sample and an honest empty array says so; reopen any sampled closure whose evidence has drifted, and archived closures citing volatile living sources inherit the citation ledger's drift triggers). COUNTS: report ledger_closure_lines (closure-index line count) and archive_blocks (archive record count) in the envelope — they must match. PROBE CLASSES (W1.10): any required_fix that demands a probe MUST class it — DOCUMENT-PROBE (executable now against shipped artifacts: a read, diff, version check) or LIVE-PROBE (requires built artifacts; in a design-phase debate it is DEFERRABLE and is discharged by naming it as an acceptance test with its pass condition). An unclassed probe demand risks blue overclaiming a file read as a probe or stalling on an impossible obligation.
FOUND_BY: on every gap, record which lens seats surfaced it (found_by: ["L1","L4",...]) — auditable against the candidate files.
Gap ids are stable across rounds (R1-1 stays R1-1); assign fresh R${round}-N ids to genuinely new gaps only. LINEAGE IS MANDATORY: when you close a gap WITH REGRESSION and mint a successor, the successor gap's "supersedes" array MUST name the closed gap's id, and your envelope's "closures" array MUST record the closure with class "closed_with_regression" — the docket detector follows these chains and an undeclared lineage is a protocol violation the script rejects.${adjudicated.length ? ` Gaps already adjudicated by the lead-judge and EXCLUDED from your verdict: ${JSON.stringify(adjudicated.map(x => x.gap_id))}.` : ''}${gradeAdjustments.length ? ` GRADE ADJUSTMENTS RULED BY THE JUDGE last round (apply each in the ledger, and list the delta in your "### RED" entry): ${JSON.stringify(gradeAdjustments)}.` : ''}${pendingDisputes.length ? ` BLUE'S GRADE DISPUTES from last round — you MUST answer EVERY one in the envelope's dispute_responses (accepted or rejected, with rationale; an unaddressed dispute is treated as rejected and auto-docketed to the judge): ${JSON.stringify(pendingDisputes)}. For each ACCEPTED dispute: apply the new grade in the ledger AND list the delta (gap id, dimension, old -> new) in your "### RED" entry — pending deltas are watched there by blue, the judge, and the operator.` : ''}
BOARD TELEMETRY (run-4 §2.5 — the signal the stopping judgment reads): append ONE JSON line to ${TELEMETRY} (bash: cat >> — the file is git-tracked): {"round": ${round}, "mapping_version": "${MASS_MAPPING_VERSION}", "open_count", "max_severity", "new_mint": {"count", "by_severity"}, "mass", "accepted_deltas": [...], "realized_open", "excluded_mass_memo", "found_by_summary", "repair_regression": {"closures", "lineage_mints", "ratio"}, "edge_deltas": {"down_mass", "up_mass"}} — mass = sum over OPEN unadjudicated gaps of L×I under the pinned mapping ${JSON.stringify(MASS)} (realized contributes 0 and is counted in realized_open; list any excluded gaps in excluded_mass_memo). repair_regression (W2b, blue's headline scorecard number — baseline 0.37-0.72): closures = prior-round gaps you closed this round; lineage_mints = fresh gaps whose supersedes names a gap closed THIS round; ratio = lineage_mints/closures (null when closures=0). edge_deltas (W2b — the accepted-delta blind spot: ~20 mass moved per run along supersedes edges the dispute-channel accounting cannot see): for each supersedes edge minted this round, mass(successor) - mass(ancestor) under the pinned mapping; report the summed magnitudes split by sign. This line is the convenience copy, never the evidence of record — grades of record live in the ledger; a pending-window dispute delta is EXPECTED divergence, carried in the line's own delta record. GAP RECORDS: every gap you mint carries acceptance_check — the exact falsifiable check you will run at re-audit (probe command, recount command, quote-anchor); a fix-spec naming instances states the class-closure rule or declares the enumeration open (the sweep clause).
Decide the binary verdict — PASS only when every remaining unadjudicated gap is closed, evidence-rebutted, or risk-accepted. Append the round-${round} "### RED" section to ${runDir}/debate.md per the debate template. CLOSING ARGUMENTS: any gap you RE-RAISE from a prior round, any successor you mint via supersedes, and any grade dispute you REJECT is docket-bound — for EACH such item, also append a "### RED CLOSING (round ${round})" entry to ${runDir}/debate.md — max 120 words per item: your strongest evidence the gap is real and graded correctly, and your answer to blue's best rebuttal so far. The judge rules AFTER blue responds this round, on the closings, the transcript, and the final artifact state only — your closing is your case; overstatement the record does not support counts against you.${petitionClause(`red-merge-r${round}`)}${frictionClause(`red-merge-r${round}`)}${speedClause}${recordClause(`red-merge-r${round}`, 'red-merge')} Return the red envelope.`,
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
  for (const c of redEnv.closures || []) {
    if (c.class === 'closed_with_regression' && !redEnv.gaps.some(g => (g.supersedes || []).includes(c.id))) {
      throw new Error(`red-merge round ${round} closed gap ${c.id} WITH REGRESSION but no successor gap names it in supersedes — lineage silently dropped`)
    }
  }
  // Sharding observables (run-4 §4.5 conds 5+7): spot-check floor from round 2; count
  // arithmetic catches a self-inconsistent self-report (shape/consistency tier only —
  // true counts are audited post-hoc over the git-tracked shards).
  if (round >= 2 && prevArchiveBlocks > 0 && (!redEnv.archive_spot_checks || redEnv.archive_spot_checks.length === 0)) {
    throw new Error(`red-merge round ${round} reported no archive spot-checks with ${prevArchiveBlocks} archived closure(s) on the board at round start — the floor is never zero when there is something to sample (run-4 §4.5 condition 5; W1.8)`)
  }
  if (typeof redEnv.ledger_closure_lines === 'number' && typeof redEnv.archive_blocks === 'number' && redEnv.ledger_closure_lines !== redEnv.archive_blocks) {
    throw new Error(`red-merge round ${round}: closure-index lines (${redEnv.ledger_closure_lines}) != archive blocks (${redEnv.archive_blocks}) — self-inconsistent shard self-report`)
  }

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
  const hasNew = redEnv.gaps.some(g => !allPriorGapIds.has(g.id) && !(g.supersedes || []).some(id => allPriorGapIds.has(id)))
  for (const g of redEnv.gaps) allPriorGapIds.add(g.id)

  const adjudicatedIds = new Set(adjudicated.map(x => x.gap_id))
  const openGaps = redEnv.gaps.filter(g => !adjudicatedIds.has(g.id))
  blueEnv = await agent(
    `Blue response, round ${round}, topic "${topic}". Red's verdict: FAIL.${reliefInEffect.length ? ` BENCH RELIEF IN EFFECT (binding): ${JSON.stringify(reliefInEffect)}.` : ''} FIRST ACTION (read batching, W1.12): concatenate your working set with a single bash call — \`cat ${LEDGER} ${runDir}/debate.md ${runDir}/inputs/red-gap-patterns.md > <your session scratchpad>/respond-${round}-workset.md\` (ABSOLUTE scratchpad path, never under ${runDir}) — then read that one file instead of three separate reads; report.md you edit in place as usual. PRE-FLIGHT: re-check your planned repairs against red's gap patterns (in the workset). Open gaps (adjudicated ones excluded): ${JSON.stringify(openGaps)}. Corroboration flags: ${JSON.stringify(redEnv.corroboration || [])}. BEFORE drafting, read the latest "### RED" section of ${runDir}/debate.md — the gap JSON above is a lossy summary of the transcript, and it lists any accepted grade-dispute deltas pending their contest window — and the latest "### LEAD" section if one exists: any gap the judge CARRIED comes with a stated research direction you owe. Address every open gap ADDITIVELY in ${runDir}/blue/report.md — expand and repair where red is right, rebut in writing (with evidence) where red is wrong, and argue risk-acceptance where the fix's complexity exceeds its likelihood x impact; never subtract substance, and propagate every correction to ALL sites that state the corrected claim, not only the flagged sentence (incomplete propagation was run 3's dominant blue failure class — 5 regressions in 5 rounds; grep the corrected strings/figures report-wide and log the sites checked in the CHANGELOG).${contested.length ? ` CLOSING ARGUMENTS: the following items are DOCKETED for adjudication AFTER your response this round: ${JSON.stringify(contested)}. For EACH docketed item, after your repairs, append a "### BLUE CLOSING (round ${round})" entry to ${runDir}/debate.md — max 120 words per item: why your response resolves it, or why red's grade/claim is wrong, citing the exact section and evidence. The judge rules on the closings, the transcript, and the final state of the artifacts — your closing is your case; material not in the record cannot help you.` : ''} GRADE DISPUTES (your machine-readable contest path on red's grading): you MAY dispute a gap's severity/likelihood/impact/complexity_cost via the envelope's grade_disputes (gap_id, dimension, proposed grade, evidence) — max ${DISPUTE_CAP} per round (overflow is batch-docketed to the judge as one item)${heldDisputes.size ? `; disputes red REJECTED last round (re-dispute any of these to send it to the judge): ${JSON.stringify([...heldDisputes.values()])}` : ''}. PROBE CLASSES (W1.10): where a fix demands a probe, discharge a DOCUMENT-PROBE now (against shipped artifacts) or, for a LIVE-PROBE needing built artifacts, name it as a deferred acceptance test with its pass condition — never call a file read a live probe. THE MANIFEST (W2b — your self-audit's receipt; your constitution carries the full eight rows): for EVERY gap you repair, run the correctness manifest — figures recomputed, universals enumerated, consistency sites swept report-wide, the repair's own boundary case asked, compositions noted where edits share text, sibling sweep or declared-open enumeration, the gap's acceptance_check RUN with its result, new claims tagged verified-at-leaf/derived/asserted — and return one manifest row per repaired gap in the envelope's manifest array ({gap_id, row}: what you checked and what it showed, compressed). An unmanifested repair is unaudited by your own standard; the script rejects an EMPTY manifest on a round with open gaps, and coverage is scored at capture. CLAIM UNIT (pinned): claim_count = FOOTNOTED declarative claims (a footnote-bearing sentence counts once; unfootnoted prose counts zero). Log edits to ${runDir}/blue/CHANGELOG.md (Round ${round}, including claim_count); append your "### BLUE" section for round ${round} to ${runDir}/debate.md. ROUND RECORD (W1.7): round_record_appended = TRUE only after BOTH the "### BLUE" round-${round} section in debate.md AND the CHANGELOG Round ${round} entry exist on disk — the script aborts on false; a revision is not on the record until the transcript carries it (the round-2 desync misled a lens and blinded the judge, run 5).${petitionClause(`blue-respond-r${round}`)}${frictionClause(`blue-respond-r${round}`)}${speedClause}${recordClause(`blue-respond-r${round}`, 'blue')} Return the blue envelope.`,
    { ...bulk, label: `blue-respond-r${round} · ${slug}`, phase: 'Debate', agentType: 'frank-exchange-of-views:blue-researcher', schema: BLUE_ENVELOPE })
  takeFriction(`blue-respond-r${round}`, blueEnv)
  if (!blueEnv) throw new Error(`blue response round ${round} returned null (agent failed) — aborting cleanly`)
  if (blueEnv.round_record_appended !== true) throw new Error(`round-parity (W1.7): blue-respond round ${round} revised the report without attesting the "### BLUE" section + CHANGELOG entry — a revision is not on the record until the transcript carries it`)
  if (openGaps.length > 0 && (!Array.isArray(blueEnv.manifest) || blueEnv.manifest.length === 0)) {
    throw new Error(`manifest (W2b): blue-respond round ${round} repaired ${openGaps.length} gap(s) with an EMPTY correctness manifest — an unmanifested repair is unaudited by blue's own standard`)
  }
  if (openGaps.length > 0) {
    const covered = new Set((blueEnv.manifest || []).map((m) => m.gap_id))
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
      `Adjudication, round ${round}, topic "${topic}". Contested docket: ${JSON.stringify(contested)}. Both sides have filed closings for this docket: red's "### RED CLOSING (round ${round})" and blue's "### BLUE CLOSING (round ${round})" in ${runDir}/debate.md. YOUR RULING BASIS IS CONFINED TO: (1) the two closings, (2) the full transcript ${runDir}/debate.md, and (3) the final state of the artifacts — ${LEDGER} and ${runDir}/blue/report.md as they now stand. Weigh the closings as each side's best case; a claim in a closing that the record does not support counts AGAINST the side that made it.${lawClause} Resolution set (full, for every gap class — blue has answered everything docketed): closed (blue's response resolves it) | rebuttal_sustained (blue's evidence beats the challenge) | risk_accepted (valid, rejected on likelihood x impact x complexity) | carried (still live — state what further research blue owes) | unresolved | moot (the predicate expired — the claim or artifact it attached to no longer exists in the report) | grade_adjusted (for grade_dispute_* items: state the corrected grade in the rationale; red applies it next round) | routed_to_infrastructure (valid finding whose FIX is owned outside the debate — run tooling, harness, or the lead; state the owed fix in the rationale; it leaves red's verdict pool and ships as a named infrastructure debt, recorded never dropped). DEMANDED READS: for every ruling on a gap with a supersedes chain, you MUST read the named ancestors' records in ${ARCHIVE} first and NAME the records read in your rationale. deadlock is true only if no gap is carried AND ${hasNew ? 'false (new gaps were raised this round)' : 'no new gaps were raised this round (none were)'}. Append your "### LEAD" resolutions to ${runDir}/debate.md.${frictionClause(`judge-r${round}`)}${speedClause}${recordClause(`judge-r${round}`, 'bench')} Return the judge envelope.`,
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
const verdict = halted ? 'HALTED' : redEnv && redEnv.verdict === 'PASS' ? 'VERIFIED' : 'UNVERIFIED'
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
    `Terminal dispute disposition for topic "${topic}" (debate ended ${verdict} after round ${round}; this docket fires at the exit boundary). Undisposed grade disputes: ${JSON.stringify(terminalDisputes)}. Read ${runDir}/debate.md and ${LEDGER} in full. ${lawClause} The resolution set at a terminal exit EXCLUDES carried — rule each dispute grade_adjusted (state the corrected grade in the rationale; assembly records the delta) or unresolved (the contested grade ships, recorded as contested in the report). deadlock: false. Append your "### LEAD (terminal disputes)" ruling to ${runDir}/debate.md.${frictionClause('judge-terminal')}${speedClause}${recordClause('judge-terminal', 'bench')} Return the judge envelope.`,
    { ...judgment, label: `judge-terminal · ${slug}`, phase: 'Assemble', agentType: 'frank-exchange-of-views:lead-judge', schema: JUDGE_ENVELOPE })
  if (terminalJudge) {
    takeFriction('judge-terminal', terminalJudge)
    for (const r of terminalJudge.resolutions) {
      if (r.resolution === 'grade_adjusted') gradeAdjustments.push({ gap_id: r.gap_id, rationale: r.rationale })
    }
  }
}

// ---- Assemble: union, not summary ----
phase('Assemble')
await agent(
  `Final assembly for topic "${topic}", run directory ${runDir}. Debate outcome: ${verdict} after ${round} round(s)${deadlocked ? ' by judged deadlock' : ''}${exhausted ? ' by safety ceiling' : ''}. Assemble ${runDir}/report.md by UNION per the report template (references/report_template.md): verdict stamp, TL;DR, the Catechism (references/catechism_template.md — the AGREED answers: the case against at full strength, of-interest-vs-merely-interesting, cost and stopping points), technical foundations, analysis, graded risk matrix (including risk_accepted items with rationale), then blue/report.md IN FULL, red's board IN FULL (${LEDGER} then ${ARCHIVE}), per-round debate synopsis pointing at debate.md, an "Open questions carried past this run" section from blue's final envelope: ${JSON.stringify((blueEnv && blueEnv.open_questions) || [])}, and the consolidated footnotes. Never compress the research into a digest. ${verdict === 'UNVERIFIED' ? 'Stamp UNVERIFIED and list every outstanding gap with its disposition and the compromise rationale.' : ''}${verdict === 'HALTED' ? ` Stamp HALTED at the top: the bench ended this run. Quote the halt opinion IN FULL and VERBATIM — it is relayed to the human, never smoothed: ${JSON.stringify(haltOpinion)}.` : ''}${petitionLog.length ? ` JUDICIAL RECORD — petitions (list each with its ruling and opinion, never dropped): ${JSON.stringify(petitionLog)}.` : ''}${gradeAdjustments.length ? ` Terminal grade adjustments ruled by the judge (record each delta in the risk matrix): ${JSON.stringify(gradeAdjustments)}.` : ''}${infraDebts.length ? ` INFRASTRUCTURE DEBTS ruled routed_to_infrastructure (W1.9 — the lead's owed fixes; list them in their own report section, never dropped): ${JSON.stringify(infraDebts)}.` : ''} Collated friction so far (report any of your own as well): ${JSON.stringify(friction)}.${frictionClause('assemble')}${speedClause}${recordClause('assemble', 'bench')} Return a 5-line synopsis of the final report, plus your own friction if any.`,
  { ...judgment, label: `assemble · ${slug}`, agentType: 'frank-exchange-of-views:lead-judge' })

return {
  runDir,
  verdict,
  rounds: round,
  lanes,
  deadlocked,
  gaps_outstanding: redEnv && redEnv.verdict !== 'PASS' ? redEnv.gaps.length : 0,
  blue_claims: blueEnv ? blueEnv.claim_count : null,
  infra_debts: infraDebts,
  petitions: petitionLog,
  halted,
  halt_opinion: haltOpinion,
  friction,
}

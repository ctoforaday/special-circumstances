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
// The lead is a script: mechanics, round-keeping, termination. All file writes
// belong to the agents (the filesystem is the blackboard; the script has no
// filesystem access by design). Judgment calls go to lead-judge, never round-to-round.
// Defensive arg handling: args may arrive JSON-encoded (resume path); a stringified
// object destructures to undefined and every agent gets literal 'undefined' paths —
// the exact harness defect red graded high/high/low in run 1. Parse, then guard.
const a = typeof args === 'string' ? JSON.parse(args) : args
const { topic, runDir, lanes = 3, maxRounds = 12, model = null, judgmentModel = null, laneFloorOverride = null } = a
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

// Compound grades allowed: red's protocol grades finer than a 3-point scale
// (retrospective friction #6 — forced rounding lost information every round).
const GRADE = { type: 'string', enum: ['low', 'low-medium', 'medium', 'medium-high', 'high', 'certain', 'realized', 'trivial'] }

const BLUE_ENVELOPE = {
  type: 'object',
  required: ['path', 'tldr', 'claim_count', 'saturation_reached', 'open_questions'],
  properties: {
    path: { type: 'string' },
    tldr: { type: 'string' },
    claim_count: { type: 'number' },
    saturation_reached: { type: 'boolean' },
    open_questions: { type: 'array', items: { type: 'string' } },
    friction: { type: 'array', items: { type: 'string' } },
  },
}

const RED_ENVELOPE = {
  type: 'object',
  required: ['verdict', 'gaps', 'citations_checked'],
  properties: {
    verdict: { type: 'string', enum: ['PASS', 'FAIL'] },
    gaps: {
      type: 'array',
      items: {
        type: 'object',
        required: ['id', 'location', 'problem', 'required_fix', 'severity', 'likelihood', 'impact', 'complexity_cost'],
        properties: {
          id: { type: 'string' }, location: { type: 'string' }, problem: { type: 'string' },
          required_fix: { type: 'string' },
          severity: GRADE, likelihood: GRADE, impact: GRADE, complexity_cost: GRADE,
          // Lineage (retrospective §3 row 23): a successor gap MUST name the prior-round
          // gap id(s) it descends from, so the contested docket can follow regression chains.
          supersedes: { type: 'array', items: { type: 'string' } },
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
          resolution: { type: 'string', enum: ['closed', 'rebuttal_sustained', 'risk_accepted', 'carried', 'unresolved'] },
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

// ---- Frontier ----
phase('Frontier')
await agent(
  `Research debate opening for topic: "${topic}". Formulate 3-5 frontier hypotheses (what would be true if each candidate answer were right) per the research protocol, and write them to ${runDir}/blue/frontier.md. Return one line per hypothesis.`,
  { ...bulk, label: `frontier · ${slug}`, agentType: 'frank-exchange-of-views:blue-researcher' })

// ---- Blue: best-of-N lanes with method diversity, then additive synthesis ----
phase('Blue')
await parallel(Array.from({ length: lanes }, (_, i) => () => agent(
  `Blue lane ${i + 1} of ${lanes} for topic: "${topic}". Read ${runDir}/blue/frontier.md; research your assigned slice to saturation per the research protocol (spend at least one search in five on disconfirming evidence; semantic footnotes with access dates). Your assigned METHOD LENS: ${LANE_METHODS[i % LANE_METHODS.length]} — work primarily through this method's source class; take hypothesis ${i + 1} first, then breadth. Write your full candidate draft to ${runDir}/blue/candidates/lane-${i + 1}.md. Return a 3-line synopsis.`,
  { ...bulk, label: `blue-lane-${i + 1} · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-researcher' })))

let blueEnv = await agent(
  `Blue synthesis for topic: "${topic}". Read every draft in ${runDir}/blue/candidates/ and synthesize ${runDir}/blue/report.md by UNION: deduplicate overlapping claims, reorganize, and never drop substantive content — structural merge (append + dedup), not a free-form rewrite. CLAIM PROVENANCE (cheap manifest): while merging, tag every claim that appears in exactly ONE lane's draft with its lane marker (e.g. "[minority: lane-2/practitioner]") — single-lane claims are minority reports red must weigh differently from convergent ones; the set-difference exists transiently in this merge and must not be discarded. Follow the report conventions (semantic footnotes with access dates). Start ${runDir}/blue/CHANGELOG.md with a Round 0 entry describing the synthesis.${frictionClause('blue-synthesize')} Return the blue envelope.`,
  { ...judgment, label: `blue-synthesize · ${slug}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-researcher', schema: BLUE_ENVELOPE })

if (!blueEnv) throw new Error('blue synthesis returned null (agent failed) — aborting cleanly')
// ---- Debate loop: red audits gate; termination is judged, never counted ----
let round = 0
let redEnv = null
let deadlocked = false
const allPriorGapIds = new Set() // every gap id from every prior round — the docket window is the whole debate, not one round
const adjudicated = [] // judge-ruled gaps (closed / rebuttal_sustained / risk_accepted) — out of red's verdict
const friction = [] // capability complaints from any agent, aggregated for /self-improve
const takeFriction = (who, env) => { if (env && env.friction) for (const f of env.friction) friction.push(`${who}: ${f}`) }
takeFriction('blue-synthesize', blueEnv)

while (round < maxRounds) {
  round++

  // Citation verification scales with the CURRENT report size — recomputed every round
  // (retrospective §3 row 2b: computed once, later rounds were systematically under-scaled
  // exactly when the report was largest).
  const citationPasses = Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count || 20) / 40)))

  const lensPasses = []
  // Citation ledger: verified citations don't un-verify. Each citation pass reads the
  // ledger first and skips claims already graded high-confidence in a prior round — but the
  // skip-trigger covers SOURCE drift as well as prose drift (retrospective §3 row 10: the
  // prose-only trigger suppressed exactly the re-check that catches a source moving).
  const ledgerClause = ` CITATION LEDGER: read ${runDir}/red/citation-ledger.md first if it exists; a claim verified at HIGH confidence in a prior round stays verified — do not re-fetch it UNLESS ${runDir}/blue/CHANGELOG.md shows its section changed this round, OR more than 2 rounds have elapsed since it was last verified, OR its recorded access date and source volatility suggest drift (living documents, issue trackers, README stats). Append every claim you verify to the ledger (one line: claim | reference | confidence | round | access-date).`
  for (let c = 0; c < citationPasses; c++) {
    lensPasses.push(`${RED_LENSES[0]}${citationPasses > 1 ? ` — instance ${c + 1} of ${citationPasses}: divide the report's sections evenly among instances and take slice ${c + 1}` : ''}.${ledgerClause}`)
  }
  lensPasses.push(RED_LENSES[1], RED_LENSES[2])

  await parallel(lensPasses.map((lens, i) => () => agent(
    `Red audit, round ${round}, lens: ${lens}. Re-read the FULL living report ${runDir}/blue/report.md in context (the whole document — never just a diff)${round > 1 ? `; blue's change log ${runDir}/blue/CHANGELOG.md is a navigation hint only` : ''}. Anchor every finding to a section heading plus a quoted sentence. Label your findings with LENS-SCOPED ids (L${i + 1}-F1, L${i + 1}-F2, ...) — stable R${round}-N ids are assigned by the merge, never by a lens. Write your pass to ${runDir}/red/candidates/round-${round}-lens-${i + 1}.md. You MUST NOT write to ${runDir}/debate.md — only red-merge writes the round's "### RED" section. Return a 3-line synopsis.`,
    { ...bulk, label: `red-lens-${i + 1}-r${round} · ${slug}`, phase: 'Red', agentType: 'frank-exchange-of-views:red-auditor' })))

  redEnv = await agent(
    `Red merge, round ${round}. Read the round-${round} lens passes in ${runDir}/red/candidates/ and consolidate into the LIVING ${runDir}/red/findings.md (cumulative across rounds: update prior gaps' status, add new ones, keep graded corroboration and likelihood/impact/complexity on every gap; every gap's location = section heading + quoted sentence). Gap ids are stable across rounds (R1-1 stays R1-1); assign fresh R${round}-N ids to genuinely new gaps only. LINEAGE IS MANDATORY: when you close a gap WITH REGRESSION and mint a successor, the successor gap's "supersedes" array MUST name the closed gap's id, and your envelope's "closures" array MUST record the closure with class "closed_with_regression" — the docket detector follows these chains and an undeclared lineage is a protocol violation the script rejects.${adjudicated.length ? ` Gaps already adjudicated by the lead-judge and EXCLUDED from your verdict: ${JSON.stringify(adjudicated.map(a => a.gap_id))}.` : ''} Decide the binary verdict — PASS only when every remaining unadjudicated gap is closed, evidence-rebutted, or risk-accepted. Append the round-${round} "### RED" section to ${runDir}/debate.md per the debate template.${frictionClause(`red-merge-r${round}`)} Return the red envelope.`,
    { ...judgment, label: `red-merge-r${round} · ${slug}`, phase: 'Red', agentType: 'frank-exchange-of-views:red-auditor', schema: RED_ENVELOPE })

  takeFriction(`red-merge-r${round}`, redEnv)
  if (!redEnv) throw new Error(`red-merge round ${round} returned null (agent failed) — aborting cleanly`)
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
  if (redEnv.verdict === 'PASS') break

  // Contested docket: a dispute that persists — same id re-raised from ANY prior round
  // (window is the whole debate, not just last round), or a successor gap whose supersedes
  // chain descends from a prior-round gap (retrospective §3 row 23: regression chains ran
  // four generations in run 3 and the id-equality detector never armed; the judge was
  // dispatched ZERO times in the entire corpus). Detection is set arithmetic in the script;
  // the judgment belongs to the lead-judge.
  const contested = redEnv.gaps.filter(g =>
    allPriorGapIds.has(g.id) || (g.supersedes || []).some(id => allPriorGapIds.has(id)))
  const hasNew = redEnv.gaps.some(g => !allPriorGapIds.has(g.id) && !(g.supersedes || []).some(id => allPriorGapIds.has(id)))
  if (contested.length > 0) {
    const judge = await agent(
      `Adjudication, round ${round}, topic "${topic}". Contested docket (persisting disputes: re-raised ids and regression-chain successors): ${JSON.stringify(contested)}. New gaps were ${hasNew ? 'ALSO raised' : 'NOT raised'} this round. Read ${runDir}/debate.md and ${runDir}/red/findings.md in full. Rule per contested gap (closed | rebuttal_sustained | risk_accepted | carried | unresolved) with rationale — for carried, state what further research blue owes. deadlock is true only if no gap is carried AND no new gaps were raised. Append your "### LEAD" resolutions to ${runDir}/debate.md.${frictionClause(`judge-r${round}`)} Return the judge envelope.`,
      { ...judgment, label: `judge-r${round} · ${slug}`, phase: 'Debate', agentType: 'frank-exchange-of-views:lead-judge', schema: JUDGE_ENVELOPE })
    if (!judge) throw new Error(`judge round ${round} returned null (agent failed) — aborting cleanly`)
    for (const r of judge.resolutions) {
      if (r.resolution === 'closed' || r.resolution === 'rebuttal_sustained' || r.resolution === 'risk_accepted') adjudicated.push(r)
    }
    takeFriction(`judge-r${round}`, judge)
    if (judge.deadlock) { deadlocked = true; break }
  }
  for (const g of redEnv.gaps) allPriorGapIds.add(g.id)

  const adjudicatedIds = new Set(adjudicated.map(a => a.gap_id))
  const openGaps = redEnv.gaps.filter(g => !adjudicatedIds.has(g.id))
  blueEnv = await agent(
    `Blue response, round ${round}, topic "${topic}". Red's verdict: FAIL. Open gaps (adjudicated ones excluded): ${JSON.stringify(openGaps)}. Corroboration flags: ${JSON.stringify(redEnv.corroboration || [])}. BEFORE drafting, read the latest "### RED" section of ${runDir}/debate.md — the gap JSON above is a lossy summary of the transcript — and the latest "### LEAD" section if one exists: any gap the judge CARRIED comes with a stated research direction you owe. Address every open gap ADDITIVELY in ${runDir}/blue/report.md — expand and repair where red is right, rebut in writing (with evidence) where red is wrong, and argue risk-acceptance where the fix's complexity exceeds its likelihood x impact; never subtract substance, and propagate every correction to ALL sites that state the corrected claim, not only the flagged sentence (incomplete propagation was run 3's dominant blue failure class — 5 regressions in 5 rounds). Log edits to ${runDir}/blue/CHANGELOG.md (Round ${round}); append your "### BLUE" section for round ${round} to ${runDir}/debate.md.${frictionClause(`blue-respond-r${round}`)} Return the blue envelope.`,
    { ...bulk, label: `blue-respond-r${round} · ${slug}`, phase: 'Debate', agentType: 'frank-exchange-of-views:blue-researcher', schema: BLUE_ENVELOPE })
  takeFriction(`blue-respond-r${round}`, blueEnv)
  if (!blueEnv) throw new Error(`blue response round ${round} returned null (agent failed) — aborting cleanly`)
}

const exhausted = round >= maxRounds && redEnv && redEnv.verdict !== 'PASS' && !deadlocked
const verdict = redEnv && redEnv.verdict === 'PASS' ? 'VERIFIED' : 'UNVERIFIED'
log(`debate ended: ${verdict} after ${round} round(s)${deadlocked ? ' (judged deadlock)' : exhausted ? ' (safety ceiling hit)' : ''}`)

// ---- Assemble: union, not summary ----
phase('Assemble')
await agent(
  `Final assembly for topic "${topic}", run directory ${runDir}. Debate outcome: ${verdict} after ${round} round(s)${deadlocked ? ' by judged deadlock' : ''}${exhausted ? ' by safety ceiling' : ''}. Assemble ${runDir}/report.md by UNION per the report template (references/report_template.md): verdict stamp, TL;DR, the Catechism (references/catechism_template.md — the AGREED answers: the case against at full strength, of-interest-vs-merely-interesting, cost and stopping points), technical foundations, analysis, graded risk matrix (including risk_accepted items with rationale), then blue/report.md IN FULL, red/findings.md IN FULL, per-round debate synopsis pointing at debate.md, an "Open questions carried past this run" section from blue's final envelope: ${JSON.stringify((blueEnv && blueEnv.open_questions) || [])}, and the consolidated footnotes. Never compress the research into a digest. ${verdict === 'UNVERIFIED' ? 'Stamp UNVERIFIED and list every outstanding gap with its disposition and the compromise rationale.' : ''} Collated friction so far (report any of your own as well): ${JSON.stringify(friction)}.${frictionClause('assemble')} Return a 5-line synopsis of the final report, plus your own friction if any.`,
  { ...judgment, label: `assemble · ${slug}`, agentType: 'frank-exchange-of-views:lead-judge' })

return {
  runDir,
  verdict,
  rounds: round,
  lanes,
  deadlocked,
  gaps_outstanding: redEnv && redEnv.verdict !== 'PASS' ? redEnv.gaps.length : 0,
  blue_claims: blueEnv ? blueEnv.claim_count : null,
  friction,
}

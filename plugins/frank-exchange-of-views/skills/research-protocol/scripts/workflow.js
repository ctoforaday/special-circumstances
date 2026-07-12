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

// The lead is a script: mechanics, round-keeping, termination. All file writes
// belong to the agents (the filesystem is the blackboard; the script has no
// filesystem access by design). Judgment calls go to lead-judge, never round-to-round.
const { topic, runDir, lanes = 3, maxRounds = 12 } = args

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
          severity: { type: 'string', enum: ['low', 'medium', 'high'] },
          likelihood: { type: 'string', enum: ['low', 'medium', 'high'] },
          impact: { type: 'string', enum: ['low', 'medium', 'high'] },
          complexity_cost: { type: 'string', enum: ['low', 'medium', 'high'] },
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

// ---- Frontier ----
phase('Frontier')
await agent(
  `Research debate opening for topic: "${topic}". Formulate 3-5 frontier hypotheses (what would be true if each candidate answer were right) per the research protocol, and write them to ${runDir}/blue/frontier.md. Return one line per hypothesis.`,
  { label: 'frontier', agentType: 'frank-exchange-of-views:blue-researcher' })

// ---- Blue: best-of-N lanes, then additive synthesis ----
phase('Blue')
await parallel(Array.from({ length: lanes }, (_, i) => () => agent(
  `Blue lane ${i + 1} of ${lanes} for topic: "${topic}". Read ${runDir}/blue/frontier.md; research your assigned slice to saturation per the research protocol (spend at least one search in five on disconfirming evidence; semantic footnotes). Divide the frontier: lane ${i + 1} takes hypothesis ${i + 1} first, then breadth. Write your full candidate draft to ${runDir}/blue/candidates/lane-${i + 1}.md. Return a 3-line synopsis.`,
  { label: `blue-lane-${i + 1}`, phase: 'Blue', agentType: 'frank-exchange-of-views:blue-researcher' })))

let blueEnv = await agent(
  `Blue synthesis for topic: "${topic}". Read every draft in ${runDir}/blue/candidates/ and synthesize ${runDir}/blue/report.md by UNION: deduplicate overlapping claims, reorganize, and never drop substantive content — structural merge (append + dedup), not a free-form rewrite. Follow the report conventions (semantic footnotes). Start ${runDir}/blue/CHANGELOG.md with a Round 0 entry describing the synthesis. Return the blue envelope.`,
  { label: 'blue-synthesize', phase: 'Blue', agentType: 'frank-exchange-of-views:blue-researcher', schema: BLUE_ENVELOPE })

// ---- Debate loop: red audits gate; termination is judged, never counted ----
// Citation verification scales with report size: one pass per ~40 claims (capped).
const citationPasses = Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count || 20) / 40)))
let round = 0
let redEnv = null
let prevGapIds = new Set()
let deadlocked = false
const adjudicated = [] // judge-ruled gaps (closed / rebuttal_sustained / risk_accepted) — out of red's verdict
const friction = [] // capability complaints from any agent, aggregated for /self-improve
const takeFriction = (who, env) => { if (env && env.friction) for (const f of env.friction) friction.push(`${who}: ${f}`) }

while (round < maxRounds) {
  round++

  const lensPasses = []
  for (let c = 0; c < citationPasses; c++) {
    lensPasses.push(`${RED_LENSES[0]}${citationPasses > 1 ? ` — instance ${c + 1} of ${citationPasses}: divide the report's sections evenly among instances and take slice ${c + 1}` : ''}`)
  }
  lensPasses.push(RED_LENSES[1], RED_LENSES[2])

  await parallel(lensPasses.map((lens, i) => () => agent(
    `Red audit, round ${round}, lens: ${lens}. Re-read the FULL living report ${runDir}/blue/report.md in context (the whole document — never just a diff)${round > 1 ? `; blue's change log ${runDir}/blue/CHANGELOG.md is a navigation hint only` : ''}. Anchor every finding to a section heading plus a quoted sentence. Write your pass to ${runDir}/red/candidates/round-${round}-lens-${i + 1}.md. Return a 3-line synopsis.`,
    { label: `red-lens-${i + 1}-r${round}`, phase: 'Red', agentType: 'frank-exchange-of-views:red-auditor' })))

  redEnv = await agent(
    `Red merge, round ${round}. Read the round-${round} lens passes in ${runDir}/red/candidates/ and consolidate into the LIVING ${runDir}/red/findings.md (cumulative across rounds: update prior gaps' status, add new ones, keep graded corroboration and likelihood/impact/complexity on every gap; every gap's location = section heading + quoted sentence). Gap ids are stable across rounds (R1-1 stays R1-1).${adjudicated.length ? ` Gaps already adjudicated by the lead-judge and EXCLUDED from your verdict: ${JSON.stringify(adjudicated.map(a => a.gap_id))}.` : ''} Decide the binary verdict — PASS only when every remaining unadjudicated gap is closed, evidence-rebutted, or risk-accepted. Append the round-${round} "### RED" section to ${runDir}/debate.md per the debate template. Return the red envelope.`,
    { label: `red-merge-r${round}`, phase: 'Red', agentType: 'frank-exchange-of-views:red-auditor', schema: RED_ENVELOPE })

  takeFriction(`red-merge-r${round}`, redEnv)
  if (redEnv.verdict === 'PASS') break

  // Contested docket: gaps raised, rebutted, and re-raised go to the judge EARLY,
  // so debates converge instead of grinding. Detection is a set comparison in the
  // script; the judgment belongs to the lead-judge.
  const gapIds = new Set(redEnv.gaps.map(g => g.id))
  const contested = redEnv.gaps.filter(g => prevGapIds.has(g.id))
  const hasNew = [...gapIds].some(id => !prevGapIds.has(id))
  if (contested.length > 0) {
    const judge = await agent(
      `Adjudication, round ${round}, topic "${topic}". Contested docket (raised, rebutted by blue, re-raised by red): ${JSON.stringify(contested)}. New gaps were ${hasNew ? 'ALSO raised' : 'NOT raised'} this round. Read ${runDir}/debate.md and ${runDir}/red/findings.md in full. Rule per contested gap (closed | rebuttal_sustained | risk_accepted | carried | unresolved) with rationale — for carried, state what further research blue owes. deadlock is true only if no gap is carried AND no new gaps were raised. Append your "### LEAD" resolutions to ${runDir}/debate.md. Return the judge envelope.`,
      { label: `judge-r${round}`, phase: 'Debate', agentType: 'frank-exchange-of-views:lead-judge', schema: JUDGE_ENVELOPE })
    for (const r of judge.resolutions) {
      if (r.resolution === 'closed' || r.resolution === 'rebuttal_sustained' || r.resolution === 'risk_accepted') adjudicated.push(r)
    }
    takeFriction(`judge-r${round}`, judge)
    if (judge.deadlock) { deadlocked = true; break }
  }
  prevGapIds = gapIds

  const adjudicatedIds = new Set(adjudicated.map(a => a.gap_id))
  const openGaps = redEnv.gaps.filter(g => !adjudicatedIds.has(g.id))
  blueEnv = await agent(
    `Blue response, round ${round}, topic "${topic}". Red's verdict: FAIL. Open gaps (adjudicated ones excluded): ${JSON.stringify(openGaps)}. Corroboration flags: ${JSON.stringify(redEnv.corroboration || [])}. Address every open gap ADDITIVELY in ${runDir}/blue/report.md — expand and repair where red is right, rebut in writing (with evidence) where red is wrong, and argue risk-acceptance where the fix's complexity exceeds its likelihood x impact; never subtract substance. Log edits to ${runDir}/blue/CHANGELOG.md (Round ${round}); append your "### BLUE" section for round ${round} to ${runDir}/debate.md. Return the blue envelope.`,
    { label: `blue-respond-r${round}`, phase: 'Debate', agentType: 'frank-exchange-of-views:blue-researcher', schema: BLUE_ENVELOPE })
  takeFriction(`blue-respond-r${round}`, blueEnv)
}

const exhausted = round >= maxRounds && redEnv && redEnv.verdict !== 'PASS' && !deadlocked
const verdict = redEnv && redEnv.verdict === 'PASS' ? 'VERIFIED' : 'UNVERIFIED'
log(`debate ended: ${verdict} after ${round} round(s)${deadlocked ? ' (judged deadlock)' : exhausted ? ' (safety ceiling hit)' : ''}`)

// ---- Assemble: union, not summary ----
phase('Assemble')
await agent(
  `Final assembly for topic "${topic}", run directory ${runDir}. Debate outcome: ${verdict} after ${round} round(s)${deadlocked ? ' by judged deadlock' : ''}${exhausted ? ' by safety ceiling' : ''}. Assemble ${runDir}/report.md by UNION per the report template (references/report_template.md): verdict stamp, TL;DR, Heilmeier Catechism (references/heilmeier_template.md — the AGREED answers), technical foundations, analysis, graded risk matrix (including risk_accepted items with rationale), then blue/report.md IN FULL, red/findings.md IN FULL, per-round debate synopsis pointing at debate.md, and the consolidated footnotes. Never compress the research into a digest. ${verdict === 'UNVERIFIED' ? 'Stamp UNVERIFIED and list every outstanding gap with its disposition and the compromise rationale.' : ''} Return a 5-line synopsis of the final report.`,
  { label: 'assemble', agentType: 'frank-exchange-of-views:lead-judge' })

return {
  runDir,
  verdict,
  rounds: round,
  deadlocked,
  gaps_outstanding: redEnv && redEnv.verdict !== 'PASS' ? redEnv.gaps.length : 0,
  blue_claims: blueEnv ? blueEnv.claim_count : null,
  friction,
}

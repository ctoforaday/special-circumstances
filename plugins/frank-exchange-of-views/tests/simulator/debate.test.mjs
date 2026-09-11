import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { loadDebateScript, makeWorld, makeResponder, blueEnv, chairEnv, passChair, plan, passPlan, ceilingPlan, party, judgeEnv, petitionRulingEnv } from './harness.mjs'

// THE SIMULATOR DRIVES debate.js WITH STUBBED SEATS. There are no rounds (plans/roundless.md
// §III.B.1): the chair sits, its envelope relays the plan `dispatch next` recorded, the loop
// dispatches the parties the plan names, and the debate ends when the plan is empty. So a
// stubbed chair IS a stubbed plan, and a scenario is a sequence of plans.
const script = loadDebateScript(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url))
const ARGS = { topic: 'test topic', runDir: 'research/2026-01-01_test', lanes: 3, model: 'sonnet', judgmentModel: 'sonnet', binDir: '/opt/feov/bin' }
const isJudgmentSeat = (l) => /^(blue-synthesize|red-chair|judge|assemble)/.test(l)
const labelsOf = (world, prefix) => world.calls.filter((c) => c.opts.label.startsWith(prefix))
const firstPrompt = (world, prefix) => labelsOf(world, prefix)[0].prompt
const CITATION_CLAUSE = 'THE REPORT DOES NOT NAME ITS SOURCES'

// ── founding regressions ────────────────────────────────────────────────────────────────────

test('founding regression 1: JSON-stringified args parse; no undefined leaks into prompts', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  const out = await world.run(script, JSON.stringify(ARGS))
  assert.equal(out.verdict, 'VERIFIED')
  for (const c of world.calls) assert.ok(!c.prompt.includes('undefined'), `undefined in: ${c.opts.label}`)
})

test('founding regression 1b: unbound topic/runDir refuses dispatch before any agent spawns', async () => {
  const world = makeWorld(makeResponder())
  await assert.rejects(world.run(script, '{}'), /refusing dispatch/)
  assert.equal(world.calls.length, 0)
})

test('maxRounds is refused as an argument, with the terms that replaced it', async () => {
  const world = makeWorld(makeResponder())
  await assert.rejects(world.run(script, { ...ARGS, maxRounds: 12 }), (e) => /not a term of this engine/.test(e.message) && /k-max/.test(e.message) && /mint-budget/.test(e.message) && !/--k-max/.test(e.message))
  assert.equal(world.calls.length, 0, 'refused before anything was dispatched')
})

test('null chair, blue synthesis, blue response and bench abort cleanly, each naming its seat', async () => {
  const nullFor = (prefix, chair) => makeWorld((p, o) => (o.label.startsWith(prefix) ? null : makeResponder({ chair })(p, o)))
  await assert.rejects(nullFor('red-chair').run(script, ARGS), /red-chair sitting 1 returned null/)
  await assert.rejects(nullFor('blue-synthesize').run(script, ARGS), /blue synthesis returned null/)
  await assert.rejects(nullFor('blue-respond').run(script, ARGS), /blue response \(epoch 1\) returned null/)
  const benchPlan = chairEnv({ plan: plan([party('judge', 'G1')], { docket: ['G1'] }) })
  await assert.rejects(nullFor('judge', [benchPlan, passChair()]).run(script, ARGS), /bench sitting \(epoch 1\) returned null/)
})

test('a chair envelope without a plan aborts: the plan is the verb\'s JSON, relayed verbatim', async () => {
  const world = makeWorld(makeResponder({ chair: [{ unruled_motions: 0, log: [] }] }))
  await assert.rejects(world.run(script, ARGS), /relayed no plan/)
})

// ── the dispatch loop ───────────────────────────────────────────────────────────────────────

test('the plan is what dispatches: the parties it names sit, in role order — lenses, then blue, then the bench', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G1'), party('judge', 'G2'), party('red-lens-logic', 'G1'), party('red-lens-evidence')], { docket: ['G2'] }) }), passChair()],
  }))
  const out = await world.run(script, ARGS)
  const debate = world.calls.filter((c) => /^(red-lens|blue-respond|judge |judge#|judge ·|judge \#)/.test(c.opts.label) || c.opts.label.startsWith('judge #'))
  const order = debate.map((c) => c.opts.label.split(' ')[0])
  assert.deepEqual(order, ['red-lens-logic', 'red-lens-evidence', 'blue-respond', 'judge'], 'red parties first, then blue, then the bench — an exchange is red followed by blue')
  assert.equal(out.verdict, 'VERIFIED')
  assert.equal(out.epochs, 2, 'two chair sittings')
})

test('a lens engaged with no gaps is told the head moved; one engaged on gaps is told which, and that silence is a null turn', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('red-lens-logic', 'G3', 'G4')], { head: 7 }) }), passChair()],
  }))
  await world.run(script, ARGS)
  const evidence = firstPrompt(world, 'red-lens-evidence')
  const logic = firstPrompt(world, 'red-lens-logic')
  assert.ok(/HEAD MOVED PAST YOUR LAST SITTING \(head 7\)/.test(evidence), 'rule 1: the head is past the pin')
  assert.ok(/YOU ARE ENGAGED ON: G3, G4/.test(logic), 'the engaged lens is told its gaps')
  assert.ok(/NULL TURN/.test(logic) && /silence is a turn taken/.test(logic), 'a null turn counts toward impasse')
  assert.ok(/DID BLUE ACTUALLY DO WHAT YOU ASKED/.test(logic), 'the lens compares its fix to blue\'s edits')
})

test('a party the workflow has no sitting for aborts — the cast and the workflow must agree', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('frontier')]) })] }))
  await assert.rejects(world.run(script, ARGS), /names frontier, which this workflow has no sitting for/)
})

test('labels carry the seat and its sitting ordinal, never a round or a lens number', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv(), chairEnv({ plan: plan([party('red-lens-evidence', 'G1')]) }), passChair()],
  }))
  await world.run(script, ARGS)
  const labels = world.calls.map((c) => c.opts.label)
  assert.ok(labels.includes('red-chair #1 · test'), labels.join('\n'))
  assert.ok(labels.includes('red-chair #3 · test'))
  assert.ok(labels.includes('red-lens-evidence #2 · test'), 'the lens sat twice')
  for (const l of labels) {
    assert.ok(!/-r\d+/.test(l), `a label carries a round: ${l}`)
    assert.ok(!/^red-lens-\d/.test(l), `a lens is identified by number: ${l}`)
  }
})

test('heartbeats: the narrator logs each epoch\'s head, parties, docket and the chair\'s verdict', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('red-lens-evidence', 'G1'), party('blue-respond', 'G1'), party('judge', 'G2')], { head: 5, docket: ['G2'] }) }), passChair()],
  }))
  await world.run(script, ARGS)
  const all = world.logs.join('\n')
  assert.ok(/epoch 1: head 5 — 3 party\(ies\) ready, docketed G2/.test(all), all)
  assert.ok(/epoch 1: dispatching 1 red lens\(es\): evidence\[G1\]/.test(all))
  assert.ok(/epoch 1: blue responded on G1/.test(all))
  assert.ok(/epoch 1: the bench sits on G2/.test(all))
  assert.ok(/epoch 2: head 2 — 0 party\(ies\) ready, chair recorded PASS — PASS permitted/.test(all))
})

// ── termination: three ways, from the plan ──────────────────────────────────────────────────

test('VERIFIED: the chair records PASS on a plan that permits it; phases in order; lanes honoured', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  const out = await world.run(script, ARGS)
  assert.deepEqual({ verdict: out.verdict, epochs: out.epochs, gaps: out.gaps_outstanding, term: out.termination },
    { verdict: 'VERIFIED', epochs: 1, gaps: 0, term: { pass_permitted: true, ceiling: false, why: [] } })
  assert.deepEqual(world.phases, ['Frontier', 'Blue', 'Red', 'Assemble'])
  assert.equal(labelsOf(world, 'blue-lane').length, 3)
  assert.ok(world.logs.some((m) => m.includes('researching: test topic')))
  assert.ok(!world.calls.some((c) => c.opts.label.startsWith('judge-terminal')), 'no unruled motion, no terminal sitting')
})

test('CEILING: nobody ready and every open material gap at its limit — the stamp says so and names no deadlock', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: ceilingPlan() })] }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'CEILING')
  assert.equal(out.termination.ceiling, true)
  const asm = firstPrompt(world, 'assemble')
  assert.ok(/it is CEILING/.test(asm) && /NOT a judged failure to verify/.test(asm), asm.slice(0, 600))
  assert.ok(/Record it as the run's outcome — CEILING, ended at the ceiling/.test(asm), 'the bench is told the exact stamp, as an act and not a command line')
  assert.ok(!/deadlock/i.test(asm) || /not a judged/i.test(asm), 'deadlock is per gap and on the record, not a run-level stamp')
})

test('UNVERIFIED: nobody ready, neither PASS nor CEILING — the plan\'s reasons are the account', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([], { why: ['G1: docketed, the bench sat and ruled nothing'] }) })] }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'UNVERIFIED')
  assert.deepEqual(out.termination, { pass_permitted: false, ceiling: false, why: ['G1: docketed, the bench sat and ruled nothing'] })
  const asm = firstPrompt(world, 'assemble')
  assert.ok(/ended UNVERIFIED/.test(asm) && /the bench sat and ruled nothing/.test(asm), 'the reason reaches the stamp')
})

test('a chair that records PASS ends the debate even if it relayed parties; a FAIL with parties continues', async () => {
  const passing = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('red-lens-evidence')]), verdict: 'PASS' })] }))
  const out = await passing.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  assert.equal(labelsOf(passing, 'red-lens').length, 0, 'PASS ends the sitting sequence')
  const failing = makeWorld(makeResponder({ chair: [chairEnv({ verdict: 'FAIL' }), passChair()] }))
  const out2 = await failing.run(script, ARGS)
  assert.equal(out2.verdict, 'VERIFIED')
  assert.equal(out2.epochs, 2)
})

test('the terminal bench sitting fires only when the chair reports unruled motions, and before assembly', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: ceilingPlan(), unruled_motions: 2 })] }))
  await world.run(script, ARGS)
  const terminal = labelsOf(world, 'judge-terminal')[0]
  const asm = labelsOf(world, 'assemble')[0]
  assert.ok(terminal, 'the terminal sitting fired')
  assert.ok(terminal.n < asm.n, 'disposition precedes assembly')
  assert.ok(/2 motion\(s\) stand unruled/.test(terminal.prompt) && /NOTHING CAN BE CARRIED AT A TERMINAL EXIT/.test(terminal.prompt))
})

// ── the seats' contracts ────────────────────────────────────────────────────────────────────

test('the chair runs the debate: relays the plan verbatim, mints nothing, closes nothing, owns the verdict', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  await world.run(script, ARGS)
  const chair = labelsOf(world, 'red-chair')[0]
  const p = chair.prompt
  assert.ok(/you mint nothing and you close nothing/.test(p) && /ASK THE RECORD WHO SITS — the dispatch/.test(p) && /Relay its JSON VERBATIM as `plan`/.test(p))
  assert.ok(/THE STOPPING JUDGMENT IS YOURS/.test(p) && /record a PASS verdict/.test(p) && /FAIL over a converged board is refused/.test(p))
  assert.ok(/CLOSING ARGUMENTS/.test(p) && /~120 words/.test(p) && /the spot-check; its assertable empty form/.test(p) && /VOTE EVERY LINE OF INQUIRY/.test(p) && /RULE ON BLUE'S DIRECTIONS/.test(p) && /RULE THE GRADE MOTIONS/.test(p))
  assert.ok(!/dispatch next|--as |show (board|work|motions)|motion (grade|inquiry) rule/.test(p), 'the chair is told the ACT; the word it types is the help page\'s')
  assert.ok(!/COALESCE/.test(p) && !/only writer/.test(p), 'the coalescing chair is gone with the mechanism')
  assert.ok(chair.opts.schema.properties.plan && chair.opts.schema.required.includes('plan'), 'the envelope carries the plan')
  assert.ok(!chair.opts.schema.properties.gaps, 'and no gap list — the record holds the board')
  assert.ok(/YOUR IN-RUN SCORECARD/.test(p) && !/YOUR CHAIR'S SCORECARD/.test(p), 'the in-run self-read, never a cross-run seed')
})

test('the lens mints its own gaps, screens first, spends a budget, and closes as the originator', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('red-lens-logic'), party('red-lens-dark-side')]) }), passChair()] }))
  await world.run(script, ARGS)
  const evidence = firstPrompt(world, 'red-lens-evidence')
  for (const want of ['YOU MINT YOUR OWN GAPS', 'for a near match', 'then mint', 'registering a new class first', 'mintBudget', 'names the gaps it supersedes', "says so in the check's kind", 'THE ORIGINATOR CLOSES', 'LINEAGE IS NEVER DROPPED',
    'ASK FOR THE ANSWER TO BE PRODUCED, NOT ASSERTED', 'DOCUMENT-PROBE', 'LIVE-PROBE', 'BELIEVE NO BYTES', 'PRESCRIBE TEXT ONLY WHERE THE DEFECT IS TEXTUAL',
    'ANCHOR EVERY FINDING TO A QUOTED SENTENCE', "labels on your findings are the tool's to assign", 'read it whole in consecutive windows',
    'counts LINES, not occurrences', 'prefer the Write tool over quoted heredocs']) {
    assert.ok(evidence.includes(want), `the lens prompt lost: ${want}`)
  }
  assert.ok(!/gap ids are the chair's/.test(evidence), 'the ids are the tool\'s, minted by the lens')
  assert.ok(evidence.includes(CITATION_CLAUSE) && /VERBATIM READS ONLY/.test(evidence) && /--comments/.test(evidence), 'the evidence lens carries the ledger clause')
  assert.ok(/STEELMAN DUTY/.test(firstPrompt(world, 'red-lens-logic')) && /STEELMAN DUTY/.test(firstPrompt(world, 'red-lens-dark-side')), 'logic and dark-side audit the declines')
  assert.ok(!/STEELMAN DUTY/.test(evidence) && !/take slice|instance \d+ of/.test(evidence), 'the evidence seat verifies sources and owns no slice')
})

test('blue is engaged on named gaps, told the board is authoritative, and files closings only when the plan docketed', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G1', 'G2')], { docket: ['G2'] }) }), chairEnv({ plan: plan([party('blue-respond', 'G3')]) }), passChair()],
  }))
  await world.run(script, ARGS)
  const [first, second] = labelsOf(world, 'blue-respond').map((c) => c.prompt)
  assert.ok(/You are engaged on: G1, G2/.test(first))
  for (const want of ['YOUR FIRST READ COMES AFTER THE MANUAL', 'red-gap-patterns.md', 'in one pass rather than three', 'lossy summary', "bench's latest resolutions",
    'CARRIED comes with a stated research direction you owe', 'which patterns you checked', 'YOU MAY COMPUTE AN ANSWER', 'DOCUMENT-PROBE', 'deferred acceptance test',
    'LINES OF INQUIRY ARE A LIVING RECORD', 'THREE paths', 'ESTOPS', 'OWNERSHIP BINDS, AS IT DID AT SYNTHESIS', 'each edit naming the gap it answers', 'a grade motion on the axis', 'Compact and reorganize prose', 'retired on the record',
    'PROPAGATE EVERY CORRECTION TO ALL SITES', 'NULL TURN', 'AUDIT YOUR OWN REPAIRS, ONE RECEIPT PER GAP', 'manifest array', 'claim_count', 'never hand-count']) {
    assert.ok(first.includes(want), `blue lost: ${want}`)
  }
  assert.ok(/CLOSING ARGUMENTS: the following are DOCKETED for adjudication AFTER your response this sitting: G2/.test(first) && /argue in ~120 words/.test(first))
  assert.ok(!/CLOSING ARGUMENTS/.test(second), 'no docket this sitting, no closing demanded')
  assert.ok(!/round \d/.test(first), 'no round is named to blue')
})

test('the bench rules on docketed gaps from the closings, the transcript and the live artifact; carried is the gap\'s deadlock', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('judge', 'G1', 'G2')], { docket: ['G1'] }) }), passChair()] }))
  await world.run(script, ARGS)
  const bench = firstPrompt(world, 'judge')
  for (const want of ['Docketed for you: G1, G2', 'THE DOCKET IS A ROUTING LIST, NOT THE EVIDENCE', 'read them FRESH before ruling', 'AS IT NOW STANDS', 'Re-run each document-probe acceptance check',
    'RULING BASIS IS CONFINED TO', 'counts AGAINST the side that made it', "READ THE NAMED ANCESTORS' RECORDS", 'the docket ruling', 'barring as settled', 'CARRY', 'NAMED infrastructure debt',
    'rules nothing is a workflow error', 'PRECEDENT IS ARGUMENT, NOT EVIDENCE']) {
    assert.ok(bench.includes(want), `the bench lost: ${want}`)
  }
  const schema = labelsOf(world, 'judge')[0].opts.schema
  assert.ok(!schema.properties.deadlock && !schema.required.includes('deadlock'), 'deadlock is not a flag the bench returns')
  assert.ok(schema.properties.resolutions.items.properties.resolution.enum.includes('defect_owed_elsewhere'))
})

test('W1.9: defect_owed_elsewhere ships as a named infra debt with the epoch it was ruled in', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('judge', 'G1')], { docket: ['G1'] }) }), passChair()],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'G1', resolution: 'defect_owed_elsewhere', rationale: 'setup tooling must stage it' }] })],
  }))
  const out = await world.run(script, ARGS)
  assert.deepEqual(out.infra_debts, [{ gap_id: 'G1', owed_fix: 'setup tooling must stage it', epoch: 1 }])
  assert.ok(firstPrompt(world, 'assemble').includes('setup tooling must stage it'), 'assembly is handed the named debts')
})

test('no seat prompt names a command path or spells a flag — the help page is the only page', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G1')], { docket: ['G1'] }) }), chairEnv({ plan: plan([party('judge', 'G1')], { docket: ['G1'] }) }), passChair()],
  }))
  await world.run(script, ARGS)
  for (const c of world.calls) {
    const body = c.prompt.replace(/--help\b|--seat-id\b|--run\b|--reason(-file)?\b|--comments\b/g, '')
    assert.ok(!/--[a-z][a-z-]+\b/.test(body), `${c.opts.label} spells a flag: ${(body.match(/.{40}--[a-z][a-z-]+.{20}/) || [''])[0]}`)
    assert.ok(!/\b(show (board|work|motions|closings|changes)|class new|dispatch next|motion (grade|inquiry|docket|petition) (file|rule|appeal)|near-match)\b/.test(body),
      `${c.opts.label} names a command path: ${(body.match(/.{40}(show \w+|class new|dispatch next|motion \w+ \w+|near-match).{20}/) || [''])[0]}`)
  }
})

test('the bench\'s rulings travel to both parties, with blue told its duty and red estopped, and the reasoning stays on the record', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G1')]) }), chairEnv({ plan: plan([party('judge', 'G1')], { docket: ['G1'] }) }),
      chairEnv({ plan: plan([party('red-lens-evidence', 'G1'), party('blue-respond', 'G1')]) }), passChair()],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'G1', resolution: 'not_a_defect', rationale: 'THE OPINION', settled: 'THE BARRED PROPOSITION', reopens_on: 'A NEW SOURCE', final: false }] })],
  }))
  await world.run(script, ARGS)
  const [blue1, blue2] = labelsOf(world, 'blue-respond').map((c) => c.prompt)
  assert.ok(!/GAPS THE BENCH HAS RULED/.test(blue1), 'nothing ruled yet, nothing claimed')
  for (const want of ['GAPS THE BENCH HAS RULED', 'THE BENCH FOUND NO DEFECT', 'rely on this ruling as established', 'THE BARRED PROPOSITION', 'A NEW SOURCE', 'THE BAR IS ON THE PROPOSITION, NOT THE GAP']) {
    assert.ok(blue2.includes(want), `blue lost: ${want}`)
  }
  const lens = firstPrompt(world, 'red-lens-evidence')
  assert.ok(/YOU ARE ESTOPPED/.test(lens) && lens.includes('THE BARRED PROPOSITION') && lens.includes('A NEW SOURCE'), 'red is handed the bar and what reopens it')
  assert.ok(!blue2.includes('THE OPINION') && !lens.includes('THE OPINION'), 'the reasoning is on the record, not in the prompt')
  assert.ok(!/your_duty/.test(lens), 'the duty table is blue\'s')
})

test('the assembler authors nothing, stamps the outcome the record derives, and lists trifles as not certified against', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  await world.run(script, ARGS)
  const asm = firstPrompt(world, 'assemble')
  assert.ok(/you author NOTHING, you copy NOTHING/.test(asm) && /do not copy anything into it yourself/.test(asm) && /cannot mis-author a synthesis surface/.test(asm))
  assert.ok(/it is VERIFIED/.test(asm) && /Record it as the run's outcome — VERIFIED/.test(asm) && !/ended at the ceiling/.test(asm))
  // THE REASON IS A JUDGEMENT, NOT A RECAP. `--reason` carries why this outcome is the right read
  // of the board; the ledger already holds what the sitting did, in order.
  assert.ok(/WHY this outcome is the right read of the board/.test(asm) && !/your account of the sitting/.test(asm),
    'the assemble prompt asks for a recap of the sitting rather than the judgement')
  assert.ok(/open, below material, not certified against/.test(asm))
  assert.ok(!asm.includes('open_questions'), 'assembly lifts blue\'s audited section, never receives it')
})

// ── bookends and prompt clauses carried forward ─────────────────────────────────────────────

test('an unknown lens area is refused before any seat is dispatched', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  await assert.rejects(() => world.run(script, { ...ARGS, lensAreas: ['evidence', 'adversarial'] }),
    (e) => /unknown lens area/.test(e.message) && /adversarial/.test(e.message) && /architecture/.test(e.message))
  assert.equal(world.calls.length, 0)
})

test('lane floor: lanes below 3 refuses dispatch unless an override reason is given', async () => {
  await assert.rejects(makeWorld(makeResponder()).run(script, { ...ARGS, lanes: 1 }), /below the floor of 3/)
  const world2 = makeWorld(makeResponder({ chair: [passChair()] }))
  const out = await world2.run(script, { ...ARGS, lanes: 1, laneFloorOverride: 'smoke run' })
  assert.equal(out.lanes, 1)
})

test('lane methods and the redundancy-floor seat at lanes=5', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  await world.run(script, ARGS)
  const lanes = labelsOf(world, 'blue-lane').map((c) => c.prompt)
  assert.equal(lanes.length, 3)
  assert.ok(lanes[0].includes('adversarial-disconfirming-first') && lanes[1].includes('primary-literature') && lanes[2].includes('local-repo critical-stance'))
  for (const [i, p] of lanes.entries()) assert.ok(p.includes('SOURCE NOTES') && p.includes('Do NOT mint footnote labels'), `lane ${i + 1} source-note convention`)
  const five = makeWorld(makeResponder({ chair: [passChair()] }))
  await five.run(script, { ...ARGS, lanes: 5 })
  const prompts = labelsOf(five, 'blue-lane').map((c) => c.prompt)
  assert.equal(prompts.length, 5)
  assert.ok(prompts[3].includes('practitioner-production') && prompts[4].includes('adversarial-disconfirming-first, second seat'))
})

test('synthesis: provenance tagging, open questions, the catechism, and ownership of only what blue can author', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()], blueSynth: [blueEnv({ open_questions: ['does the schema guarantee conformance-or-null?'] })] }))
  await world.run(script, ARGS)
  const synth = firstPrompt(world, 'blue-synthesize')
  // PROVENANCE IS NOT WRITTEN INTO PROSE. The synthesis prompt used to order an inline
  // `[minority: lane-N]` tag on every single-lane claim — which blue's own constitution names
  // as the measured failure it exists to prevent, and which reportvoice flags. Six of them
  // shipped in a VERIFIED report before this was found. The fact is real and the record
  // already holds it: the lane drafts, the frozen base and every recorded edit.
  assert.ok(!/minority|lane marker|exactly ONE lane/.test(synth), 'the synthesis prompt orders provenance into prose again')
  // THE CONTENT RULE IS THE VERBS', and the prompt points at it rather than restating it: the rule is
  // on every verb whose text the report prints (edit, ingest, and line-of-inquiry's propose and move),
  // so a copy here is a second statement of a contract nothing keeps in step.
  assert.ok(/THE REPORT'S CONTENT RULE .* is on the verbs that write report text/.test(synth),
    'the synthesis prompt must point at where the report content rule lives')
  assert.ok(!/RESEARCH PROSE AND THE TOOL'S OWN MARKERS/.test(synth) && !/EXTRACTED from the record/.test(synth),
    'the synthesis prompt restates the verbs\' report content rule instead of pointing at it')
  assert.ok(synth.includes('## Open questions') && synth.includes('THE CATECHISM IS YOURS') && synth.includes('## The Catechism') && synth.includes('every risk-accepted residual'))
  assert.ok(synth.includes('## The board') && synth.includes('is FABRICATION') && synth.includes('ONLY WHAT YOU CAN AUTHOR'))
  assert.ok(/reorganize freely/i.test(synth) && /retired on the record/.test(synth))
  assert.ok(/LINES OF INQUIRY/.test(synth) && /dead ends matter most/.test(synth) && /CONSIDERED, not only the one you took/.test(synth))
  assert.ok(/what you weighed and rejected are three things a reader needs/.test(synth))
  assert.ok(/YOUR IN-RUN SCORECARD/.test(synth) && /YOUR CHAIR/.test(synth) && /the seat you registered as/.test(synth) && !synth.includes('scorecards.mjs') && !/--bin\b/.test(synth))
})

test('priors-are-poison: no cross-run scorecard seed reaches any chair, even when scorecards are supplied', async () => {
  const scorecards = { 'blue-synthesize': { repair_regression_ratio: 0.63 }, 'red-chair': { anchored_closures_pct: 89 }, assemble: { carried_share: 0.98 } }
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  await world.run(script, { ...ARGS, scorecards })
  assert.ok(!firstPrompt(world, 'blue-synthesize').includes('0.63') && !firstPrompt(world, 'red-chair').includes('89') && !firstPrompt(world, 'assemble').includes('0.98'))
  assert.ok(!world.calls.some((c) => /YOUR CHAIR'S SCORECARD/.test(c.prompt)))
  for (const c of world.calls.filter((c) => /YOUR IN-RUN SCORECARD/.test(c.prompt))) {
    assert.ok(/projection of this run's record/.test(c.prompt))
    assert.ok(!/\d+\.\d\d/.test(c.prompt.match(/YOUR IN-RUN SCORECARD[^.]*\./)[0]), 'no prior number is seeded')
  }
})

test('every seat prompt carries the log clause, the speed clause and the record contract; bench sittings carry the law clause', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G1'), party('judge', 'G1')], { docket: ['G1'] }) }), passChair({ unruled_motions: 1 })],
  }))
  await world.run(script, ARGS)
  for (const seat of ['blue-synthesize', 'red-chair', 'red-lens-evidence', 'blue-respond', 'judge #', 'judge-terminal', 'assemble']) {
    const c = labelsOf(world, seat)[0]
    assert.ok(c, `${seat} sat`)
    assert.ok(c.prompt.includes("envelope's log field") && /AUDIENCE IS THE OPERATOR/.test(c.prompt) && /JUDGEMENT rather than for want of occasion/.test(c.prompt), `${seat} lost the operator channel`)
    assert.ok(!/OWES THE SURVEY/.test(c.prompt) && !/SILENCE IS NOT THE EMPTY CASE/.test(c.prompt), `${seat} restates a retired rule`)
    assert.ok(c.prompt.includes('SPEED:') && c.prompt.includes('batch INDEPENDENT tool calls'), `${seat} lost the speed clause`)
    assert.ok(c.prompt.includes('SEAT_ID:') && c.prompt.includes('/opt/feov/bin/feov-record'), `${seat} lost the record contract`)
  }
  for (const seat of ['judge #', 'judge-terminal', 'assemble']) {
    const p = labelsOf(world, seat)[0].prompt
    assert.ok(p.includes('PRECEDENT IS ARGUMENT, NOT EVIDENCE') && p.includes('the leaf wins') && p.includes('only AFFIRMED ones bind'), `${seat} lost the law clause`)
    assert.ok(!/DECLARE:/.test(p), `${seat} re-teaches declare beside its help page`)
  }
  const lens = firstPrompt(world, 'red-lens-evidence')
  // ONE CALL READS THE SURFACE: `manual` prints every command's own --help, and a single page is the re-check.
  assert.ok(/--seat-id red-lens-evidence manual — every command on your surface/.test(lens) && lens.includes('<command> --help') && /IMMEDIATELY AFTER `register`/.test(lens), 'the lens lost the manual directive')
  assert.ok(!/for EVERY group that page listed/.test(lens) && !lens.includes('<group> --help'), 'the lens still carries the page-by-page walk the manual replaced')
  assert.ok(/KNOWN HARNESS LIMIT/.test(lens) && /SANCTIONED fallback/.test(lens), 'the Glob/Grep fallback is sanctioned everywhere via the speed clause')
})

test('the record contract binds each seat to the id it hands the tool, petition sittings included', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G1')]), petitions: [{ class: 'ethical', ask: 'x', relief: 'narrow' }] }), passChair()],
    blueRespond: [blueEnv({ petitions: [{ class: 'procedural', ask: 'y', relief: 'z' }] })],
  }))
  await world.run(script, ARGS)
  const ids = labelsOf(world, 'judge-petition').map((c) => c.opts.label.split(' ')[0])
  assert.deepEqual(ids, ['judge-petition-red-chair', 'judge-petition-blue-respond'], 'each petition sitting is named for its filer')
  for (const c of world.calls) {
    const seat = c.opts.label.split(' ')[0]
    if (/^(frontier|blue-lane)/.test(seat)) continue
    assert.ok(c.prompt.includes(`SEAT_ID: ${seat}`), `${c.opts.label} declares a SEAT_ID other than its own`)
  }
})

test('W2c: a petition dispatches a bench sitting before the next seat; denied continues; a halt ends the run HALTED', async () => {
  const denied = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G1')]), petitions: [{ class: 'ethical', ask: 'x', relief: 'narrow' }] }), passChair()],
  }))
  const out = await denied.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  const sitting = denied.calls.findIndex((c) => c.opts.label.startsWith('judge-petition'))
  const blue = denied.calls.findIndex((c) => c.opts.label.startsWith('blue-respond'))
  assert.ok(sitting >= 0 && sitting < blue, 'the sitting fired before the parties sat')
  assert.equal(out.petitions.length, 1)
  assert.ok(denied.calls[sitting].prompt.includes('never sanctioned') && /A HALT IS A DIFFERENT DECISION FROM A RULING/.test(denied.calls[sitting].prompt))
  const halting = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G1')]), petitions: [{ class: 'ethical', ask: 'x', relief: 'stop' }] })],
    petition: [petitionRulingEnv({ halt: { opinion: 'the human must decide this' } })],
  }))
  const out2 = await halting.run(script, ARGS)
  assert.equal(out2.verdict, 'HALTED')
  assert.equal(out2.halted, true)
  assert.ok(out2.halt_opinion.includes('the human must decide'))
  assert.ok(!halting.calls.some((c) => c.opts.label.startsWith('blue-respond')), 'nothing sat past the halt')
  const asm = firstPrompt(halting, 'assemble')
  assert.ok(asm.includes('it is HALTED') && asm.includes('the human must decide'))
})

test('W2c: no petitions, no sitting; granted relief binds the party it names', async () => {
  const quiet = makeWorld(makeResponder({ chair: [passChair()] }))
  await quiet.run(script, ARGS)
  assert.ok(!quiet.calls.some((c) => c.opts.label.startsWith('judge-petition')))
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G1')]), petitions: [{ class: 'procedural', ask: 'x', relief: 'scope narrowed to §3' }] }), passChair()],
    petition: [petitionRulingEnv({ rulings: [{ petitioner: 'red-chair', class: 'procedural', ruling: 'granted', relief: 'scope narrowed to §3', binds: 'blue' }] })],
  }))
  await world.run(script, ARGS)
  assert.ok(firstPrompt(world, 'blue-respond').includes('BENCH RELIEF IN EFFECT') && firstPrompt(world, 'blue-respond').includes('scope narrowed'))
  assert.ok(firstPrompt(world, 'blue-synthesize').includes('PETITIONS:'), 'the right is stated at the seat')
})

test('W2j: a bench holding binds every seat that follows it, across sittings', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('judge', 'G1')], { docket: ['G1'] }) }), chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G2')]) }), passChair()],
    judge: [judgeEnv({ holdings: [{ term: 'material', construed: 'medium or above on current severity' }] })],
  }))
  await world.run(script, ARGS)
  for (const seat of ['red-chair #2', 'red-lens-evidence', 'blue-respond', 'assemble']) {
    assert.ok(labelsOf(world, seat)[0].prompt.includes('BENCH HOLDINGS IN EFFECT'), `${seat} was not bound by the holding`)
  }
  assert.ok(!labelsOf(world, 'red-chair #1')[0].prompt.includes('BENCH HOLDINGS IN EFFECT'), 'nothing binds before it is held')
})

test('integrity inspection arms only with transcriptDir, on the bench, and binds integrity-not-merits', async () => {
  const chair = [chairEnv({ plan: plan([party('judge', 'G1')], { docket: ['G1'] }) }), passChair()]
  const armed = makeWorld(makeResponder({ chair }))
  await armed.run(script, { ...ARGS, transcriptDir: '/sess/wf-123' })
  const bench = firstPrompt(armed, 'judge')
  assert.ok(/INTEGRITY INSPECTION/.test(bench) && /\/sess\/wf-123\/agent-\*\.jsonl/.test(bench) && /MUST NOT use trajectory material to decide the MERITS/.test(bench) && /DECLARE every inspection/.test(bench))
  const unarmed = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('judge', 'G1')], { docket: ['G1'] }) }), passChair()] }))
  await unarmed.run(script, ARGS)
  assert.ok(!unarmed.calls.some((c) => /INTEGRITY INSPECTION/.test(c.prompt)))
})

test('the sitting record (W1.7): blue is re-prompted once, continues with friction, and a recovered attestation logs none', async () => {
  const chair = [chairEnv({ plan: plan([party('blue-respond', 'G1')]) }), passChair()]
  const unresolved = makeWorld(makeResponder({ chair, blueRespond: [blueEnv({ sitting_record_appended: false })] }))
  const out = await unresolved.run(script, ARGS)
  assert.ok(unresolved.calls.some((c) => c.opts.label.startsWith('blue-respond-sitting-record')), 'the seat was re-prompted for its sitting record')
  assert.ok(out.friction.some((f) => /sitting-record.*UNRESOLVED/.test(f)))
  const synth = makeWorld(makeResponder({ chair: [passChair()], blueSynth: [blueEnv({ sitting_record_appended: false })] }))
  const out2 = await synth.run(script, ARGS)
  assert.ok(synth.calls.some((c) => c.opts.label.startsWith('blue-synthesize-sitting-record')) && labelsOf(synth, 'red-chair').length === 1 && out2.friction.some((f) => /sitting-record.*UNRESOLVED/.test(f)))
  const recovered = makeWorld((p, o) => {
    if (o.label.startsWith('blue-synthesize-sitting-record')) return { sitting_record_appended: true }
    return makeResponder({ chair: [passChair()], blueSynth: [blueEnv({ sitting_record_appended: false })] })(p, o)
  })
  const out3 = await recovered.run(script, ARGS)
  assert.ok(!out3.friction.some((f) => /sitting-record/.test(f)), 'a recovered attestation logs no friction')
})

// W2b, OWED GAPS ONLY, NEVER AN ABORT (gblock's ruling on #868). The B4 ordering: the plan engaged
// the lenses and blue on G1 and G2, the lenses sat first and closed both, and blue correctly found
// nothing to repair. The engine checked blue against the plan and threw, killing the run.
const engagedOn = (...gaps) => [chairEnv({ plan: plan([party('red-lens-logic', ...gaps), party('blue-respond', ...gaps)]) }), passChair()]
test('W2b: the B4 ordering — lenses close G1 and G2 first, blue files no row — continues and logs nothing unmanifested', async () => {
  const world = makeWorld(makeResponder({ chair: engagedOn('G1', 'G2'), blueRespond: [blueEnv({ manifest: [], found_closed: ['G1', 'G2'] })] }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  assert.ok(!world.logs.some((l) => /no row for/.test(l)), `a gap closed before blue sat is not owed: ${world.logs.join(' | ')}`)
  assert.ok(world.logs.some((l) => l.includes('blue found G1, G2 closed before it sat')))
})
test('W2b: a gap still open with no row is logged and left to capture; none of N is no longer fatal', async () => {
  const oneOpen = makeWorld(makeResponder({ chair: engagedOn('G1', 'G2'), blueRespond: [blueEnv({ manifest: [], found_closed: ['G1'] })] }))
  assert.equal((await oneOpen.run(script, ARGS)).verdict, 'VERIFIED')
  assert.ok(oneOpen.logs.some((l) => l.includes('0/1 gap(s) open when blue sat — no row for: G2') && l.includes('scored at capture')), oneOpen.logs.join(' | '))
  const noneOfTwo = makeWorld(makeResponder({ chair: engagedOn('G1', 'G2'), blueRespond: [blueEnv({ manifest: [] })] }))
  assert.equal((await noneOfTwo.run(script, ARGS)).verdict, 'VERIFIED')
  assert.ok(noneOfTwo.logs.some((l) => l.includes('0/2 gap(s) open when blue sat — no row for: G1, G2')))
})
test('W2b: partial coverage is logged, never fatal; a found_closed id blue was not engaged on is ignored', async () => {
  const partial = makeWorld(makeResponder({ chair: engagedOn('G1', 'G2'), blueRespond: [blueEnv({ manifest: ['G1'], found_closed: ['G9'] })] }))
  const out = await partial.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  assert.ok(partial.logs.some((l) => l.includes('1/2 gap(s) open when blue sat') && l.includes('no row for: G2')))
  assert.ok(!partial.logs.some((l) => l.includes('G9')), 'a claimed closure outside the engagement excuses nothing')
})

test('the operator channel aggregates from every seat with attribution, and assembly receives it', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G1')]), log: ['no PDF extraction'] }), passChair()],
    blueRespond: [blueEnv({ log: ['rate-limited on WebFetch'] })],
    blueSynth: [blueEnv({ log: ['write-block on blue/report.md'] })],
  }))
  const out = await world.run(script, ARGS)
  assert.ok(out.friction.includes('red-chair: no PDF extraction') && out.friction.includes('blue-respond: rate-limited on WebFetch') && out.friction.includes('blue-synthesize: write-block on blue/report.md'))
  assert.ok(firstPrompt(world, 'assemble').includes('no PDF extraction'))
})

test('per-role models: bulk seats get `model`, judgment seats get `judgmentModel`; unset either throws; binDir is required', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G1'), party('judge', 'G1')]) }), passChair()] }))
  await world.run(script, { ...ARGS, model: 'haiku', judgmentModel: 'opus' })
  for (const c of world.calls) assert.equal(c.opts.model, isJudgmentSeat(c.opts.label) ? 'opus' : 'haiku', `${c.opts.label} takes its class tier`)
  const base = { topic: 't', runDir: 'research/x', binDir: '/b' }
  await assert.rejects(makeWorld(makeResponder()).run(script, { ...base, judgmentModel: 'sonnet' }), /refusing dispatch — model unset/)
  await assert.rejects(makeWorld(makeResponder()).run(script, { ...base, model: 'sonnet' }), /refusing dispatch — judgmentModel unset/)
  await assert.rejects(makeWorld(makeResponder()).run(script, { topic: 't', runDir: 'research/x', model: 'sonnet', judgmentModel: 'sonnet' }), /binDir unset/)
})

test('W1.6: the pinned claim unit reaches synthesis and every blue response', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('blue-respond', 'G1')]) }), passChair()] }))
  await world.run(script, ARGS)
  for (const seat of ['blue-synthesize', 'blue-respond']) {
    const p = firstPrompt(world, seat)
    assert.ok(/claim_count with the tool/.test(p) && /never hand-count/.test(p), `${seat} missing the pinned claim unit`)
  }
})

test('the lens agent type is one configuration per area, and the areas dispatched are the plan\'s', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('red-lens-adversary'), party('red-lens-architecture')]) }), passChair()] }))
  await world.run(script, { ...ARGS, lensAreas: ['evidence', 'adversary', 'architecture'] })
  const types = labelsOf(world, 'red-lens').map((c) => c.opts.agentType)
  assert.deepEqual(types, ['frank-exchange-of-views:red-lens-adversary', 'frank-exchange-of-views:red-lens-architecture'])
  assert.equal(labelsOf(world, 'red-lens-evidence').length, 0, 'an area the plan did not ready does not sit — the record decides, not the roster')
})

test('the lens constitutions say a lens mints and the chair runs the debate', () => {
  const chair = readFileSync(new URL('../../agents/red-chair.md', import.meta.url), 'utf8')
  const lens = readFileSync(new URL('../../agents/red-lens-evidence.md', import.meta.url), 'utf8')
  assert.ok(!/COALESCE/.test(chair) && !/board's only writer/.test(chair), 'the chair constitution still describes the coalescing merge')
  assert.ok(/dispatch/.test(chair), 'the chair constitution names its dispatch duty')
  assert.ok(/mint/i.test(lens), 'the lens constitution names minting')
})

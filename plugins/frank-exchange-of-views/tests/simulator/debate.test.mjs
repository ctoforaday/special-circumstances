import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync, readdirSync } from 'node:fs'
import { loadDebateScript, makeWorld, makeResponder, blueEnv, chairEnv, passChair, plan, exchanges, passPlan, ceilingPlan, party, petitionBench, blocker, withKey, judgeEnv, petitionRulingEnv, schemaAccepts, strictTypesWarnings } from './harness.mjs'

// THE SIMULATOR DRIVES debate.js WITH STUBBED SEATS. There are no rounds (plans/roundless.md
// §III.B.1): the chair sits, its envelope relays the plan `dispatch next` recorded, the loop
// dispatches the parties the plan names, and the debate ends when the plan is empty. So a
// stubbed chair IS a stubbed plan, and a scenario is a sequence of plans.
const script = loadDebateScript(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url))
const ARGS = { topic: 'test topic', runDir: 'research/2026-01-01_test', lanes: 3, model: 'sonnet', judgmentModel: 'sonnet', binDir: '/opt/feov/bin' }
const labelsOf = (world, prefix) => world.calls.filter((c) => c.opts.label.startsWith(prefix))
const firstPrompt = (world, prefix) => labelsOf(world, prefix)[0].prompt
const CITATION_CLAUSE = 'THE REPORT DOES NOT NAME ITS SOURCES'
// THE LENS AREAS ARE READ FROM THE LENS DEFINITIONS, one `agents/red-lens-<area>.md` each, so a
// test over every area holds no list of them: the script's own areas are its generated seat block,
// and a definition the block does not seat fails the default-cast test below.
const LENS_AREAS = readdirSync(new URL('../../agents/', import.meta.url)).map((f) => /^red-lens-(.+)\.md$/.exec(f)).filter(Boolean).map((m) => m[1])
if (!LENS_AREAS.length) throw new Error('no agents/red-lens-<area>.md — every test over the lens areas would pass on nothing')

// THE SPEED CLAUSE IS READ FROM ITS ONE DEFINITION, NOT RESTATED HERE.
//
// This test asserted `includes('batch INDEPENDENT tool calls')` — a phrase of the clause's PROSE —
// so rewording the clause broke a test about whether seats CARRY it. That is the defect it is
// meant to catch, one level up: a fact recovered from a string. Reading the literal means any
// wording change is covered by construction and only a seat actually losing the clause reds.
//
// IT REFUSES WHAT IT CANNOT READ. The clause is a plain template literal today; an interpolation
// added to it would make this text differ from what a prompt carries, so `${` here is a hard
// failure rather than a silently weakened assertion.
const speedClauseText = (() => {
  const src = readFileSync(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url), 'utf8')
  const m = /^const speedClause = `([^`]*)`/m.exec(src)
  if (!m) throw new Error('debate.js no longer declares `const speedClause` as a single template literal — this bind cannot read it')
  if (m[1].includes('${')) throw new Error('speedClause now interpolates; teach this bind to render it rather than comparing the literal')
  return m[1]
})()

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
  await assert.rejects(world.run(script, { ...ARGS, maxRounds: 12 }), (e) => /not a term of this engine/.test(e.message) && /k-max/.test(e.message) && /mint-budget/.test(e.message) && /max-epochs/.test(e.message) && !/--k-max/.test(e.message))
  assert.equal(world.calls.length, 0, 'refused before anything was dispatched')
})

// ── a seat's failed sitting ─────────────────────────────────────────────────────────────────

// A SEAT'S FAILED SITTING DOES NOT STOP THE RUN (ruled by gblock 2026-10-27/28; universe-m35 died of
// one at epoch 6). Two failures reach the engine: the harness THROWS when a seat dispatched with a
// schema ends without its envelope, and agent() returns NULL when the agent dies on an API error.
// failing(prefix, n, how) fails the first n sittings whose label starts with prefix, then serves the
// responder's envelopes.
const NUDGE_ERROR = 'agent({schema}): subagent completed without calling StructuredOutput (after in-conversation nudge)'
const failing = (prefix, n, how, seats) => {
  let left = n
  return (p, o) => {
    if (o.label.startsWith(prefix) && left > 0) {
      left--
      if (how === 'throws') throw new Error(NUDGE_ERROR)
      return null
    }
    return seats(p, o)
  }
}
const engaged = () => chairEnv() // the evidence lens and blue on G-00000001
const docketed = () => chairEnv({ plan: plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }) })
const petitioned = () => chairEnv({ plan: plan([petitionBench()], { blockers: [blocker('M-00000001', 'judge')] }) })

// EACH SINGLE DISPATCH IN THE LOOP: the seat fails once, the run goes on, the next plan readies it
// again (the record still lists what it owes), it sits, and the run reaches its outcome.
for (const how of ['throws', 'returns null']) {
  for (const [name, prefix, owes] of [
    ['blue', 'blue-respond', engaged],
    ['the bench on its docket', 'judge #', docketed],
    ['the bench hearing a petition', 'judge · petition', petitioned],
  ]) {
    test(`${name}: a sitting that ${how} is dispatched again in the next epoch, and the run reaches its outcome`, async () => {
      const world = makeWorld(failing(prefix, 1, how, makeResponder({ chair: [owes(), owes(), passChair()] })))
      const out = await world.run(script, ARGS)
      assert.equal(out.verdict, 'VERIFIED')
      assert.equal(labelsOf(world, prefix).length, 2, `${prefix} sat in epoch 1, failed, and sat again in epoch 2`)
      assert.ok(world.logs.some((l) => l.includes(prefix) && (how === 'throws' ? l.includes(NUDGE_ERROR) : l.includes('returned nothing'))), 'the failure is logged with what the harness said')
      assert.equal(labelsOf(world, 'judge · assemble').length, 1, 'the assembly sitting records the outcome')
    })
  }
}

// THE CHAIR RELAYS THE PLAN, so a chair sitting that fails dispatches nobody that epoch and the chair
// sits again; failures that are not in a row stop nothing.
test('a chair sitting that fails is sat again in the next epoch, and failures not in a row stop nothing', async () => {
  const seats = makeResponder({ chair: [engaged(), engaged(), passChair()] })
  let n = 0
  const world = makeWorld((p, o) => {
    if (o.label.startsWith('red-chair')) {
      n++
      if (n === 1 || n === 3 || n === 4) throw new Error(NUDGE_ERROR)
    }
    return seats(p, o)
  })
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  assert.equal(labelsOf(world, 'red-chair').length, 6, 'fail, plan, fail, fail, plan, PASS')
  assert.equal(labelsOf(world, 'blue-respond').length, 2, 'nobody sits in an epoch whose chair relayed no plan')
})

// NO_PROGRESS_EPOCHS CHAIR FAILURES IN A ROW STOP THE DEBATE UNVERIFIED, naming the chair, and the
// assembly sitting stamps UNVERIFIED with that reason.
test('three chair failures in a row stop the run UNVERIFIED, naming the chair', async () => {
  const world = makeWorld(failing('red-chair', 3, 'throws', makeResponder()))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'UNVERIFIED')
  assert.equal(labelsOf(world, 'red-chair').length, NO_PROGRESS_EPOCHS)
  assert.deepEqual(out.termination, { seat_stop: { seat: 'red-chair', failures: NO_PROGRESS_EPOCHS } })
  const assembly = firstPrompt(world, 'judge · assemble')
  assert.ok(/it is UNVERIFIED, ended UNVERIFIED — the engine stopped the debate on a seat that kept failing: seat failure: red-chair failed 3 sittings in a row/.test(assembly), `the stamp names the chair: ${assembly.slice(0, 600)}`)
})

// THE SYNTHESIS AND THE ASSEMBLY HAVE NO EPOCH TO BE RE-DISPATCHED IN, so each is sat again in place.
test('a failed synthesis is sat again in place; three in a row stop the debate before it opens, UNVERIFIED', async () => {
  const once = makeWorld(failing('blue-synthesize', 1, 'returns null', makeResponder()))
  assert.equal((await once.run(script, ARGS)).verdict, 'VERIFIED')
  assert.equal(labelsOf(once, 'blue-synthesize').length, 2)
  const thrice = makeWorld(failing('blue-synthesize', 3, 'throws', makeResponder()))
  const out = await thrice.run(script, ARGS)
  assert.equal(out.verdict, 'UNVERIFIED')
  assert.equal(labelsOf(thrice, 'red-chair').length, 0, 'the debate does not open')
  assert.deepEqual(out.termination, { seat_stop: { seat: 'blue-synthesize', failures: NO_PROGRESS_EPOCHS } })
})

test('a failed assembly sitting is sat again in place; three in a row stop the run naming the verdict it could not record', async () => {
  const once = makeWorld(failing('judge · assemble', 2, 'throws', makeResponder({ chair: [passChair()] })))
  const out = await once.run(script, ARGS)
  assert.equal(out.gaps_outstanding, 0)
  assert.equal(labelsOf(once, 'judge · assemble').length, 3)
  const thrice = makeWorld(failing('judge · assemble', 3, 'returns null', makeResponder({ chair: [passChair()] })))
  await assert.rejects(thrice.run(script, ARGS), (e) => /^the assembly sitting failed 3 sittings in a row/.test(e.message) && /the debate had ended VERIFIED/.test(e.message))
})

// THE FRONTIER AND THE TERMINAL SITTING: a failure is logged and the run goes on.
test('a failed frontier or terminal sitting is logged and the run goes on', async () => {
  const frontier = makeWorld(failing('frontier', 1, 'throws', makeResponder({ chair: [passChair()] })))
  assert.equal((await frontier.run(script, ARGS)).verdict, 'VERIFIED')
  const held = chairEnv({ plan: plan([], { blockers: [blocker('M-00000001', 'judge')], why: ['nobody ready'] }) })
  const terminal = makeWorld(failing('judge · terminal', 1, 'throws', makeResponder({ chair: [held] })))
  const out = await terminal.run(script, ARGS)
  assert.equal(labelsOf(terminal, 'judge · terminal').length, 1)
  assert.equal(labelsOf(terminal, 'judge · assemble').length, 1)
  assert.equal(out.verdict, 'UNVERIFIED')
})

test('a chair envelope without the plan\'s fields aborts: the envelope is the verb\'s JSON, relayed verbatim', async () => {
  const world = makeWorld(makeResponder({ chair: [{ notes: 'n' }] }))
  await assert.rejects(world.run(script, ARGS), /red-chair sitting 1 relayed a plan whose `key` is undefined/)
})

// ── the dispatch loop ───────────────────────────────────────────────────────────────────────

test('the plan is what dispatches: the parties it names sit, in role order — lenses, then blue, then the bench', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G-00000001'), party('judge', 'G-00000002'), party('red-lens-logic', 'G-00000001'), party('red-lens-evidence')], { docket: ['G-00000002'] }) }), passChair()],
  }))
  const out = await world.run(script, ARGS)
  // The bench's closing sittings head with `judge` as well now, so the debate filter names the
  // docket sitting by its ordinal form and excludes `judge · terminal` / `judge · assemble`.
  const debate = world.calls.filter((c) => /^(red-lens|blue-respond|judge #)/.test(c.opts.label))
  const order = debate.map((c) => c.opts.label.split(' ')[0])
  assert.deepEqual(order, ['red-lens-logic', 'red-lens-evidence', 'blue-respond', 'judge'], 'red parties first, then blue, then the bench — an exchange is red followed by blue')
  assert.equal(out.verdict, 'VERIFIED')
  assert.equal(out.epochs, 2, 'two chair sittings')
})

test('a lens is told to read its pin from its work list; one engaged on gaps is told which, and that silence is a null turn', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('red-lens-logic', 'G-00000003', 'G-00000004')], { head: 7 }) }), passChair()],
  }))
  await world.run(script, ARGS)
  const evidence = firstPrompt(world, 'red-lens-evidence')
  const logic = firstPrompt(world, 'red-lens-logic')
  for (const [name, p] of [['evidence', evidence], ['logic', logic]]) {
    assert.ok(/YOUR PIN IS ON YOUR WORK LIST: read `sitting\.last_read` first/.test(p), `${name}: the sitting is read from the record`)
    for (const kind of ['first', 'behind', 'unchanged', 'served', 'undispatched']) assert.ok(p.includes(`\`${kind}\``), `${name}: the ${kind} kind is explained`)
    assert.ok(/why you missed it then/.test(p), `${name}: a fresh gap on already-read text owes why it was missed`)
    assert.ok(!/head 7/.test(p) && !/HEAD MOVED/.test(p), `${name}: no head clause is built from the relay`)
  }
  assert.ok(/YOU ARE ENGAGED ON: G-00000003, G-00000004/.test(logic), 'the engaged lens is told its gaps')
  assert.ok(/NULL TURN/.test(logic) && /silence is a turn taken/.test(logic), 'a null turn counts toward impasse')
  assert.ok(/DID BLUE ACTUALLY DO WHAT YOU ASKED/.test(logic), 'the lens compares its fix to blue\'s edits')
})

test('a party the workflow has no sitting for aborts — the cast and the workflow must agree', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('frontier')]) })] }))
  await assert.rejects(world.run(script, ARGS), /names frontier, which this workflow has no sitting for/)
})

test('labels carry the seat and its sitting ordinal, never a round or a lens number', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv(), chairEnv({ plan: plan([party('red-lens-evidence', 'G-00000001')]) }), passChair()],
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
    chair: [chairEnv({ plan: plan([party('red-lens-evidence', 'G-00000001'), party('blue-respond', 'G-00000001'), party('judge', 'G-00000002')], { head: 5, docket: ['G-00000002'] }) }), passChair()],
  }))
  await world.run(script, ARGS)
  const all = world.logs.join('\n')
  assert.ok(/epoch 1: head 5 — 3 party\(ies\) ready, docketed G-00000002/.test(all), all)
  assert.ok(/epoch 1: dispatching 1 red lens\(es\): evidence\[G-00000001\]/.test(all))
  assert.ok(/epoch 1: blue responded on G-00000001/.test(all))
  assert.ok(/epoch 1: the bench sits on G-00000002/.test(all))
  assert.ok(/epoch 2: head 2 — 0 party\(ies\) ready, chair recorded PASS — PASS permitted/.test(all))
})

// ── termination: three ways, from the plan ──────────────────────────────────────────────────

test('VERIFIED: the chair records PASS on a plan that permits it; phases in order; lanes honoured', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  const out = await world.run(script, ARGS)
  assert.deepEqual({ verdict: out.verdict, epochs: out.epochs, gaps: out.gaps_outstanding, term: out.termination },
    { verdict: 'VERIFIED', epochs: 1, gaps: 0, term: { pass_permitted: true, ceiling: false, epoch_limit_reached: false, why: [], no_progress: null, seat_stop: null } })
  assert.deepEqual(world.phases, ['Frontier', 'Blue', 'Red', 'Assemble'])
  assert.equal(labelsOf(world, 'blue-lane').length, 3)
  assert.ok(world.logs.some((m) => m.includes('researching: test topic')))
  assert.ok(!world.calls.some((c) => c.opts.label.startsWith('judge · terminal')), 'no unruled motion, no terminal sitting')
})

test('CEILING: the run at its epoch limit — the stamp says so and names no deadlock', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: ceilingPlan() })] }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'CEILING')
  assert.equal(out.termination.ceiling, true)
  const asm = firstPrompt(world, 'judge · assemble')
  assert.ok(/it is CEILING/.test(asm) && /NOT a judged failure to verify/.test(asm), asm.slice(0, 600))
  assert.ok(/Record it as the run's outcome — CEILING, ended at the ceiling/.test(asm), 'the bench is told the exact stamp, as an act and not a command line')
  assert.ok(!/deadlock/i.test(asm) || /not a judged/i.test(asm), 'deadlock is per gap and on the record, not a run-level stamp')
})

// A CEILING over an unchecked citation or proof, or one on a failed check, names it in the reason the
// run ends on, and names no other blocker; with none standing, each reason is the sentence it is
// without the rule. The record's derived reason carries the same clause (TestACeilingReasonNamesOnlyAStandingReopenedAnchor).
test('CEILING: the reason names each unchecked citation or proof and each on a failed check, and only those', async () => {
  const ended = async (chair) => {
    const world = makeWorld(makeResponder({ chair }))
    await world.run(script, ARGS)
    return world.logs.find((m) => m.startsWith('debate ended: CEILING')).replace(/^.*? chair sitting\(s\) — /, '')
  }
  const standing = { blockers: [blocker('G-00000001', 'red-lens-logic', 'material_gap'), blocker('C-706fda63', 'red-lens-evidence', 'unchecked'), blocker('P-cd9f4c5a', 'red-lens-evidence', 'unchecked'), blocker('C-0b0e0d0a', 'red-lens-evidence', 'failed_check'), blocker('C-0b0e0d0a', 'blue-respond', 'failed_check')] }
  const none = { blockers: [blocker('G-00000001', 'red-lens-logic', 'material_gap')] }
  const term = "epoch limit 4 reached — the run's term on chair sittings; the parties the board still readied were not dispatched"
  const clause = '; unchecked: C-706fda63, P-cd9f4c5a; on a failed check: C-0b0e0d0a'
  assert.equal(await ended(toTheLimit(limitPlan(none))), term)
  assert.equal(await ended(toTheLimit(limitPlan(standing))), term + clause)
})

test('UNVERIFIED: nobody ready, neither PASS nor CEILING — the plan\'s reasons are the account', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([], { why: ['G-00000001: docket motion M-00000001 stands unruled and the bench has sat since it was filed — one bench sitting per docketing, so this gap is not re-readied'], blockers: [blocker('M-00000001', 'judge')] }) })] }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'UNVERIFIED')
  assert.deepEqual(out.termination, { pass_permitted: false, ceiling: false, epoch_limit_reached: false, why: ['G-00000001: docket motion M-00000001 stands unruled and the bench has sat since it was filed — one bench sitting per docketing, so this gap is not re-readied'], no_progress: null, seat_stop: null })
  const asm = firstPrompt(world, 'judge · assemble')
  assert.ok(/ended UNVERIFIED/.test(asm) && /the bench has sat since it was filed/.test(asm), 'the reason reaches the stamp')
})

// ── the no-progress valve: a plan whose key repeats stops the debate ──
//
// The B5 loop (2026-09-11): the voice lens was readied alone at head 144 for eleven epochs, because
// it never registered and so never sat. The chair's stub here repeats that plan well past the
// valve and passes only at its tenth sitting, so a loop with no valve ends VERIFIED at epoch 10 and
// fails every assertion below rather than hanging.
const NO_PROGRESS_EPOCHS = 3
const b5Plan = () => chairEnv({ plan: plan([party('red-lens-voice')], { head: 144, why: ['red-lens-voice: active, never sat — audits the report at head 144'] }) })
const repeated = (make, n) => Array.from({ length: n }, make)

test('NO PROGRESS: a plan the same for NO_PROGRESS_EPOCHS epochs stops the debate UNVERIFIED, naming the stuck parties and the head', async () => {
  const world = makeWorld(makeResponder({ chair: [...repeated(b5Plan, 9), passChair()] }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'UNVERIFIED')
  assert.equal(out.epochs, NO_PROGRESS_EPOCHS, 'the valve stops at the sitting that repeats the plan the third time')
  assert.equal(labelsOf(world, 'red-lens-voice').length, NO_PROGRESS_EPOCHS - 1, 'the repeated plan is not dispatched again')
  assert.deepEqual(out.termination.no_progress, { epochs: NO_PROGRESS_EPOCHS, head: 144, parties: [{ seat_id: 'red-lens-voice', gap_ids: [] }] })
  assert.equal(out.termination.ceiling, false)
  const line = world.logs.find((m) => /NO PROGRESS/.test(m))
  assert.ok(line && line.includes(`NO_PROGRESS_EPOCHS = ${NO_PROGRESS_EPOCHS}`) && /head 144/.test(line) && /red-lens-voice/.test(line), `the log names the constant, the head and the stuck party: ${line}`)
  const asm = firstPrompt(world, 'judge · assemble')
  assert.ok(/ended UNVERIFIED/.test(asm) && /no progress/.test(asm) && /head 144/.test(asm) && /red-lens-voice/.test(asm) && /not CEILING/.test(asm), asm.slice(0, 900))
  const lastChair = world.calls.map((c) => c.opts.label).lastIndexOf(labelsOf(world, 'red-chair').at(-1).opts.label)
  assert.deepEqual(world.calls.slice(lastChair + 1).map((c) => c.opts.label.replace(/ · [^·]+$/, '')), ['judge · assemble'],
    'nobody sits after the stop but the assembly, which is the bench')
})

test('a plan that changes keeps going — the head moving, an exchange count advancing, or a repeat that is not consecutive', async () => {
  const heads = [2, 3, 4, 5, 6].map((head) => chairEnv({ plan: plan([party('red-lens-voice')], { head }) }))
  const moving = await makeWorld(makeResponder({ chair: [...heads, passChair()] })).run(script, ARGS)
  assert.deepEqual([moving.verdict, moving.epochs, moving.termination.no_progress], ['VERIFIED', 6, null])

  // k and k-max bound this march; the valve must not cut it short at a head that does not move.
  // The reasons are the same at every sitting: the count is what marches.
  const marchers = [party('red-lens-logic', 'G-00000001'), party('blue-respond', 'G-00000001')]
  const march = [0, 1, 2, 3].map((n) => chairEnv({ plan: plan(marchers, { head: 9, exchanges: exchanges(marchers, n), why: ['G-00000001: open, material — below its limits'] }) }))
  const marching = await makeWorld(makeResponder({ chair: [...march, passChair()] })).run(script, ARGS)
  assert.deepEqual([marching.verdict, marching.epochs], ['VERIFIED', 5])

  // A sitting the record cannot close is not a count of zero, and it moves the plan as a count does.
  const unclosed = [0, 1, 0, 1].map((unresolved) => chairEnv({ plan: plan(marchers, { head: 9, exchanges: [{ gap_id: 'G-00000001', count: 0, unresolved }] }) }))
  const measuring = await makeWorld(makeResponder({ chair: [...unclosed, passChair()] })).run(script, ARGS)
  assert.deepEqual([measuring.verdict, measuring.epochs], ['VERIFIED', 5])

  const other = chairEnv({ plan: plan([party('red-lens-evidence')], { head: 144 }) })
  const broken = await makeWorld(makeResponder({ chair: [b5Plan(), b5Plan(), other, b5Plan(), b5Plan(), passChair()] })).run(script, ARGS)
  assert.deepEqual([broken.verdict, broken.epochs], ['VERIFIED', 6])
})

// THE VALVE READS THE KEY AND NOTHING ELSE. Plans the same in every field but their key are different
// plans to it, and plans different in every field but their key are one: the key is the record's
// statement of what the plan decides, and capture holds the relay to the stored plan, key included.
test('the valve keys on the plan\'s key alone', async () => {
  const sameKey = [0, 1, 2, 3].map((n) => chairEnv({ plan: plan([party('red-lens-voice')], { head: 144 + n, key: '00000000000000aa' }) }))
  const stopped = await makeWorld(makeResponder({ chair: [...sameKey, passChair()] })).run(script, ARGS)
  assert.deepEqual([stopped.verdict, stopped.epochs], ['UNVERIFIED', NO_PROGRESS_EPOCHS])
  const keys = [0, 1, 2, 3].map((n) => chairEnv({ plan: plan([party('red-lens-voice')], { head: 144, key: `00000000000000a${n}` }) }))
  const moving = await makeWorld(makeResponder({ chair: [...keys, passChair()] })).run(script, ARGS)
  assert.deepEqual([moving.verdict, moving.epochs], ['VERIFIED', 5])
})

// A BLOCKER OF ANY KIND THAT CHANGES MOVES THE PLAN: the record's key covers the whole plan, so a
// sitting whose only change is what holds the PASS is progress the valve does not cut short.
test('a plan whose blockers alone change keeps going, whatever their kind', async () => {
  const stuck = (b) => chairEnv({ plan: plan([party('red-lens-voice')], { head: 144, blockers: [b] }) })
  for (const changing of [
    [0, 1, 2, 3].map((n) => stuck(blocker(`G-0000000${n}`, 'red-lens-logic', 'material_gap'))),
    [0, 1, 2, 3].map((n) => stuck(blocker(`M-0000000${n}`, 'judge'))),
  ]) {
    const moving = await makeWorld(makeResponder({ chair: [...changing, passChair()] })).run(script, ARGS)
    assert.deepEqual([moving.verdict, moving.epochs], ['VERIFIED', 5])
  }
})

// ── the epoch limit: a run TERM, and the record's word ────────────
const limitPlan = (over = {}) => plan([], { ceiling: true, epoch_limit_reached: true, max_epochs: 4,
  why: ['G-00000001: open, material, 1 exchange(s) (0 stalled) — below its limits', "epoch limit 4 reached — this chair sitting opens the run's last epoch, so the 2 party(ies) above are not dispatched"],
  blockers: [{ kind: 'material_gap', subject: 'G-00000001', owner: 'red-lens-evidence' }], ...over })
const toTheLimit = (last) => [...[11, 12, 13].map((head) => chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G-00000001')], { head }) })), chairEnv({ plan: last }), passChair()]

test('EPOCH LIMIT: the plan at the last epoch ends the debate CEILING with the limit named', async () => {
  const world = makeWorld(makeResponder({ chair: toTheLimit(limitPlan()) }))
  const out = await world.run(script, ARGS)
  assert.deepEqual([out.verdict, out.epochs, out.termination.epoch_limit_reached, out.termination.ceiling], ['CEILING', 4, true, true])
  assert.ok(out.termination.why.some((w) => /epoch limit 4 reached/.test(w)), 'the record\'s reason travels in the termination')
  const asm = firstPrompt(world, 'judge · assemble')
  assert.ok(/it is CEILING/.test(asm) && /epoch limit 4 reached/.test(asm) && !/every open material gap reached its limit/.test(asm), asm.slice(0, 900))
  assert.ok(world.logs.some((m) => /epoch 4: the epoch limit \(4\) is reached/.test(m)))

})

test('EPOCH LIMIT: a relayed plan that says the limit is reached dispatches nobody, whatever parties it still lists', async () => {
  const world = makeWorld(makeResponder({ chair: toTheLimit(limitPlan({ parties: [party('red-lens-evidence'), party('blue-respond', 'G-00000001')] })) }))
  const out = await world.run(script, ARGS)
  assert.deepEqual([out.verdict, out.epochs], ['CEILING', 4])
  assert.equal(labelsOf(world, 'red-lens-evidence').length, 3, 'three epochs dispatched the lens; the fourth dispatched nobody')
  assert.equal(labelsOf(world, 'blue-respond').length, 3)
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

test('the terminal bench sitting fires only when the last plan holds a blocker the bench owns, and before assembly', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: ceilingPlan({ blockers: [blocker('M-00000001', 'judge'), blocker('M-00000003', 'judge'), blocker('M-00000002', 'red-chair')] }) })] }))
  await world.run(script, ARGS)
  const terminal = labelsOf(world, 'judge · terminal')[0]
  const asm = labelsOf(world, 'judge · assemble')[0]
  assert.ok(terminal, 'the terminal sitting fired')
  assert.ok(terminal.n < asm.n, 'disposition precedes assembly')
  assert.ok(/2 motion\(s\) whose gavel is yours stand unruled on the record: M-00000001, M-00000003/.test(terminal.prompt) && /NOTHING CAN BE REMANDED AT A TERMINAL EXIT/.test(terminal.prompt))
  // A motion whose gavel is the chair's never convenes it: the bench has no verb to rule one.
  const chairOnly = makeWorld(makeResponder({ chair: [chairEnv({ plan: ceilingPlan({ blockers: [blocker('M-00000002', 'red-chair')] }) })] }))
  await chairOnly.run(script, ARGS)
  assert.ok(!labelsOf(chairOnly, 'judge · terminal').length, 'a chair-owned blocker convened the terminal bench')
  // Only a MOTION the bench owns convenes it: the sitting is told to rule motions.
  const notAMotion = makeWorld(makeResponder({ chair: [chairEnv({ plan: ceilingPlan({ blockers: [blocker('G-00000001', 'judge', 'stranded_gap')] }) })] }))
  await notAMotion.run(script, ARGS)
  assert.ok(!labelsOf(notAMotion, 'judge · terminal').length, 'a bench-owned blocker that is not a motion convened the terminal bench')
})

// A PETITION FILED IN THE FINAL EPOCH IS HEARD AT THE TERMINAL SITTING, and what it rules travels as a
// petition sitting's does: the ruling is in the returned petitions, and a halt stamps the run HALTED.
test('a petition standing at the exit: the terminal sitting rules it, and its halt stamps the run HALTED', async () => {
  const lastEpoch = () => chairEnv({ plan: plan([], { ceiling: true, epoch_limit_reached: true, max_epochs: 2, blockers: [blocker('M-00000001', 'judge')] }) })
  const ruling = { petitioner: 'red-lens-dark-side', class: 'safety', ruling: 'granted', opinion: 'the hazard is real', relief: 'withhold the protocol', binds: 'both' }
  const halting = makeWorld(makeResponder({ chair: [chairEnv(), lastEpoch()], judge: [judgeEnv({ rulings: [ruling], halt: { opinion: 'the human must decide this' } })] }))
  const out = await halting.run(script, ARGS)
  const terminal = labelsOf(halting, 'judge · terminal')[0]
  assert.ok(terminal, 'the terminal sitting fired on the standing petition')
  assert.equal(terminal.opts.schema.properties.halt.required[0], 'opinion', 'the terminal envelope carries the halt')
  assert.ok(terminal.opts.schema.properties.rulings, 'the terminal envelope carries petition rulings')
  assert.ok(/A HALT IS A DIFFERENT DECISION FROM A RULING/.test(terminal.prompt) && /stamps the run HALTED/.test(terminal.prompt))
  assert.equal(out.verdict, 'HALTED')
  assert.equal(out.halted, true)
  assert.ok(out.halt_opinion.includes('the human must decide'))
  assert.deepEqual(out.petitions, [{ petitioner: 'red-lens-dark-side', class: 'safety', ruling: 'granted', opinion: 'the hazard is real' }])
  const asm = firstPrompt(halting, 'judge · assemble')
  assert.ok(asm.includes('it is HALTED') && asm.includes('the human must decide'), 'the assembly stamps the halt the terminal sitting recorded')
  // Ruled without a halt, the run keeps its stamp and the ruling still travels.
  const ruled = makeWorld(makeResponder({ chair: [chairEnv(), lastEpoch()], judge: [judgeEnv({ rulings: [{ ...ruling, ruling: 'denied', relief: undefined }] })] }))
  const out2 = await ruled.run(script, ARGS)
  assert.equal(out2.verdict, 'CEILING')
  assert.equal(out2.petitions.length, 1)
  assert.equal(out2.petitions[0].ruling, 'denied')
  // THE STAMP IS THE LAST PLAN'S, WHATEVER THE TERMINAL SITTING RULED. The engine reads the record
  // only through the chair's plan, so the assembly is told CEILING by the epoch limit after the
  // ruling that left nobody ready — the word `bench outcome` must then derive from the record.
  const stamp = firstPrompt(ruled, 'judge · assemble')
  assert.ok(/it is CEILING/.test(stamp) && /epoch limit 2 reached/.test(stamp) && /Record it as the run's outcome — CEILING/.test(stamp), stamp.slice(0, 900))
  assert.equal(out2.termination.epoch_limit_reached, true)
})

// ── the seats' contracts ────────────────────────────────────────────────────────────────────

test('the chair runs the debate: relays the plan verbatim, mints nothing, closes nothing, owns the verdict', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  await world.run(script, ARGS)
  const chair = labelsOf(world, 'red-chair')[0]
  const p = chair.prompt
  assert.ok(/you mint nothing and you close nothing/.test(p) && /ASK THE RECORD WHO SITS — the dispatch/.test(p) && /Your envelope IS the object it prints, its fields copied VERBATIM/.test(p) && /when you have filed or ruled a motion since you asked, and relay THAT plan; your verdict prints the plan you relay/.test(p))
  assert.ok(/THE STOPPING JUDGMENT IS YOURS/.test(p) && /record a PASS verdict/.test(p) && /or a FAIL if the board as it stands still stops you, which it does until it converges/.test(p))
  assert.ok(!/for its lens to mint|raise something material/i.test(p), 'the chair mints nothing, so its stopping judgment names no lens act for a defect it finds')
  assert.ok(/ANSWER EVERY INQUIRY ON ONE READ/.test(p) && /where neither has changed, your last read stands/.test(p) && /RULE ON BLUE'S INQUIRIES/.test(p) && /RULE THE GRADE MOTIONS/.test(p))
  // THE POSITION AND THE SPOT-CHECK ARE THE CHAIR'S DUTIES, and the constitution is their one carrier:
  // the prompt, which every chair sitting receives unchanged, restates neither. A docketed gap is its
  // parties' to argue, never the chair's.
  const chairMd = readFileSync(new URL('../../agents/red-chair.md', import.meta.url), 'utf8')
  assert.ok(/A docketed gap is argued by its parties — its minting lens and blue — in their closings, never by you/.test(chairMd) && /name in it every stale area the plan lists, and any archived closure you re-read/.test(chairMd) && /YOUR POSITION ARGUES TO THE BENCH/.test(chairMd), 'the chair constitution lost the position, the parties\' closings or the spot-check')
  assert.ok(!/CLOSING ARGUMENTS/.test(p) && !/YOUR POSITION ARGUES TO THE BENCH/.test(p) && !/A CLOSURE IS A CLAIM/.test(p), 'the chair prompt restates a duty its constitution carries')
  assert.ok(!/dispatch next|--as |show (board|work|motions)|motion (grade|inquiry) rule/.test(p), 'the chair is told the ACT; the word it types is the help page\'s')
  assert.ok(!/COALESCE/.test(p) && !/only writer/.test(p), 'the coalescing chair is gone with the mechanism')
  const schema = chair.opts.schema
  assert.ok(schema.required.includes('parties') && schema.properties.verdict && !schema.properties.plan && schema.additionalProperties === false, 'the envelope IS the plan, closed, and the verdict is its field')
  assert.ok(!chair.opts.schema.properties.gaps, 'and no gap list — the record holds the board')
  assert.ok(!/SCORECARD/i.test(p), 'no prompt sends a seat to a scorecard: the card is retired and `show work` carries the situation (#1126)')
})

// THE CHAIR IS TOLD THE READINESS IT RELAYS AND THE PASS IT MAY RECORD: a lens sits again only when
// the report head has moved past its pin and, owing nothing, retires at three sittings in a
// row with no fresh material, a PASS needs no lens ready, the stale areas are spot-checked first, and the PASS
// lists every open gap that is not material by class. Each clause is the text the gate enforces.
test('the chair is told a lens sits again only when the head passes its pin and retires at three barren sittings, a PASS needs no lens ready, stale areas are spot-checked, and the PASS lists what is not material by class', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  await world.run(script, ARGS)
  const p = labelsOf(world, 'red-chair')[0].prompt
  assert.ok(/RECORDS who sits — each lens that has never sat or whose pin the report head has moved past,/.test(p), 'the dispatch is described by the head and the lens\'s pin')
  assert.ok(!/active lenses/.test(p) && !/re-arm/.test(p), 'no lens is described as ready at a head it has sat at')
  assert.ok(!/every sitting it is not empty/.test(p) && !/Re-sample the archive/.test(p), 'no sitting without a PASS in view owes a spot-check')
  // THE READINESS RULE AND THE PASS IT PERMITS ARE THE CONSTITUTION'S, ONCE. They hold at every chair
  // sitting, so the prompt carries none of them (TestEveryFragmentRendersOnceWhereItIsIncluded, rule 2).
  const chairMd = readFileSync(new URL('../../agents/red-chair.md', import.meta.url), 'utf8')
  assert.ok(/A lens with no gap engaged is readied when it has never sat or its pin is behind the head, and not otherwise: its pin is the head the tool last served it the whole report at, in a sitting it sat through, so a lens whose sitting ended without the report served to it at the head is readied again\. Three sittings in a row with no fresh material mint retire a lens that owes nothing, and a head move does not ready a retired lens; a gap a lens holds owes only what is new on it since that lens last sat\./.test(chairMd), 'ready only on a moved head, retired at three owing nothing')
  assert.ok(/permits a PASS only when no lens is ready — every lens has been served the report at the head or has retired — and nothing another seat must clear holds it\. BEFORE a PASS/.test(chairMd), 'the no-lens-ready PASS condition')
  assert.ok(/BEFORE a PASS, read the changes since each stale area's pin and name those areas in your spot-check; a defect you find there goes in that spot-check, and you record no verdict/.test(chairMd), 'the stale areas')
  assert.ok(/The spot-check is owed before a PASS and at no other time: name in it every stale area the plan lists/.test(chairMd), 'the spot-check is owed before a PASS, and names the stale areas')
  assert.ok(/YOUR PASS LISTS EVERY OPEN GAP THAT IS NOT MATERIAL, BY CLASS — your work list marks each — with one line on why it changes no reader decision, on the record/.test(chairMd), 'the by-class listing')
  assert.ok(!/A LENS SITS AGAIN/.test(p) && !/YOUR PASS LISTS EVERY OPEN GAP/.test(p) && !/The spot-check is owed before a PASS/.test(p), 'the chair prompt restates the readiness rule its constitution carries')
  for (const clause of ['is readied again. Three sittings in a row with no fresh material mint retire a lens that owes nothing', 'The plan permits a PASS only when no lens is ready', "each stale area's pin", 'YOUR PASS LISTS EVERY OPEN GAP THAT IS NOT MATERIAL, BY CLASS', 'each lens that has never sat or whose pin — the head the tool last served it the whole report at — the report head has moved past']) {
    assert.ok(chairMd.includes(clause), `the chair constitution and prompt disagree: ${clause}`)
  }
})

test('the lens mints its own gaps, screens first, spends a budget, and closes as the originator', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('red-lens-logic'), party('red-lens-dark-side')]) }), passChair()] }))
  await world.run(script, ARGS)
  const evidence = firstPrompt(world, 'red-lens-evidence')
  for (const want of ['YOU MINT YOUR OWN GAPS', 'read the board first', 'then mint', 'or the catch-all with the reason none does', 'mintBudget', 'names the gaps it supersedes', "says so in the check's kind", 'THE ORIGINATOR CLOSES', 'LINEAGE IS NEVER DROPPED',
    'ASK FOR THE ANSWER TO BE PRODUCED, NOT ASSERTED', 'DOCUMENT-PROBE', 'LIVE-PROBE', 'BELIEVE NO BYTES', 'PRESCRIBE TEXT ONLY WHERE THE DEFECT IS TEXTUAL',
    // A FIX BLUE CANNOT MAKE HOLDS THE GATE ON NOTHING.
    'A REQUIRED FIX AND ITS ACCEPTANCE CHECK NAME ONLY WHAT THE ANSWERING SEAT CAN CHANGE', 'close as a defect owed elsewhere, and log',
    'ANCHOR EVERY FINDING TO A QUOTED SENTENCE', "ids of your findings and of the gaps you mint are the tool's to assign", 'read it whole in consecutive windows',
    'counts LINES, not occurrences',
    "a gap that is not material — by its class, or graded below medium — does not hold the gate", "the stored script is what the re-run executes; a clean exit or '0 failing' is not an output", "read at the computation's anchor, gives the script and the output blue recorded"]) {
    assert.ok(evidence.includes(want), `the lens prompt lost: ${want}`)
  }
  assert.ok(!/gap ids are the chair's/.test(evidence), 'the ids are the tool\'s, minted by the lens')
  assert.ok(evidence.includes(CITATION_CLAUSE) && /VERBATIM READS ONLY/.test(evidence) && /--comments/.test(evidence), 'the evidence lens carries the ledger clause')
  // #1209: the manual's shared block MANDATES a quoted heredoc for free text, and a source a lens
  // finds is read through the run's cache so every later seat reads the same bytes.
  assert.ok(!/heredoc/.test(evidence) && !/curl/.test(evidence), 'the lens prompt contradicts its manual on heredocs or bypasses the shared source cache')
  assert.ok(/STEELMAN DUTY/.test(firstPrompt(world, 'red-lens-logic')) && /STEELMAN DUTY/.test(firstPrompt(world, 'red-lens-dark-side')), 'logic and dark-side audit the declines')
  // AN INQUIRY THAT CONTRADICTS WHAT THE READER IS TOLD IS A GAP, NOT A FINDING. The inquiries print
  // in the report beside the body, so the reader sees both halves. The lens reads the inquiry
  // sections as they ship before it grades one.
  for (const seat of ['red-lens-logic', 'red-lens-dark-side']) {
    const p = firstPrompt(world, seat)
    assert.ok(/that is a candidate gap, anchored to the inquiry/.test(p) && /one approach under two fates/.test(p) && /a fate the report's own text or a proof contradicts/.test(p) && /a section built on an approach that is no inquiry/.test(p),
      `${seat} is not told that an inquiry contradicting the report is a gap`)
    assert.ok(/read them in the report's assembled form before you grade an inquiry/.test(p), `${seat} is not told to read the inquiry sections as they ship`)
    assert.ok(/A contradiction left as a finding ships: a finding about an inquiry is on no seat's work list and obliges nobody/.test(p) && !/Both are findings/.test(p), `${seat} is told to file an inquiry contradiction as a finding`)
    // THE BAR IS THE ONE EVERY GAP MEETS. A clause that orders a mint for every mismatch spends a
    // lens's budget — three open gaps, in a smoke — on a duplicate row, against the same prompt's "not on
    // nitpicks"; and the lens that filed findings on universe-m21 had judged by the answer alone.
    assert.ok(/weighed like any other: by what it changes for a reader, who decides by what was tried as well as by the answer/.test(p) && !/YOU MINT A GAP/.test(p),
      `${seat} is told to mint whatever the mismatch changes for a reader`)
  }
  assert.ok(!/STEELMAN DUTY/.test(evidence) && !/take slice|instance \d+ of/.test(evidence), 'the evidence seat verifies sources and owns no slice')

  // THE FULL RE-READ IS SCOPED TO THE KIND THAT ORDERS IT, and this is the measurement that put it
  // there. #1089 made an empty sitting free, and across 13 empty sittings of the 2026-09-22 smoke
  // the affordance was taken 0 times: 13 of 13 registered, logged, and re-read the whole report.
  // The constitution had already been narrowed to excuse them; this prompt still said "RE-READ THE
  // FULL REPORT IN CONTEXT" flat in its body, one sentence after the clause explaining `unchanged`,
  // and the unconditional rule won. A seat obeys the strictest thing it is told.
  assert.ok(!/RE-READ THE FULL REPORT IN CONTEXT/.test(evidence),
    'the re-read must not be ordered unconditionally in the prompt body — scope it to `first` and `behind`')
  assert.ok(/`first`: audit the report in full/.test(evidence) && /`behind`:[^`]*in full/.test(evidence),
    'the kinds that DID move still order the full read')
  assert.ok(/`unchanged`:[^`]*NOTHING TO AUDIT/.test(evidence),
    'the kind that did not move must say there is nothing to audit')
  assert.ok(/the sitting owes no entry/.test(evidence),
    'an unchanged sitting with nothing engaging it owes no log entry')
})

test('blue is engaged on named gaps, works from the report and its work list, and files closings only where the plan holds one owed', async () => {
  const owed = [{ kind: 'material_gap', subject: 'G-00000001', owner: 'red-lens-logic' }, { kind: 'closing_owed', subject: 'G-00000002', owner: 'blue-respond' },
    { kind: 'closing_owed', subject: 'G-00000002', owner: 'red-lens-logic' }, { kind: 'unruled_motion', subject: 'M-00000001', owner: 'judge' }]
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G-00000001')], { docket: ['G-00000002'], blockers: owed }) }), chairEnv({ plan: plan([party('blue-respond', 'G-00000003')], { docket: ['G-00000003'] }) }), passChair()],
  }))
  await world.run(script, ARGS)
  const [first, second] = labelsOf(world, 'blue-respond').map((c) => c.prompt)
  assert.ok(/You are engaged on: G-00000001\./.test(first))
  for (const want of ['READ YOUR MANUAL FIRST', 'together, in one message', 'the report and your work list', 'YOUR WORK LIST CARRIES YOUR GAPS, EACH AS ONE WORK ITEM', 'that text beside the span it replaces',
    'the work items carry it, and the record is authoritative', 'the gaps the bench has ruled closed are on the list too', 'you name it in found_closed from that list',
    'REMANDED comes with a stated research direction you owe', 'YOU MAY COMPUTE AN ANSWER', 'DOCUMENT-PROBE', 'deferred acceptance test',
    'INQUIRIES ARE A LIVING RECORD', 'THREE paths', 'ESTOPS', 'OWNERSHIP BINDS, AS IT DID AT SYNTHESIS', "TL;DR, Catechism, the report's own sections, open questions, your citations", 'and the report carries no list of them', 'each edit naming the gap it answers', 'a grade motion on the axis', 'Compact and reorganize prose', 'retired on the record',
    'PROPAGATE EVERY CORRECTION TO ALL SITES', 'NULL TURN', 'AUDIT YOUR OWN REPAIRS, ONE RECEIPT PER GAP', 'manifest array',
    'where the gap changes no reader decision or asks for complexity that does not pay, argue `defect_accepted` with that reason', "materiality is the class's default: always, never, or by grade from medium"]) {
    assert.ok(first.includes(want), `blue lost: ${want}`)
  }
  // BLUE'S WORKING SET IS THE REPORT AND ITS WORK ITEMS (gblock's ruling, 2026-10-09): no sentence sends
  // it to the board, and none to red's argument for a gap — the lens's reason and the basis of a regrade.
  for (const banned of [/READ THE BOARD/, /the board and work projections/, /the board shows/, /the board you read/, /red's argument for it/, /basis of each change of grade/]) {
    assert.ok(!banned.test(first), `blue is sent to the board or to red's argument: ${banned}`)
  }
  // THE CLOSINGS OWED ARE THE PLAN'S closing_owed BLOCKERS THIS SEAT OWNS — not the lens's, and not
  // the docket the chair's sitting filed.
  assert.ok(/YOUR CLOSING ARGUMENT IS OWED ON: G-00000002 — each is DOCKETED/.test(first) && /argue in ~120 words/.test(first))
  assert.ok(!/YOUR CLOSING ARGUMENT IS OWED/.test(second), 'a docket with no closing the plan holds blue to owes none')
  assert.ok(!/round \d/.test(first), 'no round is named to blue')
  // #1209: blue rules nothing, so the bench's law clause is not blue's; every read is a projection,
  // so no scratch-file copy of one; and a clean sitting owes no log entry.
  assert.ok(!/ LAW: /.test(first) && !/concatenat/.test(first) && !/log entry saying so/.test(first), 'blue was handed a duty that is not its own')
})

test('the bench rules on docketed gaps from the closings, the transcript and the live artifact; a remand sends the gap back for one more exchange', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('judge', 'G-00000001', 'G-00000002')], { docket: ['G-00000001'] }) }), passChair()] }))
  await world.run(script, ARGS)
  const bench = firstPrompt(world, 'judge')
  for (const want of ['Docketed for you: G-00000001, G-00000002', 'THE DOCKET IS A ROUTING LIST, NOT THE EVIDENCE', 'read them FRESH before ruling', 'AS IT NOW STANDS', 'Re-run each document-probe acceptance check',
    'RULING BASIS IS CONFINED TO', 'counts AGAINST the side that made it', "READ THE NAMED ANCESTORS' RECORDS", 'the docket ruling', 'barring as settled', 'a gap you REMAND was never closed, so a remand is never final — it goes back to the debate for ONE more exchange', 'the research direction you state, which the record refuses blank',
    "where none pulled, the ruling's weakest point", 'where nothing needs a human, what on the record already settles it', 'for a ruling that closes the gap, what would reopen it', 'a second remand at impasse is refused, and its refusal names the dispositions that close the gap', 'NAMED infrastructure debt',
    'rules nothing is a workflow error', 'PRECEDENT IS ARGUMENT, NOT EVIDENCE']) {
    assert.ok(bench.includes(want), `the bench lost: ${want}`)
  }
  const schema = labelsOf(world, 'judge')[0].opts.schema
  assert.ok(!schema.properties.deadlock && !schema.required.includes('deadlock'), 'deadlock is not a flag the bench returns')
  assert.ok(schema.properties.dispositions.items.properties.disposition.enum.includes('defect_owed_elsewhere'))
})

test('W1.9: defect_owed_elsewhere ships as a named infra debt with the epoch it was ruled in', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }) }), passChair()],
    judge: [judgeEnv({ dispositions: [{ gap_id: 'G-00000001', disposition: 'defect_owed_elsewhere', rationale: 'setup tooling must stage it' }] })],
  }))
  const out = await world.run(script, ARGS)
  assert.deepEqual(out.infra_debts, [{ gap_id: 'G-00000001', owed_fix: 'setup tooling must stage it', epoch: 1 }])
  assert.ok(firstPrompt(world, 'judge · assemble').includes('setup tooling must stage it'), 'assembly is handed the named debts')
})

test('no seat prompt names a command path or spells a flag — the help page is the only page', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G-00000001')], { docket: ['G-00000001'] }) }), chairEnv({ plan: plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }) }), passChair()],
  }))
  await world.run(script, ARGS)
  for (const c of world.calls) {
    const body = c.prompt.replace(/--help\b|--seat-id\b|--run\b|--reason(-file)?\b|--comments\b/g, '')
    assert.ok(!/--[a-z][a-z-]+\b/.test(body), `${c.opts.label} spells a flag: ${(body.match(/.{40}--[a-z][a-z-]+.{20}/) || [''])[0]}`)
    assert.ok(!/\b(show (board|work|motions|closings|changes)|dispatch next|motion (grade|inquiry|docket|petition) (file|rule|appeal))\b/.test(body),
      `${c.opts.label} names a command path: ${(body.match(/.{40}(show \w+|dispatch next|motion \w+ \w+).{20}/) || [''])[0]}`)
  }
})

test('the bench\'s rulings travel to both parties, with blue told its duty and red estopped, each relying on the ruling as stated', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G-00000001')]) }), chairEnv({ plan: plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }) }),
      chairEnv({ plan: plan([party('red-lens-evidence', 'G-00000001'), party('blue-respond', 'G-00000001')]) }), passChair()],
    judge: [judgeEnv({ dispositions: [{ gap_id: 'G-00000001', disposition: 'not_a_defect', rationale: 'THE OPINION', settled: 'THE BARRED PROPOSITION', reopens_on: 'A NEW SOURCE', final: false }] })],
  }))
  await world.run(script, ARGS)
  const [blue1, blue2] = labelsOf(world, 'blue-respond').map((c) => c.prompt)
  assert.ok(!/GAPS THE BENCH HAS RULED/.test(blue1), 'nothing ruled yet, nothing claimed')
  for (const want of ['GAPS THE BENCH HAS RULED', 'THE BENCH FOUND NO DEFECT', 'rely on this ruling as established', 'THE BARRED PROPOSITION', 'A NEW SOURCE', 'THE BAR IS ON THE PROPOSITION, NOT THE GAP']) {
    assert.ok(blue2.includes(want), `blue lost: ${want}`)
  }
  const lens = firstPrompt(world, 'red-lens-evidence')
  assert.ok(/YOU ARE ESTOPPED/.test(lens) && lens.includes('THE BARRED PROPOSITION') && lens.includes('A NEW SOURCE'), 'red is handed the bar and what reopens it')
  // A RULING IS RELIED ON AS ITS ROW STATES IT (gblock's ruling, 2026-10-09): the bench's opinion is on the
  // bench's and the chair's read group, and this clause renders to a lens and to blue, who hold neither.
  for (const [who, prompt] of [['blue', blue2], ['a lens', lens]]) {
    assert.ok(prompt.includes('RELY ON EACH RULING AS ITS ROW STATES IT'), `${who} is not told to rely on the ruling as stated`)
    assert.ok(!/read the bench's opinion|THE REASONING IS ON THE RECORD/.test(prompt), `${who} is sent to the bench's opinion`)
  }
  assert.ok(!blue2.includes('THE OPINION') && !lens.includes('THE OPINION'), 'the opinion is not in the prompt')
  assert.ok(!/your_duty/.test(lens), 'the duty table is blue\'s')
})

test('an empty settled bars nothing: the ruling still travels, and red is handed no estoppel over a blank', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G-00000001')]) }), chairEnv({ plan: plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }) }),
      chairEnv({ plan: plan([party('red-lens-evidence', 'G-00000001'), party('blue-respond', 'G-00000001')], { remand_owed: [{ gap_id: 'G-00000001', direction: 'THE RECORD DIRECTION' }] }) }),
      chairEnv({ plan: plan([party('red-lens-evidence', 'G-00000001'), party('blue-respond', 'G-00000001')]) }), passChair()],
    judge: [judgeEnv({ dispositions: [{ gap_id: 'G-00000001', disposition: 'remanded', rationale: 'THE OPINION', settled: '' }] })],
  }))
  await world.run(script, ARGS)
  const lens = firstPrompt(world, 'red-lens-evidence')
  assert.ok(lens.includes('GAPS THE BENCH HAS RULED') && lens.includes('"gap_id":"G-00000001"') && lens.includes('THE RECORD DIRECTION'), 'the ruling and the direction its exchange owes still reach red')
  assert.ok(!/"settled"/.test(lens), 'an empty settled renders as a proposition')
  assert.ok(!/YOU ARE ESTOPPED/.test(lens), 'red is estopped on a proposition nobody stated')
  // #1210: a remand readies blue for ONE more exchange, and the direction it owes travels with it.
  const blue2 = labelsOf(world, 'blue-respond').map((c) => c.prompt)[1]
  assert.ok(blue2.includes('back in the debate for ONE more exchange') && blue2.includes('THE RECORD DIRECTION'), 'blue is told the remand readies it for one more exchange, and the direction it owes')
  // The duty lasts while the plan readies the exchange: a later plan that does not name G-00000001 in
  // remand_owed tells blue the exchange is not owed, never that it is dispatched on it for one.
  const blue3 = labelsOf(world, 'blue-respond').map((c) => c.prompt)[2]
  assert.ok(blue3.includes('is not owed now') && !blue3.includes('back in the debate for ONE more exchange'), 'blue is still told a spent remand readies it for one more exchange')
  assert.ok(!blue3.includes('THE RECORD DIRECTION'), 'a spent remand still hands blue its direction as a duty')
})

// #1249 review: A REMAND'S DIRECTION IS THE RECORD'S. The bench's envelope is its own account of
// the ruling it recorded; the direction blue and the minting lens owe comes from the plan, which
// reads the record. An envelope remand with a different reopens_on, or none, changes nothing.
test('a remand\'s direction reaches blue and the lens from the plan, never from the bench envelope', async () => {
  for (const account of [{}, { reopens_on: 'THE ENVELOPE ACCOUNT' }, { reopens_on: 'THE ENVELOPE ACCOUNT', final: true }]) {
    const world = makeWorld(makeResponder({
      chair: [chairEnv({ plan: plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }) }),
        chairEnv({ plan: plan([party('red-lens-evidence', 'G-00000001'), party('blue-respond', 'G-00000001')], { remand_owed: [{ gap_id: 'G-00000001', direction: 'THE RECORD DIRECTION' }] }) }), passChair()],
      judge: [judgeEnv({ dispositions: [{ gap_id: 'G-00000001', disposition: 'remanded', rationale: 'THE OPINION', settled: '', ...account }] })],
    }))
    await world.run(script, ARGS)
    for (const [who, prompt] of [['blue', firstPrompt(world, 'blue-respond')], ['the lens', firstPrompt(world, 'red-lens-evidence')]]) {
      assert.ok(prompt.includes('"direction":"THE RECORD DIRECTION"'), `${who} is not handed the record's direction (envelope ${JSON.stringify(account)})`)
      assert.ok(!prompt.includes('THE ENVELOPE ACCOUNT'), `${who} is handed the envelope's account of the direction`)
      assert.ok(!/"final"|"reopens_on"/.test(prompt.slice(prompt.indexOf('GAPS THE BENCH HAS RULED'))), `${who} is handed a remand as a reopen condition or a final ruling`)
      assert.ok(prompt.includes('A REMANDED gap was never closed') && prompt.includes('what this exchange owes, not a condition for reopening'), `${who} reads a remand's direction as a reopen condition`)
    }
  }
})

test('the assembler authors nothing, stamps the outcome the record derives, and is told an open gap that is not material stays on the board', async () => {
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  await world.run(script, ARGS)
  const asm = firstPrompt(world, 'judge · assemble')
  assert.ok(/you author NOTHING, you copy NOTHING/.test(asm) && /do not copy anything into it yourself/.test(asm) && /cannot mis-author a synthesis surface/.test(asm))
  assert.ok(/it is VERIFIED/.test(asm) && /Record it as the run's outcome — VERIFIED/.test(asm) && !/ended at the ceiling/.test(asm))
  // THE REASON IS A JUDGEMENT, NOT A RECAP. `--reason` carries why this outcome is the right read
  // of the board; the ledger already holds what the sitting did, in order.
  assert.ok(/WHY this outcome is the right read of the board/.test(asm) && !/your account of the sitting/.test(asm),
    'the assemble prompt asks for a recap of the sitting rather than the judgement')
  assert.ok(/An open gap that is not material stays open on the board and in the risk matrix; the chair's PASS listed it by class, on the record\./.test(asm))
  assert.ok(!asm.includes('open_questions'), 'assembly reads blue\'s audited section from the record, never receives it')
})

// ── bookends and prompt clauses carried forward ─────────────────────────────────────────────

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
  for (const [i, p] of lanes.entries()) assert.ok(p.includes('YOUR SOURCES, IN PROSE') && p.includes('Do NOT mint footnote labels'), `lane ${i + 1} source-note convention`)
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
  assert.ok(/every section names the reader's question it answers/.test(synth), "the synthesis is written to the reader's question")
  // THE CONTENT RULE IS THE VERBS', and the prompt points at it rather than restating it: the rule is
  // on every verb whose text the report prints (edit, ingest, the inquiry's propose and record, and
  // compute's note and cite's title, which the Bibliography prints),
  // so a copy here is a second statement of a contract nothing keeps in step.
  // The pointer is the synthesizer constitution's, once: it holds at every synthesis.
  const synthMd = readFileSync(new URL('../../agents/blue-synthesizer.md', import.meta.url), 'utf8')
  assert.ok(/The full content rule .* is on the verbs that write report text/.test(synthMd) && !/CONTENT RULE/i.test(synth),
    'the synthesizer constitution, and not its prompt, points at where the report content rule lives')
  assert.ok(!/RESEARCH PROSE AND THE TOOL'S OWN MARKERS/.test(synth) && !/EXTRACTED from the record/.test(synth),
    'the synthesis prompt restates the verbs\' report content rule instead of pointing at it')
  assert.ok(synth.includes('## Open questions') && synth.includes('THE CATECHISM IS YOURS') && synth.includes('## The Catechism') && synth.includes('every risk-accepted residual'))
  assert.ok(synth.includes('## The board') && synth.includes('is FABRICATION') && synth.includes('ONLY WHAT YOU CAN AUTHOR'))
  // THE REPORT'S SECTIONS ARE ITS AUTHOR'S TO NAME. The prompt listed "technical foundations,
  // analysis" among what blue authors, the same prompt asks that each section name the reader's
  // question, and on universe-m21 the report had neither heading: it shipped a "not authored"
  // line under each and a gap for lacking them.
  assert.ok(/the sections of your argument under headings you choose/.test(synth) && !/technical foundations|Additional findings/i.test(synth),
    'the synthesis prompt names a heading the report owes, or one the tool writes above its sections')
  assert.ok(/reorganize freely/i.test(synth) && /retired on the record/.test(synth))
  // THE SYNTHESIZER RECORDS NO INQUIRY, AND ITS PROMPT ASKS FOR NONE. It merges the drafts; the
  // approaches are the frontier's and the lanes', who did the work. What it is told is what the
  // report holds of them: the tool composes the three sections, and the report carries no list of
  // approaches under any heading.
  assert.ok(!/INQUIRIES:/.test(synth) && !/the hypothesis that would be true if it paid off/.test(synth) && !/NO FATE IS OWED/.test(synth) && !/The question you propose/.test(synth),
    'the synthesis prompt carries the inquiry paragraph: the synthesizer is asked to record approaches')
  assert.ok(/The approaches the research considered are among them: each is an inquiry on the record, where what became of it is said, and the report carries no list of them\./.test(synth),
    'the synthesis prompt must say the report carries no list of the approaches')
  for (const h of ['Risk matrix', 'Research areas', 'Future research directions', 'Alternatives considered', 'Bibliography']) {
    assert.ok(synth.includes('Do NOT write any of') && synth.includes('`## ' + h + '`'), 'the synthesis prompt does not name ## ' + h + ' among the sections assembly composes')
  }
  assert.ok(!/reach the reader under their own headings|left for a later run/.test(synth), 'the synthesis prompt cues blue to head a section of its own with the inquiries')
  // THE FRONTIER WRITES THE QUESTIONS EVERY LATER SEAT INHERITS, before any research: the
  // prediction goes in the hypothesis, never in the question.
  const frontier = firstPrompt(world, 'frontier')
  assert.ok(/ITS QUESTION IS WHAT the inquiry asks .* with no answer in it/.test(frontier) && /ITS HYPOTHESIS IS YOUR PREDICTION/.test(frontier) && /A result you expect goes in the hypothesis, never in the question\./.test(frontier),
    'the frontier prompt must say the question holds no answer and the prediction is the hypothesis')
  assert.ok(!/it takes the approach you would follow and the hypothesis/.test(frontier), 'the frontier prompt runs the question and the hypothesis together')
  // AN APPROACH IS RECORDED WITH THE FATE IT HAD, AND NO FATE IS OWED AN ENTRY: the lanes' paragraph.
  // A prompt that asks for "what you considered and declined" is read as asking for a declined
  // inquiry. "Your part is the record: you put each approach on it" presses the same way.
  for (const lane of labelsOf(world, 'blue-lane').map((c) => c.prompt)) {
    assert.ok(/INQUIRIES:/.test(lane) && /NO FATE IS OWED AN ENTRY/.test(lane) && /A sitting that weighed nothing and set it down has no declined inquiry, and that record is complete/.test(lane),
      'the inquiry clause must say an approach carries the fate it had and that having no declined inquiry is a complete record')
    // A REFUTATION HAS A FATE OF ITS OWN. The clause tells the lane to record the fate the approach
    // really had and sends it to the page of the act that records it for what each fate means.
    assert.ok(/Record the fate the approach really had; what each fate means is on the page of the act that records it/.test(lane),
      'the inquiry clause must tell the lane to record the fate the approach had and where each fate is defined')
    assert.ok(!/one you ran that died is abandoned/.test(lane), 'the inquiry clause calls an approach that ran abandoned')
    // THE QUESTION HOLDS NO PREDICTED RESULT. A question that carries its predicted result is printed
    // as the question beside whatever the research finds, and no act rewrites it.
    assert.ok(/THE QUESTION HOLDS NO PREDICTED RESULT: what you expect to find is the hypothesis/.test(lane),
      'the inquiry clause must keep the predicted result out of the question')
    assert.ok(!/not only the one you took/.test(lane) && !/considered and declined is what makes/.test(lane) && !/dead ends matter most/.test(lane) && !/you put each approach on it/.test(lane),
      'the inquiry clause presses for an entry')
    // THE INQUIRY SECTIONS HAVE AN ACTOR, AND IT IS THE TOOL: who composes the three sections, which
    // fate lands under which, and that blue writes none of them.
    assert.ok(/THE TOOL WRITES THE REPORT'S ACCOUNT OF THEM, NOT YOU/.test(lane) && /You write none of those three headings in the report/.test(lane))
    assert.ok(/Research areas holds each inquiry that is proposed, underway, confirmed or refuted, Future research directions each one deferred, and Alternatives considered each one declined or abandoned/.test(lane))
  }
  // THE CARD IS RETIRED, VERB AND CLAUSE (#1126). `show work` and the scorecard were the same
  // intent from two designs; work is the one that kept it. Measured 2026-09-22: 36 reads and 36 of
  // 39 results carried the literal "not computed", which the verb's own help concedes — the rows
  // fill in at capture, not in the sitting reading them. A prompt must not send a seat to it.
  assert.ok(!/SCORECARD/i.test(synth) && !synth.includes('scorecards.mjs') && !/--bin\b/.test(synth))
})

test('priors-are-poison: no cross-run scorecard seed reaches any chair, even when scorecards are supplied', async () => {
  const scorecards = { 'blue-synthesize': { repair_regression_ratio: 0.63 }, 'red-chair': { anchored_closures_pct: 89 }, assemble: { remanded_share: 0.98 } }
  const world = makeWorld(makeResponder({ chair: [passChair()] }))
  await world.run(script, { ...ARGS, scorecards })
  assert.ok(!firstPrompt(world, 'blue-synthesize').includes('0.63') && !firstPrompt(world, 'red-chair').includes('89') && !firstPrompt(world, 'judge · assemble').includes('0.98'))
  // STRONGER THAN IT WAS: the seat-side card is retired, so no prompt mentions a scorecard at all
  // and a supplied seed has nowhere to land. The arg is still supplied here on purpose.
  assert.ok(!world.calls.some((c) => /SCORECARD/i.test(c.prompt)), 'no prompt names a scorecard')
})

test('every seat prompt carries the speed clause and the record contract, and leaves the log duty to the constitution; bench sittings carry the law clause', async () => {
  const world = makeWorld(makeResponder({
    // A stub plan: the record never permits a PASS over a bench motion, but the terminal sitting's
    // prompt is what this reads, and the bench-owned blocker is what convenes it.
    chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G-00000001'), party('judge', 'G-00000001')], { docket: ['G-00000001'] }) }), passChair({ plan: passPlan({ blockers: [blocker('M-00000001', 'judge')] }) })],
  }))
  await world.run(script, ARGS)
  for (const seat of ['blue-synthesize', 'red-chair', 'red-lens-evidence', 'blue-respond', 'judge #', 'judge · terminal', 'judge · assemble']) {
    const c = labelsOf(world, seat)[0]
    assert.ok(c, `${seat} sat`)
    assert.ok(!/envelope's log/.test(c.prompt) && !/says so in the envelope/.test(c.prompt), `${seat} is told to log in its envelope — the record is the one channel`)
    // THE LOG DUTY IS THE CONSTITUTION'S, ONCE. It is unconditional, so the seat's definition states
    // it (the log fragments under scripts/agentgen/src/fragments) and the prompt restates none of it.
    assert.ok(!/LOG \([^)]*\)/.test(c.prompt) && !/cost you a call, a guess or an act/.test(c.prompt) && !/RETOOL YOU/i.test(c.prompt), `${seat}: the prompt restates the log duty its constitution states`)
    const definition = readFileSync(new URL(`../../agents/${c.opts.agentType.split(':')[1]}.md`, import.meta.url), 'utf8')
    assert.ok(definition.includes('THE LOG IS FOR WHATEVER COST YOU A CALL, A GUESS OR AN ACT') && /what you expected, and where the expectation came from/.test(definition), `${seat} lost the operator channel: its constitution does not state the log duty`)
    // THE RETIRED SHAPES, in both directions. The first four told a seat the log was owed whatever
    // happened, and seats padded it; the last two narrowed it to what BLOCKED a seat, and on
    // universe-m13 every seat classed its refusals and guesses as its own mistakes and filed nothing.
    assert.ok(!/OWES THE SURVEY/.test(c.prompt) && !/SILENCE IS NOT THE EMPTY CASE/.test(c.prompt) && !/CLOSE THIS CHANNEL/.test(c.prompt) && !/JUDGEMENT rather than for want of occasion/.test(c.prompt) && !/ONLY WHAT GOT IN YOUR WAY/.test(c.prompt) && !/FILE NOTHING/.test(c.prompt), `${seat} restates a retired rule`)
    assert.ok(c.prompt.includes(speedClauseText), `${seat} lost the speed clause`)
    assert.ok(c.prompt.includes('SEAT_ID:') && c.prompt.includes('/opt/feov/bin/feov-record'), `${seat} lost the record contract`)
  }
  for (const seat of ['judge #', 'judge · terminal', 'judge · assemble']) {
    const p = labelsOf(world, seat)[0].prompt
    assert.ok(p.includes('PRECEDENT IS ARGUMENT, NOT EVIDENCE') && p.includes('the leaf wins') && p.includes('only AFFIRMED ones bind'), `${seat} lost the law clause`)
    assert.ok(!/DECLARE:/.test(p), `${seat} re-teaches declare beside its help page`)
  }
  const lens = firstPrompt(world, 'red-lens-evidence')
  // THE SURFACE ARRIVES WITH THE SEAT, AND SO DOES THE DUTY TO READ IT: the constitution carries
  // every command's own --help and the one statement of the read-it-whole duty (the
  // surface-discovery fragment). The prompt adds what only the sitting knows — where one page prints.
  assert.ok(lens.includes('--seat-id red-lens-evidence <command> --help'), 'the lens lost the pointer to one page of its surface')
  assert.ok(!/IS A GUESS/.test(lens) && !/READ IT BEFORE YOU CHOOSE/.test(lens) && !/YOUR CONFIGURATION HOLDS YOUR WHOLE SURFACE/.test(lens), 'the lens prompt restates the surface-reading duty its constitution states')
  const audit = readFileSync(new URL('../../skills/adversarial-audit/SKILL.md', import.meta.url), 'utf8')
  assert.ok(/Before you act, read your WHOLE SURFACE/.test(audit) && /is a guess/.test(audit), 'the red seats lost the surface-reading directive: the skill they preload does not state it')
  // AND IT DOES NOT SEND THE SEAT AFTER A DOCUMENT IT WAS GIVEN. Both retired shapes: the
  // page-by-page walk `manual` replaced, and the `manual` fetch the constitution replaced.
  assert.ok(!/for EVERY group that page listed/.test(lens) && !lens.includes('<group> --help'), 'the lens still carries the page-by-page walk the manual replaced')
  assert.ok(!/manual\.txt/.test(lens) && !/ manual —/.test(lens) && !/IMMEDIATELY AFTER `register`/.test(lens), 'the lens is still told to fetch a manual its constitution already carries')
  // THE SPEED CLAUSE NAMES NO SEARCH TOOL AND FORBIDS NO SEARCH. The tool list is the harness's: a
  // session with no Grep or Glob answers "No such tool available" to a seat told to use them, and a
  // seat forbidden grep through the shell is then left with no search at all.
  assert.ok(/KNOWN HARNESS LIMIT/.test(lens) && /search through the shell/.test(lens), 'the speed clause lost the search a seat always has: the shell')
  assert.ok(!/\b(Grep|Glob)\b/.test(lens) && !/cat\/grep through/.test(lens), 'the speed clause names a search tool the session may not have, or forbids grep through the shell')
})

test('the record contract binds each seat to the id it hands the tool, petition sittings included', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([petitionBench(), party('blue-respond', 'G-00000001')], { blockers: [blocker('M-00000001', 'judge')] }) }), passChair()],
  }))
  await world.run(script, ARGS)
  // A PETITION SITTING IS THE BENCH. The filer is on the petition the sitting rules, on the record,
  // never in the seat id or the label.
  const petitions = labelsOf(world, 'judge · petition')
  assert.deepEqual(petitions.map((c) => c.opts.label.replace(/ · [^·]+$/, '')), ['judge · petition'],
    'one petition sitting, the bench')
  for (const c of world.calls) {
    const seat = c.opts.label.split(' ')[0]
    if (/^(frontier|blue-lane)/.test(seat)) continue
    assert.ok(c.prompt.includes(`SEAT_ID: ${seat}`), `${c.opts.label} declares a SEAT_ID other than its own`)
  }
})

test('W2c: a plan that convenes the bench for petitions seats it before any party; denied continues; a halt ends the run HALTED', async () => {
  const convened = () => chairEnv({ plan: plan([party('red-lens-evidence'), petitionBench(), party('blue-respond', 'G-00000001')], { blockers: [blocker('M-00000001', 'judge')] }) })
  const denied = makeWorld(makeResponder({ chair: [convened(), passChair()] }))
  const out = await denied.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  const sitting = denied.calls.findIndex((c) => c.opts.label.startsWith('judge · petition'))
  const lens = denied.calls.findIndex((c) => c.opts.label.startsWith('red-lens-evidence'))
  const blue = denied.calls.findIndex((c) => c.opts.label.startsWith('blue-respond'))
  assert.ok(sitting >= 0 && sitting < lens && sitting < blue, 'the sitting fired before any party sat')
  assert.equal(out.petitions.length, 1)
  assert.equal(out.petitions[0].petitioner, 'x', 'the petitioner is the ruling\'s, read off the record by the bench')
  assert.ok(denied.calls[sitting].prompt.includes('never sanctioned') && /A HALT IS A DIFFERENT DECISION FROM A RULING/.test(denied.calls[sitting].prompt))
  const halting = makeWorld(makeResponder({
    chair: [convened()],
    petition: [petitionRulingEnv({ halt: { opinion: 'the human must decide this' } })],
  }))
  const out2 = await halting.run(script, ARGS)
  assert.equal(out2.verdict, 'HALTED')
  assert.equal(out2.halted, true)
  assert.ok(out2.halt_opinion.includes('the human must decide'))
  assert.ok(!halting.calls.some((c) => c.opts.label.startsWith('blue-respond')), 'nothing sat past the halt')
  const asm = firstPrompt(halting, 'judge · assemble')
  assert.ok(asm.includes('it is HALTED') && asm.includes('the human must decide'))
})

test('W2c: no plan convenes it, no sitting — whatever an envelope carries; granted relief binds the party it names', async () => {
  const quiet = makeWorld(makeResponder({ chair: [passChair()] }))
  await quiet.run(script, ARGS)
  assert.ok(!quiet.calls.some((c) => c.opts.label.startsWith('judge · petition')))
  // AN ENVELOPE IS NOT A PETITION CHANNEL: the record is, relayed in the plan. Every seat here
  // returns a `petitions` field and nothing convenes the bench.
  const petitions = [{ class: 'ethical', basis: 'x', relief: 'narrow' }]
  const envelopes = makeWorld(makeResponder({
    chair: [chairEnv({ petitions }), passChair({ petitions })],
    blueSynth: [blueEnv({ petitions })], blueRespond: [blueEnv({ petitions })],
  }))
  await envelopes.run(script, ARGS)
  assert.ok(!envelopes.calls.some((c) => c.opts.label.startsWith('judge · petition')), 'an envelope field convened the bench')
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([petitionBench(), party('blue-respond', 'G-00000001')], { blockers: [blocker('M-00000001', 'judge')] }) }), passChair()],
    petition: [petitionRulingEnv({ rulings: [{ petitioner: 'red-chair', class: 'procedural', ruling: 'granted', relief: 'scope narrowed to §3', binds: 'blue' }] })],
  }))
  await world.run(script, ARGS)
  assert.ok(firstPrompt(world, 'blue-respond').includes('BENCH RELIEF IN EFFECT') && firstPrompt(world, 'blue-respond').includes('scope narrowed'))
  // THE RIGHT IS STATED AT THE SEAT, ONCE: in its constitution, and in no prompt.
  assert.ok(!world.calls.some((c) => /PETITIONS:|you may petition the bench/.test(c.prompt)), 'a prompt restates the petition right')
  for (const held of ['../../agents/blue-synthesizer.md', '../../agents/blue-researcher.md', '../../skills/adversarial-audit/SKILL.md']) {
    assert.ok(readFileSync(new URL(held, import.meta.url), 'utf8').includes('**PETITION RIGHT**'), `${held}: the right is stated at the seat`)
  }
})

test('W2j: a bench holding reaches a seat from the record alone — every prompt states the duty, and none carries a holding', async () => {
  // A BENCH THAT RETURNS A HOLDING ON ITS ENVELOPE ANYWAY binds nobody through it: no schema
  // declares the field, and a holding is a declaration on the record, which every seat's work list
  // carries with the board's status beside each gap it names.
  const stray = 'G-00000002 remains open and is material'
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }) }), chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G-00000002')]) }), passChair()],
    judge: [judgeEnv({ holdings: [stray] })],
  }))
  await world.run(script, ARGS)
  assert.ok(world.calls.length > 0, 'nothing was dispatched')
  for (const c of world.calls) {
    assert.ok(!c.prompt.includes(stray), `${c.opts.label} carries a holding the envelope returned`)
    assert.ok(c.prompt.includes("THIS RUN'S BENCH HOLDINGS BIND HOW YOU READ THE RECORD; each is on your work list"), `${c.opts.label} is not told what a recorded holding obliges`)
    assert.ok(c.prompt.includes('the board, not a holding, says where a gap stands'), `${c.opts.label} is not told which wins`)
  }
  for (const seat of ['red-chair #1', 'red-chair #2', 'red-lens-evidence', 'blue-respond', 'judge · assemble']) {
    assert.ok(labelsOf(world, seat).length > 0, `${seat} never sat`)
  }
})

test('integrity inspection arms only with transcriptDir, on the bench, and binds integrity-not-merits', async () => {
  const chair = [chairEnv({ plan: plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }) }), passChair()]
  const armed = makeWorld(makeResponder({ chair }))
  await armed.run(script, { ...ARGS, transcriptDir: '/sess/wf-123' })
  const bench = firstPrompt(armed, 'judge')
  assert.ok(/INTEGRITY INSPECTION/.test(bench) && /\/sess\/wf-123\/agent-\*\.jsonl/.test(bench) && /MUST NOT use trajectory material to decide the MERITS/.test(bench) && /DECLARE every inspection/.test(bench))
  const unarmed = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }) }), passChair()] }))
  await unarmed.run(script, ARGS)
  assert.ok(!unarmed.calls.some((c) => /INTEGRITY INSPECTION/.test(c.prompt)))
})

// THE ENGINE RE-PROMPTS NO SEAT FOR ITS SITTING RECORD, AND NO ENVELOPE RELAYS A FACT THE RECORD
// HOLDS. A claim count is counted from the report by whoever wants it; what a sitting put on the
// record is read off the record — blue-respond's own work list blocks on a position it owes and
// capture's record-parity names a sitting that closed without one. So an envelope that carries
// neither key is the whole envelope, one that carries them anyway moves nothing, and nothing the
// engine returns or hands to assembly is a list of unattested sittings.
test('the engine re-prompts no seat for its sitting record and reads no relayed claim count', async () => {
  const src = readFileSync(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url), 'utf8')
  for (const gone of ['ensureSittingRecord', 'sitting_record_appended', 'sitting_record_unresolved', 'claim_count', 'blue_claims']) {
    assert.ok(!src.includes(gone), `debate.js names ${gone}: the engine reads the record for nothing and relays nothing it could compute`)
  }
  const chair = () => [chairEnv({ plan: plan([party('blue-respond', 'G-00000001')]) }), passChair()]
  const bare = makeWorld(makeResponder({ chair: chair() }))
  const out = await bare.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  for (const c of bare.calls) {
    if (!/^blue-(synthesize|respond) /.test(c.opts.label)) continue
    assert.ok(c.opts.schema, `${c.opts.label} is dispatched with no schema`)
    assert.deepEqual(c.opts.schema.required ?? [], [], `${c.opts.label}: a blue envelope requires nothing — what blue did is on the record`)
    for (const key of ['claim_count', 'sitting_record_appended']) assert.ok(!(key in c.opts.schema.properties), `${c.opts.label}: the schema declares ${key}`)
    // The synthesizer's envelope declares nothing, and is closed: its whole envelope is the empty one.
    assert.ok(schemaAccepts(c.opts.schema, /^blue-synthesize /.test(c.opts.label) ? {} : blueEnv()), `${c.opts.label}: an envelope carrying neither key is refused`)
    assert.ok(!schemaAccepts(c.opts.schema, { ...blueEnv(), claim_count: 7 }), `${c.opts.label}: a key the schema does not declare is accepted`)
  }
  // The old attestation, returned false by both seats, dispatches nobody and is named nowhere.
  const stale = makeWorld(makeResponder({ chair: chair(), blueSynth: [blueEnv({ sitting_record_appended: false, claim_count: 7 })], blueRespond: [blueEnv({ sitting_record_appended: false, claim_count: 9 })] }))
  const out2 = await stale.run(script, ARGS)
  assert.deepEqual(stale.calls.map((c) => c.opts.label.replace(/ · .*$/, '')), bare.calls.map((c) => c.opts.label.replace(/ · .*$/, '')), 'an unattested envelope changed who was dispatched')
  assert.ok(!stale.calls.some((c) => /sitting-record/.test(c.opts.label)), 'a seat was re-prompted for its sitting record')
  for (const key of ['sitting_record_unresolved', 'blue_claims']) assert.ok(!(key in out2), `the workflow returns ${key}`)
  assert.ok(!/sitting[_ -]record|re-prompt|did not attest/i.test(firstPrompt(stale, 'judge · assemble')), 'the assembly dispatch carries an unresolved-sitting list')
  assert.ok(!stale.logs.some((l) => /claims|attest/.test(l)), stale.logs.join(' | '))
  // blue's "found every engaged gap closed" path needs no attestation: it is the envelope's own list.
  const closed = makeWorld(makeResponder({ chair: chair(), blueRespond: [blueEnv({ manifest: [], found_closed: ['G-00000001'] })] }))
  assert.equal((await closed.run(script, ARGS)).verdict, 'VERIFIED')
  assert.ok(closed.logs.some((l) => l.includes('blue found every engaged gap (G-00000001) closed before it sat — nothing to answer')), closed.logs.join(' | '))
})

// W2b, OWED GAPS ONLY, NEVER AN ABORT (gblock's ruling on #868). The B4 ordering: the plan engaged
// the lenses and blue on G-00000001 and G-00000002, the lenses sat first and closed both, and blue correctly found
// nothing to repair. The engine checked blue against the plan and threw, killing the run.
const engagedOn = (...gaps) => [chairEnv({ plan: plan([party('red-lens-logic', ...gaps), party('blue-respond', ...gaps)]) }), passChair()]
test('W2b: the B4 ordering — lenses close G-00000001 and G-00000002 first, blue files no row — continues and logs nothing unmanifested', async () => {
  const world = makeWorld(makeResponder({ chair: engagedOn('G-00000001', 'G-00000002'), blueRespond: [blueEnv({ manifest: [], found_closed: ['G-00000001', 'G-00000002'] })] }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  assert.ok(!world.logs.some((l) => /no row for/.test(l)), `a gap closed before blue sat is not owed: ${world.logs.join(' | ')}`)
  assert.ok(world.logs.some((l) => l.includes('blue found G-00000001, G-00000002 closed before it sat')))
})
test('W2b: a gap still open with no row is logged and left to capture; none of N is no longer fatal', async () => {
  const oneOpen = makeWorld(makeResponder({ chair: engagedOn('G-00000001', 'G-00000002'), blueRespond: [blueEnv({ manifest: [], found_closed: ['G-00000001'] })] }))
  assert.equal((await oneOpen.run(script, ARGS)).verdict, 'VERIFIED')
  assert.ok(oneOpen.logs.some((l) => l.includes('0/1 gap(s) open when blue sat — no row for: G-00000002') && l.includes('scored at capture')), oneOpen.logs.join(' | '))
  const noneOfTwo = makeWorld(makeResponder({ chair: engagedOn('G-00000001', 'G-00000002'), blueRespond: [blueEnv({ manifest: [] })] }))
  assert.equal((await noneOfTwo.run(script, ARGS)).verdict, 'VERIFIED')
  assert.ok(noneOfTwo.logs.some((l) => l.includes('0/2 gap(s) open when blue sat — no row for: G-00000001, G-00000002')))
})
// A MANIFEST ROW IS A GAP ID, AND PROSE IS REFUSED AT THE RETURN. Respond seats wrote rows such as
// "G-237fed96: repaired (…)" into `manifest` (m29, and 4 of 32 replays); the coverage log matches
// ids, so each read as "no row for" a gap that had one. GAP_ID is generated from anchor's matcher.
test('a respond envelope takes gap ids in manifest and found_closed, and refuses prose there', async () => {
  const w = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('blue-respond', 'G-00000001')]) }), passChair()] }))
  await w.run(script, ARGS)
  const respond = w.calls.find((c) => /^blue-respond /.test(c.opts.label))
  assert.ok(respond, 'blue-respond was never dispatched')
  const schema = respond.opts.schema
  assert.ok(schemaAccepts(schema, { manifest: ['G-00000001'], found_closed: ['G-0a1b2c3d'] }), 'gap ids are refused')
  assert.ok(!schemaAccepts(schema, { manifest: ['G-00000001: repaired (the sum was wrong)'] }), 'a prose manifest row is accepted')
  assert.ok(!schemaAccepts(schema, { found_closed: ['the evidence gap'] }), 'a prose found_closed entry is accepted')
})

// R2 (gblock, 2026-10-23): BLUE SITS AGAIN ON THE SUCCESSOR OF A GAP IT FOUND CLOSED. universe-m31 epoch
// 10: the chair engaged logic and blue on G-77cc033b, logic closed it and minted G-1ac85c33 superseding
// it, and blue sat on a closed gap and recorded nothing — the successor waited for an epoch that never
// came. Blue engages its next sitting on the successor and relays it; the engine sits it once more, and
// no more: a second sitting's successors, and a successor already engaged this epoch, are not sat.
test('R2: blue sits a second time on the successor it engaged, once per successor, and never a third time', async () => {
  const blueRespond = [
    blueEnv({ manifest: [], found_closed: ['G-77cc033b'], successors: ['G-1ac85c33'] }),
    blueEnv({ manifest: ['G-1ac85c33'], found_closed: ['G-1ac85c33'], successors: ['G-00000003'] }),
  ]
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('red-lens-logic', 'G-77cc033b'), party('blue-respond', 'G-77cc033b')]) }), passChair()], blueRespond }))
  assert.equal((await world.run(script, ARGS)).verdict, 'VERIFIED')
  const blues = world.calls.filter((c) => /^blue-respond /.test(c.opts.label))
  assert.equal(blues.length, 2, `blue sat ${blues.length} time(s) in the epoch: ${world.logs.join(' | ')}`)
  assert.ok(blues[0].prompt.includes('You are engaged on: G-77cc033b.'), blues[0].prompt.slice(0, 200))
  assert.ok(blues[1].prompt.includes('You are engaged on: G-1ac85c33.'), blues[1].prompt.slice(0, 200))
  assert.ok(world.logs.some((l) => l.includes('engaged its next sitting on the successor(s) G-1ac85c33 — blue sits again this epoch')), world.logs.join(' | '))
  assert.ok(world.logs.some((l) => l.includes('G-00000003 it does not sit again on — this was its second sitting this epoch')), world.logs.join(' | '))
})
test('R2: no second sitting without a gap found closed, or on a successor this epoch already engaged blue on', async () => {
  const sittings = async (env, ...gaps) => {
    const w = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('blue-respond', ...gaps)]) }), passChair()], blueRespond: [env] }))
    await w.run(script, ARGS)
    return w.calls.filter((c) => /^blue-respond /.test(c.opts.label)).length
  }
  assert.equal(await sittings(blueEnv({ manifest: [], successors: ['G-00000002'] }), 'G-00000001'), 1, 'a successor relayed with no gap found closed was sat')
  assert.equal(await sittings(blueEnv({ manifest: [], found_closed: ['G-00000001'], successors: ['G-00000002'] }), 'G-00000001', 'G-00000002'), 1, 'a successor the plan already engaged blue on was sat twice')
  assert.equal(await sittings(blueEnv({ manifest: [], found_closed: ['G-00000001'] }), 'G-00000001'), 1, 'a gap closed with no successor sat blue again')
})
test('a respond envelope takes gap ids in successors, and refuses prose there', async () => {
  const w = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('blue-respond', 'G-00000001')]) }), passChair()] }))
  await w.run(script, ARGS)
  const schema = w.calls.find((c) => /^blue-respond /.test(c.opts.label)).opts.schema
  assert.ok(schemaAccepts(schema, { successors: ['G-0a1b2c3d'] }), 'a gap id is refused')
  assert.ok(!schemaAccepts(schema, { successors: ['the successor'] }), 'a prose successor is accepted')
})

test('W2b: partial coverage is logged, never fatal; a found_closed id blue was not engaged on is ignored', async () => {
  const partial = makeWorld(makeResponder({ chair: engagedOn('G-00000001', 'G-00000002'), blueRespond: [blueEnv({ manifest: ['G-00000001'], found_closed: ['G-00000009'] })] }))
  const out = await partial.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  assert.ok(partial.logs.some((l) => l.includes('1/2 gap(s) open when blue sat') && l.includes('no row for: G-00000002')))
  assert.ok(!partial.logs.some((l) => l.includes('G-00000009')), 'a claimed closure outside the engagement excuses nothing')
})

test('a log a seat returns in its envelope reaches nothing: not the workflow return, not assembly', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G-00000001')]), log: ['no PDF extraction'] }), passChair()],
    blueRespond: [blueEnv({ log: ['rate-limited on WebFetch'] })],
    blueSynth: [blueEnv({ log: ['write-block on blue/report.md'] })],
  }))
  const out = await world.run(script, ARGS)
  // THE NEGATIVE CONTROL. A seat's log entries are on the record; an envelope copy is read by
  // nobody, so three seats returning one leave no trace in what the engine returns or hands on.
  const returned = JSON.stringify(out)
  const handed = world.calls.map((c) => c.prompt).join('\n')
  for (const text of ['no PDF extraction', 'rate-limited on WebFetch', 'write-block on blue/report.md']) {
    assert.ok(!returned.includes(text), `the workflow returned a seat's envelope log: ${text}`)
    assert.ok(!handed.includes(text), `a prompt carries a seat's envelope log: ${text}`)
  }
})

// WHICH SEAT RIDES WHICH TIER, UNDER WHICH CONFIGURATION, IS NOT ASSERTED HERE. Both are the
// roster's, written into the script's generated seat block, and the dispatch identity test in
// tools/integration/surface holds every dispatch to them.
test('both tiers are required and every dispatch carries one of them; binDir is required', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G-00000001'), party('judge', 'G-00000001')]) }), passChair()] }))
  await world.run(script, { ...ARGS, model: 'haiku', judgmentModel: 'opus' })
  for (const c of world.calls) assert.ok(['haiku', 'opus'].includes(c.opts.model), `${c.opts.label} is dispatched on a tier the run named`)
  assert.deepEqual([...new Set(world.calls.map((c) => c.opts.model))].sort(), ['haiku', 'opus'], 'both tiers are in use')
  const base = { topic: 't', runDir: 'research/x', binDir: '/b' }
  await assert.rejects(makeWorld(makeResponder()).run(script, { ...base, judgmentModel: 'sonnet' }), /refusing dispatch — model unset/)
  await assert.rejects(makeWorld(makeResponder()).run(script, { ...base, model: 'sonnet' }), /refusing dispatch — judgmentModel unset/)
  await assert.rejects(makeWorld(makeResponder()).run(script, { topic: 't', runDir: 'research/x', model: 'sonnet', judgmentModel: 'sonnet' }), /binDir unset/)
})

test('the lenses dispatched are the plan\'s, each under a configuration of its own', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('red-lens-adversary'), party('red-lens-architecture')]) }), passChair()] }))
  await world.run(script, ARGS)
  assert.deepEqual(labelsOf(world, 'red-lens').map((c) => c.opts.label), ['red-lens-adversary #1 · test', 'red-lens-architecture #1 · test'])
  assert.equal(new Set(labelsOf(world, 'red-lens').map((c) => c.opts.agentType)).size, 2, 'two lenses, two configurations')
  assert.equal(labelsOf(world, 'red-lens-evidence').length, 0, 'an area the plan did not ready does not sit — the record decides, not the roster')
})

// EVERY LENS SITTING ENDS BY RETURNING AN ENVELOPE, so the record is asked whether it refuses it: the
// PreToolUse gate on the call that returns one refuses it while a blocking item on the lens's own
// work list stands. A lens dispatched with no schema ends with text and that gate never runs. The
// envelope carries nothing — what the lens did is on the record — and it is closed.
test('every lens is dispatched with the closed, empty lens envelope', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan(LENS_AREAS.map((a) => party(`red-lens-${a}`))) }), passChair()] }))
  await world.run(script, ARGS)
  const lenses = labelsOf(world, 'red-lens')
  assert.equal(lenses.length, LENS_AREAS.length, 'the drive dispatches every lens')
  for (const c of lenses) {
    assert.ok(c.opts.schema, `${c.opts.label} is dispatched with no schema`)
    assert.ok(schemaAccepts(c.opts.schema, {}), `${c.opts.label}: the empty envelope is refused`)
    assert.ok(!schemaAccepts(c.opts.schema, { found: 'G-00000001' }), `${c.opts.label}: a key the envelope does not declare is accepted`)
  }
})

test('the lens constitutions say a lens mints and the chair runs the debate', () => {
  const chair = readFileSync(new URL('../../agents/red-chair.md', import.meta.url), 'utf8')
  const lens = readFileSync(new URL('../../agents/red-lens-evidence.md', import.meta.url), 'utf8')
  assert.ok(!/COALESCE/.test(chair) && !/board's only writer/.test(chair), 'the chair constitution still describes the coalescing merge')
  assert.ok(/dispatch/.test(chair), 'the chair constitution names its dispatch duty')
  assert.ok(/mint/i.test(lens), 'the lens constitution names minting')
  for (const area of LENS_AREAS) {
    const md = readFileSync(new URL(`../../agents/red-lens-${area}.md`, import.meta.url), 'utf8')
    assert.ok(md.includes('the record readies you for your first look, again whenever the report head is past your pin — the head the tool last served you the whole report at, in a sitting you sat through, so a sitting that ends without the whole report served to you leaves you ready at the head — and whenever your gap needs acting on. Three sittings in a row without fresh material retire you, and a head move does not ready you after that — unless your work list still owes something'), `red-lens-${area}: readiness is the pin behind the head, until retirement owing nothing`)
    assert.ok(!/re-arm/.test(md) && !/two sittings without one retire you/.test(md), `red-lens-${area}: no lens is told it sits twice at one head`)
    assert.ok(md.includes("a corroborating source's title ships in the source's note and Bibliography entry") && md.includes('replacement text you prescribe'), `red-lens-${area}: the title and the prescription are red text the report prints`)
  }
})

// EVERY FIELD THE WORKFLOW READS FROM THE RELAY IS REFUSED LOUD WHEN MISSING OR MISTYPED — one table,
// so the check in debate.js and this list cannot drift. A fallback here once read a dropped docket as
// "nothing docketed"; each row now throws, naming the field.
test('a relayed plan missing a required field, or carrying the wrong type, aborts naming the field', async () => {
  const cases = [
    ['head', (p) => { delete p.head }], ['head', (p) => { p.head = '7' }],
    ['parties', (p) => { delete p.parties }], ['parties', (p) => { p.parties = null }],
    ['pass_permitted', (p) => { delete p.pass_permitted }], ['pass_permitted', (p) => { p.pass_permitted = 'yes' }],
    ['ceiling', (p) => { delete p.ceiling }], ['ceiling', (p) => { p.ceiling = 0 }],
    ['docket', (p) => { delete p.docket }], ['docket', (p) => { p.docket = null }],
    ['why', (p) => { delete p.why }], ['why', (p) => { p.why = null }],
    ['max_epochs', (p) => { delete p.max_epochs }], ['max_epochs', (p) => { p.max_epochs = '3' }],
    ['epoch_limit_reached', (p) => { delete p.epoch_limit_reached }], ['epoch_limit_reached', (p) => { p.epoch_limit_reached = 1 }],
    ['stale_areas', (p) => { delete p.stale_areas }], ['stale_areas', (p) => { p.stale_areas = null }],
    ['remand_owed', (p) => { delete p.remand_owed }], ['remand_owed', (p) => { p.remand_owed = 'G-00000001' }],
    ['remand_owed[0].gap_id', (p) => { p.remand_owed = [{ direction: 'D' }] }],
    ['remand_owed[0].direction', (p) => { p.remand_owed = [{ gap_id: 'G-00000001' }] }],
    ['remand_owed[0].direction', (p) => { p.remand_owed = [{ gap_id: 'G-00000001', direction: '  ' }] }],
    ['parties[0].seat_id', (p) => { p.parties[0].seat_id = 7 }],
    ['parties[0].gap_ids', (p) => { p.parties[0].gap_ids = null }],
    ['stale_areas[0].seat_id', (p) => { p.stale_areas = [{ seat_id: 7, pin: 3 }] }],
    ['stale_areas[0].pin', (p) => { p.stale_areas = [{ seat_id: 'red-lens-voice', pin: '3' }] }],
  ]
  for (const [field, spoil] of cases) {
    const p = plan([party('red-lens-evidence'), party('blue-respond', 'G-00000001')])
    spoil(p)
    const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: p })] }))
    await assert.rejects(world.run(script, ARGS), (err) => err.message.includes(`\`${field}\``), `a plan spoiled at ${field} must abort naming it`)
  }
  // And a faithful plan with a stale area relays.
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([party('red-lens-evidence')], { stale_areas: [{ seat_id: 'red-lens-voice', pin: 3 }] }) }), passChair()] }))
  assert.equal((await world.run(script, ARGS)).verdict, 'VERIFIED')
})

// A plan naming every area dispatches one lens sitting per area, one agent configuration per area.
test('every lens area sits under a configuration of its own', async () => {
  const areas = LENS_AREAS
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan(areas.map((a) => party(`red-lens-${a}`))) }), passChair()] }))
  await world.run(script, ARGS)
  assert.deepEqual(labelsOf(world, 'red-lens').map((c) => c.opts.label).sort(), areas.map((a) => `red-lens-${a} #1 · test`).sort(), 'one lens sitting per area')
  assert.equal(new Set(labelsOf(world, 'red-lens').map((c) => c.opts.agentType)).size, areas.length, 'one configuration per area')
})

// THE PATTERN DUTY REACHES THE SEAT THAT OWES IT, and nothing pinned that until now.
//
// The clause had no call site: defined, never invoked, so red's accumulated memory never reached a
// repairing seat at all. `grep -rl 'PATTERN DUTY' --include='*.golden'` returned nothing, because a
// clause that never renders cannot move a golden — the same plausible zero this engine keeps
// finding, here in the delivery path of a duty both constitutions advertise as the one that works.
// NO SEAT IS SENT TO A PATTERN CORPUS. The promoted corpus is retired: the gap class's closed set is
// the vocabulary, enforced at mint, and a seat reads its situation from the record. Measured
// before the removal: all three repairing seats carried the duty and none opened the file.
test('no prompt sends a repairing seat to a gap-pattern corpus', async () => {
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G-00000001')]) }), passChair()],
  }))
  await world.run(script, ARGS)
  for (const c of world.calls) {
    assert.ok(!/PATTERN DUTY/.test(c.prompt), `a pattern duty reached ${c.opts.label}`)
    assert.ok(!/gap-patterns-by-class|red-gap-patterns/.test(c.prompt), `a corpus path reached ${c.opts.label}`)
  }
})

// THE CORPUS IS NOT RELAYED THROUGH THE WORKFLOW ARGUMENT. Passing the parsed index cost 38,473
// characters of `args` and 125 seconds of generation on one smoke — the lead retyping a file that
// was already on disk beside the seat that needed it. A reintroduced relay would not fail any other
// test: it would just make every run slower, which is exactly the kind of regression nothing
// notices.
test('a gapPatterns argument is ignored: the corpus never travels through args', async () => {
  const withCorpus = { ...ARGS, gapPatterns: { 'unproven-claim': [{ file: 'f.md', title: 'T', hook: 'H' }] } }
  const world = makeWorld(makeResponder({
    chair: [chairEnv({ plan: plan([party('blue-respond', 'G-00000001')]) }), passChair()],
  }))
  await world.run(script, withCorpus)
  for (const c of world.calls) {
    assert.ok(!c.prompt.includes('f.md'), `a relayed pattern filename reached ${c.opts.label}`)
    assert.ok(!c.prompt.includes('T — H'), `a relayed pattern body reached ${c.opts.label}`)
  }
})

// ── the envelope schema and the plan check agree ────────────────────────────────────────────

// The schema the chair's dispatch carries, as the engine hands it to the harness.
const chairSchema = async () => {
  const world = makeWorld(makeResponder())
  await world.run(script, ARGS)
  return labelsOf(world, 'red-chair')[0].opts.schema
}
const relayEnds = async (env) => {
  const world = makeWorld(makeResponder({ chair: [env, passChair()] }))
  return { world, out: await world.run(script, ARGS) }
}

// m16 (2026-10-08): chair sitting 6 relayed a plan it had rebuilt by hand, with `occasions: []` on
// both lenses and blue. The schema accepted the envelope; the plan check threw on it; the run ended
// there, 42 minutes in. The fixture is that sitting's plan, verdict and notes, off its transcript; the
// envelope below is its plan and notes, as the chair's envelope now carries them.
// `exchanges` and the key are the fields written into it here: the plan carries an entry per engaged gap, and
// the sitting's transcript states no count for any of the four. The rebuilt plan also dropped every
// blocker, and a plan that permits no PASS names what holds it, so the schema refuses that relay
// now; the empty occasions lists are accepted once its blockers stand.
test('m16: chair sitting 6\'s relay — an empty occasions list on each lens and blue — is accepted by the schema and dispatches its parties', async () => {
  const relayed = JSON.parse(readFileSync(new URL('./testdata/m16-chair6-relay.json', import.meta.url), 'utf8'))
  assert.deepEqual(relayed.plan.parties.map((p) => p.occasions), [[], [], []], 'the fixture is the relay that ended the run')
  assert.deepEqual([relayed.plan.blockers, relayed.plan.pass_permitted], [[], false], 'the fixture dropped every blocker')
  assert.ok(!schemaAccepts(await chairSchema(), { ...relayed.plan, notes: relayed.notes }), 'the schema refuses a plan that permits no PASS and names nothing holding it')
  const env = { ...withKey({ ...relayed.plan, blockers: ['G14', 'G15', 'G16', 'G17'].map((g) => blocker(g, 'red-lens-adversary', 'material_gap')) }), notes: relayed.notes }
  assert.ok(schemaAccepts(await chairSchema(), env), 'the schema accepts the empty occasions lists')
  const { world, out } = await relayEnds(env)
  assert.deepEqual(world.calls.map((c) => c.opts.label).filter((l) => / #1 /.test(l) && !l.startsWith('red-chair')),
    ['red-lens-adversary #1 · test', 'red-lens-architecture #1 · test', 'blue-respond #1 · test'])
  assert.ok(labelsOf(world, 'blue-respond #1')[0].prompt.includes('You are engaged on: G14, G15, G16, G17.'))
  assert.equal(out.verdict, 'VERIFIED')
})

// o3 (2026-10-13): chair sitting 5 nested its whole envelope under `plan`; the harness refused it for
// missing every plan field, and the chair re-sent the nest with each field added beside it, emptied.
// The open schema took that relay, the engine read its empty `parties` as the record's word that
// nobody sits, and the run ended UNVERIFIED with the four parties the record dispatched never
// sitting. Both relays and the plan the verb printed are off the sitting's transcript.
test('o3: chair sitting 5\'s re-sent relay — the plan nested under `plan`, the fields beside it emptied — is refused; the plan the verb printed is accepted and dispatches its four parties', async () => {
  const { printed, relays } = JSON.parse(readFileSync(new URL('./testdata/o3-chair5-relays.json', import.meta.url), 'utf8'))
  assert.deepEqual(relays.map((r) => r.accepted), [false, true], 'the fixture is the refused relay and the one the harness took')
  const [, taken] = relays
  assert.deepEqual([Object.keys(taken.envelope.plan).includes('plan'), taken.envelope.plan.parties, taken.envelope.plan.blockers], [true, [], []], 'the taken relay is the nest with its fields emptied')
  const schema = await chairSchema()
  for (const r of relays) assert.ok(!schemaAccepts(schema, r.envelope), 'the schema refuses each relay the chair sent')
  const { plan: _nested, verdict: _verdict, ...beside } = taken.envelope.plan
  assert.ok(!schemaAccepts(schema, beside), 'the emptied fields alone are refused: a plan that permits no PASS names what holds it')
  assert.ok(schemaAccepts(schema, printed), 'the plan the verb printed is accepted')
  assert.ok(!schemaAccepts(schema, { ...printed, plan: printed }), 'a key the verb does not print is refused, beside a faithful plan too')
  const { world } = await relayEnds(printed)
  assert.deepEqual(['red-lens-logic', 'red-lens-voice', 'blue-respond', 'judge'].map((s) => world.calls.some((c) => c.opts.label.startsWith(`${s} `))), [true, true, true, true], 'each party the record dispatched sits')
})

// m3 (2026-10-13): chair sitting 3 relayed the plan with `blockers` emptied, where the verb printed
// eleven; the schema took it, and capture's dispatch-parity audit failed the run on it afterwards.
test('m3: chair sitting 3\'s relay — every blocker dropped from a plan with eight parties — is refused', async () => {
  const { printed, relays } = JSON.parse(readFileSync(new URL('./testdata/m3-chair3-relays.json', import.meta.url), 'utf8'))
  const [only] = relays
  assert.deepEqual([only.accepted, only.envelope.plan.blockers.length, printed.blockers.length, only.envelope.plan.parties.length], [true, 0, 11, 8], 'the fixture is the relay the harness took')
  const schema = await chairSchema()
  assert.ok(!schemaAccepts(schema, { ...only.envelope.plan, notes: only.envelope.notes }), 'the schema refuses the relay')
  assert.ok(schemaAccepts(schema, { ...printed, notes: only.envelope.notes }), 'the plan the verb printed is accepted')
})

// A RELAY THE HARNESS LETS THROUGH EMPTIED STOPS THE RUN, NAMING IT. The schema refuses it first; the
// engine's own check stands behind the schema, and an emptied plan never reads as "nobody sits".
test('a plan that permits no PASS and names nothing holding it stops the run rather than ending the debate', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv({ plan: plan([], { blockers: [] }) }), passChair()] }))
  await assert.rejects(world.run(script, ARGS), /red-chair sitting 1 relayed a plan whose `blockers` is empty and whose `pass_permitted` is false/)
  assert.equal(labelsOf(world, 'judge · assemble').length, 0, 'no outcome is recorded on an emptied plan')
})

// EACH SHAPE THE PLAN CHECK REFUSES IS ONE THE SCHEMA REFUSES FIRST — named, so a reader sees which
// shapes those are; the sweep below is what holds the two together.
test('the schema refuses what the plan check refuses: a bench with no occasion, an occasion on a lens or blue, a seat with no sitting, a blank direction', async () => {
  const schema = await chairSchema()
  const refused = {
    'a bench party with no occasions': plan([{ seat_id: 'judge', gap_ids: ['G-00000001'] }]),
    'a bench party with an empty occasions list': plan([{ seat_id: 'judge', gap_ids: ['G-00000001'], occasions: [] }]),
    'a bench party convened for a word that is not an occasion the plan names': plan([{ seat_id: 'judge', gap_ids: [], occasions: ['terminal'] }]),
    'a lens convened for an occasion': plan([{ seat_id: 'red-lens-logic', gap_ids: [], occasions: ['docket'] }]),
    'blue convened for an occasion': plan([{ seat_id: 'blue-respond', gap_ids: ['G-00000001'], occasions: ['petition'] }]),
    'a seat the workflow has no sitting for': plan([party('frontier')]),
    'a lens area the workflow does not declare': plan([party('red-lens-typography')]),
    'a remand with a blank direction': plan([], { remand_owed: [{ gap_id: 'G-00000001', direction: ' ' }] }),
  }
  for (const [what, p] of Object.entries(refused)) {
    assert.ok(!schemaAccepts(schema, chairEnv({ plan: p })), `the schema must refuse ${what}`)
  }
  const accepted = {
    'the bench on its docket': plan([party('judge', 'G-00000001')], { docket: ['G-00000001'] }),
    'the bench for a petition and its docket': plan([{ seat_id: 'judge', gap_ids: ['G-00000001'], occasions: ['petition', 'docket'] }]),
    'a lens with no occasions field': plan([party('red-lens-logic')]),
    'a lens with an empty occasions list': plan([{ seat_id: 'red-lens-logic', gap_ids: [], occasions: [] }]),
  }
  for (const [what, p] of Object.entries(accepted)) {
    assert.ok(schemaAccepts(schema, chairEnv({ plan: p })), `the schema must accept ${what}`)
  }
})

// THE GATE: NO RELAY THE SCHEMA ACCEPTS ENDS THE RUN. Every position in a full plan is replaced by
// every value in the pool — and deleted — and each resulting relay the schema accepts is run
// through the engine. A check added to the engine with no shape for it in the schema fails here,
// naming the relay; so does a shape loosened in the schema that a check still throws on.
// THE CHAIR'S ENVELOPE IS THE PLAN, AND DECLARES NOTHING BESIDE IT THE ENGINE DOES NOT READ. The
// plan's fields are the record's, relayed whole for capture to compare; a field beside them is one
// the seat composes into, so nothing it writes there goes unread. Each such key is read below as
// `chairEnv.<key>`; a key added beside the plan must be read there.
test('the chair envelope is the plan, and declares no field beside it the engine does not read', async () => {
  const schema = await chairSchema()
  const src = readFileSync(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url), 'utf8')
  const relayed = new Set([...Object.keys(plan()), 'verdict'])
  assert.deepEqual([...relayed].filter((k) => !schema.properties[k]), [], 'every field the verb prints is a field of the envelope')
  const beside = Object.keys(schema.properties).filter((k) => !relayed.has(k))
  assert.deepEqual(beside, ['notes'], 'the envelope adds one field to the plan')
  for (const key of beside) {
    assert.ok(src.includes(`chairEnv.${key}`), `CHAIR_ENVELOPE declares \`${key}\` beside the plan, which the engine never reads as chairEnv.${key}`)
  }
})

// THE ENVELOPE KEEPS A FREE FIELD BESIDE THE PLAN'S, UNDESCRIBED. Replaying an m28 chair at its return
// (2026-10-18), an envelope of `plan` and `verdict` alone had it wrap the whole envelope inside `plan`
// on 4 of 9 first returns — m28 itself did so in 11 of 12 sittings — against 0 of 16 with a bare
// `notes`; a described `notes` drew 2 of 12. Cutting or describing it is a measured regression.
test('the chair envelope keeps its bare notes field', async () => {
  const schema = await chairSchema()
  assert.deepEqual(schema.properties.notes, { type: 'string' }, 'CHAIR_ENVELOPE.notes is a bare string: removed or described, the m28 replay wrapped the envelope more often')
})

// THE CHAIR'S SCHEMA IS A TOOL'S INPUT SCHEMA, AND THE API REFUSES ONE WITH oneOf, allOf OR anyOf AT ITS
// TOP LEVEL: every chair sitting's request fails with a 400 before the chair is asked anything. The
// envelope is the plan, so the plan's cross-field rule is a conditional there, never a combinator.
test('the chair envelope schema holds no oneOf, allOf or anyOf at its top level', async () => {
  const schema = await chairSchema()
  assert.deepEqual(['oneOf', 'allOf', 'anyOf'].filter((k) => k in schema), [], 'the API refuses this tool schema')
  assert.ok(schema.if && schema.else, 'the empty-only-with rule stands, as a conditional')
})

// EVERY DISPATCHED ENVELOPE IS CLOSED AT ITS TOP LEVEL: a key no schema declares is refused, never
// dropped. Driven through a run that reaches the petition bench, the docket, the terminal sitting
// and assembly, so every schema the script dispatches is read as dispatched.
test('every envelope schema the script dispatches is closed at its top level', async () => {
  const chair = [chairEnv({ plan: plan([petitionBench(), party('red-lens-evidence'), party('blue-respond', 'G-00000001'), party('judge', 'G-00000001')], { docket: ['G-00000001'], blockers: [blocker('M-00000001', 'judge')] }) }),
    chairEnv({ plan: plan([], { blockers: [blocker('M-00000002', 'judge')] }) })]
  const world = makeWorld(makeResponder({ chair }))
  await world.run(script, ARGS)
  const schemas = new Map(world.calls.filter((c) => c.opts.schema).map((c) => [c.opts.label.replace(/ #\d+/, '').replace(/ · test$/, ''), c.opts.schema]))
  assert.deepEqual([...schemas.keys()].sort(), ['blue-respond', 'blue-synthesize', 'judge', 'judge · assemble', 'judge · petition', 'judge · terminal', 'red-chair', 'red-lens-evidence'], 'the drive reaches every envelope')
  for (const [label, schema] of schemas) assert.equal(schema.additionalProperties, false, `${label}'s envelope is open at its top level`)
})

test('a relay the envelope schema accepts is never one the engine throws on', async () => {
  const schema = await chairSchema()
  const full = () => chairEnv({
    plan: plan(
      [party('red-lens-evidence', 'G-00000001'), party('blue-respond', 'G-00000001', 'G-00000002'), { seat_id: 'judge', gap_ids: ['G-00000002'], occasions: ['petition', 'docket'] }],
      { docket: ['G-00000002'], remand_owed: [{ gap_id: 'G-00000001', direction: 'find the primary source' }], exchanges: [{ gap_id: 'G-00000001', count: 2, unresolved: 1 }, { gap_id: 'G-00000002', count: 3, unresolved: 0 }], stale_areas: [{ seat_id: 'red-lens-voice', pin: 3 }], blockers: [blocker('M-00000001', 'judge')], why: ['G-00000001: open'], max_epochs: 12 }),
    notes: 'n',
  })
  const paths = []
  const walk = (v, path) => {
    paths.push(path)
    if (Array.isArray(v)) v.forEach((x, i) => walk(x, [...path, i]))
    else if (v && typeof v === 'object') for (const k of Object.keys(v)) walk(v[k], [...path, k])
  }
  walk(full(), [])
  const DELETE = Symbol('delete')
  const pool = [DELETE, null, '', ' ', 'x', 'judge', 'blue-respond', 'red-lens-logic', 'frontier', 'docket', 0, -1, 1.5, 7, true, false, {}, [], [''], ['x'], ['docket'], ['petition'], ['terminal'], [{}], 'PASS', 'FAIL']
  // A key the verb omits and a seat may add: `occasions` on the lens and on blue, and the verdict the
  // verb prints only once the sitting has recorded one.
  const added = [['parties', 0, 'occasions'], ['parties', 1, 'occasions'], ['verdict']]
  let accepted = 0, refused = 0
  for (const path of [...paths.filter((p) => p.length), ...added]) {
    for (const value of pool) {
      const env = full()
      let at = env
      for (const k of path.slice(0, -1)) at = at[k]
      const last = path[path.length - 1]
      if (value === DELETE) { if (Array.isArray(at)) at.splice(last, 1); else delete at[last] } else at[last] = value
      const what = `${path.join('.')} = ${value === DELETE ? '(deleted)' : JSON.stringify(value)}`
      if (!schemaAccepts(schema, env)) { refused++; continue }
      accepted++
      const { out } = await relayEnds(env).catch((err) => assert.fail(`the schema accepts the relay with ${what}, and the engine throws on it: ${err.message}`))
      assert.ok(!out.failure, `the schema accepts the relay with ${what}, and the engine ends the run on it: ${out.failure}`)
    }
  }
  // Both arms ran: a sweep in which the schema refused everything, or nothing, measured no agreement.
  assert.ok(accepted > 50 && refused > 50, `the sweep must exercise both arms; it accepted ${accepted} and refused ${refused}`)
})

// ── a run the engine stops ──────────────────────────────────────────────────────────────────

// AN ENGINE STOP RECORDS NO OUTCOME, AND THE RUN STAYS OPEN. Only the bench's assembly sitting writes
// an outcome, and an outcome closes the record to every seat — so a stop that convened it would
// finish a run an operator can still resume. The workflow rejects with the error and that is all it
// does: the stop itself reaches the record from the lead, which holds the error.
const strayPlan = () => chairEnv({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G-00000001'), party('frontier')], { blockers: [blocker('M-00000001', 'judge')] }) })
test('an engine stop convenes nobody: no party of the plan it stopped on, no terminal sitting, and no assembly sitting to record an outcome', async () => {
  const world = makeWorld(makeResponder({ chair: [chairEnv(), strayPlan(), passChair()] }))
  await assert.rejects(world.run(script, ARGS), /epoch 2: the plan names frontier/)
  assert.equal(world.calls[world.calls.length - 1].opts.label, 'red-chair #2 · test', 'the sitting the engine stopped on is the last one dispatched')
  assert.equal(labelsOf(world, 'judge · terminal').length, 0, 'the plan the engine stopped on convenes no terminal sitting, whatever blockers it lists')
  assert.equal(labelsOf(world, 'judge · assemble').length, 0, 'the assembly sitting records the outcome, and a stopped run has none')
})

// A RESUME REPLAYS THE SITTINGS THAT COMPLETED AND RUNS THE REST. The Workflow tool's resume hands a
// completed sitting its cached result when the script asks for it again; the stand-in here is a
// cache keyed by the dispatch's label and prompt, so a replayed sitting hits only if the script
// asks for it in the same words. What the tool itself keys on is not measured here.
const cachingWorld = (cache, fresh) => {
  const tally = { hits: 0, ran: [] }
  const world = makeWorld((p, o) => {
    const key = `${o.label}\n${p}`
    if (cache.has(key)) { tally.hits++; return cache.get(key) }
    const out = fresh(p, o)
    cache.set(key, out)
    tally.ran.push(o.label)
    return out
  })
  return { world, tally }
}
test('a stopped run resumes: every completed sitting replays under the prompt it first had, an unchanged resume stops at the same sitting, and one that sits it again reaches the outcome', async () => {
  const cache = new Map()
  const first = cachingWorld(cache, makeResponder({ chair: [chairEnv(), strayPlan()] }))
  await assert.rejects(first.world.run(script, ARGS), /the plan names frontier/)
  const sat = first.tally.ran.length
  assert.ok(sat > 5 && first.tally.hits === 0, `the first run sits every seat fresh; it sat ${sat} and replayed ${first.tally.hits}`)

  // Nothing changed: the stop replays from the cache, and no seat sits again.
  const same = cachingWorld(cache, () => { throw new Error('a replay with nothing changed dispatched a seat') })
  await assert.rejects(same.world.run(script, ARGS), /the plan names frontier/)
  assert.equal(same.tally.hits, sat, 'every sitting of the stopped run is asked for again in the same words')

  // The sitting the engine stopped on sits again: everything before it replays, and the run goes on.
  for (const k of [...cache.keys()]) if (k.startsWith('red-chair #2')) cache.delete(k)
  const resumed = cachingWorld(cache, makeResponder({ chair: [passChair()] }))
  const out = await resumed.world.run(script, ARGS)
  assert.equal(resumed.tally.hits, sat - 1, 'every sitting but the one that stopped the run replays')
  assert.deepEqual(resumed.tally.ran, ['red-chair #2 · test', 'judge · assemble · test'], 'only the stopped sitting and what follows it run')
  assert.equal(out.verdict, 'VERIFIED')
  assert.equal(out.gaps_outstanding, 0)
})

// ── every dispatched schema compiles strict ────────────────────────────────────────────────

// EVERY SCHEMA THE ENGINE DISPATCHES COMPILES UNDER STRICT TYPES WITH NO WARNING. The harness
// compiles each one at launch and logs what strict types finds, and keeps going: the chair's PLAN
// carried a `minItems` with no `type` in an anyOf branch, and every smoke run's launch log said so
// while every test here stayed green. The worlds below reach each sitting the engine dispatches
// with a schema, and the census is held to the source: one schema per `schema: <NAME>` it declares.
test('every envelope schema the engine dispatches compiles under strict types with no warning', async () => {
  const src = readFileSync(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url), 'utf8')
  const declared = new Set([...src.matchAll(/\bschema: ([A-Z_]+)\b/g)].map((m) => m[1]))
  const worlds = [
    makeResponder(),
    makeResponder({ chair: [chairEnv({ plan: plan([petitionBench(), party('blue-respond', 'G-00000001')], { blockers: [blocker('M-00000001', 'judge')] }) }), passChair()] }),
    makeResponder({ chair: [chairEnv({ plan: plan([party('judge', 'G-00000001')]) }), passChair()] }),
    makeResponder({ chair: [chairEnv({ plan: ceilingPlan({ blockers: [blocker('M-00000001', 'judge')] }) })] }),
  ]
  const seen = new Map()
  for (const respond of worlds) {
    const world = makeWorld(respond)
    await world.run(script, ARGS)
    for (const c of world.calls) if (c.opts.schema) seen.set(JSON.stringify(c.opts.schema), c.opts)
  }
  assert.equal(seen.size, declared.size, `the worlds reach ${seen.size} distinct schemas and the engine declares ${declared.size} (${[...declared].join(', ')}) — a schema no world dispatches is not checked`)
  for (const opts of seen.values()) assert.deepEqual(strictTypesWarnings(opts.schema), [], `${opts.label}: the schema warns under strict types`)
})

// The check is ajv's, line for line, on the schema that tripped it: a branch's property with a
// typed keyword and no type, and the same property typed.
test('the strict-types check reports a typed keyword with no type, and only that', () => {
  const branch = (blockers) => ({ type: 'object', properties: { blockers: { type: 'array' } }, anyOf: [{ required: ['blockers'], properties: { blockers } }] })
  assert.deepEqual(strictTypesWarnings(branch({ minItems: 1 })), ['strict mode: missing type "array" for keyword "minItems" at "#/anyOf/0/properties/blockers" (strictTypes)'])
  assert.deepEqual(strictTypesWarnings(branch({ type: 'array', minItems: 1 })), [])
  assert.deepEqual(strictTypesWarnings({ type: 'string', anyOf: [{ type: 'integer' }] }), ['strict mode: type "integer" not allowed by context "string" at "#/anyOf/0" (strictTypes)'])
  assert.deepEqual(strictTypesWarnings({ type: ['string', 'integer'] }), ['strict mode: use allowUnionTypes to allow union type keyword at "#" (strictTypes)'])
  assert.deepEqual(strictTypesWarnings({ type: ['string', 'null'], enum: ['PASS', null] }), [])
})

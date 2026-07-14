// node --test — the debate engine's regression suite. Run:
//   node --test plugins/frank-exchange-of-views/tests/simulator/
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { loadDebateScript, makeWorld, makeResponder, blueEnv, redEnv, gap, judgeEnv } from './harness.mjs'

const script = loadDebateScript(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url))
const ARGS = { topic: 'test topic', runDir: 'research/2026-01-01_test', lanes: 2 }
const isJudgmentSeat = (l) => /^(blue-synthesize|red-merge|judge|assemble)/.test(l)

test('founding regression 1: JSON-stringified args parse; no undefined leaks into prompts', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  const out = await world.run(script, JSON.stringify(ARGS))
  assert.equal(out.verdict, 'VERIFIED')
  for (const c of world.calls) assert.ok(!c.prompt.includes('undefined/'), `undefined path in: ${c.opts.label}`)
})

test('founding regression 1b: unbound topic/runDir refuses dispatch before any agent spawns', async () => {
  const world = makeWorld(makeResponder())
  await assert.rejects(world.run(script, '{}'), /refusing dispatch/)
  assert.equal(world.calls.length, 0)
})

test('founding regression 2: null red-merge aborts cleanly, not with a TypeError', async () => {
  const world = makeWorld((p, o) => (o.label.startsWith('red-merge') ? null : makeResponder()(p, o)))
  await assert.rejects(world.run(script, ARGS), /red-merge round 1 returned null/)
})

test('founding regression 2b: null blue synthesis aborts cleanly', async () => {
  const world = makeWorld((p, o) => (o.label.startsWith('blue-synthesize') ? null : makeResponder()(p, o)))
  await assert.rejects(world.run(script, ARGS), /blue synthesis returned null/)
})

test('happy path: red PASS on round 1 -> VERIFIED, phases in order, lanes honored', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  const out = await world.run(script, ARGS)
  assert.deepEqual({ verdict: out.verdict, rounds: out.rounds, deadlocked: out.deadlocked, gaps: out.gaps_outstanding },
    { verdict: 'VERIFIED', rounds: 1, deadlocked: false, gaps: 0 })
  assert.deepEqual(world.phases, ['Frontier', 'Blue', 'Assemble'])
  assert.equal(world.calls.filter((c) => c.opts.label.startsWith('blue-lane')).length, 2)
  assert.ok(world.logs.some((m) => m.includes('researching: test topic')))
})

test('per-role models: bulk seats get `model`; judgment seats inherit unless judgmentModel set', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, { ...ARGS, model: 'sonnet' })
  for (const c of world.calls) {
    if (isJudgmentSeat(c.opts.label)) assert.equal(c.opts.model, undefined, `judgment seat ${c.opts.label} must inherit session model`)
    else assert.equal(c.opts.model, 'sonnet', `bulk seat ${c.opts.label} must take the bulk model`)
  }

  const world2 = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world2.run(script, { ...ARGS, model: 'haiku', judgmentModel: 'opus' })
  for (const c of world2.calls) assert.equal(c.opts.model, isJudgmentSeat(c.opts.label) ? 'opus' : 'haiku')
})

test('contested docket: a re-raised gap goes to the judge; adjudicated gaps leave red verdict scope', async () => {
  const world = makeWorld(makeResponder({
    red: [
      redEnv({ gaps: [gap('R1-1'), gap('R1-2')] }),         // round 1: two new gaps, no docket
      redEnv({ gaps: [gap('R1-1')] }),                       // round 2: R1-1 re-raised -> contested
      redEnv({ verdict: 'PASS' }),                           // round 3
    ],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'closed', rationale: 'blue fixed it' }] })],
  }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  assert.equal(out.rounds, 3)
  const judgeCalls = world.calls.filter((c) => c.opts.label.startsWith('judge'))
  assert.equal(judgeCalls.length, 1, 'judge invoked exactly once, only when a gap recurs')
  assert.ok(judgeCalls[0].prompt.includes('"R1-1"'))
  assert.ok(!judgeCalls[0].prompt.includes('"R1-2"'), 'un-recurred gap is not on the docket')
  const merge3 = world.calls.find((c) => c.opts.label.startsWith('red-merge-r3'))
  assert.ok(merge3.prompt.includes('EXCLUDED from your verdict: ["R1-1"]'))
})

test('deadlock: judge deadlock=true ends the debate UNVERIFIED with the deadlock stamp', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] })],
    judge: [judgeEnv({ deadlock: true, resolutions: [{ gap_id: 'R1-1', resolution: 'unresolved', rationale: 'stuck' }] })],
  }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'UNVERIFIED')
  assert.equal(out.deadlocked, true)
  assert.ok(world.calls.find((c) => c.opts.label.startsWith('assemble')).prompt.includes('judged deadlock'))
})

test('safety ceiling: fresh gaps every round never trigger the judge; ceiling stamps the assembly', async () => {
  let n = 0
  const world = makeWorld((p, o) => {
    if (o.label.startsWith('red-merge')) return redEnv({ gaps: [gap(`R${++n}-1`)] }) // always-new ids
    return makeResponder()(p, o)
  })
  const out = await world.run(script, { ...ARGS, maxRounds: 2 })
  assert.equal(out.verdict, 'UNVERIFIED')
  assert.equal(out.rounds, 2)
  assert.equal(out.deadlocked, false)
  assert.equal(world.calls.filter((c) => c.opts.label.startsWith('judge')).length, 0)
  assert.ok(world.calls.find((c) => c.opts.label.startsWith('assemble')).prompt.includes('safety ceiling'))
})

test('citation passes scale with claim_count and carry the ledger clause', async () => {
  const lensCount = async (claims) => {
    const world = makeWorld(makeResponder({ blueSynth: [blueEnv({ claim_count: claims })], red: [redEnv({ verdict: 'PASS' })] }))
    await world.run(script, ARGS)
    const lenses = world.calls.filter((c) => c.opts.label.startsWith('red-lens'))
    for (const c of lenses.slice(0, lenses.length - 2)) assert.ok(c.prompt.includes('CITATION LEDGER'), 'citation lens carries the ledger clause')
    return lenses.length
  }
  assert.equal(await lensCount(10), 1 + 2)   // floor: one citation pass + logic + risk
  assert.equal(await lensCount(200), 4 + 2)  // cap: four citation passes + logic + risk
})

test('friction aggregates from every seat with attribution', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')], friction: ['no PDF extraction'] }), redEnv({ verdict: 'PASS', friction: [] })],
    blueRespond: [blueEnv({ friction: ['rate-limited on WebFetch'] })],
  }))
  const out = await world.run(script, ARGS)
  assert.ok(out.friction.includes('red-merge-r1: no PDF extraction'))
  assert.ok(out.friction.includes('blue-respond-r1: rate-limited on WebFetch'))
  const assemble = world.calls.find((c) => c.opts.label.startsWith('assemble'))
  assert.ok(assemble.prompt.includes('no PDF extraction'), 'assembly receives the collated friction')
})

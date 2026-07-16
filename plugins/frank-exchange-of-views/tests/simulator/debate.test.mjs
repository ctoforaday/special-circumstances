// node --test — the debate engine's regression suite. Run:
//   node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs (directory form fails on Windows node — list files explicitly)
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { loadDebateScript, makeWorld, makeResponder, blueEnv, redEnv, gap, judgeEnv } from './harness.mjs'

const script = loadDebateScript(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url))
const ARGS = { topic: 'test topic', runDir: 'research/2026-01-01_test', lanes: 3 }
const isJudgmentSeat = (l) => /^(blue-synthesize|red-merge|judge|assemble)/.test(l)

// ---- Founding regressions (runs 1-2) ----

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

test('null blue response aborts cleanly', async () => {
  const world = makeWorld((p, o) => {
    if (o.label.startsWith('blue-respond')) return null
    return makeResponder({ red: [redEnv({ gaps: [gap('R1-1')] })] })(p, o)
  })
  await assert.rejects(world.run(script, ARGS), /blue response round 1 returned null/)
})

// ---- Run-3 docket row 2: the judge call site is guarded ----

test('null judge aborts cleanly instead of TypeError on judge.resolutions', async () => {
  const world = makeWorld((p, o) => {
    if (o.label.startsWith('judge')) return null
    return makeResponder({
      red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] })],
    })(p, o)
  })
  await assert.rejects(world.run(script, ARGS), /judge round 2 returned null/)
})

// ---- Run-3 docket row 2b: citation passes rescale every round ----

test('citationPasses recompute: lens count follows the CURRENT claim_count, round by round', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ claim_count: 10 })],           // round 1: 1 citation pass + 2 -> 3 lenses
    blueRespond: [blueEnv({ claim_count: 200 })],        // report grew: round 2 must rescale
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'more research owed' }] })],
  }))
  await world.run(script, ARGS)
  const lensesByRound = (r) => world.calls.filter((c) => c.opts.label.match(new RegExp(`^red-lens-\\d+-r${r} `))).length
  assert.equal(lensesByRound(1), 3, 'round 1: 1 citation pass + logic + risk')
  assert.equal(lensesByRound(2), 6, 'round 2 rescales to 4 citation passes + logic + risk')
})

// ---- Run-3 docket row 20: degenerate FAIL-with-empty-gaps throws, never loops ----

test('degenerate {FAIL, gaps: []} throws a distinguishing error instead of burning rounds', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'FAIL', gaps: [] })] }))
  await assert.rejects(world.run(script, ARGS), /FAIL with an empty gaps array — degenerate merge/)
})

// ---- Run-3 docket row 23: lineage-following docket ----

test('lineage: a successor gap with supersedes arms the docket even under a fresh id', async () => {
  const world = makeWorld(makeResponder({
    red: [
      redEnv({ gaps: [gap('R1-1')] }),
      redEnv({
        gaps: [gap('R2-1', { supersedes: ['R1-1'] })],
        closures: [{ id: 'R1-1', class: 'closed_with_regression' }],
      }),
      redEnv({ verdict: 'PASS' }),
    ],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R2-1', resolution: 'closed', rationale: 'chain resolved' }] })],
  }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  const judgeCalls = world.calls.filter((c) => c.opts.label.startsWith('judge'))
  assert.equal(judgeCalls.length, 1, 'regression-chain successor must reach the judge')
  assert.ok(judgeCalls[0].prompt.includes('"R2-1"'))
})

test('lineage enforcement: closed_with_regression without a successor naming it throws (R5-5)', async () => {
  const world = makeWorld(makeResponder({
    red: [
      redEnv({ gaps: [gap('R1-1')] }),
      redEnv({
        gaps: [gap('R2-1')], // fresh id, NO supersedes — the silent-lineage-drop shape
        closures: [{ id: 'R1-1', class: 'closed_with_regression' }],
      }),
    ],
  }))
  await assert.rejects(world.run(script, ARGS), /closed gap R1-1 WITH REGRESSION but no successor gap names it/)
})

test('docket window is the whole debate: an id re-raised after skipping a round still arms the judge', async () => {
  const world = makeWorld(makeResponder({
    red: [
      redEnv({ gaps: [gap('R1-1')] }),
      redEnv({ gaps: [gap('R2-1')] }),          // R1-1 absent this round; R2-1 is new
      redEnv({ gaps: [gap('R1-1')] }),          // R1-1 re-raised two rounds later
      redEnv({ verdict: 'PASS' }),
    ],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'risk_accepted', rationale: 'tradeoff' }] })],
  }))
  await world.run(script, ARGS)
  const judgeCalls = world.calls.filter((c) => c.opts.label.startsWith('judge'))
  assert.equal(judgeCalls.length, 1, 'skip-a-round re-raise must still reach the judge')
  assert.ok(judgeCalls[0].prompt.includes('"R1-1"'))
})

// ---- Run-3 docket rows 21 + 24: friction everywhere, persisted in prompts ----

test('blue-synthesize friction reaches the aggregate (row 21)', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ friction: ['write-block on blue/report.md'] })],
    red: [redEnv({ verdict: 'PASS' })],
  }))
  const out = await world.run(script, ARGS)
  assert.ok(out.friction.includes('blue-synthesize: write-block on blue/report.md'))
})

test('every seat prompt carries the friction-to-file clause (row 24); lenses are transcript-forbidden and lens-scoped', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  for (const seat of ['blue-synthesize', 'red-merge-r1', 'blue-respond-r1', 'assemble']) {
    const c = world.calls.find((c) => c.opts.label.startsWith(seat))
    assert.ok(c.prompt.includes('append each entry to') && c.prompt.includes('friction.md'), `${seat} missing friction-persist clause`)
  }
  const lens = world.calls.find((c) => c.opts.label.startsWith('red-lens-1-r1'))
  assert.ok(lens.prompt.includes('MUST NOT write to') && lens.prompt.includes('debate.md'), 'lens must be transcript-forbidden')
  assert.ok(lens.prompt.includes('L1-F1'), 'lens must use lens-scoped finding labels')
})

// ---- Run-3 docket rows 6/7: lane diversity + floor ----

test('lane floor: lanes below 3 refuses dispatch unless an override reason is given', async () => {
  const world = makeWorld(makeResponder())
  await assert.rejects(world.run(script, { ...ARGS, lanes: 1 }), /below the floor of 3/)
  assert.equal(world.calls.length, 0)
  const world2 = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  const out = await world2.run(script, { ...ARGS, lanes: 1, laneFloorOverride: 'smoke run' })
  assert.equal(out.lanes, 1, 'return object must carry lanes for observability')
})

test('lane methods: each lane prompt carries its assigned method lens', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const lanePrompts = world.calls.filter((c) => c.opts.label.startsWith('blue-lane')).map((c) => c.prompt)
  assert.equal(lanePrompts.length, 3)
  assert.ok(lanePrompts[0].includes('adversarial-disconfirming-first'))
  assert.ok(lanePrompts[1].includes('primary-literature'))
  assert.ok(lanePrompts[2].includes('local-repo critical-stance'))
})

// ---- Run-3 docket rows 5/9/22 + 10: provenance, open questions, transcript-first, ledger drift ----

test('synthesis prompt demands minority-claim provenance tagging (row 5, cheap manifest)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const synth = world.calls.find((c) => c.opts.label.startsWith('blue-synthesize'))
  assert.ok(synth.prompt.includes('exactly ONE lane') && synth.prompt.includes('minority'), 'provenance tagging missing from synthesis')
})

test('assembly receives blue open_questions for the template section (row 9)', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ open_questions: ['does the schema guarantee conformance-or-null?'] })],
    red: [redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const asm = world.calls.find((c) => c.opts.label.startsWith('assemble'))
  assert.ok(asm.prompt.includes('Open questions carried past this run'))
  assert.ok(asm.prompt.includes('conformance-or-null'))
})

test('blue response reads transcript RED/LEAD sections first and must propagate corrections everywhere (row 22)', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const resp = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1'))
  assert.ok(resp.prompt.includes('lossy summary'), 'gap JSON must be flagged as lossy')
  assert.ok(resp.prompt.includes('### LEAD'), 'carried-gap rationale delivery missing')
  assert.ok(resp.prompt.includes('propagate every correction to ALL sites'), 'propagation clause missing')
})

test('citation lens ledger clause includes time and access-date drift triggers, not prose-only (row 10)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const lens = world.calls.find((c) => c.opts.label.startsWith('red-lens-1-r1'))
  assert.ok(lens.prompt.includes('more than 2 rounds have elapsed'), 'time trigger missing')
  assert.ok(lens.prompt.includes('access date'), 'access-date trigger missing')
})

// ---- Original behavioral suite (runs 1-2 era, still binding) ----

test('happy path: red PASS on round 1 -> VERIFIED, phases in order, lanes honored', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  const out = await world.run(script, ARGS)
  assert.deepEqual({ verdict: out.verdict, rounds: out.rounds, deadlocked: out.deadlocked, gaps: out.gaps_outstanding },
    { verdict: 'VERIFIED', rounds: 1, deadlocked: false, gaps: 0 })
  assert.deepEqual(world.phases, ['Frontier', 'Blue', 'Assemble'])
  assert.equal(world.calls.filter((c) => c.opts.label.startsWith('blue-lane')).length, 3)
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

test('safety ceiling: fresh unrelated gaps every round never trigger the judge; ceiling stamps the assembly', async () => {
  let n = 0
  const world = makeWorld((p, o) => {
    if (o.label.startsWith('red-merge')) return redEnv({ gaps: [gap(`R${++n}-1`)] }) // always-new ids, no lineage
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

// ---- Efficiency phase (run-4 ratified levers; plans/efficiency-phase.md PR-A) ----

test('telemetry: red-merge prompt carries the board-telemetry append with the pinned v1 mapping', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const merge = world.calls.find(c => c.opts.label.startsWith('red-merge'))
  assert.ok(merge.prompt.includes('trajectories/board-telemetry.jsonl'), 'telemetry sink named')
  assert.ok(merge.prompt.includes('"mapping_version": "v1"'), 'mapping version pinned in the line spec')
  assert.ok(merge.prompt.includes('"realized":0'), 'realized pinned to 0 in the mapping (excluded from mass)')
  assert.ok(merge.prompt.includes('"trivial":0.5'), 'trivial assigned, not left to seat convention')
})

test('batching: merge concatenates lens passes to an absolute scratchpad path, never under the run dir', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const merge = world.calls.find(c => c.opts.label.startsWith('red-merge'))
  assert.ok(merge.prompt.includes('FIRST ACTION'), 'batching is the first action')
  assert.ok(merge.prompt.includes('session scratchpad'), 'output path is the seat scratchpad')
  assert.ok(merge.prompt.includes('never under research/2026-01-01_test'), 'run dir explicitly excluded')
})

test('sharding: ledger+archive replace findings.md in every seat prompt', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  for (const c of world.calls) assert.ok(!c.prompt.includes('red/findings.md'), `findings.md leaked into: ${c.opts.label}`)
  const merge = world.calls.find(c => c.opts.label.startsWith('red-merge'))
  assert.ok(merge.prompt.includes('red/ledger.md') && merge.prompt.includes('red/archive.md'), 'merge maintains both shards')
  assert.ok(merge.prompt.includes('NEAR-MATCH RULE'), 'near-match forces the archive read before a fresh id (§4.5 cond 3)')
  assert.ok(merge.prompt.includes('drift triggers'), 'volatile-source closures inherit drift re-checks (§4.5 cond 4)')
  const judge = world.calls.find(c => c.opts.label.startsWith('judge'))
  assert.ok(judge.prompt.includes('red/ledger.md') && judge.prompt.includes('DEMANDED READS'), 'judge reads ledger + demanded ancestor archive reads')
})

test('spot-check floor: an empty archive_spot_checks from round 2 aborts; round 1 is exempt', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')], archive_spot_checks: [] }),
          redEnv({ gaps: [gap('R1-1')], archive_spot_checks: [] })],
  }))
  await assert.rejects(world.run(script, { ...ARGS, maxRounds: 3 }), /round 2 reported no archive spot-checks/)
})

test('shard counts: closure-index lines != archive blocks is a self-inconsistent self-report and aborts', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ verdict: 'PASS', ledger_closure_lines: 3, archive_blocks: 2 })],
  }))
  await assert.rejects(world.run(script, ARGS), /self-inconsistent shard self-report/)
})

test('dispute routing: an UNADDRESSED dispute auto-dockets (default-to-docket punishes silence)', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R2-1')] })],
    blueRespond: [blueEnv({ grade_disputes: [{ gap_id: 'R1-1', dimension: 'impact', proposed: 'low', evidence: 'e' }] }), blueEnv()],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  const merge2 = world.calls.find(c => c.opts.label.startsWith('red-merge-r2'))
  assert.ok(merge2.prompt.includes("BLUE'S GRADE DISPUTES"), 'red is shown the pending disputes')
  const judge = world.calls.find(c => c.opts.label.startsWith('judge-r2'))
  assert.ok(judge, 'judge dispatched for the unaddressed dispute')
  assert.ok(judge.prompt.includes('grade_dispute_unaddressed'), 'dispute traffic class named')
})

test('dispute routing: an explicit rejection is HELD (no docket) and dockets only on re-dispute', async () => {
  const d = { gap_id: 'R1-1', dimension: 'impact', proposed: 'low', evidence: 'e' }
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }),
          redEnv({ gaps: [gap('R2-1')], dispute_responses: [{ ...d, response: 'rejected', rationale: 'no' }] }),
          redEnv({ gaps: [gap('R3-1')] })],
    blueRespond: [blueEnv({ grade_disputes: [d] }), blueEnv({ grade_disputes: [d] }), blueEnv()],
  }))
  await world.run(script, { ...ARGS, maxRounds: 3 })
  assert.ok(!world.calls.some(c => c.opts.label.startsWith('judge-r2')), 'explicit rejection does NOT docket in its round')
  const judge3 = world.calls.find(c => c.opts.label.startsWith('judge-r3'))
  assert.ok(judge3 && judge3.prompt.includes('grade_dispute_re_raised'), 're-dispute reaches the judge')
})

test('accepted deltas: cumulative magnitude over the threshold batch-dockets for judge review', async () => {
  const d = { gap_id: 'R1-1', dimension: 'impact', proposed: 'certain', evidence: 'e' }
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }),
          redEnv({ gaps: [gap('R1-1', { impact: 'low' })], dispute_responses: [{ gap_id: 'R1-1', dimension: 'impact', response: 'accepted', rationale: 'ok' }] })],
    blueRespond: [blueEnv({ grade_disputes: [d] }), blueEnv()],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  const judge2 = world.calls.find(c => c.opts.label.startsWith('judge-r2'))
  assert.ok(judge2 && judge2.prompt.includes('accepted_delta_overflow'), 'overflow deltas reviewed before they stand')
})

test('carried persistence: a carried gap with unchanged grades never re-dockets; a grade change re-dockets it', async () => {
  const carriedJudge = judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'owe more research' }] })
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] }),
          redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1', { severity: 'high' })] })],
    judge: [carriedJudge, carriedJudge],
  }))
  await world.run(script, { ...ARGS, maxRounds: 4 })
  const judgeRounds = world.calls.filter(c => c.opts.label.startsWith('judge-r')).map(c => c.opts.label)
  assert.ok(judgeRounds.some(l => l.startsWith('judge-r2')), 'first re-raise dockets')
  assert.ok(!judgeRounds.some(l => l.startsWith('judge-r3')), 'carried ruling absorbs the unchanged re-raise (no repeat sitting)')
  assert.ok(judgeRounds.some(l => l.startsWith('judge-r4')), 'a script-visible grade change re-dockets')
})

test('terminal disputes: pending disputes at the ceiling dispatch the terminal judge BEFORE assembly, carried excluded', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] })],
    blueRespond: [blueEnv({ grade_disputes: [{ gap_id: 'R1-1', dimension: 'severity', proposed: 'low', evidence: 'e' }] })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 1 })
  const terminal = world.calls.find(c => c.opts.label.startsWith('judge-terminal'))
  const assemble = world.calls.find(c => c.opts.label.startsWith('assemble'))
  assert.ok(terminal, 'terminal dispute docket fires at the exit boundary')
  assert.ok(terminal.prompt.includes('EXCLUDES carried'), 'no carrying into a round that does not exist')
  assert.ok(terminal.n < assemble.n, 'disposition precedes assembly')
})

test('dispute cap: disputes beyond the per-round cap batch-docket as one overflow item', async () => {
  const many = Array.from({ length: 7 }, (_, i) => ({ gap_id: `R1-${i + 1}`, dimension: 'impact', proposed: 'low', evidence: 'e' }))
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: many.map(d => gap(d.gap_id)) }), redEnv({ gaps: [gap('R2-1')] })],
    blueRespond: [blueEnv({ grade_disputes: many }), blueEnv()],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  const judge2 = world.calls.find(c => c.opts.label.startsWith('judge-r2'))
  assert.ok(judge2 && judge2.prompt.includes('grade_dispute_over_cap'), 'overflow rides the docket as a batch')
})

test('grade_adjusted: a judge grade ruling reaches the next red-merge as an instruction to apply', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R2-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'grade_adjusted', rationale: 'impact is low: evidence X' }] })],
    blueRespond: [blueEnv({ grade_disputes: [{ gap_id: 'R1-1', dimension: 'impact', proposed: 'low', evidence: 'e' }] }), blueEnv()],
  }))
  await world.run(script, { ...ARGS, maxRounds: 3 })
  const merge3 = world.calls.find(c => c.opts.label.startsWith('red-merge-r3'))
  assert.ok(merge3 && merge3.prompt.includes('GRADE ADJUSTMENTS'), 'adjustment applied by the seat that owns the ledger')
})

test('traffic classes: regression successors are marked first-raise; re-raised ids keep the full resolution set', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1'), gap('R1-2')] }),
          redEnv({
            gaps: [gap('R1-2'), gap('R2-1', { supersedes: ['R1-1'] })],
            closures: [{ id: 'R1-1', class: 'closed_with_regression' }],
          })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  const judge2 = world.calls.find(c => c.opts.label.startsWith('judge-r2'))
  assert.ok(judge2.prompt.includes('"first_raise_successor"'), 'successor marked as first-raise traffic')
  assert.ok(judge2.prompt.includes('"re_raised"'), 're-raised id marked with full resolution set')
  assert.ok(judge2.prompt.includes('structurally unavailable'), 'judge told which options are dead for first-raise traffic')
})

test('lane footnote namespaces: every lane prompt assigns a lane-prefixed label convention', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const lanes = world.calls.filter(c => c.opts.label.startsWith('blue-lane'))
  assert.equal(lanes.length, 3)
  for (const [i, c] of lanes.entries()) {
    assert.ok(c.prompt.includes('FOOTNOTE NAMESPACE') && c.prompt.includes(`[^L${i + 1}`), `lane ${i + 1} prefix`)
  }
})

// ---- Coverage-audit gap closures (2026-07-16 audit vs runs 1-4 problem classes) ----

test('lens prompts carry harness notes: windowed full read, Grep counts lines, no heredocs (3 live recurrences)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  for (const c of world.calls.filter(c => c.opts.label.startsWith('red-lens'))) {
    assert.ok(c.prompt.includes('read it whole in consecutive windows'), 'windowed full-read clause missing')
    assert.ok(c.prompt.includes('counts LINES, not occurrences'), 'Grep footgun note missing')
    assert.ok(c.prompt.includes('prefer the Write tool over quoted heredocs'), 'heredoc note missing')
  }
})

test('multi-instance citation slices are assigned footnote-block ownership', async () => {
  const world = makeWorld(makeResponder({ blueSynth: [blueEnv({ claim_count: 200 })], red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const citation = world.calls.filter(c => c.opts.label.startsWith('red-lens')).slice(0, 4)
  assert.equal(citation.length, 4, 'claim_count 200 scales to 4 citation instances')
  for (const c of citation) assert.ok(c.prompt.includes('footnote-block ownership follows the slice'))
})

test('claim_count is echoed to the tracked CHANGELOG at synthesis and every blue response', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  assert.ok(world.calls.find(c => c.opts.label.startsWith('blue-synthesize')).prompt.includes('stating claim_count'))
  assert.ok(world.calls.find(c => c.opts.label.startsWith('blue-respond-r1')).prompt.includes('including claim_count'))
})

test('lanes=5: the full roster deploys and disconfirming-first holds its redundancy-floor second seat', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, { ...ARGS, lanes: 5 })
  const prompts = world.calls.filter(c => c.opts.label.startsWith('blue-lane')).map(c => c.prompt)
  assert.equal(prompts.length, 5)
  assert.ok(prompts[3].includes('practitioner-production'))
  assert.ok(prompts[4].includes('adversarial-disconfirming-first, second seat'), 'redundancy floor seat missing')
})

test('GRADE enum carries compound grades and the pinned mass mapping is total over it (R4-5)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const merge = world.calls.find(c => c.opts.label.startsWith('red-merge'))
  const en = merge.opts.schema.properties.gaps.items.properties.likelihood.enum
  assert.deepEqual(en, ['low', 'low-medium', 'medium', 'medium-high', 'high', 'certain', 'realized', 'trivial'])
  const mapping = JSON.parse(merge.prompt.match(/pinned mapping (\{.*?\})/)[1])
  for (const g of en) assert.ok(g in mapping, `mass mapping not total: ${g} unmapped`)
})

test('shard creator: round-1 merge creates both shards; round 2 updates, never recreates (R4-6 one-creator)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  assert.ok(world.calls.find(c => c.opts.label.startsWith('red-merge-r1')).prompt.includes('ROUND 1: create BOTH files'))
  const m2 = world.calls.find(c => c.opts.label.startsWith('red-merge-r2'))
  assert.ok(m2.prompt.includes('Update the ledger in place') && !m2.prompt.includes('ROUND 1: create BOTH'), 'round 2 must not recreate')
})

test('gap-pattern memory: all three blue seat classes are pointed at the staged inventory (GAP-33)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  for (const seat of ['blue-lane-1', 'blue-synthesize', 'blue-respond-r1']) {
    const c = world.calls.find(x => x.opts.label.startsWith(seat))
    assert.ok(c.prompt.includes('red-gap-patterns.md'), `${seat} missing the gap-pattern pre-flight path`)
  }
})

test('moot: a predicate-expired ruling adjudicates the gap out of red verdict scope (GAP-35, live in run 4)', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'moot', rationale: 'predicate expired — the claim it attached to left the report' }] })],
  }))
  const out = await world.run(script, { ...ARGS, maxRounds: 3 })
  const m3 = world.calls.find(c => c.opts.label.startsWith('red-merge-r3'))
  assert.ok(m3.prompt.includes('EXCLUDED from your verdict: ["R1-1"]'), 'moot gap leaves red scope')
  assert.equal(out.verdict, 'VERIFIED')
})

test('MUST-try observable: graded-down citations require an attempt-or-impossibility line (false-paywall class)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const citation = world.calls.find(c => c.opts.label.startsWith('red-lens-1'))
  assert.ok(citation.prompt.includes('attempt-or-impossibility line'), 'observable missing from ledger clause')
  assert.ok(citation.prompt.includes('an untried "unable to corroborate" is an incomplete audit'))
})

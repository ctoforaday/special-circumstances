// node --test — the debate engine's regression suite. Run:
//   node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs (directory form fails on Windows node — list files explicitly)
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { loadDebateScript, makeWorld, makeResponder, blueEnv, redEnv, gap, judgeEnv, petitionRulingEnv } from './harness.mjs'

const script = loadDebateScript(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url))
// model + judgmentModel are REQUIRED (#111): debate.js throws without both, so the shared ARGS
// carries them and every run/spread inherits them. Tests that exercise the unset-tier THROW build
// their own model-less args explicitly rather than spreading ARGS.
// binDir is REQUIRED now: omitting it used to select a legacy prompt set that told seats to
// hand-write debate.md, red/citation-ledger.md and blue/CHANGELOG.md — files setup no longer
// creates and nothing reads, producing a run that recorded nothing with every gate green.
const ARGS = { topic: 'test topic', runDir: 'research/2026-01-01_test', lanes: 3, model: 'sonnet', judgmentModel: 'sonnet', binDir: '/opt/feov/bin' }
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

// ---- Run-3 docket row 2b + W2i: citation passes rescale every round, on the round's input ----

const lensesByRound = (world, r) => world.calls.filter((c) => c.opts.label.match(new RegExp(`^red-lens-\\d+-r${r} `)))
// A CITATION SEAT IS IDENTIFIED BY A PHRASE IN ITS PROMPT, which is a pattern standing in for a
// schema: reword the clause and every one of these tests reports zero citation seats, which reads
// exactly like a dispatcher that stopped dispatching them. It broke once already, when the clause
// was retitled for the evidence view. The field that should carry this is the lens ROLE — 1-4 are
// the citation slices, 5 logic, 6 risk, and the role is already in the label — so the marker lives
// here in ONE place until the tests key on the number instead.
const CITATION_CLAUSE = 'HOW YOU RESOLVE AN ANCHOR'
const citationSeats = (world, r) => lensesByRound(world, r).filter((c) => c.prompt.includes(CITATION_CLAUSE))

test('citationPasses recompute: round 1 sizes on the corpus, round 2 on the DELTA (W2i)', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ claim_count: 10 })],           // round 1: 1 citation pass + 2 -> 3 lenses
    blueRespond: [blueEnv({ claim_count: 200 })],        // +190 claims: a big delta, but capped at 2
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'more research owed' }] })],
  }))
  await world.run(script, ARGS)
  assert.equal(lensesByRound(world, 1).length, 3, 'round 1: 1 citation pass + logic + risk')
  assert.equal(lensesByRound(world, 2).length, 4, 'round 2: delta of 190 rescales to the cap of 2 citation + logic + risk')
})

test('W2i: a small round-2 delta drops to ONE citation seat — sized to input, not halved by rule', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ claim_count: 200 })],          // round 1: the cap, 4 citation passes
    blueRespond: [blueEnv({ claim_count: 210 })],        // +10 claims: one seat's worth of new surface
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'more research owed' }] })],
  }))
  await world.run(script, ARGS)
  assert.equal(citationSeats(world, 1).length, 4, 'round 1 coverage is untouched by W2i')
  assert.equal(citationSeats(world, 2).length, 1, 'round 2 sizes on the 10-claim delta, not the 210-claim corpus')
  assert.equal(lensesByRound(world, 2).length, 3, 'L5 + L6 dispatch every round regardless')
})

test('W2i: round 3 restores both citation seats — the staleness sweep is O(corpus), not O(delta)', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ claim_count: 200 })],
    blueRespond: [blueEnv({ claim_count: 202 }), blueEnv({ claim_count: 203 })], // near-zero deltas throughout
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R2-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [
      judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'more research owed' }] }),
      judgeEnv({ resolutions: [{ gap_id: 'R2-1', resolution: 'carried', rationale: 'still owed' }] }),
    ],
  }))
  await world.run(script, ARGS)
  assert.equal(citationSeats(world, 2).length, 1, 'round 2: delta of 2 buys one seat')
  assert.equal(citationSeats(world, 3).length, 2, 'round 3: the >2-rounds staleness trigger re-staffs the second seat')
})

test('W2i: lens numbers are ROLES — logic is always L5, dark-side always L6, whatever the citation count', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ claim_count: 10 })],           // ONE citation pass: positionally L5/L6 would slide to L2/L3
    red: [redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const labels = lensesByRound(world, 1).map((c) => c.opts.label.match(/^red-lens-(\d+)-/)[1])
  assert.deepEqual(labels.sort(), ['1', '5', '6'], 'citation slice L1, logic L5, dark-side L6 — no positional slide')
  const logic = lensesByRound(world, 1).find((c) => c.opts.label.startsWith('red-lens-5-'))
  assert.ok(logic.prompt.includes('logic and completeness'), 'L5 is the logic seat')
  assert.ok(logic.prompt.includes('L5-F'), 'the tool-assigned finding label carries the ROLE number (L5-F{N})')
  assert.ok(logic.prompt.includes('lens number 5'), 'the prompt names this lens its role (5)')
})

test('W2i: the consolidated-citation duty binds rounds 2+ citation seats only, and keeps coverage observable', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ claim_count: 200 })],
    blueRespond: [blueEnv({ claim_count: 210 })],
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'more research owed' }] })],
  }))
  await world.run(script, ARGS)
  for (const c of citationSeats(world, 1)) assert.ok(!c.prompt.includes('CONSOLIDATED CITATION SEAT'), 'round 1 is not consolidated')
  const r2 = citationSeats(world, 2)[0]
  assert.ok(r2.prompt.includes('CONSOLIDATED CITATION SEAT'), 'round 2 citation seat carries the consolidated duty')
  assert.ok(r2.prompt.includes('SPOT-CHECK'), 'the duty keeps a spot-check of already-verified pairs')
  assert.ok(r2.prompt.includes('COVERAGE IS AN OBSERVABLE'), 'what went unexamined must be stated, not assumed')
  const dark = lensesByRound(world, 2).find((c) => c.opts.label.startsWith('red-lens-6-'))
  assert.ok(!dark.prompt.includes('CONSOLIDATED CITATION SEAT'), 'L6 is untouched by citation consolidation')
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

// LINEAGE: the guard reads the ENVELOPE, which is a lossy summary of the record — so it reports
// and continues rather than killing the run. Measured 2026-08-04: red closed R1-1 with regression
// and correctly minted R2-1 with --supersedes R1-1 ON THE RECORD (the board showed
// `R2-1 supersedes -> [R1-1]`), then omitted the field from its envelope. The old hard throw
// discarded a 12-agent, 723k-token run whose lineage was entirely intact. `verify`'s
// supersedes-resolve checks lineage where the truth lives; this clause only stops the engine
// trusting the summary over the source.
test('lineage: an envelope missing supersedes REPORTS and continues — the record is authoritative', async () => {
  const world = makeWorld(makeResponder({
    red: [
      redEnv({ gaps: [gap('R1-1')] }),
      redEnv({
        gaps: [gap('R2-1')], // fresh id, NO supersedes in the ENVELOPE — the lossy-report shape
        closures: [{ id: 'R1-1', class: 'closed_with_regression' }],
      }),
      redEnv({ verdict: 'PASS' }),
    ],
  }))
  const out = await world.run(script, ARGS) // MUST NOT throw — that is the fix
  assert.ok(
    out.friction.some((f) => /closed R1-1 WITH REGRESSION and no successor in the ENVELOPE/.test(f)),
    'the reporting gap is logged as friction so it is still visible',
  )
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

test('every seat prompt carries the friction clause (envelope + verb, not a hand-written file); lenses are transcript-forbidden and record findings via the tool', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  // THE LENS IS IN THIS LIST NOW. It was the one seat class the clause was never appended to —
  // found 2026-08-13 when the orphan gate reported `lens friction` as named nowhere a seat reads.
  // Four lens seats per round were told to close a channel nobody had told them about.
  for (const seat of ['blue-synthesize', 'red-merge-r1', 'blue-respond-r1', 'assemble', 'red-lens-1-r1']) {
    const c = world.calls.find((c) => c.opts.label.startsWith(seat))
    // The COMMAND, not the phrase "friction verb": the clause names `<role> friction` now, so a
    // seat can run what it is told to run. Asserting the old prose would pass over a clause that
    // never shows the invocation — which is exactly how these verbs went uninstructed.
    assert.ok(/\b(lens|merge|blue|bench) friction --none\b/.test(c.prompt) && c.prompt.includes("envelope's friction field"),
      `${seat} missing friction-report clause (envelope + the role-qualified command)`)
  }
  const lens = world.calls.find((c) => c.opts.label.startsWith('red-lens-1-r1'))
  // Transcript-forbidden, stated POSITIVELY: the clause used to prohibit writing to debate.md,
  // a file setup no longer creates. The contract it was protecting is who owns the round's
  // narrative — one seat, through one verb — and that is the form a lens can actually act on.
  assert.ok(/Only red-merge records the round's "### RED" narrative, via a position event/.test(lens.prompt),
    'a lens is not a debate party — the round narrative is the merge\'s, via the position verb')
  // The COMMAND again, not the phrase "finding verb" — same reason as the friction clause above.
  assert.ok(/lens finding --key /.test(lens.prompt) && lens.prompt.includes('L1-F'),
    'lens records findings via the verb with a tool-assigned role-scoped label (L1-F{N})')
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

test('blue authors ## Open questions inside the audited report; assembly lifts it, never receives it as input', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ open_questions: ['does the schema guarantee conformance-or-null?'] })],
    red: [redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const synth = world.calls.find((c) => c.opts.label.startsWith('blue-synthesize'))
  assert.ok(synth.prompt.includes('## Open questions'), 'blue authors the open-questions section inside the audited report')
  const asm = world.calls.find((c) => c.opts.label.startsWith('assemble'))
  assert.ok(!asm.prompt.includes('open_questions'), 'assembly does not receive open_questions as an input — it lifts blue’s audited section verbatim')
})

test('blue response reads transcript RED/LEAD sections first and must propagate corrections everywhere (row 22)', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const resp = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1'))
  assert.ok(resp.prompt.includes('lossy summary'), 'gap JSON must be flagged as lossy')
  assert.ok(resp.prompt.includes('### LEAD'), 'carried-gap rationale delivery missing')
  assert.ok(/[Pp]ropagate every correction to ALL sites/.test(resp.prompt), 'propagation clause missing')
})

test('additive is a CLAIMS invariant, not a prose one: compaction is free, removal is recorded', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const resp = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1')).prompt
  const synth = world.calls.find((c) => c.opts.label.startsWith('blue-synthesize')).prompt

  // The old rule forbade rewriting to prevent silent deletion, and paid for it in
  // a report that could only grow. Both halves must now be stated at both seats.
  assert.ok(/compact and reorganize prose/i.test(resp), 'respond may compact')
  assert.ok(/reorganize freely/i.test(synth), 'synthesis may reorganize')
  assert.ok(/retire verb/.test(resp) && /retire verb/.test(synth), 'both seats name the one exit for substance')
  assert.ok(/claim_count fall against the retire events|unaccounted drop/.test(resp), 'the seat is told the arithmetic is checked')
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

test('per-role models: bulk seats get `model`, judgment seats get `judgmentModel`; unset either throws', async () => {
  // Both tiers set and DISTINCT, so each seat class is checked against its own flag — judgment
  // seats now carry `judgmentModel` (never undefined; the pre-#111 inheritance is gone).
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, { ...ARGS, model: 'haiku', judgmentModel: 'opus' })
  for (const c of world.calls) assert.equal(c.opts.model, isJudgmentSeat(c.opts.label) ? 'opus' : 'haiku', `${c.opts.label} takes its class tier`)

  // The engine never guesses or inherits a tier: a missing tier refuses dispatch, naming the flag.
  // Build args WITHOUT spreading ARGS (which carries both) so the omission is real.
  const base = { topic: 'test topic', runDir: 'research/2026-01-01_test', lanes: 3 }
  await assert.rejects(makeWorld(makeResponder()).run(script, { ...base, judgmentModel: 'sonnet' }), /refusing dispatch — model unset/)
  await assert.rejects(makeWorld(makeResponder()).run(script, { ...base, model: 'sonnet' }), /refusing dispatch — judgmentModel unset/)
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
  // CEILING, not UNVERIFIED (rulebook audit item 7). The protocol named two
  // terminators and this is the third: the run did not fail to converge, it ran
  // out of budget mid-flight. Run 5 ended exactly here — the bench had ruled
  // deadlock FALSE and the final blue revision was never audited — and the
  // taxonomy had no way to say so, which is how the obligation left the run by
  // hand instead of on the record.
  assert.equal(out.verdict, 'CEILING')
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
    for (const c of lenses.slice(0, lenses.length - 2)) assert.ok(c.prompt.includes(CITATION_CLAUSE), 'citation lens carries the ledger clause')
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

test('telemetry: red-merge is told the line is TOOL-COMPUTED, not hand-written', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const merge = world.calls.find(c => c.opts.label.startsWith('red-merge'))
  assert.ok(merge.prompt.includes('COMPUTED BY THE TOOL') && merge.prompt.includes('you do NOT hand-write it'), 'telemetry is tool-computed, not appended by the seat')
  assert.ok(!merge.prompt.includes('feov-record merge render'), 'the manual render step is retired — telemetry derives just-in-time on read, not from a seat-run render')
  assert.ok(!merge.prompt.includes('trajectories/board-telemetry.jsonl'), 'the seat no longer hand-writes a telemetry sink — the self-report is retired')
})

test('the merge reads this round\'s findings from the record view, not a candidate-file cat', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const merge = world.calls.find(c => c.opts.label.startsWith('red-merge'))
  assert.ok(merge.prompt.includes('FIRST ACTION'), 'reading findings is the first action')
  assert.ok(merge.prompt.includes('show findings'), 'the merge reads the findings view from the record')
  assert.ok(!merge.prompt.includes('red/candidates'), 'the candidate-file cat is retired')
})

test('the board is the tool: merge mints through feov-record, downstream seats pull via show, no findings.md', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  for (const c of world.calls) assert.ok(!c.prompt.includes('red/findings.md'), `findings.md leaked into: ${c.opts.label}`)
  const merge = world.calls.find(c => c.opts.label.startsWith('red-merge'))
  assert.ok(merge.prompt.includes('MINT each open gap') && merge.prompt.includes('feov-record merge'), 'merge writes the board through the tool, not hand-written markdown')
  assert.ok(merge.prompt.includes('NEAR-MATCH RULE'), 'near-match forces the archive read before a fresh id (§4.5 cond 3)')
  assert.ok(merge.prompt.includes('drift triggers'), 'volatile-source closures inherit drift re-checks (§4.5 cond 4)')
  const judge = world.calls.find(c => c.opts.label.startsWith('judge'))
  // `show` became a GROUP: `show ledger`, not `show --view ledger`. A flag's VALUE space has no
  // --help and no completion of its own, which is the undiscoverability the motion collapse fixed
  // one layer up. The assertion is on the READ still being pulled through the tool, which is the
  // property that matters — the spelling changed, the contract did not.
  // Two tokens, not one phrase: the projection now comes FIRST (`show board --run <dir> --format
  // markdown`) so the group's subcommand is adjacent to `show`, which is what makes a stale name
  // catchable. Asserting the contiguous string would have been asserting the argument ORDER.
  assert.ok(judge.prompt.includes('show board') && judge.prompt.includes('--format markdown') && judge.prompt.includes('DEMANDED READS'), 'judge ACTIVELY PULLS the board through the tool (no materialized-path read); ledger and archive collapsed into board --format markdown')
})

test('spot-check floor: an empty archive_spot_checks from round 2 aborts; round 1 is exempt', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')], archive_spot_checks: [], ledger_closure_lines: 1, archive_blocks: 1 }),
          redEnv({ gaps: [gap('R1-1')], archive_spot_checks: [] })],
  }))
  await assert.doesNotReject(world.run(script, { ...ARGS, maxRounds: 3 }), 'RETIRED: the spot-check floor trusted red-merge self-report; the tool board is authoritative now')
})

test('shard counts: closure-index lines != archive blocks is a self-inconsistent self-report and aborts', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ verdict: 'PASS', ledger_closure_lines: 3, archive_blocks: 2 })],
  }))
  await assert.doesNotReject(world.run(script, ARGS), 'RETIRED: the count gate compared two self-reported numbers; the tool board is authoritative now')
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

test('closing arguments: judge sits AFTER blue, both sides file closings, ruling basis confined', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1'), gap('R1-2')] }),
          redEnv({
            gaps: [gap('R1-2'), gap('R2-1', { supersedes: ['R1-1'] })],
            closures: [{ id: 'R1-1', class: 'closed_with_regression' }],
          })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  const judge2 = world.calls.find(c => c.opts.label.startsWith('judge-r2'))
  const blue2 = world.calls.find(c => c.opts.label.startsWith('blue-respond-r2'))
  const merge2 = world.calls.find(c => c.opts.label.startsWith('red-merge-r2'))
  assert.ok(judge2 && blue2, 'both seats ran in round 2')
  assert.ok(judge2.n > blue2.n, 'the judge rules AFTER blue has answered — never on unanswered material')
  assert.ok(merge2.prompt.includes('### RED CLOSING'), 'red files its closing at the merge')
  assert.ok(blue2.prompt.includes('### BLUE CLOSING'), 'blue files its closing with its response')
  assert.ok(blue2.prompt.includes('DOCKETED for adjudication AFTER your response'), 'blue told the docket before the judge sits')
  assert.ok(judge2.prompt.includes('RULING BASIS IS CONFINED TO'), 'ruling basis: closings + transcript + final state')
  assert.ok(judge2.prompt.includes('counts AGAINST the side that made it'), 'overstatement penalty stated to the judge')
  assert.ok(!judge2.prompt.includes('structurally unavailable'), 'dead-options patch removed — full resolution set for all classes')
})
test('lanes record source notes as prose and do NOT mint footnote labels (citations are tool-managed at synthesis)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const lanes = world.calls.filter(c => c.opts.label.startsWith('blue-lane'))
  assert.equal(lanes.length, 3)
  for (const [i, c] of lanes.entries()) {
    assert.ok(c.prompt.includes('SOURCE NOTES') && c.prompt.includes('Do NOT mint footnote labels'), `lane ${i + 1} source-note convention`)
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

test('multi-instance citation slices are assigned citation ownership', async () => {
  const world = makeWorld(makeResponder({ blueSynth: [blueEnv({ claim_count: 200 })], red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const citation = world.calls.filter(c => c.opts.label.startsWith('red-lens')).slice(0, 4)
  assert.equal(citation.length, 4, 'claim_count 200 scales to 4 citation instances')
  for (const c of citation) assert.ok(c.prompt.includes('citation ownership follows the slice'))
})

// claim_count reaches the ROUND RECORD at synthesis and every blue response.
//
// It used to say "the tracked CHANGELOG", and the assertion still passes for the right reason:
// the number belongs in whatever carries the round record. That file is retired (#251) and the
// carrier is the `revision` event — which capture already counted, while the prompts demanded
// the file. Two channels for one fact, and the audit read the one nobody was told to write.
test('claim_count reaches the round record at synthesis and every blue response', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const synth = world.calls.find(c => c.opts.label.startsWith('blue-synthesize')).prompt
  const respond = world.calls.find(c => c.opts.label.startsWith('blue-respond-r1')).prompt
  assert.ok(synth.includes('claim_count') && synth.includes('blue revision'), 'synthesis records the round with claim_count')
  assert.ok(respond.includes('claim_count') && respond.includes('blue revision'), 'each response records the round with claim_count')
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
  // The mass mapping no longer rides in the prompt — the tool computes telemetry, and the
  // Go MASS table (record.go) is total over this enum, asserted by the record package tests.
  assert.ok(!/pinned mapping \{/.test(merge.prompt), 'the MASS json no longer bloats the merge prompt (retired with the hand-written telemetry line)')
})

test('shard creator: round-1 merge creates both shards; round 2 updates, never recreates (R4-6 one-creator)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  assert.ok(world.calls.find(c => c.opts.label.startsWith('red-merge-r1')).prompt.includes('MINT each open gap'), 'round-1 merge mints through the tool, does not hand-write shards')
  const m2 = world.calls.find(c => c.opts.label.startsWith('red-merge-r2'))
  assert.ok(m2.prompt.includes('MINT each open gap') && !m2.prompt.includes('create BOTH'), 'round 2 also mints through the tool, never hand-writes')
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

test('speed doctrine: every seat prompt carries the batching + native-peek clause (run-4 forensics)', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  for (const c of world.calls) {
    assert.ok(c.prompt.includes('SPEED:'), `seat missing speed clause: ${c.opts.label}`)
    assert.ok(c.prompt.includes('batch INDEPENDENT tool calls'), `batching directive missing: ${c.opts.label}`)
  }
})

test('heartbeats: the panel narrator logs round start, red verdict + mass, docket size, and blue response', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  const all = world.logs.join('\n')
  assert.ok(/round 1: dispatching \d+ red lenses/.test(all), 'round-start heartbeat')
  assert.ok(/round 1: red FAIL — 1 gaps open, mass 4\.0/.test(all), 'verdict heartbeat carries live mass')
  assert.ok(/round 2: docket — 1 contested item/.test(all), 'docket heartbeat')
  assert.ok(/round 1: blue responded — 1 gaps addressed/.test(all), 'blue heartbeat')
})

// ---- Wave 1 engine plumbing (W1.6-W1.12) ----

test('W1.6: pinned claim unit reaches synthesis and response prompts', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  for (const seat of ['blue-synthesize', 'blue-respond-r1']) {
    const c = world.calls.find((c) => c.opts.label.startsWith(seat))
    // `CITED` was the LEGACY arm's wording. The record arm names the command that computes the
    // number — which is the point of pinning it: two honest merges differed 2x by hand.
    assert.ok(c.prompt.includes('CLAIM UNIT') && c.prompt.includes('count-claims'), `${seat} missing pinned claim unit`)
  }
})

// W1.7 + #249: the attestation duty stays; the CONSEQUENCE is recovery, not a dead run. Two
// consecutive haiku validation runs died on this guard after 16-22 completed agents, so a missed
// round record must degrade, never discard.
test('W1.7/#249: blue-respond without the attestation is RE-PROMPTED, then continues with friction', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    blueRespond: [blueEnv({ round_record_appended: false })], // the retry re-serves the same false envelope
  }))
  const out = await world.run(script, ARGS) // MUST NOT throw — that is the whole fix
  assert.ok(world.calls.some((c) => c.opts.label.startsWith('blue-respond-r1-round-record')), 'the seat was re-prompted for its round record')
  assert.ok(out.friction.some((f) => /round-parity.*UNRESOLVED/.test(f)), 'the unresolved parity gap is logged as friction')
})

test('W1.7/#249: blue-synthesize without the Round 0 attestation continues — red still audits', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ round_record_appended: false })],
    red: [redEnv({ verdict: 'PASS' })],
  }))
  const out = await world.run(script, ARGS)
  assert.ok(world.calls.some((c) => c.opts.label.startsWith('blue-synthesize-round-record')), 'synthesize was re-prompted')
  assert.ok(world.calls.some((c) => c.opts.label.startsWith('red-')), 'red seats still dispatch — a bookkeeping miss no longer discards the run')
  assert.ok(out.friction.some((f) => /round-parity.*UNRESOLVED/.test(f)), 'the gap is scored as friction')
})

test('W1.7/#249: a seat that attests ON THE RETRY continues with NO friction', async () => {
  const world = makeWorld(makeResponder({
    // first envelope misses the attestation; the re-prompt returns an attested one
    blueSynth: [blueEnv({ round_record_appended: false }), blueEnv({ round_record_appended: true })],
    red: [redEnv({ verdict: 'PASS' })],
  }))
  const out = await world.run(script, ARGS)
  assert.ok(world.calls.some((c) => c.opts.label.startsWith('blue-synthesize-round-record')), 'the retry happened')
  assert.ok(!out.friction.some((f) => /round-parity/.test(f)), 'a recovered attestation logs no parity friction')
})

test('W1.8: empty spot-checks are EXEMPT when the archive entered the round with zero records (round 1 closed nothing)', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')], archive_spot_checks: [], ledger_closure_lines: 0, archive_blocks: 0 }),
          redEnv({ verdict: 'PASS', archive_spot_checks: [], ledger_closure_lines: 0, archive_blocks: 0 })],
  }))
  const result = await world.run(script, { ...ARGS, maxRounds: 3 })
  assert.equal(result.verdict, 'VERIFIED', 'run completes — the spot-check floor gate is RETIRED, so empty spot-checks never abort')
})

test('W1.9: routed_to_infrastructure leaves red verdict pool and ships as a named infra debt', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }),
          redEnv({ gaps: [gap('R1-1')] }), // re-raise -> dockets
          redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'routed_to_infrastructure', rationale: 'pin validation is setup tooling, owed by the lead' }] })],
  }))
  const result = await world.run(script, { ...ARGS, maxRounds: 4 })
  assert.equal(result.infra_debts.length, 1)
  assert.equal(result.infra_debts[0].gap_id, 'R1-1')
  assert.ok(result.infra_debts[0].owed_fix.includes('setup tooling'))
  const judgePrompt = world.calls.find((c) => c.opts.label.startsWith('judge-r'))
  assert.ok(judgePrompt.prompt.includes('routed_to_infrastructure'), 'the disposition is offered in the resolution set')
  const assemble = world.calls.find((c) => c.opts.label.startsWith('assemble'))
  assert.ok(assemble.prompt.includes('opinions and petition rulings'), 'infra debts ship as routed_to_infrastructure opinions the tool composes from the debate record — not hand-carried inputs')
})

test('W1.10-W1.12: probe classes, sanctioned Glob/Grep fallback, respond workset batching, large-source rule', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const merge = world.calls.find((c) => c.opts.label.startsWith('red-merge-r1'))
  const respond = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1'))
  const lens = world.calls.find((c) => c.opts.label.startsWith('red-lens'))
  assert.ok(merge.prompt.includes('PROBE CLASSES') && merge.prompt.includes('DOCUMENT-PROBE'), 'merge demands classed probes (W1.10)')
  assert.ok(respond.prompt.includes('PROBE CLASSES') && respond.prompt.includes('deferred acceptance test'), 'respond discharges by class (W1.10)')
  assert.ok(lens.prompt.includes('KNOWN HARNESS LIMIT') && lens.prompt.includes('SANCTIONED fallback'), 'Glob/Grep fallback sanctioned everywhere via speedClause (W1.11)')
  assert.ok(respond.prompt.includes('respond-1-workset'), 'respond FIRST ACTION batches its working set (W1.12)')
  const citLens = world.calls.filter((c) => c.opts.label.startsWith('red-lens')).find((c) => c.prompt.includes(CITATION_CLAUSE))
  assert.ok(citLens && citLens.prompt.includes('VERBATIM READS ONLY') && citLens.prompt.includes('--comments'), 'citation lens carries the verbatim-reads rule (no WebFetch, curl/gh only)')
})

// ---- W2b engine mechanics ----

test('W2b: empty manifest on a repair round aborts; the harness default satisfies the floor', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] })],
    blueRespond: [blueEnv({ manifest: [] })],
  }))
  await assert.rejects(world.run(script, ARGS), /EMPTY correctness manifest/)
})

test('W2b: manifest coverage gap is LOGGED (scored at capture), never fatal', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1'), gap('R1-2')] }), redEnv({ verdict: 'PASS' })],
  }))
  const result = await world.run(script, ARGS)
  assert.equal(result.verdict, 'VERIFIED', 'partial coverage does not kill the round')
  assert.ok(world.logs.some((l) => l.includes('manifest coverage 1/2') && l.includes('R1-2')), 'the uncovered gap is named')
})

test('W2b: script-side scorecard — repair_regression ratio and edge deltas logged from envelopes', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1', { likelihood: 'high', impact: 'high' })] }),
          redEnv({ gaps: [gap('R2-1', { supersedes: ['R1-1'], likelihood: 'medium', impact: 'medium' })] }),
          redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 4 })
  const line = world.logs.find((l) => l.includes('scorecard'))
  assert.ok(line, 'scorecard line emitted')
  assert.ok(line.includes('repair_regression 1/1 = 1.00'), `ratio computed: ${line}`)
  assert.ok(line.includes('down 5.0'), `edge delta down 9-4=5 mass: ${line}`)
})

test('W2b: convergence-vs-verdict detector fires on converged fallout-only FAIL, visibility only', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }),
          redEnv({ gaps: [gap('R2-1', { supersedes: ['R1-1'], severity: 'low', likelihood: 'low', impact: 'low' })] }),
          redEnv({ verdict: 'PASS' })],
  }))
  const result = await world.run(script, { ...ARGS, maxRounds: 4 })
  assert.equal(result.verdict, 'VERIFIED', 'detector never blocks — the run continued to PASS')
  assert.ok(world.logs.some((l) => l.includes('convergence-vs-verdict divergence')), 'detector line emitted')
})

test('W2b: telemetry spec and gap records carry the new fields in the merge prompt', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const merge = world.calls.find((c) => c.opts.label.startsWith('red-merge-r1'))
  assert.ok(merge.prompt.includes('repair_regression') && merge.prompt.includes('edge_deltas'), 'telemetry line spec extended')
  assert.ok(merge.prompt.includes('acceptance_check'), 'acceptance-check-at-mint demanded')
  const respond = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1'))
  assert.ok(respond.prompt.includes('THE MANIFEST') && respond.prompt.includes('manifest array'), 'manifest demanded at respond')
})

// THE RECORD CONTRACT IS UNCONDITIONAL. This was a DUAL-MODE test: binDir armed the record
// clause and omitting it 'left prompts untouched' — a legacy set that told seats to hand-write
// debate.md, red/citation-ledger.md and blue/CHANGELOG.md. setup stopped creating those files,
// so that mode produced a run recording nothing with every gate green. binDir is required now.
test('the record contract arms SEAT_ID and the binary path on every seat', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, { ...ARGS, binDir: '/plug/bin' })
  const seats = ['blue-synthesize', 'red-lens', 'red-merge-r1', 'blue-respond-r1', 'assemble']
  for (const s of seats) {
    const c = world.calls.find((x) => x.opts.label.startsWith(s))
    assert.ok(c.prompt.includes('SEAT_ID:') && c.prompt.includes('/plug/bin/feov-record'), `${s} missing record clause`)
  }
  // Each seat is handed its ROLE, not a script path: seat identity is bound to
  // the role namespace, so a seat pointed at the wrong one is refused by the tool.
  const roleOf = (label) => world.calls.find((x) => x.opts.label.startsWith(label)).prompt
  // The path is QUOTED (plugin roots can contain spaces), so the role token sits
  // after the closing quote — the same shape capture's join audit must match.
  const dispatches = (label, role) => new RegExp(`feov-record"?\\s+${role}\\s+register`).test(roleOf(label))
  assert.ok(dispatches('red-lens', 'lens'), 'lens dispatched to the lens role')
  assert.ok(dispatches('red-merge-r1', 'merge'), 'merge dispatched to the merge role')
  assert.ok(dispatches('blue-respond-r1', 'blue'), 'blue dispatched to the blue role')
  assert.ok(dispatches('assemble', 'bench'), 'assembly dispatched to the bench role')

  // THE HELP IS THE CONTRACT, and the seat is told so explicitly — including what
  // to do when what it needs is absent, which is friction rather than improvising
  // or hand-writing the artifact the record layer exists to replace.
  const lens = roleOf('red-lens')
  assert.ok(lens.includes('feov-record lens --help'), 'the seat is pointed at its own contract')
  assert.ok(/does not exist for you/.test(lens), 'absence is stated as absence')
  assert.ok(/friction verb/.test(lens), 'the escalation path is named')
})

// ---- W2c: the petition short-circuit + judicial halt ----

test('W2c: a petition dispatches a bench sitting BEFORE the next seat; denied -> the debate continues to its normal end', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ petitions: [{ class: 'ethical', basis: 'asked to assert what I believe false', relief: 'reframe the claim' }] })],
    red: [redEnv({ verdict: 'PASS' })],
    petition: [petitionRulingEnv()],
  }))
  const result = await world.run(script, ARGS)
  assert.equal(result.verdict, 'VERIFIED', 'denied petition never blocks the run')
  const idx = world.calls.findIndex((c) => c.opts.label.startsWith('judge-petition'))
  const redIdx = world.calls.findIndex((c) => c.opts.label.startsWith('red-lens'))
  assert.ok(idx >= 0 && idx < redIdx, 'the sitting fired before round 1 dispatched')
  assert.equal(result.petitions.length, 1)
  assert.equal(result.petitions[0].ruling, 'denied')
  assert.ok(world.calls[idx].prompt.includes('never sanctioned'), 'anti-chilling stated at the sitting')
})

test('W2c: a HALT ruling ends the run — verdict HALTED, opinion carried verbatim to assembly and the envelope', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')], petitions: [{ class: 'safety', basis: 'the topic requires publishing an exploit chain', relief: 'halt the run' }] })],
    // #329: a halt is NOT a ruling value. The petition is ruled on its merits, and the halt
    // arrives on its own channel — because it is recorded through `bench halt`, whose opinion
    // reaches the human verbatim. While `halt` sat in the ruling enum, the judge recorded it with
    // `petition-rule`, the record REFUSED the write, and the engine halted off the envelope with
    // no halt event anywhere: the report never said the bench halted.
    petition: [petitionRulingEnv({
      rulings: [{ petitioner: 'red-merge-r1', class: 'safety', ruling: 'granted', relief: 'the objection is sound' }],
      halt: { opinion: 'continuing would compromise safety; the human must decide' },
    })],
  }))
  const result = await world.run(script, { ...ARGS, maxRounds: 5 })
  assert.equal(result.verdict, 'HALTED')
  assert.equal(result.halted, true)
  assert.ok(result.halt_opinion.includes('the human must decide'))
  // THE SEAT MUST BE TOLD WHICH VERB RECORDS IT. The engine stopping is not the same as the halt
  // being on the record, and that gap is exactly what #329 was.
  const sitting = world.calls.find((c) => c.opts.label.startsWith('judge-petition'))
  assert.ok(!/rule granted[^.]*\| halt/.test(sitting.prompt), 'the prompt must not offer halt as a petition ruling')
})

// #329 IN RECORD MODE, which is what every real run uses: the sitting must name the verb that
// actually records a halt. The engine stopping is not the same as the halt being on the record —
// the prompt used to say "each ruling is recorded via the petition-rule verb" while the record's
// enum refused `halt`, so the run ended with no halt event and the report never said so.
test('W2c: the petition sitting names `bench halt` as the halt channel, not petition-rule (#329)', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')], petitions: [{ class: 'safety', basis: 'b', relief: 'halt the run' }] })],
    petition: [petitionRulingEnv({
      rulings: [{ petitioner: 'red-merge-r1', class: 'safety', ruling: 'granted', relief: 'sound' }],
      halt: { opinion: 'the human must decide' },
    })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 5, binDir: '/bin' })
  const sitting = world.calls.find((c) => c.opts.label.startsWith('judge-petition'))
  assert.ok(sitting.prompt.includes('bench halt'), 'the sitting prompt names the halt verb')
  assert.ok(sitting.prompt.includes('VERBATIM'), 'and says the opinion reaches the human verbatim')
  assert.ok(!world.calls.some((c) => c.opts.label.startsWith('blue-respond')), 'the round never continued past the halt')
  const assemble = world.calls.find((c) => c.opts.label.startsWith('assemble'))
  assert.ok(assemble.prompt.includes('--as HALTED') && assemble.prompt.includes('terminal halt'), 'the halt is recorded via bench outcome (verdict HALTED) and the tool composes the terminal halt from the record')
})

test('W2c: no petitions -> no sitting (zero cost); granted relief binds subsequent seats', async () => {
  const quiet = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await quiet.run(script, ARGS)
  assert.ok(!quiet.calls.some((c) => c.opts.label.startsWith('judge-petition')), 'no petition, no sitting')
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ petitions: [{ class: 'constitutional', basis: 'b', relief: 'narrow the demanded scope' }] })],
    petition: [petitionRulingEnv({ rulings: [{ petitioner: 'blue-synthesize', class: 'constitutional', ruling: 'granted', relief: 'scope narrowed to the shipped artifacts' }] })],
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  const result = await world.run(script, ARGS)
  assert.equal(result.verdict, 'VERIFIED')
  const respond = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1'))
  assert.ok(respond.prompt.includes('BENCH RELIEF IN EFFECT') && respond.prompt.includes('scope narrowed'), 'granted relief surfaced to the party')
  assert.ok(world.calls.find((c) => c.opts.label.startsWith('blue-synthesize')).prompt.includes('PETITIONS:'), 'the right is stated at the seat')
})

test('W2e: every bench sitting carries the law clause — precedent is argument, the leaf wins, persuasive vs affirmed', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    blueSynth: [blueEnv({ petitions: [{ class: 'ethical', basis: 'b', relief: 'r' }] })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 4 })
  for (const seat of ['judge-petition', 'judge-r']) {
    const c = world.calls.find((x) => x.opts.label.startsWith(seat))
    assert.ok(c, `${seat} sat`)
    assert.ok(c.prompt.includes('PRECEDENT IS ARGUMENT, NOT EVIDENCE') && c.prompt.includes('the leaf wins') && c.prompt.includes('only AFFIRMED ones bind'), `${seat} missing law clause`)
  }
})

// ---- W2g: MASS v2 (existence/consequence split) + catechism into blue's template ----

test('W2g: gaps carry existence; merge prompt redefines likelihood as consequence-only', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const merge = world.calls.find((c) => c.opts.label.startsWith('red-merge-r1'))
  assert.ok(merge.prompt.includes('existence (verified') && merge.prompt.includes('CONSEQUENCE ONLY'), 'grading redefinition stated')
})

test('W2g: blue authors the catechism at round 0 inside the audited report; assembly union-copies and never authors', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const synth = world.calls.find((c) => c.opts.label.startsWith('blue-synthesize'))
  assert.ok(synth.prompt.includes('THE CATECHISM IS YOURS') && synth.prompt.includes('## The Catechism'), 'catechism is blue round-0 duty')
  assert.ok(synth.prompt.includes('every risk-accepted residual'), 'against-case at full strength demanded (the E0.5h/catechism-audit omission class)')
  // Blue authors ONLY its surfaces — not the tool-composed ones. A "## Red team findings"
  // blue writes is fabrication (it cannot know red's findings) and duplicates the tool.
  assert.ok(synth.prompt.includes('## Red team findings') && synth.prompt.includes('is FABRICATION'), 'blue is forbidden the record-composed sections; authoring them is fabrication')
  assert.ok(synth.prompt.includes('ONLY WHAT YOU CAN AUTHOR'), 'blue report scope is bounded to its own surfaces')
  const assemble = world.calls.find((c) => c.opts.label.startsWith('assemble'))
  assert.ok(assemble.prompt.includes('cannot mis-author a synthesis surface') && assemble.prompt.includes('do not copy anything yourself'), 'the tool composes/lifts; the seat is told to author nothing and copy nothing')
})

// ---- W2h: the scorecard visibility loop reaches the seat ----

test('priors-are-poison: the cross-run scorecard SEED is not injected into any chair prompt, even when scorecards are supplied', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, {
    ...ARGS,
    scorecards: {
      blue: ['repair_regression_ratio 0.63 [BENCHMARK]'],
      red: ['anchored_closures_pct 89 [BENCHMARK]'],
      bench: ['carried_share 0.98 [BENCHMARK]'],
    },
  })
  const promptOf = (label) => world.calls.find((c) => c.opts.label.startsWith(label)).prompt

  // A prior run's numbers are Goodhart bait, topic-confounded, cross-model, and
  // salience-priming (plans/priors-are-poison.md). No chair prompt carries the seed —
  // the `scorecards` arg feeds operator analytics only. A chair reads its OWN in-run
  // scorecard via `feov-record scorecard`, never a predecessor's numbers.
  assert.ok(!promptOf('blue-synthesize').includes('repair_regression_ratio 0.63'), 'blue is not seeded with its prior number')
  assert.ok(!promptOf('red-merge-r1').includes('anchored_closures_pct 89'), 'the merge is not seeded')
  assert.ok(!promptOf('assemble').includes('carried_share 0.98'), 'assembly is not seeded')
  assert.ok(!world.calls.some((c) => /YOUR CHAIR'S SCORECARD/.test(c.prompt)), 'the seed clause is gone entirely')
})

test('priors-are-poison half-2 (binDir): the in-run self-read runs `feov-record scorecard`, gated on binDir, and takes no --bin (#121 slice 3/5)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, {
    ...ARGS,
    binDir: '/plug/bin',
    scorecards: { blue: ['repair_regression_ratio 0.63 [BENCHMARK]'] },
  })
  const promptOf = (label) => world.calls.find((c) => c.opts.label.startsWith(label)).prompt
  const blue = promptOf('blue-synthesize')
  // With binDir set, the self-read is the Go subcommand — the command IS the tool, so no --bin.
  assert.ok(blue.includes('/plug/bin/feov-record scorecard --run') && blue.includes('--chair blue'), 'blue reads its own scorecard via feov-record scorecard')
  assert.ok(promptOf('red-merge-r1').includes('--chair red'), 'the merge, a red chair, reads the red scorecard')
  assert.ok(!blue.includes('scorecards.mjs'), 'the clause is binary-only — the node scorecards.mjs fallback is retired')
  assert.ok(!/feov-record scorecard[^\n]*--bin/.test(blue), 'the scorecard self-read takes no --bin')
  assert.ok(!blue.includes('repair_regression_ratio 0.63'), 'the cross-run seed is still not injected')
})

// The in-run scorecard is computed from THIS run's record by `feov-record scorecard`; the
// `scorecards` arg is operator analytics and feeds no prompt (priors-are-poison). So a first
// run — no priors to pass — still gets the clause, and is never told a number it did not earn.
test('W2h: no scorecards arg -> the chair still gets its in-run scorecard, with no prior numbers in it', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const chairs = world.calls.filter((c) => /YOUR IN-RUN SCORECARD/.test(c.prompt))
  assert.ok(chairs.length > 0, 'the clause does not depend on the scorecards arg')
  for (const c of chairs) {
    assert.ok(/feov-record scorecard --run/.test(c.prompt), 'it is the tool that computes it')
    assert.ok(!/\d+\.\d\d/.test(c.prompt.match(/YOUR IN-RUN SCORECARD[^.]*\./)[0]), 'no prior number is seeded')
  }
})

// ---- memory-as-duty: the class join delivers patterns to the repair, not the seat ----

const PATTERNS = {
  'citation-figure-misattribution': [{ file: 'cite.md', title: 'Citation misattribution', hook: 'real figures cited to the wrong source' }],
  'false-universal': [{ file: 'universal.md', title: 'False universal', hook: '"every" claims that hold for the sampled cases only' }],
  'live-source-drift': [{ file: 'drift.md', title: 'Live-source drift', hook: 'volatile figures only catchable live' }],
}

test('memory-as-duty: only the patterns matching the REPAIRED gaps class are delivered', async () => {
  const world = makeWorld(makeResponder({
    red: [
      redEnv({ gaps: [{ ...gap('R1-1'), class: 'false-universal' }] }),
      redEnv({ verdict: 'PASS' }),
    ],
  }))
  await world.run(script, { ...ARGS, gapPatterns: PATTERNS })
  const respond = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1')).prompt

  assert.ok(respond.includes('False universal'), 'the gap class selects its pattern')
  assert.ok(!respond.includes('Citation misattribution'), 'an unrelated class is NOT delivered — the point is the join, not the corpus')
  assert.ok(!respond.includes('Live-source drift'), 'nor is any other unrelated class')
  // The duty has to be a duty: checked before closure, recorded in the manifest.
  assert.ok(/manifest row which patterns you checked/.test(respond), 'the check is recorded, so a skipped duty is visible')
})

test('memory-as-duty: a gap whose class has no patterns adds no clause (no noise, no ritual)', async () => {
  const world = makeWorld(makeResponder({
    red: [
      redEnv({ gaps: [{ ...gap('R1-1'), class: 'a-class-with-no-memory' }] }),
      redEnv({ verdict: 'PASS' }),
    ],
  }))
  await world.run(script, { ...ARGS, gapPatterns: PATTERNS })
  const respond = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1')).prompt
  assert.ok(!/PATTERN DUTY/.test(respond), 'no matching memory -> no clause')
})

test('memory-as-duty: patterns are deduped across gaps sharing a class', async () => {
  const world = makeWorld(makeResponder({
    red: [
      redEnv({ gaps: [{ ...gap('R1-1'), class: 'false-universal' }, { ...gap('R1-2'), class: 'false-universal' }] }),
      redEnv({ verdict: 'PASS' }),
    ],
  }))
  await world.run(script, { ...ARGS, gapPatterns: PATTERNS })
  const respond = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1')).prompt
  assert.equal((respond.match(/False universal/g) || []).length, 1, 'one pattern, once — a repeated duty reads as two duties')
})

// ---- Lines of Inquiry: exploration becomes a record, and a surface for red ----

test('lines of inquiry: every blue seat is told to record avenues; red L5/L6 audit them', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, { ...ARGS, binDir: '/plug/bin' })
  const p = (label) => world.calls.find((c) => c.opts.label.startsWith(label)).prompt

  for (const seat of ['blue-lane-1', 'blue-synthesize', 'blue-respond-r1']) {
    assert.ok(/LINES OF INQUIRY/.test(p(seat)), `${seat} records avenues`)
    assert.ok(/abandoned/.test(p(seat)), `${seat} is told the abandoned ones matter most`)
  }

  // The STEELMAN duty is the point of recording declines: E0.5h measured that the
  // case AGAINST a design attracts no adversary, so it lands on the two lenses
  // that audit arguments rather than sources.
  assert.ok(/STEELMAN DUTY/.test(p('red-lens-5-r1')), 'the logic lens audits the declines')
  assert.ok(/STEELMAN DUTY/.test(p('red-lens-6-r1')), 'so does dark-side')
  assert.ok(!/STEELMAN DUTY/.test(p('red-lens-1-r1')), 'citation slices verify sources, not arguments')

  // The exploration space reaches the reader, not just the record.
  assert.ok(/expansions and alternatives-considered/.test(p('assemble')), 'the avenues reach the report as the expansions (pursued) and the alternatives considered (abandoned/declined)')
})

// THE VERB IS ALWAYS THERE, so the duty is always instructed. This asserted the INVERSE — that
// without the record tool the clause is absent — which was the honest shape while a toolless run
// existed. It does not: binDir is required, so an uninstructed lines-of-inquiry duty would now be
// a capability nobody is told about rather than a duty nobody could discharge.
test('lines of inquiry: the duty is instructed, because the verb that discharges it exists', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  assert.ok(world.calls.some((c) => /LINES OF INQUIRY|STEELMAN DUTY/.test(c.prompt)),
    'the avenue verb exists on every run, so the duty must be stated somewhere a seat reads')
})

// ---- integrity inspection: the bench may read trajectories, on two conditions ----

test('integrity inspection arms only with transcriptDir, and binds integrity-not-merits', async () => {
  const armed = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    blueRespond: [blueEnv({ grade_disputes: [{ gap_id: 'R1-1', dimension: 'impact', proposed: 'low', evidence: 'e' }] })],
  }))
  await armed.run(script, { ...ARGS, transcriptDir: '/sess/wf-123' })
  const bench = armed.calls.find((c) => c.opts.label.startsWith('judge')).prompt

  assert.ok(/INTEGRITY INSPECTION/.test(bench), 'the bench is told it may read')
  assert.ok(/\/sess\/wf-123\/agent-\*\.jsonl/.test(bench), 'and where')
  // The separation is the whole reason this is oversight rather than surveillance.
  assert.ok(/MUST NOT use trajectory material to decide the MERITS/.test(bench), 'integrity only, never merits')
  assert.ok(/DECLARE every inspection/.test(bench), 'the finding goes on the record even though the looking did not')

  // Without the path there is nothing to read: the capability is honest about
  // when it exists rather than instructing a seat to reach for what is not there.
  const unarmed = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await unarmed.run(script, ARGS)
  assert.ok(!unarmed.calls.some((c) => /INTEGRITY INSPECTION/.test(c.prompt)), 'no transcriptDir -> no clause')
})

test('CEILING is distinct from UNVERIFIED: a judged deadlock and a PASS keep their own names', async () => {
  // A ceiling exit must not be read as a judged failure to verify, and the two
  // real terminators must not be swallowed by the new class.
  const passing = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  assert.equal((await passing.run(script, ARGS)).verdict, 'VERIFIED')

  const ceiling = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R2-1')] })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'owed' }] })],
  }))
  const out = await ceiling.run(script, { ...ARGS, maxRounds: 2 })
  assert.equal(out.verdict, 'CEILING')
  const asm = ceiling.calls.find((c) => c.opts.label.startsWith('assemble')).prompt
  assert.ok(/--as CEILING/.test(asm), 'the tool is driven to record the CEILING verdict via bench outcome; the stamp text (never red-audited, re-audit debt OUT of the run) is tested in internal/report')
})

// The docket carries PERSISTING disputes: re-raised, or descending by supersedes. A gap
// minted fresh this round is not docket-bound by that rule — but both seats file closing
// arguments by their own reading of the prose contract, and in the 2026-07-18 run red and
// blue both argued R3-2 to closing while the engine never docketed it. It reached no
// ruling and returned to red's verdict pool as though nobody had considered it. The
// engine cannot pre-compute the docket (it depends on red's own output from the same
// turn), so nothing may be withheld SILENTLY.
test('a fresh gap kept off the docket still reaches the bench, with the reason', async () => {
  const persisting = gap({ id: 'R1-1', severity: 'high' })
  const fresh = gap({ id: 'R2-9', severity: 'medium' })
  const world = makeWorld(makeResponder({
    red: [
      redEnv({ verdict: 'FAIL', gaps: [persisting] }),
      redEnv({ verdict: 'FAIL', gaps: [persisting, fresh] }),
      redEnv({ verdict: 'PASS' }),
    ],
  }))
  await world.run(script, JSON.stringify({ ...ARGS, maxRounds: 3 }))
  const judgePrompt = world.calls.filter((c) => c.opts.label.startsWith('judge-r')).map((c) => c.prompt).join('\n')
  assert.ok(judgePrompt.includes('WITHHELD FROM THE DOCKET'), 'the bench is told what was withheld')
  assert.ok(judgePrompt.includes('R2-9'), 'the fresh gap is named rather than silently dropped')
  assert.ok(/minted fresh this round/.test(judgePrompt), 'and the reason it was withheld is given')
  assert.ok(/docket defect, not a decision/.test(judgePrompt), 'the bench may rule on one it judges wrongly withheld')
})

// Docket text is snapshotted when red merges; blue repairs afterwards, in the same round.
// In the 2026-07-18 run BOTH docketed premises asserted "blue took no round-3 turn" and
// were false by the time the bench sat — a bench that credited the problem statement
// instead of re-running the check would have carried two already-discharged gaps.
test('the bench is told the docket premise may be stale and to re-check the live artifact', async () => {
  const g = gap({ id: 'R1-1', severity: 'high' })
  const world = makeWorld(makeResponder({
    red: [redEnv({ verdict: 'FAIL', gaps: [g] }), redEnv({ verdict: 'FAIL', gaps: [g] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, JSON.stringify({ ...ARGS, maxRounds: 3 }))
  const judgePrompt = world.calls.filter((c) => c.opts.label.startsWith('judge-r')).map((c) => c.prompt).join('\n')
  assert.ok(/STALENESS/.test(judgePrompt))
  assert.ok(/AS IT NOW STANDS/.test(judgePrompt), 'rule on the artifact as it stands, not as the docket describes it')
  assert.ok(/DOCUMENT-PROBE acceptance check/.test(judgePrompt), 'and re-run the checks that can be re-run')
})

// #394: ONE SEAT ID NAMES ONE SITTING.
//
// hearPetitions used to dispatch every sitting as the literal `judge-petition` — once after
// synthesis and twice per round. Each register rotates the nonce, so N sittings meant N shards
// under one seat id, and replay keeps one: a petition sitting writes no `verdict` and no
// `revision`, so the terminal pool is empty and selection falls to latest mtime. Every earlier
// sitting's rulings were dropped while the run reported success.
test('W2c: each petition sitting gets its own seat id, derived from the petitioner', async () => {
  const petition = [{ class: 'constitutional', basis: 'b', relief: 'r' }]
  const world = makeWorld(makeResponder({
    // Three sittings in one run: after synthesis, after the round-1 merge, after the round-1
    // response. Under the old id all three were `judge-petition`.
    blueSynth: [blueEnv({ petitions: petition })],
    red: [redEnv({ gaps: [gap('R1-1')], petitions: petition }), redEnv({ verdict: 'PASS' })],
    blueRespond: [blueEnv({ petitions: petition })],
  }))
  await world.run(script, ARGS)

  const sittings = world.calls.filter((c) => c.opts.label.startsWith('judge-petition'))
  assert.ok(sittings.length >= 2, `want multiple sittings to have something to collide, got ${sittings.length}`)

  const ids = sittings.map((c) => c.opts.label.replace(/ · .*$/, ''))
  assert.equal(new Set(ids).size, ids.length, `every sitting needs its OWN id, got ${JSON.stringify(ids)}`)
  assert.ok(!ids.includes('judge-petition'), 'the bare id is the collision — it must not be dispatched')

  // The id names its petitioner, which is what makes it unique by construction rather than by a
  // counter. And it is the SEAT_ID the prompt hands the seat, not merely a display label.
  for (const c of sittings) {
    const id = c.opts.label.replace(/ · .*$/, '')
    assert.ok(c.prompt.includes(`SEAT_ID: ${id}`), `the record contract must carry the same id: ${id}`)
  }
  assert.ok(ids.some((i) => i === 'judge-petition-blue-synthesize'), `the pre-round sitting names its filer: ${JSON.stringify(ids)}`)
  assert.ok(ids.some((i) => /^judge-petition-(red-merge|blue-respond)-r1$/.test(i)),
    `an in-round sitting carries the round in a position RoundOf reads: ${JSON.stringify(ids)}`)
})

// node --test — the debate engine's regression suite. Run:
//   node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs (directory form fails on Windows node — list files explicitly)
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
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
// THE CITATION SEAT IS IDENTIFIED BY ITS ROLE, NOT BY A SENTENCE IN ITS PROMPT.
//
// This counter used to grep the ledger clause's opening words. Keyed that way it returned 0 both
// when no citation seat was dispatched and when the clause was reworded — the miss and the honest
// zero were the same number, and rewording the clause reported as "the sizing rule stopped
// working". L1-L4 are the citation slices BY ROLE (L5 logic, L6 dark-side), the role is in the
// label, and it is stable across rounds by construction.
const lensRole = (c) => Number(c.opts.label.match(/^red-lens-(\d+)-/)[1])
const citationSeats = (world, r) => lensesByRound(world, r).filter((c) => lensRole(c) <= 4)
// The ledger clause's own presence is asserted separately, on the clause's subject rather than
// on its first four words.
const CITATION_CLAUSE = 'THE REPORT DOES NOT NAME ITS SOURCES'

// THE ROSTER IS A LIST OF AREAS, NOT A FUNCTION OF THE CORPUS (#771).
//
// Three tests stood here pinning citationPasses: round 1 sized on the corpus, round 2 on the
// delta, round 3 restored by the staleness trigger. That arithmetic is gone with the instancing
// it sized, and the replacement invariant is stronger and simpler — the number of lenses is a
// property of the ROSTER. A corpus twenty times larger dispatches exactly the same seats.
test('the lens roster does not scale with the corpus — one seat per strategic area', async () => {
  const count = async (claims) => {
    const world = makeWorld(makeResponder({
      blueSynth: [blueEnv({ claim_count: claims })], red: [redEnv({ verdict: 'PASS' })],
    }))
    await world.run(script, ARGS)
    return world.calls.filter((c) => c.opts.label.startsWith('red-lens')).length
  }
  assert.equal(await count(10), 4, 'evidence + logic + risk + voice')
  assert.equal(await count(200), 4, 'a 20x corpus dispatches the SAME four areas')
})

test('exactly one evidence seat sits per round, in every round', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ claim_count: 200 })],
    blueRespond: [blueEnv({ claim_count: 400 }), blueEnv({ claim_count: 401 })],
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R2-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [
      judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'owed' }] }),
      judgeEnv({ resolutions: [{ gap_id: 'R2-1', resolution: 'carried', rationale: 'still owed' }] }),
    ],
  }))
  await world.run(script, ARGS)
  for (const r of [1, 2, 3]) {
    assert.equal(citationSeats(world, r).length, 1, `round ${r} dispatches one evidence seat`)
  }
})

test('W2i: the consolidated duty binds the evidence seat from round 2, and keeps coverage observable', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ claim_count: 200 })],
    blueRespond: [blueEnv({ claim_count: 210 })],
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'more research owed' }] })],
  }))
  await world.run(script, ARGS)
  for (const c of citationSeats(world, 1)) assert.ok(!c.prompt.includes('YOUR ROUND'), 'round 1 sweeps the corpus, not the delta')
  const r2 = citationSeats(world, 2)[0]
  assert.ok(r2.prompt.includes('YOUR ROUND'), 'the round-2 evidence seat carries the consolidated duty')
  assert.ok(r2.prompt.includes('YOU OWN THE WHOLE EVIDENCE PICTURE'), 'one seat means the cross-corpus defect is reachable')
  assert.ok(r2.prompt.includes('SPOT-CHECK'), 'the duty keeps a spot-check of already-verified pairs')
  assert.ok(r2.prompt.includes('COVERAGE IS AN OBSERVABLE'), 'what went unexamined must be stated, not assumed')
  const dark = lensesByRound(world, 2).find((c) => c.opts.label.startsWith('red-lens-6-'))
  assert.ok(!dark.prompt.includes('YOUR ROUND'), 'L6 is untouched by the evidence seat\'s duty')
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
        closures: [{ id: 'R1-1', class: 'repaired_with_regression' }],
      }),
      redEnv({ verdict: 'PASS' }),
    ],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R2-1', resolution: 'repaired', rationale: 'chain resolved' }] })],
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
        closures: [{ id: 'R1-1', class: 'repaired_with_regression' }],
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
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'defect_accepted', rationale: 'tradeoff' }] })],
  }))
  await world.run(script, ARGS)
  const judgeCalls = world.calls.filter((c) => c.opts.label.startsWith('judge'))
  assert.equal(judgeCalls.length, 1, 'skip-a-round re-raise must still reach the judge')
  assert.ok(judgeCalls[0].prompt.includes('"R1-1"'))
})

// ---- Run-3 docket rows 21 + 24: friction everywhere, persisted in prompts ----

test('blue-synthesize log entries reach the aggregate (row 21)', async () => {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ log: ['write-block on blue/report.md'] })],
    red: [redEnv({ verdict: 'PASS' })],
  }))
  const out = await world.run(script, ARGS)
  assert.ok(out.friction.includes('blue-synthesize: write-block on blue/report.md'))
})

test('every seat prompt carries the log clause (envelope + verb, not a hand-written file); lenses are transcript-forbidden and record findings via the tool', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  // THE LENS IS IN THIS LIST NOW. It was the one seat class the clause was never appended to —
  // found 2026-08-13 when the orphan gate reported `lens friction` as named nowhere a seat reads.
  // Four lens seats per round were told to close a channel nobody had told them about.
  for (const seat of ['blue-synthesize', 'red-merge-r1', 'blue-respond-r1', 'assemble', 'red-lens-1-r1']) {
    const c = world.calls.find((c) => c.opts.label.startsWith(seat))
    // WHAT IS LEFT HERE AFTER THE SUBTRACTION. The clause used to carry the duty, the explicit
    // empty form, and the fact that silence is not the empty case — all three of which the
    // `friction` verb's own help now states, in those words, on the page the tree walk opens.
    // What is genuinely the PROMPT's is the ENVELOPE half (a schema field, not a verb) and the
    // AUDIENCE: this channel is read by the operator who can retool the seat, which no help page
    // can tell a seat about its own sitting.
    //
    // THE SURVEY IS GONE, AND IT WAS MEASURED OUT. This assertion used to require the prompt to
    // demand a survey of "the verbs you read and did NOT use", defended here as the instrument
    // the traversal is measured with. Measured on 2026-09-02_quadratic-formula: the surveys are
    // 64,960 of 142,891 characters of this channel — 45.5% — and a mechanical search of all of
    // them for a single proposed fix returns ZERO. The instrument is also unfalsifiable, and
    // false where it can be checked: red-lens-r4-L2's survey states it rejected `reproduce` and
    // `finding`, while that seat's own events record both. A check that cannot fail, and does not
    // hold where it can be tested, is what the bench docked blue for at R3-5 in that same run.
    // Both seats, asked afterwards, called it duty rather than judgement — one said it wrote to
    // be "visibly compliant with a prompt that pre-accuses the seat", knowing the difference.
    //
    // WHAT REPLACED IT IS THE PART THEY DEFENDED: a decline that was a JUDGEMENT rather than an
    // absence of occasion. The record holds every verb a seat ran, so "what I used" is derivable
    // and never needs asking; what nothing derives is what a seat weighed and set aside. That is
    // the one sentence this clause still asks for, and the only part of the survey with a reader.
    assert.ok(c.prompt.includes("envelope's log field"), `${seat} lost the envelope half of the operator channel`)
    assert.ok(/AUDIENCE IS THE OPERATOR/.test(c.prompt),
      `${seat} lost the audience half — a capability gap is addressed to whoever can retool the seat, not to the debate`)
    assert.ok(/JUDGEMENT rather than for want of occasion/.test(c.prompt),
      `${seat} lost the judgement-decline ask — the one part of the retired survey the record cannot derive`)
    assert.ok(!/OWES THE SURVEY/.test(c.prompt) && !/did NOT use/.test(c.prompt),
      `${seat} still demands the verb survey — 45.5% of the channel, zero fix proposals, and unfalsifiable`)
    assert.ok(!/SILENCE IS NOT THE EMPTY CASE/.test(c.prompt),
      `${seat} restates the empty-form rule the verb's help states — two copies of one rule is what this pass removed`)
  }
  const lens = world.calls.find((c) => c.opts.label.startsWith('red-lens-1-r1'))
  // Transcript-forbidden, stated POSITIVELY: the clause used to prohibit writing to debate.md,
  // a file setup no longer creates. The contract it was protecting is who owns the round's
  // narrative — one seat, through one verb — and that is the form a lens can actually act on.
  assert.ok(/Only red-merge writes the round's RED narrative/.test(lens.prompt),
    'a lens is not a debate party — the round narrative belongs to one seat')
  // The ACT and the fact the LABEL IS THE TOOL'S — same reason as the friction clause above.
  assert.ok(/ANCHOR EVERY FINDING TO A QUOTED SENTENCE/.test(lens.prompt) && /the labels on your findings are the tool's to assign/.test(lens.prompt),
    'lens records findings as events with a tool-assigned role-scoped label (L1-F{N})')
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
  // The bench's rulings reach blue through the transcript, and a CARRIED gap arrives owing a
  // stated research direction. Asserted on the fact rather than on the "### LEAD" heading the
  // transcript happens to render them under — that heading is the tool's, and the prompt naming
  // it was the prompt describing the tool's output format back to the seat.
  assert.ok(/bench's latest resolutions/.test(resp.prompt), 'blue is not sent to read the bench rulings it must answer')
  assert.ok(/gap the judge CARRIED comes with a stated research direction you owe/.test(resp.prompt), 'carried-gap direction delivery missing')
  assert.ok(/propagate every correction to ALL sites/i.test(resp.prompt), 'propagation clause missing')
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
  // BOTH SEATS ARE TOLD SUBSTANCE LEAVES ON THE RECORD — which is the invariant. What the VERB
  // is called, and that capture compares the claim-count fall against the retire events, are
  // both in `retire --help`; the prompt saying them too is the duplication this pass removed.
  assert.ok(/retired on the record/.test(resp) && /retired on the record/.test(synth), 'both seats are told substance leaves only on the record')
  assert.ok(!/unaccounted drop/.test(resp), 'the prompt restates the detector the verb\'s own help describes')
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
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'repaired', rationale: 'blue fixed it' }] })],
  }))
  const out = await world.run(script, ARGS)
  assert.equal(out.verdict, 'VERIFIED')
  assert.equal(out.rounds, 3)
  const judgeCalls = world.calls.filter((c) => c.opts.label.startsWith('judge'))
  assert.equal(judgeCalls.length, 1, 'judge invoked exactly once, only when a gap recurs')
  assert.ok(judgeCalls[0].prompt.includes('"R1-1"'))
  assert.ok(!judgeCalls[0].prompt.includes('"R1-2"'), 'un-recurred gap is not on the docket')
  const merge3 = world.calls.find((c) => c.opts.label.startsWith('red-merge-r3'))
  // THE FATE TRAVELS WITH THE ID, and that is the assertion. Red used to be handed bare ids, so
  // the bar was enforced by making the gap invisible — indistinguishable, from red's side, from a
  // gap nobody ever raised. `repaired` and `not_a_defect` and `defect_accepted` estop red from
  // completely different propositions, and it cannot honour a bar whose grounds it cannot see.
  assert.ok(merge3.prompt.includes('{"gap_id":"R1-1","resolution":"repaired"}'),
    'red must be told the FATE of an adjudicated gap, not merely its id')
  assert.ok(merge3.prompt.includes('ESTOPPEL'),
    'the exclusion must be named as estoppel — red may reopen its OWN closure, but not a bench ruling')
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

test('the evidence seat carries the ledger clause at any corpus size', async () => {
  const evidence = async (claims) => {
    const world = makeWorld(makeResponder({ blueSynth: [blueEnv({ claim_count: claims })], red: [redEnv({ verdict: 'PASS' })] }))
    await world.run(script, ARGS)
    const lenses = world.calls.filter((c) => c.opts.label.startsWith('red-lens'))
    assert.equal(lenses.length, 4, 'four areas')
    return lenses[0]
  }
  for (const claims of [10, 200]) {
    assert.ok((await evidence(claims)).prompt.includes(CITATION_CLAUSE), 'the evidence seat carries the ledger clause')
  }
})

test('the operator channel aggregates from every seat with attribution', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')], log: ['no PDF extraction'] }), redEnv({ verdict: 'PASS', log: [] })],
    blueRespond: [blueEnv({ log: ['rate-limited on WebFetch'] })],
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
  // THE PROMPT NO LONGER MENTIONS TELEMETRY AT ALL, AND THAT IS THE ASSERTION.
  //
  // It used to spend a paragraph naming the per-round fields and insisting the seat not hand-write
  // them. There is no verb that writes telemetry — the tool recomputes the series from the record
  // on every read — so the paragraph was warning the seat off something it structurally cannot do,
  // in a prompt where every sentence competes for the seat's attention. `show telemetry`'s own
  // help says what the series carries and that it is the stopping signal, to the seat that goes
  // looking for it.
  for (const gone of ['COMPUTED BY THE TOOL', 'you do NOT hand-write it', 'open_count, mass, max_severity',
                      'feov-record merge render', 'trajectories/board-telemetry.jsonl']) {
    assert.ok(!merge.prompt.includes(gone), `the merge prompt still carries telemetry mechanics (${gone}) — the tool computes the series and its own help describes it`)
  }
  assert.ok(merge.opts.schema.properties.gaps, 'the merge still returns a board-derived envelope')
})

test('the merge reads this round\'s findings from the record view, not a candidate-file cat', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const merge = world.calls.find(c => c.opts.label.startsWith('red-merge'))
  // NOT `FIRST ACTION` ANY MORE, and this assertion used to pin that phrase. Reading the findings
  // is the merge's first READ; the tree walk comes before it. Two clauses both claiming the
  // opening move is `co-resident-rules-disagree`, and it was measured: every blue board obeyed the
  // batched-read clause at the top of its prompt and never traversed, while merge — whose
  // competing clause is a single read — traversed anyway. What this test is FOR is that the merge
  // reads findings from the record view rather than catting a candidate file, so it pins that and
  // the ordering, not the words that happened to carry them.
  // NO CLAUSE CLAIMS THE OPENING MOVE ANY MORE, AND THAT IS THE FIX.
  //
  // This assertion has now pinned two different clauses that each announced the merge's first act
  // — `FIRST ACTION`, then `FIRST READ … AFTER THE TREE WALK` — and the second existed only to
  // subordinate itself to the tree walk it was competing with. The reading order the merge needs
  // is stated once, in the record contract, where every seat gets it. The findings are not a file
  // and never were; `show findings` says what the view is and that the merge coalesces it.
  assert.ok(!/FIRST ACTION|YOUR FIRST READ/.test(merge.prompt), 'a clause in the merge prompt claims the opening move again — the tree walk is the opening move and only the record contract may say so')
  assert.ok(!merge.prompt.includes('red/candidates'), 'the candidate-file cat is retired')
  assert.ok(/COALESCE — DO NOT TRANSCRIBE/.test(merge.prompt), 'the merge is told its job is to coalesce findings into problem-classes')
})

test('the board is the tool: merge mints through feov-record, downstream seats pull via show, no findings.md', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  for (const c of world.calls) assert.ok(!c.prompt.includes('red/findings.md'), `findings.md leaked into: ${c.opts.label}`)
  const merge = world.calls.find(c => c.opts.label.startsWith('red-merge'))
  // THE BOARD IS THE TOOL'S, AND THE MERGE IS TOLD SO BY ITS OWN SURFACE: it has mint and close
  // verbs and no markdown to write. What the prompt owes is the DISCIPLINE — coalesce rather than
  // transcribe, one considered gap at a time, screen before minting — which no help page can
  // supply, because the tool cannot tell a considered mint from a transcribed one.
  assert.ok(/one considered gap at a time/.test(merge.prompt), 'merge is told to mint deliberately, one gap at a time')
  assert.ok(/[Ss]creen every candidate against the board first/.test(merge.prompt), 'the screen-before-mint ordering is stated (§4.5 cond 3)')
  assert.ok(merge.prompt.includes('drift triggers'), 'volatile-source closures inherit drift re-checks (§4.5 cond 4)')
  const judge = world.calls.find(c => c.opts.label.startsWith('judge'))
  // `show` became a GROUP: `show ledger`, not `show --view ledger`. A flag's VALUE space has no
  // --help and no completion of its own, which is the undiscoverability the motion collapse fixed
  // one layer up. The assertion is on the READ still being pulled through the tool, which is the
  // property that matters — the spelling changed, the contract did not.
  // Two tokens, not one phrase: the projection now comes FIRST (`show board --run <dir> --format
  // markdown`) so the group's subcommand is adjacent to `show`, which is what makes a stale name
  // catchable. Asserting the contiguous string would have been asserting the argument ORDER.
  // WHICH PROJECTION AND HOW TO PULL IT IS THE TOOL'S; that the bench must READ RATHER THAN
  // INFER is the prompt's, and it is the whole reason this assertion exists — the bench once
  // ruled on a docket snapshot that was false by the time it sat.
  assert.ok(/read them FRESH before ruling/.test(judge.prompt), 'the bench is not told to read the board fresh rather than trust the docket')
  assert.ok(/READ THE NAMED ANCESTORS' RECORDS first and NAME what you read/.test(judge.prompt), 'the demanded lineage read is not stated')
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
  assert.ok(/NOTHING CAN BE CARRIED AT A TERMINAL EXIT/.test(terminal.prompt), 'no carrying into a round that does not exist')
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
            closures: [{ id: 'R1-1', class: 'repaired_with_regression' }],
          })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 2 })
  const judge2 = world.calls.find(c => c.opts.label.startsWith('judge-r2'))
  const blue2 = world.calls.find(c => c.opts.label.startsWith('blue-respond-r2'))
  const merge2 = world.calls.find(c => c.opts.label.startsWith('red-merge-r2'))
  assert.ok(judge2 && blue2, 'both seats ran in round 2')
  assert.ok(judge2.n > blue2.n, 'the judge rules AFTER blue has answered — never on unanswered material')
  // WHICH HEADING THE TRANSCRIPT RENDERS A CLOSING UNDER IS THE TOOL'S BUSINESS, and `closing
  // --help` states it at both seats. What the prompts owe is that each side is told it is filing
  // one, on which items, and what a closing is FOR.
  assert.ok(/docket-bound/.test(merge2.prompt) && /argue it in ~120 words/.test(merge2.prompt), 'red files its closing at the merge')
  assert.ok(/CLOSING ARGUMENTS/.test(blue2.prompt) && /argue in ~120 words/.test(blue2.prompt), 'blue files its closing with its response')
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

// NO SLICE IS ASSIGNED, AND THAT IS THE POINT (#771).
//
// This asserted that four citation instances each carried "citation ownership follows the slice".
// Splitting bought only duplicate-avoidance and cost the findings worth having — two of the three
// findings the citation lenses ever raised are unreachable from a slice. One seat owns the whole
// evidence picture, so there is nothing to divide and no partition prose to keep in step.
test('the evidence seat is given no slice to own', async () => {
  const world = makeWorld(makeResponder({ blueSynth: [blueEnv({ claim_count: 200 })], red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  for (const c of world.calls.filter((c) => c.opts.label.startsWith('red-lens'))) {
    assert.ok(!/take slice|ownership follows the slice|instance \d+ of/.test(c.prompt),
      `a lens was told to take a slice: ${c.opts.label}`)
  }
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
  assert.ok(synth.includes('claim_count') && /RECORD THE ROUND/.test(synth), 'synthesis records the round with claim_count')
  assert.ok(respond.includes('claim_count') && /Record the round/i.test(respond), 'each response records the round with claim_count')
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
  // UNDERSCORES. debate.js's GRADE enum has spelled `low_medium` and `medium_high` since the
  // record's grade vocabulary moved; this assertion kept the hyphens, and nothing ran it — the
  // simulator was on the "owed before PR2 closes" list, so a carrier still speaking the old
  // model sat green-by-absence for the whole migration. Seventh instance of one value spelled
  // two ways across a boundary with one side moved, and the only one where the JS was the
  // stale half: the engine and the tool agreed, and the test between them did not.
  assert.deepEqual(en, ['low', 'low_medium', 'medium', 'medium_high', 'high', 'certain', 'realized', 'trivial'])
  // The mass mapping no longer rides in the prompt — the tool computes telemetry, and the
  // Go MASS table (record.go) is total over this enum, asserted by the record package tests.
  assert.ok(!/pinned mapping \{/.test(merge.prompt), 'the MASS json no longer bloats the merge prompt (retired with the hand-written telemetry line)')
})

test('shard creator: round-1 merge creates both shards; round 2 updates, never recreates (R4-6 one-creator)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  // THE SHARD IS GONE, WHICH IS WHY THIS PINS AN ABSENCE. There is no ledger file and no archive
  // file to create, in round 1 or any other: the board is the record, rendered on read. The
  // negative is the whole assertion — a prompt that starts telling a seat which artifact to
  // create has reintroduced the hand-written board this migration removed.
  for (const r of [1, 2]) {
    const m = world.calls.find(c => c.opts.label.startsWith(`red-merge-r${r}`))
    assert.ok(m, `round ${r} merge sat`)
    assert.ok(!/create BOTH|ledger\.md|archive\.md/.test(m.prompt), `round ${r} merge is told to hand-write a board artifact`)
    assert.ok(m.opts.schema.properties.gaps, `round ${r} merge returns the board-derived envelope`)
  }
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
  // `moot` is the fate that most needs to travel: the predicate expired, so NOBODY decided the
  // merits. A seat told only that R1-1 is excluded would read a live question as a settled one.
  assert.ok(m3.prompt.includes('{"gap_id":"R1-1","resolution":"moot"}'),
    'a moot gap leaves red scope AND says it went moot — the merits were never reached')
  assert.equal(out.verdict, 'VERIFIED')
})

test('MUST-try observable: graded-down citations require an attempt-or-impossibility line (false-paywall class)', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  const citation = world.calls.find(c => c.opts.label.startsWith('red-lens-1'))
  // THE OBSERVABLE MOVED TO THE GRADE ITSELF. `unreachable` is a value of the verify verb's --as,
  // and its enum help carries the duty at the point of use: "Say what you tried in --reason; an
  // untried 'unable to corroborate' is an incomplete audit." A seat grading a citation down reads
  // that line while choosing the value; a paragraph three thousand characters up its prompt was
  // where the false-paywall got past the rule the first time.
  assert.ok(!/attempt-or-impossibility line/.test(citation.prompt), 'the prompt restates a duty the grade\'s own enum states where the grade is chosen')
  assert.ok(/VERBATIM READS ONLY/.test(citation.prompt) && /A TRUNCATED READ IS NOT A READ/.test(citation.prompt),
    'the lens keeps the reading discipline the tool cannot enforce: read the leaf, whole, or say you could not')
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
    // THE POINT IS THAT THE NUMBER IS COMPUTED, NOT THAT A COMMAND IS SPELLED. Two honest merges
    // differed 2x by hand, so the seat is told to recompute it with the tool and never count it
    // itself; WHICH command does the computing is on the seat's own surface, where naming it in
    // the prompt hands out a slice of the tree and stops the seat reading the rest.
    assert.ok(/claim_count with the tool/.test(c.prompt) && /never hand-count/.test(c.prompt), `${seat} missing pinned claim unit`)
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

test('W1.9: defect_owed_elsewhere leaves red verdict pool and ships as a named infra debt', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }),
          redEnv({ gaps: [gap('R1-1')] }), // re-raise -> dockets
          redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'defect_owed_elsewhere', rationale: 'pin validation is setup tooling, owed by the lead' }] })],
  }))
  const result = await world.run(script, { ...ARGS, maxRounds: 4 })
  assert.equal(result.infra_debts.length, 1)
  assert.equal(result.infra_debts[0].gap_id, 'R1-1')
  assert.ok(result.infra_debts[0].owed_fix.includes('setup tooling'))
  // THE RESOLUTION SET IS THE SCHEMA'S, AND THE BENCH IS HANDED THE SCHEMA. The prompt used to
  // list all eight fates in prose beside an envelope whose enum already refused anything else —
  // and beside `opinion --help`, which says what a ruling owes. What the prompt owes is the one
  // thing neither states: that this fate routes the work OUT of the debate rather than ending it,
  // and that the debt must be NAMED or it is simply dropped.
  const judgePrompt = world.calls.find((c) => c.opts.label.startsWith('judge-r'))
  assert.ok(judgePrompt.opts.schema.properties.resolutions.items.properties.resolution.enum.includes('defect_owed_elsewhere'),
    'the disposition is offered in the resolution set the bench is handed')
  assert.ok(/NAMED infrastructure debt/.test(judgePrompt.prompt), 'the bench is told a routed finding must be named as a debt, not merely disposed of')
  const assemble = world.calls.find((c) => c.opts.label.startsWith('assemble'))
  assert.ok(/author NOTHING, you copy NOTHING/.test(assemble.prompt),
    'infra debts ship as opinions the tool composes from the debate record — the assembler is told it authors and copies nothing')
})

test('W1.10-W1.12: probe classes, sanctioned Glob/Grep fallback, respond workset batching, large-source rule', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const merge = world.calls.find((c) => c.opts.label.startsWith('red-merge-r1'))
  const respond = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1'))
  const lens = world.calls.find((c) => c.opts.label.startsWith('red-lens'))
  // THE TWO PROBE CLASSES ARE ONE VOCABULARY OR THEY ARE NOTHING: the seat that DEMANDS a probe
  // and the seat that DISCHARGES it must name the same two things, and there is no field for it —
  // it lives in prose in a required_fix, which is why it stays in the prompts.
  assert.ok(merge.prompt.includes('DOCUMENT-PROBE') && merge.prompt.includes('LIVE-PROBE'), 'merge demands classed probes (W1.10)')
  assert.ok(respond.prompt.includes('DOCUMENT-PROBE') && respond.prompt.includes('LIVE-PROBE') && respond.prompt.includes('deferred acceptance test'), 'respond discharges by class (W1.10)')
  assert.ok(lens.prompt.includes('KNOWN HARNESS LIMIT') && lens.prompt.includes('SANCTIONED fallback'), 'Glob/Grep fallback sanctioned everywhere via speedClause (W1.11)')
  // The BATCHING is the fact; the scratchpad filename it used to spell was a worked example of
  // one shell command, and a seat that copies it learns a filename rather than the habit.
  assert.ok(/in one pass rather than three/.test(respond.prompt), 'respond batches its working-set read (W1.12)')
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
  // The telemetry field list is the TOOL's (it computes the series; `show telemetry` documents
  // it), and the acceptance check is a REQUIRED flag on the mint whose usage line says what it is
  // for. What survives here is the demand that red's check be falsifiable and pre-agreed, and
  // that blue audit each repair against it.
  assert.ok(!merge.prompt.includes('repair_regression'), 'the merge prompt restates the telemetry fields the tool computes')
  assert.ok(/acceptance check/i.test(merge.prompt), 'the merge is still told its check is what settles the gap')
  const respond = world.calls.find((c) => c.opts.label.startsWith('blue-respond-r1'))
  assert.ok(/AUDIT YOUR OWN REPAIRS, ONE RECEIPT PER GAP/.test(respond.prompt) && respond.prompt.includes('manifest array'), 'manifest demanded at respond')
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
  const roleOf = (label) => world.calls.find((x) => x.opts.label.startsWith(label)).prompt
  // THE SUBJECT IS ROLE BINDING, and what CARRIES it moved. This used to assert the prompt told
  // the seat which command GROUP was its own ('Yours is `lens`'). There are no role groups: the
  // tree is scoped to the injected seat identity, so the ID is the binding — pass it and you get
  // your surface, and there is no wider one behind it. What the prompt owes is therefore that the
  // seat id it DECLARES is the one it tells the seat to PASS. A prompt stating one id and handing
  // over another would bind a seat to a surface that is not its own, which is precisely what the
  // role-group sentence used to guard.
  //
  // WHICH id belongs to which role is not asserted here and must not be: `assemble` carries no
  // role in its name at all, so recovering one from the string is the shape this suite keeps
  // deleting. That mapping is a record (record.roleSeats), pinned by seat-roster.golden.
  const bindsItsOwnID = (label) => {
    const p = roleOf(label)
    const declared = /SEAT_ID: (\S+?)\./.exec(p)
    if (!declared) return false
    return p.includes('--seat-id ' + declared[1] + ' ')
  }
  for (const s of ['red-lens', 'red-merge-r1', 'blue-respond-r1', 'assemble']) {
    assert.ok(bindsItsOwnID(s), `${s} declares a SEAT_ID it does not hand the tool — the id IS the surface`)
  }

  // THE HELP IS THE ONLY PAGE THAT INSTRUCTS, and the seat is REQUIRED to WALK THE WHOLE TREE
  // before it chooses — root, then every group, then every group nested in those — with the
  // command rung read per act.
  //
  // THIS ASSERTION USED TO PIN THE OPPOSITE, and pinned it by quoting the clause verbatim:
  // `BEFORE using any command in a group you have not yet opened`. That is a LOOKUP trigger —
  // it fires on an act already chosen — and under it a seat obeying perfectly reads the root
  // once and then opens only the page for the verb it had already picked. Measured over nine
  // sittings: 6 of 51 group pages opened, 90% of commands run without their own page read, and
  // 18 of the 23 pages that were opened were for verbs the seat went on to run. The rule was not
  // being ignored; its ceiling was confirmation. A test that quotes a clause pins whatever that
  // clause says, including its defect, so the negative below is what keeps the old shape out.
  const lens = roleOf('red-lens')
  assert.ok(/REQUIRED/.test(lens), 'reading the help is stated as required, not suggested')
  assert.ok(/--help — your whole surface/.test(lens), 'step 1: the root help, which IS the seat surface')
  // AND THE PAGE SAYS WHICH ENTRIES HIDE COMMANDS, which it did not until the usage template
  // stopped dropping cobra's command groups. The prompt used to carry that sentence — the tool
  // describing its own output format, in the prompt, because the output would not describe
  // itself. Step 2 is only followable if step 1's page marks its groups.
  assert.ok(/marks which entries hold commands it does not list/.test(lens), 'step 2 has nothing to enumerate unless step 1 says which entries are groups')
  assert.ok(lens.includes('<group> --help'), 'step 2: the group help')
  assert.ok(lens.includes('<group> <command> --help'), 'step 3: the command help, before running it')
  assert.ok(/for EVERY group that page listed/.test(lens), 'step 2 is exhaustive: every group, not the one holding the verb you want')
  assert.ok(/including the groups nested inside those/.test(lens), 'step 2 reaches leaf depth — `motion` alone holds three subgroups')
  assert.ok(/before you have decided what to do/.test(lens), 'the traversal precedes the DECISION, not merely the act')
  assert.ok(!/BEFORE using any command in a group you have not yet opened/.test(lens),
    'the group rung fires on an act already chosen again — that trigger is what made the ceiling of this rule confirmation rather than survey')
  assert.ok(/<group> <command> --help — BEFORE you run it/.test(lens), 'the command rung is ordered before use')
  // ABSENCE IS STATED AS ABSENCE — ON THE PAGE WHERE THE SEAT LOOKS FOR THE THING. The prompt
  // carried a paragraph of it; the friction footer says it verbatim at the foot of EVERY help
  // page, including the one the seat opens at the moment it fails to find a verb, which is when
  // the sentence has to arrive. (TestRoleHelpCarriesTheFrictionFooter holds that end.) The prompt
  // keeps only the consequence it can state and the footer cannot: hand-writing the artifact is
  // the failure the whole contract exists to prevent.
  assert.ok(!/DOES NOT EXIST FOR YOU/.test(lens), 'the prompt restates the friction footer that closes every help page')
  assert.ok(/Routing around it into markdown is the failure this contract exists to prevent/.test(lens), 'the prompt no longer names what routing around the tool costs')
  // The escalation path travels with the same footer, on every page. The prompt's own carrier for
  // it is the log clause, asserted at every seat by the log test above.
  assert.ok(/LOG \(/.test(lens), 'the escalation path is stated')

  // AND IT NAMES NO COMMAND. The contract is the ladder, not a list — a partial list satisfies
  // the seat's need to know what exists and stops it looking, which is the whole reason the
  // naming experiment moved this out of the prompt.
  assert.ok(!/\b(register|mint|close|verdict|regrade|retire|opinion|certify|halt)\s+--/.test(lens),
    'the record contract hands over a typeable invocation')
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
  // THE DISTINCTION IS THE PROMPT'S; THE VERB AND ITS RELAY ARE THE TOOL'S. `halt --help` says it
  // is the bench's own terminal act and that capture relays the opinion to the human verbatim.
  // What the sitting must be told is that halting and ruling are DIFFERENT DECISIONS and both can
  // be true at once — the confusion that once ended a run with no halt event on the record.
  assert.ok(/A HALT IS A DIFFERENT DECISION FROM A RULING/.test(sitting.prompt), 'the sitting is told halting is not a disposition of the petition')
  assert.ok(/both can be true at once/.test(sitting.prompt), 'and that it may rule and halt in the same sitting')
  assert.ok(/Record the halt, and ALSO return the envelope/.test(sitting.prompt), 'the halt reaches BOTH the record and the engine — the engine stopping is not the halt being recorded')
  assert.ok(!world.calls.some((c) => c.opts.label.startsWith('blue-respond')), 'the round never continued past the halt')
  const assemble = world.calls.find((c) => c.opts.label.startsWith('assemble'))
  assert.ok(assemble.prompt.includes('it is HALTED'), 'the halt reaches assembly as the run\'s terminal verdict')
  assert.ok(/author NOTHING, you copy NOTHING/.test(assemble.prompt), 'and the tool composes the terminal halt from the record — the assembler writes none of it')
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
  // `existence` WAS REMOVED IN 0.65.0 and the prompt kept instructing the seat to carry it for
  // three releases — a carrier still speaking the old model, which read as done because every
  // gate passed. The likelihood semantics that replaced it live on the flag: `--likelihood`'s
  // usage line says it grades the CONSEQUENCE and never how likely the defect is to BE there.
  assert.ok(!/existence/.test(merge.prompt), 'the merge prompt instructs the seat to carry a field the tool removed')
  const grades = merge.opts.schema.properties.gaps.items.properties
  assert.ok(!grades.existence, 'the envelope schema still declares the removed field')
  assert.ok(grades.likelihood && grades.impact, 'the consequence axes are what the envelope carries')
})

test('W2g: blue authors the catechism at round 0 inside the audited report; assembly union-copies and never authors', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, ARGS)
  const synth = world.calls.find((c) => c.opts.label.startsWith('blue-synthesize'))
  assert.ok(synth.prompt.includes('THE CATECHISM IS YOURS') && synth.prompt.includes('## The Catechism'), 'catechism is blue round-0 duty')
  assert.ok(synth.prompt.includes('every risk-accepted residual'), 'against-case at full strength demanded (the E0.5h/catechism-audit omission class)')
  // Blue authors ONLY its surfaces — not the tool-composed ones. A "## The board"
  // blue writes is fabrication (it cannot know red's findings) and duplicates the tool.
  assert.ok(synth.prompt.includes('## The board') && synth.prompt.includes('is FABRICATION'), 'blue is forbidden the record-composed sections; authoring them is fabrication')
  assert.ok(synth.prompt.includes('ONLY WHAT YOU CAN AUTHOR'), 'blue report scope is bounded to its own surfaces')
  const assemble = world.calls.find((c) => c.opts.label.startsWith('assemble'))
  assert.ok(assemble.prompt.includes('cannot mis-author a synthesis surface') && /do not copy anything into it yourself/.test(assemble.prompt), 'the tool composes/lifts; the seat is told to author nothing and copy nothing')
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

test('priors-are-poison half-2: the in-run self-read names the ACT, not a command, and carries no chair selector', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, {
    ...ARGS,
    binDir: '/plug/bin',
    scorecards: { blue: ['repair_regression_ratio 0.63 [BENCHMARK]'] },
  })
  const promptOf = (label) => world.calls.find((c) => c.opts.label.startsWith(label)).prompt
  const blue = promptOf('blue-synthesize')
  // The ACT is asserted and the command is not: which verb performs it belongs to --help, and
  // promptverbs' catalogue gate pins debate.js at ZERO named commands.
  //
  // THE CHAIR NAME LEFT THE PROMPT TOO, and that is the fix rather than a regression. It used to
  // interpolate `YOUR CHAIR (\`blue\`) through the tool — the operator command and the selector
  // that names a chair are in the root --help`, which was false three ways: the seat's root
  // --help is scoped to the seat and never listed that command, reaching it meant overriding
  // --seat-id to `operator`, and the prompt was teaching a capability the surface lacked. Four
  // seats across three runs filed friction about it. The chair is now resolved from the seat's
  // own registration, so there is nothing for the prompt to name and nothing to select.
  assert.ok(/YOUR IN-RUN SCORECARD/.test(blue), 'blue is told to read its own in-run scorecard')
  assert.ok(/YOUR CHAIR/.test(blue) && /your own surface/.test(blue), 'it is on the seat\'s own surface')
  // Asserted on the DEFECT's own phrasing, not on the bare words: the clause legitimately says
  // the read "needs no selector", and an unrelated line elsewhere names the operator as a reader
  // of red's narrative. A test that banned the words would fail on both and teach nothing.
  assert.ok(!/the operator command/.test(blue), 'the prompt no longer teaches an operator escape hatch')
  assert.ok(!/through the tool/.test(blue), 'nor the vague "through the tool" that named no verb and no page')
  assert.ok(/the seat you registered as/.test(blue), 'the chair is resolved from the registration, so there is nothing to select')
  assert.ok(/YOUR IN-RUN SCORECARD/.test(promptOf('red-merge-r1')), 'the merge, a red chair, gets the clause too')
  assert.ok(!blue.includes('scorecards.mjs'), 'the clause is binary-only — the node scorecards.mjs fallback is retired')
  assert.ok(!/--bin\b/.test(blue), 'the scorecard self-read takes no --bin')
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
    assert.ok(/YOUR CHAIR/.test(c.prompt) && /projection of this run's record/.test(c.prompt), 'it is the tool that computes it, from THIS run')
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

test('lines of inquiry: every blue seat is told to record lines of inquiry; red L5/L6 audit them', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
  }))
  await world.run(script, { ...ARGS, binDir: '/plug/bin' })
  const p = (label) => world.calls.find((c) => c.opts.label.startsWith(label)).prompt

  for (const seat of ['blue-lane-1', 'blue-synthesize', 'blue-respond-r1']) {
    assert.ok(/LINES OF INQUIRY/.test(p(seat)), `${seat} records lines of inquiry`)
    // The four FATES and what each means are the `--as` enum's, at the point of choosing one.
    // What the prompt owes is the reason to record at all: the dead ends are worth more than the
    // conclusion that survived, and that is an argument, not a field.
    assert.ok(/dead ends matter most/.test(p(seat)), `${seat} is told the dead ends matter most`)
    assert.ok(/CONSIDERED, not only the one you took/.test(p(seat)), `${seat} is told to record the alternatives it did not take`)
  }

  // The STEELMAN duty is the point of recording declines: E0.5h measured that the
  // case AGAINST a design attracts no adversary, so it lands on the two lenses
  // that audit arguments rather than sources.
  assert.ok(/STEELMAN DUTY/.test(p('red-lens-5-r1')), 'the logic lens audits the declines')
  assert.ok(/STEELMAN DUTY/.test(p('red-lens-6-r1')), 'so does dark-side')
  assert.ok(!/STEELMAN DUTY/.test(p('red-lens-1-r1')), 'citation slices verify sources, not arguments')

  // The exploration space reaches the reader, not just the record.
  // WHERE THEY LAND IS THE ASSEMBLER'S COMPOSITION, not something the assembler is told to do —
  // it authors nothing. The seats that RECORD lines are the ones who need to know a reader sees
  // them, and under which headings, because that is what makes recording a decline worth doing.
  assert.ok(/Alternatives reach the reader under their own headings/.test(p('blue-respond-r1')), 'the recording seats are not told the alternatives reach a reader')
  assert.ok(/what you weighed and rejected are three things a reader needs/.test(p('blue-synthesize')), 'the declined and abandoned lines are not stated as reader-facing')
})

// THE VERB IS ALWAYS THERE, so the duty is always instructed. This asserted the INVERSE — that
// without the record tool the clause is absent — which was the honest shape while a toolless run
// existed. It does not: binDir is required, so an uninstructed lines-of-inquiry duty would now be
// a capability nobody is told about rather than a duty nobody could discharge.
test('lines of inquiry: the duty is instructed, because the verb that discharges it exists', async () => {
  const world = makeWorld(makeResponder({ red: [redEnv({ verdict: 'PASS' })] }))
  await world.run(script, ARGS)
  assert.ok(world.calls.some((c) => /LINES OF INQUIRY|STEELMAN DUTY/.test(c.prompt)),
    'the line of inquiry verb exists on every run, so the duty must be stated somewhere a seat reads')
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
  assert.ok(/it is CEILING/.test(asm), 'the assembler is told the run\'s terminal verdict is CEILING; the stamp text (never red-audited, re-audit debt OUT of the run) is tested in internal/report')
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
  assert.ok(/THE DOCKET IS A ROUTING LIST, NOT THE EVIDENCE/.test(judgePrompt), 'the bench is told the docket is refs, not the case')
  assert.ok(/AS IT NOW STANDS/.test(judgePrompt), 'rule on the artifact as it stands, not as the docket describes it')
  assert.ok(/Re-run each document-probe acceptance check/.test(judgePrompt), 'and re-run the checks that can be re-run')
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

// #361: THE DECLARE VERB REACHES THE SEATS THAT RULE.
//
// `bench declare` shipped with a prompt clause — in `assemble` only, the one bench seat that
// rules on nothing. All three RULING sittings were silent about it, including the petition
// sitting whose failure the verb was built for: a bench with a construction both parties needed,
// `opinion` demanding an id and a fate it did not want to move, and the holding going into a
// petition ruling's opinion text where red never read it.
//
// The carrier set is the one every bench-wide duty uses — whatever carries lawClause.
test('W2e: the declare capability is on the bench surface, and no prompt re-teaches it', async () => {
  const world = makeWorld(makeResponder({
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    blueSynth: [blueEnv({ petitions: [{ class: 'ethical', basis: 'b', relief: 'r' }] })],
  }))
  await world.run(script, { ...ARGS, maxRounds: 4 })

  // The three that rule. `judge-terminal` only sits when disputes survive the exit boundary, so
  // it is asserted through the same law-clause carrier rather than by forcing a fourth scenario.
  for (const seat of ['judge-petition', 'judge-r']) {
    const c = world.calls.find((x) => x.opts.label.startsWith(seat))
    assert.ok(c, `${seat} sat`)
    // THIS TEST'S SUBJECT MOVED, AND THAT IS THE POINT OF W2e RATHER THAN A RETREAT FROM IT.
    // The verb existed for two releases while no prompt and no constitution named it, so the
    // bench could not know it had it — the paragraph here was the repair. The durable repair is
    // that `declare` is now ON THE BENCH'S SURFACE with its own page, saying what binds how the
    // record is READ, why `opinion` cannot carry it, and the measured case of a bench that put
    // exactly such a holding in a petition ruling where nobody read it. A prompt paragraph
    // teaching a verb is a workaround for a help page that does not; both is duplication.
    assert.ok(!/DECLARE:/.test(c.prompt), `${seat} teaches the declare verb in prose beside a help page that teaches it`)
    // AND THE PAGE IT MOVED TO IS READ HERE, so this test still fails if the capability goes
    // silent again — which is the failure W2e was filed for, not the wording of any paragraph.
    const page = readFileSync(new URL('../../tools/internal/cli/seat/help/declare.md', import.meta.url), 'utf8')
    assert.ok(/binds how the whole record is READ/.test(page), 'the declare help does not say what distinguishes a holding from an opinion')
    // THE ASSERTION NAMES THE LIVE VERB, and it used to name a dead one. `opinion` retired in
    // #681 Scope 2 and this regex went on passing, so the gate that exists to keep the page
    // teaching the distinction was instead PINNING the page to a verb the bench cannot run:
    // correcting the prose would have broken a green test. A check keyed on a retired name is
    // not a weaker check, it is a check pointed at the wrong thing.
    assert.ok(/docket ruling cannot carry it/.test(page), 'the declare help does not say why a docket ruling cannot hold it')
    assert.ok(!/`opinion`/.test(page), 'the declare help still names the retired `opinion` verb')
  }

  // The law clause is still bench-wide and still the prompt's: statute over precedent over
  // case-local argument is a RANKING the bench applies, not a thing the tool does. It stays as
  // the carrier-set check it always was — no ruling sitting may lose it.
  const lawSeats = world.calls.filter((c) => c.prompt.includes('PRECEDENT IS ARGUMENT, NOT EVIDENCE'))
  assert.ok(lawSeats.length >= 2, 'the law clause must reach at least the two sittings above')
  for (const c of lawSeats) {
    assert.ok(!/DECLARE:/.test(c.prompt),
      `${c.opts.label} teaches the declare verb in prose — the bench's own help page is where that lives now`)
  }
})

// A HOLDING REACHES EVERY SEAT THAT COMES AFTER IT (#503).
//
// `bench declare` exists because the bench sometimes states a CONSTRUCTION rather than disposing
// of a gap — how a term is READ, for the rest of the run, by both parties. The verb shipped, the
// render shipped, the law harvest shipped, and the delivery never did: measured on the
// 2026-08-22 sqlite run, the one holding the bench issued construed the deadlock test at the
// level of defect CLASS rather than raw mint count — the construction that let the run terminate
// — and no seat that followed it was told it existed.
//
// The same defect relief had before reliefFor (#360), and the same repair.
test('W2j: a bench holding binds every seat that follows it, and does not expire with the round', async () => {
  const world = makeWorld(makeResponder({
    // RE-RAISED, which is what docketes a gap and makes the bench sit at all. My first fixture
    // raised a fresh id each round, no docket was ever contested, the judge never sat — and the
    // test then asserted nothing about delivery while appearing to.
    // FOUR RED SITTINGS, because the bench first sits in round 2 (a gap must be RE-RAISED to be
    // docketed) and blue must respond at least once AFTER that for the both-parties claim below
    // to be testable at all.
    red: [
      redEnv({ gaps: [gap('R1-1')] }), redEnv({ gaps: [gap('R1-1')] }),
      redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' }),
    ],
    judge: [
      judgeEnv({
        resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'more research owed' }],
        holdings: ['"new this round" is measured by defect CLASS, not by raw mint count'],
      }),
      judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'still owed' }] }),
      judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'repaired', rationale: 'now fixed' }] }),
    ],
  }))
  await world.run(script, ARGS)
  assert.ok(world.calls.some((c) => c.opts.label.startsWith('judge')), 'the bench never sat, so no holding was ever laid down')

  const judgedAt = world.calls.findIndex((c) => c.opts.label.startsWith('judge'))
  const after = world.calls.slice(judgedAt + 1).filter((c) => /blue-|red-/.test(c.opts.label))
  assert.ok(after.length > 0, 'the fixture produced no round-2 seats, so this asserts nothing')
  for (const c of after) {
    assert.ok(c.prompt.includes('BENCH HOLDINGS IN EFFECT'),
      `${c.opts.label} was not told a holding binds it`)
    assert.ok(c.prompt.includes('defect CLASS'),
      `${c.opts.label} got the heading without the holding's own words`)
  }
  // BOTH PARTIES, not one. Relief is addressed and can bind a single side; a construction of a
  // term cannot — red reading it one way and blue the other is the disagreement it exists to end.
  assert.ok(after.some((c) => /blue-/.test(c.opts.label)), 'no blue seat in the sample')
  assert.ok(after.some((c) => /red-/.test(c.opts.label)), 'no red seat in the sample')

  // AND IT DID NOT REACH THE SEATS THAT PRECEDED IT, which is not a nicety: a holding rendered
  // into round 1 would be the engine asserting the bench had ruled before it sat.
  const before = world.calls.slice(0, judgedAt)
  for (const c of before) {
    assert.ok(!c.prompt.includes('BENCH HOLDINGS IN EFFECT'),
      `${c.opts.label} carried a holding laid down after it sat`)
  }
})

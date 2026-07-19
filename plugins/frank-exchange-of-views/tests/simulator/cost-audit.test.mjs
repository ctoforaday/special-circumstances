// node --test — cost-audit.mjs against a fixture transcript dir (zero tokens, no live runs).
// The doctrine behind the repo's Go hooks is "tested tools, never untested scripts" — this
// file is that test for the workflow layer's one script tool.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const tool = new URL('../../skills/research-protocol/scripts/cost-audit.mjs', import.meta.url).pathname
  .replace(/^\/([A-Za-z]):\//, '$1:/') // Windows: strip the leading slash file URLs add

const usageLine = (model, inp, out, cr, cw) =>
  JSON.stringify({ message: { model, usage: { input_tokens: inp, output_tokens: out, cache_read_input_tokens: cr, cache_creation_input_tokens: cw } } })

test('cost-audit: parses fixture transcripts, prices by tier, aggregates per seat-round', () => {
  const dir = mkdtempSync(join(tmpdir(), 'cost-audit-fixture-'))
  // A sonnet red lens, round 2: two API turns.
  writeFileSync(join(dir, 'agent-lens.jsonl'), [
    JSON.stringify({ message: { role: 'user', content: 'Red audit, round 2, lens: leaf-node citation verification' } }),
    usageLine('claude-sonnet-5', 1_000_000, 100_000, 10_000_000, 500_000),
    usageLine('claude-sonnet-5', 0, 100_000, 10_000_000, 0),
  ].join('\n'))
  // A session-model (fable) merge, round 2: one API turn.
  writeFileSync(join(dir, 'agent-merge.jsonl'), [
    JSON.stringify({ message: { role: 'user', content: 'Red merge, round 2. Read the round-2 lens passes' } }),
    usageLine('claude-fable-5', 100_000, 50_000, 2_000_000, 100_000),
  ].join('\n'))

  const out = execFileSync(process.execPath, [tool, dir], { encoding: 'utf8' })

  // Lens (sonnet, intro pricing): 1M*$2 + 0.2M*$10 + 20M*$0.2 + 0.5M*$2.5 = 2+2+4+1.25 = $9.25
  assert.match(out, /\| 2 \| red-lens \| sonnet \| 1 \| 2 \|.*\$9\.25 \|/)
  // Merge (fable): 0.1M*$10 + 0.05M*$50 + 2M*$1 + 0.1M*$12.5 = 1+2.5+2+1.25 = $6.75
  assert.match(out, /\| 2 \| red-merge \| fable \| 1 \| 1 \|.*\$6\.75 \|/)
  // Total: $16.00
  assert.match(out, /\*\*TOTAL\*\*.*\*\*\$16\.00\*\*/)
  // Cache share: (22M+0.6M)/(1.1M+0.25M+22M+0.6M) = 22.6/23.95 ≈ 94%
  assert.match(out, /Cache traffic is 94% of all tokens/)
})

test('cost-audit: refuses to run without a transcript dir', () => {
  assert.throws(() => execFileSync(process.execPath, [tool], { encoding: 'utf8', stdio: 'pipe' }))
})

// The arithmetic above is now reachable as functions rather than only through the
// CLI's stdout. The extraction is behaviour-preserving: the end-to-end test above
// still drives the same fixture through the same report.

test('tier: recognized models price by their row; an UNKNOWN model over-reports rather than under-reports', async () => {
  const { tier, PRICES } = await import('../../skills/research-protocol/scripts/cost-audit.mjs')
  assert.equal(tier('claude-sonnet-5'), 'sonnet')
  assert.equal(tier('claude-haiku-4-5'), 'haiku')
  assert.equal(tier('claude-opus-4-8'), 'opus')
  assert.equal(tier('CLAUDE-OPUS-4-8'), 'opus', 'model ids are matched case-insensitively')
  // A model nobody has taught this table about must not quietly cost nothing: an
  // under-reported bill is the error that goes unnoticed, so the fallback is the
  // most expensive row.
  assert.equal(tier('some-future-model'), 'fable')
  assert.equal(tier(null), 'fable')
  assert.equal(tier(''), 'fable')
  const worst = Object.keys(PRICES).reduce((a, b) => (PRICES[a][0] >= PRICES[b][0] ? a : b))
  assert.equal(tier(null), worst, 'the fallback IS the dearest row — if the table grows, this must be re-checked')
})

test('classifyTranscript: seat identity from the prompt head, unknown heads land in `other`', async () => {
  const { classifyTranscript } = await import('../../skills/research-protocol/scripts/cost-audit.mjs')
  assert.deepEqual(classifyTranscript('Red audit, round 7, lens: x'), { seat: 'red-lens', round: 7 })
  assert.deepEqual(classifyTranscript('Red merge, round 11.'), { seat: 'red-merge', round: 11 })
  assert.deepEqual(classifyTranscript('Blue response, round 2.'), { seat: 'blue-respond', round: 2 })
  assert.deepEqual(classifyTranscript('Adjudication, round 3.'), { seat: 'judge', round: 3 })
  assert.deepEqual(classifyTranscript('Blue synthesis'), { seat: 'blue-synthesize', round: 0 })
  assert.deepEqual(classifyTranscript('Blue lane 2'), { seat: 'blue-lane', round: 0 })
  assert.deepEqual(classifyTranscript('…formulate frontier hypotheses'), { seat: 'frontier', round: 0 })
  assert.deepEqual(classifyTranscript('Final assembly'), { seat: 'assemble', round: 0 })
  assert.deepEqual(classifyTranscript('unrecognized'), { seat: 'other', round: 0 })
  // A round-bearing prompt that mentions a round-0 marker is still its own seat —
  // the cost of a mis-attributed seat is a cost table that blames the wrong chair.
  assert.deepEqual(classifyTranscript('Red merge, round 4. Prior: Blue synthesis'), { seat: 'red-merge', round: 4 })
})

test('scanTranscript: a half-written final line costs the report nothing (runs get killed mid-append)', async () => {
  const { scanTranscript } = await import('../../skills/research-protocol/scripts/cost-audit.mjs')
  const txt = [
    JSON.stringify({ message: { role: 'user', content: 'Red merge, round 2. FIRST ACTION' } }),
    usageLine('claude-sonnet-5', 1_000_000, 0, 0, 0),
    JSON.stringify({ message: { role: 'assistant', content: [] } }), // no usage: not an API turn
    '',                                                              // blank lines are routine
    '{"message":{"usage":{"input_tokens":999',                       // the process died here
  ].join('\n')
  const r = scanTranscript(txt)
  assert.equal(r.seat, 'red-merge')
  assert.equal(r.round, 2)
  assert.equal(r.turns, 1, 'only records CARRYING usage are API turns')
  assert.equal(r.inp, 1_000_000)
  assert.equal(r.cost, 2, 'sonnet input at $2/MTok')
  // A transcript with no usage at all still produces a row: an agent that burned
  // nothing is a fact worth seeing, and dropping it would hide a stalled seat.
  const bare = scanTranscript(JSON.stringify({ message: { role: 'user', content: 'Final assembly' } }))
  assert.deepEqual([bare.seat, bare.turns, bare.cost], ['assemble', 0, 0])
  assert.equal(bare.t, 'fable', 'no model observed falls back to the dearest tier')
  // The model is remembered across turns: a later usage record without a model
  // field belongs to the model already seen, not to the unknown-model fallback.
  const mixed = scanTranscript([
    JSON.stringify({ message: { role: 'user', content: 'Blue lane 1' } }),
    usageLine('claude-haiku-4-5', 1_000_000, 0, 0, 0),
    JSON.stringify({ message: { usage: { input_tokens: 1_000_000 } } }),
  ].join('\n'))
  assert.equal(mixed.t, 'haiku')
  assert.equal(mixed.cost, 2, 'both turns priced at haiku input, $1/MTok')
})

test('aggregate: seat-rounds bucket together and round 10 sorts AFTER round 2, not before it', async () => {
  const { aggregate } = await import('../../skills/research-protocol/scripts/cost-audit.mjs')
  const row = (seat, round, cost) => ({ seat, round, t: 'sonnet', turns: 1, inp: 1, out: 0, cr: 0, cw: 0, cost })
  const agg = aggregate([row('red-lens', 2, 1), row('red-lens', 2, 2), row('red-lens', 10, 5), row('judge', 2, 3)])
  assert.equal(Object.keys(agg).length, 3, 'same seat+round+tier collapses into one bucket')
  const r2 = agg['02|red-lens|sonnet']
  assert.deepEqual([r2.n, r2.turns, r2.cost], [2, 2, 3], 'agents, turns and dollars all sum')
  // The key is zero-padded precisely so the report's plain lexicographic sort is
  // numeric — unpadded, round 10 would print between rounds 1 and 2.
  const order = Object.keys(agg).sort()
  assert.ok(order.indexOf('02|red-lens|sonnet') < order.indexOf('10|red-lens|sonnet'))
})

test('cacheShare: an empty run reports 0%, never NaN% written into a run record', async () => {
  const { cacheShare } = await import('../../skills/research-protocol/scripts/cost-audit.mjs')
  assert.equal(cacheShare({ inp: 0, out: 0, cr: 0, cw: 0 }), 0)
  assert.equal(cacheShare({ inp: 1e6, out: 0, cr: 3e6, cw: 0 }), 75)
  assert.equal(cacheShare({ inp: 1e6, out: 1e6, cr: 0, cw: 0 }), 0, 'output counts against the denominator')
})

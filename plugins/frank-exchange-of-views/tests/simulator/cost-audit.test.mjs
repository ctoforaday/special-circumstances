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

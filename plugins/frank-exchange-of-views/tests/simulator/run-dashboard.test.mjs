// node --test — render-run-dashboard.mjs against temp-dir fixtures.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, writeFileSync, mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { buildModel, renderHtml } from '../../skills/research-protocol/scripts/render-run-dashboard.mjs'

const tmp = () => mkdtempSync(join(tmpdir(), 'feov-dash-'))

function fixture() {
  const runDir = tmp(), transcriptDir = tmp()
  mkdirSync(join(runDir, 'trajectories'), { recursive: true })
  mkdirSync(join(runDir, 'red'), { recursive: true })
  mkdirSync(join(runDir, 'blue'), { recursive: true })
  writeFileSync(join(runDir, 'trajectories', 'board-telemetry.jsonl'), [
    JSON.stringify({ round: 1, mass: 118.75, open_count: 30, max_severity: 'high', new_mint: { count: 30, by_severity: { high: 3, medium: 9 } }, accepted_deltas: [] }),
    JSON.stringify({ round: 2, mass: 81.5, open_count: 23, max_severity: 'high', new_mint: { count: 22, by_severity: { high: 1 } }, accepted_deltas: [] }),
  ].join('\n') + '\n')
  writeFileSync(join(runDir, 'blue', 'report.md'), '# report\ncontent\n')
  writeFileSync(join(transcriptDir, 'journal.jsonl'), [
    JSON.stringify({ type: 'started', label: 'red-merge-r2 · demo', timestamp: '2026-07-17T20:00:00Z' }),
    JSON.stringify({ type: 'result', label: 'blue-synthesize · demo', result: { claim_count: 100 } }),
  ].join('\n') + '\n')
  writeFileSync(join(transcriptDir, 'agent-a1.jsonl'),
    JSON.stringify({ message: { role: 'assistant', usage: { input_tokens: 100, output_tokens: 50, cache_read_input_tokens: 1000000, cache_creation_input_tokens: 0 }, content: [] } }) + '\n')
  return { runDir, transcriptDir }
}

test('dashboard model: telemetry series, live vs done seats, cost estimate, blackboard sizes', () => {
  const { runDir, transcriptDir } = fixture()
  const m = buildModel(runDir, transcriptDir)
  assert.equal(m.telemetry.length, 2)
  assert.equal(m.latest.mass, 81.5)
  assert.equal(m.seats.filter((s) => !s.done).length, 1, 'started-without-result is live')
  assert.equal(m.seats.filter((s) => s.done).length, 1)
  assert.ok(m.cost > 1, 'cache reads priced')
  assert.equal(m.blackboard['red/ledger.md'], null, 'uncreated shard reported as such')
  assert.ok(m.blackboard['blue/report.md'] > 0)
})

test('dashboard html: mass line + both round rows + live seat + dark-mode roles, all escaped', () => {
  const { runDir, transcriptDir } = fixture()
  const html = renderHtml(buildModel(runDir, transcriptDir))
  assert.ok(html.includes('polyline'), 'mass trend rendered')
  assert.ok(html.includes('118.75') && html.includes('81.5'), 'both telemetry rounds in the table')
  assert.ok(html.includes('red-merge-r2'), 'live seat listed')
  assert.ok(html.includes('data-theme="dark"'), 'dark mode selected, not flipped')
  assert.ok(html.includes('prefers-color-scheme'), 'OS scheme honored')
  assert.ok(!html.includes('<script'), 'no scripts — static artifact')
})

test('dashboard degrades: no telemetry yet renders the stated absence, not a crash', () => {
  const runDir = tmp(), transcriptDir = tmp()
  const html = renderHtml(buildModel(runDir, transcriptDir))
  assert.ok(html.includes('no telemetry yet'))
})

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
  writeFileSync(join(runDir, 'red', 'ledger.md'), '# ledger\n## open\nR2-1 | high | loc | problem\nR2-2 | medium | loc | problem\n## closure index\nR1-1 | closed | fixed | -\nR1-2 | closed_with_regression | superseded | R2-1 R2-2\n')
  writeFileSync(join(runDir, 'red', 'archive.md'), '# archive\n## R1-1 — closed\nprose\n')
  writeFileSync(join(runDir, 'friction.md'), '# friction\n- seat-a: pain one\n- seat-b: pain two\n')
  // Journal uses the REAL harness schema — {type, key, agentId}, NO labels (production
  // divergence caught live 2026-07-17: the fixture had invented a label field).
  writeFileSync(join(transcriptDir, 'journal.jsonl'), [
    JSON.stringify({ type: 'started', key: 'v2:a', agentId: 'idmerge' }),
    JSON.stringify({ type: 'started', key: 'v2:b', agentId: 'idsynth' }),
    JSON.stringify({ type: 'result', key: 'v2:b', agentId: 'idsynth', result: { claim_count: 100, grade_disputes: [{ id: 'R1-1', proposed: 'medium' }] } }),
    JSON.stringify({ type: 'started', key: 'v2:c', agentId: 'idfront' }),
    JSON.stringify({ type: 'result', key: 'v2:c', agentId: 'idfront', result: 'hypotheses' }),
    // Judiciary fixtures: two red envelopes with a supersedes chain (R1-1 -> R2-1, grade
    // migrating high/high -> medium/medium), one judge sitting with mixed rulings, one
    // dispute round-trip (blue raised above, red accepts here).
    JSON.stringify({ type: 'started', key: 'v2:d', agentId: 'idred1' }),
    JSON.stringify({ type: 'result', key: 'v2:d', agentId: 'idred1', result: { verdict: 'FAIL', gaps: [{ id: 'R1-1', likelihood: 'high', impact: 'high' }, { id: 'R1-2', likelihood: 'low', impact: 'low' }] } }),
    JSON.stringify({ type: 'started', key: 'v2:e', agentId: 'idjudge' }),
    JSON.stringify({ type: 'result', key: 'v2:e', agentId: 'idjudge', result: { resolutions: [{ id: 'R1-2', resolution: 'carried' }, { id: 'R1-3', resolution: 'risk_accepted' }] } }),
    JSON.stringify({ type: 'started', key: 'v2:f', agentId: 'idred2' }),
    JSON.stringify({ type: 'result', key: 'v2:f', agentId: 'idred2', result: { verdict: 'FAIL', gaps: [{ id: 'R2-1', supersedes: ['R1-1'], likelihood: 'medium', impact: 'medium' }], dispute_responses: [{ id: 'R1-1', response: 'accepted' }] } }),
  ].join('\n') + '\n')
  // Seat identity comes from each transcript's first user message (cost-audit's method).
  writeFileSync(join(transcriptDir, 'agent-idmerge.jsonl'), [
    JSON.stringify({ timestamp: new Date(Date.now() - 300000).toISOString(), message: { role: 'user', content: 'Red merge, round 2. FIRST ACTION...' } }),
    JSON.stringify({ message: { role: 'assistant', usage: { input_tokens: 100, output_tokens: 50, cache_read_input_tokens: 1000000, cache_creation_input_tokens: 0 }, content: [] } }),
  ].join('\n') + '\n')
  writeFileSync(join(transcriptDir, 'agent-idsynth.jsonl'),
    JSON.stringify({ message: { role: 'user', content: 'Blue synthesis for topic...' } }) + '\n')
  writeFileSync(join(transcriptDir, 'agent-idfront.jsonl'),
    JSON.stringify({ message: { role: 'user', content: 'Research debate opening for topic: x. Formulate 3-5 frontier hypotheses...' } }) + '\n')
  writeFileSync(join(transcriptDir, 'agent-idred1.jsonl'),
    JSON.stringify({ message: { role: 'user', content: 'Red merge, round 1. FIRST ACTION...' } }) + '\n')
  writeFileSync(join(transcriptDir, 'agent-idjudge.jsonl'),
    JSON.stringify({ message: { role: 'user', content: 'Judge sitting, round 1. Contested docket...' } }) + '\n')
  writeFileSync(join(transcriptDir, 'agent-idred2.jsonl'),
    JSON.stringify({ message: { role: 'user', content: 'Red merge, round 2. FIRST ACTION...' } }) + '\n')
  return { runDir, transcriptDir }
}

test('dashboard model: telemetry series, live vs done seats, cost estimate, blackboard sizes', () => {
  const { runDir, transcriptDir } = fixture()
  const m = buildModel(runDir, transcriptDir)
  assert.equal(m.telemetry.length, 2)
  assert.equal(m.latest.mass, 81.5)
  assert.equal(m.seats.filter((s) => !s.done).length, 1, 'started-without-result is live')
  assert.equal(m.seats.filter((s) => s.done).length, 5)
  assert.ok(m.cost > 1, 'cache reads priced')
  assert.equal(m.shards.openRows, 2, 'ledger open rows counted')
  assert.equal(m.shards.openBySeverity.high, 1)
  assert.equal(m.shards.closureIndexRows, 2, 'LINE count — a supersedes-bearing row counts once, not per id named')
  assert.equal(m.shards.archiveRecords, 1)
  assert.equal(m.friction.count, 2, 'friction is a count of pain points, not bytes')
  assert.equal(m.blueClaims, 100)
  assert.ok(m.steps.some(s2 => s2.name === 'frontier' && s2.state === 'done'), 'progress steps derived from journal')
  assert.equal(m.rates[1].closed, 30 + 22 - 23, 'close rate math: prev_open + minted - open')
})

test('dashboard html: mass line + both round rows + live seat + dark-mode roles, all escaped', () => {
  const { runDir, transcriptDir } = fixture()
  const html = renderHtml(buildModel(runDir, transcriptDir))
  assert.ok(html.includes('polyline'), 'mass trend rendered')
  assert.ok(html.includes('118.75') && html.includes('81.5'), 'both telemetry rounds in the table')
  assert.ok(html.includes('red-merge-r2'), 'live seat listed (classified from transcript head)')
  assert.ok(/running \d+ min/.test(html), 'live seat shows elapsed minutes from transcript timestamp')
  assert.ok(html.includes('class="bar"'), 'progress bar rendered')
  assert.ok(html.includes('close rate'), 'open/close rates table')
  assert.ok(html.includes('friction entries'), 'friction as count tile')
  assert.ok(!html.includes('KB<'), 'no byte sizes as content metrics')
  assert.ok(html.includes('data-theme="dark"'), 'dark mode selected, not flipped')
  assert.ok(html.includes('prefers-color-scheme'), 'OS scheme honored')
  assert.ok(!html.includes('<script'), 'no scripts — static artifact')
})

test('judiciary: rulings tallied, chains follow supersedes edges, migrations graded by mass', () => {
  const { runDir, transcriptDir } = fixture()
  const m = buildModel(runDir, transcriptDir)
  const j = m.judiciary
  assert.equal(j.judgeSittings, 1)
  assert.deepEqual(j.rulings, { carried: 1, risk_accepted: 1 }, 'rulings tallied by type')
  assert.equal(j.disputes.raised, 1, 'blue grade_disputes counted')
  assert.equal(j.disputes.accepted, 1, 'red dispute_responses counted')
  assert.equal(j.chains, 2, 'R1-1 and R2-1 collapse into one chain via supersedes; R1-2 is its own')
  assert.deepEqual(j.chainSpans, { 1: 1, 2: 1 }, 'chain longevity spans rounds, not id lifetimes')
  assert.equal(j.migDown, 1, 'high/high -> medium/medium along the chain is a downgrade')
  assert.equal(j.migUp + j.migFlat, 0)
  const html = renderHtml(m)
  assert.ok(html.includes('Judiciary'), 'judiciary section rendered')
  assert.ok(html.includes('carried: 1') && html.includes('risk_accepted: 1'), 'ruling distribution shown')
  assert.ok(html.includes('1×1r') && html.includes('1×2r'), 'chain-longevity histogram shown')
})

test('dashboard degrades: no telemetry yet renders the stated absence, not a crash', () => {
  const runDir = tmp(), transcriptDir = tmp()
  const html = renderHtml(buildModel(runDir, transcriptDir))
  assert.ok(html.includes('no telemetry yet'))
})

test('phone UX pass: severity codes, live-seat grouping, summarized envelopes, telemetry fallback', async () => {
  const { summarizeResult } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  assert.equal(summarizeResult(JSON.stringify({ verdict: 'FAIL', citations_checked: 28, friction: ['x'] })), 'verdict FAIL · 28 citations checked · 1 friction')
  assert.equal(summarizeResult('Pass written to somewhere long...'), 'Pass written to somewhere long...')
  const { runDir, transcriptDir } = fixture()
  const html = renderHtml(buildModel(runDir, transcriptDir))
  assert.ok(html.includes('3h') && html.includes('9m'), 'severity codes compact (3h · 9m style)')
  assert.ok(html.includes('class="scrollx"'), 'rates table scrolls in its own container')
  assert.ok(html.includes('title="frontier"') && html.includes('>front<'), 'short bar labels with full-name tooltips')
  assert.ok(html.includes('100 claims'), 'completions summarize the raw envelope end-to-end (untruncated at capture)')
})

test('W1.5: telemetry is authoritative when the ledger parse under-reads the open count', () => {
  const { runDir, transcriptDir } = fixture()
  // Ledger open section parses 2 rows but telemetry says 23 open — the run-5 "open gaps: 1"
  // regression class: the old fallback engaged only at zero.
  const html = renderHtml(buildModel(runDir, transcriptDir))
  assert.ok(html.includes('23 <span class="muted">(from telemetry'), 'under-read parse defers to telemetry')
  assert.ok(html.includes('found 2 row(s)'), 'the parse count is disclosed, not hidden')
})

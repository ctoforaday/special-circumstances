// node --test — render-run-dashboard.mjs against temp-dir fixtures.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, writeFileSync, mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { buildModel, renderHtml, classifySeat, scorecardSection } from '../../skills/research-protocol/scripts/render-run-dashboard.mjs'

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

test('classifySeat: every prompt shape maps to its seat; an unknown head is `other`, never a guess', () => {
  // Seat identity is recovered by regex over prose, so the failure mode is silent
  // misattribution: a head that matches nothing must land in `other` rather than
  // in whichever branch happens to be last.
  assert.deepEqual(classifySeat('Red audit, round 3, lens: citations'), { seat: 'red-lens', round: 3 })
  assert.deepEqual(classifySeat('Red merge, round 12. FIRST ACTION'), { seat: 'red-merge', round: 12 })
  assert.deepEqual(classifySeat('Blue response, round 2.'), { seat: 'blue-respond', round: 2 })
  assert.deepEqual(classifySeat('Adjudication, round 1. Contested docket'), { seat: 'judge', round: 1 })
  assert.deepEqual(classifySeat('Terminal dispute disposition for the run'), { seat: 'judge-terminal', round: 0 })
  assert.deepEqual(classifySeat('Blue synthesis for topic x'), { seat: 'blue-synthesize', round: 0 })
  assert.deepEqual(classifySeat('Blue lane 3: prior art'), { seat: 'blue-lane', round: 0 })
  assert.deepEqual(classifySeat('Formulate 3-5 frontier hypotheses'), { seat: 'frontier', round: 0 })
  assert.deepEqual(classifySeat('Final assembly of the report'), { seat: 'assemble', round: 0 })
  assert.deepEqual(classifySeat(''), { seat: 'other', round: 0 })
  assert.deepEqual(classifySeat('some unrelated preamble'), { seat: 'other', round: 0 })
  // Round-bearing prompts win over round-0 markers that appear in their body —
  // a red merge prompt quoting "Blue synthesis" is still a red merge.
  assert.deepEqual(classifySeat('Red merge, round 4. Prior step was Blue synthesis.'), { seat: 'red-merge', round: 4 })
})

test('DEFECT: compound grades are counted as themselves, not as the simple grade they contain', () => {
  // `medium-high` contains the substring `high`, and `low-medium` contains
  // `medium`. With the grade list ordered simple-first, the substring match
  // classified every compound row as the simple grade — inflating the
  // high-severity count on the tile a human reads to decide whether the board is
  // getting worse. The list is now ordered most-specific-first.
  const runDir = tmp(), transcriptDir = tmp()
  mkdirSync(join(runDir, 'red'), { recursive: true })
  writeFileSync(join(runDir, 'red', 'ledger.md'),
    '# ledger\n## open\nR1-1 | medium-high | loc | p\nR1-2 | low-medium | loc | p\nR1-3 | high | loc | p\n## closure index\n')
  const m = buildModel(runDir, transcriptDir)
  assert.equal(m.shards.openRows, 3)
  assert.deepEqual(m.shards.openBySeverity, { 'medium-high': 1, 'low-medium': 1, high: 1 })
})

test('DEFECT: scorecard rows whose CLAUSE contains a colon reach the dashboard', () => {
  // `[^:]+` cannot backtrack past a colon, so `LOSS: additive violations` never
  // matched and the row vanished from the dashboard entirely — and that row is
  // `unrecorded_claim_loss`, a DETECTOR carrying the whole additive invariant.
  // The same defect was just repaired in scorecards.mjs; this is its twin.
  const runDir = tmp()
  mkdirSync(join(runDir, 'inputs'), { recursive: true })
  writeFileSync(join(runDir, 'inputs', 'red-scorecard.md'), [
    '# red scorecard', '',
    '## 2026-07-01_old-run', '',
    '- `stale_metric` [benchmark] — An earlier run: **99**', '',
    '## 2026-07-18_this-run', '',
    '- `unrecorded_claim_loss` [detector] — LOSS: additive violations: **2**',
    '- `round_parity_failures` [detector] — Round on the record: **0**',
    '- `finding_precision` [benchmark] — Certification: earned PASS/FAIL: **0.8**',
    '- `lines_of_inquiry` [diagnostic] — Alternatives explored: **7**',
    '- `carried_share` [benchmark] — Not a router: _not computed_ — the bench did not sit',
    '',
  ].join('\n'))
  const html = scorecardSection(runDir)

  assert.ok(html.includes('unrecorded_claim_loss'), 'the colon-bearing detector row is rendered')
  assert.ok(html.includes('LOSS: additive violations'), 'the clause is rendered WHOLE, not truncated at its colon')
  assert.ok(html.includes('Certification: earned PASS/FAIL'), 'the second colon-bearing clause too')
  assert.ok(!html.includes('stale_metric'), 'only the LATEST capture is shown — the series lives in the file')
  // A tripped detector is an alarm; a detector reading 0 is not. Class, not value,
  // drives the styling — treating a diagnostic as a target is the defect this guards.
  const rowOf = (metric) => html.split('<tr>').find((r) => r.includes(metric))
  assert.ok(rowOf('unrecorded_claim_loss').includes('#b00'), 'nonzero detector reads as an alarm')
  assert.ok(!rowOf('round_parity_failures').includes('#b00'), 'a detector at 0 has not fired')
  assert.ok(rowOf('finding_precision').includes('font-weight:700'), 'benchmarks read bold')
  assert.ok(rowOf('lines_of_inquiry').includes('opacity:.65'), 'diagnostics read muted — not a target')
  assert.ok(rowOf('carried_share').includes('not computed'), 'an uncomputed metric is stated, not dropped')
})

test('scorecardSection: absent inputs, no scorecards, and unparseable rows all render nothing (never a crash)', () => {
  // The dashboard runs as a live watcher over a run in progress, so every one of
  // these states is reached on a normal run before the first capture exists.
  assert.equal(scorecardSection(tmp()), '', 'no inputs/ yet')
  const noCards = tmp(); mkdirSync(join(noCards, 'inputs'), { recursive: true })
  assert.equal(scorecardSection(noCards), '', 'inputs/ exists but holds no scorecards')
  const junk = tmp(); mkdirSync(join(junk, 'inputs'), { recursive: true })
  writeFileSync(join(junk, 'inputs', 'blue-scorecard.md'), '# blue scorecard\n\n## run\n\nprose, no rows\n')
  assert.equal(scorecardSection(junk), '', 'a card with no parseable rows contributes no empty table')
  // Chair names and values are escaped — the file is machine-written, but it is
  // still untrusted input to an HTML document.
  const esc = tmp(); mkdirSync(join(esc, 'inputs'), { recursive: true })
  writeFileSync(join(esc, 'inputs', 'red-scorecard.md'), '## r\n\n- `m` [benchmark] — Clause: **<script>x</script>**\n')
  const out = scorecardSection(esc)
  assert.ok(!out.includes('<script>'), 'values are escaped, never injected')
  assert.ok(out.includes('&lt;script&gt;'))
})

test('W1.5: telemetry is authoritative when the ledger parse under-reads the open count', () => {
  const { runDir, transcriptDir } = fixture()
  // Ledger open section parses 2 rows but telemetry says 23 open — the run-5 "open gaps: 1"
  // regression class: the old fallback engaged only at zero.
  const html = renderHtml(buildModel(runDir, transcriptDir))
  assert.ok(html.includes('23 <span class="muted">(from telemetry'), 'under-read parse defers to telemetry')
  assert.ok(html.includes('found 2 row(s)'), 'the parse count is disclosed, not hidden')
})

// friction.md holds ATTRIBUTED LINES, not markdown bullets. The dashboard counted
// /^- / and therefore reported "none logged yet" while seven real entries sat in the
// file — including every tool failure the run hit. Fixture is the real format.
test('friction entries are counted in the format seats actually write', async () => {
  const { buildModel } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  const runDir = tmp()
  mkdirSync(join(runDir, 'trajectories'), { recursive: true })
  writeFileSync(join(runDir, 'friction.md'), [
    '# friction.md — some topic',
    '',
    'blue-synthesize: the Write tool refused to create blue/report.md, a generic subagent guard.',
    'red-merge-r1: `merge mint` has no amend path — a first mint with placeholder payload is permanent.',
    'red-merge-r2: acceptance-check contracts have no field for CONDITION-PINNING.',
    '',
  ].join('\n'))
  const m = buildModel(runDir, tmp())
  assert.equal(m.friction.count, 3, 'three attributed lines, none of them bullets')
  assert.match(m.friction.last, /CONDITION-PINNING/, 'the latest entry is the last one written')
})

test('an empty friction log still reports zero rather than throwing', async () => {
  const { buildModel } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  const runDir = tmp()
  mkdirSync(join(runDir, 'trajectories'), { recursive: true })
  writeFileSync(join(runDir, 'friction.md'), '# friction.md — topic\n\n')
  const m = buildModel(runDir, tmp())
  assert.equal(m.friction.count, 0, 'a header-only file is genuinely empty')
  assert.equal(m.friction.last, null)
})

// A run's pace depends on its model tier, its report size and how hard red is working,
// so the only honest basis for "how much longer" is THIS run's own measured seats.
test('projectCompletion ranges over measured spans and never returns a point estimate', async () => {
  const { projectCompletion } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  const t = (min) => min * 60000
  const now = t(100)
  const seats = [
    { seat: 'red-merge', label: 'red-merge-r1', done: true, startedMs: t(0), endedMs: t(11) },
    { seat: 'red-merge', label: 'red-merge-r2', done: true, startedMs: t(20), endedMs: t(37) },
    { seat: 'blue-synthesize', label: 'blue-synthesize', done: true, startedMs: t(40), endedMs: t(58) },
    { seat: 'red-merge', label: 'red-merge-r3', done: false, startedMs: t(95) },
  ]
  const p = projectCompletion(seats, now)
  // The live merge is 5 minutes in: 11-5 low, 17-5 high, plus assembly's 18 (no
  // assemble precedent, so blue-synthesize stands in).
  assert.equal(p.lowMin, 24)
  assert.equal(p.highMin, 30)
  assert.ok(p.highMin > p.lowMin, 'merge seats slowed 11 -> 17 as the board grew; one number would be wrong in a predictable direction')
  assert.match(p.basis, /3 completed seat/)
})

test('projectCompletion reports what an extra round costs, since it cannot know the ceiling', async () => {
  const { projectCompletion } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  const t = (min) => min * 60000
  const seats = [
    { seat: 'red-lens', label: 'red-lens-r1', done: true, startedMs: t(0), endedMs: t(4) },
    { seat: 'red-merge', label: 'red-merge-r1', done: true, startedMs: t(4), endedMs: t(15) },
    { seat: 'blue-respond', label: 'blue-respond-r1', done: true, startedMs: t(15), endedMs: t(20) },
  ]
  const p = projectCompletion(seats, t(20))
  assert.equal(p.perRoundLowMin, 20, 'lens + merge + respond is what one more round adds')
  assert.equal(p.perRoundHighMin, 20)
})

// An estimate that quietly omits an unmeasured phase reads as more confident than it is.
test('projectCompletion names the phases it has no precedent for', async () => {
  const { projectCompletion } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  const t = (min) => min * 60000
  const p = projectCompletion([{ seat: 'red-lens', label: 'red-lens-r1', done: false, startedMs: t(0) }], t(1))
  assert.ok(p.unmeasured.includes('red-lens-r1'), 'a live seat with no completed sibling cannot be projected')
  assert.ok(p.unmeasured.includes('assembly'), 'and neither can assembly on a first run')
})

test('projectCompletion reports complete once assembly is done', async () => {
  const { projectCompletion } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  const t = (min) => min * 60000
  const p = projectCompletion([{ seat: 'assemble', label: 'assemble', done: true, startedMs: t(0), endedMs: t(10) }], t(11))
  assert.equal(p.state, 'complete')
  assert.equal(p.highMin, 0)
})

// The dashboard kept its own single price row at fable tier and applied it to every
// seat regardless of model. With haiku bulk seats that read $189.59 against a true
// $53.92 — a 3.5x overstatement on the number a human watches while deciding whether
// to let an expensive run keep going. Both modules now price from one table.
test('the dashboard prices each seat at its own model rate, agreeing with the cost audit', async () => {
  const { buildModel } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  const { PRICES, tier } = await import('../../skills/research-protocol/scripts/cost-audit.mjs')
  const runDir = tmp(), tdir = tmp()
  mkdirSync(join(runDir, 'trajectories'), { recursive: true })
  const usage = { input_tokens: 1000, output_tokens: 2000, cache_read_input_tokens: 1e6, cache_creation_input_tokens: 5e4 }
  writeFileSync(join(tdir, 'journal.jsonl'), JSON.stringify({ agentId: 'a1', result: 'ok' }) + '\n')
  writeFileSync(join(tdir, 'agent-a1.jsonl'), [
    JSON.stringify({ timestamp: '2026-07-19T00:00:00Z', message: { role: 'user', content: 'Blue lane 1' } }),
    JSON.stringify({ message: { model: 'claude-haiku-4-5-20251001', usage } }),
  ].join('\n') + '\n')

  const [pin, pout, pcr, pcw] = PRICES[tier('claude-haiku-4-5-20251001')]
  const expected = (usage.input_tokens * pin + usage.output_tokens * pout +
    usage.cache_read_input_tokens * pcr + usage.cache_creation_input_tokens * pcw) / 1e6
  const m = buildModel(runDir, tdir)
  assert.ok(Math.abs(m.cost - expected) < 1e-9, `dashboard ${m.cost} should equal haiku-priced ${expected}`)

  const fableRow = PRICES.fable
  const asFable = (usage.input_tokens * fableRow[0] + usage.output_tokens * fableRow[1] +
    usage.cache_read_input_tokens * fableRow[2] + usage.cache_creation_input_tokens * fableRow[3]) / 1e6
  assert.ok(m.cost < asFable / 2, 'haiku traffic must not be billed at the dearest tier')
})

// An unknown model must over-report: a silently cheap estimate is the failure nobody
// notices, and the cost tile is what a human uses to decide whether to stop a run.
test('an unrecognised model is priced at the dearest tier, never the cheapest', async () => {
  const { buildModel } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  const { PRICES } = await import('../../skills/research-protocol/scripts/cost-audit.mjs')
  const runDir = tmp(), tdir = tmp()
  mkdirSync(join(runDir, 'trajectories'), { recursive: true })
  const usage = { input_tokens: 0, output_tokens: 0, cache_read_input_tokens: 1e6, cache_creation_input_tokens: 0 }
  writeFileSync(join(tdir, 'journal.jsonl'), JSON.stringify({ agentId: 'a1', result: 'ok' }) + '\n')
  writeFileSync(join(tdir, 'agent-a1.jsonl'), [
    JSON.stringify({ timestamp: '2026-07-19T00:00:00Z', message: { role: 'user', content: 'Blue lane 1' } }),
    JSON.stringify({ message: { model: 'claude-something-unreleased', usage } }),
  ].join('\n') + '\n')
  const m = buildModel(runDir, tdir)
  assert.ok(Math.abs(m.cost - PRICES.fable[2]) < 1e-9, 'unknown model falls back to the fable row')
})

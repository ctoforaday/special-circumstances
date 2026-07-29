// scorecards.mjs — W2h's measurement layer. Eight of its ten exports shipped with
// no test at all, and the two defects below were both live at the first real run:
// a metric that could never reach a seat, and an end-of-input anchor JavaScript
// does not have. Both are invisible to a reader and obvious to a test.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, mkdirSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import {
  row, readTelemetry, blueRows, redRows, citationYieldByRole, bucketFindingsByRole,
  benchRows, computeScorecards, renderChair, headline, chairHeader, readResults,
  computeDirectionUptake, computeAnchoredClosures,
} from '../../skills/research-protocol/scripts/scorecards.mjs'

const tmp = () => mkdtempSync(join(tmpdir(), 'sc-'))
const runWith = (files = {}) => {
  const d = tmp()
  for (const [rel, body] of Object.entries(files)) {
    const p = join(d, rel)
    mkdirSync(join(p, '..'), { recursive: true })
    writeFileSync(p, body)
  }
  return d
}

// ---- the contract row() enforces ----

test('row refuses an unknown class and an uncomputed row with no reason', () => {
  assert.throws(() => row({ clause: 'c', metric: 'm', cls: 'nonsense' }), /unknown class/)
  assert.throws(() => row({ clause: 'c', metric: 'm', cls: 'benchmark' }), /no value and no reason/)
  // Zero is a value, not an absence — the tripwire classes depend on this.
  assert.equal(row({ clause: 'c', metric: 'm', cls: 'detector', value: 0 }).value, 0)
})

// ---- readTelemetry ----

test('readTelemetry survives a torn last line and a missing file', () => {
  assert.deepEqual(readTelemetry(tmp()), [], 'absent telemetry is empty, not a throw')
  const d = runWith({ 'trajectories/board-telemetry.jsonl': '{"round":1}\n{"round":2}\n{"round":3' })
  const t = readTelemetry(d)
  assert.equal(t.length, 2, 'a half-written final line is dropped, not fatal — runs get killed mid-append')
})

// ---- blue ----

test('blueRows: unrecorded claim loss nets retires against the fall', () => {
  const results = [
    { claim_count: 100 }, { claim_count: 90, retired: [{ id: 'c1' }, { id: 'c2' }] }, { claim_count: 95 },
  ]
  const rows = blueRows(runWith(), results, [])
  const loss = rows.find((r) => r.metric === 'unrecorded_claim_loss')
  assert.equal(loss.value, 8, '10 lost, 2 retired on the record, 8 unaccounted')
  assert.equal(loss.cls, 'detector')
})

test('blueRows: a single round cannot compute loss and says so', () => {
  const rows = blueRows(runWith(), [{ claim_count: 10 }], [])
  const loss = rows.find((r) => r.metric === 'unrecorded_claim_loss')
  assert.equal(loss.value, null)
  assert.ok(/two rounds/.test(loss.note), 'an uncomputed row states its precondition')
})

test('blueRows: round parity counts claimed-but-unattested, not absent fields', () => {
  const rows = blueRows(runWith(), [
    { round_record_appended: true }, { round_record_appended: false }, {},
  ], [])
  const p = rows.find((r) => r.metric === 'round_parity_failures')
  assert.equal(p.value, 1, 'the envelope with no attestation field is not a failure — it is silence')
})

test('blueRows: a thin declined-avenue reason is caught; a pursued one is not judged', () => {
  const rows = blueRows(runWith(), [{
    avenues: [
      { line: 'a', status: 'declined', reason: 'no' },
      { line: 'b', status: 'declined', reason: 'the corpus does not contain per-claim confidence records, so this cannot be computed' },
      { line: 'c', status: 'pursued', reason: 'ok' },
    ],
  }], [])
  const thin = rows.find((r) => r.metric === 'thin_avenue_reasons')
  assert.equal(thin.value, 1, 'only the shrug counts')
  const loi = rows.find((r) => r.metric === 'lines_of_inquiry')
  assert.equal(loi.cls, 'diagnostic', 'counting avenues must never be a benchmark — it invites avenue-padding')
  assert.deepEqual(loi.value, { declined: 2, pursued: 1 })
})

// ---- red ----

test('bucketFindingsByRole buckets the findings view by ROLE, not by position', () => {
  const y = bucketFindingsByRole([
    { role: 'L1', round: 1, seat_id: 'red-lens-r1-L1' },
    { role: 'L1', round: 1, seat_id: 'red-lens-r1-L1' },
    { role: 'L5', round: 1, seat_id: 'red-lens-r1-L5' },
    { role: 'L6', round: 1, seat_id: 'red-lens-r1-L6' },
  ])
  assert.equal(y[1].citation, 2); assert.equal(y[1].logic, 1); assert.equal(y[1].darkside, 1)
  assert.deepEqual(y[1].seats, { citation: 1, logic: 1, darkside: 1 }, 'one seat per role dispatched')
  assert.deepEqual(y[1].per_seat, { citation: 2, logic: 1, darkside: 1 })
  assert.equal(bucketFindingsByRole([]), null, 'no findings is null, not an empty answer')
  assert.equal(citationYieldByRole(tmp(), undefined), null, 'no --bin is null (same "not computed" as no findings)')
})

// anchored_closures_pct now reads the board view's closure anchors, not archive.md. A closure
// is anchored iff it carries the full seat|tool|target triple OR a carried-from; a partial or
// absent anchor is not. computeAnchoredClosures is the pure kernel; the fixture is the board
// view's `closed[]` shape (absent anchor keys omitted, as payloadMap leaves them).
test('computeAnchoredClosures: full triple and carried-from count; partial and absent do not', () => {
  const board = { closed: [
    { id: 'R1-1', closure: { anchor_seat: 'red-lens-1', anchor_tool: 'Grep', anchor_target: 'report.md' } }, // anchored (triple)
    { id: 'R1-2', closure: { carried_from: '1' } },                                                            // anchored (carried)
    { id: 'R1-3', closure: { anchor_seat: 'red-lens-1', anchor_target: 'report.md' } },                        // partial — NOT anchored
    { id: 'R1-4', closure: {} },                                                                               // no anchor — NOT
    { id: 'R1-5', closure: null },                                                                             // no closure payload — NOT
  ] }
  assert.deepEqual(computeAnchoredClosures(board), { anchored: 2, total: 5 })
  // The classic 1-of-3: triple, undefined, none → 33%.
  const three = { closed: [
    { closure: { anchor_seat: 'L1', anchor_tool: 'Grep', anchor_target: 'report.md' } },
    { closure: {} },
    { closure: null },
  ] }
  assert.equal(Math.round(computeAnchoredClosures(three).anchored / computeAnchoredClosures(three).total * 100), 33)
})

test('redRows anchored_closures_pct: without --bin it says it needs the tool, not a stub 0', () => {
  const noBin = redRows(runWith(), [], [], undefined).find((r) => r.metric === 'anchored_closures_pct')
  assert.equal(noBin.value, null, 'no tool binary -> not computed')
  assert.match(noBin.note, /needs the tool|board view/, 'the row states it needs the record, not archive.md')
})

// ---- bench ----

test('benchRows: carried_share is the not-a-router scoreboard', () => {
  const rows = benchRows(runWith(), [{
    resolutions: [
      { resolution: 'carried' }, { resolution: 'carried' },
      { resolution: 'granted', principle: 'p', tension: 't', review_flag: false },
    ],
  }])
  assert.equal(rows.find((r) => r.metric === 'carried_share').value, 0.67)
  assert.equal(rows.find((r) => r.metric === 'rulings_without_opinion').value, 2, 'a ruling stating no principle is a disposition wearing a ruling name')
})

// Direction-uptake reads the debate VIEW (structured), so a blue section is a whole string
// the tool returns — the JS-regex `\Z`-truncation defect (which once cut the last blue section
// at a literal 'Z' and dropped the citing words after it) cannot recur. computeDirectionUptake
// is the pure kernel; the fixture is the debate view's shape.
test('computeDirectionUptake: a blue section citing the direction counts in full, past any "Zero"', () => {
  const debate = { rounds: [{
    round: 1,
    red: [],
    blue: ['Zero regressions this round; acted on the judge direction as carried.'],
    lead: [{ gap_id: 'R1-1', disposition: 'carried' }],
  }] }
  assert.deepEqual(computeDirectionUptake(debate), { leadSections: 1, blueCitesLead: 1 },
    'the section cites the direction AFTER the word "Zero" — the whole string is read, not truncated')
  // A round with a LEAD sitting but a blue section that does not reference the bench.
  const noCite = { rounds: [{ round: 1, red: [], blue: ['unrelated repair notes'], lead: [{ gap_id: 'R1-1', disposition: 'carried' }] }] }
  assert.deepEqual(computeDirectionUptake(noCite), { leadSections: 1, blueCitesLead: 0 })
})

test('benchRows direction-uptake: with a view reader it computes; without --bin it says it needs the tool', () => {
  // With a reader (bin), benchRows spawns the tool; here we cannot spawn in-test, so the
  // pure-kernel path above covers the computed case. Without bin, the row must be uncomputed
  // WITH a reason — never a false "no LEAD sections" derived from the debate.md stub.
  const noBin = benchRows(runWith(), [], undefined).find((r) => r.metric === 'blue_sections_citing_direction')
  assert.equal(noBin.value, null, 'no tool binary -> not computed')
  assert.match(noBin.note, /needs the tool|debate view/, 'the row states it needs the record, not a stub')
})

test('benchRows: an empty bench states it rather than dividing by zero', () => {
  const rows = benchRows(runWith(), [{}])
  assert.equal(rows.find((r) => r.metric === 'carried_share').value, null)
  assert.equal(rows.find((r) => r.metric === 'petitions_filed').value, 0, 'a measure with nothing to record is 0, not absent')
})

// ---- render / headline ----

test('headline ranks nonzero detectors first and drops uncomputed rows', () => {
  const rows = [
    row({ clause: 'a', metric: 'benchy', cls: 'benchmark', value: 1 }),
    row({ clause: 'b', metric: 'quiet_tripwire', cls: 'detector', value: 0 }),
    row({ clause: 'c', metric: 'live_tripwire', cls: 'detector', value: 3 }),
    row({ clause: 'd', metric: 'blocked', cls: 'benchmark', note: 'BLOCKED until W2f' }),
  ]
  const h = headline(rows)
  assert.ok(h[0].startsWith('live_tripwire'), 'a tripped detector outranks a benchmark')
  assert.ok(!h.join(' ').includes('quiet_tripwire'), 'a detector reading zero is not news')
  assert.ok(!h.join(' ').includes('blocked'), 'an uncomputed row is never a headline')
})

// THE DEFECT: renderChair emits real clauses that contain a colon
// ('LOSS: additive violations'). Setup parses the rendered file to build the
// prompt headlines, and its parser stopped at the first colon after the class
// tag — so that metric could never reach the seat it governs. It is a DETECTOR,
// the class headline() ranks first, and the whole enforcement of the additive
// invariant. The scorecard was written, mirrored, and silently unreadable.
test('renderChair output stays parseable when a clause contains a colon', async () => {
  const rows = [row({ clause: 'LOSS: additive violations', metric: 'unrecorded_claim_loss', cls: 'detector', value: 4, note: '6 lost, 2 retired' })]
  const rendered = renderChair('blue', rows, 'run-6')
  const { mirrorScorecards } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const mem = tmp(), runDir = tmp()
  mkdirSync(join(runDir, 'inputs'), { recursive: true })
  writeFileSync(join(mem, 'blue-scorecard.md'), chairHeader('blue') + rendered)
  const r = mirrorScorecards(mem, runDir)
  assert.ok(r.written)
  assert.ok(r.headlines.blue, 'the chair produced a headline at all')
  assert.ok(
    r.headlines.blue.some((h) => h.startsWith('unrecorded_claim_loss 4')),
    'a colon in the CLAUSE must not hide the metric from the seat prompt',
  )
})

test('renderChair marks an uncomputed row with its reason instead of a bare blank', () => {
  const out = renderChair('blue', [row({ clause: 'Calibration', metric: 'confidence_vs_survival', cls: 'benchmark', note: 'BLOCKED until W2f' })], 'run-6')
  assert.ok(/_not computed_ — BLOCKED until W2f/.test(out))
  assert.ok(!/undefined/.test(out), 'no rendered scorecard line may contain the word undefined')
})

test('computeScorecards returns all three chairs even on an empty run', () => {
  const s = computeScorecards(runWith(), [])
  assert.deepEqual(Object.keys(s), ['blue', 'red', 'bench'])
  for (const [chair, rows] of Object.entries(s)) {
    assert.ok(rows.length, `${chair} produced rows`)
    for (const r of rows) assert.ok(r.value !== null || r.note, `${chair}.${r.metric} is either computed or explains itself`)
  }
})

test('chairHeader states every class, so a reader knows which numbers not to optimize', () => {
  const h = chairHeader('red')
  for (const c of ['BENCHMARK', 'DIAGNOSTIC', 'DETECTOR', 'MEASURE']) assert.ok(h.includes(c), `${c} documented`)
})

// The emitted HEADLINE is the single authority. Setup used to re-derive it from the
// rendered rows in FILE order, so a tripped detector could lose its prompt slot to a
// benchmark that merely appeared above it — the ranking headline() exists to impose.
test('the seat prompt gets headline()-RANKED numbers, not file order', async () => {
  const { mirrorScorecards } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const rows = [
    row({ clause: 'A', metric: 'bench_one', cls: 'benchmark', value: 1 }),
    row({ clause: 'B', metric: 'bench_two', cls: 'benchmark', value: 2 }),
    row({ clause: 'C', metric: 'bench_three', cls: 'benchmark', value: 3 }),
    row({ clause: 'LOSS: additive violations', metric: 'tripped', cls: 'detector', value: 7 }),
  ]
  const mem = tmp(), runDir = tmp()
  mkdirSync(join(runDir, 'inputs'), { recursive: true })
  writeFileSync(join(mem, 'blue-scorecard.md'), chairHeader('blue') + renderChair('blue', rows, 'run-6'))
  const r = mirrorScorecards(mem, runDir)
  assert.ok(r.headlines.blue[0].startsWith('tripped 7'), 'the detector leads even though three benchmarks precede it in the file')
  assert.equal(r.headlines.blue.length, 3)
})

// Scorecards written by an older capture have no HEADLINE line; the fallback parser
// must still read them, colons and all.
test('a pre-HEADLINE scorecard still yields a prompt headline', async () => {
  const { mirrorScorecards } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const mem = tmp(), runDir = tmp()
  mkdirSync(join(runDir, 'inputs'), { recursive: true })
  writeFileSync(join(mem, 'red-scorecard.md'), [
    '# red scorecard', '', '## run-5', '',
    '- `unrecorded_claim_loss` [detector] — LOSS: additive violations: **4** (6 lost, 2 retired)',
    '- `anchored_closures_pct` [benchmark] — Attestation-format invariant: **89**',
    '',
  ].join('\n'))
  const r = mirrorScorecards(mem, runDir)
  assert.ok(r.headlines.red.some((h) => h.startsWith('unrecorded_claim_loss 4')), 'the colon-bearing clause is readable by the fallback')
  assert.ok(r.headlines.red.some((h) => h.startsWith('anchored_closures_pct 89')))
})

// ---- one parser, one authority ----

test('parseRenderedRows reads a colon-bearing clause and distinguishes zero from unmeasured', async () => {
  const { parseRenderedRows } = await import('../../skills/research-protocol/scripts/scorecards.mjs')
  const rows = [
    row({ clause: 'LOSS: additive violations', metric: 'unrecorded_claim_loss', cls: 'detector', value: 0 }),
    row({ clause: 'Certification: earned PASS/FAIL', metric: 'finding_precision', cls: 'benchmark', note: 'needs adjudication outcomes' }),
    row({ clause: 'Petition handling', metric: 'petitions_filed', cls: 'measure', value: 2 }),
  ]
  const parsed = parseRenderedRows(renderChair('red', rows, 'run-6'))
  const byMetric = Object.fromEntries(parsed.map((r) => [r.metric, r]))
  assert.equal(byMetric.unrecorded_claim_loss.value.trim(), '0', 'a detector reading zero is MEASURED, not missing')
  assert.equal(byMetric.unrecorded_claim_loss.clause, 'LOSS: additive violations', 'the clause survives its own colon intact')
  assert.equal(byMetric.finding_precision.value, null, 'an uncomputed row parses to null, never to a string')
  assert.ok(byMetric.petitions_filed, 'measure rows are parsed; excluding them is the caller\'s choice, not the parser\'s')
})

// STRUCTURAL GUARD. The colon defect was found three times in three private copies
// of this regex — setup's headline mirror, the dashboard's table, and the renderer
// itself. Fixing it three times would only have reset the clock, so the parser now
// lives in one place and this test fails if a fourth copy appears.
test('no module re-implements the scorecard row parser', async () => {
  const { readdirSync, readFileSync } = await import('node:fs')
  const dir = new URL('../../skills/research-protocol/scripts/', import.meta.url)
  const offenders = []
  for (const f of readdirSync(dir).filter((f) => /\.(mjs|js)$/.test(f))) {
    if (f === 'scorecards.mjs') continue // the owner of the format may parse it
    const body = readFileSync(new URL(f, dir), 'utf8')
    if (/\[\(benchmark\|detector\|diagnostic/.test(body)) offenders.push(f)
  }
  assert.deepEqual(offenders, [], 'import parseRenderedRows from scorecards.mjs instead of re-deriving the row format')
})

// ---- one seat table ----

// The dashboard and the cost audit each carried a private copy of the seat
// classifier, and they had already drifted: cost-audit lacked the
// `Terminal dispute disposition` case, so those transcripts were filed as `other`
// and their spend was misattributed in the report the bench's economics decisions
// rest on. Both now re-export one table, and this test fails if they diverge again.
test('the dashboard and the cost audit classify every seat identically', async () => {
  const { classifySeat, KNOWN_SEATS } = await import('../../skills/research-protocol/scripts/seat-classify.mjs')
  const { classifySeat: dash } = await import('../../skills/research-protocol/scripts/render-run-dashboard.mjs')
  const { classifyTranscript: cost } = await import('../../skills/research-protocol/scripts/cost-audit.mjs')
  const heads = [
    'Red audit, round 3 — verify at the leaf',
    'Red merge, round 2', 'Blue response, round 4', 'Adjudication, round 1',
    'Terminal dispute disposition', 'Petition sitting, topic "x"', 'Blue synthesis', 'Blue lane 2',
    'frontier hypotheses', 'Final assembly', 'something unrecognised',
  ]
  for (const h of heads) {
    assert.deepEqual(dash(h), classifySeat(h), `dashboard agrees on: ${h}`)
    assert.deepEqual(cost(h), classifySeat(h), `cost audit agrees on: ${h}`)
  }
  assert.deepEqual(cost('Terminal dispute disposition'), { seat: 'judge-terminal', round: 0 },
    'the case cost-audit was missing — its absence misattributed real spend')
  const named = new Set(heads.map((h) => classifySeat(h).seat))
  for (const s of KNOWN_SEATS) assert.ok(named.has(s), `${s} is exercised by this test`)
})

test('a seat table entry carries its round; an unrounded seat reports round 0', async () => {
  const { classifySeat } = await import('../../skills/research-protocol/scripts/seat-classify.mjs')
  assert.deepEqual(classifySeat('Red audit, round 11'), { seat: 'red-lens', round: 11 }, 'a two-digit round is not truncated')
  assert.deepEqual(classifySeat('Blue synthesis'), { seat: 'blue-synthesize', round: 0 })
  assert.deepEqual(classifySeat(''), { seat: 'other', round: 0 }, 'empty input is other, not a throw')
  assert.deepEqual(classifySeat(undefined), { seat: 'other', round: 0 }, 'undefined is other, not a throw')
})

// W2i dispatches FEWER citation lenses in later rounds. A raw round-over-round
// comparison therefore measures the CUT as though it were the collapse that
// justified the cut. This fixture is the first live run's actual shape: citation
// 3 findings over 2 seats in r1, 1 over 1 seat in r2. Raw reads -67%; per seat it
// is -33%, while L5 (one seat in both rounds) really did fall -67%.
test('bucketFindingsByRole reports per-seat yield, so a dispatch cut is not read as a collapse', () => {
  const f = (role, round, n) => Array.from({ length: n }, () => ({ role, round, seat_id: `red-lens-r${round}-${role}` }))
  const y = bucketFindingsByRole([
    ...f('L1', 1, 2), ...f('L2', 1, 1), ...f('L5', 1, 9), // r1: citation across 2 seats (L1,L2), logic 9 on 1 seat
    ...f('L1', 2, 1), ...f('L5', 2, 3),                   // r2: citation 1 on 1 seat, logic 3 on 1 seat
  ])
  assert.equal(y[1].citation, 3); assert.equal(y[1].seats.citation, 2)
  assert.equal(y[2].citation, 1); assert.equal(y[2].seats.citation, 1)
  assert.equal(y[1].per_seat.citation, 1.5, 'r1 citation: 3 findings over 2 dispatched seats')
  assert.equal(y[2].per_seat.citation, 1, 'r2 citation: 1 over 1 — a 33% decline, not the 67% the raw count shows')
  assert.equal(y[1].per_seat.logic, 9)
  assert.equal(y[2].per_seat.logic, 3, 'L5 held one seat in both rounds, so its fall is real and unconfounded')
  assert.equal(y[1].per_seat.darkside, null, 'a role that did not sit reports null, never 0 — absence is not a measurement')
})

// The in-run self-read (priors-are-poison half-2): a chair reads its OWN scorecard for THIS
// run before its docket. Mid-run there is no journal, so envelope rows read "not computed"
// and the file-derived headline still computes — never a stale cross-run number.
test('in-run scorecard: file-derived headline computes without a journal; a journal parses when present', () => {
  const d = tmp()
  mkdirSync(join(d, 'records', 'render-shadow'), { recursive: true })
  writeFileSync(join(d, 'records', 'render-shadow', 'board-telemetry.jsonl'),
    JSON.stringify({ round: 2, repair_regression: { closures: 2, lineage_mints: 1, ratio: 0.5 } }) + '\n')

  // No journal yet (mid-run) — not an error.
  assert.deepEqual(readResults(d), [], 'no journal mid-run yields empty results, not a throw')
  const cards = computeScorecards(d, readResults(d))
  const repair = cards.blue.find((r) => r.metric === 'repair_regression_ratio')
  assert.equal(repair.value, 0.5, 'blue reads its headline from telemetry alone, mid-run')
  const manifest = cards.blue.find((r) => r.metric === 'manifest_coverage')
  assert.equal(manifest.value, 0, 'an envelope-derived row is present (0/not-computed), never a stale number')

  // A journal, once assembled, is parsed into results — junk lines skipped.
  mkdirSync(join(d, 'trajectories'), { recursive: true })
  writeFileSync(join(d, 'trajectories', 'journal.jsonl'),
    JSON.stringify({ result: { claim_count: 20 } }) + '\nnot json\n' + JSON.stringify({ result: { claim_count: 22 } }) + '\n')
  const parsed = readResults(d)
  assert.equal(parsed.length, 2, 'readResults collects result objects and skips unparseable lines')
  assert.equal(parsed[1].claim_count, 22)
})

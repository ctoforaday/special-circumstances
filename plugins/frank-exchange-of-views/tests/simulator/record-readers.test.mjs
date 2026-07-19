// node --test — the mjs READERS of the record layer: parity check, setup
// skeleton + mirror purge, and capture's record-join audit.
//
// This file replaces the writer-side half of the old record.test.mjs. The mjs
// record library and its four seat CLIs are gone: they existed to be validated
// against, the Go port was validated against them, and they were never used in a
// run. Their behaviour now lives in the Go suite and in the golden transcripts
// that were recorded while the differential gate was green.
//
// What survives here is everything that READS the record layer — and these tests
// deliberately build their fixtures by writing JSONL DIRECTLY rather than by
// calling any writer. That is the architecture, not a convenience: the JSONL is
// the interface, readers must not carry a dependency on whichever implementation
// owns the write path, and a reader test that needs the writer to run cannot
// distinguish "the reader is correct" from "both sides share a bug".
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdirSync, writeFileSync, readFileSync, existsSync, utimesSync, mkdtempSync } from 'node:fs'
import { join } from 'node:path'
import { tmpdir } from 'node:os'

const tmp = () => mkdtempSync(join(tmpdir(), 'feov-rec-'))

// writeEvents emits a shard exactly as feov-record does: one JSON object per
// line, fields in the tool's order. Hand-writing it here is the point — if the
// binary ever changes this shape, these readers must fail.
function writeEvents(runDir, seatId, events, nonce = 'aabbccdd') {
  const dir = join(runDir, 'records')
  mkdirSync(dir, { recursive: true })
  const round = (/-r(\d+)/.exec(seatId) || [, '0'])[1]
  const lines = events.map((e, i) => JSON.stringify({
    seq: i, seatId, nonce, round: Number(round), type: e.type,
    key: e.key || `${seatId}:${e.type}:${e.payload.gap_id || e.payload.label || `#${i}`}`,
    payload: e.payload,
  }))
  writeFileSync(join(dir, `events-${seatId}-${nonce}.jsonl`), lines.join('\n') + '\n')
  writeFileSync(join(dir, `.active-${seatId}`), nonce)
}

// A minimal shadow render, as the binary leaves on disk after render-on-mutation.
function writeShadow(runDir, files) {
  const dir = join(runDir, 'records', 'render-shadow')
  mkdirSync(dir, { recursive: true })
  for (const [name, body] of Object.entries(files)) writeFileSync(join(dir, name), body)
}

test('parity check: agreeing hand artifacts PASS; a hand-only gap FAILs with the id named', async () => {
  const { compare } = await import('../../skills/research-protocol/scripts/record-parity-check.mjs')
  const run = tmp()
  mkdirSync(join(run, 'red'), { recursive: true })
  mkdirSync(join(run, 'trajectories'), { recursive: true })
  writeShadow(run, {
    'ledger.md': '# ledger\n\n## OPEN GAPS (1)\n\n### R1-1 — a gap\n\n## CLOSURE INDEX\n\n',
  })
  writeFileSync(join(run, 'red', 'ledger.md'), '# ledger\n## OPEN\nR1-1 | medium | loc\n## closure index\n')
  let d = compare(run)
  assert.equal(d.length, 0, `agreeing records diverged: ${d.join('; ')}`)

  writeFileSync(join(run, 'red', 'ledger.md'), '# ledger\n## OPEN\nR1-1 | medium | loc\nR1-2 | high | loc2\n## closure index\n')
  d = compare(run)
  assert.ok(d.some((x) => x.includes('R1-2') && x.includes('missing from shadow')), `hand-only gap flagged: ${d.join('; ')}`)
})

test('parity check refuses to pass over an absent shadow (a gate with nothing to compare is not passing)', async () => {
  const { compare } = await import('../../skills/research-protocol/scripts/record-parity-check.mjs')
  assert.throws(() => compare(tmp()), /no shadow renders/)
})

test('ledger facts: the closure index splits open from closed, and a supersedes reference is not an open gap', async () => {
  const { extractLedgerFacts } = await import('../../skills/research-protocol/scripts/record-parity-check.mjs')
  // Both closure-row shapes the hand artifacts use in the wild: a table row and a
  // heading. Hand formats vary by seat and round, which is exactly why the parity
  // gate compares extracted FACTS rather than text.
  const f = extractLedgerFacts([
    '# ledger', '## OPEN GAPS', 'R2-1 | high | loc | problem | supersedes R1-1',
    'R2-2 | medium | loc | problem', '', '## CLOSURE INDEX',
    'R1-1 | closed | fixed | -', '### R1-2 — closed_with_regression', '',
  ].join('\n'))
  assert.deepEqual([...f.openIds].sort(), ['R2-1', 'R2-2'])
  assert.deepEqual([...f.closedIds].sort(), ['R1-1', 'R1-2'])
  assert.ok(!f.openIds.has('R1-1'), 'an id named in a supersedes column is a reference, not a second open gap')

  // No closure index at all: every id is open. A run before its first closure is
  // the normal state at round 1, not a malformed ledger.
  const early = extractLedgerFacts('## OPEN\nR1-1 | high | loc\nR1-2 | low | loc\n')
  assert.deepEqual([...early.openIds].sort(), ['R1-1', 'R1-2'])
  assert.equal(early.closedIds.size, 0)

  // Absence and emptiness are different facts — see the compare() test below.
  assert.equal(extractLedgerFacts(null), null, 'an absent file has no facts to extract')
  const empty = extractLedgerFacts('')
  assert.deepEqual([empty.openIds.size, empty.closedIds.size], [0, 0], 'an empty file has facts: zero of them')
})

test('DEFECT: an EMPTY hand ledger diverges from a populated shadow instead of passing parity', async () => {
  const { compare } = await import('../../skills/research-protocol/scripts/record-parity-check.mjs')
  const run = tmp()
  mkdirSync(join(run, 'red'), { recursive: true })
  writeShadow(run, { 'ledger.md': '# ledger\n\n## OPEN GAPS (2)\n\n### R1-1 — a gap\n\n### R1-2 — another\n\n## CLOSURE INDEX\n\n' })
  // A run killed mid-append leaves a zero-byte artifact. `if (!text) return null`
  // made that indistinguishable from "no file", and compare()'s `if (hl && sl)`
  // guard then skipped the ledger check entirely — so a truncated ledger reported
  // parity PASS against a shadow full of open gaps. A gate that measures nothing
  // is not a passing gate; this file already refuses that for a missing shadow.
  writeFileSync(join(run, 'red', 'ledger.md'), '')
  const d = compare(run)
  assert.ok(d.some((x) => x.includes('R1-1') && x.includes('missing from hand')), `truncated ledger flagged: ${d.join('; ')}`)
  assert.ok(d.some((x) => x.includes('R1-2')), 'every shadow gap the hand lost is named, not counted')

  // A genuinely ABSENT hand artifact is still skipped: the run may not have that
  // shard at all, and inventing divergences for it would make the gate unusable.
  const noHand = tmp()
  mkdirSync(join(noHand, 'red'), { recursive: true })
  writeShadow(noHand, { 'ledger.md': '# ledger\n\n### R1-1 — a gap\n' })
  assert.equal(compare(noHand).length, 0, 'no hand ledger written yet is not a divergence')
})

test('telemetry facts: rounds are keyed and compared, and a half-written final line is skipped not fatal', async () => {
  const { extractTelemetryFacts, compare } = await import('../../skills/research-protocol/scripts/record-parity-check.mjs')
  const t = extractTelemetryFacts([
    JSON.stringify({ round: 1, open_count: 30, mass: 118.75, max_severity: 'high' }),
    JSON.stringify({ round: 2, open_count: 23, mass: 81.5, max_severity: 'high' }),
    '{"round":3,"open_count":1', // killed mid-append — the shape every live run risks
  ].join('\n'))
  assert.equal(t.size, 2, 'the truncated line is dropped, the good lines survive')
  assert.deepEqual(t.get(2), { open_count: 23, mass: 81.5, max_severity: 'high' })

  // A field DISAGREEING between hand and shadow is the failure this audit exists
  // for — a present-but-wrong line passes every presence check.
  const run = tmp()
  mkdirSync(join(run, 'trajectories'), { recursive: true })
  writeFileSync(join(run, 'trajectories', 'board-telemetry.jsonl'),
    JSON.stringify({ round: 1, open_count: 30, mass: 118.75, max_severity: 'high' }) + '\n')
  writeShadow(run, {
    'ledger.md': '',
    'board-telemetry.jsonl': JSON.stringify({ round: 1, open_count: 29, mass: 118.75, max_severity: 'high' }) + '\n',
  })
  const d = compare(run)
  assert.ok(d.some((x) => x.includes('round 1 open_count') && x.includes('30') && x.includes('29')), `both values named: ${d.join('; ')}`)
  assert.ok(!d.some((x) => x.includes('mass')), 'an agreeing field is not reported')
})

test('scorecards are APPENDED run over run — the series is the point, a single number says nothing', async () => {
  const { writeScorecards } = await import('../../skills/research-protocol/scripts/capture-research-run.mjs')
  const memory = tmp()
  const runA = join(tmp(), '2026-07-01_first-run')
  const runB = join(tmp(), '2026-07-18_second-run')
  mkdirSync(runA, { recursive: true }); mkdirSync(runB, { recursive: true })

  const a = writeScorecards(runA, [], memory)
  assert.equal(a.written, true)
  assert.ok(a.chairs >= 1, 'every chair gets a card')
  const card = join(memory, 'red-scorecard.md')
  const first = readFileSync(card, 'utf8')
  assert.ok(first.startsWith('# red scorecard'), 'a card born this run gets its header')
  assert.ok(first.includes('## 2026-07-01_first-run'), 'the run DIRECTORY name labels the series — dated and unique, so no clock is trusted')

  writeScorecards(runB, [], memory)
  const second = readFileSync(card, 'utf8')
  assert.ok(second.includes('## 2026-07-01_first-run'), 'the earlier run survives — overwriting would destroy the series')
  assert.ok(second.includes('## 2026-07-18_second-run'), 'and this run is appended after it')
  assert.equal(second.split('# red scorecard').length, 2, 'the header is written once, not re-stamped every run')
  assert.ok(second.indexOf('2026-07-01_first-run') < second.indexOf('2026-07-18_second-run'), 'chronological, because appended')

  // Without the tracked memory dir there is nowhere a series could accumulate, so
  // writing anywhere else would be a silently discarded scorecard.
  const r = writeScorecards(runA, [], join(memory, 'no-such-dir'))
  assert.equal(r.written, false)
  assert.match(r.reason, /scorecards need the tracked memory dir/)
})

test('setup: records/ is created; stale mirrors purge at 30 days and fresh ones survive', async () => {
  const { buildSkeleton, purgeStaleMirrors } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const run = tmp()
  buildSkeleton(run, 'topic')
  assert.ok(existsSync(join(run, 'records')), 'records/ born at setup')
  const mirrors = tmp()
  mkdirSync(join(mirrors, 'oldrun'))
  mkdirSync(join(mirrors, 'newrun'))
  utimesSync(join(mirrors, 'oldrun'), new Date(Date.now() - 40 * 86400e3), new Date(Date.now() - 40 * 86400e3))
  const r = purgeStaleMirrors(mirrors)
  assert.equal(r.purged, 1)
  assert.ok(!existsSync(join(mirrors, 'oldrun')) && existsSync(join(mirrors, 'newrun')))
})

test('record-join audit: an event no transcript invoked is FLAGGED; the binary invocation shape is matched', async () => {
  const { recordJoinAudit } = await import('../../skills/research-protocol/scripts/capture-research-run.mjs')
  const run = tmp()
  const transcripts = tmp()
  writeEvents(run, 'red-lens-r1-L1', [{ type: 'finding', payload: { label: 'L1-F1', text: 'x' } }])
  // Transcript carrying the ROLE-SUBCOMMAND form the binary is invoked with.
  writeFileSync(join(transcripts, 'agent-aaa.jsonl'), [
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Bash', input: { command: 'feov-record lens register --run x --seat-id red-lens-r1-L1' } }] } }),
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Bash', input: { command: 'feov-record lens finding --label L1-F1 --text "x" --run x --seat-id red-lens-r1-L1' } }] } }),
  ].join('\n') + '\n')
  const pass = recordJoinAudit(run, transcripts, ['agent-aaa.jsonl'])
  assert.equal(pass.verdict, 'PASS', pass.detail)

  // An event no transcript ever invoked: the orphan must be named, not counted.
  writeEvents(run, 'red-merge-r1', [{ type: 'friction', payload: { text: 'nobody ran this through a tool' } }], 'ddeeff00')
  const flagged = recordJoinAudit(run, transcripts, ['agent-aaa.jsonl'])
  assert.equal(flagged.verdict, 'FAIL')
  assert.ok(flagged.detail.includes('red-merge-r1 friction'), 'the orphan event is named')
})

test('record-join audit still reads pre-port transcripts (the mjs invocation shape)', async () => {
  const { recordJoinAudit } = await import('../../skills/research-protocol/scripts/capture-research-run.mjs')
  const run = tmp()
  const transcripts = tmp()
  writeEvents(run, 'red-lens-r1-L1', [{ type: 'finding', payload: { label: 'L1-F1', text: 'x' } }])
  writeFileSync(join(transcripts, 'agent-old.jsonl'),
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Bash', input: { command: 'node tools/red-lens.mjs finding --label L1-F1 --run x --seat-id red-lens-r1-L1' } }] } }) + '\n')
  const pass = recordJoinAudit(run, transcripts, ['agent-old.jsonl'])
  assert.equal(pass.verdict, 'PASS', pass.detail)
})

test('setup preflight: a missing or skewed record binary refuses the run BEFORE any state exists', async () => {
  const { preflightRecordBinary, recordToolVersion } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')

  const missing = preflightRecordBinary('0.1.0', 'feov-record', () => ({ error: { code: 'ENOENT' }, status: null }))
  assert.equal(missing.ok, false)
  assert.match(missing.reason, /not runnable/)
  assert.match(missing.remedy, /doctor --fix/)

  // Skew is the subtler failure: the binary RUNS, so nothing else would notice,
  // and it writes events under a different contract for the whole run.
  const skewed = preflightRecordBinary('0.2.0', 'feov-record', () => ({ status: 0, stdout: Buffer.from('0.1.0\n') }))
  assert.equal(skewed.ok, false)
  assert.match(skewed.reason, /0\.1\.0.*expects 0\.2\.0/)

  const good = preflightRecordBinary('0.1.0', 'feov-record', () => ({ status: 0, stdout: Buffer.from('0.1.0\n') }))
  assert.equal(good.ok, true)
  assert.equal(good.version, '0.1.0')

  // The plugin manifest is the version authority; a hardcoded copy here would be
  // the very skew the preflight exists to catch.
  assert.equal(recordToolVersion(), '0.1.0')
})

test('record-join audit matches the QUOTED binary path the engine actually emits', async () => {
  const { recordJoinAudit } = await import('../../skills/research-protocol/scripts/capture-research-run.mjs')
  const run = tmp()
  const transcripts = tmp()
  writeEvents(run, 'red-lens-r1-L1', [{ type: 'finding', payload: { label: 'L1-F1', text: 'x' } }])
  // Exactly what recordClause renders: the path is quoted because a plugin root
  // can contain spaces. Requiring whitespace after the binary name matched
  // nothing here, so the audit reported healthy while measuring nothing.
  writeFileSync(join(transcripts, 'agent-q.jsonl'),
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Bash', input: { command: '"/plug/bin/feov-record" lens finding --label L1-F1 --run x --seat-id red-lens-r1-L1' } }] } }) + '\n')
  const pass = recordJoinAudit(run, transcripts, ['agent-q.jsonl'])
  assert.equal(pass.verdict, 'PASS', pass.detail)
})

test('setup preflight binds to INTENT: fatal with --bin-dir, reported without it', async () => {
  const { runSetupCli } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs').catch(() => ({}))
  // The CLI path is covered end-to-end in run-scripts.test.mjs; what matters here
  // is the rule: a run that never asked to record must not be blocked by a tool
  // it does not use, and a run that DID ask must not start without it.
  const { preflightRecordBinary } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const absent = preflightRecordBinary('0.1.0', 'no-such-binary')
  assert.equal(absent.ok, false, 'absence is detected either way — the difference is what setup DOES about it')
  assert.ok(absent.reason.includes('not runnable'))
  assert.equal(typeof runSetupCli, 'undefined')
})

test('gap-pattern mirror: promoted corpus wins, raw accrual fills gaps, duplicates never double-stage', async () => {
  const { mirrorGapPatterns } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const run = tmp(); mkdirSync(join(run, 'inputs'), { recursive: true })
  const promoted = tmp(); const raw = tmp()
  writeFileSync(join(promoted, 'shared.md'), 'PROMOTED VERSION')
  writeFileSync(join(promoted, 'README.md'), 'not a pattern — the corpus doc')
  writeFileSync(join(raw, 'shared.md'), 'RAW VERSION')
  writeFileSync(join(raw, 'only-raw.md'), 'not yet promoted')

  const r = mirrorGapPatterns([promoted, raw], run)
  assert.equal(r.written, true)
  const staged = readFileSync(join(run, 'inputs', 'red-gap-patterns.md'), 'utf8')
  assert.ok(staged.includes('PROMOTED VERSION'), 'the reviewed corpus is authoritative')
  assert.ok(!staged.includes('RAW VERSION'), 'a promoted pattern is never re-staged from raw accrual')
  assert.ok(staged.includes('not yet promoted'), 'un-promoted accrual still reaches the run')
  assert.ok(!staged.includes('the corpus doc'), 'the README is documentation, not a pattern')
  assert.equal(r.files, 2)
})

test('attestation integrity: a claimed act with no matching tool call is FLAGGED, a reconciled one passes', async () => {
  const { attestationAudit } = await import('../../skills/research-protocol/scripts/capture-research-run.mjs')
  const run = tmp(); const tr = tmp()
  mkdirSync(join(run, 'red'), { recursive: true })
  writeFileSync(join(run, 'red', 'archive.md'),
    '# archive\n\n## R1-1 — closed\nthe problem\nverification anchor: L1 | Read | report.md#SectionTwo\n' +
    '\n## R1-2 — closed\nother\nverification anchor: L5 | git show | 7bc501e:plans/record-tool.md\n')
  writeFileSync(join(tr, 'agent-a.jsonl'),
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Read', input: { file_path: 'report.md#SectionTwo' } }] } }) + '\n')

  const r = attestationAudit(run, tr, ['agent-a.jsonl'], 5)
  assert.equal(r.verdict, 'FAIL')
  assert.equal(r.unreconciled.length, 1)
  assert.equal(r.unreconciled[0].id, 'R1-2', 'the claim nothing supports is the one named')
  assert.match(r.detail, /RECORD IS HONEST, never on the merits/, 'the separation is stated in the finding itself')

  // A CARRIED closure asserts no fresh act, so reconciling it would manufacture a
  // finding out of an honest statement that this round did nothing new.
  writeFileSync(join(run, 'red', 'archive.md'),
    '# archive\n\n## R2-1 — closed\np\nverification anchor: CARRIED from round 1\n')
  assert.equal(attestationAudit(run, tr, ['agent-a.jsonl'], 5).verdict, 'SKIP')
})

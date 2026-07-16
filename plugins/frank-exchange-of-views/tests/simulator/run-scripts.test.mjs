// node --test — run-setup.mjs / run-capture.mjs against temp-dir fixtures. Zero tokens,
// zero network; the automation-doctrine counterpart of the debate simulator: mechanics
// that moved from prose to scripts get tests the prose never had.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, existsSync, readFileSync, writeFileSync, mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { buildSkeleton, buildPinned, mirrorGapPatterns, writeRunLiveMarker } from '../../skills/research-protocol/scripts/run-setup.mjs'
import { readJournal, telemetryAudit, shardAudit, frictionAudit } from '../../skills/research-protocol/scripts/run-capture.mjs'

const tmp = () => mkdtempSync(join(tmpdir(), 'feov-runscripts-'))

// ---- run-setup ----

test('skeleton: creates stubs with topic headers; ledger/archive/telemetry are NOT created (red-merge-born)', () => {
  const dir = tmp()
  const { created } = buildSkeleton(dir, 'test topic')
  assert.equal(created.length, 7)
  assert.ok(readFileSync(join(dir, 'blue', 'report.md'), 'utf8').includes('test topic'))
  assert.ok(existsSync(join(dir, 'red', 'candidates')))
  assert.ok(!existsSync(join(dir, 'red', 'ledger.md')), 'ledger must be red-merge-born')
  assert.ok(!existsSync(join(dir, 'red', 'archive.md')), 'archive must be red-merge-born')
  assert.ok(!existsSync(join(dir, 'trajectories', 'board-telemetry.jsonl')))
})

test('skeleton: idempotent — pre-staged files are never overwritten', () => {
  const dir = tmp()
  mkdirSync(join(dir, 'blue'), { recursive: true })
  writeFileSync(join(dir, 'blue', 'report.md'), 'PRE-STAGED CONTENT\n')
  const { created, skipped } = buildSkeleton(dir, 'topic')
  assert.ok(skipped.includes('blue/report.md'))
  assert.equal(created.length, 6)
  assert.equal(readFileSync(join(dir, 'blue', 'report.md'), 'utf8'), 'PRE-STAGED CONTENT\n')
})

test('pinned: HEAD row + per-cite pins, honoring explicit @pin; pre-staged PINNED kept', () => {
  const dir = tmp()
  buildSkeleton(dir, 'topic')
  const r = buildPinned(dir, 'abc1234', ['research/old-run@def5678', 'ideas/backlog.md'])
  assert.ok(r.written)
  const txt = readFileSync(r.path, 'utf8')
  assert.ok(txt.includes('`abc1234`') && txt.includes('`def5678`'), 'explicit pin honored, HEAD default applied')
  assert.ok(txt.includes('ideas/backlog.md'))
  const again = buildPinned(dir, 'zzz9999', [])
  assert.equal(again.written, false, 'pre-staged PINNED never overwritten')
})

test('gap-pattern mirror: concatenates memory files; absent/empty memory is a stated no-op', () => {
  const dir = tmp(); buildSkeleton(dir, 'topic')
  const mem = tmp()
  writeFileSync(join(mem, 'pattern_a.md'), '# pattern A\n')
  writeFileSync(join(mem, 'pattern_b.md'), '# pattern B\n')
  const r = mirrorGapPatterns(mem, dir)
  assert.equal(r.files, 2)
  const out = readFileSync(join(dir, 'inputs', 'red-gap-patterns.md'), 'utf8')
  assert.ok(out.includes('pattern A') && out.includes('pattern B') && out.includes('read-only copy'))
  const none = mirrorGapPatterns(join(mem, 'nope'), tmp())
  assert.equal(none.written, false)
})

test('run-live marker: commitment-as-state with the pinned paths for hook guards', () => {
  const project = tmp()
  const p = writeRunLiveMarker(project, 'research/x', ['research/old-run', 'ideas/backlog.md'])
  const j = JSON.parse(readFileSync(p, 'utf8'))
  assert.equal(j.runDir, 'research/x')
  assert.deepEqual(j.pinnedPaths, ['research/old-run', 'ideas/backlog.md'])
})

// ---- run-capture audits ----

function fixtureRun({ telemetryRounds = 2, redRounds = 2, ledgerLines = 2, archiveBlocks = 2, frictionInFile = true } = {}) {
  const dir = tmp()
  mkdirSync(join(dir, 'red'), { recursive: true })
  mkdirSync(join(dir, 'trajectories'), { recursive: true })
  writeFileSync(join(dir, 'debate.md'), Array.from({ length: redRounds }, (_, i) => `## Round ${i + 1}\n### RED\nverdict\n`).join('\n'))
  if (telemetryRounds >= 0) {
    writeFileSync(join(dir, 'trajectories', 'board-telemetry.jsonl'),
      Array.from({ length: telemetryRounds }, (_, i) => JSON.stringify({ round: i + 1, mass: 4, open_count: 2 })).join('\n') + '\n')
  }
  writeFileSync(join(dir, 'red', 'ledger.md'), '# ledger\n## open\n\n## closure index\n' +
    Array.from({ length: ledgerLines }, (_, i) => `R1-${i + 1} | closed | fixed | -`).join('\n') + '\n')
  writeFileSync(join(dir, 'red', 'archive.md'), '# archive\n' +
    Array.from({ length: archiveBlocks }, (_, i) => `## R1-${i + 1} — closed\nprose record\n`).join('\n'))
  writeFileSync(join(dir, 'friction.md'), frictionInFile ? '- red-merge-r1: needed a PDF extractor for X\n' : '# friction\n')
  writeFileSync(join(dir, 'trajectories', 'journal.jsonl'), [
    JSON.stringify({ type: 'result', result: { verdict: 'FAIL', ledger_closure_lines: ledgerLines, archive_blocks: archiveBlocks, friction: ['red-merge-r1: needed a PDF extractor for X'] } }),
  ].join('\n') + '\n')
  return dir
}

test('telemetry audit: PASS when lines cover red rounds; FAIL when a round is missing; FAIL when absent with rounds on record', () => {
  assert.equal(telemetryAudit(fixtureRun()).verdict, 'PASS')
  assert.equal(telemetryAudit(fixtureRun({ telemetryRounds: 1 })).verdict, 'FAIL')
  const noFile = fixtureRun({ telemetryRounds: -1 })
  assert.equal(telemetryAudit(noFile).verdict, 'FAIL')
})

test('shard audit: measured counts vs envelope self-report — vacuity-adjacent inconsistency FAILs', () => {
  const ok = fixtureRun()
  const { results } = readJournal(join(ok, 'trajectories'))
  assert.equal(shardAudit(ok, results).verdict, 'PASS')
  // Envelope claims 2/2 but the files hold 3 index lines — the self-report diverges from disk.
  const lying = fixtureRun({ ledgerLines: 3 })
  writeFileSync(join(lying, 'trajectories', 'journal.jsonl'),
    JSON.stringify({ type: 'result', result: { ledger_closure_lines: 2, archive_blocks: 2, friction: [] } }) + '\n')
  const r2 = shardAudit(lying, readJournal(join(lying, 'trajectories')).results)
  assert.equal(r2.verdict, 'FAIL')
  assert.ok(r2.detail.includes('measured (heuristic)'))
})

test('friction parity: an envelope entry missing from friction.md is named in the FAIL detail', () => {
  const ok = fixtureRun()
  assert.equal(frictionAudit(ok, readJournal(join(ok, 'trajectories')).friction).verdict, 'PASS')
  const missing = fixtureRun({ frictionInFile: false })
  const r = frictionAudit(missing, readJournal(join(missing, 'trajectories')).friction)
  assert.equal(r.verdict, 'FAIL')
  assert.ok(r.detail.includes('needed a PDF extractor'))
})

test('shard audit: pre-sharding run (no ledger/archive) is SKIP, not FAIL', () => {
  const dir = tmp()
  mkdirSync(join(dir, 'trajectories'), { recursive: true })
  assert.equal(shardAudit(dir, []).verdict, 'SKIP')
})

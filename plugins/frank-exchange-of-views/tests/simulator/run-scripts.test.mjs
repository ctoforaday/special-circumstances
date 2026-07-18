// node --test — setup-research-run.mjs / capture-research-run.mjs against temp-dir fixtures. Zero tokens,
// zero network; the automation-doctrine counterpart of the debate simulator: mechanics
// that moved from prose to scripts get tests the prose never had.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, existsSync, readFileSync, writeFileSync, mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { buildSkeleton, buildPinned, mirrorGapPatterns, writeRunLiveMarker } from '../../skills/research-protocol/scripts/setup-research-run.mjs'
import { readJournal, telemetryAudit, shardAudit, frictionAudit } from '../../skills/research-protocol/scripts/capture-research-run.mjs'

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

// ---- Integration tier (the glue the pure-function tests skipped) ----

import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { dirname } from 'node:path'
import { qmdRefresh } from '../../skills/research-protocol/scripts/setup-research-run.mjs'
import { capture } from '../../skills/research-protocol/scripts/capture-research-run.mjs'

const SCRIPTS = join(dirname(fileURLToPath(import.meta.url)), '..', '..', 'skills', 'research-protocol', 'scripts')

function fixtureTranscript() {
  const dir = tmp()
  writeFileSync(join(dir, 'journal.jsonl'), [
    JSON.stringify({ type: 'result', result: { verdict: 'FAIL', ledger_closure_lines: 1, archive_blocks: 1, friction: ['red-merge-r1: probe friction entry'] } }),
  ].join('\n') + '\n')
  // One minimal agent transcript with a usage record so cost-audit produces a row.
  writeFileSync(join(dir, 'agent-abc123.jsonl'), [
    JSON.stringify({ message: { role: 'assistant', model: 'claude-fable-5', usage: { input_tokens: 100, output_tokens: 50, cache_read_input_tokens: 1000, cache_creation_input_tokens: 200 }, content: [] } }),
  ].join('\n') + '\n')
  return dir
}

test('capture(): end-to-end mechanics — journal copy, tarball, cost.md with telemetry join, audit report, marker removal', () => {
  const runDir = fixtureRun({ ledgerLines: 1, archiveBlocks: 1 })
  const transcriptDir = fixtureTranscript()
  // A live marker in cwd/.claude — capture must remove it.
  const marker = join(process.cwd(), '.claude', 'run-live.json')
  const hadMarker = existsSync(marker)
  const priorMarker = hadMarker ? readFileSync(marker, 'utf8') : null
  writeFileSync(marker, JSON.stringify({ runDir, test: true }))
  const { audits } = capture(runDir, transcriptDir)
  assert.ok(existsSync(join(runDir, 'trajectories', 'journal.jsonl')), 'journal copied')
  assert.ok(existsSync(join(runDir, 'trajectories', 'agent-transcripts.tar.gz')), 'tarball built')
  const cost = readFileSync(join(runDir, 'cost.md'), 'utf8')
  assert.ok(cost.includes('# Cost audit'), 'cost.md written via cost-audit.mjs')
  assert.ok(cost.includes('## Board telemetry'), 'telemetry join section present')
  assert.ok(cost.includes('CUMULATIVE ARCHIVE'), 'corrected physics finding in notes')
  const audit = readFileSync(join(runDir, 'run-record-audit.md'), 'utf8')
  assert.ok(audit.includes('telemetry: PASS') && audit.includes('shards: PASS') && audit.includes('friction-parity: FAIL'),
    'audit verdicts rendered (fixture has envelope friction not present in friction.md)')
  assert.ok(!existsSync(marker), 'run-live marker removed')
  if (hadMarker) writeFileSync(marker, priorMarker) // restore whatever the session had
})

test('run-capture CLI: exit code 2 on any audit FAIL (integrity findings are never smoothed over)', () => {
  const runDir = fixtureRun({ ledgerLines: 1, archiveBlocks: 1, frictionInFile: false })
  const transcriptDir = fixtureTranscript()
  const r = spawnSync(process.execPath, [join(SCRIPTS, 'capture-research-run.mjs'), runDir, transcriptDir], { cwd: tmp() }) // cwd MUST be a temp dir: capture removes cwd/.claude/run-live.json, and an inherited repo cwd would delete a LIVE session marker (happened 2026-07-17 — killed the run-5 watcher)
  assert.equal(r.status, 2, `expected exit 2, got ${r.status}: ${r.stderr}`)
  assert.ok(r.stdout.toString().includes('friction-parity: FAIL'))
})

test('run-setup CLI: arg parsing end-to-end — topic header, multi-cite pins, --no-qmd, summary lines', () => {
  const cwd = tmp()
  const runDir = join(cwd, 'research', '2026-01-01_cli-test')
  const r = spawnSync(process.execPath, [join(SCRIPTS, 'setup-research-run.mjs'), runDir,
    '--topic', 'cli parse topic', '--cite', 'a/path@abc1234', '--cite', 'b/path', '--no-qmd'], { cwd })
  assert.equal(r.status, 0, r.stderr.toString())
  const out = r.stdout.toString()
  assert.ok(out.includes('skeleton: 7 created') && out.includes('skipped (--no-qmd)'), out)
  assert.ok(readFileSync(join(runDir, 'blue', 'report.md'), 'utf8').includes('cli parse topic'))
  const pinned = readFileSync(join(runDir, 'inputs', 'PINNED.md'), 'utf8')
  assert.ok(pinned.includes('`abc1234`') && pinned.includes('b/path'), 'both cites pinned, explicit pin honored')
  assert.ok(existsSync(join(cwd, '.claude', 'run-live.json')), 'marker written under the invoking cwd')
})

test('run-setup CLI: refuses to run without a runDir', () => {
  const r = spawnSync(process.execPath, [join(SCRIPTS, 'setup-research-run.mjs'), '--topic', 'x'])
  assert.equal(r.status, 1)
  assert.ok(r.stderr.toString().includes('usage:'))
})

test('qmdRefresh: not-installed branch is a stated no-op (injected missing binary)', () => {
  const r = qmdRefresh('definitely-not-a-real-binary-xyz')
  assert.equal(r.ran, false)
  assert.ok(r.reason.includes('not installed'))
})

test('cost-audit CLI: without runDir no telemetry section; with runDir but no telemetry file, the absent-file note names capture-audit', () => {
  const transcriptDir = fixtureTranscript()
  const bare = spawnSync(process.execPath, [join(SCRIPTS, 'cost-audit.mjs'), transcriptDir])
  assert.equal(bare.status, 0)
  assert.ok(!bare.stdout.toString().includes('## Board telemetry'), 'no join without runDir arg')
  const runDir = tmp()
  const withRun = spawnSync(process.execPath, [join(SCRIPTS, 'cost-audit.mjs'), transcriptDir, runDir])
  assert.ok(withRun.stdout.toString().includes('no board-telemetry.jsonl'), 'absent-file branch stated, not silent')
})

test('batch-collapse CLI: pairs candidate-read ingestions and reports per-round collapse dollars', () => {
  const dir = tmp()
  const toolUse = { message: { role: 'assistant', model: 'claude-fable-5', usage: { input_tokens: 10, output_tokens: 5, cache_read_input_tokens: 100, cache_creation_input_tokens: 10 }, content: [{ type: 'tool_use', id: 'tu1', name: 'Read', input: { file_path: 'red/candidates/round-1-lens-1.md' } }] } }
  const toolResult = { message: { role: 'user', content: [{ type: 'tool_result', tool_use_id: 'tu1' }] } }
  const ingest = { message: { role: 'assistant', model: 'claude-fable-5', usage: { input_tokens: 20, output_tokens: 5, cache_read_input_tokens: 50000, cache_creation_input_tokens: 1000 }, content: [] } }
  writeFileSync(join(dir, 'agent-merge1.jsonl'),
    [JSON.stringify({ message: { role: 'user', content: 'Red merge, round 1. ...' } }), JSON.stringify(toolUse), JSON.stringify(toolResult), JSON.stringify(ingest)].map(String).join('\n') + '\n')
  const r = spawnSync(process.execPath, [join(SCRIPTS, 'measure-read-batching.mjs'), dir])
  assert.equal(r.status, 0, r.stderr.toString())
  assert.ok(/round 1/.test(r.stdout.toString()), `expected a round-1 row, got: ${r.stdout}`)
})

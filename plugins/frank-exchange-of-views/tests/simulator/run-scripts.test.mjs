// node --test — setup-research-run.mjs / capture-research-run.mjs against temp-dir fixtures. Zero tokens,
// zero network; the automation-doctrine counterpart of the debate simulator: mechanics
// that moved from prose to scripts get tests the prose never had.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, existsSync, readFileSync, writeFileSync, mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { buildSkeleton, buildPinned, mirrorGapPatterns, writeRunLiveMarker, validatePins } from '../../skills/research-protocol/scripts/setup-research-run.mjs'
import { readJournal, telemetryAudit, shardAudit, frictionAudit, contextUse, assemblyScreen, recordParityAudit } from '../../skills/research-protocol/scripts/capture-research-run.mjs'

const tmp = () => mkdtempSync(join(tmpdir(), 'feov-runscripts-'))

// ---- run-setup ----

test('skeleton: creates stubs with topic headers; ledger/archive/telemetry are NOT created (red-merge-born)', () => {
  const dir = tmp()
  const { created } = buildSkeleton(dir, 'test topic')
  assert.equal(created.length, 6) // friction.md retired — friction is a record event, projected to render-shadow/friction.md
  assert.ok(!existsSync(join(dir, 'friction.md')), 'friction.md is not pre-created — it is a render projection of the friction verb')
  assert.ok(readFileSync(join(dir, 'blue', 'report.md'), 'utf8').includes('test topic'))
  assert.ok(!existsSync(join(dir, 'red', 'candidates')), 'red/candidates is retired — lens findings are record events, read via `show --view findings`')
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
  assert.equal(created.length, 5) // 6 stubs (friction.md retired) minus the 1 pre-staged
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

test('pin validation (W1.1): missing path at pin is named; explicit pin honored; non-git context is a stated skip', () => {
  const calls = []
  const gitOk = (args) => { calls.push(args.join(' ')); return { status: 0 } }
  const ok = validatePins(['plans/x.md@abc1234', 'ideas/y.md'], 'headddd', gitOk)
  assert.equal(ok.missing.length, 0)
  assert.equal(ok.checked, 2)
  assert.ok(calls[0].includes('abc1234:plans/x.md'), 'explicit pin used')
  assert.ok(calls[1].includes('headddd:ideas/y.md'), 'HEAD default used')
  const gitMiss = (args) => ({ status: args.join(' ').includes('plans/gone.md') ? 128 : 0 })
  const bad = validatePins(['plans/gone.md@abc1234', 'ideas/y.md'], 'headddd', gitMiss)
  assert.equal(bad.missing.length, 1)
  assert.equal(bad.missing[0].path, 'plans/gone.md')
  const skip = validatePins(['a@b'], 'unknown')
  assert.ok(skip.skipped && skip.skipped.includes('UNVALIDATED'), 'no git repo -> stated skip, never a silent pass')
})

// ---- run-capture audits ----

function fixtureRun({ telemetryRounds = 2, redRounds = 2, ledgerLines = 2, archiveBlocks = 2, frictionOnRecord = true, debateOnShadow = false } = {}) {
  const dir = tmp()
  mkdirSync(join(dir, 'red'), { recursive: true })
  mkdirSync(join(dir, 'trajectories'), { recursive: true })
  mkdirSync(join(dir, 'records'), { recursive: true })
  const debateBody = Array.from({ length: redRounds }, (_, i) => `## Round ${i + 1}\n### RED\nverdict\n### BLUE\nresponse\n`).join('\n')
  if (debateOnShadow) {
    // The tool projects the transcript here; the root debate.md is a stub post-#109. Writing the
    // populated projection to render-shadow and a STUB to root proves readers prefer the projection.
    mkdirSync(join(dir, 'records', 'render-shadow'), { recursive: true })
    writeFileSync(join(dir, 'records', 'render-shadow', 'debate.md'), debateBody)
    writeFileSync(join(dir, 'debate.md'), '# debate.md — stub\n')
  } else {
    writeFileSync(join(dir, 'debate.md'), debateBody)
  }
  // Friction is a record event now — write a friction event shard iff the seat recorded it.
  if (frictionOnRecord) {
    writeFileSync(join(dir, 'records', 'events-red-merge-r1-fix.jsonl'),
      JSON.stringify({ seq: 1, ts: '2026-07-26T00:00:00Z', seatId: 'red-merge-r1', nonce: 'fix', round: 1, type: 'friction', key: 'red-merge-r1:friction:#1', payload: { text: 'red-merge-r1: needed a PDF extractor for X' } }) + '\n')
  } else {
    writeFileSync(join(dir, 'records', 'events-red-merge-r1-fix.jsonl'), '') // records/ exists but carries no friction event
  }
  mkdirSync(join(dir, 'blue'), { recursive: true })
  writeFileSync(join(dir, 'blue', 'CHANGELOG.md'), Array.from({ length: redRounds }, (_, i) => `## Round ${i + 1}\nedits\n`).join('\n'))
  if (telemetryRounds >= 0) {
    writeFileSync(join(dir, 'trajectories', 'board-telemetry.jsonl'),
      Array.from({ length: telemetryRounds }, (_, i) => JSON.stringify({ round: i + 1, mass: 4, open_count: 2 })).join('\n') + '\n')
  }
  writeFileSync(join(dir, 'red', 'ledger.md'), '# ledger\n## open\n\n## closure index\n' +
    Array.from({ length: ledgerLines }, (_, i) => `R1-${i + 1} | closed | fixed | -`).join('\n') + '\n')
  writeFileSync(join(dir, 'red', 'archive.md'), '# archive\n' +
    Array.from({ length: archiveBlocks }, (_, i) => `## R1-${i + 1} — closed\nprose record\n`).join('\n'))
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

test('friction parity (#87): envelope friction must match a friction EVENT on the record; a gap FAILs; no records/ SKIPs', () => {
  // PASS: the seat's envelope friction has a matching friction event on the record.
  const ok = fixtureRun()
  assert.equal(frictionAudit(ok, readJournal(join(ok, 'trajectories')).friction).verdict, 'PASS')
  // FAIL: envelope reports friction but no friction event on the record — a seat that skipped the verb.
  const missing = fixtureRun({ frictionOnRecord: false })
  const r = frictionAudit(missing, readJournal(join(missing, 'trajectories')).friction)
  assert.equal(r.verdict, 'FAIL')
  assert.ok(r.detail.includes('needed a PDF extractor') && r.detail.includes('skipped the friction verb'))
  // SKIP: a legacy no-record run has only the envelope channel.
  const legacy = tmp()
  assert.equal(frictionAudit(legacy, ['some friction']).verdict, 'SKIP')
})

test('class sweep (#112): telemetry + record-parity read the render-shadow debate projection, not the root stub', () => {
  // Post-#109 the root debate.md is a stub; the populated transcript lives in render-shadow. A
  // stub-read would see 0 red rounds and go vacuous (telemetry always-PASS off 0, parity spurious SKIP).
  const dir = fixtureRun({ debateOnShadow: true })
  const t = telemetryAudit(dir)
  assert.equal(t.verdict, 'PASS')
  assert.ok(t.detail.includes('vs 2 red round'), 'sees the projection\'s 2 red rounds, not the stub\'s 0')
  const rp = recordParityAudit(dir)
  assert.notEqual(rp.verdict, 'SKIP', 'runs against the record, not the spurious "no red rounds" SKIP off the stub')
  assert.equal(rp.verdict, 'PASS')
})

test('context-use (W1.3): haiku seat over 50% of its 200k window trips the WARN tripwire', () => {
  const dir = tmp()
  writeFileSync(join(dir, 'agent-big.jsonl'),
    JSON.stringify({ message: { role: 'assistant', model: 'claude-haiku-4-5', usage: { input_tokens: 1000, cache_read_input_tokens: 150000, cache_creation_input_tokens: 0, output_tokens: 10 }, content: [] } }) + '\n')
  writeFileSync(join(dir, 'agent-small.jsonl'),
    JSON.stringify({ message: { role: 'assistant', model: 'claude-fable-5', usage: { input_tokens: 1000, cache_read_input_tokens: 150000, cache_creation_input_tokens: 0, output_tokens: 10 }, content: [] } }) + '\n')
  const r = contextUse(dir, ['agent-big.jsonl', 'agent-small.jsonl'])
  assert.equal(r.verdict, 'WARN')
  assert.ok(r.detail.includes('big') && r.detail.includes('50% tripwire'), 'the breaching seat is named')
  const calm = contextUse(dir, ['agent-small.jsonl'])
  assert.equal(calm.verdict, 'PASS', 'same tokens on a 1M window is 15% — no flag')
})

test('assembly screen (W1.4): a REFUTED-row token in assembly-owned text WARNs with the token named', () => {
  const dir = tmp()
  mkdirSync(join(dir, 'red'), { recursive: true })
  writeFileSync(join(dir, 'red', 'citation-ledger.md'),
    '"#32191 open / leaf-checked OPEN" | live fetch | LOW — REFUTED: Closed as duplicate | r1 | 2026-07-17\n' +
    '"solid claim" | source | high | r1 | 2026-07-17\n')
  writeFileSync(join(dir, 'report.md'),
    '# assembled report\nThe open MCP-headless bug trio (#76239, #68375, #32191) blocks headless runs.\n\n## Blue team report (in full)\naudited body text\n')
  const r = assemblyScreen(dir)
  assert.equal(r.verdict, 'WARN')
  assert.ok(r.detail.includes('#32191'), 'the candidate regression token is named for human eyes')
  writeFileSync(join(dir, 'report.md'), '# assembled report\nTwo open bugs (#76239).\n\n## Blue team report (in full)\n#32191 appears only in the audited body, where red already re-reads it.\n')
  assert.equal(assemblyScreen(dir).verdict, 'PASS', 'tokens below the cut line are the audited body — not screened')
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
import * as capMod from '../../skills/research-protocol/scripts/capture-research-run.mjs'

const SCRIPTS = join(dirname(fileURLToPath(import.meta.url)), '..', '..', 'skills', 'research-protocol', 'scripts')

function fixtureTranscript() {
  const dir = tmp()
  writeFileSync(join(dir, 'journal.jsonl'), [
    // Friction text matches the friction EVENT fixtureRun writes to records/, so capture's
    // friction-parity (envelope-from-this-journal vs record events) PASSes (#87).
    JSON.stringify({ type: 'result', result: { verdict: 'FAIL', ledger_closure_lines: 1, archive_blocks: 1, friction: ['red-merge-r1: needed a PDF extractor for X'] } }),
  ].join('\n') + '\n')
  // One minimal agent transcript with a usage record so cost-audit produces a row.
  writeFileSync(join(dir, 'agent-abc123.jsonl'), [
    JSON.stringify({ message: { role: 'assistant', model: 'claude-fable-5', usage: { input_tokens: 100, output_tokens: 50, cache_read_input_tokens: 1000, cache_creation_input_tokens: 200 }, content: [] } }),
  ].join('\n') + '\n')
  return dir
}

test('capture(): end-to-end mechanics — journal copy, tarball, cost.md with telemetry join, audit report, marker removal', (t) => {
  const runDir = fixtureRun({ ledgerLines: 1, archiveBlocks: 1 })
  const transcriptDir = fixtureTranscript()
  // capture() removes cwd/.claude/run-live.json — run under a TEMP cwd so the test never
  // touches a live session's marker (2026-07-17 incident class) and passes in worktrees
  // where the repo .claude dir doesn't exist.
  const prevCwd = process.cwd()
  const testCwd = tmp()
  mkdirSync(join(testCwd, '.claude'), { recursive: true })
  process.chdir(testCwd)
  t.after(() => process.chdir(prevCwd))
  const marker = join(testCwd, '.claude', 'run-live.json')
  writeFileSync(marker, JSON.stringify({ runDir, test: true }))
  const { audits } = capture(runDir, transcriptDir)
  assert.ok(existsSync(join(runDir, 'trajectories', 'journal.jsonl')), 'journal copied')
  assert.ok(existsSync(join(runDir, 'trajectories', 'agent-transcripts.tar.gz')), 'tarball built')
  const cost = readFileSync(join(runDir, 'cost.md'), 'utf8')
  assert.ok(cost.includes('# Cost audit'), 'cost.md written via cost-audit.mjs')
  assert.ok(cost.includes('## Board telemetry'), 'telemetry join section present')
  assert.ok(cost.includes('CUMULATIVE ARCHIVE'), 'corrected physics finding in notes')
  const audit = readFileSync(join(runDir, 'run-record-audit.md'), 'utf8')
  assert.ok(audit.includes('telemetry: PASS') && audit.includes('shards: PASS') && audit.includes('friction-parity: PASS'),
    'audit verdicts rendered (fixture envelope friction matches the friction event on the record — #87)')
  assert.ok(audit.includes('context-use: PASS'), 'context-use telemetry in the audit report (W1.3)')
  assert.ok(!existsSync(join(runDir, 'friction.md')), 'capture no longer writes friction.md — friction is a record event (#87)')
  assert.ok(!existsSync(marker), 'run-live marker removed')
})

test('run-capture CLI: exit code 2 on any audit FAIL (integrity findings are never smoothed over)', () => {
  // The FAIL source here is a shard self-report diverging from disk (the vacuity-adjacent class
  // that stays fatal); friction is on the record (default) so friction-parity PASSes and does not
  // muddy which check failed.
  const runDir = fixtureRun({ ledgerLines: 3, archiveBlocks: 1 })
  const transcriptDir = fixtureTranscript()
  const r = spawnSync(process.execPath, [join(SCRIPTS, 'capture-research-run.mjs'), runDir, transcriptDir], { cwd: tmp() }) // cwd MUST be a temp dir: capture removes cwd/.claude/run-live.json, and an inherited repo cwd would delete a LIVE session marker (happened 2026-07-17 — killed the run-5 watcher)
  assert.equal(r.status, 2, `expected exit 2, got ${r.status}: ${r.stderr}`)
  assert.ok(r.stdout.toString().includes('shards: FAIL'))
  assert.ok(r.stdout.toString().includes('friction-parity: PASS'), 'friction is on the record — parity PASSes independently of the shard FAIL (#87)')
})

test('run-setup CLI (W1.1): a cite whose path does not exist at its pin fails setup loudly, creating nothing', () => {
  const cwd = tmp()
  const g = (args) => spawnSync('git', args, { cwd })
  g(['init', '-q'])
  g(['config', 'user.email', 't@t']); g(['config', 'user.name', 't'])
  writeFileSync(join(cwd, 'real.md'), 'exists\n')
  g(['add', 'real.md']); g(['commit', '-q', '-m', 'fixture'])
  const runDir = join(cwd, 'research', 'pin-test')
  const bad = spawnSync(process.execPath, [join(SCRIPTS, 'setup-research-run.mjs'), runDir,
    '--topic', 't', '--model', 'haiku', '--judgment-model', 'haiku', '--cite', 'plans/does-not-exist.md', '--no-qmd'], { cwd })
  assert.equal(bad.status, 2, `expected exit 2, got ${bad.status}: ${bad.stderr}`)
  assert.ok(bad.stderr.toString().includes('PIN VALIDATION FAILED') && bad.stderr.toString().includes('does-not-exist.md'))
  assert.ok(bad.stderr.toString().includes('inputs/'), 'the staging remedy is stated')
  assert.ok(!existsSync(join(runDir, 'blue')), 'nothing was created — validation runs before the skeleton')
  const good = spawnSync(process.execPath, [join(SCRIPTS, 'setup-research-run.mjs'), runDir,
    '--topic', 't', '--model', 'haiku', '--judgment-model', 'haiku', '--cite', 'real.md', '--no-qmd'], { cwd })
  assert.equal(good.status, 0, good.stderr.toString())
  assert.ok(good.stdout.toString().includes('1 cite(s) verified at their pins'))
})

test('run-setup CLI: arg parsing end-to-end — topic header, multi-cite pins, --no-qmd, summary lines', () => {
  const cwd = tmp()
  const runDir = join(cwd, 'research', '2026-01-01_cli-test')
  const r = spawnSync(process.execPath, [join(SCRIPTS, 'setup-research-run.mjs'), runDir,
    '--topic', 'cli parse topic', '--model', 'haiku', '--judgment-model', 'haiku', '--cite', 'a/path@abc1234', '--cite', 'b/path', '--no-qmd'], { cwd })
  assert.equal(r.status, 0, r.stderr.toString())
  const out = r.stdout.toString()
  assert.ok(out.includes('skeleton: 6 created') && out.includes('skipped (--no-qmd)'), out)
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

test('record parity (W1.7 post-hoc): missing BLUE blocks or CHANGELOG rounds FAIL; the PASS-exit final round is floored out', () => {
  const { recordParityAudit } = capMod
  const healthy = fixtureRun()
  assert.equal(recordParityAudit(healthy).verdict, 'PASS')
  const desynced = fixtureRun()
  writeFileSync(join(desynced, 'debate.md'), '## Round 1\n### RED\nv\n### BLUE\nr\n## Round 2\n### RED\nv\n## Round 3\n### RED\nv\n')
  const r = recordParityAudit(desynced)
  assert.equal(r.verdict, 'FAIL', '3 red rounds with 1 blue block is below the redRounds-1 floor')
  assert.ok(r.detail.includes('3 red round(s)'))
  const passExit = fixtureRun()
  writeFileSync(join(passExit, 'debate.md'), '## Round 1\n### RED\nv\n### BLUE\nr\n## Round 2\n### RED\nPASS\n')
  writeFileSync(join(passExit, 'blue', 'CHANGELOG.md'), '## Round 1\nedits\n')
  assert.equal(recordParityAudit(passExit).verdict, 'PASS', 'a PASS exit has no final blue response — floored')
})

test('W2e law mirror: repo law/ stages read-only into inputs/law; absent law dir is a stated no-op', async () => {
  const { mirrorLaw } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const repo = tmp()
  mkdirSync(join(repo, 'law'), { recursive: true })
  writeFileSync(join(repo, 'law', 'README.md'), '# law\nstatute > precedent > argument\n')
  writeFileSync(join(repo, 'law', 'precedents.md'), '# precedents\n## some-holding [AFFIRMED 2026-07-18]\n')
  const run = tmp()
  const r = mirrorLaw(join(repo, 'law'), run)
  assert.equal(r.files, 2)
  const staged = readFileSync(join(run, 'inputs', 'law', 'precedents.md'), 'utf8')
  assert.ok(staged.includes('read-only copy') && staged.includes('AFFIRMED'), 'mirrored with provenance banner')
  const none = mirrorLaw(join(repo, 'nope'), tmp())
  assert.equal(none.written, false)
})

test('W2e precedent harvest: rulings become PERSUASIVE proposals with defeasible form; harvest never invents facts', async () => {
  const { harvestPrecedents } = await import('../../skills/research-protocol/scripts/capture-research-run.mjs')
  const repo = tmp()
  mkdirSync(join(repo, 'law'), { recursive: true })
  const runDir = join(repo, 'research', '2026-07-18_law-test')
  mkdirSync(runDir, { recursive: true })
  // A rationale longer than the old 600-char cap, whose ACTIONABLE tail sat past the cut.
  const longRationale = 'context clause '.repeat(45) + 'Direction owed: TRAILING_ACTIONABLE_TAIL'
  const results = [
    { resolutions: [{ gap_id: 'R2-3', resolution: 'risk_accepted', rationale: 'complexity exceeds bounded likelihood x impact' }] },
    { rulings: [{ petitioner: 'blue-respond-r2', ruling: 'granted', opinion: 'scope narrowed to shipped artifacts' }] },
    { resolutions: [{ gap_id: 'R1-9', resolution: 'carried', rationale: longRationale }] },
  ]
  const r = harvestPrecedents(runDir, results, join(repo, 'law'))
  assert.equal(r.count, 3)
  const out = readFileSync(r.path, 'utf8')
  assert.ok(out.includes('[PERSUASIVE]') && !out.includes('[AFFIRMED'), 'everything starts persuasive')
  assert.ok(out.includes('holding: risk_accepted') && out.includes('holding: granted'))
  assert.ok(out.includes('facts: <reviewer: fill from the cited record'), 'the harvest never invents facts')
  assert.ok(out.includes('source: 2026-07-18_law-test, R2-3'), 'holdings carry their source anchors')
  // The rationale is stored WHOLE — the actionable tail past the old 600-char cut survives.
  assert.ok(longRationale.length > 600, 'fixture exceeds the old cap')
  assert.ok(out.includes('TRAILING_ACTIONABLE_TAIL'), 'full rationale preserved — no mid-sentence truncation')
  const noLaw = harvestPrecedents(runDir, results, join(repo, 'absent'))
  assert.equal(noLaw.written, false)
  assert.ok(noLaw.reason.includes('law'))
})

// Two readers of one corpus must agree on identity. mirrorGapPatterns dedups
// promoted-first; buildPatternIndex did not, so the raw accrual path's
// PRE-PROMOTION copies were re-counted as an unclassified backlog. Found by
// running setup for real: it reported 55 patterns needing classification when
// the true number was zero.
test('buildPatternIndex dedups promoted-first, like its sibling mirror', async () => {
  const { buildPatternIndex } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const promoted = tmp(), raw = tmp()
  // Same filename in both tiers: classified in the promoted corpus, classless in raw.
  writeFileSync(join(promoted, 'p.md'), '---\nmetadata:\n  classes: [false-universal]\ndescription: hook\n---\n# P\n')
  writeFileSync(join(raw, 'p.md'), '---\ndescription: pre-promotion copy\n---\n# P\n')
  const r = buildPatternIndex([promoted, raw])
  assert.deepEqual(r.unclassified, [], 'the raw copy must not resurrect as a backlog item')
  assert.equal(r.byClass['false-universal'].length, 1, 'and must not double-deliver either')
  // Order encodes authority: promoted first wins, so reversing it is a different answer.
  const reversed = buildPatternIndex([raw, promoted])
  assert.deepEqual(reversed.unclassified, ['p.md'], 'raw-first surfaces the unclassified copy — order is the policy')
})

// A harness-limit pattern is deliberately classless and is NOT unfinished work.
test('buildPatternIndex keeps harness-limit distinct from unclassified', async () => {
  const { buildPatternIndex } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const d = tmp()
  writeFileSync(join(d, 'h.md'), '---\nmetadata:\n  classes: []\n  class_note: harness-limit — a tooling constraint\n---\n# H\n')
  writeFileSync(join(d, 'u.md'), '---\nmetadata:\n  classes: []\n---\n# U\n')
  const r = buildPatternIndex(d)
  assert.deepEqual(r.harnessLimit, ['h.md'])
  assert.deepEqual(r.unclassified, ['u.md'], 'only the genuinely unclassified counts as backlog')
})

// Run 1 has no scorecards by construction — they are written at capture and read
// by the next run. The absent-reason printed a bare "undefined", which reads as a
// defect in a feature that was working exactly as designed.
test('mirrorScorecards explains an empty corpus instead of printing undefined', async () => {
  const { mirrorScorecards } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const mem = tmp(), runDir = tmp()
  mkdirSync(join(runDir, 'inputs'), { recursive: true })
  const r = mirrorScorecards(mem, runDir)
  assert.equal(r.written, false)
  assert.equal(typeof r.reason, 'string')
  assert.ok(r.reason.length > 0, 'a false result must carry a printable reason')
  assert.ok(/capture/.test(r.reason), 'and the reason should say where scorecards come from')
})

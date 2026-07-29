// node --test — capture-research-run.mjs against temp-dir fixtures. Zero tokens,
// zero network; the automation-doctrine counterpart of the debate simulator: mechanics
// that moved from prose to scripts get tests the prose never had.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, existsSync, readFileSync, writeFileSync, mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { readJournal, telemetryAudit, shardAudit, frictionAudit, contextUse, assemblyScreen } from '../../skills/research-protocol/scripts/capture-research-run.mjs'

const tmp = () => mkdtempSync(join(tmpdir(), 'feov-runscripts-'))

// ---- run-capture audits ----

function fixtureRun({ telemetryRounds = 2, redRounds = 2, ledgerLines = 2, archiveBlocks = 2, frictionInFile = true } = {}) {
  const dir = tmp()
  mkdirSync(join(dir, 'red'), { recursive: true })
  mkdirSync(join(dir, 'trajectories'), { recursive: true })
  writeFileSync(join(dir, 'debate.md'), Array.from({ length: redRounds }, (_, i) => `## Round ${i + 1}\n### RED\nverdict\n### BLUE\nresponse\n`).join('\n'))
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
  writeFileSync(join(dir, 'friction.md'), frictionInFile ? '- red-merge-r1: needed a PDF extractor for X\n' : '# friction\n')
  writeFileSync(join(dir, 'trajectories', 'journal.jsonl'), [
    JSON.stringify({ type: 'result', result: { verdict: 'FAIL', ledger_closure_lines: ledgerLines, archive_blocks: archiveBlocks, friction: ['red-merge-r1: needed a PDF extractor for X'] } }),
  ].join('\n') + '\n')
  return dir
}

// The record-backed audits (telemetry, record-parity, friction-parity) read the run through
// the tool's views (`show --view debate --json`, `--view friction`). In production that
// reader wraps the feov-record binary; here we inject a fake so the audit LOGIC is exercised
// without a real binary — the same seam the dashboard tests use. `friction` is the set of
// friction texts ON THE RECORD (what the friction verb captured).
function views({ redRounds = 2, blueRounds = 2, friction = [] } = {}) {
  return (view) => {
    if (view === 'debate') {
      const n = Math.max(redRounds, blueRounds)
      return {
        rounds: Array.from({ length: n }, (_, i) => ({
          round: i + 1,
          red: i < redRounds ? [`red r${i + 1}`] : [],
          blue: i < blueRounds ? [`blue r${i + 1}`] : [],
          lead: [],
        })),
      }
    }
    if (view === 'friction') {
      return { friction: friction.map((t) => ({ seat_id: 'red-merge-r1', round: 1, text: String(t) })), counts: { total: friction.length } }
    }
    throw new Error(`views() fake got an unexpected view: ${view}`)
  }
}

test('telemetry audit: PASS when lines cover red rounds; FAIL when a round is missing; FAIL when absent with rounds on record; SKIP without --bin', () => {
  const viewJSON = views({ redRounds: 2 })
  assert.equal(telemetryAudit(fixtureRun(), { viewJSON }).verdict, 'PASS')
  assert.equal(telemetryAudit(fixtureRun({ telemetryRounds: 1 }), { viewJSON }).verdict, 'FAIL')
  const noFile = fixtureRun({ telemetryRounds: -1 })
  assert.equal(telemetryAudit(noFile, { viewJSON }).verdict, 'FAIL')
  // Without a view reader (no --bin) the audit reads nothing and SKIPs — never a false PASS.
  assert.equal(telemetryAudit(fixtureRun()).verdict, 'SKIP')
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

test('friction parity: envelope friction on the record PASSes; an entry the record never got FAILs and is named; no --bin SKIPs', () => {
  const dir = fixtureRun()
  const envFriction = readJournal(join(dir, 'trajectories')).friction // ['red-merge-r1: needed a PDF extractor for X']
  // The friction reached the record (the friction verb captured it): PASS.
  const onRecord = views({ friction: envFriction })
  assert.equal(frictionAudit(dir, envFriction, { viewJSON: onRecord }).verdict, 'PASS')
  // The envelope reported friction the record never received — a first-class pipeline error.
  const r = frictionAudit(dir, envFriction, { viewJSON: views({ friction: [] }) })
  assert.equal(r.verdict, 'FAIL')
  assert.ok(r.detail.includes('needed a PDF extractor'), 'the entry that never reached the record is named')
  assert.ok(r.detail.includes('never reached the record'), 'the finding states the pipeline gap, not a file miss')
  // Without a view reader (no --bin) the audit cannot read the record and SKIPs.
  assert.equal(frictionAudit(dir, envFriction).verdict, 'SKIP')
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
import { capture } from '../../skills/research-protocol/scripts/capture-research-run.mjs'
import * as capMod from '../../skills/research-protocol/scripts/capture-research-run.mjs'

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
  // No --bin passed here (this test exercises capture() mechanics, not the audit internals —
  // those have unit tests, and the real --bin path is driven end-to-end in the real-data
  // check). The three record-backed audits therefore SKIP rather than reading a stale file.
  assert.ok(audit.includes('shards: PASS'), 'shards audit runs (it reads files, not the record views)')
  assert.ok(audit.includes('telemetry: SKIP') && audit.includes('friction-parity: SKIP') && audit.includes('record-parity: SKIP'),
    'the record-backed audits SKIP without --bin — never a false PASS on an unread record')
  assert.ok(audit.includes('context-use: PASS'), 'context-use telemetry in the audit report (W1.3)')
  assert.ok(!existsSync(marker), 'run-live marker removed')
})

test('run-capture CLI: exit code 2 on any audit FAIL (integrity findings are never smoothed over)', () => {
  // The FAIL source is a shard self-report diverging from disk (the fixture's journal claims
  // 1/1 while the runDir files hold 3/1) — the vacuity-adjacent class that stays fatal. The
  // record-backed audits SKIP without --bin and so cannot be the FAIL here.
  const runDir = fixtureRun({ ledgerLines: 3, archiveBlocks: 1, frictionInFile: false })
  const transcriptDir = fixtureTranscript()
  const r = spawnSync(process.execPath, [join(SCRIPTS, 'capture-research-run.mjs'), runDir, transcriptDir], { cwd: tmp() }) // cwd MUST be a temp dir: capture removes cwd/.claude/run-live.json, and an inherited repo cwd would delete a LIVE session marker (happened 2026-07-17 — killed the run-5 watcher)
  assert.equal(r.status, 2, `expected exit 2, got ${r.status}: ${r.stderr}`)
  assert.ok(r.stdout.toString().includes('shards: FAIL'))
  assert.ok(r.stdout.toString().includes('friction-parity: SKIP'), 'friction-parity SKIPs without --bin — the exit-2 is shards, not friction')
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

test('record parity (W1.7 post-hoc): missing blue sittings or CHANGELOG rounds FAIL; the PASS-exit final round is floored out; no --bin SKIPs', () => {
  const { recordParityAudit } = capMod
  // healthy: 2 red, 2 blue on the record; fixtureRun writes a 2-round CHANGELOG → PASS.
  const healthy = fixtureRun()
  assert.equal(recordParityAudit(healthy, { viewJSON: views({ redRounds: 2, blueRounds: 2 }) }).verdict, 'PASS')
  // desynced: 3 red rounds, 1 blue sitting on the record — below the redRounds-1 floor.
  const desynced = fixtureRun()
  const r = recordParityAudit(desynced, { viewJSON: views({ redRounds: 3, blueRounds: 1 }) })
  assert.equal(r.verdict, 'FAIL', '3 red rounds with 1 blue sitting is below the redRounds-1 floor')
  assert.ok(r.detail.includes('3 red round(s)'))
  // PASS exit: 2 red, 1 blue (blue never took the final turn), CHANGELOG 1 round → floored to PASS.
  const passExit = fixtureRun()
  writeFileSync(join(passExit, 'blue', 'CHANGELOG.md'), '## Round 1\nedits\n')
  assert.equal(recordParityAudit(passExit, { viewJSON: views({ redRounds: 2, blueRounds: 1 }) }).verdict, 'PASS', 'a PASS exit has no final blue response — floored')
  // Without a view reader (no --bin) the audit SKIPs rather than reading a stale debate.md.
  assert.equal(recordParityAudit(healthy).verdict, 'SKIP')
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

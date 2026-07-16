#!/usr/bin/env node
// Run capture — the mechanical half of /research's run-record step, PLUS the mechanized
// post-hoc auditor. The run-4 report's attestation ceiling (§6.2) names "independent seats
// over git-tracked artifacts" as the enforcement tier for vacuity — a seat asserting work it
// did not do. This script IS that named auditor for the mechanical checks: it recomputes
// counts from the FILES and diffs them against the envelopes' self-reports.
//
// Usage: node run-capture.mjs <runDir> <workflow-transcript-dir>
// Writes <runDir>/capture-audit.md, <runDir>/cost.md, trajectories/journal.jsonl (copy),
// trajectories/agent-transcripts.tar.gz; removes the .run-live marker.
import { existsSync, readFileSync, writeFileSync, copyFileSync, readdirSync, rmSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { spawnSync } from 'node:child_process'
import { pathToFileURL, fileURLToPath } from 'node:url'

// Tolerant journal walk: collect every result object and any friction arrays inside.
export function readJournal(transcriptDir) {
  const p = join(transcriptDir, 'journal.jsonl')
  if (!existsSync(p)) return { results: [], friction: [] }
  const results = [], friction = []
  for (const line of readFileSync(p, 'utf8').split('\n')) {
    if (!line.trim()) continue
    let j; try { j = JSON.parse(line) } catch { continue }
    const r = j.result
    if (r && typeof r === 'object') {
      results.push(r)
      if (Array.isArray(r.friction)) friction.push(...r.friction)
    }
  }
  return { results, friction }
}

// AUDIT 1 — telemetry presence: one line per red round (a missing line is caught here; a
// WRONG line is beyond any presence check — recompute on actuation, per §2.5).
export function telemetryAudit(runDir) {
  const debate = existsSync(join(runDir, 'debate.md')) ? readFileSync(join(runDir, 'debate.md'), 'utf8') : ''
  const redRounds = (debate.match(/^### RED/gm) || []).length
  const p = join(runDir, 'trajectories', 'board-telemetry.jsonl')
  if (!existsSync(p)) {
    return { check: 'telemetry', verdict: redRounds === 0 ? 'SKIP' : 'FAIL', detail: `board-telemetry.jsonl absent; debate.md shows ${redRounds} red round(s)` }
  }
  const lines = readFileSync(p, 'utf8').split('\n').filter(Boolean)
  const rounds = new Set(lines.map((l) => { try { return JSON.parse(l).round } catch { return null } }).filter((r) => r != null))
  return {
    check: 'telemetry',
    verdict: rounds.size >= redRounds ? 'PASS' : 'FAIL',
    detail: `${rounds.size} telemetry round(s) vs ${redRounds} red round(s) in debate.md`,
  }
}

// AUDIT 2 — shard self-report vs files: the closure index (`id | class | summary | supersedes`
// lines in ledger.md) and heading-anchored archive records, recounted from disk and diffed
// against the last red envelope's integers. Heuristic counts, labeled as such.
export function shardAudit(runDir, results) {
  const ledgerP = join(runDir, 'red', 'ledger.md')
  const archiveP = join(runDir, 'red', 'archive.md')
  if (!existsSync(ledgerP) && !existsSync(archiveP)) return { check: 'shards', verdict: 'SKIP', detail: 'no ledger/archive (pre-sharding run)' }
  const idLine = /R\d+-\d+\s*\|/
  const ledgerCount = existsSync(ledgerP) ? readFileSync(ledgerP, 'utf8').split('\n').filter((l) => idLine.test(l)).length : 0
  const archiveCount = existsSync(archiveP) ? (readFileSync(archiveP, 'utf8').match(/^#{1,4}\s+.*R\d+-\d+/gm) || []).length : 0
  const lastRed = [...results].reverse().find((r) => typeof r.ledger_closure_lines === 'number' || typeof r.archive_blocks === 'number')
  const self = lastRed ? `self-reported ${lastRed.ledger_closure_lines}/${lastRed.archive_blocks}` : 'no envelope self-report found in journal'
  const consistent = !lastRed || (lastRed.ledger_closure_lines === ledgerCount && lastRed.archive_blocks === archiveCount)
  return {
    check: 'shards',
    verdict: consistent ? 'PASS' : 'FAIL',
    detail: `measured (heuristic) closure-index lines=${ledgerCount}, archive records=${archiveCount}; ${self}`,
  }
}

// AUDIT 3 — friction parity: every envelope friction entry must also live in friction.md
// (seats are instructed to append as they go; the file copy survives aborts — D24).
export function frictionAudit(runDir, envelopeFriction) {
  const p = join(runDir, 'friction.md')
  const file = existsSync(p) ? readFileSync(p, 'utf8') : ''
  const missing = envelopeFriction.filter((f) => !file.includes(String(f).slice(0, 60)))
  return {
    check: 'friction-parity',
    verdict: missing.length === 0 ? 'PASS' : 'FAIL',
    detail: missing.length ? `${missing.length} envelope entr${missing.length === 1 ? 'y' : 'ies'} missing from friction.md:\n${missing.map((m) => `    - ${String(m).slice(0, 120)}`).join('\n')}` : `${envelopeFriction.length} envelope entries all present in friction.md`,
  }
}

export function capture(runDir, transcriptDir) {
  const lines = []
  // Mechanics: journal copy, transcript tarball, cost.md (with telemetry join).
  copyFileSync(join(transcriptDir, 'journal.jsonl'), join(runDir, 'trajectories', 'journal.jsonl'))
  const agentFiles = readdirSync(transcriptDir).filter((f) => f.startsWith('agent-') && f.endsWith('.jsonl'))
  const tar = spawnSync('tar', ['czf', join(runDir, 'trajectories', 'agent-transcripts.tar.gz'), ...agentFiles], { cwd: transcriptDir })
  lines.push(`tarball: ${tar.status === 0 ? `${agentFiles.length} transcript(s)` : 'FAILED — ' + (tar.stderr || '').toString().slice(0, 200)}`)
  const cost = spawnSync(process.execPath, [join(dirname(fileURLToPath(import.meta.url)), 'cost-audit.mjs'), transcriptDir, runDir])
  if (cost.status === 0) { writeFileSync(join(runDir, 'cost.md'), cost.stdout); lines.push('cost.md: written (telemetry join included)') }
  else lines.push(`cost.md: FAILED — ${(cost.stderr || '').toString().slice(0, 200)}`)

  // The audits (journal read from the copy just made — the git-tracked artifact).
  const { results, friction } = readJournal(join(runDir, 'trajectories'))
  const audits = [telemetryAudit(runDir), shardAudit(runDir, results), frictionAudit(runDir, friction)]

  // Marker removal: the run is no longer live; hook guards stand down.
  const marker = join(process.cwd(), '.claude', 'run-live.json')
  if (existsSync(marker)) { rmSync(marker); lines.push('run-live marker: removed') }

  const report = [
    '# capture-audit — mechanized post-hoc checks (run-capture.mjs)',
    '',
    'Presence/consistency tier only: these checks catch a missing line and a self-inconsistent',
    'self-report; a plausible-but-wrong value is vacuity, whose auditor is the next run/',
    'retrospective over these same git-tracked artifacts.',
    '',
    ...audits.map((a) => `- **${a.check}: ${a.verdict}** — ${a.detail}`),
    '',
    ...lines.map((l) => `- ${l}`),
    '',
  ].join('\n')
  writeFileSync(join(runDir, 'capture-audit.md'), report)
  return { audits, report }
}

function main() {
  const [runDir, transcriptDir] = process.argv.slice(2)
  if (!runDir || !transcriptDir) { console.error('usage: node run-capture.mjs <runDir> <workflow-transcript-dir>'); process.exit(1) }
  const { audits, report } = capture(runDir, transcriptDir)
  console.log(report)
  process.exitCode = audits.some((a) => a.verdict === 'FAIL') ? 2 : 0
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main()

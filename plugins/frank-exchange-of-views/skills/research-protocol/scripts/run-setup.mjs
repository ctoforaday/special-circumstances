#!/usr/bin/env node
// Run setup — the mechanical half of /research steps 1-3, as a script instead of prose.
// Doctrine (design-by-contract, "a script must do what a script can do"): prose in a command
// is for decisions (topic, cited corpora, launch); mechanics executed by an LLM are an
// unenforced good-faith contract — the exact failure class red keeps catching in the engine.
//
// Usage:
//   node run-setup.mjs <runDir> --topic "<topic>" [--cite <path>[@<pin>]]... [--memory-dir <dir>] [--no-qmd]
//
// Idempotent: existing files are never overwritten (pre-staged inputs survive). Creates the
// blackboard skeleton (NOT red/ledger.md, red/archive.md, or board-telemetry.jsonl — those
// are red-merge-born; single-creator provenance is a run-4 §4.5 ratification condition),
// pins the evidence base, mirrors red's gap-pattern memory into inputs/, writes the
// .run-live marker (consumed by hook guards; removed by run-capture), and refreshes the
// qmd recall index when installed.
import { existsSync, mkdirSync, writeFileSync, readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { spawnSync } from 'node:child_process'

// Stub files with a one-line header each — the write-guard workaround (agents append to
// existing artifacts; the guard is filename-keyed and path-independent, run-3 experiment).
const STUBS = [
  ['debate.md', 'debate.md'],
  ['report.md', 'report.md'],
  ['friction.md', 'friction.md'],
  ['blue/frontier.md', 'blue frontier'],
  ['blue/report.md', 'blue report'],
  ['blue/CHANGELOG.md', 'blue CHANGELOG'],
  ['red/citation-ledger.md', 'red citation-ledger'],
]
const DIRS = ['blue/candidates', 'red/candidates', 'trajectories', 'inputs']

export function buildSkeleton(runDir, topic) {
  const created = [], skipped = []
  for (const d of DIRS) mkdirSync(join(runDir, d), { recursive: true })
  for (const [rel, name] of STUBS) {
    const p = join(runDir, rel)
    if (existsSync(p)) { skipped.push(rel); continue }
    writeFileSync(p, `# ${name} — ${topic}\n`)
    created.push(rel)
  }
  return { created, skipped }
}

export function buildPinned(runDir, head, cites) {
  const p = join(runDir, 'inputs', 'PINNED.md')
  if (existsSync(p)) return { written: false, path: p }
  const rows = cites.map((c) => {
    const [path, pin] = c.split('@')
    return `| \`${path}\` | \`${pin || head}\` |`
  })
  writeFileSync(p, [
    '# PINNED — the evidence base, frozen at launch',
    '',
    'Agents MUST cite these commits for repo and cross-corpus references; the evidence base',
    'does not move under the report. Pushes to cited paths are frozen while the run is live',
    '(.run-live marker; hook-guarded).',
    '',
    '| Corpus | Pin |',
    '|---|---|',
    `| Repo HEAD at launch | \`${head}\` |`,
    ...rows,
    '',
  ].join('\n'))
  return { written: true, path: p }
}

// Red's gap-pattern memory, mirrored to a path every seat can read (run-4 friction: the
// pre-flight MUST was unsatisfiable at four seat classes). Source of truth stays the agent
// memory; this is a read-only staged copy.
export function mirrorGapPatterns(memoryDir, runDir) {
  const out = join(runDir, 'inputs', 'red-gap-patterns.md')
  if (existsSync(out)) return { written: false, reason: 'already staged' }
  if (!memoryDir || !existsSync(memoryDir)) return { written: false, reason: 'no memory dir' }
  const parts = []
  for (const f of readdirSync(memoryDir).filter((f) => f.endsWith('.md'))) {
    parts.push(`\n<!-- mirrored from agent memory: ${f} -->\n` + readFileSync(join(memoryDir, f), 'utf8'))
  }
  if (!parts.length) return { written: false, reason: 'memory dir empty' }
  writeFileSync(out, '# red gap-pattern inventory (mirrored at run setup — read-only copy)\n' + parts.join('\n'))
  return { written: true, files: parts.length }
}

// The commitment-as-state pattern: promises about future behavior (push freeze, no mid-run
// plugin updates) become an observable marker that deterministic guards consult.
export function writeRunLiveMarker(projectDir, runDir, pinnedPaths) {
  const dir = join(projectDir, '.claude')
  mkdirSync(dir, { recursive: true })
  const p = join(dir, 'run-live.json')
  writeFileSync(p, JSON.stringify({ runDir, pinnedPaths, started: new Date().toISOString() }, null, 2) + '\n')
  return p
}

export function qmdRefresh(bin = 'qmd') {
  const probe = spawnSync(bin, ['--version'], { shell: process.platform === 'win32' })
  if (probe.error || probe.status !== 0) return { ran: false, reason: 'qmd not installed (optional — doctor installs it)' }
  const upd = spawnSync(bin, ['update'], { shell: process.platform === 'win32' })
  const emb = spawnSync(bin, ['embed'], { shell: process.platform === 'win32' })
  return { ran: true, update: upd.status === 0, embed: emb.status === 0 }
}

function main() {
  const [runDir, ...rest] = process.argv.slice(2)
  if (!runDir || runDir.startsWith('--')) {
    console.error('usage: node run-setup.mjs <runDir> --topic "<topic>" [--cite <path>[@pin]]... [--memory-dir <dir>] [--no-qmd]')
    process.exit(1)
  }
  const arg = (name) => { const i = rest.indexOf(name); return i >= 0 ? rest[i + 1] : null }
  const topic = arg('--topic') || '(topic not stated)'
  const cites = rest.flatMap((a, i) => (a === '--cite' ? [rest[i + 1]] : []))
  const head = (spawnSync('git', ['rev-parse', '--short', 'HEAD']).stdout || '').toString().trim() || 'unknown'

  const skel = buildSkeleton(runDir, topic)
  const pinned = buildPinned(runDir, head, cites)
  const memDir = arg('--memory-dir') || join(process.cwd(), '.claude', 'agent-memory', 'frank-exchange-of-views-red-auditor')
  const mirror = mirrorGapPatterns(memDir, runDir)
  const marker = writeRunLiveMarker(process.cwd(), runDir, cites.map((c) => c.split('@')[0]))
  const qmd = rest.includes('--no-qmd') ? { ran: false, reason: 'skipped (--no-qmd)' } : qmdRefresh()

  console.log(`run-setup: ${runDir}`)
  console.log(`  skeleton: ${skel.created.length} created, ${skel.skipped.length} pre-staged (kept)`)
  console.log(`  NOT created (red-merge-born): red/ledger.md, red/archive.md, trajectories/board-telemetry.jsonl`)
  console.log(`  pinned: ${pinned.written ? `HEAD ${head} + ${cites.length} cited path(s)` : 'inputs/PINNED.md pre-staged (kept)'}`)
  console.log(`  gap-patterns: ${mirror.written ? `${mirror.files} memory file(s) mirrored` : mirror.reason}`)
  console.log(`  run-live marker: ${marker}`)
  console.log(`  qmd refresh: ${qmd.ran ? `update ${qmd.update ? 'ok' : 'FAILED'}, embed ${qmd.embed ? 'ok' : 'FAILED'}` : qmd.reason}`)
}

import { pathToFileURL } from 'node:url'
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main()

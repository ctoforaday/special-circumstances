#!/usr/bin/env node
// Run setup — the mechanical half of /research steps 1-3, as a script instead of prose.
// Doctrine (design-by-contract, "a script must do what a script can do"): prose in a command
// is for decisions (topic, cited corpora, launch); mechanics executed by an LLM are an
// unenforced good-faith contract — the exact failure class red keeps catching in the engine.
//
// Usage:
//   node setup-research-run.mjs <runDir> --topic "<topic>" [--cite <path>[@<pin>]]... [--memory-dir <dir>] [--no-qmd]
//
// Idempotent: existing files are never overwritten (pre-staged inputs survive). Creates the
// blackboard skeleton (NOT red/ledger.md, red/archive.md, or board-telemetry.jsonl — those
// are red-merge-born; single-creator provenance is a run-4 §4.5 ratification condition),
// pins the evidence base, mirrors red's gap-pattern memory into inputs/, writes the
// .run-live marker (consumed by hook guards; removed by run-capture), and refreshes the
// qmd recall index when installed.
import { existsSync, mkdirSync, writeFileSync, readdirSync, readFileSync, rmSync, statSync } from 'node:fs'
import { homedir } from 'node:os'
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
const DIRS = ['blue/candidates', 'red/candidates', 'trajectories', 'inputs', 'records']

// Record-tool housekeeping (plan §III R2): stale checkpoint mirrors from crashed
// runs are purged after 30 days — the ONLY deletion anywhere in the record
// system, and it touches recovery copies of runs long since captured or dead.
export function purgeStaleMirrors(mirrorRoot, now = Date.now(), maxAgeDays = 30) {
  if (!existsSync(mirrorRoot)) return { purged: 0 }
  let purged = 0
  for (const d of readdirSync(mirrorRoot)) {
    const p = join(mirrorRoot, d)
    try {
      if (now - statSync(p).mtimeMs > maxAgeDays * 86400e3) { rmSync(p, { recursive: true, force: true }); purged++ }
    } catch {}
  }
  return { purged }
}

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

// W2e — the law corpus, mirrored read-only into the run (statute > precedent >
// argument; precedent is argument never evidence; persuasive-until-affirmed —
// see law/README.md). The bench reads inputs/law/ at every sitting.
export function mirrorLaw(repoLawDir, runDir) {
  const outDir = join(runDir, 'inputs', 'law')
  if (existsSync(outDir)) return { written: false, reason: 'already staged' }
  if (!repoLawDir || !existsSync(repoLawDir)) return { written: false, reason: 'no law dir' }
  mkdirSync(outDir, { recursive: true })
  let files = 0
  for (const f of readdirSync(repoLawDir).filter((x) => x.endsWith('.md'))) {
    writeFileSync(join(outDir, f), '<!-- mirrored from law/ at run setup — read-only copy -->\n' + readFileSync(join(repoLawDir, f), 'utf8'))
    files++
  }
  return { written: files > 0, files }
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

// W1.1 — pin validation (R1-7, judge-r2 ruling: "setup tooling must validate every pinned
// path and fail loudly on a miss, or stage cross-corpus artifacts into inputs/"). A cite
// asserting a path that does not exist at its pin poisons every citation of that corpus for
// the whole run — snapshot-grade at best, and the defect recurs every run until caught here.
export function validatePins(cites, head, git = (args) => spawnSync('git', args)) {
  if (!head || head === 'unknown') return { checked: 0, missing: [], skipped: 'not a git repo — pins UNVALIDATED' }
  const missing = []
  for (const c of cites) {
    const [path, pin] = c.split('@')
    const at = pin || head
    const r = git(['cat-file', '-e', `${at}:${path}`])
    if (r.error || r.status !== 0) missing.push({ path, pin: at })
  }
  return { checked: cites.length, missing }
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

// Single command string with shell (not args + shell): Windows needs the shell for the
// npm .cmd shim, and node deprecates args-array-with-shell (DEP0190, smoke finding #3).
// bin is caller-controlled ('qmd' or an injected test double), never user input.
export function qmdRefresh(bin = 'qmd') {
  const sh = { shell: true }
  const probe = spawnSync(`${bin} --version`, sh)
  if (probe.error || probe.status !== 0) return { ran: false, reason: 'qmd not installed (optional — doctor installs it)' }
  const upd = spawnSync(`${bin} update`, sh)
  const emb = spawnSync(`${bin} embed`, sh)
  return { ran: true, update: upd.status === 0, embed: emb.status === 0 }
}

function main() {
  const [runDir, ...rest] = process.argv.slice(2)
  if (!runDir || runDir.startsWith('--')) {
    console.error('usage: node setup-research-run.mjs <runDir> --topic "<topic>" [--cite <path>[@pin]]... [--memory-dir <dir>] [--no-qmd]')
    process.exit(1)
  }
  const arg = (name) => { const i = rest.indexOf(name); return i >= 0 ? rest[i + 1] : null }
  const topic = arg('--topic') || '(topic not stated)'
  const cites = rest.flatMap((a, i) => (a === '--cite' ? [rest[i + 1]] : []))
  const head = (spawnSync('git', ['rev-parse', '--short', 'HEAD']).stdout || '').toString().trim() || 'unknown'

  // Pins validated BEFORE anything is built: a bad cite fails the whole setup loudly.
  const pv = validatePins(cites, head)
  if (pv.missing && pv.missing.length) {
    console.error('run-setup: PIN VALIDATION FAILED — refusing to create the run:')
    for (const m of pv.missing) console.error(`  - ${m.path} does not exist at pin ${m.pin} (git cat-file -e ${m.pin}:${m.path})`)
    console.error('  remedies: fix the cite (right path / right pin), or stage the artifact into')
    console.error('  <runDir>/inputs/ BEFORE setup and cite the staged copy (setup keeps pre-staged files).')
    process.exit(2)
  }

  const skel = buildSkeleton(runDir, topic)
  const mirrors = purgeStaleMirrors(join(homedir(), '.cache', 'feov', 'run-mirror'))
  if (mirrors.purged) console.log(`  mirror purge: ${mirrors.purged} stale checkpoint mirror(s) removed`)
  const pinned = buildPinned(runDir, head, cites)
  // Red's agent memory accrues under the LAUNCHING session's project (smoke finding #2):
  // CLAUDE_PROJECT_DIR names it when set; cwd is the repo-local fallback; --memory-dir wins.
  const memHome = (d) => join(d, '.claude', 'agent-memory', 'frank-exchange-of-views-red-auditor')
  const memDir = arg('--memory-dir') ||
    [process.env.CLAUDE_PROJECT_DIR, process.cwd()].filter(Boolean).map(memHome).find(existsSync)
  const mirror = mirrorGapPatterns(memDir, runDir)
  const law = mirrorLaw(join(process.cwd(), 'law'), runDir)
  const marker = writeRunLiveMarker(process.cwd(), runDir, cites.map((c) => c.split('@')[0]))
  const qmd = rest.includes('--no-qmd') ? { ran: false, reason: 'skipped (--no-qmd)' } : qmdRefresh()

  console.log(`run-setup: ${runDir}`)
  console.log(`  skeleton: ${skel.created.length} created, ${skel.skipped.length} pre-staged (kept)`)
  console.log(`  NOT created (red-merge-born): red/ledger.md, red/archive.md, trajectories/board-telemetry.jsonl`)
  console.log(`  pinned: ${pinned.written ? `HEAD ${head} + ${cites.length} cited path(s)` : 'inputs/PINNED.md pre-staged (kept)'}`)
  console.log(`  pin validation: ${pv.skipped ? pv.skipped : `${pv.checked} cite(s) verified at their pins`}`)
  console.log(`  gap-patterns: ${mirror.written ? `${mirror.files} memory file(s) mirrored` : mirror.reason}`)
  console.log(`  law: ${law.written ? `${law.files} file(s) mirrored (statute > precedent > argument)` : law.reason}`)
  console.log(`  run-live marker: ${marker}`)
  console.log(`  qmd refresh: ${qmd.ran ? `update ${qmd.update ? 'ok' : 'FAILED'}, embed ${qmd.embed ? 'ok' : 'FAILED'}` : qmd.reason}`)
}

import { pathToFileURL } from 'node:url'
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main()

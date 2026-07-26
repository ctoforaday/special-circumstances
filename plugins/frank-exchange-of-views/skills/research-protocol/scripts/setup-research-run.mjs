#!/usr/bin/env node
// Run setup — the mechanical half of /research steps 1-3, as a script instead of prose.
// Doctrine (design-by-contract, "a script must do what a script can do"): prose in a command
// is for decisions (topic, cited corpora, launch); mechanics executed by an LLM are an
// unenforced good-faith contract — the exact failure class red keeps catching in the engine.
//
// Usage:
//   node setup-research-run.mjs <runDir> --topic "<topic>" [--cite <path>[@<pin>]]... [--bin-dir <dir>] [--memory-dir <dir>] [--no-qmd]
//
// Idempotent: existing files are never overwritten (pre-staged inputs survive). Creates the
// blackboard skeleton (NOT red/ledger.md, red/archive.md, or board-telemetry.jsonl — those
// are red-merge-born; single-creator provenance is a run-4 §4.5 ratification condition),
// pins the evidence base, mirrors red's gap-pattern memory into inputs/, writes the
// .run-live marker (consumed by hook guards; removed by run-capture), and refreshes the
// qmd recall index when installed.
import { existsSync, mkdirSync, writeFileSync, readdirSync, readFileSync, rmSync, statSync } from 'node:fs'
import { homedir } from 'node:os'
import { join, dirname } from 'node:path'
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { parseRenderedRows } from './scorecards.mjs'

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
const DIRS = ['blue/candidates', 'red', 'trajectories', 'inputs', 'records']

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
  const dirs = (Array.isArray(memoryDir) ? memoryDir : [memoryDir]).filter((d) => d && existsSync(d))
  if (!dirs.length) return { written: false, reason: 'no memory dir' }
  const parts = []
  const seen = new Set()
  for (const dir of dirs) {
    for (const f of readdirSync(dir).filter((f) => f.endsWith('.md') && f !== 'README.md')) {
      // First source wins: the PROMOTED corpus is authoritative, and raw accrual
      // only contributes patterns that have not been promoted yet. A file present
      // in both must not be staged twice — red would read the same pattern under
      // two headings and could not tell which one the run is holding it to.
      if (seen.has(f)) continue
      seen.add(f)
      parts.push(`\n<!-- mirrored from ${dir}: ${f} -->\n` + readFileSync(join(dir, f), 'utf8'))
    }
  }
  if (!parts.length) return { written: false, reason: 'memory dir empty' }
  writeFileSync(out, '# red gap-pattern inventory (mirrored at run setup — read-only copy)\n' + parts.join('\n'))
  return { written: true, files: parts.length, sources: dirs.length }
}

// W1.1 — pin validation (R1-7, judge-r2 ruling: "setup tooling must validate every pinned
// path and fail loudly on a miss, or stage cross-corpus artifacts into inputs/"). A cite
// asserting a path that does not exist at its pin poisons every citation of that corpus for
// the whole run — snapshot-grade at best, and the defect recurs every run until caught here.
// W2h — the visibility loop's second leg: stage each chair's scorecard where the
// run can see it, and hand back the HEADLINE numbers for the prompts.
//
// The file mirror is for the human and the audit trail. The headline is what
// actually reaches a seat, because E0.5 measured what staging alone is worth:
// run 5's lanes verifiably READ the staged gap patterns and committed both warned
// patterns anyway. A number a seat must go and look up is a number it will not
// look up; a number in its prompt is one it cannot avoid.
export function mirrorScorecards(memoryDir, runDir) {
  if (!memoryDir || !existsSync(memoryDir)) return { written: false, reason: 'no feov-memory dir', headlines: {} }
  const staged = []
  const headlines = {}
  for (const f of readdirSync(memoryDir).filter((f) => f.endsWith('-scorecard.md'))) {
    const chair = f.replace('-scorecard.md', '')
    const body = readFileSync(join(memoryDir, f), 'utf8')
    writeFileSync(join(runDir, 'inputs', f), body)
    staged.push(chair)
    // The LAST section is the most recent run; its bolded values are the
    // headline. Parsed from the rendered file rather than recomputed, so the
    // seat sees exactly what the dashboard and the human see.
    const sections = body.split(/^## /m)
    const latest = sections[sections.length - 1] || ''
    // PREFERRED: the emitted HEADLINE line, which scorecards.mjs computes with
    // headline() — one authority for what a seat sees, correctly RANKED (a tripped
    // detector first, never displaced by a benchmark that merely sits higher).
    const emitted = /^HEADLINE:\s*(.+)$/m.exec(latest)
    if (emitted) {
      headlines[chair] = emitted[1].split(' · ').map((s) => s.trim()).filter(Boolean).slice(0, 3)
      continue
    }
    // FALLBACK for scorecards written before the HEADLINE line existed. It uses
    // scorecards.mjs's parser rather than a fourth private regex — that
    // duplication is what hid a detector from every prompt in the first place.
    // `measure` is dropped here on purpose: it is recorded, never targeted, so it
    // is not headline material (headline() makes the same exclusion).
    const picks = parseRenderedRows(latest)
      .filter((r) => r.value !== null && r.cls !== 'measure')
      .map((r) => `${r.metric} ${r.value.trim()} [${r.cls.toUpperCase()}]`)
    if (picks.length) headlines[chair] = picks.slice(0, 3)
  }
  // A run-1 corpus legitimately has no scorecards: they are WRITTEN at capture and
  // READ by the next run, so the visibility loop cannot close on its own first turn.
  // That is an expected state, not a failure, and it must say so — the caller prints
  // this reason, and an absent one printed a bare "undefined" that read like a bug.
  if (!staged.length) return { written: false, reason: 'no scorecards yet — written at capture, consumed by the next run', chairs: [], headlines: {} }
  return { written: true, chairs: staged, headlines }
}

// buildPatternIndex — the class join's lookup table.
//
// MEMORY THAT BINDS IS DELIVERED, NOT STAGED. E0.5 settled this empirically: run
// 4's "read red's gap patterns" clause was unsatisfiable at four blue seats, and
// run 5 was worse — lanes verifiably READ the staged file and committed both
// warned patterns anyway. Fifty-odd patterns at seat start is a salience problem,
// not a knowledge problem.
//
// So patterns are indexed BY CLASS here, and the engine hands a repairing seat
// only the patterns whose class matches the gap in front of it. That makes the
// duty deterministic (a join, not a search), auditable (the manifest row records
// which patterns applied and what checking them showed), and small enough to
// survive the seat's attention.
//
// A pattern with no classes is deliberately NOT delivered. Promotion into the
// corpus is where classification happens, and an unclassified pattern is a
// promotion that skipped its review step rather than a pattern that applies
// everywhere.
export function buildPatternIndex(dirs) {
  const byClass = {}
  const unclassified = []
  const harnessLimit = []
  // FIRST SOURCE WINS, exactly as mirrorGapPatterns does it — the promoted corpus is
  // authoritative and the raw accrual path holds PRE-PROMOTION copies of the same
  // filenames. Without this, every promoted pattern was counted twice: once classified
  // from feov-memory/, once classless from .claude/agent-memory/, and setup reported a
  // 55-item classification backlog that did not exist. Two readers of one corpus must
  // agree on identity; when only one of them dedups, the other one lies.
  const seen = new Set()
  for (const dir of (Array.isArray(dirs) ? dirs : [dirs]).filter((d) => d && existsSync(d))) {
    for (const f of readdirSync(dir).filter((f) => f.endsWith('.md') && f !== 'README.md' && f !== 'MEMORY.md')) {
      if (seen.has(f)) continue
      seen.add(f)
      const body = readFileSync(join(dir, f), 'utf8')
      const fm = /^---\n([\s\S]*?)\n---/.exec(body)
      const title = (/^#\s+(.+)$/m.exec(body) || [, f])[1]
      const desc = fm ? (/^description:\s*(.+)$/m.exec(fm[1]) || [, ''])[1].replace(/^["']|["']$/g, '') : ''
      const classLine = fm ? /^\s*classes:\s*\[([^\]]*)\]/m.exec(fm[1]) : null
      const classes = classLine ? classLine[1].split(',').map((c) => c.trim().replace(/^["']|["']$/g, '')).filter(Boolean) : []
      // A harness-limit pattern is DELIBERATELY classless: it records a tooling
      // constraint, not a kind of research defect, so it has no gap class to join
      // against and must not be reported as unfinished classification work. The
      // two are different states and conflating them would make the honest
      // backlog number lie.
      if (!classes.length) {
        (/^\s*class_note:\s*harness-limit/m.test(fm ? fm[1] : '') ? harnessLimit : unclassified).push(f)
        continue
      }
      for (const c of classes) (byClass[c] ||= []).push({ file: f, title, hook: desc })
    }
  }
  return { byClass, unclassified, harnessLimit }
}

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
// The record binary is checked BEFORE the run exists, not when a seat first
// reaches for it. A missing or version-skewed feov-record discovered mid-round
// costs a seat: the agent is already dispatched, its context is spent, and its
// record for that round is simply absent — while the same failure at setup costs
// nothing but a re-run. (The suite's own rule that plugins are never updated
// mid-run exists for the same reason; this is its enforceable half.)
//
// Skew is reported, not just absence: the binary is released per plugin tag, so
// a stale one on PATH silently writes events under an older contract.
// The plugin manifest is the version authority — hardcoding it here would create
// exactly the skew this preflight exists to catch.
export function recordToolVersion(pluginJson = null) {
  try {
    const p = pluginJson || join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..', '.claude-plugin', 'plugin.json')
    return JSON.parse(readFileSync(p, 'utf8')).recordToolVersion || null
  } catch { return null }
}

export function preflightRecordBinary(expectVersion, bin = 'feov-record', run = (b, a) => spawnSync(b, a)) {
  const r = run(bin, ['--version'])
  if (r.error || r.status !== 0) {
    return {
      ok: false,
      reason: `${bin} not runnable (${r.error ? r.error.code || r.error.message : `exit ${r.status}`})`,
      remedy: 'install it with /prosthetic-conscience:doctor --fix, or build it: ' +
        '(cd plugins/frank-exchange-of-views/tools && go build ./cmd/feov-record)',
    }
  }
  const got = (r.stdout || '').toString().trim().split(/\s+/).pop() || ''
  if (expectVersion && got && got !== expectVersion) {
    return {
      ok: false,
      reason: `${bin} is ${got}, the plugin expects ${expectVersion} — a skewed binary writes events under a different contract`,
      remedy: 'refresh it with /prosthetic-conscience:doctor --fix (never mid-run)',
    }
  }
  return { ok: true, version: got }
}

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
    console.error('usage: node setup-research-run.mjs <runDir> --topic "<topic>" [--cite <path>[@pin]]... [--bin-dir <dir>] [--memory-dir <dir>] [--no-qmd]')
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
    console.error('  <runDir>/inputs/ AND COMMIT IT before setup, then cite the committed copy.')
    console.error('  Staging alone is NOT enough: an uncommitted file exists at no pin, so it')
    console.error('  cannot be cited — evidence that can still change underneath the run is not')
    console.error('  evidence. (setup keeps pre-staged files; committing is the missing step.)')
    process.exit(2)
  }

  // Record-binary preflight, before ANY run state is created — but bound to
  // INTENT rather than to mere availability. --bin-dir is how the operator says
  // this run will record through the tool (it is the same value the engine takes
  // as binDir), so a missing or skewed binary there is fatal: the run cannot do
  // what it was asked to do. Without --bin-dir the run is not using the tool, and
  // a hard failure would block legacy and no-record runs over a dependency they
  // never take. Absence is still REPORTED, because silence would let a run
  // intended to record start without the means to.
  const binDir = arg('--bin-dir')
  const recordBin = binDir ? join(binDir, 'feov-record') : 'feov-record'
  const pre = preflightRecordBinary(recordToolVersion(), recordBin)
  if (!pre.ok && binDir) {
    console.error('run-setup: RECORD BINARY PREFLIGHT FAILED — refusing to create the run:')
    console.error(`  ${pre.reason}`)
    console.error(`  remedy: ${pre.remedy}`)
    console.error('  (failing here costs a re-run; failing mid-round costs a seat its whole record)')
    process.exit(2)
  }

  const skel = buildSkeleton(runDir, topic)
  const mirrors = purgeStaleMirrors(join(homedir(), '.cache', 'feov', 'run-mirror'))
  if (mirrors.purged) console.log(`  mirror purge: ${mirrors.purged} stale checkpoint mirror(s) removed`)
  const pinned = buildPinned(runDir, head, cites)
  // Red's agent memory accrues under the LAUNCHING session's project (smoke finding #2):
  // CLAUDE_PROJECT_DIR names it when set; cwd is the repo-local fallback; --memory-dir wins.
  // TWO TIERS, promoted first. feov-memory/ is the tracked, reviewed corpus that
  // binds seats; .claude/agent-memory/ is the harness's raw accrual path, which is
  // machine-local and gitignored. Sourcing only from the latter is what nearly
  // started red amnesiac at the 2026-07-18 cwd move — a corpus that survives on
  // one disk is not a corpus. Clone the repo, get the memory.
  const memHome = (d) => join(d, '.claude', 'agent-memory', 'frank-exchange-of-views-red-auditor')
  const promoted = join(process.cwd(), 'feov-memory', 'red-gap-patterns')
  const raw = [process.env.CLAUDE_PROJECT_DIR, process.cwd()].filter(Boolean).map(memHome).find(existsSync)
  const memDirs = arg('--memory-dir') ? [arg('--memory-dir')] : [promoted, raw]
  const mirror = mirrorGapPatterns(memDirs, runDir)
  const patternIndex = buildPatternIndex(memDirs)
  writeFileSync(join(runDir, 'inputs', 'gap-patterns-by-class.json'), JSON.stringify(patternIndex.byClass, null, 2) + '\n')
  // Canonical run config — models per stage, the round ceiling, lane count. The engine
  // (debate.js) takes these as launch args but is sandboxed and cannot write them down; setup
  // can, so post-hoc readers (the dashboard) read one file instead of being told separately.
  const runConfig = { topic, model: arg('--model'), judgmentModel: arg('--judgment-model'), maxRounds: arg('--max-rounds'), lanes: arg('--lanes') }
  writeFileSync(join(runDir, 'inputs', 'run-config.json'), JSON.stringify(runConfig, null, 2) + '\n')
  const law = mirrorLaw(join(process.cwd(), 'law'), runDir)
  const cards = mirrorScorecards(join(process.cwd(), 'feov-memory'), runDir)
  const marker = writeRunLiveMarker(process.cwd(), runDir, cites.map((c) => c.split('@')[0]))
  const qmd = rest.includes('--no-qmd') ? { ran: false, reason: 'skipped (--no-qmd)' } : qmdRefresh()

  console.log(`run-setup: ${runDir}`)
  console.log(`  skeleton: ${skel.created.length} created, ${skel.skipped.length} pre-staged (kept)`)
  console.log(`  NOT created (red-merge-born): red/ledger.md, red/archive.md, trajectories/board-telemetry.jsonl`)
  console.log(`  pinned: ${pinned.written ? `HEAD ${head} + ${cites.length} cited path(s)` : 'inputs/PINNED.md pre-staged (kept)'}`)
  console.log(`  pin validation: ${pv.skipped ? pv.skipped : `${pv.checked} cite(s) verified at their pins`}`)
  console.log(`  gap-patterns: ${mirror.written ? `${mirror.files} pattern(s) mirrored from ${mirror.sources} source(s) (promoted corpus first)` : mirror.reason}`)
  console.log(`  law: ${law.written ? `${law.files} file(s) mirrored (statute > precedent > argument)` : law.reason}`)
  console.log(`  gap-pattern index: ${Object.keys(patternIndex.byClass).length} class(es) -> inputs/gap-patterns-by-class.json${patternIndex.unclassified.length ? ` (${patternIndex.unclassified.length} UNCLASSIFIED, not delivered — classify them to make them bind)` : ''}${patternIndex.harnessLimit.length ? ` (${patternIndex.harnessLimit.length} harness-limit, classless by design)` : ''}`)
  console.log(`  scorecards: ${cards.written ? `${cards.chairs.join(', ')} staged into inputs/` : cards.reason}`)
  if (Object.keys(cards.headlines).length) {
    // Printed as a ready-to-pass workflow arg: debate.js is SANDBOXED (zero
    // imports) and cannot read this file itself, so the numbers travel as an
    // argument rather than as a path the engine would have to open.
    // Written as a FILE as well as printed. The workflow script cannot read it —
    // verified empirically, not assumed: a zero-agent probe found require,
    // process, fetch and import() all absent, with import() explicitly refused
    // ("import() is not available in workflow scripts"). But whoever LAUNCHES the
    // run can read it, and a machine-readable file beats a human retyping a
    // printed line into an argument.
    writeFileSync(join(runDir, 'inputs', 'scorecards.json'), JSON.stringify(cards.headlines, null, 2) + '\n')
    console.log(`  scorecards arg: pass inputs/scorecards.json as the workflow's "scorecards" arg`)
    console.log(`    ${JSON.stringify(cards.headlines)}`)
  }
  console.log(`  record binary: ${pre.ok ? `${recordBin} ${pre.version || '(version unreported)'}` : `NOT AVAILABLE — ${pre.reason} (this run will not record through the tool; pass --bin-dir to require it)`}`)
  console.log(`  run-live marker: ${marker}`)
  console.log(`  qmd refresh: ${qmd.ran ? `update ${qmd.update ? 'ok' : 'FAILED'}, embed ${qmd.embed ? 'ok' : 'FAILED'}` : qmd.reason}`)
}

import { pathToFileURL } from 'node:url'
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main()

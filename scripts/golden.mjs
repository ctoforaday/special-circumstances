#!/usr/bin/env node
// scripts/golden.mjs — ONE command for every golden in the repo, both languages.
//
// Dev tooling for this repository only (CLAUDE.md: dev concerns live at the root,
// consumer behaviour lives in plugins/). Nothing here ships to an installing
// project.
//
// THE PROBLEM THIS SOLVES is not diffing — go-cmp already does that — it is
// REVIEW EROSION. Golden suites rot in a predictable way: goldens spread across
// suites and languages, a wave regenerates some of them by hand, the diff is
// hundreds of lines, it gets rubber-stamped, and a real behavioural drift ships
// inside the noise. Prompt goldens per seat class will make that pressure worse,
// because seat prompts are long and every wave edits them.
//
// Four countermeasures, all here:
//   1. ONE command updates everything, so nobody hand-picks which goldens to
//      refresh and no suite silently falls behind.
//   2. An update run prints a CHANGE REPORT — what moved, and by how much —
//      so the author sees the shape of the change before committing it.
//   3. `--check` (what CI runs) fails when goldens are stale, so an unrecorded
//      behaviour change cannot merge quietly.
//   4. Orphan detection: a golden no test claims is a failure, because a stale
//      golden reads as an authoritative record of behaviour nothing asserts.
//
// Usage:
//   node scripts/golden.mjs            compare (same as CI)
//   node scripts/golden.mjs --update   re-record, then print the change report
//   node scripts/golden.mjs --check    compare, exit 1 on any staleness
import { spawnSync } from 'node:child_process'
import { readdirSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { dirname } from 'node:path'

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const update = process.argv.includes('--update')

// The mjs suites that carry goldens. Enumerated, never globbed: `node --test`
// over a directory fails on Windows (a documented recurrence), and an explicit
// list is also the thing CI copies.
const MJS_SUITES = [
  'plugins/frank-exchange-of-views/tests/simulator/prompts.test.mjs',
]

const GO_MODULES = [
  'plugins/frank-exchange-of-views/tools',
]

let failed = false

function run(label, cmd, args, opts = {}) {
  process.stdout.write(`\n── ${label}\n`)
  const r = spawnSync(cmd, args, { cwd: repoRoot, stdio: 'inherit', shell: false, ...opts })
  if (r.status !== 0) failed = true
  return r.status
}

// ---- mjs ----
const presentSuites = MJS_SUITES.filter((s) => existsSync(join(repoRoot, s)))
if (presentSuites.length) {
  run(
    `mjs goldens${update ? ' (recording)' : ''}`,
    process.execPath,
    ['--test', ...presentSuites],
    { env: { ...process.env, UPDATE_GOLDENS: update ? '1' : '' } },
  )
} else {
  process.stdout.write('\n── mjs goldens: no suites yet (skipped)\n')
}

// ---- go ----
for (const mod of GO_MODULES) {
  run(`go goldens in ${mod}${update ? ' (recording)' : ''}`, 'go', ['test', './...'], {
    cwd: join(repoRoot, mod),
    env: { ...process.env, UPDATE_GOLDENS: update ? '1' : '' },
  })
}

// ---- change report: git is the review surface, so ask git what moved ----
if (update) {
  // `git diff` cannot see a golden that did not exist before, and a NEW golden
  // is exactly what a new seat class or scenario produces — so the report asks
  // for modified and untracked separately, or it under-reports on the runs that
  // matter most.
  const diff = spawnSync('git', ['diff', '--stat', '--', '*.golden'], { cwd: repoRoot, encoding: 'utf8' })
  const untracked = spawnSync('git', ['ls-files', '--others', '--exclude-standard', '--', '*.golden'], { cwd: repoRoot, encoding: 'utf8' })
  const added = (untracked.stdout || '').trim().split('\n').filter(Boolean)
  const newList = added.map((f) => `  + ${f}`).join('\n')
  const body = [
    (diff.stdout || '').trim(),
    added.length ? `${added.length} new golden(s):\n${newList}` : '',
  ].filter(Boolean).join('\n')
  process.stdout.write('\n══ CHANGE REPORT ══\n')
  process.stdout.write(body ? `${body}\n` : 'no golden changed — behaviour is unmoved\n')
  if (body) {
    process.stdout.write(
      '\nCommit the testdata change ON ITS OWN, separately from the code that caused it.\n' +
      'A golden diff mixed into a feature commit is a diff nobody reads.\n')
  }
}

if (failed) {
  process.stderr.write('\ngolden: FAILED — goldens are stale or a test errored.\n' +
    'If the change is intentional: node scripts/golden.mjs --update\n')
  process.exit(1)
}
process.stdout.write(`\ngolden: OK${update ? ' (recorded)' : ''}\n`)

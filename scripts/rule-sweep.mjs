#!/usr/bin/env node
// scripts/rule-sweep.mjs — a rule patch names its class and sweeps the siblings.
//
// Dev tooling for this repository only (CLAUDE.md: dev concerns live at the root).
// Nothing here ships to an installing project.
//
// THE PROBLEM. Every protocol patch that shipped recurred in adjacent form —
// lens labels patched, blue-lane footnote namespaces bit next run;
// blue-reads-transcript patched, the lossy gap summary bit next run; grade enums
// widened, the mass mapping distorted next run; friction-to-file shipped, lens
// seats still had no write path. Four for four. Each fix was correct and each was
// an INSTANCE fix: the class survived, re-emerged at the adjacent seat, and
// passed every test the patch had added, because those tests were about the
// instance too.
//
// THE RULE. A change to a PROTOCOL SURFACE must name its class from
// feov-memory/protocol-class-registry.md and state the sibling sweep:
//
//     Rule-Class: adjacent-seat-omission
//     Sibling-Sweep: checked blue-lane, blue-synthesize, blue-respond — all three
//                    carry the clause; red-merge does not share this duty
//
// WHY MECHANICAL. A sweep requirement enforced by good intentions is an instance
// of policy-without-mechanism, which is a class in the registry it would be
// enforcing. This script is the mechanism; CI runs it on every pull request.
//
// It is deliberately a PROMPT, not a proof: it cannot know whether a sweep was
// honest, only whether one was stated. That is the same tier as the round-parity
// check — presence, not truth — and it is the tier that stops the failure we
// actually measured, which was nobody looking at the siblings at all.
//
// Usage:
//   node scripts/rule-sweep.mjs               check HEAD against origin/main
//   node scripts/rule-sweep.mjs --base <ref>  check against another base
//   node scripts/rule-sweep.mjs --selftest    verify this script's own logic
import { spawnSync } from 'node:child_process'
import { readFileSync, existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..')

// PROTOCOL SURFACES: the files that govern how the engine behaves, as opposed to
// the code that implements it. A change to debate.js's dispatch arithmetic is
// engineering; a change to what a seat is TOLD is protocol, and it is the second
// kind whose fixes kept recurring.
export const PROTOCOL_SURFACES = [
  /\/skills\/[^/]+\/SKILL\.md$/,
  /\/agents\/[^/]+\.md$/,
  /\/references\/[^/]+\.md$/,
  /\/commands\/[^/]+\.md$/,
  /\/scripts\/debate\.js$/,
  /^law\//,
]

export function touchesProtocol(files) {
  return files.filter((f) => PROTOCOL_SURFACES.some((re) => re.test(f)))
}

export function knownClasses(registryPath) {
  if (!existsSync(registryPath)) return []
  // Classes are the ### headings of the registry — one source, so the checker and
  // the document cannot drift into disagreeing about what a valid class is.
  return [...readFileSync(registryPath, 'utf8').matchAll(/^###\s+([a-z][a-z0-9-]+)\s*$/gm)].map((m) => m[1])
}

// Trailers are read from the message's LAST paragraph only — the git trailer
// block — never from anywhere in the body.
//
// The first version scanned the whole message, and the commit introducing this
// gate exposed it: prose EXPLAINING the mechanism ("...must carry Rule-Class:
// and Sibling-Sweep: trailers...") satisfied the gate. A check that passes on a
// DESCRIPTION of the act rather than the act is precisely the defect class this
// gate exists to prevent, which made it an unusually direct way to learn it.
export function trailerBlock(message) {
  // Any paragraph in which EVERY line is a `Key: value` or an indented
  // continuation. Not merely the last one: attribution trailers
  // (Co-Authored-By, Claude-Session) conventionally sit last, so requiring the
  // final paragraph would force authors to interleave unrelated trailers — and a
  // rule that fights the convention it rides on gets worked around, not obeyed.
  //
  // The strictness that matters is EVERY line, not SOME: a prose paragraph
  // explaining the mechanism contains a line starting "Sibling-Sweep:" and must
  // not count, which is exactly how the first version of this check was fooled.
  return message.trim().split(/\n\s*\n/).filter((p) => {
    const lines = p.split('\n').filter((l) => l.trim())
    return lines.length > 0 && lines.every((l) => /^[A-Za-z][A-Za-z-]*:\s+\S/.test(l) || /^\s+\S/.test(l))
  }).join('\n')
}

export function extractTrailers(messages) {
  const cls = []
  const sweeps = []
  for (const m of messages) {
    for (const line of trailerBlock(m).split('\n')) {
      const c = /^Rule-Class:\s*(.+)$/i.exec(line.trim())
      if (c) cls.push(...c[1].split(',').map((s) => s.trim()).filter(Boolean))
      const s = /^Sibling-Sweep:\s*(.+)$/i.exec(line.trim())
      if (s && s[1].trim().length > 10) sweeps.push(s[1].trim())
    }
  }
  return { classes: cls, sweeps }
}

// evaluate is the pure core, so the self-test can exercise every branch without
// a git repository or a network.
export function evaluate({ files, messages, classes }) {
  const touched = touchesProtocol(files)
  if (!touched.length) return { ok: true, skipped: true, touched, reason: 'no protocol surface touched' }

  const { classes: named, sweeps } = extractTrailers(messages)
  const problems = []
  if (!named.length) {
    problems.push('no Rule-Class: trailer — name the class this defect belongs to (feov-memory/protocol-class-registry.md)')
  }
  const unknown = named.filter((c) => !classes.includes(c))
  if (unknown.length) {
    problems.push(`unknown class(es): ${unknown.join(', ')} — use a registry slug, or add the class WITH its definition, neighbour and distinguishing question`)
  }
  if (!sweeps.length) {
    problems.push('no Sibling-Sweep: trailer — state which adjacent seats or surfaces carry this same duty and what you found at each')
  }
  return { ok: problems.length === 0, skipped: false, touched, problems, named, sweeps }
}

function git(args) {
  const r = spawnSync('git', args, { cwd: repoRoot, encoding: 'utf8' })
  return r.status === 0 ? (r.stdout || '').trim() : ''
}

function selftest() {
  const classes = ['adjacent-seat-omission', 'unobservable-duty']
  const cases = [
    ['code-only change is not gated', { files: ['plugins/x/tools/internal/record/render.go'], messages: ['fix: thing'], classes }, true],
    ['protocol change with no trailers fails', { files: ['plugins/x/skills/y/SKILL.md'], messages: ['fix: thing'], classes }, false],
    ['protocol change with both trailers passes', {
      files: ['plugins/x/agents/red-auditor.md'],
      messages: ['fix: thing\n\nRule-Class: adjacent-seat-omission\nSibling-Sweep: checked blue and bench, neither carries this duty'],
      classes,
    }, true],
    ['an invented class fails', {
      files: ['law/README.md'],
      messages: ['x\n\nRule-Class: made-up-thing\nSibling-Sweep: checked everything everywhere'],
      classes,
    }, false],
    ['a token sweep fails', {
      files: ['plugins/x/skills/y/SKILL.md'],
      messages: ['x\n\nRule-Class: unobservable-duty\nSibling-Sweep: done'],
      classes,
    }, false],
    ['prose DESCRIBING the trailers does not satisfy them', {
      files: ['plugins/x/skills/y/SKILL.md'],
      messages: ['a change\n\nthe gate needs Rule-Class: adjacent-seat-omission and\nSibling-Sweep: a statement of what was checked\n\nbut this closing paragraph is prose, not a trailer block'],
      classes,
    }, false],
    ['seat prompts count as protocol', {
      files: ['plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js'],
      messages: ['x'],
      classes,
    }, false],
  ]
  let failed = 0
  for (const [name, input, expectOk] of cases) {
    const got = evaluate(input).ok
    const pass = got === expectOk
    if (!pass) failed++
    console.log(`${pass ? 'ok  ' : 'FAIL'} ${name}`)
  }
  return failed
}

function main() {
  if (process.argv.includes('--selftest')) process.exit(selftest() ? 1 : 0)

  const baseIdx = process.argv.indexOf('--base')
  const base = baseIdx >= 0 ? process.argv[baseIdx + 1] : 'origin/main'
  const range = `${base}...HEAD`
  const files = git(['diff', '--name-only', range]).split('\n').filter(Boolean)
  const messages = git(['log', '--format=%B%x00', range]).split('\0').filter((m) => m.trim())

  const registry = join(repoRoot, 'feov-memory', 'protocol-class-registry.md')
  const result = evaluate({ files, messages, classes: knownClasses(registry) })

  if (result.skipped) {
    console.log('rule-sweep: no protocol surface touched — nothing to sweep')
    return
  }
  console.log(`rule-sweep: ${result.touched.length} protocol surface(s) changed:`)
  for (const f of result.touched) console.log(`  ${f}`)

  if (result.ok) {
    console.log(`\nrule-sweep: OK — class ${result.named.join(', ')}`)
    for (const s of result.sweeps) console.log(`  sweep: ${s}`)
    return
  }
  console.error('\nrule-sweep: FAILED — a rule patch must name its class and sweep its siblings.\n')
  for (const p of result.problems) console.error(`  - ${p}`)
  console.error(`
Every protocol patch that shipped before this check recurred in adjacent form:
lens labels patched, blue-lane footnotes bit next run; blue-reads-transcript
patched, the lossy gap summary bit next run; grade enums widened, the mass
mapping distorted next run; friction-to-file shipped, lens seats still had no
write path. Four for four. The fixes were correct; they were INSTANCE fixes.

Add to any commit in this branch:

  Rule-Class: <slug from feov-memory/protocol-class-registry.md>
  Sibling-Sweep: <which adjacent seats or surfaces carry this same duty, and
                  what you found at each — including "none do", if that is true>
`)
  process.exit(1)
}

if (process.argv[1] && import.meta.url === new URL(`file://${process.argv[1].replace(/\\/g, '/')}`).href) main()
else if (process.argv[1] && process.argv[1].endsWith('rule-sweep.mjs')) main()

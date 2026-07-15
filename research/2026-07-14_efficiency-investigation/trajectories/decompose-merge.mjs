#!/usr/bin/env node
// Merge-context decomposition for a debate run — the [^MergeDecomposition] instrument,
// committed per red R3-1 (round 3): the round-2 measurement ran from a session-scratchpad
// script that was not retained; this is the same documented method, re-implemented and
// re-run against the pinned tarball so the figure's derivation is reproducible from the
// run directory alone.
//
// Method (as documented in blue/report.md [^MergeDecomposition]):
//   1. extract <runDir>/trajectories/agent-transcripts.tar.gz (RETENTION ASSUMPTION:
//      the tarball is gitignored — present in the producing run's working tree, NOT in
//      the git object store; this script is reproducible only where the tarball survives)
//   2. identify red-merge transcripts by cost-audit.mjs's own "Red merge, round N" head match
//   3. attribute tool_result bytes to source files via the paired tool_use's
//      input.file_path (Read/Edit/Write) or input.command (Bash)
//   4. cache-weighting = bytes x remaining assistant turns after the result lands
//      (approximates prefix re-billing at cache-read rate; ignores system prompt and
//      harness overhead)
//   5. dollars = cache-weighted bytes / 4 (bytes->tokens) / 1e6 x cache-read $/MTok
//
// Usage: node decompose-merge.mjs <extracted-transcript-dir> [cacheReadPerMTok=1.00]

import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const dir = process.argv[2]
if (!dir) { console.error('usage: node decompose-merge.mjs <extracted-transcript-dir> [cacheReadPerMTok]'); process.exit(1) }
const RATE = Number(process.argv[3] || 1.0) // fable cache-read $/MTok per cost-audit.mjs PRICES

const BUCKETS = [
  ['blue/report.md', (s) => s.includes('blue/report.md')],
  ['red/findings.md', (s) => s.includes('red/findings.md') || /(^|[\/\s])findings\.md/.test(s)],
  ['lens candidates', (s) => s.includes('red/candidates/')],
  ['debate.md', (s) => s.includes('debate.md')],
]
const classify = (src) => (BUCKETS.find(([, f]) => f(src.replace(/\\+/g, '/'))) || ['other'])[0]

const flat = (c) => typeof c === 'string' ? c
  : Array.isArray(c) ? c.map((b) => b.text || (typeof b === 'string' ? b : '')).join('') : ''

for (const f of readdirSync(dir).filter((f) => f.startsWith('agent-') && f.endsWith('.jsonl')).sort()) {
  const txt = readFileSync(join(dir, f), 'utf8')
  const m = txt.slice(0, 2000).match(/Red merge, round (\d+)/)
  if (!m) continue
  const round = +m[1]
  const lines = txt.split('\n').filter((l) => l.trim()).map((l) => { try { return JSON.parse(l) } catch { return null } }).filter(Boolean)
  // pass 1: tool_use id -> source string; count assistant turns; index of each line's preceding assistant-turn count
  const src = {}
  let turns = 0
  const events = [] // { id, bytes, turnIndex }
  for (const j of lines) {
    const msg = j.message
    if (!msg || !msg.content) continue
    if (msg.role === 'assistant') {
      if (msg.usage) turns++
      for (const b of Array.isArray(msg.content) ? msg.content : []) {
        if (b.type === 'tool_use') src[b.id] = String(b.input?.file_path || b.input?.command || b.input?.pattern || b.name || '')
      }
    } else if (msg.role === 'user') {
      for (const b of Array.isArray(msg.content) ? msg.content : []) {
        if (b.type === 'tool_result') events.push({ id: b.tool_use_id, bytes: flat(b.content).length, turnIndex: turns })
      }
    }
  }
  const raw = {}, weighted = {}
  for (const e of events) {
    const bucket = classify(src[e.id] || '')
    raw[bucket] = (raw[bucket] || 0) + e.bytes
    weighted[bucket] = (weighted[bucket] || 0) + e.bytes * Math.max(0, turns - e.turnIndex)
  }
  const total = Object.values(raw).reduce((a, b) => a + b, 0)
  const KB = (n) => (n / 1024).toFixed(1) + 'KB'
  console.log(`\nred-merge round ${round}: ${events.length} tool results, ${turns} assistant turns, total ingest ${KB(total)}`)
  for (const k of ['blue/report.md', 'red/findings.md', 'lens candidates', 'debate.md', 'other']) {
    const r = raw[k] || 0, w = weighted[k] || 0
    console.log(`  ${k.padEnd(18)} raw ${KB(r).padStart(9)} (${((100 * r) / total).toFixed(1).padStart(5)}%)  cache-weighted $${((w / 4 / 1e6) * RATE).toFixed(2)}`)
  }
}

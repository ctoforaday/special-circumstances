#!/usr/bin/env node
// Standard instrument (promoted from run-4 trajectories/, where it was seat-improvised —
// run-4 friction: "run 5 should get this as standard cost-audit output, not seat
// improvisation"). Measures the read-batching saving: what the merge seat paid to ingest
// lens candidates turn-by-turn, i.e. what the batching sentence collapses.
// Usage: node batch-collapse.mjs <workflow-transcript-dir>
// R4-4 turn-collapse measurement, tool_result-paired (same pairing as decompose-merge.mjs):
// a candidate-read tool_result's ingesting call is the next assistant turn with usage;
// prompt-level read batching collapses k candidate-read ingestions into ~1, avoiding the
// billed input of k-1 of those calls. Rates: fable session tier per cost-audit.mjs PRICES
// (cache-read 1.0 $/MTok default, matching decompose-merge.mjs; cache-write 6.25; input 5.0).
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
const dir = process.argv[2]
const RATE_READ = 1.0, RATE_WRITE = 6.25, RATE_IN = 5.0
const norm = (s) => s.replace(/\\+/g, '/')
const out = []
for (const f of readdirSync(dir).filter(f => f.startsWith('agent-') && f.endsWith('.jsonl')).sort()) {
  const txt = readFileSync(join(dir, f), 'utf8')
  const m = txt.slice(0, 2000).match(/Red merge, round (\d+)/)
  if (!m) continue
  const lines = txt.split('\n').filter(l => l.trim()).map(l => { try { return JSON.parse(l) } catch { return null } }).filter(Boolean)
  const candIds = new Set()
  for (const j of lines) {
    const msg = j.message; if (!msg || msg.role !== 'assistant' || !Array.isArray(msg.content)) continue
    for (const b of msg.content) if (b.type === 'tool_use' &&
      norm(String(b.input?.file_path || b.input?.command || '')).includes('red/candidates/')) candIds.add(b.id)
  }
  let pendingCand = false; const ingest = []
  for (const j of lines) {
    const msg = j.message; if (!msg || !Array.isArray(msg.content)) continue
    if (msg.role === 'user') {
      for (const b of msg.content) if (b.type === 'tool_result' && candIds.has(b.tool_use_id)) pendingCand = true
    } else if (msg.role === 'assistant' && msg.usage && ((msg.usage.cache_read_input_tokens || 0) + (msg.usage.input_tokens || 0) + (msg.usage.cache_creation_input_tokens || 0) > 0)) {
      if (pendingCand) { ingest.push({ cr: msg.usage.cache_read_input_tokens || 0, cw: msg.usage.cache_creation_input_tokens || 0, inp: msg.usage.input_tokens || 0 }); pendingCand = false }
    }
  }
  const avoided = ingest.slice(1)
  const d = avoided.reduce((a, t) => a + (t.cr * RATE_READ + t.cw * RATE_WRITE + t.inp * RATE_IN) / 1e6, 0)
  out.push([+m[1], ingest.length, avoided.length, d, ingest.map(t => t.cr).join('/')])
}
for (const [r, k, a, d, crs] of out.sort((x, y) => x[0] - y[0]))
  console.log(`round ${r}: ${k} candidate ingestion calls; avoided ${a}; billed input avoided ~= $${d.toFixed(2)} (cache_read at those calls: ${crs})`)
console.log(`sum avoided across rounds: $${out.reduce((a, x) => a + x[3], 0).toFixed(2)}`)

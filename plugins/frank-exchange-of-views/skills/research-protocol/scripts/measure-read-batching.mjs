#!/usr/bin/env node
// Standard instrument (promoted from run-4 trajectories/, where it was seat-improvised —
// run-4 friction: "run 5 should get this as standard cost-audit output, not seat
// improvisation"). Measures the read-batching saving: what the merge seat paid to ingest
// lens candidates turn-by-turn, i.e. what the batching sentence collapses.
// Usage: node measure-read-batching.mjs <workflow-transcript-dir>
// R4-4 turn-collapse measurement, tool_result-paired (same pairing as decompose-merge.mjs):
// a candidate-read tool_result's ingesting call is the next assistant turn with usage;
// prompt-level read batching collapses k candidate-read ingestions into ~1, avoiding the
// billed input of k-1 of those calls. Rates: fable session tier per cost-audit.mjs PRICES
// (cache-read 1.0 $/MTok default, matching decompose-merge.mjs; cache-write 6.25; input 5.0).
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { pathToFileURL } from 'node:url'
export const RATE_READ = 1.0, RATE_WRITE = 6.25, RATE_IN = 5.0
const norm = (s) => s.replace(/\\+/g, '/')

// measureTranscript is the whole measurement for ONE agent transcript. Returns
// null for any transcript that is not a red-merge seat — the saving being
// measured only exists where candidates are ingested.
//
// The pairing rule, stated so a future reader can check it: a candidate-read
// tool_result is attributed to the NEXT assistant turn that carries billed
// input, and that turn is one "ingestion call". Several candidate results
// arriving before a single assistant turn are ONE ingestion — that collapse is
// exactly what prompt-level batching buys, so counting them separately would
// measure the saving as zero. The first ingestion is unavoidable; only the
// remaining k-1 are counted as avoided.
export function measureTranscript(txt) {
  const m = txt.slice(0, 2000).match(/Red merge, round (\d+)/)
  if (!m) return null
  // Malformed lines are skipped, never fatal: a run killed mid-append leaves a
  // half-written final JSONL line, and those are the runs worth measuring.
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
  const dollars = avoided.reduce((a, t) => a + (t.cr * RATE_READ + t.cw * RATE_WRITE + t.inp * RATE_IN) / 1e6, 0)
  return { round: +m[1], ingest, avoided: avoided.length, dollars, cacheReads: ingest.map(t => t.cr) }
}

export const formatRow = (r) =>
  `round ${r.round}: ${r.ingest.length} candidate ingestion calls; avoided ${r.avoided}; ` +
  `billed input avoided ~= $${r.dollars.toFixed(2)} (cache_read at those calls: ${r.cacheReads.join('/')})`

function main() {
  const dir = process.argv[2]
  // Matches cost-audit.mjs's contract rather than dying inside readdirSync(undefined)
  // with a stack trace: the two scripts take the same argument and an operator who
  // mistypes one should get the same answer from both.
  if (!dir) { console.error('usage: node measure-read-batching.mjs <workflow-transcript-dir>'); process.exit(1) }
  const out = []
  for (const f of readdirSync(dir).filter(f => f.startsWith('agent-') && f.endsWith('.jsonl')).sort()) {
    const r = measureTranscript(readFileSync(join(dir, f), 'utf8'))
    if (r) out.push(r)
  }
  for (const r of out.sort((x, y) => x.round - y.round)) console.log(formatRow(r))
  console.log(`sum avoided across rounds: $${out.reduce((a, r) => a + r.dollars, 0).toFixed(2)}`)
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main()

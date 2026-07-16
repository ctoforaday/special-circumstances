#!/usr/bin/env node
// Cost audit for a debate run — parses the Workflow harness's per-agent transcripts and
// emits a markdown report (per seat-round tokens + dollars, totals, findings) for the run
// record. Part of the run's standard output, next to friction.md (retrospective backlog:
// "a tool, not a diet" — measurement first, never a silent judgment discount).
//
// Usage: node cost-audit.mjs <workflow-transcript-dir> [<runDir>] > <runDir>/cost.md
// The transcript dir is printed by the Workflow tool at launch ("Transcript dir: ...").
// With <runDir>, joins the per-round board telemetry (trajectories/board-telemetry.jsonl)
// into the report — the named-consumer extension from run-4 report §2.5 item 1.
//
// Pricing is LIST-RATE arithmetic ($/MTok below) — plan meters typically observe less.
// Run 3 calibration: the meter drew ~0.6x of these figures.
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const PRICES = {
  //         input  output cache-read cache-write
  haiku:  [   1,     5,    0.10,      1.25 ],
  sonnet: [   2,    10,    0.20,      2.50 ], // intro pricing through 2026-08-31; list is 3/15/0.3/3.75
  opus:   [   5,    25,    0.50,      6.25 ],
  fable:  [  10,    50,    1.00,     12.50 ],
}
const tier = (m) => Object.keys(PRICES).find((k) => (m || '').toLowerCase().includes(k)) || 'fable'

const wf = process.argv[2]
if (!wf) { console.error('usage: node cost-audit.mjs <workflow-transcript-dir>'); process.exit(1) }

const rows = []
for (const f of readdirSync(wf).filter((f) => f.startsWith('agent-') && f.endsWith('.jsonl'))) {
  const txt = readFileSync(join(wf, f), 'utf8')
  const head = txt.slice(0, 2000)
  let seat = 'other', round = 0, m
  if ((m = head.match(/Red audit, round (\d+)/))) { seat = 'red-lens'; round = +m[1] }
  else if ((m = head.match(/Red merge, round (\d+)/))) { seat = 'red-merge'; round = +m[1] }
  else if ((m = head.match(/Blue response, round (\d+)/))) { seat = 'blue-respond'; round = +m[1] }
  else if ((m = head.match(/Adjudication, round (\d+)/))) { seat = 'judge'; round = +m[1] }
  else if (head.includes('Blue synthesis')) seat = 'blue-synthesize'
  else if (head.includes('Blue lane')) seat = 'blue-lane'
  else if (head.includes('frontier hypotheses')) seat = 'frontier'
  else if (head.includes('Final assembly')) seat = 'assemble'
  let model = null, inp = 0, out = 0, cr = 0, cw = 0, turns = 0
  for (const line of txt.split('\n')) {
    if (!line.trim()) continue
    let j; try { j = JSON.parse(line) } catch { continue }
    const u = j.message && j.message.usage
    if (u) {
      model = j.message.model || model; turns++
      inp += u.input_tokens || 0; out += u.output_tokens || 0
      cr += u.cache_read_input_tokens || 0; cw += u.cache_creation_input_tokens || 0
    }
  }
  const t = tier(model), p = PRICES[t]
  rows.push({ seat, round, t, turns, inp, out, cr, cw, cost: (inp * p[0] + out * p[1] + cr * p[2] + cw * p[3]) / 1e6 })
}

const M = (n) => (n / 1e6).toFixed(2) + 'M'
const agg = {}
for (const r of rows) {
  const k = `${String(r.round).padStart(2, '0')}|${r.seat}|${r.t}`
  agg[k] = agg[k] || { n: 0, turns: 0, inp: 0, out: 0, cr: 0, cw: 0, cost: 0 }
  const a = agg[k]
  a.n++; a.turns += r.turns; a.inp += r.inp; a.out += r.out; a.cr += r.cr; a.cw += r.cw; a.cost += r.cost
}

console.log('# Cost audit\n')
console.log(`Measured from ${rows.length} per-agent API transcripts in \`${wf}\`. List-rate arithmetic; see the price table in cost-audit.mjs.\n`)
console.log('## Per seat-round\n')
console.log('| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |')
console.log('|---|---|---|---|---|---|---|---|---|---|')
const T = { n: 0, turns: 0, inp: 0, out: 0, cr: 0, cw: 0, cost: 0 }
for (const k of Object.keys(agg).sort()) {
  const [round, seat, t] = k.split('|'); const a = agg[k]
  console.log(`| ${+round === 0 ? '—' : +round} | ${seat} | ${t} | ${a.n} | ${a.turns} | ${M(a.inp)} | ${M(a.out)} | ${M(a.cr)} | ${M(a.cw)} | $${a.cost.toFixed(2)} |`)
  T.n += a.n; T.turns += a.turns; T.inp += a.inp; T.out += a.out; T.cr += a.cr; T.cw += a.cw; T.cost += a.cost
}
console.log(`| | **TOTAL** | | ${T.n} | ${T.turns} | ${M(T.inp)} | ${M(T.out)} | ${M(T.cr)} | ${M(T.cw)} | **$${T.cost.toFixed(2)}** |`)
const cachePct = Math.round((100 * (T.cr + T.cw)) / (T.inp + T.out + T.cr + T.cw || 1))
console.log('\n## Notes\n')
console.log(`- Cache traffic is ${cachePct}% of all tokens; harness panel counters (input+output only) understate real flow accordingly.`)
// Findings text corrected per run-4 report §6.4 item 1: (a) merge cost tracks the CUMULATIVE
// ARCHIVE, not dispute size — run 3's own table contradicted the old claim (r5 merge > r2 on
// the second-smallest board); sharding (ledger/archive split) is the countermeasure. (b) The
// judgment-tier cache-write premium is a ~5x RATE multiplier over sonnet-intro (12.5 is the
// absolute $/MTok, not the multiplier).
console.log('- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.')

// Board-telemetry join (optional second arg): the per-round mass/board series next to the
// per-round dollars — the evidence base every deferred actuation decision needs.
const runDir = process.argv[3]
if (runDir) {
  try {
    const lines = readFileSync(join(runDir, 'trajectories', 'board-telemetry.jsonl'), 'utf8')
      .split('\n').filter(Boolean).map((l) => { try { return JSON.parse(l) } catch { return null } }).filter(Boolean)
    if (lines.length) {
      console.log('\n## Board telemetry (per round)\n')
      console.log('| round | open | max severity | new mints | mass | realized_open | accepted deltas | mapping |')
      console.log('|---|---|---|---|---|---|---|---|')
      for (const t of lines) {
        console.log(`| ${t.round} | ${t.open_count ?? '?'} | ${t.max_severity ?? '?'} | ${(t.new_mint && t.new_mint.count) ?? '?'} | ${t.mass ?? '?'} | ${t.realized_open ?? '?'} | ${(t.accepted_deltas || []).length} | ${t.mapping_version ?? '?'} |`)
      }
      console.log('\nTelemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.')
    } else {
      console.log('\n## Board telemetry\n\n(board-telemetry.jsonl present but empty)')
    }
  } catch {
    console.log('\n## Board telemetry\n\n(no board-telemetry.jsonl in the run dir — pre-telemetry run, or the merge seat never appended: check capture-audit.md)')
  }
}

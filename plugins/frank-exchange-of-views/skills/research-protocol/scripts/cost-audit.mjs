#!/usr/bin/env node
// Cost audit for a debate run — parses the Workflow harness's per-agent transcripts and
// emits a markdown report (per seat-round tokens + dollars, totals, findings) for the run
// record. Part of the run's standard output, alongside the record's friction (the friction
// verb; render projects it to render-shadow/friction.md) (retrospective backlog:
// "a tool, not a diet" — measurement first, never a silent judgment discount).
//
// Usage: node cost-audit.mjs <workflow-transcript-dir> [<runDir>] > <runDir>/cost.md
// The transcript dir is printed by the Workflow tool at launch ("Transcript dir: ...").
// With <runDir>, joins the per-round board telemetry (trajectories/board-telemetry.jsonl)
// into the report — the named-consumer extension from run-4 report §2.5 item 1.
//
// Pricing is LIST-RATE arithmetic ($/MTok below) — plan meters typically observe less.
// Run 3 calibration: the meter drew ~0.6x of these figures.
import { classifySeat as classifyTranscript } from './seat-classify.mjs'
import { tierMismatch } from './model-guard.mjs'
import { existsSync, readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { pathToFileURL } from 'node:url'

export const PRICES = {
  //         input  output cache-read cache-write
  haiku:  [   1,     5,    0.10,      1.25 ],
  sonnet: [   2,    10,    0.20,      2.50 ], // intro pricing through 2026-08-31; list is 3/15/0.3/3.75
  opus:   [   5,    25,    0.50,      6.25 ],
  fable:  [  10,    50,    1.00,     12.50 ],
}
// tier picks the price row by substring. An UNRECOGNIZED model falls back to
// `fable`, the most expensive row: an unknown model must over-report, never
// under-report, since a silently cheap estimate is the failure that goes
// unnoticed.
export const tier = (m) => Object.keys(PRICES).find((k) => (m || '').toLowerCase().includes(k)) || 'fable'

// Seat identity is recovered from the transcript's first user message — the
// harness journal carries no seat labels. Round-bearing prompts are matched
// first because a round-0 marker can also appear in a round-bearing prompt's
// preamble.
// classifyTranscript is the shared seat table. The private copy it replaced was
// missing the `Terminal dispute disposition` case, so terminal-disposition
// transcripts were filed as `other` and their spend was misattributed in the
// cost report. Adding the case CHANGES this report's output — deliberately:
// the previous output was wrong.
export { classifySeat as classifyTranscript } from './seat-classify.mjs'

// scanTranscript sums one agent's usage records and prices them. Non-JSON lines
// are skipped rather than fatal: a run killed mid-append leaves a half-written
// final line, and a cost report that throws on it reports nothing at all.
export function scanTranscript(txt) {
  const { seat, round } = classifyTranscript(txt.slice(0, 2000))
  let model = null, inp = 0, out = 0, cr = 0, cw = 0, turns = 0
  for (const line of txt.split('\n')) {
    if (!line.trim()) continue
    let j; try { j = JSON.parse(line) } catch { continue }
    const u = j.message && j.message.usage
    if (u) {
      // A recognized tier always wins; an unrecognized string (e.g. a "<synthetic>" harness turn)
      // only fills a still-empty slot. One synthetic turn must not reprice a whole seat to the
      // fallback tier — which mislabels cost AND fired a false #111 tier-guard FAIL. Recognized =
      // matches a PRICES key (the one tier source), so this adds no second copy of the tier list.
      const mm = j.message.model
      if (mm && (!model || Object.keys(PRICES).some((k) => mm.toLowerCase().includes(k)))) model = mm
      turns++
      inp += u.input_tokens || 0; out += u.output_tokens || 0
      cr += u.cache_read_input_tokens || 0; cw += u.cache_creation_input_tokens || 0
    }
  }
  const t = tier(model), p = PRICES[t]
  return { seat, round, t, turns, inp, out, cr, cw, cost: (inp * p[0] + out * p[1] + cr * p[2] + cw * p[3]) / 1e6 }
}

// aggregate groups rows into seat-round-tier buckets. The key is zero-padded so
// plain lexicographic sort orders rounds numerically (round 2 before round 10).
export function aggregate(rows) {
  const agg = {}
  for (const r of rows) {
    const k = `${String(r.round).padStart(2, '0')}|${r.seat}|${r.t}`
    agg[k] = agg[k] || { n: 0, turns: 0, inp: 0, out: 0, cr: 0, cw: 0, cost: 0 }
    const a = agg[k]
    a.n++; a.turns += r.turns; a.inp += r.inp; a.out += r.out; a.cr += r.cr; a.cw += r.cw; a.cost += r.cost
  }
  return agg
}

// Share of all tokens that is cache traffic. The `|| 1` guards an empty run:
// division by zero would print NaN% into a run record.
export const cacheShare = (T) => Math.round((100 * (T.cr + T.cw)) / (T.inp + T.out + T.cr + T.cw || 1))

// The report is emitted from main(), which runs only when this file IS the
// entry point — the pure functions above are importable without a CLI run
// firing (and calling process.exit) as a side effect of the import.
function main() {
const wf = process.argv[2]
if (!wf) { console.error('usage: node cost-audit.mjs <workflow-transcript-dir>'); process.exit(1) }

const rows = []
for (const f of readdirSync(wf).filter((f) => f.startsWith('agent-') && f.endsWith('.jsonl'))) {
  rows.push(scanTranscript(readFileSync(join(wf, f), 'utf8')))
}

const M = (n) => (n / 1e6).toFixed(2) + 'M'
const agg = aggregate(rows)

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
const cachePct = cacheShare(T)
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
  // ## Tier check (#111): actual price-tier per seat vs the tier its class was configured to run
  // on. DEARER than configured = the fable trap (FAIL); CHEAPER = discounted verification (WARN).
  // Config is inputs/run-config.json; debate.js now requires both tiers, so a current run has them.
  try {
    let cfg = {}
    try { cfg = JSON.parse(readFileSync(join(runDir, 'inputs', 'run-config.json'), 'utf8')) } catch {}
    // rows carry actual tier per transcript; several transcripts can share a seat-round, so dedupe.
    const seen = new Set()
    const findings = tierMismatch(rows, cfg).filter((f) => {
      const k = `${f.seat}|${f.round}|${f.verdict}|${f.actual}`
      if (seen.has(k)) return false
      seen.add(k); return true
    })
    console.log('\n## Tier check\n')
    if (!findings.length) console.log(`- PASS — every seat ran on its configured tier (bulk: ${cfg.model || 'undeclared'}, judgment: ${cfg.judgmentModel || 'undeclared'}).`)
    else for (const f of findings) console.log(`- **${f.verdict}** — ${f.why}`)
  } catch { /* run-config unreadable — tier check silently skipped, telemetry below still runs */ }
  try {
    const telPath = [join(runDir, 'records', 'render-shadow', 'board-telemetry.jsonl'), join(runDir, 'trajectories', 'board-telemetry.jsonl')].find(existsSync)
    const lines = (telPath ? readFileSync(telPath, 'utf8') : '')
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
      console.log(telPath
        ? '\n## Board telemetry\n\n(board-telemetry.jsonl present but empty)'
        : '\n## Board telemetry\n\n(no board-telemetry.jsonl in the run dir — pre-telemetry run, or the merge seat never rendered: check run-record-audit.md)')
    }
  } catch {
    console.log('\n## Board telemetry\n\n(no board-telemetry.jsonl in the run dir — pre-telemetry run, or the merge seat never appended: check run-record-audit.md)')
  }
}
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main()

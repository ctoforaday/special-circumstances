#!/usr/bin/env node
// Live run dashboard for a debate — renders the run's own instruments (board telemetry,
// journal lifecycle, blackboard sizes, transcript-derived cost) into a single local
// dashboard.html. The Workflow panel shows a phase name and the last-touched agent label;
// this shows what is actually happening.
//
// Usage:
//   node render-run-dashboard.mjs <runDir> <workflow-transcript-dir> [--watch]
// Writes <runDir>/dashboard.html (open it in a browser; it meta-refreshes every 20s).
// --watch regenerates every 15s until Ctrl-C. Read-only over the run; writes only the html.
import { readFileSync, writeFileSync, readdirSync, existsSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { pathToFileURL } from 'node:url'

const jsonl = (p) => existsSync(p)
  ? readFileSync(p, 'utf8').split('\n').filter(Boolean).map((l) => { try { return JSON.parse(l) } catch { return null } }).filter(Boolean)
  : []

// Same list-rate arithmetic as cost-audit.mjs (kept tiny here: fable-tier default).
const RATE = { in: 10, out: 50, cr: 1.0, cw: 12.5 }

export function buildModel(runDir, transcriptDir) {
  const telemetry = jsonl(join(runDir, 'trajectories', 'board-telemetry.jsonl'))
  const journal = jsonl(join(transcriptDir, 'journal.jsonl'))

  // Seat lifecycle: started without a matching result = live.
  const seats = new Map()
  for (const j of journal) {
    const label = j.label || (j.result !== undefined ? '(unlabeled)' : null)
    if (!label) continue
    const s = seats.get(label) || { label, started: null, done: false }
    if (j.type === 'started' || (j.result === undefined && !s.started)) s.started = j.timestamp || s.started
    if (j.result !== undefined) { s.done = true; s.result = typeof j.result === 'string' ? j.result.slice(0, 100) : JSON.stringify(j.result).slice(0, 100) }
    seats.set(label, s)
  }

  // Cost so far from transcripts (usage records).
  let cost = 0, rounds = 0, agents = 0
  if (existsSync(transcriptDir)) {
    for (const f of readdirSync(transcriptDir).filter((f) => f.startsWith('agent-') && f.endsWith('.jsonl'))) {
      agents++
      for (const j of jsonl(join(transcriptDir, f))) {
        const u = j.message && j.message.usage
        if (!u) continue
        rounds++
        cost += ((u.input_tokens || 0) * RATE.in + (u.output_tokens || 0) * RATE.out +
          (u.cache_read_input_tokens || 0) * RATE.cr + (u.cache_creation_input_tokens || 0) * RATE.cw) / 1e6
      }
    }
  }

  // CONTENTS over bytes (review feedback): counts and states, not file sizes.
  const readIf = (rel) => { const p = join(runDir, rel); return existsSync(p) ? readFileSync(p, 'utf8') : null }
  const frictionTxt = readIf('friction.md')
  const friction = {
    count: frictionTxt ? (frictionTxt.match(/^- /gm) || []).length : 0,
    last: frictionTxt ? (frictionTxt.match(/^- .*$/gm) || []).slice(-1)[0] || null : null,
  }
  const ledgerTxt = readIf('red/ledger.md')
  const archiveTxt = readIf('red/archive.md')
  const GRADES = ['certain', 'high', 'medium-high', 'medium', 'low-medium', 'low', 'trivial', 'realized']
  const idLine = /R\d+-\d+/
  let openRows = 0
  const openBySeverity = {}
  if (ledgerTxt) {
    // Heuristic over red's ledger: rows above the closure index are the open board.
    const closureAt = ledgerTxt.search(/closure index/i)
    const openSection = closureAt >= 0 ? ledgerTxt.slice(0, closureAt) : ledgerTxt
    for (const line of openSection.split('\n')) {
      if (!idLine.test(line) || !line.includes('|')) continue
      openRows++
      const low = line.toLowerCase()
      const g = GRADES.find((g2) => low.includes(g2))
      if (g) openBySeverity[g] = (openBySeverity[g] || 0) + 1
    }
  }
  const shards = {
    ledgerExists: ledgerTxt !== null,
    openRows,
    openBySeverity,
    closureIndexRows: ledgerTxt ? (ledgerTxt.slice(Math.max(0, ledgerTxt.search(/closure index/i))).match(/R\d+-\d+/g) || []).length : 0,
    archiveRecords: archiveTxt ? (archiveTxt.match(/^#{1,4}\s+.*R\d+-\d+/gm) || []).length : 0,
  }
  // Blue corpus size in CLAIMS (last blue envelope in the journal), not bytes.
  let blueClaims = null
  for (const j of journal) if (j.result && typeof j.result === 'object' && typeof j.result.claim_count === 'number') blueClaims = j.result.claim_count

  // Progress through the workflow's big steps: Frontier -> Blue lanes -> Synthesis ->
  // rounds 1..maxRounds (each: lenses -> merge -> respond [-> judge]) -> Assembly.
  // The ceiling divides the bar; judged termination may end it early — stated on the bar.
  const MAX_ROUNDS_CEILING = 8
  const seatList = [...seats.values()]
  const seen = (prefix) => seatList.some((s) => s.label.startsWith(prefix))
  const doneSeat = (prefix) => seatList.some((s) => s.label.startsWith(prefix) && s.done)
  const steps = []
  steps.push({ name: 'frontier', state: doneSeat('frontier') ? 'done' : seen('frontier') ? 'live' : 'todo' })
  steps.push({ name: 'blue lanes', state: seatList.filter((s) => s.label.startsWith('blue-lane')).every((s) => s.done) && seen('blue-lane') ? 'done' : seen('blue-lane') ? 'live' : 'todo' })
  steps.push({ name: 'synthesis', state: doneSeat('blue-synthesize') ? 'done' : seen('blue-synthesize') ? 'live' : 'todo' })
  for (let r = 1; r <= MAX_ROUNDS_CEILING; r++) {
    const respondDone = doneSeat(`blue-respond-r${r}`)
    const anySeen = seen(`red-lens-1-r${r}`) || seen(`red-merge-r${r}`) || seen(`blue-respond-r${r}`) || seen(`judge-r${r}`)
    steps.push({ name: `round ${r}`, state: respondDone ? 'done' : anySeen ? 'live' : 'todo' })
  }
  steps.push({ name: 'assembly', state: doneSeat('assemble') ? 'done' : seen('assemble') ? 'live' : 'todo' })

  // Open/close rates per round, computed from consecutive telemetry lines:
  // closed_in_round = prev_open + minted - open (round 1: minted - open).
  const rates = telemetry.map((t, i) => {
    const prevOpen = i === 0 ? 0 : (telemetry[i - 1].open_count || 0)
    const minted = (t.new_mint && t.new_mint.count) || 0
    const closed = prevOpen + minted - (t.open_count || 0)
    return { round: t.round, opened: minted, closed, open: t.open_count, closeRate: minted + prevOpen > 0 ? Math.round((100 * closed) / (prevOpen + minted)) : 0 }
  })

  const latest = telemetry[telemetry.length - 1] || null
  return { runDir, telemetry, latest, seats: seatList, cost, apiRounds: rounds, agents, friction, shards, blueClaims, steps, rates, generated: new Date().toISOString() }
}

const esc = (s) => String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

// Single-series mass trend as inline SVG: 2px line, >=8px markers, direct label on the
// last point, no legend (one series — the title names it). Table below is the data view.
function massSvg(telemetry) {
  if (telemetry.length === 0) return '<p class="muted">no telemetry yet — red-merge appends the first line at round 1</p>'
  const W = 560, H = 160, pad = 34
  const xs = telemetry.map((t, i) => pad + (telemetry.length === 1 ? (W - 2 * pad) / 2 : (i * (W - 2 * pad)) / (telemetry.length - 1)))
  const max = Math.max(...telemetry.map((t) => t.mass || 0), 1)
  const ys = telemetry.map((t) => H - pad - ((t.mass || 0) / max) * (H - 2 * pad))
  const pts = xs.map((x, i) => `${x.toFixed(1)},${ys[i].toFixed(1)}`).join(' ')
  const dots = telemetry.map((t, i) =>
    `<circle cx="${xs[i].toFixed(1)}" cy="${ys[i].toFixed(1)}" r="4" fill="var(--series-1)"><title>round ${esc(t.round)}: mass ${esc(t.mass)}, open ${esc(t.open_count)}</title></circle>` +
    `<text x="${xs[i].toFixed(1)}" y="${H - 10}" text-anchor="middle" class="axis">r${esc(t.round)}</text>`).join('')
  const last = telemetry[telemetry.length - 1]
  return `<svg viewBox="0 0 ${W} ${H}" role="img" aria-label="board mass by round">
<polyline points="${pts}" fill="none" stroke="var(--series-1)" stroke-width="2"/>${dots}
<text x="${(xs[xs.length - 1] - 8).toFixed(1)}" y="${(ys[ys.length - 1] - 10).toFixed(1)}" text-anchor="end" class="label">${esc(last.mass)}</text>
</svg>`
}

export function renderHtml(m) {
  const sevRow = (t) => t && t.new_mint && t.new_mint.by_severity
    ? Object.entries(t.new_mint.by_severity).map(([k, v]) => `${esc(k)}:${esc(v)}`).join(' · ') : '—'
  const live = m.seats.filter((s) => !s.done)
  const done = m.seats.filter((s) => s.done)
  return `<!-- generated by render-run-dashboard.mjs -->
<meta charset="utf-8"><meta http-equiv="refresh" content="20">
<title>FEOV run — ${esc(m.runDir.split(/[\\/]/).pop())}</title>
<style>
.viz-root { color-scheme: light; --surface-1:#fcfcfb; --text-primary:#0b0b0b; --text-secondary:#52514e; --series-1:#2a78d6;
  font: 14px/1.45 system-ui, sans-serif; background: var(--surface-1); color: var(--text-primary); padding: 20px; max-width: 900px; margin: auto; }
@media (prefers-color-scheme: dark) { :root:where(:not([data-theme="light"])) .viz-root { color-scheme: dark; --surface-1:#1a1a19; --text-primary:#ffffff; --text-secondary:#c3c2b7; --series-1:#3987e5; } }
:root[data-theme="dark"] .viz-root { color-scheme: dark; --surface-1:#1a1a19; --text-primary:#ffffff; --text-secondary:#c3c2b7; --series-1:#3987e5; }
h1 { font-size: 18px; } h2 { font-size: 14px; margin: 18px 0 6px; color: var(--text-secondary); }
.tiles { display: flex; gap: 12px; flex-wrap: wrap; } .tile { border: 1px solid color-mix(in oklab, var(--text-secondary) 30%, transparent); border-radius: 8px; padding: 10px 14px; min-width: 96px; }
.tile b { display: block; font-size: 22px; } .tile span { color: var(--text-secondary); font-size: 12px; }
table { border-collapse: collapse; width: 100%; } td, th { text-align: left; padding: 3px 10px 3px 0; border-bottom: 1px solid color-mix(in oklab, var(--text-secondary) 20%, transparent); }
.muted, .axis, .label { fill: var(--text-secondary); color: var(--text-secondary); font-size: 11px; }
.label { fill: var(--text-primary); font-weight: 600; }
.liveseat { font-weight: 600; }
.bar { display: flex; gap: 3px; margin: 6px 0 2px; } .seg { flex: 1; height: 14px; border-radius: 4px; background: color-mix(in oklab, var(--text-secondary) 18%, transparent); position: relative; }
.seg.done { background: var(--series-1); } .seg.live { background: color-mix(in oklab, var(--series-1) 45%, transparent); outline: 2px solid var(--series-1); }
.barlabels { display: flex; gap: 3px; } .barlabels span { flex: 1; font-size: 10px; color: var(--text-secondary); text-align: center; overflow: hidden; white-space: nowrap; }
</style>
<body class="viz-root">
<h1>FEOV run · ${esc(m.runDir.split(/[\\/]/).pop())}</h1>
<p class="muted">generated ${esc(m.generated)} · auto-refreshes every 20s · dollars are list-rate estimates</p>
<h2>Progress (rounds segmented by the ceiling — judged termination may end the run earlier)</h2>
<div class="bar">${m.steps.map((s) => `<div class="seg ${s.state}" title="${esc(s.name)}: ${esc(s.state)}"></div>`).join('')}</div>
<div class="barlabels">${m.steps.map((s) => `<span>${esc(s.name.replace('round ', 'r'))}</span>`).join('')}</div>
<div class="tiles" style="margin-top:12px">
<div class="tile"><b>${m.latest ? esc(m.latest.mass) : '—'}</b><span>board mass</span></div>
<div class="tile"><b>${m.latest ? esc(m.latest.open_count) : '—'}</b><span>open gaps</span></div>
<div class="tile"><b>${m.latest ? esc(m.latest.max_severity) : '—'}</b><span>max severity</span></div>
<div class="tile"><b>${m.blueClaims ?? '—'}</b><span>blue claims</span></div>
<div class="tile"><b>${m.friction.count}</b><span>friction entries</span></div>
<div class="tile"><b>$${m.cost.toFixed(2)}</b><span>cost so far (est.)</span></div>
<div class="tile"><b>${done.length}/${m.agents}</b><span>seats done</span></div>
</div>
<h2>Board mass by round</h2>
${massSvg(m.telemetry)}
<h2>Open / close rates (the convergence signal — is red's discovery decaying?)</h2>
<table><tr><th>round</th><th>opened</th><th>closed</th><th>still open</th><th>close rate</th><th>max sev</th><th>new mints by severity</th><th>mass</th><th>deltas</th></tr>
${m.rates.map((r, i) => { const t = m.telemetry[i]; return `<tr><td>${esc(r.round)}</td><td>${esc(r.opened)}</td><td>${esc(r.closed)}</td><td>${esc(r.open)}</td><td>${esc(r.closeRate)}%</td><td>${esc(t.max_severity ?? '—')}</td><td>${sevRow(t)}</td><td>${esc(t.mass ?? '—')}</td><td>${(t.accepted_deltas || []).length}</td></tr>` }).join('\n')}
</table>
<h2>Red's board (contents, not bytes)</h2>
${m.shards.ledgerExists ? `<table>
<tr><td>open gaps on the ledger</td><td>${m.shards.openRows}${Object.keys(m.shards.openBySeverity).length ? ' <span class="muted">(' + Object.entries(m.shards.openBySeverity).map(([k, v]) => `${esc(k)}:${esc(v)}`).join(' · ') + ')</span>' : ''}</td></tr>
<tr><td>closure index rows</td><td>${m.shards.closureIndexRows}</td></tr>
<tr><td>archived closure records</td><td>${m.shards.archiveRecords}</td></tr>
</table><p class="muted">severity counts are a heuristic parse of red's own rows — the ledger is the record</p>` : '<p class="muted">ledger not yet created (red-merge-born at round 1)</p>'}
<h2>Friction — logged pain points</h2>
${m.friction.count ? `<p>${m.friction.count} attributed entr${m.friction.count === 1 ? 'y' : 'ies'} · latest: <span class="muted">${esc((m.friction.last || '').slice(0, 160))}</span></p>` : '<p class="muted">none logged yet</p>'}
<h2>Seats live now</h2>
${live.length ? `<table>${live.map((s) => `<tr><td class="liveseat">${esc(s.label)}</td><td class="muted">since ${esc(s.started || '?')}</td></tr>`).join('')}</table>` : '<p class="muted">none — between seats or complete</p>'}
<h2>Recent completions</h2>
<table>${done.slice(-8).reverse().map((s) => `<tr><td>${esc(s.label)}</td><td class="muted">${esc(s.result || '')}</td></tr>`).join('\n')}</table>
</body>`
}

function generate(runDir, transcriptDir) {
  const html = renderHtml(buildModel(runDir, transcriptDir))
  const out = join(runDir, 'dashboard.html')
  writeFileSync(out, html)
  return out
}

function main() {
  const [runDir, transcriptDir, flag] = process.argv.slice(2)
  if (!runDir || !transcriptDir) { console.error('usage: node render-run-dashboard.mjs <runDir> <workflow-transcript-dir> [--watch]'); process.exit(1) }
  const out = generate(runDir, transcriptDir)
  console.log('dashboard:', out, flag === '--watch' ? '(watching — regenerates every 15s, Ctrl-C to stop)' : '(static snapshot — re-run or use --watch to refresh)')
  if (flag === '--watch') {
    // Marker-keyed lifetime: the watcher lives exactly as long as the run does. run-setup
    // writes .claude/run-live.json; run-capture removes it — when it goes, one final render
    // and exit. The dashboard becomes innate to the run without touching the workflow.
    const marker = join(process.cwd(), '.claude', 'run-live.json')
    const timer = setInterval(() => {
      try { generate(runDir, transcriptDir) } catch (e) { console.error('regen failed:', String(e).slice(0, 120)) }
      if (!existsSync(marker)) {
        try { generate(runDir, transcriptDir) } catch {}
        console.log('run-live marker gone — final render written, watcher exiting')
        clearInterval(timer)
      }
    }, 15000)
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main()

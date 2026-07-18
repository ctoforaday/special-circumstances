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
// Engine's pinned v1 mass mapping (for grade-migration arithmetic only).
const MASSD = { trivial: 0.5, low: 1, 'low-medium': 1.5, medium: 2, 'medium-high': 2.5, high: 3, certain: 3.5, realized: 0 }

// Seat classification from the transcript's first user message — the journal carries only
// {type, key, agentId} (no labels; labels are a panel affordance the harness does not
// persist), so seat identity is recovered exactly the way cost-audit.mjs recovers it. A
// flat-files lesson in miniature: structured state living in prose, parsed by regex.
export function classifySeat(head) {
  let m
  if ((m = head.match(/Red audit, round (\d+)/))) return { seat: 'red-lens', round: +m[1] }
  if ((m = head.match(/Red merge, round (\d+)/))) return { seat: 'red-merge', round: +m[1] }
  if ((m = head.match(/Blue response, round (\d+)/))) return { seat: 'blue-respond', round: +m[1] }
  if ((m = head.match(/Adjudication, round (\d+)/))) return { seat: 'judge', round: +m[1] }
  if (head.includes('Terminal dispute disposition')) return { seat: 'judge-terminal', round: 0 }
  if (head.includes('Blue synthesis')) return { seat: 'blue-synthesize', round: 0 }
  if (head.includes('Blue lane')) return { seat: 'blue-lane', round: 0 }
  if (head.includes('frontier hypotheses')) return { seat: 'frontier', round: 0 }
  if (head.includes('Final assembly')) return { seat: 'assemble', round: 0 }
  return { seat: 'other', round: 0 }
}

export function buildModel(runDir, transcriptDir) {
  const telemetry = jsonl(join(runDir, 'trajectories', 'board-telemetry.jsonl'))
  const journal = jsonl(join(transcriptDir, 'journal.jsonl'))

  // Lifecycle by agentId from the journal; identity by prompt classification from the
  // agent's own transcript head. started-without-result = live.
  const byId = new Map()
  for (const j of journal) {
    if (!j.agentId) continue
    const s = byId.get(j.agentId) || { agentId: j.agentId, done: false }
    // Keep the RAW result — truncating here fed the summarizer unparseable half-JSON.
    if (j.result !== undefined) { s.done = true; s.result = j.result }
    byId.set(j.agentId, s)
  }
  const seats = new Map()
  for (const [id, s] of byId) {
    const tp = join(transcriptDir, `agent-${id}.jsonl`)
    let head = ''
    try { head = readFileSync(tp, 'utf8').slice(0, 3000) } catch {}
    // Start time lives on the transcript's first line (the journal carries no timestamps).
    let startedMs = null
    try { startedMs = Date.parse(JSON.parse(head.slice(0, head.indexOf('\n'))).timestamp) || null } catch {}
    const c = classifySeat(head)
    const label = c.round ? `${c.seat}-r${c.round}` : c.seat
    seats.set(id, { ...s, label, seat: c.seat, round: c.round, startedMs })
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
    // LINE count, not id-occurrence count: a closure row also NAMES its supersedes ids in
    // the fourth column, so occurrence-counting read 88 for a 52-row index (seen live, run 5).
    closureIndexRows: ledgerTxt ? ledgerTxt.slice(Math.max(0, ledgerTxt.search(/closure index/i))).split('\n').filter((l) => idLine.test(l) && l.includes('|')).length : 0,
    archiveRecords: archiveTxt ? (archiveTxt.match(/^#{1,4}\s+.*R\d+-\d+/gm) || []).length : 0,
  }
  // Blue corpus size in CLAIMS (last blue envelope in the journal), not bytes.
  let blueClaims = null
  for (const j of journal) if (j.result && typeof j.result === 'object' && typeof j.result.claim_count === 'number') blueClaims = j.result.claim_count

  // Judiciary analytics from journal envelopes: rulings by type (the router-vs-decider
  // signal — 64 of 65 rulings across runs 4-5 were `carried` under judge-before-blue
  // ordering; the closing-arguments redesign predicts this distribution diversifies),
  // dispute traffic, and ARGUMENT longevity measured over supersedes CHAINS (raw gap ids
  // under-measure: red mints successor ids each round, so ids live ~1 round while the
  // argument persists across the chain).
  const rulings = {}
  let judgeSittings = 0
  const disputes = { raised: 0, accepted: 0, rejected: 0 }
  const gapRounds = new Map() // id -> { first, last, firstMass, lastMass }
  const parentOf = new Map() // union-find over supersedes edges
  const find = (x) => { while (parentOf.has(x) && parentOf.get(x) !== x) x = parentOf.get(x); return x }
  let redSeen = 0
  for (const j of journal) {
    const r = j.result
    if (!r || typeof r !== 'object') continue
    if (Array.isArray(r.resolutions)) { judgeSittings++; for (const x of r.resolutions) rulings[x.resolution] = (rulings[x.resolution] || 0) + 1 }
    if (Array.isArray(r.grade_disputes)) disputes.raised += r.grade_disputes.length
    if (Array.isArray(r.dispute_responses)) for (const d of r.dispute_responses) { if (d.response in disputes) disputes[d.response]++ }
    if (r.verdict && Array.isArray(r.gaps)) {
      redSeen++
      for (const g of r.gaps) {
        const gm = (MASSD[g.likelihood] ?? 0) * (MASSD[g.impact] ?? 0)
        const e = gapRounds.get(g.id) || { first: redSeen, firstMass: gm }
        e.last = redSeen; e.lastMass = gm
        gapRounds.set(g.id, e)
        for (const anc of g.supersedes || []) parentOf.set(g.id, find(anc))
      }
    }
  }
  // Chains: group ids by root; chain span = min(first)..max(last); migration = last vs first mass.
  const chains = new Map()
  for (const [id, e] of gapRounds) {
    const root = find(id)
    const c = chains.get(root) || { first: e.first, last: e.last, firstMass: e.firstMass, lastMass: e.lastMass, ids: 0 }
    c.ids++
    if (e.first <= c.first) { c.first = e.first; c.firstMass = e.firstMass }
    if (e.last >= c.last) { c.last = e.last; c.lastMass = e.lastMass }
    chains.set(root, c)
  }
  const chainSpans = {}
  let migDown = 0, migUp = 0, migFlat = 0
  for (const c of chains.values()) {
    const span = c.last - c.first + 1
    chainSpans[span] = (chainSpans[span] || 0) + 1
    if (span > 1) { const d = c.lastMass - c.firstMass; if (d < 0) migDown++; else if (d > 0) migUp++; else migFlat++ }
  }
  const judiciary = { judgeSittings, rulings, disputes, chainSpans, chains: chains.size, migDown, migUp, migFlat }

  // Progress through the workflow's big steps: Frontier -> Blue lanes -> Synthesis ->
  // rounds 1..maxRounds (each: lenses -> merge -> respond [-> judge]) -> Assembly.
  // The ceiling divides the bar; judged termination may end it early — stated on the bar.
  const MAX_ROUNDS_CEILING = 8
  const seatList = [...seats.values()]
  const seen = (seat, round = null) => seatList.some((s) => s.seat === seat && (round === null || s.round === round))
  const doneSeat = (seat, round = null) => seatList.some((s) => s.seat === seat && (round === null || s.round === round) && s.done)
  const allDone = (seat, round = null) => { const xs = seatList.filter((s) => s.seat === seat && (round === null || s.round === round)); return xs.length > 0 && xs.every((s) => s.done) }
  const steps = []
  steps.push({ name: 'frontier', state: doneSeat('frontier') ? 'done' : seen('frontier') ? 'live' : 'todo' })
  steps.push({ name: 'blue lanes', state: allDone('blue-lane') ? 'done' : seen('blue-lane') ? 'live' : 'todo' })
  steps.push({ name: 'synthesis', state: doneSeat('blue-synthesize') ? 'done' : seen('blue-synthesize') ? 'live' : 'todo' })
  for (let r = 1; r <= MAX_ROUNDS_CEILING; r++) {
    const roundDone = doneSeat('blue-respond', r)
    const anySeen = seen('red-lens', r) || seen('red-merge', r) || seen('blue-respond', r) || seen('judge', r)
    steps.push({ name: `round ${r}`, state: roundDone ? 'done' : anySeen ? 'live' : 'todo' })
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
  return { runDir, telemetry, latest, seats: seatList, cost, apiRounds: rounds, agents, friction, shards, blueClaims, steps, rates, judiciary, generated: new Date().toISOString() }
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

// Grade codes for phone-width tables (full names in the tooltip): 3h · 2mh · 9m …
const SEV_CODE = { certain: 'c', high: 'h', 'medium-high': 'mh', medium: 'm', 'low-medium': 'lm', low: 'l', trivial: 't', realized: 'r' }

// Envelope results are summarized, never dumped (phone review: raw JSON sliced mid-brace).
export function summarizeResult(raw) {
  if (typeof raw !== 'string') raw = JSON.stringify(raw)
  let j = null
  try { j = JSON.parse(raw) } catch {}
  if (!j || typeof j !== 'object') return raw.slice(0, 110)
  const bits = []
  if (j.verdict) bits.push(`verdict ${j.verdict}`)
  if (typeof j.claim_count === 'number') bits.push(`${j.claim_count} claims`)
  if (Array.isArray(j.gaps)) bits.push(`${j.gaps.length} gaps`)
  if (Array.isArray(j.resolutions)) bits.push(`${j.resolutions.length} ruling${j.resolutions.length === 1 ? '' : 's'}`)
  if (typeof j.citations_checked === 'number') bits.push(`${j.citations_checked} citations checked`)
  if (Array.isArray(j.friction) && j.friction.length) bits.push(`${j.friction.length} friction`)
  return bits.length ? bits.join(' · ') : raw.slice(0, 110)
}

export function renderHtml(m) {
  const sevRow = (t) => t && t.new_mint && t.new_mint.by_severity
    ? Object.entries(t.new_mint.by_severity).map(([k, v]) => `${esc(v)}${esc(SEV_CODE[k] || k)}`).join(' ') : '—'
  // Duplicate live seats collapse to one row with a count (six identical lens rows, phone review).
  const liveGroups = new Map()
  for (const s of m.seats.filter((x) => !x.done)) {
    const g = liveGroups.get(s.label) || { label: s.label, n: 0, oldest: null }
    g.n++
    if (s.startedMs && (!g.oldest || s.startedMs < g.oldest)) g.oldest = s.startedMs
    liveGroups.set(s.label, g)
  }
  const live = [...liveGroups.values()]
  const done = m.seats.filter((s) => s.done)
  const SHORT = { frontier: 'front', 'blue lanes': 'lanes', synthesis: 'synth', assembly: 'asm' }
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
.scrollx { overflow-x: auto; } .nowrap { white-space: nowrap; }
td, th { font-variant-numeric: tabular-nums; }
</style>
<body class="viz-root">
<h1>FEOV run · ${esc(m.runDir.split(/[\\/]/).pop())}</h1>
<p class="muted">generated ${esc(m.generated)} · auto-refreshes every 20s · dollars are list-rate estimates</p>
<h2>Progress (rounds segmented by the ceiling — judged termination may end the run earlier)</h2>
<div class="bar">${m.steps.map((s) => `<div class="seg ${s.state}" title="${esc(s.name)}: ${esc(s.state)}"></div>`).join('')}</div>
<div class="barlabels">${m.steps.map((s) => `<span title="${esc(s.name)}">${esc(SHORT[s.name] || s.name.replace('round ', 'r'))}</span>`).join('')}</div>
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
<div class="scrollx"><table><tr><th>round</th><th>opened</th><th>closed</th><th>still open</th><th>close rate</th><th>max sev</th><th title="new mints by severity: c certain, h high, mh medium-high, m medium, lm low-medium, l low, t trivial">mints (c/h/mh/m/lm/l/t)</th><th>mass</th><th>deltas</th></tr>
${m.rates.map((r, i) => { const t = m.telemetry[i]; return `<tr><td>${esc(r.round)}</td><td>${esc(r.opened)}</td><td>${esc(r.closed)}</td><td>${esc(r.open)}</td><td>${esc(r.closeRate)}%</td><td>${esc(t.max_severity ?? '—')}</td><td class="nowrap">${sevRow(t)}</td><td>${esc(t.mass ?? '—')}</td><td>${(t.accepted_deltas || []).length}</td></tr>` }).join('\n')}
</table></div>
<h2>Red's board (contents, not bytes)</h2>
${m.shards.ledgerExists ? `<table>
<tr><td>open gaps on the ledger</td><td>${m.shards.openRows === 0 && m.latest && m.latest.open_count > 0 ? `${esc(m.latest.open_count)} <span class="muted">(from telemetry — ledger open-section rows not parseable)</span>` : `${m.shards.openRows}${Object.keys(m.shards.openBySeverity).length ? ' <span class="muted">(' + Object.entries(m.shards.openBySeverity).map(([k, v]) => `${esc(k)}:${esc(v)}`).join(' · ') + ')</span>' : ''}`}</td></tr>
<tr><td>closure index rows</td><td>${m.shards.closureIndexRows}</td></tr>
<tr><td>archived closure records</td><td>${m.shards.archiveRecords}</td></tr>
</table><p class="muted">severity counts are a heuristic parse of red's own rows — the ledger is the record</p>` : '<p class="muted">ledger not yet created (red-merge-born at round 1)</p>'}
<h2>Judiciary (rulings, disputes, and how long arguments actually run)</h2>
${m.judiciary.judgeSittings ? `<table>
<tr><td>judge sittings</td><td>${m.judiciary.judgeSittings}</td></tr>
<tr><td>rulings by type</td><td>${Object.entries(m.judiciary.rulings).map(([k, v]) => `${esc(k)}: ${esc(v)}`).join(' · ') || '—'} <span class="muted">(a carried-dominated bench is routing, not deciding — the closing-arguments ordering predicts this diversifies)</span></td></tr>
<tr><td>grade disputes</td><td>raised ${m.judiciary.disputes.raised} · accepted ${m.judiciary.disputes.accepted} · rejected ${m.judiciary.disputes.rejected}</td></tr>
<tr><td>argument chains (supersedes-aware)</td><td>${m.judiciary.chains} chains · by rounds alive: ${Object.entries(m.judiciary.chainSpans).sort((a, b) => a[0] - b[0]).map(([k, v]) => `${esc(v)}×${esc(k)}r`).join(' · ')}</td></tr>
<tr><td>grade migration on multi-round chains</td><td>down ${m.judiciary.migDown} · up ${m.judiciary.migUp} · flat ${m.judiciary.migFlat} <span class="muted">(first-vs-last mass along the chain — the downgrade process)</span></td></tr>
</table>` : '<p class="muted">no judge sittings yet</p>'}
<h2>Friction — logged pain points</h2>
${m.friction.count ? `<p>${m.friction.count} attributed entr${m.friction.count === 1 ? 'y' : 'ies'} · latest: <span class="muted">${esc((m.friction.last || '').slice(0, 160))}</span></p>` : '<p class="muted">none logged yet</p>'}
<h2>Seats live now</h2>
${live.length ? `<table>${live.map((g) => `<tr><td class="liveseat">${g.n > 1 ? `${g.n}× ` : ''}${esc(g.label)}</td><td class="muted">${g.oldest ? 'running ' + Math.max(1, Math.round((Date.now() - g.oldest) / 60000)) + ' min' : 'running'}</td></tr>`).join('')}</table>` : '<p class="muted">none — between seats or complete</p>'}
<h2>Recent completions</h2>
<table>${done.slice(-8).reverse().map((s) => `<tr><td class="nowrap">${esc(s.label)}</td><td class="muted">${esc(summarizeResult(s.result || ''))}</td></tr>`).join('\n')}</table>
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

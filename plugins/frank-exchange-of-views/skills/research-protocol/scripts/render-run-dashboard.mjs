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
import { parseRenderedRows, latestSection } from './scorecards.mjs'
import { classifySeat } from './seat-classify.mjs'
import { PRICES, tier } from './cost-audit.mjs'

const jsonl = (p) => existsSync(p)
  ? readFileSync(p, 'utf8').split('\n').filter(Boolean).map((l) => { try { return JSON.parse(l) } catch { return null } }).filter(Boolean)
  : []

// Pricing comes from cost-audit.mjs, the module that owns it. The dashboard used to
// keep its own single RATE row at fable-tier and apply it to EVERY seat regardless of
// model — so a run with haiku bulk seats was billed as if all of it were the most
// expensive tier. On the first live run that read $189.59 against a true $53.92: a
// 3.5x overstatement on the one number a human watches while deciding whether to let a
// run continue. Fourth instance in this codebase of two modules holding the same fact
// and disagreeing; the fix is the same one each time.
// Engine's pinned v1 mass mapping (for grade-migration arithmetic only).
const MASSD = { trivial: 0.5, low: 1, 'low-medium': 1.5, medium: 2, 'medium-high': 2.5, high: 3, certain: 3.5, realized: 0 }

// Seat classification from the transcript's first user message — the journal carries only
// {type, key, agentId} (no labels; labels are a panel affordance the harness does not
// persist), so seat identity is recovered exactly the way cost-audit.mjs recovers it. A
// flat-files lesson in miniature: structured state living in prose, parsed by regex.
// Re-exported from seat-classify.mjs, the single seat table. This used to be a
// private copy that had already drifted from cost-audit's.
export { classifySeat } from './seat-classify.mjs'

// projectCompletion answers "how much longer?" from THIS run's own measured seat
// durations — the only honest basis, since a run's pace depends on its model tier, its
// report size, and how hard red is working.
//
// It reports a RANGE from the fastest and slowest completed seat of each class, never a
// point estimate: merge seats in the first live run took 11.4, 16.2 and 16.7 minutes as
// the board grew, so a single number would have been wrong in a predictable direction.
//
// It also states its ASSUMPTION rather than hiding it. The run's round ceiling is a
// launch argument the engine never writes down, so the dashboard cannot know how many
// rounds remain; this projects the work now in flight plus assembly, and reports what
// one more round would add so the reader can do the arithmetic the tool cannot.
export function projectCompletion(seats, nowMs = Date.now()) {
  // Number.isFinite, not truthiness: a timestamp of 0 is a legitimate value and a
  // falsy one, so `s.startedMs &&` silently drops the seat. That is the same defect
  // this run found in the bench's review_flag, where `false` meant "no review needed"
  // and was scored as no opinion at all.
  const ms = (v) => Number.isFinite(v)
  const done = [...seats].filter((s) => ms(s.startedMs) && ms(s.endedMs) && s.endedMs > s.startedMs)
  const live = [...seats].filter((s) => !s.done && ms(s.startedMs))
  if (done.length && [...seats].some((s) => s.seat === 'assemble' && s.done)) {
    return { state: 'complete', lowMin: 0, highMin: 0, basis: 'assembly finished' }
  }
  const byClass = {}
  for (const s of done) (byClass[s.seat] ||= []).push((s.endedMs - s.startedMs) / 60000)
  const span = (seat) => {
    const v = byClass[seat]
    return v && v.length ? { lo: Math.min(...v), hi: Math.max(...v) } : null
  }
  // Assembly has no precedent on a first run; blue-synthesize is the nearest analogue
  // (one seat, reads everything, writes the long document).
  const assembly = span('assemble') || span('blue-synthesize')
  let lo = 0, hi = 0
  const missing = []
  for (const s of live) {
    const sp = span(s.seat)
    const elapsed = (nowMs - s.startedMs) / 60000
    if (!sp) { missing.push(s.label); continue }
    lo += Math.max(0, sp.lo - elapsed)
    hi += Math.max(0, sp.hi - elapsed)
  }
  if (![...seats].some((s) => s.seat === 'assemble')) {
    if (assembly) { lo += assembly.lo; hi += assembly.hi } else missing.push('assembly')
  }
  const roundCost = ['red-lens', 'red-merge', 'blue-respond']
    .map((c) => span(c)).filter(Boolean)
    .reduce((a, sp) => ({ lo: a.lo + sp.lo, hi: a.hi + sp.hi }), { lo: 0, hi: 0 })
  return {
    state: live.length || !byClass.assemble ? 'running' : 'complete',
    lowMin: Math.round(lo), highMin: Math.round(hi),
    perRoundLowMin: Math.round(roundCost.lo), perRoundHighMin: Math.round(roundCost.hi),
    basis: `${done.length} completed seat(s) in this run`,
    // Named so the reader can discount the estimate rather than trust it blindly.
    unmeasured: missing,
  }
}

export function buildModel(runDir, transcriptDir, config = {}) {
  // Canonical run config is inputs/run-config.json (written by setup — the engine is
  // sandboxed and cannot write it). Any CLI --model/--max-rounds/… override a stored value.
  let fileConfig = {}
  try { fileConfig = JSON.parse(readFileSync(join(runDir, 'inputs', 'run-config.json'), 'utf8')) } catch {}
  const cfg = { ...fileConfig, ...Object.fromEntries(Object.entries(config).filter(([, v]) => v != null)) }
  config = cfg
  // Live board ground truth comes from the TOOL's render (migrated 2026-07-19), with the
  // legacy trajectories/ path as a fallback for pre-migration runs. Reading only the old
  // path left the live panel blank while the seed panel looked populated — backwards.
  const renderTel = join(runDir, 'records', 'render-shadow', 'board-telemetry.jsonl')
  const legacyTel = join(runDir, 'trajectories', 'board-telemetry.jsonl')
  const telemetry = jsonl(existsSync(renderTel) ? renderTel : legacyTel)
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
    // Start and end come from the FILE, not its contents.
    //
    // The old form parsed head.slice(0, head.indexOf('\n')) as JSON to read a
    // "timestamp" field. It returned null for 22 of 23 seats in the first live run and
    // failed silently, because a transcript's first record serialises the seat's entire
    // opening prompt BEFORE its timestamp — many thousands of characters — so within a
    // 3000-byte head there is no newline (indexOf gives -1, slice(0,-1) is invalid
    // JSON) and no timestamp to find either. Only `frontier`, whose prompt is short,
    // ever worked. Reading far enough into 23 multi-megabyte files to reach the field
    // would make the dashboard the most expensive process in the run.
    //
    // birthtime is when the harness created the transcript, which is when the seat
    // started; mtime is its last write, which for a finished seat is when it ended.
    let startedMs = null
    let endedMs = null
    try {
      const st = statSync(tp)
      startedMs = st.birthtimeMs || st.ctimeMs || null
      if (s.done) endedMs = st.mtimeMs
    } catch {}
    const c = classifySeat(head)
    const label = c.round ? `${c.seat}-r${c.round}` : c.seat
    seats.set(id, { ...s, label, seat: c.seat, round: c.round, startedMs, endedMs })
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
        // Price each turn at ITS OWN model's rate; an unrecognised model falls back to
        // the dearest row, so an estimate errs upward rather than flattering the run.
        const [pin, pout, pcr, pcw] = PRICES[tier((j.message && j.message.model) || '')]
        cost += ((u.input_tokens || 0) * pin + (u.output_tokens || 0) * pout +
          (u.cache_read_input_tokens || 0) * pcr + (u.cache_creation_input_tokens || 0) * pcw) / 1e6
      }
    }
  }

  // CONTENTS over bytes (review feedback): counts and states, not file sizes.
  const readIf = (rel) => { const p = join(runDir, rel); return existsSync(p) ? readFileSync(p, 'utf8') : null }
  const frictionTxt = readIf('friction.md')
  // Entries are ATTRIBUTED LINES ("blue-synthesize: ...", "red-merge-r2: ..."), not
  // markdown bullets. Counting /^- / found zero of the seven real entries in the first
  // live run and the dashboard reported "none logged yet" — the writer and the reader
  // disagreeing about the format, which is the same defect class as the scorecard
  // parser and the seat table. What it hid was the entire tool-failure surface of the
  // run: a Write guard blocking the one seat whose deliverable is a file, `merge mint`
  // having no amend path, `spot-check` unable to record an honestly-empty round.
  const frictionLines = frictionTxt
    ? frictionTxt.split('\n').map((l) => l.trim()).filter((l) => l && !l.startsWith('#'))
    : []
  const friction = {
    count: frictionLines.length,
    last: frictionLines.length ? frictionLines[frictionLines.length - 1] : null,
  }
  // Projections all land in records/render-shadow/ (the tool's render output); read there
  // first, with the legacy red/ path as a fallback for pre-migration runs. Reading only red/
  // left the red board blank on every post-migration run — the same miss the telemetry read
  // above already fixed.
  const ledgerTxt = readIf('records/render-shadow/ledger.md') || readIf('red/ledger.md')
  const archiveTxt = readIf('records/render-shadow/archive.md') || readIf('red/archive.md')
  // Lens findings now live on the record; findings.md is the render projection (in
  // records/render-shadow/, where every projection lands — NOT red/, which is why this reads
  // there first). The count is red's raw leaf-audit volume BEFORE the merge coalesces it into
  // gaps, so it appears during the red rounds, ahead of the merge-born ledger.
  const findingsTxt = readIf('records/render-shadow/findings.md') || readIf('red/findings.md')
  // Citations are the other half of the record-canonical pair: cite events, projected to the
  // citation-ledger. One data row per verified claim (the header line is skipped).
  const citationsTxt = readIf('records/render-shadow/citation-ledger.md') || readIf('red/citation-ledger.md')
  // Ordered MOST SPECIFIC FIRST, and it must stay that way: the match below is a
  // substring test, so listing `high` before `medium-high` made every
  // medium-high row report as high (and every low-medium row as medium) —
  // the compound grades have the simple ones as substrings. That silently
  // inflated the high-severity count on the tile a human uses to judge whether
  // the board is getting worse.
  const GRADES = ['medium-high', 'low-medium', 'certain', 'high', 'medium', 'low', 'trivial', 'realized']
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
    findings: findingsTxt ? (findingsTxt.match(/^- /gm) || []).length : 0,
    citations: citationsTxt ? citationsTxt.split('\n').filter((l) => l.includes(' | ')).length : 0,
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
  let latestVerdict = null, verdictRound = 0 // red's most recent round verdict (FAIL until PASS or ceiling)
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
      latestVerdict = r.verdict; verdictRound = redSeen
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
  const judiciary = { judgeSittings, rulings, disputes, chainSpans, chains: chains.size, migDown, migUp, migFlat, latestVerdict, verdictRound }

  // Progress through the workflow's big steps: Frontier -> Blue lanes -> Synthesis ->
  // rounds 1..maxRounds (each: lenses -> merge -> respond [-> judge]) -> Assembly.
  // The ceiling divides the bar; judged termination may end it early — stated on the bar.
  const MAX_ROUNDS_CEILING = config.maxRounds ? Number(config.maxRounds) : 8
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
  const eta = projectCompletion(seatList)
  return { runDir, telemetry, latest, seats: seatList, cost, apiRounds: rounds, agents, friction, shards, blueClaims, steps, rates, judiciary, eta, config, generated: new Date().toISOString() }
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

// W2h — the per-chair scoreboard. The visibility loop's third leg: the same
// numbers the seats were given, in front of the human watching the run.
//
// The CLASS is rendered, not just the value. A benchmark and a diagnostic look
// identical as bare figures, and treating a diagnostic as a target is how red
// learns that grade stability is a virtue rather than a symptom. Benchmarks read
// bold, diagnostics muted, and a detector that has fired reads as an alarm
// because any nonzero detector is a finding.
export function scorecardSection(runDir) {
  // GROUND TRUTH NOW, or nothing. The dashboard used to render inputs/*-scorecard.md —
  // but those are the PRIOR run's scorecards, staged as the chairs' SEED (their memory,
  // shown in their prompts). Rendering the seed on the live dashboard reported a
  // predecessor's numbers as this run's, so a fresh run showed a populated bench and a
  // repair_regression_ratio before anything had been repaired — limbo, not ground truth.
  //
  // The seed belongs in the chairs' prompts; this run's own scorecards do not exist until
  // CAPTURE (post-run) computes them. So the live dashboard shows THIS run's computed
  // scorecards when they exist, and a clean blank slate otherwise — never the seed.
  const captured = join(runDir, 'records', 'render-shadow', 'scorecards')
  const inputs = captured
  if (!existsSync(inputs)) return ''
  const cards = readdirSync(inputs).filter((f) => f.endsWith('-scorecard.md'))
  if (!cards.length) return ''
  const blocks = []
  for (const f of cards) {
    const chair = f.replace('-scorecard.md', '')
    const body = readFileSync(join(inputs, f), 'utf8')
    // Parsed by scorecards.mjs, which owns the format it renders. Three
    // hand-rolled copies of this parser each carried the same colon defect.
    const latest = latestSection(body)
    const rows = parseRenderedRows(latest)
    if (!rows.length) continue
    blocks.push(`<h3>${esc(chair)}</h3><table>` + rows.map((r) => {
      const { metric, cls, clause, value } = r
      const fired = cls === 'detector' && value && value.trim() !== '0'
      const style = fired ? 'color:#b00;font-weight:700'
        : cls === 'benchmark' ? 'font-weight:700'
          : cls === 'diagnostic' ? 'opacity:.65' : 'opacity:.8'
      return `<tr><td style="${style}">${esc(metric)}</td><td>${esc(cls)}</td>` +
        `<td style="${style}">${esc(value || 'not computed')}</td><td>${esc(clause.trim())}</td></tr>`
    }).join('') + '</table>')
  }
  return blocks.length ? `<h2>scorecards — this run</h2>${blocks.join('')}` : ''
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
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
<title>FEOV run — ${esc(m.runDir.split(/[\\/]/).pop())}</title>
<style>
.viz-root { color-scheme: light; --surface-1:#fcfcfb; --text-primary:#0b0b0b; --text-secondary:#52514e; --series-1:#2a78d6;
  font: 14px/1.45 system-ui, sans-serif; background: var(--surface-1); color: var(--text-primary); padding: 20px; max-width: 900px; margin: auto; }
@media (prefers-color-scheme: dark) { :root:where(:not([data-theme="light"])) .viz-root { color-scheme: dark; --surface-1:#1a1a19; --text-primary:#ffffff; --text-secondary:#c3c2b7; --series-1:#3987e5; } }
:root[data-theme="dark"] .viz-root { color-scheme: dark; --surface-1:#1a1a19; --text-primary:#ffffff; --text-secondary:#c3c2b7; --series-1:#3987e5; }
h1 { font-size: 18px; margin-bottom: 2px; } h2 { font-size: 14px; margin: 18px 0 6px; color: var(--text-secondary); }
.topic { font-size: 15px; font-weight: 600; margin: 0 0 6px; line-height: 1.35; }
.topic::before { content: "researching: "; font-weight: 400; color: var(--text-secondary); }
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
@media (max-width: 560px) { .viz-root { padding: 12px; } h1 { font-size: 16px; } .tile { flex: 1 1 40%; min-width: 0; } table { display: block; overflow-x: auto; } }
</style>
<body class="viz-root">
<h1>FEOV run · ${esc(m.runDir.split(/[\\/]/).pop())}</h1>
${m.config && m.config.topic ? `<p class="topic">${esc(m.config.topic)}</p>` : ''}
<p class="muted">generated ${esc(m.generated)} · auto-refreshes every 20s · dollars are list-rate estimates</p>
${m.config && (m.config.model || m.config.judgmentModel || m.config.maxRounds || m.config.lanes) ? `<h2>Run configuration</h2>
<table>
<tr><td>bulk seats <span class="muted">frontier · lanes · red lenses · blue responses</span></td><td class="nowrap"><b>${esc(m.config.model || 'session default')}</b></td></tr>
<tr><td>judgment seats <span class="muted">synthesis · red-merge · judge · assembly</span></td><td class="nowrap"><b>${esc(m.config.judgmentModel || 'session default')}</b></td></tr>
<tr><td>round ceiling <span class="muted">cost bound — the terminator is red-PASS or judged deadlock</span></td><td class="nowrap">${m.config.maxRounds ? esc(m.config.maxRounds) + ' rounds' : '—'}</td></tr>
<tr><td>blue lanes <span class="muted">best-of-N candidate drafts</span></td><td class="nowrap">${m.config.lanes ? esc(m.config.lanes) : '—'}</td></tr>
</table>` : ''}
<h2>Progress (rounds segmented by the ceiling — judged termination may end the run earlier)</h2>
${m.eta && m.eta.state === 'running' && (m.eta.lowMin || m.eta.highMin) ? `<p class="muted">projected <b>${m.eta.lowMin}–${m.eta.highMin} min</b> remaining for the work now in flight plus assembly, from ${esc(m.eta.basis)}. Each ADDITIONAL round would add roughly ${m.eta.perRoundLowMin}–${m.eta.perRoundHighMin} min: the round ceiling is a launch argument this run never writes down, so the estimate cannot know how many remain.${m.eta.unmeasured && m.eta.unmeasured.length ? ` No completed precedent yet for: ${esc(m.eta.unmeasured.join(', '))} — discount accordingly.` : ''}</p>` : ''}
<div class="bar">${m.steps.map((s) => `<div class="seg ${s.state}" title="${esc(s.name)}: ${esc(s.state)}"></div>`).join('')}</div>
<div class="barlabels">${m.steps.map((s) => `<span title="${esc(s.name)}">${esc(SHORT[s.name] || s.name.replace('round ', 'r'))}</span>`).join('')}</div>
<div class="tiles" style="margin-top:12px">
<div class="tile"><b>${m.latest ? esc(m.latest.mass) : '—'}</b><span>board mass</span></div>
<div class="tile"><b>${m.latest ? esc(m.latest.open_count) : '—'}</b><span>open gaps</span></div>
<div class="tile"><b>${m.latest ? esc(m.latest.max_severity) : '—'}</b><span>max severity</span></div>
<div class="tile"><b>${m.blueClaims ?? '—'}</b><span>blue claims</span></div>
<div class="tile"><b>${m.shards.findings}</b><span>lens findings</span></div>
<div class="tile"><b>${m.shards.citations}</b><span>citations checked</span></div>
<div class="tile"><b>${m.judiciary.latestVerdict || '—'}</b><span>latest verdict${m.judiciary.verdictRound ? ` (r${m.judiciary.verdictRound})` : ''}</span></div>
<div class="tile"><b>${m.friction.count}</b><span>friction entries</span></div>
${m.eta && m.eta.state === 'running' && (m.eta.lowMin || m.eta.highMin)
  ? `<div class="tile"><b>${m.eta.lowMin}–${m.eta.highMin}m</b><span>projected remaining</span></div>`
  : ''}
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
<tr><td>open gaps on the ledger</td><td>${m.latest && m.latest.open_count > 0 && m.shards.openRows < m.latest.open_count ? `${esc(m.latest.open_count)} <span class="muted">(from telemetry — heuristic ledger parse found ${esc(m.shards.openRows)} row(s); telemetry wins when the parse under-reads)</span>` : `${m.shards.openRows}${Object.keys(m.shards.openBySeverity).length ? ' <span class="muted">(' + Object.entries(m.shards.openBySeverity).map(([k, v]) => `${esc(k)}:${esc(v)}`).join(' · ') + ')</span>' : ''}`}</td></tr>
<tr><td>closure index rows</td><td>${m.shards.closureIndexRows}</td></tr>
<tr><td>archived closure records</td><td>${m.shards.archiveRecords}</td></tr>
<tr><td>lens findings recorded</td><td>${m.shards.findings} <span class="muted">(raw leaf audit, before the merge coalesces into gaps)</span></td></tr>
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
</body>${scorecardSection(m.runDir || "")}`
}

function generate(runDir, transcriptDir, config) {
  const html = renderHtml(buildModel(runDir, transcriptDir, config))
  const out = join(runDir, 'dashboard.html')
  writeFileSync(out, html)
  return out
}

function parseFlag(argv, name) { const i = argv.indexOf(name); return i >= 0 && i + 1 < argv.length ? argv[i + 1] : null }

function main() {
  const argv = process.argv.slice(2)
  const [runDir, transcriptDir] = argv
  if (!runDir || !transcriptDir) { console.error('usage: node render-run-dashboard.mjs <runDir> <workflow-transcript-dir> [--watch] [--model M] [--judgment-model M] [--max-rounds N] [--lanes N]'); process.exit(1) }
  // The run config (models per stage, round ceiling, lanes) is a launch argument the sandboxed
  // engine cannot write down — the invoker passes it here so the dashboard can display it.
  const config = {
    model: parseFlag(argv, '--model'),
    judgmentModel: parseFlag(argv, '--judgment-model'),
    maxRounds: parseFlag(argv, '--max-rounds'),
    lanes: parseFlag(argv, '--lanes'),
  }
  const watch = argv.includes('--watch')
  const out = generate(runDir, transcriptDir, config)
  console.log('dashboard:', out, watch ? '(watching — regenerates every 15s, Ctrl-C to stop)' : '(static snapshot — re-run or use --watch to refresh)')
  if (watch) {
    // Marker-keyed lifetime: the watcher lives exactly as long as the run does. run-setup
    // writes .claude/run-live.json; run-capture removes it — when it goes, one final render
    // and exit. The dashboard becomes innate to the run without touching the workflow.
    const marker = join(process.cwd(), '.claude', 'run-live.json')
    const timer = setInterval(() => {
      try { generate(runDir, transcriptDir, config) } catch (e) { console.error('regen failed:', String(e).slice(0, 120)) }
      if (!existsSync(marker)) {
        try { generate(runDir, transcriptDir, config) } catch {}
        console.log('run-live marker gone — final render written, watcher exiting')
        clearInterval(timer)
      }
    }, 15000)
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main()

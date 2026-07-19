#!/usr/bin/env node
// record-parity-check.mjs — the R2.5 gate (plans/record-tool.md §III). Compares
// the HAND-WRITTEN artifacts against the SHADOW renders on extracted facts:
// gap ids, severities, closure classes, telemetry fields, round-section
// presence, changelog rounds, citation row counts. Prose compares
// presence-not-text. Exit 2 on any divergence — zero-FAIL gates R3.
//
// Usage: node record-parity-check.mjs <runDir>
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const read = (p) => (existsSync(p) ? readFileSync(p, 'utf8') : null)
const idRe = /R\d+-\d+/g

export function extractLedgerFacts(text) {
  // ABSENT (null from read) and EMPTY (a present, zero-byte file) are different
  // facts and must not collapse. `!text` treated an empty ledger as "nothing to
  // compare", and compare()'s `if (hl && sl)` guard then skipped the ledger
  // check entirely — so a run whose ledger was truncated to zero bytes (killed
  // mid-append, the exact failure this repo keeps hitting) reported parity PASS
  // against a shadow full of open gaps. That is the same "gate with nothing to
  // compare" this file already refuses to do for a missing shadow.
  if (text == null) return null
  const cut = text.search(/closure index/i)
  const openSec = cut >= 0 ? text.slice(0, cut) : text
  const closeSec = cut >= 0 ? text.slice(cut) : ''
  const openIds = new Set(openSec.match(idRe) || [])
  const closedIds = new Set()
  for (const line of closeSec.split('\n')) {
    const m = line.match(/^(R\d+-\d+)\s*\|/) || line.match(/^###?\s+(R\d+-\d+)/)
    if (m) closedIds.add(m[1])
  }
  // supersedes columns name closed ids inside OPEN rows — strip ids that also
  // appear in the closure index from the open set only if row-anchored parsing
  // is unavailable (hand formats vary; the fact compared is the closure set +
  // the open set MINUS closure references).
  for (const id of closedIds) openIds.delete(id)
  return { openIds, closedIds }
}

export function extractTelemetryFacts(text) {
  // Same absent-vs-empty distinction as extractLedgerFacts: an empty telemetry
  // file is a run that recorded nothing, which must be COMPARED against the
  // shadow (and diverge), not waved through.
  if (text == null) return null
  const rounds = new Map()
  for (const l of text.split('\n').filter(Boolean)) {
    try { const j = JSON.parse(l); rounds.set(j.round, { open_count: j.open_count, mass: j.mass, max_severity: j.max_severity }) } catch {}
  }
  return rounds
}

export function compare(runDir) {
  // READERS READ. This script used to call the mjs record library's render() to
  // refresh the shadow first; it no longer does, and the point is architectural
  // rather than incidental. The writer is now the compiled feov-record binary,
  // and every seat verb renders on mutation, so the shadow on disk is already
  // current — re-rendering here would mean a reader carrying a dependency on
  // whichever implementation happens to own the write path this month. The
  // JSONL and its projections are the interface; that was the whole point of
  // the format.
  const shadow = join(runDir, 'records', 'render-shadow')
  if (!existsSync(shadow)) {
    // Distinguish "no records yet" from "parity passed with nothing to compare":
    // silently comparing two absent things is how a gate reports success while
    // measuring nothing.
    throw new Error(`record-parity-check: no shadow renders at ${shadow} — the run has no records/ yet, ` +
      `or no seat has written through feov-record. A parity gate with nothing to compare is not a passing gate.`)
  }
  const divergences = []
  const push = (artifact, what, hand, sh) => divergences.push(`${artifact}: ${what} — hand=${hand} shadow=${sh}`)

  const hl = extractLedgerFacts(read(join(runDir, 'red', 'ledger.md')))
  const sl = extractLedgerFacts(read(join(shadow, 'ledger.md')))
  if (hl && sl) {
    for (const id of hl.openIds) if (!sl.openIds.has(id)) push('ledger', `open ${id} missing from shadow`, 'open', 'absent')
    for (const id of sl.openIds) if (!hl.openIds.has(id)) push('ledger', `open ${id} missing from hand`, 'absent', 'open')
    for (const id of hl.closedIds) if (!sl.closedIds.has(id)) push('ledger', `closure ${id} missing from shadow`, 'closed', 'absent')
    for (const id of sl.closedIds) if (!hl.closedIds.has(id)) push('ledger', `closure ${id} missing from hand`, 'absent', 'closed')
  }

  const ht = extractTelemetryFacts(read(join(runDir, 'trajectories', 'board-telemetry.jsonl')))
  const st = extractTelemetryFacts(read(join(shadow, 'board-telemetry.jsonl')))
  if (ht && st) {
    for (const [r, h] of ht) {
      const s = st.get(r)
      if (!s) { push('telemetry', `round ${r} absent from shadow`, 'present', 'absent'); continue }
      for (const f of ['open_count', 'mass', 'max_severity']) if (String(h[f]) !== String(s[f])) push('telemetry', `round ${r} ${f}`, h[f], s[f])
    }
  }

  const hd = read(join(runDir, 'debate.md'))
  const sd = read(join(shadow, 'debate.md'))
  if (hd && sd) {
    for (const sec of ['### RED', '### BLUE', '### LEAD']) {
      const hc = (hd.match(new RegExp(`^${sec}\\b`, 'gm')) || []).length
      const sc = (sd.match(new RegExp(`^${sec}\\b`, 'gm')) || []).length
      if (hc !== sc) push('debate', `${sec} section count`, hc, sc)
    }
  }

  const hc2 = read(join(runDir, 'blue', 'CHANGELOG.md'))
  const sc2 = read(join(shadow, 'CHANGELOG.md'))
  if (hc2 && sc2) {
    const roundsOf = (t) => new Set((t.match(/^#+.*Round (\d+)/gim) || []).map((h) => h.match(/Round (\d+)/i)[1]))
    const hr = roundsOf(hc2), sr = roundsOf(sc2)
    for (const r of hr) if (!sr.has(r)) push('changelog', `Round ${r} missing from shadow`, 'present', 'absent')
    for (const r of sr) if (!hr.has(r)) push('changelog', `Round ${r} missing from hand`, 'absent', 'present')
  }

  const hcl = read(join(runDir, 'red', 'citation-ledger.md'))
  const scl = read(join(shadow, 'citation-ledger.md'))
  if (hcl && scl) {
    const rows = (t) => t.split('\n').filter((l) => l.includes(' | ')).length
    if (Math.abs(rows(hcl) - rows(scl)) > 0) push('citation-ledger', 'row count', rows(hcl), rows(scl))
  }

  return divergences
}

import { pathToFileURL } from 'node:url'
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const runDir = process.argv[2]
  if (!runDir) { console.error('usage: node record-parity-check.mjs <runDir>'); process.exit(1) }
  const d = compare(runDir)
  if (d.length === 0) console.log('record-parity: PASS — hand and shadow agree on every extracted fact')
  else { console.log(`record-parity: FAIL — ${d.length} divergence(s):`); for (const x of d) console.log(`  - ${x}`) }
  process.exit(d.length ? 2 : 0)
}

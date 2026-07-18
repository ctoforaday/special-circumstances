// lib/record.mjs — the record-management library (plans/record-tool.md R1).
//
// Physics, not legislation: an append-only per-process-shard event log with
// structural idempotency keys, render-time deterministic merge, and projections
// (ledger/archive/telemetry) written atomically. Seats write ONLY through the
// per-seat CLIs (tools/*.mjs) whose verb sets encode role boundaries; this
// library owns append, validation, id minting, replay, merge, dedup, render.
//
// Binding plan-audit directives (sitting 4): renders are temp+atomic-rename;
// winner selection under nonce multiplicity = terminal-event nonce, then
// latest-mtime; every multi-nonce seatId is reported in the render anomaly
// footer — duplicate dispatch is dedup'd AND visible, never silently normal.
import { appendFileSync, existsSync, mkdirSync, readFileSync, readdirSync, renameSync, statSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'
import { randomBytes } from 'node:crypto'

export const GRADES = ['low', 'low-medium', 'medium', 'medium-high', 'high', 'certain', 'realized', 'trivial']
// Pinned v1 mass mapping — mirrors debate.js; changing values bumps the version.
export const MASS_MAPPING_VERSION = 'v1'
export const MASS = { trivial: 0.5, low: 1, 'low-medium': 1.5, medium: 2, 'medium-high': 2.5, high: 3, certain: 3.5, realized: 0 }
export const gapMass = (g) => (MASS[g.likelihood] ?? 0) * (MASS[g.impact] ?? 0)

const recordsDir = (runDir) => join(runDir, 'records')
const pointerPath = (runDir, seatId) => join(recordsDir(runDir), `.active-${seatId}`)
const shardPath = (runDir, seatId, nonce) => join(recordsDir(runDir), `events-${seatId}-${nonce}.jsonl`)

export function roundOf(seatId) {
  const m = /-r(\d+)/.exec(seatId)
  return m ? Number(m[1]) : 0
}

// register: called as every seat's FIRST record action. Duplicate dispatches are
// SEQUENTIAL retries (measured: 8/50 in run 5, all crash re-runs) — register
// ROTATES the nonce, so a re-dispatched seat writes its own shard and the stale
// instance's shard stays inert. Two registers for one seatId are themselves an
// event (visible to the join audit and the anomaly footer).
export function registerSeat(runDir, seatId) {
  if (!seatId || !/^[a-z][a-z0-9-]*$/i.test(seatId)) throw new Error(`record: invalid --seat-id ${JSON.stringify(seatId)}`)
  mkdirSync(recordsDir(runDir), { recursive: true })
  const nonce = randomBytes(4).toString('hex')
  writeFileSync(pointerPath(runDir, seatId), nonce)
  const line = { seq: 0, seatId, nonce, round: roundOf(seatId), type: 'register', key: `${seatId}:register:${nonce}`, payload: {} }
  appendFileSync(shardPath(runDir, seatId, nonce), JSON.stringify(line) + '\n')
  return { nonce, shard: shardPath(runDir, seatId, nonce) }
}

function activeNonce(runDir, seatId) {
  const p = pointerPath(runDir, seatId)
  if (!existsSync(p)) return registerSeat(runDir, seatId).nonce // implicit register — tolerant
  return readFileSync(p, 'utf8').trim()
}

// Structural idempotency keys (plan §II): singleton verbs key on seat+round+verb;
// multi-instance verbs on their stable labels; friction/observe on ordinals.
const SINGLETON = new Set(['position', 'revision', 'verdict', 'spot-check', 'register'])
function deriveKey(seatId, round, type, payload, shardEvents) {
  if (SINGLETON.has(type)) return `${seatId}:${type}`
  const label = payload.gap_id || payload.label || payload.id || payload.observation || payload.reference
  if (label) return `${seatId}:${type}:${label}`
  const ordinal = shardEvents.filter((e) => e.type === type).length + 1
  return `${seatId}:${type}:#${ordinal}`
}

export function readShard(path) {
  if (!existsSync(path)) return []
  return readFileSync(path, 'utf8').split('\n').filter(Boolean).map((l) => { try { return JSON.parse(l) } catch { return null } }).filter(Boolean)
}

// append: the ONLY write path. Validates, derives key, assigns per-shard seq.
export function append(runDir, seatId, type, payload = {}) {
  const nonce = activeNonce(runDir, seatId)
  const shard = shardPath(runDir, seatId, nonce)
  const events = readShard(shard)
  const round = roundOf(seatId)
  validate(runDir, type, payload)
  const key = deriveKey(seatId, round, type, payload, events)
  const line = { seq: events.length, seatId, nonce, round, type, key, payload }
  // Torn-line healing: a crash mid-append can leave an unterminated fragment as the
  // file's last bytes; appending straight onto it would destroy THIS event too
  // (caught by the crash-mid-append test). If the shard does not end in a newline,
  // terminate the fragment first — the fragment stays visibly unparseable, the new
  // event stays whole.
  let prefix = ''
  if (existsSync(shard)) {
    const raw = readFileSync(shard, 'utf8')
    if (raw.length && !raw.endsWith('\n')) prefix = '\n'
  }
  appendFileSync(shard, prefix + JSON.stringify(line) + '\n')
  return line
}

// Validation at append time — enums, anchors, lineage existence, class registry.
function validate(runDir, type, p) {
  for (const g of ['severity', 'likelihood', 'impact', 'complexity_cost']) {
    if (p[g] !== undefined && !GRADES.includes(p[g])) throw new Error(`record: ${type}.${g}=${JSON.stringify(p[g])} not in grade enum`)
  }
  if (type === 'mint') {
    if (!p.acceptance_check) throw new Error('record: mint requires --check (the acceptance check red will run at re-audit — the pre-agreed contract)')
    if (!p.class) throw new Error('record: mint requires --class (or --class-new with --definition/--neighbor/--distinguisher)')
    validateClass(runDir, p)
    for (const anc of p.supersedes || []) {
      if (!allGapIds(runDir).has(anc)) throw new Error(`record: mint supersedes ${anc}, which no mint event has created — dangling lineage refused`)
    }
  }
  if (type === 'close') {
    if (!p.gap_id) throw new Error('record: close requires --id')
    if (!allGapIds(runDir).has(p.gap_id)) throw new Error(`record: close of unknown gap ${p.gap_id}`)
    const anchored = p.anchor_seat && p.anchor_tool && p.anchor_target
    if (!anchored && !p.carried_from) throw new Error('record: close requires the attestation anchor (--anchor-seat --anchor-tool --anchor-target) or --carried-from <round> — an unanchored closure is unauditable (E0.5a)')
    if (p.closure_class === 'closed_with_regression' && !p.successor) throw new Error('record: closed_with_regression requires --successor (lineage never drops)')
  }
  if (type === 'dispose' && !p.disposition) throw new Error('record: dispose requires --as minted-as|folded-into|declined|banked')
  if (type === 'regrade' && !p.basis) throw new Error('record: regrade requires --basis (grade movement is recorded with its reason)')
  if (type === 'opinion') {
    for (const f of ['gap_id', 'disposition', 'principle', 'tension', 'review_flag']) {
      if (p[f] === undefined) throw new Error(`record: opinion requires --${f.replace('_', '-')} (opinions, not dispositions)`)
    }
  }
}

function loadRegistry(runDir) {
  const seats = ['class-registry.json']
  for (const f of seats) {
    const p = join(recordsDir(runDir), f)
    if (existsSync(p)) { try { return JSON.parse(readFileSync(p, 'utf8')) } catch { return null } }
  }
  return null
}

function validateClass(runDir, p) {
  const reg = loadRegistry(runDir)
  const extensions = mergedEvents(runDir).events.filter((e) => e.type === 'class-new').map((e) => e.payload.slug)
  if (p.class_new) {
    if (!p.definition || !p.neighbor || !p.distinguisher) throw new Error('record: --class-new requires --definition, --neighbor <existing-slug>, and --distinguisher (the tie-break question)')
    const known = new Set([...(reg ? reg.classes.map((c) => c.slug) : []), ...extensions])
    if (reg && !known.has(p.neighbor)) throw new Error(`record: --neighbor ${p.neighbor} is not a known class`)
    return
  }
  if (!reg) return // no registry staged — advisory mode (R1 tolerance; R4 makes it strict)
  const known = new Set([...reg.classes.map((c) => c.slug), ...extensions])
  if (!known.has(p.class)) {
    const hint = reg.classes.slice(0, 6).map((c) => c.slug).join(', ')
    throw new Error(`record: unknown class ${p.class} — use a registry slug (e.g. ${hint}, ...) or mint the class with --class-new (definition + neighbor + distinguisher required)`)
  }
}

function allGapIds(runDir) {
  return new Set(mergedEvents(runDir).events.filter((e) => e.type === 'mint').map((e) => e.payload.gap_id))
}

// mintGapId: ids are tool-minted, sequential per round — the collision class dies.
export function mintGapId(runDir, round) {
  const existing = mergedEvents(runDir).events.filter((e) => e.type === 'mint' && e.round === round).length
  return `R${round}-${existing + 1}`
}

// Mint idempotency (plan §II, crash-retry safe): a seat whose message died after a
// successful mint retries the SAME command; with --key (the stable local label, e.g.
// the source lens finding), the retry returns the EXISTING id instead of double-minting.
export function existingMintByKey(runDir, seatId, key) {
  if (!key) return null
  const hit = mergedEvents(runDir).events.find((e) => e.type === 'mint' && e.seatId === seatId && e.payload.mint_key === key)
  return hit ? hit.payload.gap_id : null
}

// ---- Replay: shard discovery, winner selection, deterministic merge, dedup ----

export function mergedEvents(runDir) {
  const dir = recordsDir(runDir)
  if (!existsSync(dir)) return { events: [], anomalies: [] }
  const shardFiles = readdirSync(dir).filter((f) => f.startsWith('events-') && f.endsWith('.jsonl')).sort()
  const bySeat = new Map() // seatId -> [{nonce, file, events, mtime}]
  for (const f of shardFiles) {
    const m = /^events-(.+)-([0-9a-f]{8})\.jsonl$/.exec(f)
    if (!m) continue
    const [, seatId, nonce] = m
    const arr = bySeat.get(seatId) || []
    arr.push({ nonce, file: join(dir, f), events: readShard(join(dir, f)), mtime: statSync(join(dir, f)).mtimeMs })
    bySeat.set(seatId, arr)
  }
  const anomalies = []
  const winners = []
  for (const [seatId, shards] of bySeat) {
    if (shards.length === 1) { winners.push(...shards[0].events); continue }
    // Multi-nonce: winner = the nonce whose shard carries the seat's TERMINAL event
    // (verdict or revision — the last verb of a seat contract); neither-terminal
    // falls EXPLICITLY to latest mtime (plan-audit sitting-4 note 2).
    const terminal = shards.filter((s) => s.events.some((e) => e.type === 'verdict' || e.type === 'revision'))
    const pool = terminal.length ? terminal : shards
    const winner = pool.reduce((a, b) => (b.mtime > a.mtime ? b : a))
    anomalies.push(`multi-nonce seat ${seatId}: ${shards.length} dispatches (${shards.map((s) => s.nonce).join(', ')}); winner ${winner.nonce} by ${terminal.length ? 'terminal event' : 'mtime fallback'}`)
    winners.push(...winner.events)
  }
  // Deterministic global order + key-level dedup (first occurrence in merged order wins).
  winners.sort((a, b) => a.round - b.round || (a.seatId < b.seatId ? -1 : a.seatId > b.seatId ? 1 : 0) || a.seq - b.seq)
  const seen = new Set()
  const events = []
  for (const e of winners) {
    if (e.type !== 'register' && seen.has(e.key)) { anomalies.push(`duplicate key dedup'd: ${e.key} (nonce ${e.nonce})`); continue }
    seen.add(e.key)
    events.push(e)
  }
  return { events, anomalies }
}

// ---- Board state from events ----

export function boardState(runDir) {
  const { events, anomalies } = mergedEvents(runDir)
  const gaps = new Map() // id -> {mint payload, closed?, closure, regrades: []}
  const observations = [] // {seatId, key, payload, disposition?}
  for (const e of events) {
    if (e.type === 'mint') gaps.set(e.payload.gap_id, { ...e.payload, round: e.round, open: true, regrades: [] })
    if (e.type === 'regrade') { const g = gaps.get(e.payload.gap_id); if (g) { Object.assign(g, pickGrades(e.payload)); g.regrades.push(e.payload) } }
    if (e.type === 'close') { const g = gaps.get(e.payload.gap_id); if (g) { g.open = false; g.closure = e.payload; g.closedRound = e.round } }
    if (e.type === 'finding' || e.type === 'observe') observations.push({ seatId: e.seatId, key: e.key, kind: e.type, payload: e.payload })
    if (e.type === 'dispose') { const o = observations.find((x) => (x.payload.label || x.key) === e.payload.observation); if (o) o.disposition = e.payload }
  }
  return { events, anomalies, gaps, observations }
}

const pickGrades = (p) => Object.fromEntries(Object.entries(p).filter(([k]) => ['severity', 'likelihood', 'impact', 'complexity_cost'].includes(k)))

// ---- Projections (atomic writes: temp + rename — sitting-4 note 1) ----

function writeAtomic(path, content) {
  const tmp = path + '.tmp-' + randomBytes(3).toString('hex')
  writeFileSync(tmp, content)
  renameSync(tmp, path)
}

export function render(runDir, { outDir = null } = {}) {
  // Shadow by default during dual-mode (plan R2); real paths are opt-in at R3.
  const out = outDir || join(recordsDir(runDir), 'render-shadow')
  mkdirSync(out, { recursive: true })
  const { anomalies, gaps, observations } = boardState(runDir)
  const open = [...gaps.values()].filter((g) => g.open)
  const closed = [...gaps.values()].filter((g) => !g.open)

  const anomalyFooter = anomalies.length ? `\n## render anomalies (never silently normalized)\n\n${anomalies.map((a) => `- ${a}`).join('\n')}\n` : ''
  const undisposed = observations.filter((o) => !o.disposition)
  const notesFooter = undisposed.length ? `\n## undisposed lens observations (every observation demands a merge disposition)\n\n${undisposed.map((o) => `- ${o.seatId} ${o.payload.label || o.key}: ${(o.payload.text || '').slice(0, 120)}`).join('\n')}\n` : ''

  const ledger = [
    `# red/ledger.md — RENDERED PROJECTION (source of truth: records/ event log; do not hand-edit)`,
    '',
    `## OPEN GAPS (${open.length})`,
    '',
    ...open.map((g) => `### ${g.gap_id} — ${g.problem ? g.problem.slice(0, 100) : ''}\n${g.location || ''}\nseverity ${g.severity} | ${g.likelihood} x ${g.impact} | cx ${g.complexity_cost} | class ${g.class}${g.supersedes && g.supersedes.length ? ` | supersedes ${g.supersedes.join(' ')}` : ''}${g.found_by && g.found_by.length ? ` | found_by ${g.found_by.join(',')}` : ''}${g.regrades.length ? `\nregraded x${g.regrades.length} (history in the event log; latest basis: ${g.regrades[g.regrades.length - 1].basis})` : ''}\nrequired_fix: ${g.required_fix || ''}\nacceptance_check: ${g.acceptance_check}\n`),
    '## CLOSURE INDEX',
    '',
    ...closed.map((g) => `${g.gap_id} | ${g.closure.closure_class || 'closed'} | ${(g.problem || '').slice(0, 60)} | ${g.closure.successor || '-'}`),
    anomalyFooter,
    notesFooter,
  ].join('\n')
  writeAtomic(join(out, 'ledger.md'), ledger)

  const archive = [
    `# red/archive.md — RENDERED PROJECTION (append-only by construction in the event log)`,
    '',
    ...closed.map((g) => `## ${g.gap_id} — ${g.closure.closure_class || 'closed'}\n${(g.problem || '')}\nverification anchor: ${g.closure.carried_from ? `CARRIED from round ${g.closure.carried_from}` : `${g.closure.anchor_seat} | ${g.closure.anchor_tool} | ${g.closure.anchor_target}`}${g.closure.successor ? `\nsuccessor: ${g.closure.successor}` : ''}\n`),
  ].join('\n')
  writeAtomic(join(out, 'archive.md'), archive)

  // Telemetry: one line per round that has red activity — computed, never self-reported.
  const rounds = [...new Set([...gaps.values()].map((g) => g.round))].sort((a, b) => a - b)
  const lines = []
  for (const r of rounds) {
    const openAtR = [...gaps.values()].filter((g) => g.round <= r && (g.open || (g.closedRound ?? 99) > r))
    const minted = [...gaps.values()].filter((g) => g.round === r)
    const closedAtR = [...gaps.values()].filter((g) => g.closedRound === r)
    const lineage = minted.filter((g) => (g.supersedes || []).length)
    let down = 0, up = 0
    for (const g of lineage) for (const anc of g.supersedes || []) {
      const a = gaps.get(anc); if (!a) continue
      const d = gapMass(g) - gapMass(a)
      if (d < 0) down += -d; else up += d
    }
    const bySev = {}
    for (const g of minted) bySev[g.severity] = (bySev[g.severity] || 0) + 1
    lines.push(JSON.stringify({
      round: r, mapping_version: MASS_MAPPING_VERSION,
      open_count: openAtR.length,
      max_severity: openAtR.map((g) => g.severity).sort((a, b) => (MASS[b] ?? 0) - (MASS[a] ?? 0))[0] || null,
      new_mint: { count: minted.length, by_severity: bySev },
      mass: openAtR.reduce((s, g) => s + gapMass(g), 0),
      repair_regression: { closures: closedAtR.length, lineage_mints: lineage.length, ratio: closedAtR.length ? +(lineage.length / closedAtR.length).toFixed(2) : null },
      edge_deltas: { down_mass: +down.toFixed(2), up_mass: +up.toFixed(2) },
    }))
  }
  writeAtomic(join(out, 'board-telemetry.jsonl'), lines.join('\n') + (lines.length ? '\n' : ''))
  return { out, anomalies: anomalies.length, open: open.length, closed: closed.length }
}

// ---- Shared CLI plumbing for the per-seat tools ----

export function parseArgs(argv) {
  const out = { _: [] }
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]
    if (a.startsWith('--')) {
      const k = a.slice(2)
      const next = argv[i + 1]
      if (next === undefined || next.startsWith('--')) out[k] = true
      else { out[k] = next; i++ }
    } else out._.push(a)
  }
  return out
}

export function payloadText(args) {
  if (args.file) return readFileSync(args.file, 'utf8')
  if (args.text) return String(args.text)
  return ''
}

// runVerb: the CLI entry shared by all seat tools. `verbs` maps verb -> handler;
// the verb SET is the role boundary — a missing verb is a missing capability.
export function runVerb(toolName, verbs, argv) {
  const [verb, ...rest] = argv
  const args = parseArgs(rest)
  if (!verb || verb === '--help' || verb === 'help') {
    console.log(`${toolName} — record contract (the verb set IS the role boundary)\nverbs: ${Object.keys(verbs).join(', ')}, render, help`)
    for (const [v, h] of Object.entries(verbs)) console.log(`  ${v}: ${h.help}`)
    console.log(`  render: read-only projection refresh (any seat may invoke)`)
    return 0
  }
  const runDir = args.run
  if (!runDir) { console.error(`${toolName}: --run <runDir> is required`); return 2 }
  if (verb === 'render') { const r = render(runDir); console.log(`${toolName}: rendered to ${r.out} (${r.open} open, ${r.closed} closed${r.anomalies ? `, ${r.anomalies} anomalies` : ''})`); return 0 }
  const h = verbs[verb]
  if (!h) { console.error(`${toolName}: verb '${verb}' is outside this seat's role (available: ${Object.keys(verbs).join(', ')}, render)`); return 2 }
  const seatId = args['seat-id']
  if (!seatId) { console.error(`${toolName}: --seat-id is required (the engine assigns it; it is in your prompt)`); return 2 }
  try {
    const result = h.fn(runDir, seatId, args)
    // Render-on-mutation (plan §II render locus): projections current after every write.
    if (verb !== 'register') render(runDir)
    if (result) console.log(result)
    return 0
  } catch (e) {
    console.error(`${toolName}: ${e.message}`)
    return 2
  }
}

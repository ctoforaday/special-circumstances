// scorecards.mjs — every constitutional clause gets a number the seat can SEE.
//
// The premise (plans/scorecards.md): a seat improves on metrics it is actually
// measured against, so telos and loss conditions alike must be numbers that reach
// the dashboard AND the seat. Proven negatively on ourselves — "confidence
// self-graded" was mandated, uninstrumented, invisible, and practised five times
// in 1,892 lines. A clause without an instrument and a feedback path is a dead
// letter by construction.
//
// This module is the COMPUTE half. It reads only git-tracked run artifacts
// (board telemetry, the journal's envelopes, debate.md, the ledger) and returns
// rows. Capture writes them to feov-memory/<chair>-scorecard.md; setup mirrors
// them back into the next run's inputs/; the engine injects each chair's headline
// numbers into that chair's prompts.
//
// TWO RULES THIS FILE EXISTS TO ENFORCE:
//
//  1. EVERY ROW CARRIES ITS CLASS. benchmark (optimize me), diagnostic (explains
//     you — optimizing it is a defect), detector (a loss-condition tripwire, any
//     nonzero is a finding). A number handed to a seat without its class is an
//     invitation to game it; red "improving" grade stability is stubbornness, not
//     rigour.
//  2. AN UNCOMPUTABLE ROW IS STILL A ROW. It appears with its reason instead of
//     a value. Emitting only the easy rows would let the scorecard read complete
//     while the hard clauses stayed exactly as unmeasured as before — which is
//     the failure this whole exercise is a response to.
import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { join } from 'node:path'

const read = (p) => (existsSync(p) ? readFileSync(p, 'utf8') : null)

export const CLASSES = ['benchmark', 'diagnostic', 'detector', 'measure']

// row builds one scorecard entry. `value` null means not computed, and `note`
// must then say why — the constructor enforces the pairing rather than trusting
// each call site to remember it.
export function row({ clause, metric, cls, value = null, note = '', joint = '' }) {
  if (!CLASSES.includes(cls)) throw new Error(`scorecards: unknown class ${cls} for ${metric}`)
  if (value === null && !note) throw new Error(`scorecards: ${metric} has no value and no reason — an uncomputed row must say why`)
  return { clause, metric, cls, value, note, joint }
}

export function readTelemetry(runDir) {
  const t = read(join(runDir, 'trajectories', 'board-telemetry.jsonl'))
  if (!t) return []
  return t.split('\n').filter(Boolean).map((l) => { try { return JSON.parse(l) } catch { return null } }).filter(Boolean)
}

// ---- BLUE ----

export function blueRows(runDir, results, telemetry) {
  const rows = []

  // repair_regression_ratio: lineage mints per closure. Above ~0.5 means repairs
  // are minting successors as fast as they close gaps — churn wearing the shape
  // of progress. Baseline across runs 3-5: 0.37-0.72.
  const ratios = telemetry.map((t) => t.repair_regression && t.repair_regression.ratio).filter((r) => typeof r === 'number')
  rows.push(ratios.length
    ? row({
      clause: 'Durable repairs', metric: 'repair_regression_ratio', cls: 'benchmark',
      value: +(ratios.reduce((a, b) => a + b, 0) / ratios.length).toFixed(2),
      joint: 'reads WITH red rigour: a low ratio under a lax adversary means nothing',
    })
    : row({ clause: 'Durable repairs', metric: 'repair_regression_ratio', cls: 'benchmark', note: 'no telemetry rounds with closures' }))

  // Manifest coverage: repaired gaps that came back with a manifest row. An
  // unmanifested repair is unaudited by blue's OWN standard.
  const blue = results.filter((r) => Array.isArray(r.manifest))
  const manifested = blue.reduce((n, r) => n + r.manifest.length, 0)
  const repaired = blue.reduce((n, r) => n + (Array.isArray(r.repaired_gaps) ? r.repaired_gaps.length : 0), 0)
  rows.push(repaired
    ? row({
      clause: 'Correctness manifest', metric: 'manifest_coverage', cls: 'benchmark',
      value: +(manifested / repaired).toFixed(2),
    })
    : row({
      clause: 'Correctness manifest', metric: 'manifest_coverage', cls: 'benchmark',
      value: manifested, note: 'manifest rows counted; envelopes do not report a repaired-gap denominator, so this is a COUNT not a ratio',
    }))

  // Round-parity: a revision that is not on the record did not happen.
  const attested = results.filter((r) => r.round_record_appended === true).length
  const claimed = results.filter((r) => 'round_record_appended' in r).length
  rows.push(row({
    clause: 'Round on the record', metric: 'round_parity_failures', cls: 'detector',
    value: claimed - attested,
    note: claimed ? '' : 'no envelope carried the attestation field',
  }))

  // The additive invariant, now that it is claims-level rather than prose-level.
  // Prose may compact freely; a claim leaves only through a retire record. So an
  // unaccounted FALL in claim_count is the whole enforcement — arithmetic, not
  // judgement, and detectable without reading a word of the report.
  const counts = results.filter((r) => typeof r.claim_count === 'number').map((r) => r.claim_count)
  const retires = results.reduce((n, r) => n + (Array.isArray(r.retired) ? r.retired.length : 0), 0)
  let drop = 0
  for (let i = 1; i < counts.length; i++) if (counts[i] < counts[i - 1]) drop += counts[i - 1] - counts[i]
  rows.push(counts.length > 1
    ? row({
      clause: 'LOSS: additive violations', metric: 'unrecorded_claim_loss', cls: 'detector',
      value: Math.max(0, drop - retires),
      note: `${drop} claim(s) lost across rounds, ${retires} retired on the record`,
      joint: 'a fall the retire events do not account for is substance leaving silently — the failure the old prose-level rule was written to stop',
    })
    : row({
      clause: 'LOSS: additive violations', metric: 'unrecorded_claim_loss', cls: 'detector',
      note: 'needs at least two rounds reporting claim_count',
    }))

  rows.push(row({
    clause: 'Calibration is craft', metric: 'confidence_vs_survival', cls: 'benchmark',
    note: 'BLOCKED until per-claim confidence records exist (W2f) — calibration cannot be computed from prose',
  }))
  return rows
}

// ---- RED ----

export function redRows(runDir, results, telemetry) {
  const rows = []
  const redEnvs = results.filter((r) => Array.isArray(r.gaps))

  // Attestation-format invariant: closures carrying seat+tool+target anchors, or
  // an explicit carried-from. E0.5a measured ~11% of the record mechanically
  // unauditable BY FORMAT while the behaviour behind it was honest.
  const archive = read(join(runDir, 'red', 'archive.md')) || ''
  const blocks = archive.split(/^## /m).slice(1)
  const anchored = blocks.filter((b) => /verification anchor:/i.test(b) && !/verification anchor:\s*(undefined|\s*\|\s*undefined)/i.test(b)).length
  rows.push(blocks.length
    ? row({
      clause: 'Attestation-format invariant', metric: 'anchored_closures_pct', cls: 'benchmark',
      value: Math.round((anchored / blocks.length) * 100),
      note: 'target 100; baseline 89 (E0.5a)',
    })
    : row({ clause: 'Attestation-format invariant', metric: 'anchored_closures_pct', cls: 'benchmark', note: 'no archive records this run' }))

  // Never-hard-fail detector: rounds that FAILed on fallout-only mints below the
  // mass threshold. Visibility only — the verdict stays red's.
  const softFails = telemetry.filter((t) => t.convergence_vs_verdict_flag).length
  rows.push(row({
    clause: 'Never-hard-fail', metric: 'convergence_vs_verdict_flags', cls: 'detector',
    value: softFails,
  }))

  // W2i's assumption under test. E0.5c measured citation yield collapsing 70-80%
  // after round 1 while L5/L6 held flat, and the round-graduated dispatch was
  // sized on that. If later-round citation yield stops collapsing, the cap comes
  // off — so the number that would tell us has to be in front of someone.
  const byRole = citationYieldByRole(runDir)
  rows.push(byRole
    ? row({
      clause: 'Lens economics (W2i assumption)', metric: 'citation_yield_by_round', cls: 'diagnostic',
      value: byRole,
      joint: 'RETUNE TRIGGER: if rounds 2+ citation yield stops collapsing versus round 1, the consolidation cap is wrong',
    })
    : row({
      clause: 'Lens economics (W2i assumption)', metric: 'citation_yield_by_round', cls: 'diagnostic',
      note: 'no per-lens candidate files found — attribution needs red/candidates/round-N-lens-M.md',
    }))

  rows.push(row({
    clause: 'Certification: earned PASS/FAIL', metric: 'finding_precision', cls: 'benchmark',
    note: 'needs adjudication outcomes per finding; the judge ruled on <5% of gaps in runs 4-5, so the denominator is not yet meaningful',
  }))
  return rows
}

// citationYieldByRole counts findings per lens ROLE per round from the candidate
// files. Role, not position: W2i pinned L1-L4 as citation slices, L5 logic, L6
// dark-side, precisely so this comparison survives a round dispatching fewer
// seats.
export function citationYieldByRole(runDir) {
  const dir = join(runDir, 'red', 'candidates')
  if (!existsSync(dir)) return null
  const perRound = {}
  for (const f of readdirSync(dir)) {
    const m = /^round-(\d+)-lens-(\d+)\.md$/.exec(f)
    if (!m) continue
    const [, round, role] = m
    const body = read(join(dir, f)) || ''
    const findings = (body.match(/^\s*(?:###\s*)?L\d+-F\d+\b/gm) || []).length
    const bucket = (perRound[round] ||= { citation: 0, logic: 0, darkside: 0 })
    if (+role <= 4) bucket.citation += findings
    else if (+role === 5) bucket.logic += findings
    else bucket.darkside += findings
  }
  return Object.keys(perRound).length ? perRound : null
}

// ---- BENCH ----

export function benchRows(runDir, results) {
  const rows = []
  const rulings = results.flatMap((r) => (Array.isArray(r.resolutions) ? r.resolutions : []))

  // Not a router. 64 of 65 rulings were `carried` across runs 4-5 — a judge that
  // only defers is a router with robes, and this is the before/after scoreboard.
  const carried = rulings.filter((r) => r.resolution === 'carried').length
  rows.push(rulings.length
    ? row({
      clause: 'Not a router', metric: 'carried_share', cls: 'benchmark',
      value: +(carried / rulings.length).toFixed(2),
      note: `${carried}/${rulings.length}; baseline 76/77`,
    })
    : row({ clause: 'Not a router', metric: 'carried_share', cls: 'benchmark', note: 'the bench did not sit this run' }))

  // Direction-uptake is the bench HEADLINE (E0.5d): in-run reversal is ~0 by
  // traffic-class construction and measures nothing, so what counts is whether
  // blue acted on the direction a carried ruling stated.
  const debate = read(join(runDir, 'debate.md')) || ''
  const leadSections = (debate.match(/^### LEAD/gm) || []).length
  const blueCitesLead = (debate.match(/^### BLUE[\s\S]*?(?=^### |\Z)/gm) || [])
    .filter((s) => /lead|judge|direction|carried/i.test(s)).length
  rows.push(leadSections
    ? row({
      clause: 'Direction-uptake (headline)', metric: 'blue_sections_citing_direction', cls: 'benchmark',
      value: `${blueCitesLead}/${leadSections}`,
      note: 'textual proxy: blue sections referencing the bench after a LEAD section; baseline ~100%',
    })
    : row({ clause: 'Direction-uptake (headline)', metric: 'blue_sections_citing_direction', cls: 'benchmark', note: 'no LEAD sections this run' }))

  // Opinion form: a ruling that states no principle is a disposition wearing a
  // ruling's name.
  const opinionated = rulings.filter((r) => r.principle && r.tension && r.review_flag).length
  rows.push(rulings.length
    ? row({
      clause: 'Opinion form', metric: 'rulings_without_opinion', cls: 'detector',
      value: rulings.length - opinionated,
    })
    : row({ clause: 'Opinion form', metric: 'rulings_without_opinion', cls: 'detector', note: 'no rulings this run' }))

  const petitions = results.flatMap((r) => (Array.isArray(r.petitions) ? r.petitions : []))
  rows.push(row({ clause: 'Petition handling', metric: 'petitions_filed', cls: 'measure', value: petitions.length }))
  return rows
}

// ---- assembly ----

export function computeScorecards(runDir, results) {
  const telemetry = readTelemetry(runDir)
  return {
    blue: blueRows(runDir, results, telemetry),
    red: redRows(runDir, results, telemetry),
    bench: benchRows(runDir, results),
  }
}

const CLASS_NOTE = {
  benchmark: 'BENCHMARK — optimize this',
  diagnostic: 'DIAGNOSTIC — this explains you; optimizing it is a defect',
  detector: 'DETECTOR — a loss-condition tripwire; any nonzero is a finding',
  measure: 'MEASURE — recorded, not targeted',
}

// renderChair produces the STATS block appended to feov-memory/<chair>-scorecard.md.
// Appended, never overwritten: the run-over-run SERIES is the point, since a
// single run's number says nothing about whether a chair is improving.
export function renderChair(chair, rows, runLabel) {
  const out = [`## ${runLabel}`, '']
  for (const r of rows) {
    const v = r.value === null
      ? `_not computed_ — ${r.note}`
      : `**${typeof r.value === 'object' ? JSON.stringify(r.value) : r.value}**${r.note ? ` (${r.note})` : ''}`
    out.push(`- \`${r.metric}\` [${r.cls}] — ${r.clause}: ${v}`)
    if (r.joint) out.push(`  - ${r.joint}`)
  }
  out.push('')
  return out.join('\n')
}

// headline picks the few numbers a seat actually sees in its prompt. A scorecard
// nobody reads is the same dead letter as a clause nobody measures, but a wall of
// rows in every prompt is how a seat learns to skip the section.
export function headline(rows, max = 3) {
  const computed = rows.filter((r) => r.value !== null)
  const ordered = [
    ...computed.filter((r) => r.cls === 'detector' && r.value !== 0),
    ...computed.filter((r) => r.cls === 'benchmark'),
    ...computed.filter((r) => r.cls === 'diagnostic'),
  ]
  return ordered.slice(0, max).map((r) => {
    const v = typeof r.value === 'object' ? JSON.stringify(r.value) : r.value
    return `${r.metric} ${v} [${CLASS_NOTE[r.cls].split(' —')[0]}]`
  })
}

export function chairHeader(chair) {
  return [
    `# ${chair} scorecard — the numbers this chair is measured against`,
    '',
    'Computed at capture from git-tracked run artifacts, appended run over run.',
    'Setup mirrors this file into the next run\'s inputs/, and the engine puts the',
    'headline numbers into this chair\'s seat prompts — the visibility loop, without',
    'which a clause is measured and still invisible to the seat it governs.',
    '',
    'CLASSES: ' + Object.values(CLASS_NOTE).join(' · '),
    '',
  ].join('\n')
}

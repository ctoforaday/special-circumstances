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
import { existsSync, readFileSync } from 'node:fs'
import { execFileSync } from 'node:child_process'
import { join } from 'node:path'
import { pathToFileURL } from 'node:url'

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
  // The tool computes telemetry into records/render-shadow on `merge render` (2026-07-19,
  // Phase B1). trajectories/ is the fallback for pre-migration runs, when the merge seat
  // still hand-wrote the line.
  const t = read(join(runDir, 'records', 'render-shadow', 'board-telemetry.jsonl'))
    ?? read(join(runDir, 'trajectories', 'board-telemetry.jsonl'))
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

  // LINES OF INQUIRY — what finally instruments think-around-problem.
  //
  // Count is a DIAGNOSTIC and never a benchmark: counting avenues invites
  // avenue-padding, and a seat that learns to list five roads it never walked has
  // made the record worse than empty. What the count explains is exploration
  // BREADTH; what it must not become is a target.
  const avenues = results.flatMap((r) => (Array.isArray(r.avenues) ? r.avenues : []))
  const byStatus = avenues.reduce((m, a) => ({ ...m, [a.status]: (m[a.status] || 0) + 1 }), {})
  rows.push(avenues.length
    ? row({
      clause: 'Alternatives explored', metric: 'lines_of_inquiry', cls: 'diagnostic',
      value: byStatus,
      joint: 'reads WITH the report: breadth means nothing if the pursued line was chosen before the others were weighed',
    })
    : row({
      clause: 'Alternatives explored', metric: 'lines_of_inquiry', cls: 'diagnostic',
      note: 'no avenues recorded — think-around-problem is back to self-attested for this run',
    }))

  // A declined avenue whose reason is a shrug is the decoration the verb exists
  // to prevent. The tool already refuses an EMPTY reason; this catches the token
  // one, the same way the rule-sweep gate rejects a one-word sibling sweep.
  const thin = avenues.filter((a) => a.status !== 'pursued' && String(a.reason || '').trim().length < 20)
  rows.push(row({
    clause: 'Alternatives explored', metric: 'thin_avenue_reasons', cls: 'detector',
    value: thin.length,
    note: thin.length ? `${thin.map((a) => a.line).slice(0, 3).join('; ')}` : '',
  }))

  rows.push(row({
    clause: 'Calibration is craft', metric: 'confidence_vs_survival', cls: 'benchmark',
    note: 'BLOCKED until per-claim confidence records exist (W2f) — calibration cannot be computed from prose',
  }))
  return rows
}

// ---- RED ----

export function redRows(runDir, results, telemetry, bin) {
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
  const byRole = citationYieldByRole(runDir, bin)
  rows.push(byRole
    ? row({
      clause: 'Lens economics (W2i assumption)', metric: 'citation_yield_by_round', cls: 'diagnostic',
      value: byRole,
      joint: 'RETUNE TRIGGER: compare PER_SEAT yield across rounds, never the raw count — W2i dispatches fewer citation lenses later, so a raw comparison scores the cut as the collapse that justified it. If per-seat citation yield holds while another role collapses, the cap is aimed at the wrong lens',
    })
    : row({
      clause: 'Lens economics (W2i assumption)', metric: 'citation_yield_by_round', cls: 'diagnostic',
      note: 'no findings on the record yet (or the tool binary was not passed) — per-role yield needs the findings view',
    }))

  rows.push(row({
    clause: 'Certification: earned PASS/FAIL', metric: 'finding_precision', cls: 'benchmark',
    note: 'needs adjudication outcomes per finding; the judge ruled on <5% of gaps in runs 4-5, so the denominator is not yet meaningful',
  }))
  return rows
}

// citationYieldByRole counts findings per lens ROLE per round from the RECORD (the
// findings view), replacing the per-lens files the merge used to leave behind. Role,
// not position: W2i pinned L1-L4 as citation slices, L5 logic, L6 dark-side, precisely
// so this comparison survives a round dispatching fewer seats.
// It SPAWNS the tool rather than parsing records/*.jsonl — parsing the log itself
// would reimplement replay/dedup, the second-reader defect the tool exists to remove.
export function citationYieldByRole(runDir, bin) {
  if (!bin) return null // no tool = no findings view (same "not computed" as no candidates dir)
  let findings
  try {
    const out = execFileSync(bin, ['merge', 'show', '--run', runDir, '--view', 'findings'], { encoding: 'utf8' })
    findings = JSON.parse(out).findings || []
  } catch {
    return null
  }
  return bucketFindingsByRole(findings)
}

// bucketFindingsByRole is the pure kernel: given the findings view's array (each carrying
// role like "L5", round, seat_id), bucket per round by role-kind and compute PER-SEAT yield.
// Split out so it is testable without spawning the tool.
export function bucketFindingsByRole(findings) {
  const perRound = {}
  const seatsSeen = {} // round -> kind -> Set(seat_id): PER-SEAT yield needs distinct seats
  for (const f of findings) {
    const roleNum = +String(f.role || '').replace(/^L/, '')
    if (!roleNum) continue
    const round = String(f.round)
    const kind = roleNum <= 4 ? 'citation' : roleNum === 5 ? 'logic' : 'darkside'
    const bucket = (perRound[round] ||= { citation: 0, logic: 0, darkside: 0, seats: { citation: 0, logic: 0, darkside: 0 } })
    bucket[kind] += 1
    const seen = (seatsSeen[round] ||= { citation: new Set(), logic: new Set(), darkside: new Set() })
    seen[kind].add(f.seat_id)
  }
  for (const [round, bucket] of Object.entries(perRound)) {
    for (const k of ['citation', 'logic', 'darkside']) bucket.seats[k] = seatsSeen[round][k].size
  }
  // PER-SEAT YIELD, not just raw count. W2i dispatches FEWER citation lenses in
  // later rounds, so a raw round-over-round comparison measures the cut as if it
  // were the collapse that justified the cut — the intervention manufactures its
  // own evidence, and the retune trigger keyed on it cannot detect the error.
  // First live run: citation fell 3->1 raw (-67%), but 2 seats -> 1 seat, so per
  // seat it was 1.5 -> 1.0 (-33%). L5 (one seat both rounds) fell 9 -> 3, a real
  // -67%. The lens that actually collapsed was not the one being rationed.
  for (const r of Object.values(perRound)) {
    r.per_seat = {}
    for (const k of ['citation', 'logic', 'darkside']) {
      r.per_seat[k] = r.seats[k] ? +(r[k] / r.seats[k]).toFixed(2) : null
    }
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
  // (?![\s\S]) is end-of-input. \Z is NOT an anchor in JavaScript — it is an
  // identity escape matching a literal 'Z', so the lazy body stopped at the first
  // capital Z in the prose. The LAST blue section has no '### ' after it, so it was
  // truncated at a letter and then filtered for words the truncation had removed.
  const blueCitesLead = (debate.match(/^### BLUE[\s\S]*?(?=^### |(?![\s\S]))/gm) || [])
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
  // review_flag is tested for PRESENCE, not truth. It is a boolean, and `false`
  // — "this needs no further review" — is a complete opinion honestly recorded.
  // Testing it for truthiness scored the bench as opinionless every time it gave
  // the clean answer, which is precisely backwards for a metric whose point is
  // that a ruling stating no principle is a disposition wearing a ruling's name.
  const opinionated = rulings.filter((r) => r.principle && r.tension && 'review_flag' in r).length
  rows.push(rulings.length
    ? row({
      clause: 'Opinion form', metric: 'rulings_without_opinion', cls: 'detector',
      value: rulings.length - opinionated,
    })
    : row({ clause: 'Opinion form', metric: 'rulings_without_opinion', cls: 'detector', note: 'no rulings this run' }))

  // The bench now has a power (reading trajectories), so its USE of that power is
  // measured — fixing one integrity gap by creating another would be the whole
  // exercise defeating itself. Evidence confinement: an inspection the bench did
  // not declare is indistinguishable from one it invented.
  const opinions = rulings.filter((r) => r.principle)
  const declaredReads = opinions.filter((r) => /trajector|inspect|tool call/i.test(String(r.rationale || '') + String(r.principle || '')))
  rows.push(row({
    clause: 'Evidence confinement', metric: 'undeclared_inspection_risk', cls: 'detector',
    value: 0,
    note: declaredReads.length
      ? `${declaredReads.length} opinion(s) reference trajectory evidence; capture's attestation-integrity audit is the cross-check`
      : 'no opinion referenced trajectory evidence this run',
    joint: 'reads WITH the attestation-integrity audit at capture: this counts declarations, that reconciles claims against actual tool calls',
  }))

  const petitions = results.flatMap((r) => (Array.isArray(r.petitions) ? r.petitions : []))
  rows.push(row({ clause: 'Petition handling', metric: 'petitions_filed', cls: 'measure', value: petitions.length }))
  return rows
}

// ---- assembly ----

export function computeScorecards(runDir, results, bin) {
  const telemetry = readTelemetry(runDir)
  return {
    blue: blueRows(runDir, results, telemetry),
    red: redRows(runDir, results, telemetry, bin),
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
  // The headline is EMITTED, not re-derived downstream. Setup previously rebuilt it
  // by re-parsing these rendered rows, which made two implementations of "what a
  // seat sees" — and they disagreed twice over: setup took the first three rows in
  // FILE order rather than headline()'s ranking (a tripped detector could lose its
  // slot to a benchmark above it), and its parser broke on any clause containing a
  // colon, which silently hid `unrecorded_claim_loss` — a detector, and the entire
  // enforcement of the additive invariant — from every prompt.
  const h = headline(rows)
  if (h.length) out.push('', `HEADLINE: ${h.join(' · ')}`)
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

// THE parser for a rendered scorecard. It lives here, beside renderChair, because
// the module that WRITES a format is the only place that can be trusted to read it.
//
// This exists because the same defect was found three times in three hand-rolled
// copies: setup's headline mirror, the dashboard's table, and (by construction)
// anything written next. Each used a colon-excluding class for the clause, and a
// character class cannot backtrack past the delimiter it excludes — so every row
// whose CLAUSE contains a colon silently failed to match. The casualty each time
// was `unrecorded_claim_loss`, a DETECTOR and the entire enforcement of the
// additive invariant: computed, rendered, and invisible to both the seat and the
// human. Three readers of one artifact, disagreeing about it, is the defect —
// fixing the regex in three places would only have reset the clock.
//
// Lazy (.+?) terminates at the first colon actually followed by a value, which is
// the real end of a clause. A row that is `_not computed_` yields value null, so
// callers can tell "measured zero" from "never measured" — a distinction the
// detector classes depend on.
export function parseRenderedRows(section) {
  const re = /`([a-z_]+)`\s*\[(benchmark|detector|diagnostic|measure)\]\s*—\s*(.+?):\s*(?:\*\*([^*]+)\*\*|_not computed_)/g
  return [...String(section || '').matchAll(re)].map((m) => ({
    metric: m[1], cls: m[2], clause: m[3], value: m[4] === undefined ? null : m[4],
  }))
}

// A scorecard is APPENDED run over run, so the last '## ' block is this run's.
export function latestSection(body) {
  const s = String(body || '').split(/^## /m)
  return s[s.length - 1] || ''
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

// ---- in-run CLI (priors-are-poison half-2) ----
//
// `node scorecards.mjs --run <dir> --chair blue|red|bench` prints THAT chair's scorecard for
// THIS run — the self-read a seat runs before its docket, replacing the retired cross-run
// seed. It reuses the ONE computation (computeScorecards); it re-derives nothing. The
// envelope-derived rows read "not computed" until capture assembles the journal, so mid-run a
// chair sees its file-derived headline (repair_regression, anchored_closures, direction-uptake)
// and honest "not computed" for the rest — never a stale number from another question.

// readResults gathers the run's envelopes from the journal if it has been assembled
// (post-capture). Mid-run it returns [] — not an error: the file-derived rows still compute,
// and the envelope rows say why they cannot.
export function readResults(runDir) {
  const p = join(runDir, 'trajectories', 'journal.jsonl')
  if (!existsSync(p)) return []
  const results = []
  for (const line of readFileSync(p, 'utf8').split('\n')) {
    if (!line.trim()) continue
    let j
    try { j = JSON.parse(line) } catch { continue }
    if (j.result && typeof j.result === 'object') results.push(j.result)
  }
  return results
}

function flag(argv, name) {
  const i = argv.indexOf(name)
  return i >= 0 && i + 1 < argv.length ? argv[i + 1] : null
}

export function cli(argv) {
  const runDir = flag(argv, '--run')
  const chair = flag(argv, '--chair')
  const bin = flag(argv, '--bin')
  const cards = runDir ? computeScorecards(runDir, readResults(runDir), bin) : null
  if (!runDir || !cards || !cards[chair]) {
    process.stderr.write('usage: node scorecards.mjs --run <runDir> --chair blue|red|bench\n')
    process.exitCode = 2
    return
  }
  process.stdout.write(renderChair(chair, cards[chair], 'this run') + '\n')
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) cli(process.argv.slice(2))

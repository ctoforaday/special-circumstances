// Model-tier guard (#111): compare each seat's ACTUAL price-tier to the tier its CLASS was
// CONFIGURED to run on. debate.js now requires both tiers explicitly (no default, no session
// inheritance), so a live run's run-config always carries them — which is what makes this guard
// possible in both directions:
//   - a seat DEARER than its configured tier is the fable trap (a bulk seat that escaped onto an
//     expensive model) -> FAIL;
//   - a seat CHEAPER than configured is the discounted-verification symptom (a strong tier was
//     asked for but a cheap model ran) -> WARN.
// Pure functions, no I/O — unit-testable for zero tokens. The price table and tier() come from
// cost-audit.mjs and the seat->class map from seat-classify.mjs; neither is re-encoded here (one
// source per fact, or the copies drift).
import { PRICES, tier } from './cost-audit.mjs'
import { classOf } from './seat-classify.mjs'

// Cost rank from the ONE price table's input-price column (haiku 1 < sonnet 2 < opus 5 < fable 10).
// An unrecognized key ranks dearest, so an anomaly over-reports rather than flatters — the same
// direction tier() already errs.
const rank = (t) => (PRICES[t] ? PRICES[t][0] : Infinity)
const dearer = (a, b) => rank(a) > rank(b)

// tierMismatch(seatRows, config) -> findings[]
//   seatRows: [{ seat, round, t }]     t = actual price-tier (cost-audit's scanTranscript emits it)
//   config:   { model, judgmentModel } the configured tier per class (bulk / judgment)
// One finding per non-PASS seat; verdict is 'FAIL' or 'WARN'. Equal tiers are PASS (omitted).
export function tierMismatch(seatRows = [], config = {}) {
  const out = []
  for (const row of seatRows) {
    const cls = classOf(row.seat)
    if (!cls) continue // other/unknown seat — not tier-bound
    const configured = cls === 'bulk' ? config.model : config.judgmentModel
    const actual = row.t
    if (!configured) {
      // Should not happen on a run launched after #111 (both tiers required), but a pre-#111
      // run-config or a malformed one can omit it — degrade honestly rather than skip silently.
      out.push({ seat: row.seat, round: row.round, cls, actual, expected: null, verdict: 'WARN',
        why: `${cls} tier not declared in run-config — cannot judge ${row.seat}` })
      continue
    }
    const expected = tier(configured)
    if (dearer(actual, expected)) {
      out.push({ seat: row.seat, round: row.round, cls, actual, expected, verdict: 'FAIL',
        why: `${row.seat} ran on ${actual}, DEARER than the configured ${cls} tier ${expected}` })
    } else if (dearer(expected, actual)) {
      out.push({ seat: row.seat, round: row.round, cls, actual, expected, verdict: 'WARN',
        why: `${row.seat} ran on ${actual}, CHEAPER than the configured ${cls} tier ${expected} — verification may be discounted` })
    }
  }
  return out
}

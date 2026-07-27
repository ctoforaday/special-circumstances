// Seat identity from a transcript head — ONE table, because two copies of it had
// already drifted.
//
// The dashboard and the cost audit each carried their own classifier. They were
// identical except that cost-audit lacked the `Terminal dispute disposition` case,
// so every terminal-disposition transcript was filed as `other` and its spend was
// misattributed in the report the bench's economics decisions are made from. The
// divergence is invisible by construction: both functions "work", they just answer
// differently, and only a side-by-side read reveals it.
//
// A seat is identified by the opening text of the prompt the engine gave it, so
// this table is coupled to debate.js's prompt wording. Change a prompt heading and
// the seat silently becomes `other` — which is why `other` must stay a visible
// bucket in both consumers rather than being folded away as noise.
const ROUNDED = [
  [/Red audit, round (\d+)/, 'red-lens'],
  [/Red merge, round (\d+)/, 'red-merge'],
  [/Blue response, round (\d+)/, 'blue-respond'],
  [/Adjudication, round (\d+)/, 'judge'],
]

const UNROUNDED = [
  ['Terminal dispute disposition', 'judge-terminal'],
  // The petition sitting hears filed petitions BEFORE a round continues. Its prompt
  // head is `Petition sitting, topic "..."` (debate.js) — without this needle it fell
  // through to `other`, so its spend was misattributed (the same defect this file's
  // header documents for terminal-disposition) and the tier guard could not see it.
  ['Petition sitting', 'judge-petition'],
  ['Blue synthesis', 'blue-synthesize'],
  ['Blue lane', 'blue-lane'],
  ['frontier hypotheses', 'frontier'],
  ['Final assembly', 'assemble'],
]

export function classifySeat(head) {
  const h = String(head || '')
  for (const [re, seat] of ROUNDED) {
    const m = h.match(re)
    if (m) return { seat, round: +m[1] }
  }
  for (const [needle, seat] of UNROUNDED) if (h.includes(needle)) return { seat, round: 0 }
  return { seat: 'other', round: 0 }
}

// The seats this table can name. Exported so a test can assert the two consumers
// agree on the full set rather than only on the cases someone thought to check.
export const KNOWN_SEATS = [...ROUNDED.map(([, s]) => s), ...UNROUNDED.map(([, s]) => s), 'other']

// The ONE seat->tier-class map. debate.js dispatches each seat with `...bulk` or
// `...judgment`; the model-tier guard needs that same split as data. Keeping it here
// (beside the seat table) rather than re-encoding it in the guard, cost-audit, or the
// Go fuzz is the point — two copies of this fact would drift like the classifier did.
// debate-dispatch.test.mjs binds this map to debate.js's actual `agent(` spreads.
//   bulk      = the high-volume seats (lenses, lanes, responses, frontier)
//   judgment  = the reasoning seats (synthesis, merge, the bench, assembly)
export const SEAT_CLASS = {
  frontier: 'bulk',
  'blue-lane': 'bulk',
  'red-lens': 'bulk',
  'blue-respond': 'bulk',
  'blue-synthesize': 'judgment',
  'red-merge': 'judgment',
  judge: 'judgment',
  'judge-petition': 'judgment',
  'judge-terminal': 'judgment',
  assemble: 'judgment',
}
// null for `other`/unknown — those seats are not tier-bound, so the guard skips them.
export const classOf = (seat) => SEAT_CLASS[seat] || null

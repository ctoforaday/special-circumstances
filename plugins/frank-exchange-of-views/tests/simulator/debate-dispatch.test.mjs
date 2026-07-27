// node:test — binds the ONE seat->class map (seat-classify.mjs SEAT_CLASS) to debate.js's actual
// dispatch. debate.js spreads `...bulk` or `...judgment` into each agent() call's opts; this reads
// debate.js SOURCE and asserts every dispatch spreads the tier its seat's class maps to. If a
// dispatch ever uses the wrong tier, or a new seat is added without a class, this fails here rather
// than on a live run. debate.js runs under goja and is not importable as a module, so a source-lint
// is the honest binding — and it keeps the map single-source (the Go fuzz cannot import it and does
// not re-encode it).
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { SEAT_CLASS, classOf } from '../../skills/research-protocol/scripts/seat-classify.mjs'

const src = readFileSync(fileURLToPath(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url)), 'utf8')
// Longest first so `judge-petition`/`judge-terminal` win over the `judge` prefix.
const seatKeys = Object.keys(SEAT_CLASS).sort((a, b) => b.length - a.length)
const stemOf = (labelTemplate) => {
  const head = labelTemplate.split('·')[0].trim() // e.g. `red-lens-${role}-r${round}`
  return seatKeys.find((s) => head.startsWith(s)) || null
}

test('every debate.js agent() dispatch spreads the tier its seat class maps to', () => {
  // Each dispatch opts object opens `{ ...bulk, label: `…` }` or `{ ...judgment, label: `…` }`.
  const re = /\{\s*\.\.\.(bulk|judgment),\s*label:\s*`([^`]+)`/g
  const seen = new Set()
  let m
  let n = 0
  while ((m = re.exec(src))) {
    n++
    const [, spread, labelTemplate] = m
    const stem = stemOf(labelTemplate)
    assert.ok(stem, `dispatch label "${labelTemplate}" matched no known seat`)
    seen.add(stem)
    assert.equal(classOf(stem), spread, `seat "${stem}" dispatches with ...${spread} but SEAT_CLASS says ${classOf(stem)}`)
  }
  assert.ok(n >= 10, `expected at least 10 dispatch sites in debate.js, found ${n}`)
  // Completeness: no dead SEAT_CLASS key — every mapped seat is actually dispatched.
  for (const s of Object.keys(SEAT_CLASS)) assert.ok(seen.has(s), `SEAT_CLASS key "${s}" is never dispatched in debate.js`)
})

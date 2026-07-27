// node:test — the model-tier guard (#111). Pure `tierMismatch` over synthetic seat rows + config.
// Tier keys align with cost-audit's PRICES ladder: haiku < sonnet < opus < fable.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { tierMismatch } from '../../skills/research-protocol/scripts/model-guard.mjs'

test('a bulk seat DEARER than its configured tier is a FAIL (the fable trap)', () => {
  const out = tierMismatch([{ seat: 'red-lens', round: 1, t: 'fable' }], { model: 'haiku', judgmentModel: 'haiku' })
  assert.equal(out.length, 1)
  assert.equal(out[0].verdict, 'FAIL')
  assert.equal(out[0].cls, 'bulk')
  assert.match(out[0].why, /DEARER/)
})

test('a bulk seat CHEAPER than its configured tier is a WARN (discounted verification)', () => {
  const out = tierMismatch([{ seat: 'red-lens', round: 1, t: 'haiku' }], { model: 'opus', judgmentModel: 'opus' })
  assert.equal(out.length, 1)
  assert.equal(out[0].verdict, 'WARN')
  assert.match(out[0].why, /CHEAPER/)
})

test('a seat on exactly its configured tier is PASS (no finding)', () => {
  const out = tierMismatch([
    { seat: 'red-lens', round: 1, t: 'sonnet' },
    { seat: 'red-merge', round: 1, t: 'opus' },
  ], { model: 'sonnet', judgmentModel: 'opus' })
  assert.deepEqual(out, [])
})

test('a class with no configured tier WARNs "not declared" rather than skipping silently', () => {
  const out = tierMismatch([{ seat: 'red-lens', round: 1, t: 'haiku' }], { judgmentModel: 'opus' }) // model omitted
  assert.equal(out.length, 1)
  assert.equal(out[0].verdict, 'WARN')
  assert.match(out[0].why, /not declared/)
})

test('an other/unknown seat is not tier-bound and is skipped', () => {
  const out = tierMismatch([{ seat: 'other', round: 0, t: 'fable' }], { model: 'haiku', judgmentModel: 'haiku' })
  assert.deepEqual(out, [])
})

test('a judgment seat is measured against judgmentModel, not model', () => {
  // Bulk is configured opus (so an opus seat would be fine as a bulk seat), but this is a JUDGMENT
  // seat whose tier is judgmentModel=haiku — opus is dearer than ITS tier, so it must FAIL, proving
  // the guard picks the class's own configured tier.
  const out = tierMismatch([{ seat: 'red-merge', round: 2, t: 'opus' }], { model: 'opus', judgmentModel: 'haiku' })
  assert.equal(out.length, 1)
  assert.equal(out[0].verdict, 'FAIL')
  assert.equal(out[0].cls, 'judgment')
})

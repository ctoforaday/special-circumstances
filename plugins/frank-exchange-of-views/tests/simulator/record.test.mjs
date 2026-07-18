// node --test — lib/record.mjs + the red seat CLIs against temp-dir fixtures.
// R1 gates from plans/record-tool.md §V: append atomicity, shard-merge
// determinism (property test), structural dedup + multi-nonce winner selection
// (the 8/50 duplicate-dispatch anomaly's named test), idempotency, dangling-
// supersedes refusal, class validation incl. --class-new, render fixtures,
// role boundaries, atomic renders.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, existsSync, readFileSync, readdirSync, writeFileSync, mkdirSync, utimesSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, dirname } from 'node:path'
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { registerSeat, append, mergedEvents, boardState, mintGapId, render, roundOf, GRADES } from '../../skills/research-protocol/scripts/lib/record.mjs'

const tmp = () => mkdtempSync(join(tmpdir(), 'feov-record-'))
const SCRIPTS = join(dirname(fileURLToPath(import.meta.url)), '..', '..', 'skills', 'research-protocol', 'scripts')

const mintArgs = (over = {}) => ({
  gap_id: over.gap_id, class: 'false-universal', location: 'loc', problem: 'p', required_fix: 'f',
  acceptance_check: 'grep the corrected figure at its anchor',
  severity: 'medium', likelihood: 'medium', impact: 'medium', complexity_cost: 'low', supersedes: [], found_by: ['L1-F1'], ...over,
})

test('roundOf parses seat ids; register creates a pointer and a nonce shard', () => {
  assert.equal(roundOf('red-merge-r3'), 3)
  assert.equal(roundOf('red-lens-r4-L5'), 4)
  assert.equal(roundOf('blue-synthesize'), 0)
  const run = tmp()
  const { nonce, shard } = registerSeat(run, 'red-merge-r1')
  assert.ok(existsSync(shard) && nonce.length === 8)
})

test('append: per-shard monotonic seq; implicit register when no pointer exists', () => {
  const run = tmp()
  append(run, 'red-lens-r1-L1', 'friction', { text: 'first' })
  append(run, 'red-lens-r1-L1', 'friction', { text: 'second' })
  const { events } = mergedEvents(run)
  const fr = events.filter((e) => e.type === 'friction')
  assert.equal(fr.length, 2)
  assert.deepEqual(fr.map((e) => e.seq), [1, 2], 'seq 0 is the implicit register')
})

test('mint validation: acceptance_check required; dangling supersedes refused; unknown grade refused', () => {
  const run = tmp()
  assert.throws(() => append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', acceptance_check: undefined })), /acceptance check/)
  assert.throws(() => append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', supersedes: ['R0-9'] })), /dangling lineage/)
  assert.throws(() => append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', severity: 'sorta-bad' })), /grade enum/)
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1' }))
  append(run, 'red-merge-r2', 'mint', mintArgs({ gap_id: 'R2-1', supersedes: ['R1-1'] }))
  assert.equal(boardState(run).gaps.size, 2, 'valid lineage accepted')
})

test('class registry: unknown class refused with hint; --class-new demands definition+neighbor+distinguisher and extends the registry', () => {
  const run = tmp()
  mkdirSync(join(run, 'records'), { recursive: true })
  writeFileSync(join(run, 'records', 'class-registry.json'), JSON.stringify({ classes: [{ slug: 'false-universal' }, { slug: 'enumeration-non-exhaustive' }] }))
  assert.throws(() => append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', class: 'made-up-class' })), /unknown class.*class-new/s)
  assert.throws(() => append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', class: undefined, class_new: true })), /definition.*neighbor.*distinguisher|--definition/s)
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', class: 'name-keying-vs-marker', class_new: true, definition: 'keyed on name not marker', neighbor: 'enumeration-non-exhaustive', distinguisher: 'is the key a NAME or a set member?' }))
  append(run, 'red-merge-r1', 'class-new', { slug: 'name-keying-vs-marker', definition: 'keyed on name not marker', neighbor: 'enumeration-non-exhaustive', distinguisher: 'q' })
  append(run, 'red-merge-r2', 'mint', mintArgs({ gap_id: 'R2-1', class: 'name-keying-vs-marker' }))
  assert.equal(boardState(run).gaps.size, 2, 'extension class usable after class-new')
})

test('close validation: anchor OR carried-from required; regression demands successor; anchors render into the archive', () => {
  const run = tmp()
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1' }))
  assert.throws(() => append(run, 'red-merge-r2', 'close', { gap_id: 'R1-1', closure_class: 'closed' }), /anchor.*carried-from/s)
  assert.throws(() => append(run, 'red-merge-r2', 'close', { gap_id: 'R1-1', closure_class: 'closed_with_regression', anchor_seat: 'L1', anchor_tool: 'git show', anchor_target: 'pin:path' }), /successor/)
  append(run, 'red-merge-r2', 'close', { gap_id: 'R1-1', closure_class: 'closed', anchor_seat: 'L1', anchor_tool: 'git show', anchor_target: '7bc501e:ideas/backlog.md' })
  const out = render(run)
  const archive = readFileSync(join(out.out, 'archive.md'), 'utf8')
  assert.ok(archive.includes('L1 | git show | 7bc501e:ideas/backlog.md'), 'attestation anchor is IN the record')
})

test('carried-from renders as CARRIED, never as a fresh act (the E0.5a inflation becomes unphraseable)', () => {
  const run = tmp()
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1' }))
  append(run, 'red-merge-r3', 'close', { gap_id: 'R1-1', closure_class: 'closed', carried_from: '2' })
  const out = render(run)
  assert.ok(readFileSync(join(out.out, 'archive.md'), 'utf8').includes('CARRIED from round 2'))
})

test('shard-merge determinism (property): shuffled shard write order renders byte-identical projections', () => {
  const build = (order) => {
    const run = tmp()
    const seats = [
      ['red-lens-r1-L1', 'finding', { label: 'L1-F1', severity: 'high', likelihood: 'high', impact: 'medium', text: 'x' }],
      ['red-lens-r1-L2', 'finding', { label: 'L2-F1', severity: 'low', likelihood: 'low', impact: 'low', text: 'y' }],
      ['red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1' })],
    ]
    for (const i of order) append(run, ...seats[i])
    render(run)
    return readFileSync(join(run, 'records', 'render-shadow', 'ledger.md'), 'utf8')
  }
  assert.equal(build([0, 1, 2]), build([2, 1, 0]), 'append order does not change the render')
})

test('multi-nonce winner selection (the 8/50 duplicate-dispatch anomaly, named test): terminal-event nonce wins; dedup is FLAGGED, never silent', () => {
  const run = tmp()
  // First dispatch: mints, then dies before verdict.
  registerSeat(run, 'red-merge-r1')
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', problem: 'first dispatch text' }))
  // Re-dispatch (rotates the nonce): re-mints with DIFFERENT prose, reaches verdict.
  registerSeat(run, 'red-merge-r1')
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', problem: 'second dispatch text' }))
  append(run, 'red-merge-r1', 'verdict', { verdict: 'FAIL' })
  const { events, anomalies } = mergedEvents(run)
  assert.equal(events.filter((e) => e.type === 'mint').length, 1, 'one mint survives')
  assert.equal(events.find((e) => e.type === 'mint').payload.problem, 'second dispatch text', 'terminal-event nonce won')
  assert.ok(anomalies.some((a) => a.includes('multi-nonce seat red-merge-r1') && a.includes('terminal event')), 'anomaly names the seat and the rule')
  const out = render(run)
  assert.ok(readFileSync(join(out.out, 'ledger.md'), 'utf8').includes('render anomalies'), 'anomaly footer rendered')
})

test('multi-nonce with NO terminal event falls to latest-mtime (explicit fallback)', () => {
  const run = tmp()
  registerSeat(run, 'red-lens-r1-L1')
  append(run, 'red-lens-r1-L1', 'finding', { label: 'L1-F1', text: 'old' })
  const oldShard = readdirSync(join(run, 'records')).find((f) => f.startsWith('events-red-lens-r1-L1'))
  registerSeat(run, 'red-lens-r1-L1')
  append(run, 'red-lens-r1-L1', 'finding', { label: 'L1-F1', text: 'new' })
  // Force the first shard older regardless of fs timestamp granularity.
  utimesSync(join(run, 'records', oldShard), new Date(Date.now() - 60000), new Date(Date.now() - 60000))
  const { events, anomalies } = mergedEvents(run)
  assert.equal(events.find((e) => e.type === 'finding').payload.text, 'new')
  assert.ok(anomalies.some((a) => a.includes('mtime fallback')))
})

test('same-label-next-round is NOT a collision: L1-F1 in r1 and r2 both survive (keys round-qualified via seatId)', () => {
  const run = tmp()
  append(run, 'red-lens-r1-L1', 'finding', { label: 'L1-F1', text: 'round one' })
  append(run, 'red-lens-r2-L1', 'finding', { label: 'L1-F1', text: 'round two' })
  assert.equal(mergedEvents(run).events.filter((e) => e.type === 'finding').length, 2)
})

test('undisposed observations surface in the render; disposal clears them', () => {
  const run = tmp()
  append(run, 'red-lens-r1-L6', 'observe', { kind: 'note', label: 'L6-N1', text: 'over-normalization can merge two causes' })
  render(run)
  let ledger = readFileSync(join(run, 'records', 'render-shadow', 'ledger.md'), 'utf8')
  assert.ok(ledger.includes('undisposed lens observations') && ledger.includes('L6-N1'), 'note-fate tracking: the E0.5b loss surface is visible')
  append(run, 'red-merge-r1', 'dispose', { observation: 'L6-N1', disposition: 'declined', reason: 'below bar; noted' })
  render(run)
  ledger = readFileSync(join(run, 'records', 'render-shadow', 'ledger.md'), 'utf8')
  assert.ok(!ledger.includes('undisposed lens observations'), 'disposition clears the debt')
})

test('telemetry render: computed repair_regression and edge_deltas, never self-reported', () => {
  const run = tmp()
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', likelihood: 'high', impact: 'high' }))
  append(run, 'red-merge-r2', 'close', { gap_id: 'R1-1', closure_class: 'closed_with_regression', anchor_seat: 'L1', anchor_tool: 'read', anchor_target: 'report:100', successor: 'R2-1' })
  append(run, 'red-merge-r2', 'mint', mintArgs({ gap_id: 'R2-1', likelihood: 'medium', impact: 'medium', supersedes: ['R1-1'] }))
  render(run)
  const lines = readFileSync(join(run, 'records', 'render-shadow', 'board-telemetry.jsonl'), 'utf8').trim().split('\n').map(JSON.parse)
  const r2 = lines.find((l) => l.round === 2)
  assert.equal(r2.repair_regression.closures, 1)
  assert.equal(r2.repair_regression.lineage_mints, 1)
  assert.equal(r2.repair_regression.ratio, 1)
  assert.equal(r2.edge_deltas.down_mass, 5, '9 - 4 mass down the edge')
  assert.equal(r2.mass, 4, 'open mass = the successor only')
})

test('renders are atomic: no .tmp leftovers after rendering', () => {
  const run = tmp()
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1' }))
  render(run)
  const leftovers = readdirSync(join(run, 'records', 'render-shadow')).filter((f) => f.includes('.tmp-'))
  assert.equal(leftovers.length, 0)
})

test('role boundary: the lens CLI has no mint verb and says so; its help carries every verb (drift test)', () => {
  const lens = join(SCRIPTS, 'tools', 'red-lens.mjs')
  const run = tmp()
  const denied = spawnSync(process.execPath, [lens, 'mint', '--run', run, '--seat-id', 'red-lens-r1-L1'])
  assert.equal(denied.status, 2)
  assert.ok(denied.stderr.toString().includes("outside this seat's role"))
  const help = spawnSync(process.execPath, [lens, '--help']).stdout.toString()
  for (const v of ['register', 'finding', 'observe', 'cite', 'friction', 'petition', 'render']) {
    assert.ok(help.includes(v), `help missing verb ${v}`)
  }
})

test('red-merge CLI end-to-end: register -> mint (tool-assigned id) -> close -> verdict renders and checkpoints', () => {
  const merge = join(SCRIPTS, 'tools', 'red-merge.mjs')
  const run = tmp()
  const sh = (args) => spawnSync(process.execPath, [merge, ...args, '--run', run, '--seat-id', 'red-merge-r1'])
  assert.equal(sh(['register']).status, 0)
  const minted = sh(['mint', '--class', 'false-universal', '--location', 'S1', '--problem', 'quantifier unchecked', '--fix', 'enumerate the cases', '--check', 'grep every status for a timer', '--severity', 'medium', '--likelihood', 'medium', '--impact', 'medium', '--cx', 'low'])
  assert.equal(minted.status, 0, minted.stderr.toString())
  assert.ok(minted.stdout.toString().includes('minted R1-1'), 'tool-assigned id')
  const closed = sh(['close', '--id', 'R1-1', '--as', 'closed', '--anchor-seat', 'merge', '--anchor-tool', 'grep', '--anchor-target', 'report:S1'])
  assert.equal(closed.status, 0, closed.stderr.toString())
  const verdict = sh(['verdict', '--as', 'PASS'])
  assert.equal(verdict.status, 0, verdict.stderr.toString())
  assert.ok(verdict.stdout.toString().includes('checkpointed'), 'recovery mirror written')
  assert.ok(existsSync(join(run, 'records', 'render-shadow', 'ledger.md')), 'render-on-mutation left current projections')
})

test('mintGapId is sequential per round across the merged view', () => {
  const run = tmp()
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: mintGapId(run, 1) }))
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: mintGapId(run, 1) }))
  assert.equal(mintGapId(run, 1), 'R1-3')
  assert.equal(mintGapId(run, 2), 'R2-1')
})

test('grades enum is the pinned eight', () => {
  assert.deepEqual([...GRADES].sort(), ['certain', 'high', 'low', 'low-medium', 'medium', 'medium-high', 'realized', 'trivial'].sort())
})

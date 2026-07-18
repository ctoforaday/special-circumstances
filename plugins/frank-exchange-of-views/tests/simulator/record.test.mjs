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

// ---- Findings-audit additions (pre-merge sweep, 2026-07-18): every measured
// failure case from the corpus gets its named test ----

test('crash-retry idempotency: a retried mint with --key returns the EXISTING id, never double-mints', () => {
  const merge = join(SCRIPTS, 'tools', 'red-merge.mjs')
  const run = tmp()
  const margs = ['mint', '--key', 'L5-F3', '--class', 'false-universal', '--location', 'S1', '--problem', 'p', '--fix', 'f', '--check', 'c', '--severity', 'medium', '--likelihood', 'medium', '--impact', 'medium', '--cx', 'low', '--run', run, '--seat-id', 'red-merge-r1']
  const first = spawnSync(process.execPath, [merge, ...margs])
  assert.ok(first.stdout.toString().includes('minted R1-1'), first.stderr.toString())
  // The seat's message died after the successful bash call; the seat retries verbatim.
  const retry = spawnSync(process.execPath, [merge, ...margs])
  assert.ok(retry.stdout.toString().includes('minted R1-1') && retry.stdout.toString().includes('idempotent'), `retry must return the existing id: ${retry.stdout}`)
  const fresh = spawnSync(process.execPath, [merge, 'mint', '--class', 'false-universal', '--location', 'S2', '--problem', 'q', '--fix', 'f', '--check', 'c', '--severity', 'low', '--likelihood', 'low', '--impact', 'low', '--cx', 'low', '--run', run, '--seat-id', 'red-merge-r1'])
  assert.ok(fresh.stdout.toString().includes('minted R1-2'), 'a genuinely new mint still advances the id')
})

test('crash mid-append: a torn final line is tolerated; prior events survive intact', () => {
  const run = tmp()
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1' }))
  const shard = readdirSync(join(run, 'records')).find((f) => f.startsWith('events-red-merge-r1'))
  // Simulate the process dying mid-write: half a JSON object, no newline.
  writeFileSync(join(run, 'records', shard), '{"seq":99,"seatId":"red-merge-r1","type":"clo', { flag: 'a' })
  const { gaps } = boardState(run)
  assert.equal(gaps.size, 1, 'the mint before the torn line is intact')
  assert.ok(gaps.get('R1-1').open, 'the torn close never happened — renders as open (torn-round semantics)')
  append(run, 'red-merge-r1', 'friction', { text: 'appends still work after a torn line' })
  assert.equal(mergedEvents(run).events.filter((e) => e.type === 'friction').length, 1)
})

test('the quoting recurrence class: hostile prose (quotes, backticks, subshells, newlines, unicode) survives --file into the event and the render', () => {
  const merge = join(SCRIPTS, 'tools', 'red-merge.mjs')
  const run = tmp()
  const hostile = 'He said "never" - then ' + String.fromCharCode(96) + 'rm -rf' + String.fromCharCode(96) + ' and $(echo gotcha) \\ backslash\nsecond line: 100% of £ costs ≈ √2 × naïve\n'
  const proseFile = join(tmp(), 'problem.md')
  writeFileSync(proseFile, hostile)
  const r = spawnSync(process.execPath, [merge, 'mint', '--class', 'false-universal', '--location', 'S1', '--file', proseFile, '--fix', 'f', '--check', 'c', '--severity', 'medium', '--likelihood', 'medium', '--impact', 'medium', '--cx', 'low', '--run', run, '--seat-id', 'red-merge-r1'])
  assert.equal(r.status, 0, r.stderr.toString())
  const mint = mergedEvents(run).events.find((e) => e.type === 'mint')
  assert.equal(mint.payload.problem, hostile, 'byte-exact round trip through the event log')
  render(run)
  const ledger = readFileSync(join(run, 'records', 'render-shadow', 'ledger.md'), 'utf8')
  assert.ok(ledger.includes('He said "never"'), 'renders carry the prose')
})

test('E0.5b unauditability case: regrade history is fully recoverable — original grade, every basis, and the render says where to look', () => {
  const run = tmp()
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1', likelihood: 'high', impact: 'high' }))
  append(run, 'red-merge-r2', 'regrade', { gap_id: 'R1-1', likelihood: 'medium', basis: 'repair narrowed the trigger set to admin-only paths' })
  append(run, 'red-merge-r3', 'regrade', { gap_id: 'R1-1', impact: 'medium', basis: 'blast radius bounded by the new fence' })
  const { events, gaps } = boardState(run)
  const g = gaps.get('R1-1')
  assert.equal(g.likelihood, 'medium')
  assert.equal(g.impact, 'medium')
  const mint = events.find((e) => e.type === 'mint')
  assert.equal(mint.payload.likelihood, 'high', 'the ORIGINAL grade is still in the log — the run-4 unauditability class cannot recur')
  assert.equal(g.regrades.length, 2, 'every movement carries its recorded basis')
  render(run)
  const ledger = readFileSync(join(run, 'records', 'render-shadow', 'ledger.md'), 'utf8')
  assert.ok(ledger.includes('regraded x2') && ledger.includes('blast radius bounded'), 'the render discloses the history and its latest basis')
})

test('CONCURRENCY: 6 truly parallel lens processes (each register + finding + cite, every verb rendering) — all events land, projections parse, no locks or temps leak', async () => {
  const { spawn } = await import('node:child_process')
  const lens = join(SCRIPTS, 'tools', 'red-lens.mjs')
  const run = tmp()
  const runOne = (args) => new Promise((res) => {
    const c = spawn(process.execPath, [lens, ...args, '--run', run])
    let err = ''
    c.stderr.on('data', (d) => { err += d })
    c.on('close', (code) => res({ code, err }))
  })
  const jobs = []
  for (let i = 1; i <= 6; i++) {
    const sid = `red-lens-r1-L${i}`
    jobs.push((async () => {
      const reg = await runOne(['register', '--seat-id', sid])
      const f = await runOne(['finding', '--label', `L${i}-F1`, '--severity', 'medium', '--likelihood', 'medium', '--impact', 'medium', '--text', `parallel finding ${i}`, '--seat-id', sid])
      const c = await runOne(['cite', '--claim', `c${i}`, '--reference', `ref-${i}`, '--confidence', 'high', '--access-date', '2026-07-18', '--seat-id', sid])
      return [reg, f, c]
    })())
  }
  const results = (await Promise.all(jobs)).flat()
  for (const r of results) assert.equal(r.code, 0, r.err)
  const { events, anomalies } = mergedEvents(run)
  assert.equal(events.filter((e) => e.type === 'finding').length, 6, 'every parallel finding landed')
  assert.equal(events.filter((e) => e.type === 'cite').length, 6, 'every parallel cite landed')
  assert.equal(anomalies.length, 0, `no dedup anomalies from clean parallelism: ${anomalies.join('; ')}`)
  const recDir = join(run, 'records')
  const leftovers = readdirSync(recDir).filter((f) => f.startsWith('.lock-'))
  assert.equal(leftovers.length, 0, `no lock leaks: ${leftovers.join(', ')}`)
  const shadowLeft = readdirSync(join(recDir, 'render-shadow')).filter((f) => f.includes('.tmp-'))
  assert.equal(shadowLeft.length, 0, 'no temp leaks under concurrent renders')
  const ledger = readFileSync(join(recDir, 'render-shadow', 'citation-ledger.md'), 'utf8')
  assert.equal(ledger.split('\n').filter((l) => l.includes(' | ')).length, 6, 'final render is full-state and complete')
})

test('CONCURRENCY: a stale lock (crashed holder) is stolen, not deadlocked on', () => {
  const run = tmp()
  mkdirSync(join(run, 'records', '.lock-render'), { recursive: true })
  utimesSync(join(run, 'records', '.lock-render'), new Date(Date.now() - 60000), new Date(Date.now() - 60000))
  append(run, 'red-merge-r1', 'friction', { text: 'proceeds despite the corpse lock' })
  const t0 = Date.now()
  render(run)
  assert.ok(Date.now() - t0 < 5000, 'stale lock stolen well inside the wait bound')
  assert.ok(existsSync(join(run, 'records', 'render-shadow', 'ledger.md')))
  assert.ok(!existsSync(join(run, 'records', '.lock-render')), 'lock released after steal')
})

// ---- R2: blue + bench CLIs, projections, parity, join ----

test('R2 role boundaries: blue has no board verbs; bench cannot mint; both help contracts complete', () => {
  const blue = join(SCRIPTS, 'tools', 'blue.mjs')
  const bench = join(SCRIPTS, 'tools', 'bench.mjs')
  const run = tmp()
  for (const [tool, verb] of [[blue, 'mint'], [blue, 'close'], [bench, 'mint']]) {
    const r = spawnSync(process.execPath, [tool, verb, '--run', run, '--seat-id', 'x-r1'])
    assert.equal(r.status, 2)
    assert.ok(r.stderr.toString().includes("outside this seat's role"))
  }
  const blueHelp = spawnSync(process.execPath, [blue, '--help']).stdout.toString()
  for (const v of ['revision', 'manifest-row', 'dispute', 'confidence', 'position', 'closing', 'friction', 'petition']) assert.ok(blueHelp.includes(v), `blue help missing ${v}`)
  const benchHelp = spawnSync(process.execPath, [bench, '--help']).stdout.toString()
  for (const v of ['opinion', 'petition-rule', 'halt', 'certify', 'friction']) assert.ok(benchHelp.includes(v), `bench help missing ${v}`)
})

test('R2 projections: debate.md assembles positions/closings/opinions by round; CHANGELOG from revisions; citation-ledger from cites', () => {
  const run = tmp()
  append(run, 'red-merge-r1', 'position', { text: 'red says the board is heavy' })
  append(run, 'red-merge-r1', 'closing', { gap_id: 'RX', text: 'red closing case' })
  append(run, 'blue-respond-r1', 'position', { text: 'blue answers' })
  append(run, 'blue-respond-r1', 'closing', { gap_id: 'RX', text: 'blue closing case' })
  append(run, 'judge-r1', 'opinion', { gap_id: 'RX', disposition: 'carried', principle: 'correctness first', tension: 'correctness vs economy', review_flag: 'first opinion under the new form' })
  append(run, 'blue-respond-r1', 'revision', { text: 'round 1 edits: S2 rewritten', claim_count: 44 })
  append(run, 'red-lens-r1-L2', 'cite', { claim: 'c', reference: 'arXiv:1234.5678', confidence: 'high', access_date: '2026-07-18' })
  render(run)
  const shadow = join(run, 'records', 'render-shadow')
  const debate = readFileSync(join(shadow, 'debate.md'), 'utf8')
  assert.ok(debate.includes('### RED\n') && debate.includes('### BLUE\n') && debate.includes('### LEAD'), 'all three section kinds')
  assert.ok(debate.includes('RED CLOSING (round 1)') && debate.includes('BLUE CLOSING (round 1)'))
  assert.ok(debate.includes('principle: correctness first'), 'opinions carry their principles')
  assert.ok(readFileSync(join(shadow, 'CHANGELOG.md'), 'utf8').includes('## Round 1 (claim_count 44)'))
  assert.ok(readFileSync(join(shadow, 'citation-ledger.md'), 'utf8').includes('arXiv:1234.5678 | high | r1'))
})

test('R2 parity check: agreeing hand artifacts PASS; a hand-ledger gap missing from events FAILs with the id named', async () => {
  const { compare } = await import('../../skills/research-protocol/scripts/record-parity-check.mjs')
  const run = tmp()
  append(run, 'red-merge-r1', 'mint', mintArgs({ gap_id: 'R1-1' }))
  mkdirSync(join(run, 'red'), { recursive: true })
  mkdirSync(join(run, 'trajectories'), { recursive: true })
  writeFileSync(join(run, 'red', 'ledger.md'), '# ledger\n## OPEN\nR1-1 | medium | loc\n## closure index\n')
  let d = compare(run)
  assert.equal(d.length, 0, `agreeing records diverged: ${d.join('; ')}`)
  writeFileSync(join(run, 'red', 'ledger.md'), '# ledger\n## OPEN\nR1-1 | medium | loc\nR1-2 | high | loc2\n## closure index\n')
  d = compare(run)
  assert.ok(d.some((x) => x.includes('R1-2') && x.includes('missing from shadow')), `hand-only gap flagged: ${d.join('; ')}`)
})

test('R2 setup: records/ is created; stale mirrors purge at 30 days and fresh ones survive', async () => {
  const { buildSkeleton, purgeStaleMirrors } = await import('../../skills/research-protocol/scripts/setup-research-run.mjs')
  const run = tmp()
  buildSkeleton(run, 'topic')
  assert.ok(existsSync(join(run, 'records')), 'records/ born at setup')
  const mirrors = tmp()
  mkdirSync(join(mirrors, 'oldrun'))
  mkdirSync(join(mirrors, 'newrun'))
  utimesSync(join(mirrors, 'oldrun'), new Date(Date.now() - 40 * 86400e3), new Date(Date.now() - 40 * 86400e3))
  const r = purgeStaleMirrors(mirrors)
  assert.equal(r.purged, 1)
  assert.ok(!existsSync(join(mirrors, 'oldrun')) && existsSync(join(mirrors, 'newrun')))
})

test('R2 record-join audit: an event with no transcript invocation is FLAGGED; matching invocations PASS', async () => {
  const { recordJoinAudit } = await import('../../skills/research-protocol/scripts/capture-research-run.mjs')
  const run = tmp()
  const transcripts = tmp()
  append(run, 'red-lens-r1-L1', 'finding', { label: 'L1-F1', text: 'x' })
  // Transcript whose Bash commands carry the seat-id + verb.
  writeFileSync(join(transcripts, 'agent-aaa.jsonl'), [
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Bash', input: { command: 'node tools/red-lens.mjs register --run x --seat-id red-lens-r1-L1' } }] } }),
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Bash', input: { command: 'node tools/red-lens.mjs finding --label L1-F1 --text "x" --run x --seat-id red-lens-r1-L1' } }] } }),
  ].join('\n') + '\n')
  const pass = recordJoinAudit(run, transcripts, ['agent-aaa.jsonl'])
  assert.equal(pass.verdict, 'PASS', pass.detail)
  // A hand-appended event no transcript ever invoked.
  append(run, 'red-merge-r1', 'friction', { text: 'nobody ran this through a tool' })
  const flagged = recordJoinAudit(run, transcripts, ['agent-aaa.jsonl'])
  assert.equal(flagged.verdict, 'FAIL')
  assert.ok(flagged.detail.includes('red-merge-r1 friction'), 'the orphan event is named')
})

test('INTEGRATION - a full round through the CLIs: 6 lenses (findings, cites, notes) then the merge (dispose, mint with found_by, verdict) yields coherent projections', () => {
  const lens = join(SCRIPTS, 'tools', 'red-lens.mjs')
  const merge = join(SCRIPTS, 'tools', 'red-merge.mjs')
  const run = tmp()
  for (let i = 1; i <= 6; i++) {
    const sid = `red-lens-r1-L${i}`
    assert.equal(spawnSync(process.execPath, [lens, 'register', '--run', run, '--seat-id', sid]).status, 0)
    assert.equal(spawnSync(process.execPath, [lens, 'finding', '--label', `L${i}-F1`, '--severity', 'medium', '--likelihood', 'medium', '--impact', 'medium', '--text', `lens ${i} finding`, '--run', run, '--seat-id', sid]).status, 0)
    assert.equal(spawnSync(process.execPath, [lens, 'cite', '--claim', `claim ${i}`, '--reference', `arXiv:250${i}.0000${i}`, '--confidence', 'high', '--access-date', '2026-07-18', '--run', run, '--seat-id', sid]).status, 0)
  }
  assert.equal(spawnSync(process.execPath, [lens, 'observe', '--kind', 'note', '--label', 'L6-N1', '--text', 'sub-bar but real', '--run', run, '--seat-id', 'red-lens-r1-L6']).status, 0)
  const m = (args) => spawnSync(process.execPath, [merge, ...args, '--run', run, '--seat-id', 'red-merge-r1'])
  assert.equal(m(['register']).status, 0)
  for (let i = 1; i <= 6; i++) assert.equal(m(['dispose', '--observation', `L${i}-F1`, '--as', i <= 2 ? 'minted-as' : 'folded-into', '--into', 'R1-1']).status, 0)
  assert.equal(m(['dispose', '--observation', 'L6-N1', '--as', 'declined', '--reason', 'below bar, noted']).status, 0)
  const minted = m(['mint', '--key', 'L1-F1', '--class', 'false-universal', '--location', 'S2', '--problem', 'merged finding', '--fix', 'fix it', '--check', 'grep it', '--severity', 'medium', '--likelihood', 'medium', '--impact', 'medium', '--cx', 'low', '--found-by', 'L1-F1,L2-F1'])
  assert.ok(minted.stdout.toString().includes('minted R1-1'), minted.stderr.toString())
  assert.equal(m(['verdict', '--as', 'FAIL']).status, 0)
  const ledger = readFileSync(join(run, 'records', 'render-shadow', 'ledger.md'), 'utf8')
  assert.ok(ledger.includes('OPEN GAPS (1)') && ledger.includes('found_by L1-F1,L2-F1'), 'board coherent, lens-finding-granular attribution rendered')
  assert.ok(!ledger.includes('undisposed'), 'every observation received its fate')
  const telemetry = readFileSync(join(run, 'records', 'render-shadow', 'board-telemetry.jsonl'), 'utf8').trim().split('\n').map(JSON.parse)
  assert.equal(telemetry[0].open_count, 1)
  assert.equal(mergedEvents(run).events.filter((e) => e.type === 'cite').length, 6, 'the citation ledger events all survive')
})

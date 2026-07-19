// node --test — measure-read-batching.mjs, the instrument that measures what the
// merge seat's prompt-level read batching actually saves.
//
// The measurement is the whole product here: this script exists to put a dollar
// figure in a run record, and a mis-paired turn changes that figure silently.
// Fixtures are hand-written transcript JSONL rather than recorded runs, for the
// same reason the record readers are tested that way — the JSONL shape IS the
// interface, and a fixture built by the thing under test cannot catch a shared
// misreading of it.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { measureTranscript, formatRow, RATE_READ, RATE_WRITE, RATE_IN }
  from '../../skills/research-protocol/scripts/measure-read-batching.mjs'

const tool = new URL('../../skills/research-protocol/scripts/measure-read-batching.mjs', import.meta.url).pathname
  .replace(/^\/([A-Za-z]):\//, '$1:/') // Windows: strip the leading slash file URLs add

const L = (o) => JSON.stringify(o)
const toolUse = (id, file_path) => L({ message: { role: 'assistant', content: [{ type: 'tool_use', id, input: { file_path } }] } })
const toolResult = (...ids) => L({ message: { role: 'user', content: ids.map((id) => ({ type: 'tool_result', tool_use_id: id })) } })
const turn = (cr, cw = 0, inp = 0) => L({ message: { role: 'assistant', usage: { cache_read_input_tokens: cr, cache_creation_input_tokens: cw, input_tokens: inp }, content: [] } })
const MERGE = L({ message: { role: 'user', content: 'Red merge, round 3. FIRST ACTION: read the lens passes' } })

test('only red-merge transcripts are measured — the saving does not exist at other seats', () => {
  // A false positive here would attribute another seat's cache reads to batching
  // and inflate the reported saving, which is the number the run record keeps.
  assert.equal(measureTranscript(L({ message: { role: 'user', content: 'Red audit, round 3, lens: x' } })), null)
  assert.equal(measureTranscript(L({ message: { role: 'user', content: 'Blue synthesis' } })), null)
  assert.equal(measureTranscript(''), null)
  assert.equal(measureTranscript('Red merge, round nine'), null, 'the round must be a number to be a round')
  assert.equal(measureTranscript(MERGE).round, 3)
})

test('k candidate ingestions cost k-1 avoidable calls; only the FIRST is unavoidable', () => {
  const txt = [
    MERGE,
    toolUse('c1', 'run/red/candidates/a.md'),
    toolUse('c2', 'run/red/candidates/b.md'),
    toolUse('c3', 'run/red/candidates/c.md'),
    toolResult('c1'), turn(1_000_000),
    toolResult('c2'), turn(2_000_000),
    toolResult('c3'), turn(3_000_000, 100_000, 1_000),
  ].join('\n')
  const r = measureTranscript(txt)
  assert.equal(r.ingest.length, 3, 'three separate ingestion calls')
  assert.equal(r.avoided, 2, 'batching collapses them into one; the other two are the saving')
  // The saving is the billed input of the avoided calls, at the fable session-tier
  // rates this instrument documents.
  const expected = (2e6 * RATE_READ + 3e6 * RATE_READ + 1e5 * RATE_WRITE + 1e3 * RATE_IN) / 1e6
  assert.equal(r.dollars.toFixed(4), expected.toFixed(4))
  assert.deepEqual(r.cacheReads, [1_000_000, 2_000_000, 3_000_000], 'per-call cache reads are disclosed, not just the total')
})

test('candidate results arriving before ONE assistant turn are ONE ingestion — that collapse IS the saving', () => {
  // Counting each batched result separately would report the fully-batched merge
  // as having avoided nothing, i.e. the instrument would read zero exactly when
  // the optimization is working perfectly.
  const txt = [
    MERGE,
    toolUse('c1', 'run/red/candidates/a.md'),
    toolUse('c2', 'run/red/candidates/b.md'),
    toolResult('c1', 'c2'),
    turn(5_000_000),
  ].join('\n')
  const r = measureTranscript(txt)
  assert.equal(r.ingest.length, 1)
  assert.equal(r.avoided, 0)
  assert.equal(r.dollars, 0, 'a perfectly batched merge has nothing left to avoid')
})

test('only CANDIDATE reads count, and Windows backslash paths are the same path', () => {
  const txt = [
    MERGE,
    toolUse('other', 'run/blue/report.md'),
    toolResult('other'), turn(9_000_000),          // not a candidate: must not be an ingestion
    toolUse('c1', 'run\\red\\candidates\\a.md'),   // the separator the harness emits on Windows
    toolResult('c1'), turn(1_000_000),
    toolUse('c2', 'run/red/candidates/b.md'),
    toolResult('c2'), turn(2_000_000),
  ].join('\n')
  const r = measureTranscript(txt)
  assert.equal(r.ingest.length, 2, 'the blue/report.md read is not a candidate ingestion')
  assert.deepEqual(r.cacheReads, [1_000_000, 2_000_000], 'the 9M turn was never attributed')
  assert.equal(r.avoided, 1)
})

test('a bash-command read of a candidate counts too — the seat is not required to use Read', () => {
  const txt = [
    MERGE,
    L({ message: { role: 'assistant', content: [{ type: 'tool_use', id: 'b1', input: { command: 'cat run/red/candidates/a.md' } }] } }),
    toolResult('b1'), turn(1_000_000),
    L({ message: { role: 'assistant', content: [{ type: 'tool_use', id: 'b2', input: { command: 'cat run/red/candidates/b.md' } }] } }),
    toolResult('b2'), turn(2_000_000),
  ].join('\n')
  assert.equal(measureTranscript(txt).avoided, 1)
})

test('zero-usage turns do not consume a pending candidate result', () => {
  // A turn that billed nothing did not ingest anything. Letting it absorb the
  // pending result would attribute the ingestion to a free turn and report the
  // saving as $0.
  const txt = [
    MERGE,
    toolUse('c1', 'run/red/candidates/a.md'),
    toolResult('c1'),
    turn(0, 0, 0),        // billed nothing
    turn(4_000_000),      // the real ingesting turn
    toolUse('c2', 'run/red/candidates/b.md'),
    toolResult('c2'), turn(1_000_000),
  ].join('\n')
  const r = measureTranscript(txt)
  assert.deepEqual(r.cacheReads, [4_000_000, 1_000_000])
  assert.equal(r.avoided, 1)
})

test('a transcript killed mid-append still measures — the half-written last line is skipped', () => {
  // Runs get killed. An instrument that throws on the resulting truncated JSONL
  // reports nothing at all for the run most worth looking at.
  const txt = [
    MERGE,
    toolUse('c1', 'run/red/candidates/a.md'),
    toolResult('c1'), turn(1_000_000),
    toolUse('c2', 'run/red/candidates/b.md'),
    toolResult('c2'), turn(2_000_000),
    '{"message":{"role":"assistant","usage":{"cache_read',
  ].join('\n')
  const r = measureTranscript(txt)
  assert.equal(r.ingest.length, 2)
  assert.equal(r.avoided, 1)
})

test('a merge that never read a candidate reports zero, not a crash and not a fabricated saving', () => {
  const r = measureTranscript([MERGE, turn(5_000_000)].join('\n'))
  assert.deepEqual([r.ingest.length, r.avoided, r.dollars], [0, 0, 0])
  assert.equal(formatRow(r), 'round 3: 0 candidate ingestion calls; avoided 0; billed input avoided ~= $0.00 (cache_read at those calls: )')
})

test('the CLI reports rounds in numeric order with a summed total', () => {
  // Round ordering is by NUMBER, not by filename: the transcripts are named by
  // agent id, so file order says nothing about which round came first.
  const dir = mkdtempSync(join(tmpdir(), 'read-batching-fixture-'))
  const merge = (round, ...rest) => [
    L({ message: { role: 'user', content: `Red merge, round ${round}. FIRST ACTION` } }), ...rest,
  ].join('\n')
  writeFileSync(join(dir, 'agent-aaa.jsonl'), merge(10,
    toolUse('c1', 'r/red/candidates/a.md'), toolResult('c1'), turn(1_000_000),
    toolUse('c2', 'r/red/candidates/b.md'), toolResult('c2'), turn(1_000_000)))
  writeFileSync(join(dir, 'agent-zzz.jsonl'), merge(2,
    toolUse('c1', 'r/red/candidates/a.md'), toolResult('c1'), turn(1_000_000),
    toolUse('c2', 'r/red/candidates/b.md'), toolResult('c2'), turn(3_000_000)))
  writeFileSync(join(dir, 'agent-lens.jsonl'),
    L({ message: { role: 'user', content: 'Red audit, round 2, lens: x' } }))

  const out = execFileSync(process.execPath, [tool, dir], { encoding: 'utf8' })
  const lines = out.trim().split('\n')
  assert.match(lines[0], /^round 2: 2 candidate ingestion calls; avoided 1; billed input avoided ~= \$3\.00/)
  assert.match(lines[1], /^round 10: 2 candidate ingestion calls; avoided 1; billed input avoided ~= \$1\.00/)
  assert.equal(lines[2], 'sum avoided across rounds: $4.00')
  assert.ok(!out.includes('red-lens') && lines.length === 3, 'the lens transcript contributes no row')
})

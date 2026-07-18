// node --test — the mjs READERS of the record layer: parity check, setup
// skeleton + mirror purge, and capture's record-join audit.
//
// This file replaces the writer-side half of the old record.test.mjs. The mjs
// record library and its four seat CLIs are gone: they existed to be validated
// against, the Go port was validated against them, and they were never used in a
// run. Their behaviour now lives in the Go suite and in the golden transcripts
// that were recorded while the differential gate was green.
//
// What survives here is everything that READS the record layer — and these tests
// deliberately build their fixtures by writing JSONL DIRECTLY rather than by
// calling any writer. That is the architecture, not a convenience: the JSONL is
// the interface, readers must not carry a dependency on whichever implementation
// owns the write path, and a reader test that needs the writer to run cannot
// distinguish "the reader is correct" from "both sides share a bug".
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdirSync, writeFileSync, existsSync, utimesSync, mkdtempSync } from 'node:fs'
import { join } from 'node:path'
import { tmpdir } from 'node:os'

const tmp = () => mkdtempSync(join(tmpdir(), 'feov-rec-'))

// writeEvents emits a shard exactly as feov-record does: one JSON object per
// line, fields in the tool's order. Hand-writing it here is the point — if the
// binary ever changes this shape, these readers must fail.
function writeEvents(runDir, seatId, events, nonce = 'aabbccdd') {
  const dir = join(runDir, 'records')
  mkdirSync(dir, { recursive: true })
  const round = (/-r(\d+)/.exec(seatId) || [, '0'])[1]
  const lines = events.map((e, i) => JSON.stringify({
    seq: i, seatId, nonce, round: Number(round), type: e.type,
    key: e.key || `${seatId}:${e.type}:${e.payload.gap_id || e.payload.label || `#${i}`}`,
    payload: e.payload,
  }))
  writeFileSync(join(dir, `events-${seatId}-${nonce}.jsonl`), lines.join('\n') + '\n')
  writeFileSync(join(dir, `.active-${seatId}`), nonce)
}

// A minimal shadow render, as the binary leaves on disk after render-on-mutation.
function writeShadow(runDir, files) {
  const dir = join(runDir, 'records', 'render-shadow')
  mkdirSync(dir, { recursive: true })
  for (const [name, body] of Object.entries(files)) writeFileSync(join(dir, name), body)
}

test('parity check: agreeing hand artifacts PASS; a hand-only gap FAILs with the id named', async () => {
  const { compare } = await import('../../skills/research-protocol/scripts/record-parity-check.mjs')
  const run = tmp()
  mkdirSync(join(run, 'red'), { recursive: true })
  mkdirSync(join(run, 'trajectories'), { recursive: true })
  writeShadow(run, {
    'ledger.md': '# ledger\n\n## OPEN GAPS (1)\n\n### R1-1 — a gap\n\n## CLOSURE INDEX\n\n',
  })
  writeFileSync(join(run, 'red', 'ledger.md'), '# ledger\n## OPEN\nR1-1 | medium | loc\n## closure index\n')
  let d = compare(run)
  assert.equal(d.length, 0, `agreeing records diverged: ${d.join('; ')}`)

  writeFileSync(join(run, 'red', 'ledger.md'), '# ledger\n## OPEN\nR1-1 | medium | loc\nR1-2 | high | loc2\n## closure index\n')
  d = compare(run)
  assert.ok(d.some((x) => x.includes('R1-2') && x.includes('missing from shadow')), `hand-only gap flagged: ${d.join('; ')}`)
})

test('parity check refuses to pass over an absent shadow (a gate with nothing to compare is not passing)', async () => {
  const { compare } = await import('../../skills/research-protocol/scripts/record-parity-check.mjs')
  assert.throws(() => compare(tmp()), /no shadow renders/)
})

test('setup: records/ is created; stale mirrors purge at 30 days and fresh ones survive', async () => {
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

test('record-join audit: an event no transcript invoked is FLAGGED; the binary invocation shape is matched', async () => {
  const { recordJoinAudit } = await import('../../skills/research-protocol/scripts/capture-research-run.mjs')
  const run = tmp()
  const transcripts = tmp()
  writeEvents(run, 'red-lens-r1-L1', [{ type: 'finding', payload: { label: 'L1-F1', text: 'x' } }])
  // Transcript carrying the ROLE-SUBCOMMAND form the binary is invoked with.
  writeFileSync(join(transcripts, 'agent-aaa.jsonl'), [
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Bash', input: { command: 'feov-record lens register --run x --seat-id red-lens-r1-L1' } }] } }),
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Bash', input: { command: 'feov-record lens finding --label L1-F1 --text "x" --run x --seat-id red-lens-r1-L1' } }] } }),
  ].join('\n') + '\n')
  const pass = recordJoinAudit(run, transcripts, ['agent-aaa.jsonl'])
  assert.equal(pass.verdict, 'PASS', pass.detail)

  // An event no transcript ever invoked: the orphan must be named, not counted.
  writeEvents(run, 'red-merge-r1', [{ type: 'friction', payload: { text: 'nobody ran this through a tool' } }], 'ddeeff00')
  const flagged = recordJoinAudit(run, transcripts, ['agent-aaa.jsonl'])
  assert.equal(flagged.verdict, 'FAIL')
  assert.ok(flagged.detail.includes('red-merge-r1 friction'), 'the orphan event is named')
})

test('record-join audit still reads pre-port transcripts (the mjs invocation shape)', async () => {
  const { recordJoinAudit } = await import('../../skills/research-protocol/scripts/capture-research-run.mjs')
  const run = tmp()
  const transcripts = tmp()
  writeEvents(run, 'red-lens-r1-L1', [{ type: 'finding', payload: { label: 'L1-F1', text: 'x' } }])
  writeFileSync(join(transcripts, 'agent-old.jsonl'),
    JSON.stringify({ message: { role: 'assistant', content: [{ type: 'tool_use', name: 'Bash', input: { command: 'node tools/red-lens.mjs finding --label L1-F1 --run x --seat-id red-lens-r1-L1' } }] } }) + '\n')
  const pass = recordJoinAudit(run, transcripts, ['agent-old.jsonl'])
  assert.equal(pass.verdict, 'PASS', pass.detail)
})

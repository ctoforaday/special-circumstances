// node --test — SEAT PROMPT GOLDENS.
//
// The seat prompt is the product. It is what a seat actually reads and acts on,
// and every wave edits it: W2g moved the catechism into blue's round-0 duty, W2i
// added the consolidated-citation duty and re-pointed lens identity at roles.
// Each of those shipped with substring assertions — `prompt.includes('<a phrase>')` —
// which prove a phrase is PRESENT and say nothing about what else moved around
// it. A clause can be deleted, duplicated, contradicted, or attached to the
// wrong seat with every such test still green.
//
// These goldens capture each seat class's prompt in full, so a wave's effect on
// the product surface arrives as a reviewable diff. The churn is the point: an
// intentional prompt change SHOULD produce a visible diff, recorded on its own
// commit (see scripts/golden.mjs).
import { test, after } from 'node:test'
import assert from 'node:assert/strict'
import { loadDebateScript, makeWorld, makeResponder, blueEnv, redEnv, gap, judgeEnv } from './harness.mjs'
import { assertGolden, orphanGoldens, goldenReport } from './golden.mjs'
import { createHash } from 'node:crypto'


// SEGMENTATION — the goldens are for a human to review, and a 2,257-byte prompt
// on ONE line defeats that completely: git renders any change as the whole line
// replaced, so the diff says "something moved" and nothing else. That is the job
// these goldens exist to do, so failing it is not cosmetic.
//
// Formatting the golden is behaviourally inert — the golden is a RECORDING, not
// the prompt; the seat still receives the original string. What the layout buys
// is line-granular diffs: edit one clause and one line moves.
//
// Breaks are inserted at SEMANTIC boundaries, not at a column width. Wrapping at
// N columns would look tidy and diff terribly — inserting one word reflows every
// line after it, which is the same "everything changed" failure in a prettier
// hat. Sentence ends and the prompts' own ALL-CAPS section labels are stable
// under edits.
const CAPS_LABEL = /(?=[A-Z][A-Z0-9]{2,}(?:[ \-][A-Z0-9()§.]{2,})*\s*(?::|\())/g
const SENTENCE_END = /(?<=[.!?])\s+(?=[A-Z(§"'`])/g

export function segment(prompt) {
  return String(prompt)
    .replace(SENTENCE_END, '\n')
    .replace(CAPS_LABEL, '\n')
    .replace(/\n{3,}/g, '\n\n')
}

// The segmented body is for reading; the HASH is what makes the golden exact.
// Segmentation normalizes whitespace, so a whitespace-only prompt change would
// otherwise slip through invisibly — the hash of the RAW prompt catches it, and
// changes on its own line when it does.
function goldenBody(prompt) {
  return `prompt sha256: ${createHash('sha256').update(prompt).digest('hex')}\n\n${segment(prompt)}\n`
}

// A segmentation bug that dropped or duplicated text would corrupt every golden
// silently, so the transform is checked for word-level losslessness each time.
function assertLossless(name, prompt) {
  const flat = (t) => String(t).replace(/\s+/g, ' ').trim()
  assert.equal(flat(segment(prompt)), flat(prompt), `${name}: segmentation changed the prompt's words`)
}

const script = loadDebateScript(new URL('../../skills/research-protocol/scripts/debate.js', import.meta.url))
const ARGS = { topic: 'the seat prompt contract', runDir: 'research/2026-01-01_golden', lanes: 3 }

// A debate that reaches every seat class: round 1 FAILs with a docketed gap (so
// the bench sits), blue responds, round 2 PASSes, assembly runs. claim_count is
// chosen to give round 1 the full four citation slices and round 2 a small
// delta, which is what exercises W2i's consolidated seat.
async function fullRun() {
  const world = makeWorld(makeResponder({
    blueSynth: [blueEnv({ claim_count: 200 })],
    // A grade dispute is what puts a judge on the record in this shape: judge-rN
    // needs a CONTESTED docket (raised, rebutted, re-raised) and judge-terminal
    // needs undisposed disputes at the exit boundary. Without one, no bench seat
    // dispatches at all — which the roster golden would faithfully record, but
    // would leave the engine's most consequential prompt uncovered.
    blueRespond: [blueEnv({
      claim_count: 210,
      grade_disputes: [{ gap_id: 'R1-1', dimension: 'likelihood', proposed: 'low', evidence: 'the harm needs two independent failures' }],
    })],
    red: [redEnv({ gaps: [gap('R1-1')] }), redEnv({ verdict: 'PASS' })],
    judge: [judgeEnv({ resolutions: [{ gap_id: 'R1-1', resolution: 'carried', rationale: 'the figure is still unrecomputed' }] })],
  }))
  await world.run(script, ARGS)
  return world
}

// One golden per seat CLASS. Round-2 lenses are captured separately from round-1
// lenses because that is precisely where W2i's dispatch graduation lives, and a
// single-round capture would have proved nothing about it.
const SEATS = [
  ['frontier', (l) => l.startsWith('frontier')],
  ['blue-lane-1', (l) => l.startsWith('blue-lane-1')],
  ['blue-lane-3', (l) => l.startsWith('blue-lane-3')],
  ['blue-synthesize', (l) => l.startsWith('blue-synthesize')],
  ['red-lens-citation-r1', (l) => l.startsWith('red-lens-1-r1')],
  ['red-lens-logic-L5-r1', (l) => l.startsWith('red-lens-5-r1')],
  ['red-lens-darkside-L6-r1', (l) => l.startsWith('red-lens-6-r1')],
  ['red-lens-citation-r2-consolidated', (l) => l.startsWith('red-lens-1-r2')],
  ['red-lens-logic-L5-r2', (l) => l.startsWith('red-lens-5-r2')],
  ['red-merge-r1', (l) => l.startsWith('red-merge-r1')],
  ['blue-respond-r1', (l) => l.startsWith('blue-respond-r1')],
  ['judge-terminal', (l) => l.startsWith('judge-terminal')],
  ['assemble', (l) => l.startsWith('assemble')],
]

test('seat prompt goldens: every seat class carries exactly its recorded contract', async () => {
  const world = await fullRun()
  const missing = []
  for (const [name, match] of SEATS) {
    const call = world.calls.find((c) => match(c.opts.label))
    if (!call) {
      missing.push(name)
      continue
    }
    assertLossless(name, call.prompt)
    assertGolden(import.meta.url, `prompt-${name}`, goldenBody(call.prompt))
  }
  assert.deepEqual(missing, [], 'a seat class vanished from dispatch — the golden set no longer covers the engine')
})

test('seat roster: the dispatched seat set itself is pinned', async () => {
  const world = await fullRun()
  // The ROSTER is a contract too: a wave that silently adds a seat, drops one,
  // or changes how many lenses a round dispatches shows up here even when every
  // individual prompt is unchanged.
  const roster = world.calls
    .map((c) => c.opts.label.replace(/ · .*$/, ''))
    .join('\n')
  assertGolden(import.meta.url, 'seat-roster', roster)
})

after(() => {
  const orphans = orphanGoldens(import.meta.url)
  if (orphans.length) {
    // A renamed or deleted seat leaves its golden behind, where it reads as an
    // authoritative record of a contract nothing asserts any more.
    console.error(`orphaned goldens (no test claims them):\n${orphans.map((o) => `  ${o}`).join('\n')}`)
    process.exitCode = 1
  }
  const report = goldenReport()
  if (report && report.length) {
    console.error(`recorded ${report.length} golden change(s): ${report.map((r) => r.name).join(', ')}`)
  }
})

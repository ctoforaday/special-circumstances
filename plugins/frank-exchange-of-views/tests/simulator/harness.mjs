// Zero-token simulator for the debate engine (skills/research-protocol/scripts/debate.js).
//
// The Workflow harness evaluates the script body as async code with agent / parallel /
// pipeline / log / phase / args / budget in scope. This harness reproduces that contract
// so every control-flow branch — args parsing, the round loop, the contested docket,
// deadlock, the safety ceiling, per-role model routing, friction aggregation — is
// exercised with canned envelopes and no live agents.
//
// Founding regressions (each cost a live run to discover):
//   1. stringified args on resume -> destructured undefined -> literal 'undefined' paths (run 1)
//   2. agent() returning null on terminal failure -> TypeError on redEnv.verdict (run 2)
import { readFileSync } from 'node:fs'

const AsyncFunction = Object.getPrototypeOf(async function () {}).constructor

// The script is not a module we can import: `export const meta` + top-level await + free
// identifiers. Strip the export and wrap the body in an AsyncFunction whose parameters
// are the Workflow globals; the script's final `return {...}` becomes the function's return.
export function loadDebateScript(path) {
  const src = readFileSync(path, 'utf8').replace(/^export const meta/m, 'const meta')
  return new AsyncFunction('agent', 'parallel', 'pipeline', 'log', 'phase', 'args', 'budget', src)
}

// A stub world. `respond(prompt, opts, call)` supplies each agent() result (return null to
// model a dead agent); every call is recorded on `calls` for assertions. parallel/pipeline
// mirror the real semantics that matter to the script: a throwing thunk resolves to null,
// never rejects the batch.
export function makeWorld(respond) {
  const calls = []
  const logs = []
  const phases = []
  const world = {
    calls, logs, phases,
    agent: async (prompt, opts = {}) => {
      const call = { n: calls.length, prompt, opts }
      calls.push(call)
      return respond(prompt, opts, call)
    },
    parallel: async (thunks) => {
      const out = []
      for (const t of thunks) {
        try { out.push(await t()) } catch { out.push(null) }
      }
      return out
    },
    pipeline: async (items, ...stages) => {
      const out = []
      for (const [i, item] of items.entries()) {
        let v = item
        try { for (const s of stages) v = await s(v, item, i); out.push(v) } catch { out.push(null) }
      }
      return out
    },
    log: (m) => logs.push(m),
    phase: (t) => phases.push(t),
    budget: { total: null, spent: () => 0, remaining: () => Infinity },
  }
  world.run = (script, args) =>
    script(world.agent, world.parallel, world.pipeline, world.log, world.phase, args, world.budget)
  return world
}

// Canned envelopes, schema-shaped.
export const blueEnv = (over = {}) => ({
  path: 'blue/report.md', tldr: 'tldr', claim_count: 40, saturation_reached: true, sitting_record_appended: true,
  manifest: ['G1'],
  open_questions: [], log: [], ...over,
})
// THE CHAIR RELAYS THE RECORD (plans/roundless.md §III.B.1): its envelope carries the plan `dispatch
// next` printed. A stubbed chair is a stubbed plan; the default sequence is one sitting that engages
// the evidence lens and blue on G1, then one with nobody ready and PASS permitted, on which the chair
// records PASS — the shortest VERIFIED run.
export const party = (seat_id, ...gap_ids) => ({ seat_id, gap_ids })
export const plan = (parties = [], over = {}) => ({ head: 2, parties, docket: [], pass_permitted: false, ceiling: false, why: [], ...over })
export const passPlan = (over = {}) => plan([], { pass_permitted: true, ...over })
export const ceilingPlan = (over = {}) => plan([], { ceiling: true, why: ['G1: at impasse, ruled remanded — at its limit'], ...over })
export const chairEnv = (over = {}) => ({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G1')]), unruled_motions: 0, log: [], ...over })
export const passChair = (over = {}) => chairEnv({ plan: passPlan(), verdict: 'PASS', ...over })
export const gap = (id, over = {}) => ({
  id, location: 'loc', problem: 'p', required_fix: 'f', acceptance_check: 'grep the corrected figure at the anchor', existence: 'verified',
  severity: 'medium', likelihood: 'medium', impact: 'medium', complexity_cost: 'low', ...over,
})
export const judgeEnv = (over = {}) => ({ resolutions: [], log: [], ...over })
export const petitionRulingEnv = (over = {}) => ({ rulings: [{ petitioner: 'x', class: 'ethical', ruling: 'denied' }], log: [], ...over })
// makeResponder serves envelopes by seat, in order; the last one repeats. Lenses answer in free text.
export const assembleEnv = (over = {}) => ({ synopsis: 'synopsis', open_gaps: 0, log: [], ...over })
export function makeResponder({ chair = [chairEnv(), passChair()], judge = [judgeEnv()], blueSynth = [blueEnv()], blueRespond = [blueEnv()], petition = [petitionRulingEnv()], assemble = [assembleEnv()] } = {}) {
  const take = (q) => (q.length > 1 ? q.shift() : q[0])
  return (prompt, opts) => {
    const label = opts.label || ''
    if (label.startsWith('red-chair')) return take(chair)
    if (label.startsWith('judge-petition')) return take(petition)
    if (label.startsWith('judge')) return take(judge)
    if (label.startsWith('blue-synthesize')) return take(blueSynth)
    if (label.startsWith('blue-respond')) return take(blueRespond)
    if (label.startsWith('assemble')) return take(assemble)
    return 'synopsis'
  }
}

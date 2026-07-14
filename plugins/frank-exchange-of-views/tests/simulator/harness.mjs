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
  path: 'blue/report.md', tldr: 'tldr', claim_count: 40, saturation_reached: true,
  open_questions: [], friction: [], ...over,
})
export const redEnv = (over = {}) => ({
  verdict: 'FAIL', gaps: [], corroboration: [], citations_checked: 10, notes: '', friction: [], ...over,
})
export const gap = (id, over = {}) => ({
  id, location: 'loc', problem: 'p', required_fix: 'f',
  severity: 'medium', likelihood: 'medium', impact: 'medium', complexity_cost: 'low', ...over,
})
export const judgeEnv = (over = {}) => ({ deadlock: false, resolutions: [], friction: [], ...over })

// Role-routing responder: dispatches on the label prefix the script assigns each seat.
// red / judge / blueSynth / blueRespond are queues consumed in call order (last entry
// repeats); bulk narrative seats (frontier, lanes, lenses, assemble) return synopsis text.
export function makeResponder({ red = [redEnv()], judge = [judgeEnv()], blueSynth = [blueEnv()], blueRespond = [blueEnv()] } = {}) {
  const take = (q) => (q.length > 1 ? q.shift() : q[0])
  return (prompt, opts) => {
    const label = opts.label || ''
    if (label.startsWith('red-merge')) return take(red)
    if (label.startsWith('judge')) return take(judge)
    if (label.startsWith('blue-synthesize')) return take(blueSynth)
    if (label.startsWith('blue-respond')) return take(blueRespond)
    return 'synopsis'
  }
}

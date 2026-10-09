// Zero-token simulator for the debate engine (skills/research-protocol/scripts/debate.js).
//
// The Workflow harness evaluates the script body as async code with agent / parallel /
// pipeline / log / phase / args / budget in scope. This harness reproduces that contract
// so every control-flow branch — args parsing, the round loop, the contested docket,
// deadlock, the safety ceiling, per-role model routing, the unresolved-sitting-record list — is
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
  manifest: ['G-00000001'],
  open_questions: [], ...over,
})
// THE CHAIR RELAYS THE RECORD (plans/roundless.md §III.B.1): its envelope carries the plan `dispatch
// next` printed. A stubbed chair is a stubbed plan; the default sequence is one sitting that engages
// the evidence lens and blue on G-00000001, then one with nobody ready and PASS permitted, on which the chair
// records PASS — the shortest VERIFIED run.
// The bench's party carries what it is convened for, as `dispatch next` prints it: `party('judge',
// …gaps)` is the bench on its docket, `petitionBench()` the bench convened to hear petitions.
export const party = (seat_id, ...gap_ids) => (seat_id === 'judge' ? { seat_id, gap_ids, occasions: ['docket'] } : { seat_id, gap_ids })
export const petitionBench = () => ({ seat_id: 'judge', gap_ids: [], occasions: ['petition'] })
// blocker is one entry of the plan's `blockers`: what holds a PASS and the seat whose act clears it.
export const blocker = (subject, owner, kind = 'unruled_motion') => ({ kind, subject, owner })
export const plan = (parties = [], over = {}) => ({ head: 2, parties, docket: [], remand_owed: [], pass_permitted: false, ceiling: false, max_epochs: 0, epoch_limit_reached: false, why: [], stale_areas: [], blockers: [], ...over })
export const passPlan = (over = {}) => plan([], { pass_permitted: true, ...over })
export const ceilingPlan = (over = {}) => plan([], { ceiling: true, why: ['G-00000001: at impasse, ruled remanded — at its limit'], ...over })
export const chairEnv = (over = {}) => ({ plan: plan([party('red-lens-evidence'), party('blue-respond', 'G-00000001')]), ...over })
export const passChair = (over = {}) => chairEnv({ plan: passPlan(), verdict: 'PASS', ...over })
export const gap = (id, over = {}) => ({
  id, location: 'loc', problem: 'p', required_fix: 'f', acceptance_check: 'grep the corrected figure at the anchor', existence: 'verified',
  severity: 'medium', likelihood: 'medium', impact: 'medium', complexity_cost: 'low', ...over,
})
export const judgeEnv = (over = {}) => ({ dispositions: [], ...over })
export const petitionRulingEnv = (over = {}) => ({ rulings: [{ petitioner: 'x', class: 'ethical', ruling: 'denied' }], ...over })
// makeResponder serves envelopes by seat, in order; the last one repeats. Lenses answer in free text.
export const assembleEnv = (over = {}) => ({ synopsis: 'synopsis', open_gaps: 0, ...over })
export function makeResponder({ chair = [chairEnv(), passChair()], judge = [judgeEnv()], blueSynth = [blueEnv()], blueRespond = [blueEnv()], petition = [petitionRulingEnv()], assemble = [assembleEnv()] } = {}) {
  const take = (q) => (q.length > 1 ? q.shift() : q[0])
  return (prompt, opts) => {
    const label = opts.label || ''
    // ROUTE ON THE QUESTION, NOT THE SEAT. The bench is one seat now, so `judge` heads the label
    // of four different sittings and a prefix match cannot tell them apart. What distinguishes
    // them is what the engine ASKED — which the label carries after the head, and which the real
    // dispatch carries as the envelope schema it demands back. Matching the head alone sent the
    // assembly sitting a judge envelope with no open_gaps in it, and the run reported null.
    if (label.startsWith('red-chair')) return take(chair)
    if (/^judge · petition/.test(label)) return take(petition)
    if (/^judge · assemble/.test(label)) return take(assemble)
    if (label.startsWith('judge')) return take(judge)
    if (label.startsWith('blue-synthesize')) return take(blueSynth)
    if (label.startsWith('blue-respond')) return take(blueRespond)
    return 'synopsis'
  }
}

// THE ENVELOPE SCHEMA, APPLIED AS THE WORKFLOW HARNESS APPLIES IT. The real harness validates a
// seat's envelope against the schema its dispatch carries and hands a miss back to the seat; this
// stub world returns whatever a test cans, so a relay the schema would have refused reaches the
// engine here and one it accepts is never known to be accepted. schemaAccepts answers that
// question for the keywords debate.js's envelopes use.
//
// IT REFUSES A KEYWORD IT DOES NOT IMPLEMENT. A validator that skips what it cannot read accepts
// everything that keyword would have refused, and reports agreement between a schema and a check
// it never applied.
const SCHEMA_KEYWORDS = new Set(['type', 'required', 'properties', 'items', 'enum', 'anyOf', 'minItems', 'maxItems', 'minimum', 'pattern', 'additionalProperties', 'description'])
const jsonTypeOf = (v) => (v === null ? 'null' : Array.isArray(v) ? 'array' : Number.isInteger(v) ? 'integer' : typeof v)
export function schemaAccepts(schema, v) {
  for (const k of Object.keys(schema)) if (!SCHEMA_KEYWORDS.has(k)) throw new Error(`schemaAccepts: the schema uses \`${k}\`, which this validator does not implement — teach it the keyword before trusting its answer`)
  if (schema.anyOf && !schema.anyOf.some((s) => schemaAccepts(s, v))) return false
  const t = jsonTypeOf(v)
  if (schema.type !== undefined) {
    const types = Array.isArray(schema.type) ? schema.type : [schema.type]
    if (!types.some((want) => want === t || (want === 'number' && t === 'integer'))) return false
  }
  if (schema.enum && !schema.enum.includes(v)) return false
  if (t === 'string' && schema.pattern !== undefined && !new RegExp(schema.pattern).test(v)) return false
  if ((t === 'integer' || t === 'number') && schema.minimum !== undefined && v < schema.minimum) return false
  if (t === 'array') {
    if (schema.minItems !== undefined && v.length < schema.minItems) return false
    if (schema.maxItems !== undefined && v.length > schema.maxItems) return false
    if (schema.items && !v.every((x) => schemaAccepts(schema.items, x))) return false
  }
  if (t === 'object') {
    for (const r of schema.required || []) if (v[r] === undefined) return false
    for (const [k, x] of Object.entries(v)) {
      const sub = (schema.properties || {})[k]
      if (sub) { if (x !== undefined && !schemaAccepts(sub, x)) return false } else if (schema.additionalProperties === false) return false
    }
  }
  return true
}

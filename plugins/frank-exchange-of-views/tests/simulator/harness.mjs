// Zero-token simulator for the debate engine (skills/research-protocol/scripts/debate.js).
//
// The Workflow harness evaluates the script body as async code with agent / parallel /
// pipeline / log / phase / args / budget in scope. This harness reproduces that contract
// so every control-flow branch — args parsing, the round loop, the contested docket,
// deadlock, the safety ceiling, per-role model routing — is
// exercised with canned envelopes and no live agents.
//
// Founding regressions (each cost a live run to discover):
//   1. stringified args on resume -> destructured undefined -> literal 'undefined' paths (run 1)
//   2. agent() returning null on terminal failure -> TypeError on redEnv.verdict (run 2)
import { createHash } from 'node:crypto'
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
export const blueEnv = (over = {}) => ({ manifest: ['G-00000001'], ...over })
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
// exchanges is the plan's entry for each gap its parties are engaged on, in the order first named,
// as the verb prints it: n counted on each, none unresolved.
export const exchanges = (parties, n = 0) => [...new Set(parties.flatMap((p) => p.gap_ids))].map((gap_id) => ({ gap_id, count: n, unresolved: 0 }))
// A PLAN THAT PERMITS NO PASS NAMES WHAT HOLDS IT, as the verb prints it: each party is readied by
// a blocker on the gate's list. A stub's default is one per party, of the kinds the engine does not
// read — a material gap per gap it is engaged on, a ready lens for a party with none — so a test that
// does not state its blockers relays a plan the schema accepts and the engine routes as before. A
// plan with no party and no PASS states its own.
const heldBy = (parties) => parties.flatMap((p) => (p.gap_ids.length ? p.gap_ids.map((g) => ({ kind: 'material_gap', subject: g, owner: p.seat_id })) : [{ kind: 'lens_ready', subject: p.seat_id, owner: p.seat_id }]))
// withKey is a stub plan carrying its key, as the verb prints one: equal for two stubs that are equal in
// every field. A stub's epoch is its own fixed 1 whatever sitting relays it, so the key a stub carries
// is the same in two sittings exactly when the record's would be. A test that states a key keeps it.
export const withKey = (p) => ({ key: createHash('sha256').update(JSON.stringify({ ...p, key: undefined })).digest('hex').slice(0, 16), ...p })
export const plan = (parties = [], over = {}) => withKey({ head: 2, parties, docket: [], remand_owed: [], exchanges: exchanges(parties), pass_permitted: false, ceiling: false, max_epochs: 0, epoch: 1, final_sitting: false, epoch_limit_reached: false, why: [], stale_areas: [], blockers: over.pass_permitted ? [] : heldBy(parties), ...over })
export const passPlan = (over = {}) => plan([], { pass_permitted: true, ...over })
// ceilingPlan is the plan at the run's epoch limit relayed in the first sitting: the one ceiling the
// record states.
export const ceilingPlan = (over = {}) => plan([], { ceiling: true, epoch_limit_reached: true, max_epochs: 1, why: ["epoch limit 1 reached — this chair sitting opens the run's last epoch, so the 2 party(ies) above are not dispatched"], blockers: [{ kind: 'material_gap', subject: 'G-00000001', owner: 'red-lens-evidence' }], ...over })
// chairEnv is a chair's envelope: the envelope IS the plan, so the plan given as `plan` is spread to
// its top level, and every other key given (the plan's `verdict`, the chair's `notes`) sits beside it.
export const chairEnv = ({ plan: p = plan([party('red-lens-evidence'), party('blue-respond', 'G-00000001')]), ...over } = {}) => ({ ...p, ...over })
export const passChair = (over = {}) => chairEnv({ plan: passPlan(), verdict: 'PASS', ...over })
export const gap = (id, over = {}) => ({
  id, location: 'loc', problem: 'p', required_fix: 'f', acceptance_check: 'grep the corrected figure at the anchor', existence: 'verified',
  severity: 'medium', likelihood: 'medium', impact: 'medium', complexity_cost: 'low', ...over,
})
export const judgeEnv = (over = {}) => ({ dispositions: [], ...over })
export const petitionRulingEnv = (over = {}) => ({ rulings: [{ petitioner: 'x', class: 'ethical', ruling: 'denied' }], ...over })
// makeResponder serves envelopes by seat, in order; the last one repeats. A lens returns its empty
// envelope; the frontier and the lanes answer in free text.
export const assembleEnv = (over = {}) => ({ open_gaps: 0, ...over })
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
    if (label.startsWith('red-lens')) return {}
    return 'sat'
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
const SCHEMA_KEYWORDS = new Set(['type', 'required', 'properties', 'items', 'enum', 'anyOf', 'if', 'then', 'else', 'minItems', 'maxItems', 'minimum', 'pattern', 'additionalProperties', 'description'])
const jsonTypeOf = (v) => (v === null ? 'null' : Array.isArray(v) ? 'array' : Number.isInteger(v) ? 'integer' : typeof v)
export function schemaAccepts(schema, v) {
  for (const k of Object.keys(schema)) if (!SCHEMA_KEYWORDS.has(k)) throw new Error(`schemaAccepts: the schema uses \`${k}\`, which this validator does not implement — teach it the keyword before trusting its answer`)
  if (schema.anyOf && !schema.anyOf.some((s) => schemaAccepts(s, v))) return false
  if (schema.if) {
    const branch = schemaAccepts(schema.if, v) ? schema.then : schema.else
    if (branch && !schemaAccepts(branch, v)) return false
  }
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

// THE STRICT-TYPES COMPILE THE WORKFLOW HARNESS RUNS ON A DISPATCH'S SCHEMA, as ajv@8 runs it by
// default (strictTypes: "log", no union types). The harness LOGS each miss at launch and compiles
// the schema anyway, so a schema it warns on still validates — and the warning is the only sign
// that a keyword sits where no `type` licenses it. The rules, from ajv's checkStrictTypes:
//   - a keyword that applies to one JSON type (minItems to arrays, required to objects, …) needs
//     that type in the schema's `type`, or in the type of the schema it is applied beside;
//   - an anyOf branch, and an if/then/else branch, reads the same value, so it inherits its parent's type; a property's or an
//     item's schema reads a new value, and inherits none;
//   - a branch's `type` must be one its context allows;
//   - a `type` names one type, or one type and null.
// It returns ajv's own message for each miss, with the path ajv prints. Like schemaAccepts, it
// refuses a keyword it does not implement rather than pass what it could not read.
const KEYWORD_TYPE = { minItems: 'array', maxItems: 'array', items: 'array', required: 'object', properties: 'object', additionalProperties: 'object', minimum: 'number', pattern: 'string' }
const includesType = (ts, t) => ts.includes(t) || (t === 'integer' && ts.includes('number'))
const hasApplicableType = (ts, t) => ts.includes(t) || (t === 'number' && ts.includes('integer'))
export function strictTypesWarnings(schema, path = '#', context = []) {
  const warn = []
  for (const k of Object.keys(schema)) if (!SCHEMA_KEYWORDS.has(k)) throw new Error(`strictTypesWarnings: the schema uses \`${k}\` at ${path}, which this check does not implement — teach it the keyword before trusting its answer`)
  const types = schema.type === undefined ? [] : Array.isArray(schema.type) ? schema.type : [schema.type]
  let dataTypes = context
  if (types.length) {
    if (!context.length) dataTypes = types
    else {
      for (const t of types) if (!includesType(context, t)) warn.push(`strict mode: type "${t}" not allowed by context "${context.join(',')}" at "${path}" (strictTypes)`)
      dataTypes = context.filter((t) => includesType(types, t))
    }
  }
  if (types.length > 1 && !(types.length === 2 && types.includes('null'))) warn.push(`strict mode: use allowUnionTypes to allow union type keyword at "${path}" (strictTypes)`)
  for (const [k, t] of Object.entries(KEYWORD_TYPE)) {
    if (schema[k] !== undefined && !hasApplicableType(dataTypes, t)) warn.push(`strict mode: missing type "${t}" for keyword "${k}" at "${path}" (strictTypes)`)
  }
  if (Array.isArray(schema.items)) throw new Error(`strictTypesWarnings: tuple \`items\` at ${path} is not implemented`)
  if (schema.items) warn.push(...strictTypesWarnings(schema.items, `${path}/items`))
  for (const [k, sub] of Object.entries(schema.properties || {})) warn.push(...strictTypesWarnings(sub, `${path}/properties/${k}`))
  if (schema.additionalProperties && typeof schema.additionalProperties === 'object') warn.push(...strictTypesWarnings(schema.additionalProperties, `${path}/additionalProperties`))
  for (const [i, sub] of (schema.anyOf || []).entries()) warn.push(...strictTypesWarnings(sub, `${path}/anyOf/${i}`, dataTypes))
  for (const k of ['if', 'then', 'else']) if (schema[k]) warn.push(...strictTypesWarnings(schema[k], `${path}/${k}`, dataTypes))
  return warn
}

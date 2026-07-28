#!/usr/bin/env node
// Run capture — the mechanical half of /research's run-record step, PLUS the mechanized
// post-hoc auditor. The run-4 report's attestation ceiling (§6.2) names "independent seats
// over git-tracked artifacts" as the enforcement tier for vacuity — a seat asserting work it
// did not do. This script IS that named auditor for the mechanical checks: it recomputes
// counts from the FILES and diffs them against the envelopes' self-reports.
//
// Usage: node capture-research-run.mjs <runDir> <workflow-transcript-dir>
// Writes <runDir>/run-record-audit.md, <runDir>/cost.md, trajectories/journal.jsonl (copy),
// trajectories/agent-transcripts.tar.gz; removes the .run-live marker.
import { existsSync, mkdirSync, readFileSync, writeFileSync, copyFileSync, readdirSync, rmSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { spawnSync, execFileSync } from 'node:child_process'
import { pathToFileURL, fileURLToPath } from 'node:url'
import { computeScorecards, renderChair, chairHeader } from './scorecards.mjs'
import { scanTranscript } from './cost-audit.mjs'
import { tierMismatch } from './model-guard.mjs'

// Tolerant journal walk: collect every result object and any friction arrays inside.
export function readJournal(transcriptDir) {
  const p = join(transcriptDir, 'journal.jsonl')
  if (!existsSync(p)) return { results: [], friction: [] }
  const results = [], friction = []
  for (const line of readFileSync(p, 'utf8').split('\n')) {
    if (!line.trim()) continue
    let j; try { j = JSON.parse(line) } catch { continue }
    const r = j.result
    if (r && typeof r === 'object') {
      results.push(r)
      if (Array.isArray(r.friction)) friction.push(...r.friction)
    }
  }
  return { results, friction }
}

// makeViewReader binds a record-tool binary + runDir into a reader the audits call for a
// STRUCTURED view: `viewJSON('debate', ['--json'])`, `viewJSON('friction')`. It is the same
// json-mode move the dashboard made (#123) — the record is read once, through the tool,
// never by re-parsing a markdown projection. Null when no --bin was given, so the audits
// that depend on it SKIP rather than reading a stale file or crashing.
export function makeViewReader(bin, runDir) {
  if (!bin) return null
  return (view, extraFlags = []) => {
    const out = execFileSync(bin, ['merge', 'show', '--view', view, ...extraFlags, '--run', runDir], { encoding: 'utf8' })
    return JSON.parse(out)
  }
}

// AUDIT 1 — telemetry presence: one telemetry line per red round (a missing line is caught
// here; a WRONG line is beyond any presence check — recompute on actuation, per §2.5). The
// red-round count comes from the STRUCTURED debate view (rounds with a red sitting), not from
// regexing a markdown transcript — the second-reader defect json-mode removes.
export function telemetryAudit(runDir, { viewJSON } = {}) {
  if (!viewJSON) return { check: 'telemetry', verdict: 'SKIP', detail: 'needs --bin (the debate view is read through the tool)' }
  let redRounds
  try {
    const debate = viewJSON('debate', ['--json'])
    redRounds = (debate.rounds || []).filter((r) => (r.red || []).length > 0).length
  } catch (e) {
    return { check: 'telemetry', verdict: 'SKIP', detail: `debate view unavailable: ${String((e && e.message) || e).slice(0, 120)}` }
  }
  const p = join(runDir, 'trajectories', 'board-telemetry.jsonl')
  if (!existsSync(p)) {
    return { check: 'telemetry', verdict: redRounds === 0 ? 'SKIP' : 'FAIL', detail: `board-telemetry.jsonl absent; the debate record shows ${redRounds} red round(s)` }
  }
  const lines = readFileSync(p, 'utf8').split('\n').filter(Boolean)
  const rounds = new Set(lines.map((l) => { try { return JSON.parse(l).round } catch { return null } }).filter((r) => r != null))
  return {
    check: 'telemetry',
    verdict: rounds.size >= redRounds ? 'PASS' : 'FAIL',
    detail: `${rounds.size} telemetry round(s) vs ${redRounds} red round(s) on the record`,
  }
}

// AUDIT 2 — shard self-report vs files: the closure index (`id | class | summary | supersedes`
// lines in ledger.md) and heading-anchored archive records, recounted from disk and diffed
// against the last red envelope's integers. Heuristic counts, labeled as such.
export function shardAudit(runDir, results) {
  const ledgerP = join(runDir, 'red', 'ledger.md')
  const archiveP = join(runDir, 'red', 'archive.md')
  if (!existsSync(ledgerP) && !existsSync(archiveP)) return { check: 'shards', verdict: 'SKIP', detail: 'no ledger/archive (pre-sharding run)' }
  const idLine = /R\d+-\d+\s*\|/
  const ledgerCount = existsSync(ledgerP) ? readFileSync(ledgerP, 'utf8').split('\n').filter((l) => idLine.test(l)).length : 0
  const archiveCount = existsSync(archiveP) ? (readFileSync(archiveP, 'utf8').match(/^#{1,4}\s+.*R\d+-\d+/gm) || []).length : 0
  const lastRed = [...results].reverse().find((r) => typeof r.ledger_closure_lines === 'number' || typeof r.archive_blocks === 'number')
  const self = lastRed ? `self-reported ${lastRed.ledger_closure_lines}/${lastRed.archive_blocks}` : 'no envelope self-report found in journal'
  const consistent = !lastRed || (lastRed.ledger_closure_lines === ledgerCount && lastRed.archive_blocks === archiveCount)
  return {
    check: 'shards',
    verdict: consistent ? 'PASS' : 'FAIL',
    detail: `measured (heuristic) closure-index lines=${ledgerCount}, archive records=${archiveCount}; ${self}`,
  }
}

// W1.2 — friction parity. Every friction a seat SELF-REPORTS in its envelope must also have
// reached the RECORD (the friction verb — the durable channel that survives a phase abort).
// An envelope entry the record never received is a first-class pipeline error (the seat
// reported friction it never recorded), so it FAILs — no longer a REPAIRED harvest-into-a-
// file. friction.md is retired: the record IS the friction channel, read here through the
// friction view. (The prior harvest-into-friction.md closed a lens/abort loss class by hand;
// with the verb universal to every role and the file gone, that repair is obsolete.)
export function frictionAudit(runDir, envelopeFriction, { viewJSON } = {}) {
  if (!viewJSON) return { check: 'friction-parity', verdict: 'SKIP', detail: 'needs --bin (the friction record is read through the tool)' }
  let onRecord
  try {
    const fj = viewJSON('friction')
    onRecord = (fj.friction || []).map((f) => String(f.text || ''))
  } catch (e) {
    return { check: 'friction-parity', verdict: 'SKIP', detail: `friction view unavailable: ${String((e && e.message) || e).slice(0, 120)}` }
  }
  // Tolerant match either direction: the envelope string may be "seat: text" while the record
  // event carries just the text (or vice versa) — a shared 60-char prefix is the same entry.
  const present = (f) => {
    const s = String(f)
    return onRecord.some((t) => t.includes(s.slice(0, 60)) || s.includes(t.slice(0, 60)))
  }
  const missing = envelopeFriction.filter((f) => !present(f))
  return {
    check: 'friction-parity',
    verdict: missing.length ? 'FAIL' : 'PASS',
    detail: missing.length
      ? `${missing.length} envelope friction entr${missing.length === 1 ? 'y' : 'ies'} never reached the record (should have been recorded via the friction verb during the run):\n${missing.map((m) => `    - ${String(m).slice(0, 120)}`).join('\n')}`
      : `${envelopeFriction.length} envelope friction entr${envelopeFriction.length === 1 ? 'y' : 'ies'} all present on the record`,
  }
}

// W1.3 — context-use telemetry: per-seat peak prompt size vs the model's window. Baseline
// (146-transcript scan, runs 3-5): zero compaction ever, peak 27% of the 1M window. The 50%
// flag is the tripwire that says within-seat checkpointing has become a real need; the
// 200k-window constraint is why judgment seats must never ride a haiku-tier model.
export function contextUse(transcriptDir, agentFiles) {
  const peaks = []
  for (const f of agentFiles) {
    let peak = 0, model = ''
    for (const line of readFileSync(join(transcriptDir, f), 'utf8').split('\n')) {
      if (!line.trim()) continue
      let j; try { j = JSON.parse(line) } catch { continue }
      const m = j.message
      if (m && m.usage) {
        const u = m.usage
        const ctx = (u.input_tokens || 0) + (u.cache_read_input_tokens || 0) + (u.cache_creation_input_tokens || 0)
        if (ctx > peak) { peak = ctx; if (m.model) model = m.model }
      }
    }
    if (peak) peaks.push({ agent: f.replace(/^agent-|\.jsonl$/g, ''), peak, window: /haiku/.test(model) ? 200_000 : 1_000_000 })
  }
  if (!peaks.length) return { check: 'context-use', verdict: 'SKIP', detail: 'no usage records found' }
  peaks.sort((a, b) => b.peak / b.window - a.peak / a.window)
  const top = peaks[0]
  const flagged = peaks.filter((p) => p.peak / p.window > 0.5)
  return {
    check: 'context-use',
    verdict: flagged.length ? 'WARN' : 'PASS',
    detail: `peak ${(top.peak / 1000).toFixed(0)}k = ${(top.peak / top.window * 100).toFixed(0)}% of its ${top.window / 1000}k window (agent ${top.agent}); ${peaks.length} seats measured; ${flagged.length} over the 50% tripwire${flagged.length ? ':\n' + flagged.map((p) => `    - ${p.agent}: ${(p.peak / 1000).toFixed(0)}k / ${p.window / 1000}k`).join('\n') : ''}`,
  }
}

// W1.4 — assembly regression screen (v1). The catechism audit (post-capture, run 5) found the
// assembly seat re-minting REFUTED round-0 phrasings from recall. Mechanical v1: harvest
// distinctive tokens (issue numbers, arXiv ids) from citation-ledger rows graded REFUTED and
// grep the ASSEMBLY-OWNED text (report.md above the embedded blue report) for them — a hit is
// a candidate regression flagged WARN for human eyes, never auto-judged (the assembly may
// carry the corrected form). The full screen (propagation-list grep) arrives with the record
// layer, which is where repaired-phrasing lists become structured.
export function assemblyScreen(runDir) {
  const reportP = join(runDir, 'report.md')
  const ledgerP = join(runDir, 'red', 'citation-ledger.md')
  if (!existsSync(reportP) || !existsSync(ledgerP)) return { check: 'assembly-screen', verdict: 'SKIP', detail: 'report.md or citation-ledger.md absent' }
  const report = readFileSync(reportP, 'utf8')
  const cut = report.search(/^## Blue team report/m)
  const assembly = cut > 0 ? report.slice(0, cut) : report
  const tokens = new Set()
  for (const row of readFileSync(ledgerP, 'utf8').split('\n')) {
    if (!/REFUTED|ABSENT/i.test(row)) continue
    for (const t of row.match(/#\d{4,6}/g) || []) tokens.add(t)
    for (const t of row.match(/arXiv:\d{4}\.\d{4,5}/gi) || []) tokens.add(t)
  }
  const hits = [...tokens].filter((t) => assembly.includes(t))
  return {
    check: 'assembly-screen',
    verdict: hits.length ? 'WARN' : 'PASS',
    detail: hits.length
      ? `${hits.length} REFUTED-row token(s) appear in assembly-owned text — verify each carries the CORRECTED form, not the round-0 one: ${hits.join(', ')}`
      : `${tokens.size} REFUTED-row token(s) screened against assembly-owned text; no hits`,
  }
}

// W1.7 (post-hoc half) — record parity: blue's in-run attestation (round_record_appended)
// is shape-tier; this recount is the vacuity check. Every red round should have a blue
// sitting on the record and a CHANGELOG round entry (the final round may lack both on a PASS
// exit — blue never responded — so the floor is redRounds - 1). Red/blue sittings come from
// the STRUCTURED debate view (rounds with a red/blue section), not from regexing debate.md;
// the CHANGELOG is blue's genuinely hand-written revision record, so it stays a file read.
export function recordParityAudit(runDir, { viewJSON } = {}) {
  if (!viewJSON) return { check: 'record-parity', verdict: 'SKIP', detail: 'needs --bin (the debate view is read through the tool)' }
  const clP = join(runDir, 'blue', 'CHANGELOG.md')
  let redRounds, blueBlocks
  try {
    const debate = viewJSON('debate', ['--json'])
    const rounds = debate.rounds || []
    redRounds = rounds.filter((r) => (r.red || []).length > 0).length
    blueBlocks = rounds.filter((r) => (r.blue || []).length > 0).length
  } catch (e) {
    return { check: 'record-parity', verdict: 'SKIP', detail: `debate view unavailable: ${String((e && e.message) || e).slice(0, 120)}` }
  }
  if (redRounds === 0) return { check: 'record-parity', verdict: 'SKIP', detail: 'no red rounds on record' }
  const clRounds = existsSync(clP) ? new Set((readFileSync(clP, 'utf8').match(/^#+.*Round \d+/gim) || []).map((h) => h.match(/Round (\d+)/i)[1])).size : 0
  const ok = blueBlocks >= redRounds - 1 && clRounds >= redRounds - 1
  return {
    check: 'record-parity',
    verdict: ok ? 'PASS' : 'FAIL',
    detail: `${redRounds} red round(s) vs ${blueBlocks} blue sitting(s) and ${clRounds} CHANGELOG round entr${clRounds === 1 ? 'y' : 'ies'} (floor: redRounds-1 — a PASS exit has no final blue response)`,
  }
}

// W2/record-tool — the record-join audit (plan §II ENFORCEMENT, behavioral tier):
// every event must trace to a tool invocation in the emitting seat's transcript.
// Transcripts self-identify via --seat-id in their own Bash commands; events with
// no matching (seatId, verb) invocation anywhere are FLAGGED (hand-appended
// lines, back-filled records). Back-fill vacuity: a seat whose record commands
// all cluster at its transcript tail gets a WARN for human review.
export function recordJoinAudit(runDir, transcriptDir, agentFiles) {
  const recDir = join(runDir, 'records')
  if (!existsSync(recDir)) return { check: 'record-join', verdict: 'SKIP', detail: 'no records/ (pre-record-tool run)' }
  const events = []
  for (const f of readdirSync(recDir).filter((x) => x.startsWith('events-') && x.endsWith('.jsonl'))) {
    for (const l of readFileSync(join(recDir, f), 'utf8').split('\n').filter(Boolean)) {
      try { const e = JSON.parse(l); if (e.type !== 'register') events.push(e) } catch {}
    }
  }
  if (!events.length) return { check: 'record-join', verdict: 'SKIP', detail: 'records/ empty' }
  // Per-transcript: the seatIds it claims and the verbs it invoked, plus tail clustering.
  const invocations = new Set() // `${seatId}::${verb}`
  const tailWarns = []
  for (const f of agentFiles) {
    const cmds = []
    let toolCallCount = 0
    for (const l of readFileSync(join(transcriptDir, f), 'utf8').split('\n')) {
      if (!l.includes('tool_use')) continue
      let j; try { j = JSON.parse(l) } catch { continue }
      const content = j.message && j.message.content
      if (!Array.isArray(content)) continue
      for (const c of content) {
        if (c.type !== 'tool_use') continue
        toolCallCount++
        const cmd = c.input && c.input.command ? String(c.input.command) : ''
        // The record tool is now the compiled binary with ROLE SUBCOMMANDS
        // (`feov-record merge mint ...`), not four mjs scripts
        // (`node tools/red-merge.mjs mint ...`). Both shapes are matched: the
        // binary form is what seats emit from R2g on, and the mjs form is kept
        // so this audit still reads any transcript captured before the port.
        // Silently matching neither is the failure that matters — the join audit
        // would report every event as an orphan, or (worse) report PASS over an
        // empty invocation set.
        // The engine QUOTES the binary path ("<binDir>/feov-record" lens register)
        // because a plugin root can contain spaces, so the closing quote sits
        // between the binary name and the role. Requiring whitespace there
        // matched nothing in a real transcript — the audit would have reported
        // every event as an orphan while looking entirely healthy.
        const m = cmd.match(/feov-record"?\s+(?:lens|merge|blue|bench)\s+([a-z-]+)[\s\S]*?--seat-id\s+(\S+)/) ||
          cmd.match(/(?:red-lens|red-merge|blue|bench)\.mjs\s+([a-z-]+)[\s\S]*?--seat-id\s+(\S+)/)
        if (m) { invocations.add(`${m[2]}::${m[1]}`); cmds.push(toolCallCount) }
      }
    }
    if (cmds.length > 5 && cmds[0] > toolCallCount - cmds.length - 2) {
      tailWarns.push(`${f}: ${cmds.length} record invocations all tail-clustered (back-fill pattern — parity may be vacuous for this seat)`)
    }
  }
  const orphans = events.filter((e) => !invocations.has(`${e.seatId}::${verbOf(e.type)}`))
  const detail = [
    `${events.length} event(s) across shards; ${invocations.size} distinct (seat, verb) invocations in transcripts`,
    orphans.length ? `FLAGGED ${orphans.length} event(s) with no matching transcript invocation:\n${orphans.slice(0, 10).map((e) => `    - ${e.seatId} ${e.type} (${e.key})`).join('\n')}` : 'every event traces to a transcript invocation',
    ...tailWarns.map((w) => `WARN ${w}`),
  ].join('\n  ')
  return { check: 'record-join', verdict: orphans.length ? 'FAIL' : tailWarns.length ? 'WARN' : 'PASS', detail }
}
// Event types map 1:1 to CLI verbs except mint-adjacent internals.
const verbOf = (type) => (type === 'class-new' ? 'mint' : type)

// model-tier audit (#111): each seat's ACTUAL price-tier vs the tier its class was CONFIGURED
// to run on. debate.js now requires both tiers, so a current run-config carries them; a seat
// DEARER than configured is the fable trap (FAIL), CHEAPER is discounted verification (WARN).
// Reuses cost-audit's scanTranscript — capture derived model ad hoc before #111, so there is no
// inline scan to share; this is the sanctioned reuse. Judgment running TOO CHEAP is the inverse
// of capture's long-standing doctrine (judgment must never ride a haiku-tier model); the
// too-cheap-judgment case is a Non-goal of #111 and is not separately gated here.
export function modelTierAudit(runDir, transcriptDir, agentFiles) {
  let cfg = {}
  try { cfg = JSON.parse(readFileSync(join(runDir, 'inputs', 'run-config.json'), 'utf8')) } catch {}
  if (!cfg.model && !cfg.judgmentModel) return { check: 'model-tier', verdict: 'SKIP', detail: 'no run-config models (pre-#111 run)' }
  const rows = agentFiles.map((f) => scanTranscript(readFileSync(join(transcriptDir, f), 'utf8')))
  const seen = new Set()
  const findings = tierMismatch(rows, cfg).filter((f) => {
    const k = `${f.seat}|${f.round}|${f.verdict}|${f.actual}`
    if (seen.has(k)) return false
    seen.add(k); return true
  })
  const fails = findings.filter((f) => f.verdict === 'FAIL')
  const warns = findings.filter((f) => f.verdict === 'WARN')
  const tiers = `bulk=${cfg.model || '?'}, judgment=${cfg.judgmentModel || '?'}`
  const detail = findings.length
    ? `${tiers}; ${findings.map((f) => `${f.verdict} ${f.why}`).join('; ')}`
    : `every seat ran on its configured tier (${tiers})`
  return { check: 'model-tier', verdict: fails.length ? 'FAIL' : warns.length ? 'WARN' : 'PASS', detail }
}

// W2e — precedent harvest: the run's judicial rulings become PERSUASIVE
// proposals in law/proposed/<slug>.md (repo-side, committed with the run
// record) for human promotion per law/README.md. The bench proposes; the
// human enacts. Derivation is mechanical from journal envelopes (resolutions
// + petition rulings); facts/scope stay stubs the reviewer fills from the
// cited record — the harvest never invents.
// INTEGRITY RECONCILIATION — the bench's mind-reading, mechanized.
//
// This exists because the question "did the seat actually do what it says it
// did" cannot be answered from the artifacts: a performative repair and a
// rigorous one look identical in the record, which is part of why the judicial
// gate measured inert at 86/87 carried. E0.5a did this audit BY HAND across one
// run and found ~11% of the record mechanically unauditable by format. This is
// that audit as a script.
//
// WHY IT LIVES AT CAPTURE AND NOT AT THE BENCH. Verified, not assumed: the
// workflow transcript directory is created BY the run and supplied to capture by
// the operator afterwards, so debate.js cannot know it at dispatch and a bench
// seat cannot reach it mid-run. Where the operator DOES know it (a resume), the
// engine takes transcriptDir and the bench gains live inspection; otherwise the
// reconciliation is post-hoc and its findings reach the next run.
//
// DESIGN CONSTRAINT, from the operator: production-by-the-party would propagate
// the lie and party-demanded discovery would encourage fabrication. So this reads
// the trajectory DIRECTLY and never asks the seat to hand over its own evidence.
//
// SEPARATION: this rules on whether the RECORD IS HONEST, never on the merits of
// a gap. Sloppy reasoning that reached a sound conclusion is not a finding; a
// clean conclusion contradicted by what the seat actually did is.
export function attestationAudit(runDir, transcriptDir, agentFiles, sampleFloor = 5) {
  const archive = existsSync(join(runDir, 'red', 'archive.md'))
    ? readFileSync(join(runDir, 'red', 'archive.md'), 'utf8') : ''
  const blocks = archive.split(/^## /m).slice(1)
  if (!blocks.length) return { check: 'attestation-integrity', verdict: 'SKIP', detail: 'no archive records to reconcile' }

  // What each closure CLAIMS was done: the anchor triple names a seat, a tool and
  // a target. Every one of those is checkable against the transcript.
  const claims = []
  for (const b of blocks) {
    const id = (/^(R\d+-\d+)/.exec(b) || [, '?'])[1]
    const anchor = /verification anchor:\s*(.+)/.exec(b)
    if (!anchor) continue
    if (/CARRIED from round/i.test(anchor[1])) continue // carried closures assert no fresh act
    const [seat, tool, target] = anchor[1].split('|').map((x) => x.trim())
    if (seat && tool && target) claims.push({ id, seat, tool, target })
  }
  if (!claims.length) return { check: 'attestation-integrity', verdict: 'SKIP', detail: 'no anchored closures (all carried, or anchors absent)' }

  // Every tool call any seat actually made, as text to match targets against.
  const calls = []
  for (const f of agentFiles) {
    for (const l of readFileSync(join(transcriptDir, f), 'utf8').split('\n')) {
      if (!l.trim()) continue
      let j; try { j = JSON.parse(l) } catch { continue }
      const content = j.message && j.message.content
      if (!Array.isArray(content)) continue
      for (const c of content) {
        if (c.type !== 'tool_use') continue
        calls.push({ agent: f, name: c.name || '', input: JSON.stringify(c.input || {}) })
      }
    }
  }

  // A RANDOM SPOT-CHECK FLOOR, not suspicion-only. Reconciling only what already
  // looks wrong means a seat needs merely to avoid arousing suspicion; sampling
  // regardless is what makes an unreconcilable claim risky to make at all. Same
  // shape as W1.8's archive spot-check floor.
  const sampled = claims.length <= sampleFloor ? claims : claims.filter((_, i) => i % Math.ceil(claims.length / sampleFloor) === 0)
  const unreconciled = []
  for (const c of sampled) {
    // The target is the load-bearing part: a file, ref or URL the seat says it
    // read. Match on its most distinctive fragment rather than the whole string,
    // since a seat may reach the same artifact by a different path.
    const needle = String(c.target).split(/[\s|]+/).filter((t) => t.length > 6).sort((a, b) => b.length - a.length)[0]
    if (!needle) { unreconciled.push({ ...c, why: 'anchor target too vague to reconcile' }); continue }
    if (!calls.some((k) => k.input.includes(needle))) {
      unreconciled.push({ ...c, why: `no tool call in any transcript touches ${needle}` })
    }
  }

  return {
    check: 'attestation-integrity',
    verdict: unreconciled.length ? 'FAIL' : 'PASS',
    detail: unreconciled.length
      ? `${unreconciled.length}/${sampled.length} sampled closure(s) NOT reconcilable against the trajectories — a claimed act with no matching tool call:\n` +
        unreconciled.map((u) => `    - ${u.id} (${u.seat} | ${u.tool}): ${u.why}`).join('\n') +
        '\n    This rules on whether the RECORD IS HONEST, never on the merits of the gap.'
      : `${sampled.length}/${claims.length} anchored closure(s) sampled and reconciled against actual tool calls`,
    unreconciled,
  }
}

export function harvestPrecedents(runDir, results, lawDir) {
  const slug = String(runDir).replace(/[\\/]+$/, '').split(/[\\/]/).pop()
  const rulings = []
  for (const r of results) {
    if (Array.isArray(r.resolutions)) for (const x of r.resolutions) rulings.push({ kind: 'docket', gap_id: x.gap_id, disposition: x.resolution, rationale: x.rationale })
    if (Array.isArray(r.rulings)) for (const x of r.rulings) rulings.push({ kind: 'petition', petitioner: x.petitioner, disposition: x.ruling, rationale: x.opinion })
  }
  if (!rulings.length) return { written: false, count: 0 }
  if (!existsSync(lawDir)) return { written: false, count: rulings.length, reason: 'no law/ dir at repo root' }
  mkdirSync(join(lawDir, 'proposed'), { recursive: true })
  const out = join(lawDir, 'proposed', slug + '.md')
  const body = ['# proposed holdings — ' + slug + ' [ALL PERSUASIVE — awaiting human review per law/README.md]', '',
    ...rulings.map((r, i) => [
      '## ' + slug + '-' + (i + 1) + ' [PERSUASIVE]',
      'facts: <reviewer: fill from the cited record — the harvest never invents>',
      'question: ' + (r.kind === 'docket' ? 'disposition of ' + r.gap_id : 'petition by ' + r.petitioner),
      'holding: ' + r.disposition,
      // Whole rationale — no truncation. The judge's opinion IS the precedent, and the
      // actionable tail ("Direction owed: ...") lived past the old 600-char cut, so every
      // harvested holding lost its point mid-sentence. Judge rationales are envelope-bounded;
      // storing them whole is honest and the proposed/ file is for a human to read.
      'rationale: ' + String(r.rationale || ''),
      'scope-limits: <reviewer: state what this assumed>',
      'source: ' + slug + ', ' + (r.kind === 'docket' ? r.gap_id : 'petition:' + r.petitioner),
      '',
    ].join('\n')),
  ].join('\n')
  writeFileSync(out, body)
  return { written: true, count: rulings.length, path: out }
}

// writeScorecards appends this run's rows to each chair's scorecard.
//
// The run label is the run DIRECTORY name, which is dated and unique, so the
// series reads chronologically without anyone stamping a timestamp — and without
// this script needing a clock it would have to be trusted about.
export function writeScorecards(runDir, results, memoryDir, bin) {
  if (!existsSync(memoryDir)) return { written: false, reason: `no ${memoryDir} — scorecards need the tracked memory dir` }
  // bin lets the record-backed rows (direction-uptake, citation-yield) compute at capture
  // instead of landing "not computed" in the run-over-run series — the same --bin the audits use.
  const cards = computeScorecards(runDir, results, bin)
  const label = runDir.split(/[\\/]/).filter(Boolean).pop() || 'run'
  let rows = 0
  for (const [chair, chairRows] of Object.entries(cards)) {
    const p = join(memoryDir, `${chair}-scorecard.md`)
    const head = existsSync(p) ? readFileSync(p, 'utf8') : chairHeader(chair)
    writeFileSync(p, head.replace(/\n+$/, '\n') + '\n' + renderChair(chair, chairRows, label))
    rows += chairRows.length
  }
  return { written: true, rows, chairs: Object.keys(cards).length }
}

export function capture(runDir, transcriptDir, { bin } = {}) {
  const lines = []
  // Mechanics: journal copy, transcript tarball, cost.md (with telemetry join).
  copyFileSync(join(transcriptDir, 'journal.jsonl'), join(runDir, 'trajectories', 'journal.jsonl'))
  const agentFiles = readdirSync(transcriptDir).filter((f) => f.startsWith('agent-') && f.endsWith('.jsonl'))
  // tar with a RELATIVE -f target, then move: GNU tar reads "C:" in an absolute Windows
  // path as a remote host ("Cannot connect to C:") — measured, not theoretical.
  const tmpTar = 'agent-transcripts.tar.gz'
  const tar = spawnSync('tar', ['czf', tmpTar, ...agentFiles], { cwd: transcriptDir })
  if (tar.status === 0) {
    copyFileSync(join(transcriptDir, tmpTar), join(runDir, 'trajectories', tmpTar))
    rmSync(join(transcriptDir, tmpTar))
  }
  lines.push(`tarball: ${tar.status === 0 ? `${agentFiles.length} transcript(s)` : 'FAILED — ' + (tar.stderr || '').toString().slice(0, 200)}`)
  const cost = spawnSync(process.execPath, [join(dirname(fileURLToPath(import.meta.url)), 'cost-audit.mjs'), transcriptDir, runDir])
  if (cost.status === 0) { writeFileSync(join(runDir, 'cost.md'), cost.stdout); lines.push('cost.md: written (telemetry join included)') }
  else lines.push(`cost.md: FAILED — ${(cost.stderr || '').toString().slice(0, 200)}`)

  // The audits (journal read from the copy just made — the git-tracked artifact). The three
  // record-backed audits read the run through the tool's views (viewJSON) — the debate view
  // for red/blue sittings, the friction view for recorded friction — instead of parsing a
  // markdown projection; without --bin they SKIP rather than read a stale file.
  const { results, friction } = readJournal(join(runDir, 'trajectories'))
  const viewJSON = makeViewReader(bin, runDir)
  const audits = [
    telemetryAudit(runDir, { viewJSON }),
    shardAudit(runDir, results),
    frictionAudit(runDir, friction, { viewJSON }),
    contextUse(transcriptDir, agentFiles),
    assemblyScreen(runDir),
    recordParityAudit(runDir, { viewJSON }),
    recordJoinAudit(runDir, transcriptDir, agentFiles),
    attestationAudit(runDir, transcriptDir, agentFiles),
    modelTierAudit(runDir, transcriptDir, agentFiles),
  ]

  // W2h — the visibility loop's first leg: compute every scorecard number from
  // the artifacts and APPEND it to the chair's file. Appended, never overwritten:
  // one run's number says nothing about whether a chair is improving, and the
  // run-over-run series is the only thing that does.
  const scorecards = writeScorecards(runDir, results, join(process.cwd(), 'feov-memory'), bin)
  lines.push(`scorecards: ${scorecards.written ? `${scorecards.rows} row(s) across ${scorecards.chairs} chair(s) -> feov-memory/` : scorecards.reason}`)

  const precedents = harvestPrecedents(runDir, results, join(process.cwd(), 'law'))
  lines.push(`precedent harvest: ${precedents.written ? `${precedents.count} ruling(s) -> ${precedents.path} (PERSUASIVE, awaiting review)` : precedents.count ? `${precedents.count} ruling(s), ${precedents.reason}` : 'no rulings this run'}`)

  // Marker removal: the run is no longer live; hook guards stand down.
  const marker = join(process.cwd(), '.claude', 'run-live.json')
  if (existsSync(marker)) { rmSync(marker); lines.push('run-live marker: removed') }

  const report = [
    '# capture-audit — mechanized post-hoc checks (capture-research-run.mjs)',
    '',
    'Presence/consistency tier only: these checks catch a missing line and a self-inconsistent',
    'self-report; a plausible-but-wrong value is vacuity, whose auditor is the next run/',
    'retrospective over these same git-tracked artifacts.',
    '',
    ...audits.map((a) => `- **${a.check}: ${a.verdict}** — ${a.detail}`),
    '',
    ...lines.map((l) => `- ${l}`),
    '',
  ].join('\n')
  writeFileSync(join(runDir, 'run-record-audit.md'), report)
  return { audits, report }
}

function main() {
  const argv = process.argv.slice(2)
  const binIdx = argv.indexOf('--bin')
  const bin = binIdx >= 0 ? argv[binIdx + 1] : undefined
  const positional = argv.filter((a, i) => a !== '--bin' && argv[i - 1] !== '--bin')
  const [runDir, transcriptDir] = positional
  if (!runDir || !transcriptDir) { console.error('usage: node capture-research-run.mjs <runDir> <workflow-transcript-dir> [--bin <feov-record executable>]'); process.exit(1) }
  const { audits, report } = capture(runDir, transcriptDir, { bin })
  console.log(report)
  process.exitCode = audits.some((a) => a.verdict === 'FAIL') ? 2 : 0
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main()

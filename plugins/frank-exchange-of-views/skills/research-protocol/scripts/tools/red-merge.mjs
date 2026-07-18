#!/usr/bin/env node
// tools/red-merge.mjs — the MERGE seat's record contract. The board's only writer:
// mint/close/dispose/regrade live here and nowhere else. Every closure carries its
// attestation anchor (seat + tool + target) or is labeled CARRIED — the E0.5a
// format invariant as required arguments, not prose.
import { existsSync, cpSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'
import { homedir } from 'node:os'
import { createHash } from 'node:crypto'
import { append, boardState, mintGapId, payloadText, registerSeat, render, roundOf, runVerb } from '../lib/record.mjs'

const list = (v) => (v ? String(v).split(',').map((s) => s.trim()).filter(Boolean) : [])

const verbs = {
  register: {
    help: 'FIRST ACTION at the seat: register --run <runDir> --seat-id <SEAT_ID>',
    fn: (runDir, seatId) => { const r = registerSeat(runDir, seatId); return `registered ${seatId} (shard nonce ${r.nonce})` },
  },
  mint: {
    help: 'mint a board gap (id is TOOL-assigned): --class <slug>|--class-new <slug> --definition --neighbor --distinguisher, --location "..." --problem "..."|--file --fix "..." --check "<acceptance check red runs at re-audit>" --severity/--likelihood/--impact/--cx <grade> [--supersedes R1-2,R1-7] [--found-by L5-F3,L6-F2]',
    fn: (runDir, seatId, a) => {
      const gap_id = mintGapId(runDir, roundOf(seatId))
      append(runDir, seatId, 'mint', {
        gap_id, class: a['class-new'] || a.class, class_new: !!a['class-new'],
        definition: a.definition, neighbor: a.neighbor, distinguisher: a.distinguisher,
        location: a.location, problem: a.problem || payloadText(a), required_fix: a.fix,
        acceptance_check: a.check,
        severity: a.severity, likelihood: a.likelihood, impact: a.impact, complexity_cost: a.cx,
        supersedes: list(a.supersedes), found_by: list(a['found-by']),
      })
      if (a['class-new']) append(runDir, seatId, 'class-new', { slug: a['class-new'], definition: a.definition, neighbor: a.neighbor, distinguisher: a.distinguisher })
      return `minted ${gap_id}`
    },
  },
  close: {
    help: 'close a gap WITH its verification anchor: --id R2-3 --as closed|closed_with_regression|... (--anchor-seat L1 --anchor-tool "git show" --anchor-target "7bc501e:path" | --carried-from <round>) [--successor R3-1] [--prose-file f]',
    fn: (runDir, seatId, a) => {
      append(runDir, seatId, 'close', {
        gap_id: a.id, closure_class: a.as || 'closed',
        anchor_seat: a['anchor-seat'], anchor_tool: a['anchor-tool'], anchor_target: a['anchor-target'],
        carried_from: a['carried-from'], successor: a.successor, prose: a['prose-file'] ? payloadText({ file: a['prose-file'] }) : undefined,
      })
      return `closed ${a.id} (${a.as || 'closed'})`
    },
  },
  dispose: {
    help: 'every lens observation gets a fate: --observation <label-or-key> --as minted-as|folded-into|declined|banked [--into R2-4] [--reason "..."]',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'dispose', { observation: a.observation, disposition: a.as, into: a.into, reason: a.reason }); return `disposed ${a.observation}: ${a.as}` },
  },
  regrade: {
    help: 'same-id grade movement, recorded with its reason: --id R2-5 [--severity/--likelihood/--impact/--cx <grade>] --basis "..."',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'regrade', { gap_id: a.id, severity: a.severity, likelihood: a.likelihood, impact: a.impact, complexity_cost: a.cx, basis: a.basis }); return `regraded ${a.id}` },
  },
  'dispute-respond': {
    help: "red's answer to a blue grade dispute: --id <gap> --as accepted|rejected --rationale \"...\"",
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'dispute-respond', { gap_id: a.id, response: a.as, rationale: a.rationale }); return `dispute on ${a.id}: ${a.as}` },
  },
  'spot-check': {
    help: 'the round archive spot-check record (W1.8 duty): --ids R1-4,R2-7 [--notes "..."]',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'spot-check', { ids: list(a.ids), notes: a.notes }); return `spot-checked ${list(a.ids).join(', ')}` },
  },
  position: {
    help: "the round's ### RED section (prose via --file)",
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'position', { text: payloadText(a) }); return 'position recorded' },
  },
  closing: {
    help: 'a ### RED CLOSING entry per docketed gap: --id <gap> --file <prose>',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'closing', { gap_id: a.id, text: payloadText(a) }); return `closing filed for ${a.id}` },
  },
  verdict: {
    help: "the seat's terminal act: --as PASS|FAIL — renders all projections and checkpoints records/ to the recovery mirror",
    fn: (runDir, seatId, a) => {
      append(runDir, seatId, 'verdict', { verdict: a.as })
      const r = render(runDir)
      const hash = createHash('sha1').update(runDir).digest('hex').slice(0, 12)
      const mirror = join(homedir(), '.cache', 'feov', 'run-mirror', hash)
      mkdirSync(mirror, { recursive: true })
      if (existsSync(join(runDir, 'records'))) cpSync(join(runDir, 'records'), mirror, { recursive: true })
      return `verdict ${a.as} — rendered (${r.open} open, ${r.closed} closed) and checkpointed to ${mirror}`
    },
  },
  friction: {
    help: 'attributed friction: --text|--file',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'friction', { text: payloadText(a) }); return 'friction recorded' },
  },
  petition: {
    help: 'petition the bench: --class ethical|safety|integrity|constitutional --basis "..." --relief "..."',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'petition', { class: a.class, basis: a.basis, relief: a.relief }); return `petition filed (${a.class})` },
  },
}

process.exitCode = runVerb('red-merge.mjs', verbs, process.argv.slice(2))

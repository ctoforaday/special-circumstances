#!/usr/bin/env node
// tools/red-lens.mjs — the LENS seat's record contract. The verb set IS the role
// boundary: a lens structurally cannot mint or close board gaps (no such verbs) —
// findings are lens-scoped observations the MERGE disposes and mints from.
import { append, registerSeat, payloadText, runVerb } from '../lib/record.mjs'

const verbs = {
  register: {
    help: 'FIRST ACTION at the seat: register --run <runDir> --seat-id <SEAT_ID from your prompt>',
    fn: (runDir, seatId) => { const r = registerSeat(runDir, seatId); return `registered ${seatId} (shard nonce ${r.nonce})` },
  },
  finding: {
    help: 'a lens-scoped finding: --label L3-F1 --severity <g> --likelihood <g> --impact <g> [--text "..."|--file f] [--location "..."]',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'finding', { label: a.label, severity: a.severity, likelihood: a.likelihood, impact: a.impact, location: a.location, text: payloadText(a) }); return `finding ${a.label} recorded` },
  },
  observe: {
    help: 'a below-bar observation with a FATE (the merge must dispose it): --kind note|checked-held [--label ...] --text|--file',
    fn: (runDir, seatId, a) => { const e = append(runDir, seatId, 'observe', { kind: a.kind || 'note', label: a.label, text: payloadText(a) }); return `observation recorded (${e.key}) — awaiting merge disposition` },
  },
  cite: {
    help: 'citation-ledger row (the cross-round re-fetch gate): --claim "..." --reference "..." --confidence high|medium|low --access-date YYYY-MM-DD',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'cite', { claim: a.claim, reference: a.reference, confidence: a.confidence, access_date: a['access-date'] }); return `citation recorded: ${a.reference}` },
  },
  friction: {
    help: 'attributed friction (survives aborts as an event): --text|--file',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'friction', { text: payloadText(a) }); return 'friction recorded' },
  },
  petition: {
    help: 'petition the bench (never sanctioned; does not pause your duties): --class ethical|safety|integrity|constitutional --basis "..." --relief "..."',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'petition', { class: a.class, basis: a.basis, relief: a.relief }); return `petition filed (${a.class}) — the bench hears it before the debate continues` },
  },
}

process.exitCode = runVerb('red-lens.mjs', verbs, process.argv.slice(2))

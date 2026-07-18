#!/usr/bin/env node
// tools/blue.mjs — the BLUE seats' record contract (synthesize, respond, lanes,
// frontier). NO board verbs at all: additive-only and never-touch-the-ledger are
// topology here, not obedience. The revision event is the round record W1.7
// attests; manifest rows are the correctness manifest's receipts.
import { append, payloadText, registerSeat, runVerb } from '../lib/record.mjs'

const verbs = {
  register: {
    help: 'FIRST ACTION at the seat: register --run <runDir> --seat-id <SEAT_ID from your prompt>',
    fn: (runDir, seatId) => { const r = registerSeat(runDir, seatId); return `registered ${seatId} (shard nonce ${r.nonce})` },
  },
  revision: {
    help: "the round-record event (the CHANGELOG entry body via --file) — singleton per seat-round; emit AFTER your report edits land",
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'revision', { text: payloadText(a), claim_count: a['claim-count'] ? Number(a['claim-count']) : undefined }); return 'revision recorded — the round is on the record' },
  },
  'manifest-row': {
    help: 'one correctness-manifest receipt per repaired gap: --id R2-3 --row "figures recomputed; acceptance check run: pass; sites swept: S2,S4"',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'manifest-row', { gap_id: a.id, row: a.row || payloadText(a) }); return `manifest row recorded for ${a.id}` },
  },
  dispute: {
    help: "contest a grade through the accounted channel: --id <gap> --dimension severity|likelihood|impact|complexity_cost --proposed <grade> --evidence \"...\"",
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'dispute', { gap_id: a.id, dimension: a.dimension, proposed: a.proposed, evidence: a.evidence }); return `dispute filed on ${a.id}.${a.dimension}` },
  },
  confidence: {
    help: 'per-claim confidence (the calibration substrate): --label <claim-label> --grade high|medium|low',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'confidence', { label: a.label, grade: a.grade }); return `confidence recorded for ${a.label}` },
  },
  position: {
    help: "the round's ### BLUE section (prose via --file)",
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'position', { text: payloadText(a) }); return 'position recorded' },
  },
  closing: {
    help: 'a ### BLUE CLOSING entry per docketed gap: --id <gap> --file <prose>',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'closing', { gap_id: a.id, text: payloadText(a) }); return `closing filed for ${a.id}` },
  },
  friction: {
    help: 'attributed friction (survives aborts as an event): --text|--file',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'friction', { text: payloadText(a) }); return 'friction recorded' },
  },
  petition: {
    help: 'petition the bench (never sanctioned; does not pause your duties): --class ethical|safety|integrity|constitutional --basis "..." --relief "..."',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'petition', { class: a.class, basis: a.basis, relief: a.relief }); return `petition filed (${a.class})` },
  },
}

process.exitCode = runVerb('blue.mjs', verbs, process.argv.slice(2))

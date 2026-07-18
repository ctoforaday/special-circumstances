#!/usr/bin/env node
// tools/bench.mjs — the BENCH's record contract (judge-r, judge-terminal,
// assemble). No mint: the bench rules, never originates. Opinions are required
// to be opinions — disposition, principle, tension, review flag — not
// dispositions. The bench keeps no memory; what it publishes here IS its
// continuity.
import { append, payloadText, registerSeat, runVerb } from '../lib/record.mjs'

const verbs = {
  register: {
    help: 'FIRST ACTION at the sitting: register --run <runDir> --seat-id <SEAT_ID from your prompt>',
    fn: (runDir, seatId) => { const r = registerSeat(runDir, seatId); return `registered ${seatId} (shard nonce ${r.nonce})` },
  },
  opinion: {
    help: 'a ruling as an OPINION: --gap-id R3-2 --disposition carried|closed|... --principle "..." --tension "correctness vs economy" --review-flag "why a human should look" [--file rationale]',
    fn: (runDir, seatId, a) => {
      append(runDir, seatId, 'opinion', { gap_id: a['gap-id'], disposition: a.disposition, principle: a.principle, tension: a.tension, review_flag: a['review-flag'], rationale: payloadText(a) })
      return `opinion recorded: ${a['gap-id']} ${a.disposition}`
    },
  },
  'petition-rule': {
    help: 'rule on a petition: --petitioner <seatId> --class <class> --as granted|denied --file <opinion> (a halt is its own verb)',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'petition-rule', { petitioner: a.petitioner, class: a.class, ruling: a.as, opinion: payloadText(a) }); return `petition ${a.as} (${a.petitioner})` },
  },
  halt: {
    help: 'the safety boundary: --file <written opinion — capture relays it like a FAIL, never smoothed>',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'halt', { opinion: payloadText(a) }); return 'JUDICIAL HALT recorded — capture relays this verbatim' },
  },
  certify: {
    help: 'the run-end certification statement ("what I would want a human to re-examine"): --file <statement>',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'certify', { statement: payloadText(a) }); return 'certification recorded' },
  },
  friction: {
    help: 'attributed friction (the bench seats had no event path before — the abort-surviving copy): --text|--file',
    fn: (runDir, seatId, a) => { append(runDir, seatId, 'friction', { text: payloadText(a) }); return 'friction recorded' },
  },
}

process.exitCode = runVerb('bench.mjs', verbs, process.argv.slice(2))

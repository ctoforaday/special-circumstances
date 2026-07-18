# round 2 — lens 2 (leaf-node citation verification; slice 2 of 4: §2 + §3 + their referenced footnote definitions)

Full living report re-read whole (3 consecutive Read windows, lines 1–1387). Citation
ledger read first; same-day access dates (2026-07-17) mean zero drift window, so claims
verified HIGH in round 1 are carried, not re-fetched. Focus of this pass: (a) round-1
repairs inside §2/§3 re-verified as new claims (repair-regression discipline); (b) every
sub-HIGH or never-fetched statement↔reference pair in the slice attempted at the leaf.

## Repairs in-slice re-verified (all faithful — no repair-regression found)

- **R1-5 (§3.3 item 2):** "TWO leaf-checked OPEN ... #76239 ... #68375 — and ONE
  historical: #32191 ... CLOSED as duplicate (canonical untraced; 2.1.58–2.1.71 era)" —
  matches the ledger-verified statuses and red's required_fix exactly. Propagated to
  [^McpHeadlessBugs]. CLEAN.
- **R1-6 (§3.2):** "the CLI exits nonzero, but the cli-reference page publishes no
  exit-code table ... the wrapper treats ANY nonzero exit as failure" — re-confirmed on a
  fresh cli-reference fetch this round: still no exit-code table; `--max-turns` still says
  only "Exits with an error." CLEAN.
- **R1-20 (§3.4 verdict):** split HIGH(non-bare)/OPEN(bare) present, with the
  "verdict stamp must not grade higher than its recommended configuration's evidence"
  clause. CLEAN.
- **R1-15/R1-22/R1-27/R1-29 (§3.2 recipe, §2.3 status field, §3.4 gate-survival table,
  §3.4 resume cap + dead-man):** design changes present as specified; no citation load.
  Gate-survival table cross-checked against ledger-verified [^RoutinesDocs]/
  [^ScheduledTasks] facts — no cell contradicts a verified source.

## Leaf verifications this round (new fetches / pin reads)

| # | Claim (section, quote) | Source followed | Result | Confidence |
|---|---|---|---|---|
| 1 | §3.2 "`--json-schema` yields schema-conforming `structured_output` (invalid schema exits with error ≥2.1.205)" | code.claude.com/docs/en/cli-reference (live fetch) | Verbatim: "Claude Code exits with an error on an invalid schema ... Before v2.1.205, an invalid schema produced unstructured output with no error"; page carries min-version 2.1.205 marker | **HIGH** |
| 2 | §3.3 "#66395 ... v2.1.161–v2.1.168; fixed v2.1.169" (round-0/1 status: title-asserted MEDIUM, body not fetched) | gh issue view 66395 --json body | Body quotes the v2.1.169 changelog verbatim: "Fixed `claude -p` being slow or appearing to hang on Windows while waiting for the slash-command/skill scan (regression in 2.1.161)"; body states the affected span "between v2.1.161 (June 2, 2026) and v2.1.168 (June 6, 2026)" | **HIGH — banked upgrade** (was MEDIUM; [^WindowsHang]'s "body not fetched" caveat can be retired) |
| 3 | §3.4 rung-3 caveat "~973MB models" | `git show 7bc501e:ideas/backlog.md` item 34 | Verbatim at the pin: "Total model footprint if daemon-only: 973MB (embed+rerank), not 2.2GB" | **HIGH** |
| 4 | §3.4 "isolated Anthropic-managed VM per session, credentials held outside the sandbox" ([^WebSandbox], round 0 "surveyed") | code.claude.com/docs/en/sandbox-environments (live fetch) | Verbatim: "runs each session in an isolated, Anthropic-managed virtual machine. A network proxy enforces a default allowlist, and a separate proxy holds your GitHub token outside the sandbox while issuing scoped credentials for repository access inside it" | **HIGH — banked upgrade** (precision note: the doc names the GitHub token specifically; "credentials" is a fair gloss) |
| 5 | §3.4 "anacron interval-since-last-run semantics" ([^MissedRun], round-1 MEDIUM, not fetched) | man7.org anacron(8) (live fetch) | Verbatim: "checks whether this job has been executed in the last n days"; "does not assume that the machine is running continuously" | **HIGH — banked upgrade** |
| 6 | §3.4 GHA "runs commonly delayed 5–30 min ... can be dropped" ([^GhaSchedule] community half, MEDIUM) | github.com/orgs/community/discussions/52477 (live fetch — MUST-TRY attempt) | Docs language ("delayed ... may be dropped") confirmed in-thread; a participant measured delays "close to an hour" around 00:00. The 5–30 min "typical" range stays community-lore-grade; worst-case observed is WORSE than the report's range, which strengthens (not weakens) the missed-run-tolerance design conclusion | **MEDIUM on the 5–30 figure (honestly labeled community-measured in the report — no gap); HIGH on delay/drop phenomenon** |
| 7 | §2.4 "DGM only admits a change to its archive after empirical validation against a benchmark, never on the proposer's say-so" ([^DGM]) | arXiv:2505.22954 abs (leaf-held from r1: abstract "empirically validates each change using coding benchmarks") + r1 ledger line "archive admission = compile+edit-ability, all agents benchmark-evaluated" (abs+html read) | Ordering true (evaluation precedes archive entry); gating implication overstated — see L2-F1 | **MEDIUM** |

MUST-TRY attempt lines for everything graded below HIGH this pass: row 6 — discussion
thread fetched live this round (attempt made; figure stays community-grade by nature of
the source, and the report already labels it so). Row 7 — no fetch failure to excuse: the
sources were read at the leaf (r1 abs+html, recorded in the citation ledger); the MEDIUM
reflects claim-vs-source mismatch, not inaccessibility. Probe P1/P2 residue (§3.1,
MEDIUM, ephemeral instrument): re-running the probes from an audit seat means spawning
paid claude sessions — not triable at a lens seat; the round-1 merge disposition of
record ("accept the re-run-and-commit offer at build; no gap minted") stands unchanged
and blue has not claimed the fix landed, so nothing regressed.

## Findings

### L2-F1 — "The DGM analogy is exact" overstates: DGM's archive admission is NOT validation-gated
- location: §2.4 "/graduate (the promotion pipeline)" — "The DGM analogy is exact and is
  the design argument: DGM only admits a change to its archive after empirical validation
  against a benchmark, never on the proposer's say-so.[^DGM]"
- problem: the abstract supports "empirically validates each change using coding
  benchmarks," and evaluation does precede archive entry — but per the round-1
  leaf read already in the citation ledger (arXiv abs+html), DGM's archive admission
  criterion is compile + retained ability to edit code; benchmark performance does NOT
  gate admission, and low scorers are deliberately retained (the open-ended-exploration
  point of the paper). Blue's /graduate gate is pass-required. So the analogy is
  STRONGER-than-DGM on exactly the dimension the sentence leans on ("only admits after
  ... validation") — directionally fine for the design, but "exact" is false and the
  sentence invites a builder to read DGM as a precedent for threshold-gated admission,
  which it is not. Within-source-condition class (right paper, right quote-family, wrong
  mechanism attribution).
- required_fix: drop "exact" (e.g. "the DGM analogy is direct and is the design
  argument"); one clause of honesty: "DGM evaluates every change empirically before
  archiving but admits even low scorers for exploration — our promotion gate is
  stricter: pass-required." No re-grading of H2 needed; the design argument survives
  intact (arguably strengthened).
- corroboration: statement↔[^DGM] pair MEDIUM as written; HIGH once restated.
- grading: certain (textual) × low × trivial → severity **LOW**

## Notes for the merge (not gaps)

- **Banked upgrades:** [^WindowsHang] MEDIUM→HIGH (row 2); [^WebSandbox] VM/credentials
  claims surveyed→HIGH (row 4); [^MissedRun] anacron MEDIUM→HIGH (row 5); the
  `--json-schema` ≥2.1.205 detail pinned HIGH (row 1); ~973MB pinned at the backlog pin
  (row 3).
- **OQ7 partially answerable from row 4's fetch (out of my slice — §8/§4 owner should
  carry):** the sandbox-environments page states the built-in sandboxed Bash tool "does
  not support native Windows. On Windows hosts, use WSL2 or one of the container or VM
  approaches"; it also documents `@anthropic-ai/sandbox-runtime` (whole-process
  Seatbelt/bubblewrap wrap, beta) — i.e., layer 4's preferred close is confirmed
  unavailable natively on this box, and the report's §4.3(b) posture is the documented
  one. Also on that page: "Auto mode replaces the prompt with a classifier" — relevant
  raw material for OQ17's `disableAutoMode` leaf-verify.
- **All other §2/§3 statement↔reference pairs** trace to ledger entries verified HIGH in
  round 1 at the same access date; carried without re-fetch per the ledger rule. No
  incomplete-repair or footnote-lag instances found in the slice (propagation greps in
  blue's CHANGELOG spot-checked against the body text during the full read: `#32191`,
  `print-only`, `Bash(node`, `leaf-checked OPEN` — all consistent in §2/§3).

## Friction

None impeding: gh, WebFetch, git-show-at-pin, and native Read all sufficed; no PDF-only
source in this slice required the extraction MCPs.

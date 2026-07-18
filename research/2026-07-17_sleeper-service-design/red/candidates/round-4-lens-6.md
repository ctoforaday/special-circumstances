# Round 4 — Lens 6 (dark-side & risk) candidate findings

Audit surface: `blue/report.md` read WHOLE in 4 consecutive windows (lines 1–1893,
2026-07-17, post-round-3 revision; CHANGELOG used as navigation hint only). Leaf checks
this pass: permissions doc re-fetched live (Bash wildcard semantics + carve-out
override — see Checked-and-HELD); special-circumstances `.claude/` dir listed;
R3-12 arithmetic recomputed; §4.2 carve-out deny enumeration diffed against the doc's
14-command set. Lens-scoped ids only; stable R4-N ids and lineage are the merge's.

## L6-F1 — out-of-contract interactive `/self-improve`: markerless + windowless run dirs COUNT AS the non-sleeper corroboration pool, and the containment sentence argues the wrong polarity; `self-improve.md` lacks the `disable-model-invocation` gate its sibling carries (challenges the R3-2 closure's residual argument)

- location: §3.4 (rung-0 execution shape, R3-2 paragraph) — "such a session carries the
  interactive profile's own permission gates rather than silently impersonating a sleeper
  run (no marker ⇒ the fence never engages, but also: no marker ⇒ its dirs are swept by
  the window log if any wrapper run brackets them, and its friction is ordinary-classed,
  requiring corroboration)"; §1.5 — "**ordinary capability/protocol classes** require at
  least one non-sleeper occurrence before entering the top-3 comparison"; §0 tree —
  graduate.md "frontmatter: disable-model-invocation: true" (self-improve.md carries no
  such line); §4.3 layer 6 — the invocation gate covers /graduate ONLY.
- problem, three legs:
  (a) **POLARITY.** The containment sentence says the out-of-contract run's friction is
  "ordinary-classed, requiring corroboration" — but untagged rows ARE the corroboration
  pool: R3-3's own fix defines the pool negatively ("excludes them from the NON-sleeper
  corroboration pool" — everything not tagged sleeper is corroboration-eligible). An
  interactive `/self-improve` run (afternoon session, no wrapper, no bracketing window —
  the sentence's own "if any wrapper run brackets them" concedes the conditionality)
  emits loop-authored friction that is markerless AND out-of-window → counted
  NON-sleeper → it simultaneously evades the 1-per-class cap AND can SUPPLY the one
  non-sleeper occurrence that ungates a sleeper-only class. The threat is not that this
  friction needs corroboration; it is that it PROVIDES corroboration. This is the
  origin-tag escape on its FOURTH surface (dir-name glob → spawned sub-run → rung-0 →
  out-of-contract interactive), and the only guard on this surface is a command-file
  header sentence — an instructional gate, on a report whose §4.1 headline evidence is
  that instructional gates degrade (0/175; "a wish, not a gate").
  (b) **INVOCATION.** self-improve.md is an installed command visible in every
  interactive session on the box and carries NO `disable-model-invocation: true` (§0
  gives that flag only to graduate.md). A prompt-injected or model-initiated
  "/self-improve" inside an interactive session executes the loop's phases un-wrapped
  and un-fenced under the interactive profile (which allows repo writes) — §4.3 layer 6
  answers this route-around for /graduate and is silent for /self-improve.
  (c) §6 row 10's round-3 re-affirmation ("the rung-0 void is closed (R3-2 — manual runs
  pass through the wrapper...)") is true for `--manual` but silent on this adjacent
  surface; the row's grade does not carry the residual.
- required_fix: (i) add `disable-model-invocation: true` to self-improve.md — one
  frontmatter line, symmetric with graduate.md; the markdown is the wrapper's phase-1
  payload in every mode (§3.4's own doctrine), so model/scheduled invocation has no
  legitimate use; (ii) close the pool-membership default: harvest QUARANTINES markerless
  run dirs bearing loop-output shape (staged docket / stub / self-improve step
  artifacts) — neither sleeper-capped nor corroboration-eligible (fail-closed toward
  "counts for nothing") — OR redefine the corroboration pool positively (only runs with
  affirmative human-present provenance corroborate) instead of everything-untagged;
  (iii) restate the §3.4 containment sentence with corrected polarity and carry the
  residual in §6 row 10.
- pattern: origin-tag surface incompleteness (4th recurrence) + corroboration-pool
  membership by default-negative + instructional gate where the sibling command got
  mechanism.
- grading: low-medium × medium × low → severity **medium**

## L6-F2 — R3-3's justifying premise contradicts the printed profile: the nightly red-merge seat CANNOT write the shared agent-memory dir (fence denies writes outside research/+ideas/; §4.2 denies `Edit(<REPO>/.claude/**)`), so the repair guards a foreclosed write — while the REAL consequence (the seat's mandatory record-new-patterns write silently fails-denied every night) is stated nowhere (challenges the R3-3 closure)

- location: §1.5 (R3-3 paragraph) — "the nightly bounded-FEOV pass spawns a red-merge
  seat that writes the SHARED agent-memory dir under hook coverage, and a pattern minted
  there has no run dir and no marker"; §0 tree — sleeper-guard "deny writes outside
  research/ + ideas/ for sleeper runs"; §4.2 deny — "Edit(<REPO>/.claude/**)"; §0
  invariant 6 — "the loop's write surface and the suite's behavior surface are disjoint".
- problem: red's project gap-pattern memory is `.claude/agent-memory/**` — in-repo,
  project scope (the design's own [^IdeasCorpus] evidence: "red-auditor wrote its
  memory: project gap-pattern file"). Under the sleeper profile that path is foreclosed
  at THREE layers for any in-session actor: the fence (outside research/+ideas/), the
  `Edit(<REPO>/.claude/**)` deny, and — for the marker-loss case — the phase-0 canary
  aborts the run before any work (no fired-record ⇒ abort, R2-2). So the write R3-3 was
  built to tag CANNOT EXECUTE in-window, and §1.5's sentence asserting it happens and
  §4.2's profile cannot both be true as printed. Whichever horn blue intends, something
  unstated surfaces:
  (a) if the denies are the intent (safe polarity — probably right), then the nightly
  red seat's own protocol clause ("AFTER catching a new gap pattern YOU MUST record it
  in your project memory") fails-denied EVERY night — a standing silent degradation of
  the nightly adversary's learning loop, plus a recurring stream of in-run fence denials
  landing in the fired-record, neither stated nor priced (are routine memory-write
  denials distinguishable from attack evidence in that record?); R3-3's live remit
  shrinks to OUT-of-window authorship — which is L6-F1's surface, not the in-window one
  §1.5 describes — and the snapshot-diff machinery is dead where aimed;
  (b) if seat memory-writes are meant to succeed, an undeclared allow widens the write
  surface beyond research/+ideas/ and invariant 6's disjointness sentence is false as
  printed.
- required_fix: pick a horn and state it. Cheapest coherent shape is (a): one paragraph
  declaring that under the sleeper profile agent-memory writes are DENIED by design
  (nightly seats do not learn; nightly-run patterns reach the corpus only via friction
  text, ordinary-classed and corroboration-requiring); reword §1.5's R3-3 paragraph to
  its true remit (belt for out-of-window/interactive authorship + defense-in-depth
  against profile drift); and state that expected nightly memory-write denials are a
  NORMAL fired-record class, so they are not misread as tamper evidence. If (b),
  enumerate the allow and re-argue invariant 6.
- pattern: repair-premise foreclosed by own profile (doctrine-vs-implementation ×
  self-defeating-mitigation) — new flavor, recorded in red gap-pattern memory this round.
- grading: certain × low-medium × low → severity **medium**

## L6-F3 — unobserved wrapper exits (power loss / OS kill / reboot): "EVERY exit path it observes" is honestly scoped but the UNCLOSED-window semantics are unstated; the snapshot chain survives by construction, the window log's open interval only "usually"

- location: §2.2 step 7 / §4.3 layer 5 — "the wrapper snapshots at EVERY exit path it
  observes — it outlives the session by construction"; §1.5 (R3-4) — "the wrapper logs
  the run-window START ... and the END at whatever exit it observes — step-7 record,
  abort, or DEAD-mark time"; §6 row 6 — sleep/laptop, likelihood High (the design's own
  environment).
- problem: a hard kill (power loss, forced reboot, OOM-kill of the wrapper) is an exit
  the wrapper does NOT observe. The snapshot-chain leg HOLDS anyway (last-recorded
  snapshot-of-any-kind = that run's START snapshot; missing link fails closed — this
  composes, but by construction of R3-7, not by stated argument). The window log does
  not obviously hold: START with no END, and harvest's reading of an UNCLOSED window is
  unspecified. If harvest requires closed [start,end] pairs, a model-authored in-session
  run dir from the killed night (`Edit(<REPO>/research/**)` is allowed) is unswept and
  markerless until some later observed exit closes the interval; the resume path usually
  closes it before the next fresh harvest, but "usually" is doing unstated work on the
  design's own high-likelihood failure environment.
- required_fix: one clause each: (i) "an unclosed run window is read as extending to the
  present" (fail-safe polarity — over-tagging is the cheap error); (ii) own the
  unobserved-exit case in the step-7/layer-5 text: its backstop is the snapshot chain +
  the resume path, not an END snapshot.
- pattern: happy-path-only-gate residue on the R3-4/R3-7 repairs themselves.
- grading: low × low-medium × trivial → severity **low**

## L6-F4 — the dead-man push channel's disarm switch (scheduling-enabled flag) has unstated custody, and the same-round R3-9 (always-print liveness) and R3-10 (condition-on-flag) repairs to the SAME line have unstated composition: does a disabled channel still print?

- location: §3.4 — "The warning is CONDITIONED on scheduling-enabled — a flag the
  operator's schedule-setup step sets and the disable step resets"; same paragraph —
  "it always prints its last-check timestamp as a liveness signal (R3-9: a silent hook
  and a passing hook must not look identical)".
- problem: two same-round repairs to one surface whose composition is unstated. If
  "conditioned" means fully silent when the flag is off/absent, then flag LOSS
  (misconfig, partial setup, uninstall/reinstall — the marker-loss class this design
  itself treats as the modal drift case, R1-28) silently disarms the dead-man channel,
  and a disarmed hook and a healthy-but-disabled hook look identical — re-opening for
  this surface exactly what the R3-9 always-print clause closed. Flag custody is also
  unstated; it must be the operator-owned dir (anywhere loop-reachable would let the
  loop silence its own dead-man — foreclosed by the deny set today, but custody should
  be one named path, not an inference).
- required_fix: two clauses: the flag lives in the operator-owned dir beside the ledger;
  when scheduling-disabled the SessionStart line still prints one token ("sleeper:
  scheduling disabled; last check <ts>") so silent-disarm ≠ disabled-by-choice.
- pattern: sibling-repair composition + fail-open flag-keyed guard.
- grading: low × low-medium × trivial → severity **low**

## Checked and HELD (negative results — merge need not re-derive)

1. **R3-15 belt-deny pattern shape is doc-legal (hypothesis refuted at the leaf).**
   Tested the sharpest available attack on the three-part close: "leading/mid-position
   wildcards in Bash rules don't match (prefix-only), so `Bash(* --output=*)` is inert."
   REFUTED — permissions doc, live fetch 2026-07-17: "Bash rules support glob patterns
   with `*`. Wildcards can appear at any position in the command", worked examples
   include `Bash(* --version)` and `Bash(* --help *)`; "A single `*` matches any
   sequence of characters including spaces." The belt denies are matchable shapes.
   Corroboration HIGH. (The `-O` short form remains hook-matcher-only — already owned
   in-report as the enforcement of record.)
2. **§4.2 carve-out deny enumeration is COMPLETE against the doc's set.** Doc names 14
   commands (ls, cat, echo, pwd, head, tail, grep, find, wc, which, diff, stat, du, cd)
   plus read-only git; all 14 are deny-enumerated in the §4.2 sample; read-only git
   intentionally retained as declared-intent allows. Diffed line-by-line. HIGH.
3. **R3-12 arithmetic recomputed and holds.** 4 nights/dir (initial fire + 3 resumes),
   × M=3 dirs → HALT at night 12; $5 × 10 ceiling nights = $50 cap trips ~night 10 < 12;
   "whichever comes first" is correct as printed.
4. **Deny-beats-carve-out is doc-stated, re-confirmed live.** "The set is not
   configurable; to require a prompt for one of these commands, add an `ask` or `deny`
   rule for it"; also "a deny rule can't carry allowlist exceptions" (deny supremacy).
   HIGH.

## Notes (not gaps — surfaced, graded below the raise bar)

- **R3-11 over-normalization residual:** stripping ALL digit runs can merge two distinct
  causes into one signature → premature HALT. Fail-closed polarity (human clears);
  adding complexity to prevent it would be worse than the failure. Note only.
- **Date-keyed idempotency interplay:** a morning `--manual` run marks today
  recorded-complete, so that night's scheduled fire exits 0 — manual use silently
  suppresses the day's scheduled run. Appears intended (one item per day); worth one
  line in scheduling.md so the operator is not surprised. Note only.

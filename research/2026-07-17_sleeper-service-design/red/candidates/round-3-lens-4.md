# Red round-3 — lens 4 (leaf-node citation verification)

Slice: §6 (risk matrix) / §7 (pre-flight self-audit) / §8 (open questions) + owned footnote
[^HooksJson]; footnote-block ownership for the issue-status claims §6/§7 assert.
Full living report re-read in context (three consecutive windows: 1–608, 609–1128,
1129–1641). Method: re-follow every leaf whose source is volatile (issue trackers) or whose
last verification is ≥2 rounds stale; carry pin-immutable and non-volatile academic leaves;
run the MUST-TRY observable on every down-grade.

## Findings

### L4-F1 — "pinned read-only git argv set" is REFUTED at the leaf: `Bash(git log *)` + `--output` is an arbitrary-file write gadget (MEDIUM)

**Location:** §6 risk matrix, row 4 disposition — quoted: *"the subprocess surface is the
pinned read-only git argv set plus the Workflow tool executing PINNED code from the
read-only plugin copy under schema-bound args ... a breach therefore needs a
permission-engine bug AND a write gadget in either pinned-argv git or pinned debate.js code
simultaneously."* Corroborating text challenged: §4.3 layer 4 (i) — *"the two Bash node
scripts are re-hosted WRAPPER-SIDE ... so the in-session Bash allowlist stays pinned-argv
read-only git commands only (no rule grants the model argv that chooses a subprocess write
target ...)"*; artifact: §4.2 allow list line `"Bash(git status)", "Bash(git diff)",
"Bash(git log *)"`.

**Leaf verification (tool run, this box, 2026-07-17):**
`git -C <repo> log -1 --oneline --output=/tmp/gitlog-gadget-test.txt` → **exit 0**, and the
file `/tmp/gitlog-gadget-test.txt` was created containing the commit line
(`6df52af docs: sharpen FEOV debate engine ...`). `git log` writes its output to an
arbitrary path via its own `--output=<file>` flag; the wildcard rule `Bash(git log *)`
grants that argv. (A second confirmatory probe was classifier-blocked, but the first run
already demonstrates the write.)

**Why this is a gap, not a nit:**
- The claim "read-only git commands only" / "no rule grants argv that chooses a subprocess
  write target" is **false as printed** — the write target is chosen by the model via
  `--output=<anywhere>`, and §4.3 layer 4 concedes in the same breath that "File rules don't
  apply to arbitrary subprocesses," which is exactly what a `git log --output` write is.
- Leg (a)'s conjunction **collapses**: it argues a breach "needs a permission-engine bug AND
  a write gadget in pinned-argv git." The write gadget is present **without any bug** —
  allowed by the rule as written. The risk-accept's channel-narrowness argument for the
  High-impact boundary breach (row 4) rests on a premise the CLI refutes.
- **Not covered by OQ18.** Open question 18 defers "Compound-command, redirection, and
  traversal matching semantics for the surviving `Bash(git ...)` allow rules." `--output` is
  none of the three: not a compound command (`;`/`&&`), not shell redirection (`>`), not
  path traversal. It is a git-native output flag. So this write channel is neither
  enumerated as read-only-safe nor listed among the deferred unknowns — it is asserted
  settled ("read-only") while an untested escape exists.
- **Backstops are incomplete for it.** Layer 5's start/end snapshot compare (R2-7) hashes
  only the *guardrail file set* + `git status --porcelain`. A `git log --output=<path OUTSIDE
  the repo>` write shows in neither the guardrail-hash set nor porcelain — so the tamper-
  evidence watchman does not see it. A write to a path *inside* the repo but outside
  research/+ideas/ would show in porcelain (caught), but an out-of-repo target (e.g. exfil
  staging under `~`, or overwriting an operator-owned file) escapes both.
- The sleeper-guard hook (layer 2) claims to block "Bash-write targets outside
  research/+ideas/," but nothing in the report asserts the hook parses `git log --output=X`
  as a write — that parsing IS the OQ18 hard problem, and `--output` is a nastier case than
  redirection because it looks like a benign read command.

**Grading:** L = the rule ships as printed (`Bash(git log *)` is literally in the §4.2
sample), so the gadget is present by construction; exploitation needs the model to weaponize
it, which the row-4 risk-accept **explicitly refuses to assume away** ("No premise about
actor intent is used"). I = High (out-of-boundary subprocess write / exfil staging = the row-4
breach class). Cx = Low (pin the argv — `Bash(git log --oneline)` / exact forms — or add
`--output`/`-O`/`--output-directory`-class flags to the hook's Bash-write matcher and to
OQ18's named scope). Corroboration confidence for the challenged claim: **LOW (refuted at
leaf).** Disposition proposed: fix (tighten the rule) OR downgrade §6 row 4 leg (a) +
§4.3 layer 4 (i) from "read-only" to "read-only EXCEPT the `git log --output` write channel,
tracked in OQ18" and re-argue narrowness accordingly. This is a build-altering catch on a
load-bearing risk-accept; surface, do not soft-pass.

## Verified clean (no gap)

**§6 row 8 / §7 Pattern A — GitHub issue statuses, live re-fetched 2026-07-17 (volatile,
≥2 rounds stale on the r1 leaves; drift-checked):**
- #76239 **OPEN** — title confirms "regression for single-turn sessions since CLI 2.1.144."
  Matches [^McpHeadlessBugs]. HIGH.
- #68375 **OPEN** — title "hangs indefinitely under `claude -p` when the full MCP fleet is
  loaded — regression in 2.1.177, works under `--strict-mcp-config`." Matches footnote
  (works around via `--strict-mcp-config`). HIGH.
- #22055 **CLOSED / NOT_PLANNED** — matches §4.1/[^PermAskBypass]. HIGH.
- #66395 **CLOSED / NOT_PLANNED** ([DOCS], span v2.1.161–v2.1.168 fixed v2.1.169 in title) —
  matches [^WindowsHang]. HIGH.
- #32191 **CLOSED / DUPLICATE** — matches the R1-5 correction (§3.3, §6 row 8, §7). HIGH.
- #23707 **CLOSED / NOT_PLANNED** — matches [^WebSandbox]/§3.4/§7. HIGH.
- #837 CLOSED COMPLETED / #14246 CLOSED DUPLICATE — **carried HIGH from r1** (ledger r1).
  MUST-TRY line: live `gh issue view 837/14246` was classifier-blocked on 3 attempts this
  pass; closed issues are low-volatility (closed states rarely revert), the supersession
  story is unaffected, and r1 leaf-confirmed both — no down-grade, impossibility recorded.

**§6 row 9 — [^HooksJson] bootstrap guard.** Pin-immutable (`git show 7bc501e:...`), verified
r1 HIGH; a git-pinned blob cannot drift — not re-fetched. HIGH (carried).

**§7 R2-14 — Pattern B/E stale-grade fix (closes the round-2 lens-4 L4-F1).** §7 bullet now
reads "pricing figures graded MEDIUM ... (upgraded to leaf-verified HIGH round 1, R1-11 —
this bullet's lag fixed round 2, R2-14 ...)" and R1-11 is added to §7's banked-upgrade list.
The round-2 stale self-report is repaired. Verified present, no residual. CLOSED.

**§7 / §8 OQ8 — STOP circumvention figures** (0.42% CI 0.31–0.57% / 0.46% CI 0.35–0.61%
insignificantly HIGHER / 10,000 sampled / syntactic detection). arXiv-pinned academic leaf
(ar5iv §6.2/Table 2), verified r1 AND r2 HIGH by three lenses; non-volatile — not re-fetched.
HIGH (carried).

**§7 Round-2 upgrade claims** ([^WindowsHang] span; [^WebSandbox] VM/token; [^MissedRun]
anacron; [^Pricing] zero-drift + Batch ≤24h; `--json-schema` invalid-schema error ≥2.1.205;
rung-3 ~973MB) — leaf-verified r2 HIGH by lenses 2/3 (ledger r2); byte/pin-stable or
same-access-date, 1 round elapsed; not re-fetched. HIGH (carried).

**§7/§8 [^ResearchCommand] mixed-locus + `--input-format stream-json`** — re-confirmed
incidentally this pass: `claude --help` on 2.1.212 shows `--input-format <format> ... "text"
(default), or "stream-json" (realtime streaming input) (only works with --print)`, and
doubts.md @7bc501e confirms "sc-quality-gate fired on workflow-agent writes; red-auditor
wrote its memory: project gap-pattern file." HIGH.

## Notes
- No new fabricated or mis-homed citation in the slice. The single finding is a design claim
  refuted by the CLI's own behavior, surfaced by running the leaf rather than reading it.
- OQ18 should be widened to name git-native output flags, independent of whether L4-F1 is
  fixed by rule-tightening.

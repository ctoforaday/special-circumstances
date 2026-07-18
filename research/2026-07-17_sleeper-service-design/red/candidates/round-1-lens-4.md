# red candidate pass — round 1, lens 4 (leaf-node citation verification, slice 4/4)

Slice: §6 (risk matrix), §7 (pre-flight self-audit), §8 (open questions) + footnote
definitions these sections reference. Full report re-read whole (2 consecutive windows,
1111 lines). Ledger was empty at start; all verifications fresh.

## Findings

### L4-F1 — §6 row 8 asserts #32191 OPEN; it is closed as duplicate — and §3.3 claims it was "leaf-checked OPEN"

- location: "## 6. Risk matrix", row 8 — "MCP headless flake (open bugs
  #76239/#68375/#32191) starves qmd recall"; cross-anchor "### 3.3 Disconfirming pass"
  — "three relevant upstream defects, status leaf-checked OPEN 2026-07-17: ... HTTP MCP
  `-p` runs exiting silently (#32191, older)".
- evidence (live fetch 2026-07-17, WebFetch of
  https://github.com/anthropics/claude-code/issues/32191): title "`claude -p` with HTTP
  MCP server silently exits with no output"; **state: Closed as duplicate**; affected
  versions 2.1.58–2.1.71. Corroboration of the "open" claim: **LOW — refuted**.
- the report's own footnote [^McpHeadlessBugs] hedges exactly this ("status per search
  listing; not individually re-fetched — grade accordingly") and §7 repeats the flag —
  so the body (§3.3 "leaf-checked OPEN", §6 "open bugs") asserts what the footnote
  disclaims. Body-overclaims-footnote class (footnote-overattribution / repair-lag
  family).
- grading: likelihood the error changes the design = LOW (closed-as-duplicate is not
  fixed; the workaround posture in row 8 stands; canonical issue untraced). Impact =
  MEDIUM on process integrity: "status leaf-checked OPEN 2026-07-17" is a false
  verification claim for one of the three ids, and the phenomenon-class evidence now
  rests on a 2.1.71-era report with no traced canonical successor. Complexity = trivial.
- required fix: §6 row 8 and §3.3 restate #32191 as "closed as duplicate (canonical
  issue untraced; phenomenon evidence 2.1.58–2.1.71 era)"; delete "leaf-checked OPEN"
  as applied to #32191. Optionally trace the duplicate target or drop the id from the
  "open bugs" list.

### L4-F2 — [^PermAskBypass] claims a "chmod 444 + PreToolUse" workaround in the #22055 thread; chmod 444 is not in the thread (adjacent-slice: footnote is referenced from §4 — route to instance 3 / merge)

- location: "## Footnotes", [^PermAskBypass] — "Community defense-in-depth workaround in
  thread: chmod 444 + PreToolUse hook"; consumed at "### 4.3", layer 3 route-around —
  "chmod-readonly on guard files as defense-in-depth (a documented community pattern for
  the #22055 gap)".
- evidence (2026-07-17): WebFetch of the issue page found no workaround mention (lossy on
  collapsed comments); `gh issue view 22055 -R anthropics/claude-code --comments` (full
  thread) grep: PreToolUse workaround PRESENT ("Until the permissions system properly
  handles Edit/Write `ask` rules, a `PreToolUse` hook provides a reliable workaround —
  hooks run at the process level and can't be bypassed by the model"); "chmod 444"
  ABSENT — "chmod" appears once, inside a commenter's allow-list settings snippet
  (`"Bash(chmod:*)"`), and the closest sibling is a recommendation that settings files
  "should be immutable from the agent's tool scope" (a recommendation, not a chmod-444
  pattern). Attempt trail: WebFetch page fetch → gh full-comments fetch → targeted greps
  (chmod / 444 / read-only / immutable / attrib) — exhausted.
- grading: corroboration HIGH for the PreToolUse half, LOW/refuted for the chmod-444
  half. Likelihood of harm LOW (chmod-readonly is a reasonable idea regardless); impact
  LOW-MEDIUM (a §4.3 defense layer is dressed as "documented community pattern" when it
  is uncited); complexity trivial.
- required fix: split the footnote — keep "PreToolUse hook workaround in thread"
  (quote above), recast chmod-readonly as design-proposed (or cite a real source);
  amend §4.3 layer 3's "documented community pattern" phrasing accordingly.

### L4-F3 — informational, instance-1 territory: gap-pattern mirror line count

- location: "### 1.3" table + [^RedPatterns] — "1,558 lines".
- evidence: `wc -l inputs/red-gap-patterns.md` = 1557 (newline count; 1,558 is exact iff
  the final line is unterminated). Not graded a defect; noting for the merge so
  instance 1 doesn't re-derive.

## Verified clean (my slice)

- §6 row 9 / [^HooksJson]: `git show 7bc501e:plugins/prosthetic-conscience/hooks/hooks.json`
  — quoted bootstrap-guard `_comment` matches verbatim; all 5 hook commands carry the
  guard. HIGH.
- §6 row 8 / #76239: live fetch — OPEN; title/behavior match the report (stdio tools
  silently missing when startup >~2s; regression since 2.1.144; single-turn sessions
  permanently lose tools). HIGH.
- §6 row 8 / #68375: live fetch — OPEN; stdio tool-call hang with multi-server fleet;
  `--strict-mcp-config` workaround; regression at 2.1.177. HIGH.
- §7 "permission issues #22055 ... fetched directly with statuses quoted": #22055
  re-fetched — "Closed as not planned", title matches [^PermAskBypass] verbatim. HIGH
  (status claim; the workaround sub-claim is L4-F2).
- §7 ephemeral-instrument disclosure: `blue/candidates/lane-3.md` carries probe P1/P2
  commands + output fields (`total_cost_usd:0.0246903`, `permission_denials:[]`,
  `terminal_reason:"completed"`) — the "commands stated for re-derivation" claim holds.
  HIGH (that the disclosure is accurate; the residue itself stays a blue-acknowledged
  limitation).
- §8 Q1/Q10 consistency with §3.1 probe scope (non-bare on 2.1.212): consistent. HIGH
  (internal).
- §6 rows 1 and 3 restate hedged evidence ([^SlashHeadlessIssues], [^Dependabot])
  without upgrading its status — no overclaim found in the rows themselves.

## Coverage boundary (for merge)

- §7's "#25621 / #6631 ... statuses quoted" NOT re-fetched here — [^DenyRWIssue] /
  [^DenyBashIssue] are slice-3 footnotes; my prior-run memory is consistent with the
  report's statuses but that is not leaf verification. If instance 3 did not fetch them,
  the merge should assign the fetch.
- STOP percentages (§8 Q8): slice-3; note for instance 3 — arXiv `/html/2310.02304v2`
  has carried section-level figures where /abs/ + PDF failed (prior-run method note).

## Friction

- WebFetch on GitHub issue pages is lossy on collapsed/paginated comments: it returned a
  confident "no workaround mentioned" for #22055 while the full thread (via
  `gh issue view --comments`) contains one. An absence-claim from a summarizer fetch of
  an issue thread is not evidence of absence; `gh` is the reliable path and should be
  the documented default for issue-thread content checks (status checks are fine either
  way).

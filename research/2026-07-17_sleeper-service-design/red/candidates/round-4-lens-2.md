# round 4 — lens 2 (leaf-node citation verification; slice 2: §2/§3 + referenced footnote definitions)

Full report re-read in 4 consecutive windows (lines 1–1893). Round state verified by direct
read: CHANGELOG ends at Round 3; report header carries the Round-3 revision paragraph —
round 4 audits the round-3 revision. New-claim surface in slice: the R3-1/R3-2/R3-4/R3-7/
R3-10/R3-11/R3-12/R3-13/R3-14 edits to §2.2/§2.3/§3.2/§3.4 plus the [^PermissionsDoc]
round-3 footnote additions.

## Leaf verifications this round

1. **[^PermissionsDoc] round-3 additions — live WebFetch, code.claude.com/docs/en/permissions,
   2026-07-17.** All verbatim, zero drift:
   - Carve-out: "Claude Code recognizes a built-in set of Bash commands as read-only and
     runs them without a permission prompt in every mode. These include `ls`, `cat`,
     `echo`, `pwd`, `head`, `tail`, `grep`, `find`, `wc`, `which`, `diff`, `stat`, `du`,
     `cd`, and read-only forms of `git`. The set is not configurable; to require a prompt
     for one of these commands, add an `ask` or `deny` rule for it." — matches §3.2 R3-14
     sentence and the footnote quote exactly. HIGH.
   - Unquoted globs: doc permits them only "for commands whose every flag is read-only";
     "Commands with write-capable or exec-capable flags, such as `find`, `sort`, `sed`,
     and `git`, still prompt when an unquoted glob is present." Footnote's find/git subset
     accurate. HIGH.
   - Exec wrappers `watch`/`setsid`/`ionice`/`flock` + `find -exec`/`-delete` always
     prompt — present verbatim. HIGH.
   - "deny, then ask, then allow" + "a deny rule can't carry allowlist exceptions" + "If a
     tool is denied at any level, no other level can allow it" — all present verbatim. HIGH.
   - dontAsk: "Auto-denies tools unless pre-approved via `/permissions` or
     `permissions.allow` rules." — §3.2's corrected sentence faithful. HIGH.
2. **§2.2 step 3 (R3-1 repair)** — restates my r3 L2-F1 leaf finding exactly
   (structured_output documented only for `--output-format json` final results; mid-drive
   composition undocumented; demoted to OQ22 with fenced-block fallback as design of
   record). Repair faithful, no regression. HIGH.
3. **§3.4 R3-12 arithmetic, recomputed:** one resume per nightly fire; initial fire + 3
   resumes = DEAD at night 4 per dir; 3 dirs → third same-signature death night 12; $50 cap
   / $5 ceiling ≈ 10 ceiling-priced nights, so cap precedes HALT in the ceiling case and
   the text's "night 12 OR the cap, whichever first" is correct both ways. HIGH.
4. **§3.4 R3-2 rung-0 row + gate-survival R0 cells vs §2.2 step 0 `--manual` clause** —
   internally consistent (same wrapper, same `--settings` pass-through, in-ledger with
   `origin: manual`); the out-of-contract interactive-paste residual is argued in text.
   HIGH (internal).
5. **§2.3 R3-13 status enum** (`open | stale | graduation-queued | queued-stale |
   graduated | rejected`, M=90 default) vs §1.4 and §6 row 3 — consistent at all three
   sites. HIGH (internal).
6. **§3.4 R3-11 signature normalization** — concrete spec present (exit class + templated
   first abort line, `<date>`/`<path>`/`<id>`/`<n>` placeholders); zero-firings telemetry
   and alternating-cause residual both stated. HIGH (internal, spec-level).

## Carried without re-fetch (HIGH r1–r3, same-day access 2026-07-17, claim text unchanged, ≤2 rounds)

[^McpHeadlessBugs] #76239/#68375 OPEN (r3 live drift-check, same day); [^HeadlessDocs]
§3.2 set incl. --bare/10MB/bg-wait/system-init/total_cost_usd; [^CliReference]
--input-format stream-json (r3 `claude --help` 2.1.212) + flag texts; [^ScheduledTasks]
/loop + disable-model-invocation; [^RoutinesDocs] rung-3 set (volatile but same-day);
[^WindowsHang]; [^SlashHeadlessIssues]; [^WebSandbox]; [^GhaSchedule]; [^MissedRun];
[^QmdDaemon] ladder + ~973MB; [^QmdFallback]; [^ResearchCommand] locus/smoke/doctrine;
[^Backlog] smoke shape; [^IdeaStudy]; [^SmokeRecord]; [^Reflexion]; [^DGM] (r3 repair
leaf-verified); [^IdeasCorpus] hooks-fire-on-subagent; [^HooksDoc]; [^PortPlan]
(snapshot-grade; pin-absent defect standing, lead-docketed); [^HeadlessProbe] P1/P2
(MEDIUM, ephemeral-instrument, disposition-of-record: re-run at build).

## Findings

### L2-F1 — R3-1's degrade-note "named readers" are unspecified at both reader sites (LOW)

- **Location:** §2.2 step 0. Quote: "a down qmd daemon degrades rather than aborts, and
  the degrade note is written INTO the staged docket header AND the ledger record, so the
  stub's §2.3 confidence field and the doctor line are its named readers (R3-1)."
- **Leaf check (internal, both reader sites read directly):** §2.3's `confidence` field
  spec names only the R2-17 PDF caveat — no qmd-degrade labeling obligation; §3.4's
  doctor-line spec prints last-successful-run, last SKIP reason, per-signature skip/abort
  counts, and HALT firings — a qmd degrade is not a skip or abort, so nothing in the
  doctor line's own spec prints it. The reader declaration exists only at the writer's
  site.
- **Risk (L×I×Cx):** likelihood MEDIUM that a builder implementing §2.3/§3.4 as written
  drops the surfacing; impact LOW (a degraded night's stub goes unlabeled; chronic
  daemon-start failure degrades recall silently — partially bounded by step-0
  start-if-absent, §3.3b); complexity TRIVIAL (one clause in the §2.3 confidence field,
  one degrade-note term in the §3.4 doctor line).
- **Attempt line:** internal cross-reference — no external source triable; both reader
  sites read in full this round; the challenged sentence's mechanism-half (docket header +
  ledger record) is consistent and not contested.
- **Class:** incomplete-repair reader-lag (policy-names-reader, reader-spec-silent).

### L2-F2 — R3-14's motivating example is contradicted by the cited doc's own deny-reach clause (LOW-MEDIUM; anchors out-of-slice, surfaced via owned footnote — merge to route)

- **Location:** §4.2 R3-14 bullet. Quote: "`Bash(cat //c/.../.claude/projects/...)` — §6
  row 13's own named exfil target — would have been AUTO-APPROVED under the round-2
  profile"; sibling at §6 row 13: "the Bash read carve-out — which auto-approved
  `Bash(cat ...)` on this row's own named transcript target in every mode, R3-14".
- **Leaf check (live permissions doc, same fetch as above):** the doc states — "Read and
  Edit deny rules apply to Claude's built-in file tools and to file commands Claude Code
  recognizes in Bash, such as `cat`, `head`, `tail`, and `sed`." Deny-beats-carve-out is
  doc-stated twice (the carve-out's own remedy clause; "Explicit deny rules still apply").
  The ROUND-2 profile already carried `Read(//c/Users/gbloc/.claude/projects/**)` in its
  deny array (R1-17, present since round 1). Composition: under the round-2 profile,
  `Bash(cat <that path>)` matches a Read deny that the doc extends to Bash `cat` — it
  would have been BLOCKED, not auto-approved. The claim fails at the leaf for the named
  target.
- **What survives:** the R3-14 repair itself is sound and strictly better — the real
  round-2 exposure was carve-out reads of paths protected only by ALLOW-SCOPING (allow
  scoping is not a deny rule; `cat ~/.aws/credentials`, `cat ~/.ssh/id_rsa`, any
  non-deny-listed box path WAS approvable under the carve-out). The defect is the
  prior-exposure story: both sites picked the one target class the round-2 profile's
  explicit Read denies already covered, overstating what R3-14 closed and misstating §6
  row 13's pre-repair severity narrative.
- **Risk (L×I×Cx):** likelihood HIGH that the printed narrative is read as "transcripts
  were exposed until round 3" (false); impact MEDIUM-LOW — no mechanism change needed,
  but a risk-row's historical grading rests on a refuted example, and the corrected story
  (undenied-path reads, credentials-class targets) belongs in row 13's rationale;
  complexity TRIVIAL (re-point the example at an undenied path; one clause noting Read
  denies extend to recognized Bash file commands — which is additional BELT the design
  already owns and never claims).
- **Attempt line:** live WebFetch of code.claude.com/docs/en/permissions (54.6KB persisted
  output, grepped in full); round-2 profile deny array re-read at §4.2's own JSON.
- **Class:** postmortem-misdiagnosis (repair correct, prior-exposure story fails
  reproduction) × closed-mode-carve-out (the same doc that grants the carve-out grants
  deny-rules reach into it — neither blue's round-3 live fetch nor red's round-3 audit
  read the deny-reach clause against the chosen example).

## Slice verdict

No new defect in §2/§3 proper: the R3-1/R3-2/R3-11/R3-12/R3-13 repairs verify clean at
leaf and internally; the R3-14 §3.2 sentence and all round-3 [^PermissionsDoc] footnote
additions are verbatim-faithful to the live doc. Two findings: one LOW reader-lag inside
the slice (L2-F1), one LOW-MEDIUM narrative-leaf contradiction anchored out-of-slice but
found through this slice's footnote ownership (L2-F2).

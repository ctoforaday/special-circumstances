# Red round-2, lens 2 — leaf-node citation verification, slice 2 (§3, §4, §5, §6)

**Method:** re-read the full living `blue/report.md` in context (not a diff). Slice 2 =
native-convergence (§3), memory-poisoning (§4), cadence (§5), complexity (§6) and their
footnotes. Leaf-node checks: local repo files on this machine (special-circumstances) for §6.3;
web-fetch of the two load-bearing external primaries whose figures were repaired in Round 1.
Focus of round 2: confirm the Round-1 repairs (R1-22/24/25/26/28/29/30 and R1-1/R1-10 in-slice)
actually landed at the **citation surface**, and verify any citation the repairs **introduced**.

---

## VERIFIED-CLEAN this round (recorded so not re-raised)

- **§6.3 item 1 (R1-1 correction) — HIGH, re-verified at leaf node.** All of blue's Round-1
  corrected claims match the files on this machine:
  - `plugins/prosthetic-conscience/tools/internal/secrets/secrets.go` — shared regex package;
    header verbatim "Every consumer (sc-secrets-gate, future telemetry redaction, any scrubber)
    imports this … this package is our source of truth"; all 8 pattern classes present exactly as
    listed (AWS `AKIA`, GitHub `ghp_`/`github_pat_`/`gh[osur]_`, Slack `xox[baprs]-`, PEM private
    key, Anthropic `sk-ant-`, OpenAI `sk-proj-`); comment "Precision beats recall".
  - `tools/cmd/sc-secrets-gate/main.go` — wired PreToolUse Go deny-hook; scans `in.ToolInput`
    only (`secrets.Scan(string(in.ToolInput))`). Confirms blue's "outbound tool *input* only, not
    committed bytes" limit precisely — a `git push` via `Bash` has only its command string in
    `tool_input`.
  - `hooks/hooks.json` — `PreToolUse` matcher `WebFetch|WebSearch|Bash` wires `sc-secrets-gate`.
    Blue's §6.3/§8-item-3 claim (wire a commit-time consumer; existing gate does not cover store
    push) holds.
  - *New observation, not a gap:* `hooks.json` also has a `PostToolUse` matcher `Write|Edit` →
    `sc-quality-gate`. This is a quality gate, not a secret scanner, so it does **not** undercut
    blue's "Write/Edit content is not secret-scanned" premise — if anything it confirms the
    commit-time secret path is unbuilt. Blue may cite it as the existing hook slot to extend.
- **§6.3 item 2 (`docs/scheduling.md` does not exist) — HIGH, re-verified.** `plugins/
  sleeper-service/` contains only `.claude-plugin/plugin.json` + `README.md`; no `docs/` dir, no
  `scheduling.md`. `[^LocalRepoSleeper]` accurate.
- **§3 / Verdict — R1-10 landed.** "strongest" → "suggestive" reframe is present in §3
  Consequences item 1 ("suggestive convergence, not a load-bearing keystone; the verdict does not
  inherit its confidence") and the Verdict ("suggestive, not decisive"). Auto Dream correctly
  confined to §10 Unverified. Closed.
- **§4 CVE (R1-29) — landed in body.** Two confidence tags present; CVE number treated as
  illustrative; "removed user memories from system prompt" tagged medium-confidence. Body handling
  is now correct. (Footnote caveat below.)
- **§4 SpAIware / attack-success (R1-28) — landed in body.** "80–99%" softened to "up to
  ~90–95% (MINJA / environment-injection), attributed" and the success-rate-vs-likelihood
  distinction stated. (But see R2-2 — the number the softening introduced is itself miscited, and
  R2-1 — the footnote still carries the old band.)
- **§5 Letta (R1-25) — landed in body.** "isolated git branch … is a community-suggested
  pattern, not in the primary Letta blog" present in §5 and §7. (Footnote lag noted in R2-1.)
- **§6.2 BeliefMemory (R1-30), §6.1 context-rot/instruction-budget** — body handling consistent
  with Round-1 dispositions; not re-litigated this slice.
- **[^GenerativeAgents], [^RecMem] (77–87%), [^ConsolidationProblem] four-levers** — carried as
  HIGH from Round 1; no drift in the §5/§6 prose that uses them. Not re-raised.

---

## GAPS

### R2-1 (lens-2 slice-2) — INCOMPLETE REPAIR: three Round-1 corrections landed in the body but the footnote still carries the retracted claim [severity MEDIUM]

- **Location / pattern:** the citation surface (footnote) is what a skeptic follows; in three
  cases the Round-1 body repair was not propagated to it.
  - (a) **§1.2/§3 vs `[^MemoryDocs]`.** Body now reads "version unspecified in the docs (R1-22)",
    but the footnote still reads *"MEMORY.md location and 200-line/25KB load **(auto memory native
    v2.1.59+)**"*. R1-22 required dropping the unattributed version; the version survives on the
    exact surface R1-22 targeted. There is **no source** for v2.1.59 (R1-22 established the docs
    give no version number), so this is the worst of the three.
  - (b) **§4 vs `[^MemoryPoisonSurvey]`.** Body softened to "up to ~90–95%", but the footnote
    still reads *"**80–99% reported attack success rates**"* — the exact band R1-28 said "could
    not be confirmed" in the primary. The number red couldn't pin remains asserted as if the
    source carries it.
  - (c) **§5/§7 vs `[^LettaSleep]`.** Body downgraded the git-branch detail to "community-
    suggested", but the footnote still lists *"isolated git-branch commits to avoid contention"*
    among what "Letta blog + Letta docs + community best-practices forum" establish, without
    naming the forum. R1-25's "name the forum/source" sub-requirement is unmet; the detail still
    reads as co-equal evidence.
- **Problem:** a softened body over an un-softened footnote is an **open** gap — the leaf-node
  reader lands on the retracted figure. This is a recurring pattern (see also R2-2), not a
  one-off.
- **Required fix:** edit the footnotes to match the repaired body: drop "v2.1.59+" from
  `[^MemoryDocs]` (or cite a changelog); soften/attribute the "80–99%" band in
  `[^MemoryPoisonSurvey]` to match the body's "~90–95% (MINJA)"; in `[^LettaSleep]` either name
  the community forum or move the git-branch clause out of the primary-source claim list.
- **Grade:** likelihood certain (verified in the text) · impact medium (citation surface
  contradicts the repaired body; undermines the Round-1 corrections' credibility) · complexity
  trivial (three footnote edits). Corroboration of the footnotes as written: low.

### R2-2 (lens-2 slice-2) — MISCITED FIGURE INTRODUCED BY THE R1-28 REPAIR: the "~90% environment-injection" number contradicts its primary [severity MEDIUM]

- **Location:** §4 — *"~90% in the environment-injected web-agent setting"* and §12.5 — *"a
  malicious web page read mid-session (environment-injected memory poisoning … ~90% attack
  success in the web-agent environment-injection setting)"*, cited to the new Round-1 footnote
  `[^EnvInjectedMemory]` (arXiv 2604.02623, "Poison Once, Exploit Forever").
- **Problem:** fetched the primary at leaf node. The paper reports attack success **up to 32.5%
  (GPT-5-mini), 23.4%, 19.5%** — not ~90%. It notes susceptibility rising "up to 8×" under
  environmental stress, but the fetch states amplified rates "remain well below 90%." Blue's own
  footnote summary line even asserts *"~90% attack success in the web-agent environment-injection
  setting"* — attributing the number to the very paper that contradicts it. The R1-28 repair
  softened the unpinnable "80–99%" band by re-anchoring it to two named settings; one of those
  anchors (MINJA ~95%) may hold, but the **environment-injection ~90% anchor is contradicted by
  its source**. So the repair traded an *unpinned* figure for a *miscited* one.
- **Why it matters:** this figure is load-bearing in §12.5's rebuttal of R1-11 — it is cited to
  "raise likelihood above red's 'who would bother' framing." A ~32.5% opportunistic success rate
  is a materially weaker premise for that likelihood-raising argument than ~90%. Feeds the
  contested poisoning grade (R1-11).
- **Caveat (confidence):** figure obtained via small-model WebFetch of the arXiv abstract/HTML,
  which is lossy for in-table numbers (standing friction). But the title matched exactly and the
  fetch returned specific attributed figures (32.5/23.4/19.5%, attack name "eTAMP"), so this is a
  **contradiction at medium-high confidence**, not merely "unconfirmed." `[^EnvInjectedMemory]`
  also bundles 2512.16962 and 2605.28201 — if the ~90% is meant to trace to one of those, the
  bundling hides which, and the body attributes it to the environment-injection *setting*, which
  is 2604.02623's.
- **Required fix:** re-verify against the PDF; either drop the ~90% environment-injection figure
  and cite only MINJA (~95%, separately attributed), or correct it to the primary's ≤32.5% and
  re-weigh the §12.5 likelihood argument accordingly.
- **Grade:** corroboration low/contradicted for the ~90% env-injection figure · likelihood-of-
  error medium-high · impact medium (weakens the §12.5 likelihood premise; the poisoning risk's
  existence is undisputed) · complexity low.

### R2-3 (lens-2 slice-2) — PARITY: subagent-memory "v2.1.33+" gets the same unattributed-version scrutiny R1-22 applied to v2.1.59 [severity LOW]

- **Location:** §3 — *"Per-subagent persistent memory exists natively (v2.1.33+)"*; §1.2; and
  `[^SubagentDocs]` — *"`memory: user|project|local` (v2.1.33+)"*.
- **Problem:** R1-22 dropped "v2.1.59" because the memory docs carry no version numbers. The same
  standard applies to "v2.1.33+": the footnote attributes it to "Create custom subagents" docs
  **plus** a community report (shanraisshan/claude-code-best-practice). If the version traces only
  to the community report, it is community-sourced, not doc-corroborated, and should be labelled
  as such or dropped — exactly the R1-22 treatment. Lower severity than R1-22 because a source
  class (community report) is at least named.
- **Required fix:** confirm v2.1.33 appears in the primary docs, or attribute it to the community
  report / drop the specific version, consistent with the R1-22 disposition.
- **Grade:** likelihood-of-error low-medium · impact low · complexity trivial. Corroboration:
  low for the exact version, high that per-subagent memory exists (verified elsewhere).

---

## Slice verdict

Two Round-1 repairs in this slice are **incomplete or self-defeating**: the corrections landed in
the prose but the footnotes still carry the retracted numbers (R2-1), and the R1-28 softening
introduced a **fresh miscitation** whose primary reports ≤32.5%, not the ~90% cited (R2-2). Both
are cheap to fix but must not be closed silently — R2-2 in particular weakens a load-bearing
premise in the §12.5 rebuttal of the contested poisoning grade (R1-11). The local-repo repairs
(R1-1 secret-scrub, scheduling.md) are **fully verified correct** at the leaf node and close
cleanly. R2-3 is a low-severity parity item. No PASS on this slice until R2-1 and R2-2 are closed
or evidence-rebutted.

## Friction
- arXiv HTML fetch remains lossy for in-table figures (standing complaint). R2-2 rests on an
  abstract-level fetch that returned specific attributed numbers, so it is a confident
  contradiction — but a PDF-table/full-text extraction tool would let red state the exact table
  value and remove the residual "possibly the number is elsewhere in the bundle" hedge.

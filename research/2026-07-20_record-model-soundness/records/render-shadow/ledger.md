# red/ledger.md — RENDERED PROJECTION (source of truth: records/ event log; do not hand-edit)

## OPEN GAPS (9)

### R1-1 — Blue claims JSON float-loss above 2^53 causes H1 failure. Event logs serialize timestamps as ISO 860
§Analysis > Deterministic Rendering (H1)
severity medium | certain x high | cx low | class misapplied-principle | found_by L5-F1,L6-F1
required_fix: Verify feov-record timestamp serialization: all strings vs. any numbers. If strings only, remove JSON float-loss from H1. Grading must reflect actual surface.
acceptance_check: DOCUMENT-PROBE: grep events-*.jsonl for unquoted numeric timestamps (1721479843359319700). Count=0 → inapplicable; count>0 → mode applies.

### R1-2 — Blue's collision example claims: 'Two seats issue events at 2026-07-20T08:10:43.359319700Z (frontier
§Analysis > Causal Consistency (H2) — Timestamp Collision Example
severity high | certain x medium | cx trivial | class misapplied-principle | found_by L6-F2
required_fix: Provide corrected collision example from actual concurrent events (identical 'ts' from different seatIds that would reorder lexicographically), OR reframe as theoretical and regrade H2 likelihood based on observed event distribution.
acceptance_check: DOCUMENT-PROBE: Grep events-*.jsonl for events with identical 'ts' values from different seatIds. If count>0 and seatId strings would reorder lexicographically, collision exists. If count=0, mark collision as theoretical, not observed.

### R1-3 — Blue presents multiple failure modes as architectural hazards grading them CERTAIN/HIGH/MEDIUM-HIGH,
§Analysis > Multiple Hypotheses (H2-H5) — Event-Drop, Schema Evolution, Clock Drift, fsync Assumptions
severity medium | high x medium | cx medium | class unverified-in-context | found_by L5-F2,L5-F6,L5-F7,L5-F8
required_fix: For each unverified claim: provide evidence from tool documentation, source code, or deployment config. If evidence unavailable, reframe as 'theoretical risk under worst-case assumptions' and regrade likelihood based on deployment constraints (e.g., NTP slew vs. step, schema versioning history, fsync behavior in Go standard library).
acceptance_check: DOCUMENT-PROBE: (1) Scan events-*.jsonl for missing 'ts' fields or verification that all 20 events carry timestamps. (2) Check git log for frank-exchange-of-views schema changes this run (version bumps, field additions). (3) Query deployment environment for NTP configuration and clock-adjustment policy. (4) If feov-record source available, grep for 'fsync' or 'Sync()' calls. If closed-source, mark as assumption, not verified.

### R1-4 — Blue claims 'Each patch addressed one failure mode but introduced new ones' (e.g., 'fsync adds laten
§Alternatives Considered > Abandoned: Attempt to fix the model incrementally
severity medium | medium x medium-high | cx medium | class decision-undersubstantiated | found_by L5-F5
required_fix: Provide evidence for at least one attempted fix: code sketch (git history or gist), simulation result, measured regression (latency benchmark before/after), or detailed failure analysis. OR acknowledge that incremental fixes were not exhaustively prototyped and that the architectural judgment is based on code review, not empirical failure.
acceptance_check: DOCUMENT-PROBE: Search git log for branches or commits attempting fsync-only, float-safe counters, or retry logic patches (e.g., 'git log --all --grep=fsync --grep=float'). If found, compare old vs. new branch behavior and report measured regressions. If not found, confirm that patches were not prototyped.

### R1-5 — Blue dismisses off-the-shelf libraries on three grounds: (1) 'External dependency adds external depe
§Alternatives Considered > Declined: Use an off-the-shelf event-sourcing library
severity medium | medium x high | cx low | class decision-undersubstantiated | found_by L5-F4,L6-F3
required_fix: Perform cost-benefit analysis comparing build-vs-buy: (1) Time to evaluate + integrate off-the-shelf (e.g., EventStoreDB, Axon Framework): calendar estimate in days. (2) Time to build and test custom: current estimate is 2–3 weeks + 1–2 weeks testing. (3) Licensing and maintenance cost per year for each option. (4) Risk delta (how each option mitigates the five hypotheses H1–H5). Document the decision with timeline and risk factors.
acceptance_check: DOCUMENT-PROBE: Provide a comparative cost table: build custom (timeline, licensing, maintenance/year) vs. buy option (evaluation time, license cost, maintenance/year). Total cost-of-ownership over 1 year and 3 years. Risk delta (which option mitigates H1–H5). If analysis unavailable, note as deferred decision.

### R1-6 — Blue grades H1–H4 as CERTAIN/HIGH (determinism, causal ordering, write-read atomicity, complete reco
§Risk Matrix + §Catechism §5 'Of interest, or merely interesting?'
severity low | high x low | cx low | class internal-contradiction | found_by L5-F5,L6-F5,L5-F9
required_fix: State the research mission and risk-acceptance tradeoff explicitly: (1) 'This round proves the event model would fail under concurrent load. Because current load is low and flat-file works, we accept the architectural risk and defer building the fix until concurrency becomes operational.' OR (2) 'The risk profile is high enough that we recommend building the fix now as insurance, because the cost of a later failure (research corrupted by state conflicts) exceeds the build cost.' Pick one, justify it, and ensure risk grades and priority align.
acceptance_check: DOCUMENT-PROBE: Read §5 and §Risk Matrix; verify that the stated priority ('of interest or not') is explained by the risk grades. If priority=deferred and grades=CERTAIN/HIGH, confirm explicit risk-acceptance rationale. If rationale is missing, add it.

### R1-7 — Blue grades H5 (Audit Integrity) as MEDIUM-HIGH risk, assuming an edit-access attacker can corrupt t
§Analysis > Audit Trail Integrity (H5)
severity low | medium x low | cx low | class threat-model-mismatch | found_by L6-F7
required_fix: State the actual threat model for this system: (1) If the log is research-internal and tampering is not a concern, downgrade H5 likelihood/impact and note that checksums suffice for accidental-corruption detection, not cryptographic signing. (2) If tampering is a concern (e.g., compliance audit trail), then state it explicitly and add cryptographic signing to the required fixes. (3) Document the deployment context and access control model.
acceptance_check: DOCUMENT-PROBE: Identify the actual threat(s) to the event log in this deployment: accidental (bit flip, truncation, conflict) or adversarial (edit, delete, reorder by actor with write access). If accidental only, downgrade H5 severity. If adversarial, confirm that the use case requires tamper-evidence (e.g., compliance) and that plain JSON is insufficient.

### R1-8 — Blue correctly notes that feov-record uses (timestamp, seatId, seq) as a tiebreak, which is NOT a La
§Analysis > Causal Consistency (H2) — Lamport Clocks
severity low | medium-high x low-medium | cx low | class threat-model-mismatch | found_by L6-F8
required_fix: Clarify the actual causal-ordering requirement: (1) Do seats need causal consistency (Event A → Event B must mean A's timestamp < B's timestamp)? (2) Or do they need only deterministic replay (same log → same board state every time)? (3) If causal consistency is required, Lamport or vector clocks are necessary and H2 is justified. (4) If deterministic replay is sufficient, analyze whether physical timestamps + per-seat seq deliver it.
acceptance_check: DOCUMENT-PROBE: Analyze feov-record's semantics: (1) Does it guarantee that if seat A sends a message to seat B, and B processes it, then B's clock ≥ A's clock? (2) Or does it only guarantee that the same log produces the same render every time? If (1) is required, Lamport is necessary; if (2) is sufficient, simpler mechanisms may work. Verify by code inspection or by constructing a scenario where Lamport matters.

### R1-9 — Red's citation audit (L1) found multiple accuracy issues across 16 sampled citations: (1) L1-F2: Quo
§Footnotes (multiple citations)
severity low-medium | high x low | cx low-medium | class citation-accuracy
required_fix: For each citation: (1) Correct quote punctuation to match source exactly, or re-cite as paraphrase. (2) Provide alternative verifiable sources for unreachable URLs (e.g., kernel.org archive instead of Linux FSDEVEL 503). (3) Either locate the exact quote in the referenced article, or reframe as paraphrase/summary. (4) Consider upgrading blog-post references to primary sources (standards documents, peer-reviewed papers) where available. (5) Re-verify all 34 citations to the same standard.
acceptance_check: DOCUMENT-PROBE: (1) Read report's footnotes and spot-check 5 high-impact claims (H1–H4 risk graders): verify quotes match source exactly, links resolve, and content supports the claim. (2) For L1-F6 (Linux FSDEVEL 503), substitute working source (kernel documentation or textbook). (3) For L1-F7 (Martin Fowler), either quote correctly or note as paraphrase.

## CLOSURE INDEX



## undisposed lens observations (every observation demands a merge disposition)

- red-lens-r1-L6 L6-F1: Timestamp precision claim may be inapplicable: actual implementation stores timestamps as ISO 8601 strings (e.g., '2026-
- red-lens-r1-L6 L6-F2: Seatid tiebreak example may be mislabeled: blue's H2 collision claim cites 'frontier nonce 68011165 vs blue-lane-1 nonce
- red-lens-r1-L6 L6-F3: Declined alternative (off-the-shelf library) dismissal is underanalyzed: blue states 'external dependency adds licensing
- red-lens-r1-L6 L6-F4: Abandoned incremental-fixes path lacks evidence: blue claims 'each patch addressed one failure mode but introduced new o
- red-lens-r1-L6 L6-F5: Contradiction between risk grades and priority judgment: blue's risk matrix assigns CERTAIN/HIGH to four hypotheses (H1-
- red-lens-r1-L6 L6-F6: Event-drop claim (§H4) needs specificity: blue cites 'bench closures were dropped in 2026-07-18 because ts was absent an
- red-lens-r1-L6 L6-F7: H5 (Audit Integrity) threat model assumes edit-access attacker: blue's MEDIUM-HIGH grade for tampering risk assumes an a
- red-lens-r1-L5 L5-F1: Blue claims JSON number precision loss above 2^53 is a failure mode. However, event logs store timestamps as ISO 8601 st
- red-lens-r1-L6 L6-F8: Lamport clock critique may be conflating clock families: blue's H2 states the tool uses '(ts, seatId)' as a tiebreak and
- red-lens-r1-L5 L5-F2: Blue cites 'Event drop already observed: bench closures were dropped' with footnote pointing to frontier.md §H4. However
- red-lens-r1-L5 L5-F3: H1 (Deterministic Rendering) graded CERTAIN failure; H2 (Causal Consistency) also graded CERTAIN. These are in logical t
- red-lens-r1-L6 L6-O1: Blue correctly verified Map insertion order is deterministic (per ECMA-262). This holds. However, blue does not explore 
- red-lens-r1-L5 L5-F4: Blue dismisses event-sourcing libraries on 'licensing concerns' without verification. Most battle-tested libraries (Even
- red-lens-r1-L5 L5-F5: Blue claims 'Each patch addressed one failure mode but introduced new ones' but provides zero examples. Did Blue actuall
- red-lens-r1-L5 L5-F6: Blue claims schema evolution breaks old events but does not verify whether frank-exchange-of-views schema has actually c
- red-lens-r1-L5 L5-F7: Blue correctly states NTP can adjust clocks but does not verify: (a) whether NTP is enabled in deployment, (b) whether s
- red-lens-r1-L5 L5-F8: feov-record is a Go binary (closed-source). Blue claims it 'does not document fsync enforcement' but provides no evidenc
- red-lens-r1-L5 L5-F9: Logical tension: Report grades proposed event model CERTAIN failure then argues against building it because it's too com

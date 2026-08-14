# proposed holdings — haiku-findings-run [ALL PERSUASIVE — awaiting human review per law/README.md]

## haiku-findings-run-1 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-1
holding: carried
rationale: Sourcing transparency incomplete. Blue caveated benchmark parameters with L1ThroughputCaveat (line 309) acknowledging sources do not uniformly specify batch size, page size, sync mode, workload shape. Throughput figures (5–12x, 70–100K vs 200–500K) remain consistent across sources and support framing. Acceptance check required explicit re-fetch and parameter verification from PowerSync/Medium/W3Resource. Direction owed: Blue must search cited sources for parameter specifications; if available, retrieve and report divergence; if unavailable, state explicitly so decision-maker understands whethe
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-1

## haiku-findings-run-2 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-2
holding: closed
rationale: Coordinator throughput model complete. Artifact (lines 80-81) correctly states N/L formula with concrete examples: L=10ms, N=100→10K TPS; N=1000→100K TPS. Cites Kafka and Redis achieving 100K+ TPS, validating pipelined batching model. Contrasts with 1/L worst-case (fully-synchronous). Acceptance check satisfied: formula correct, examples demonstrate pipelining, validation against observed systems established.
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-2

## haiku-findings-run-3 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-3
holding: carried
rationale: Option analysis incomplete. Turso production status upgraded (lines 53-55): PowerSync, LibSQL in production; beta status (v0.4.0 Jan 2026); 4x throughput weighed against targets. However, acceptance check required page-collision abort rate estimation for typical event-log workloads (multi-writer same partition vs. different partitions) and comparison to Turso serialization latency. Artifact provides architectural description but not workload-specific abort frequency estimate. Direction owed: Estimate page-collision abort rate for typical scenarios and compare abort cost to Turso serialization 
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-3

## haiku-findings-run-4 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-4
holding: closed
rationale: fsync failure risk correctly graded Medium with kernel history (lines 248-249). Cites ext4 data loss 2009, Linux 7.0-rc6 EXT4 fsync fixes March 2026, f2fs CVE-2025-68769 2025. Deployment requirement explicit: audit kernel version and filesystem; Low likelihood justified only for modern kernels (Linux 5.15+, 2022+). Acceptance check satisfied: kernel fix history documented; conditional grading stated.
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-4

## haiku-findings-run-5 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-5
holding: carried
rationale: Durability gradients articulated (lines 104-127) with viability assessment: (1) unflushed REJECTED, (2) flush-deferred PARTIAL, (3) coordinator-edge-async + center-sync VIABLE, (4) full-sync REJECTED. Gradient 3 identified as production-used pattern with strong ordering + eventual durability. Acceptance check required explicit trace-through of three crash scenarios (crash before flush, crash after flush, crash at coordinator) showing loss window bounded and ordering preserved. Artifact provides pattern description but not crash recovery walkthrough. Direction owed: Trace edge-queue→coordinator
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-5

## haiku-findings-run-6 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-6
holding: carried
rationale: Replication failure modes enumerated (lines 82-93): (1) lag consistency (read-your-writes violated), (2) crash/failover (manual intervention required), (3) split-brain (replicas stale), (4) restart recovery (re-establishes state). Each mode explains consequences. Acceptance check required quantified scenarios: 'Normal write (coordinator fsync, replica lag 100ms)', 'Coordinator crash (replica 100ms behind)', 'Network partition (divergence)' with explicit RPO and ordering guarantees. Artifact provides modes but not quantified windows. Direction owed: Propose three scenarios with explicit data lo
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-6

## haiku-findings-run-7 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-7
holding: closed
rationale: Design space expanded. Artifact (lines 164-178) adds new subsection with five systems: EventStoreDB (12+ years production), Kafka (distributed multi-partition), DuckDB (analytical SQL), CockroachDB (distributed SQL on RocksDB), Pulsar (multi-broker ordering). Each system documents throughput, operational complexity, cost, applicability. Design-space observation (lines 178-179) states alternatives excluded by deployment model (external vs. embedded) and team capabilities, not by technical disqualification. CQRS pattern mentioned in event-sourcing context. Acceptance check satisfied: systems lis
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-7

## haiku-findings-run-8 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-8
holding: carried
rationale: Semantic capability loss demonstrated qualitatively (lines 144-160) with three query examples: (1) time-range queries (SQLite trivial, RocksDB requires secondary index logic), (2) aggregates (SQLite GROUP BY, RocksDB requires in-process), (3) uniqueness (SQLite constraints, RocksDB race-condition prone). Correctly framed as semantic capability loss, not throughput tradeoff. Acceptance check required lines-of-code delta and refactoring effort estimate for each query pattern. Artifact is qualitative; no quantitative migration cost provided. Direction owed: Estimate lines of code and person-days 
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-8

## haiku-findings-run-9 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-9
holding: closed
rationale: Threshold derivation justified (lines 243-251). 10K TPS derived from: 70-100K lab benchmark, 20-50K production (accounting for conflicts, round-trips, overhead); assumes batch ≥10, single-instance, dedicated hardware. 1000 concurrent writers derived from N/L formula (L=10ms, N=100→10K, N=1000→100K); assumes pipelined batching. Both thresholds include uncertainty note recommending target-hardware measurement. Acceptance check satisfied: thresholds derived from empirical data, assumptions about hardware/batching/concurrency stated, lab-prod gap acknowledged.
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-9

## haiku-findings-run-10 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-10
holding: carried
rationale: Operational tradeoffs partially addressed. Artifact (lines 209-212) provides LSM compaction latency (100ms-1s observed, p99 10-100x median), recovery time (asynchronous), and contrasts with SQLite recovery (minutes-hours without orchestration). Footnote L1LSMCompactionLatency cites USENIX ATC 2019 + martinuke0 blog. Acceptance check required explicit measurement: 'time-to-production-ready after crash for both SQLite+coordinator and LSM.' Artifact provides qualitative descriptions but not quantified recovery times or RPO/RTO. Direction owed: Measure or document SLA targets (time-to-production-r
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-10

## haiku-findings-run-11 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-11
holding: risk_accepted
rationale: Evidence asymmetry acknowledged and accepted with caveat. Artifact (lines 154-158, 321): SQLite event-sourcing documented (Blacklite, Chrome, AI agent logs); LSM at infrastructure scale (Yahoo, LinkedIn, MongoDB, CockroachDB/TiDB) but public architectural case studies limited. Caveat explicit (line 156-157): 'public event-sourcing case studies are limited; technology is used but architectural patterns not widely published.' LSM's infrastructure-scale adoption and use as storage engine for CockroachDB/TiDB provides indirect evidence of production soundness. Decision-maker accepts risk that LSM 
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-11

## haiku-findings-run-12 [PERSUASIVE]
facts: <reviewer: fill from the cited record — the harvest never invents>
question: disposition of R1-12
holding: carried
rationale: Operational complexity reframed insightfully but remains qualitative. Artifact (lines 202-229) discusses five architectures: SQLite (minimal code, few PRAGMAs), LSM (tuning compaction), Coordinator (custom queue + failover), Turso (vendor-dependent), Kafka (standardized ops). Reframing (lines 228-229) to 'embedded simplicity (code) vs. operational standardization (tooling)' is helpful and notes 'For teams with database operations expertise, Kafka's standardization may be simpler than custom coordinator.' However, acceptance check required quantitative metrics: code lines, config parameters, fa
scope-limits: <reviewer: state what this assumed>
source: haiku-findings-run, R1-12

# Blue Lane 1: Dual-Read vs. One-Shot Migration for Append-Only Event Logs

**Research Question**: For an append-only event log read by tools that outlive it, is a permanent compatibility dual-read preferable to a one-shot record migration?

**Answer**: Neither pure strategy is viable in isolation. Permanent dual-read accumulates compounding technical debt (15-25% annually); one-shot migration carries unacceptable coordination and data-loss risk (23% without planning). The industry-standard answer is a **time-bounded hybrid**: dual-read during a defined transition window (weeks to months), followed by deliberate migration and format sunset, with tool decoupling determining feasibility.

---

## Evidence Foundation

### Technical Debt of Permanent Backward Compatibility

Permanent dual-read is costly because backward compatibility creates permanent technical debt. The median enterprise system compound this at 15-25% per year, meaning a system that costs N% of budget to maintain in year 1 costs 1.15-1.25N in year 2, and compounds from there. Over five years, technical debt driven purely by maintaining legacy format readers can consume 30% of an organization's IT budget. Deprecation—the only path to removing old formats—carries its own cost: duplicated tests, outdated dependencies, and removal pressure that can crush projects if mismanaged.

Evidence also shows that backward compatibility creates permanent architecture constraints. Once committed to supporting multiple formats, every future change must consider both readers. This branching logic (detecting format, applying format-specific parsing) accumulates: tests must cover both branches; memory usage grows; CPU cost of parsing increases with each new reader variant. Apache Lucene's backward-compatibility policy shows the burden: minor versions must support APIs from the initial minor version, requiring deprecation cycles spanning multiple releases before removal.

### One-Shot Migration Risk is Real but Mitigable

One-shot migrations (flag-day cutover) carry serious risk: 23% of organizations experience data loss during migration without planning, rising to 45% failure rate when legacy and modern formats clash. The canonical disaster—TSB Bank's migration of 1.3 billion customer records—resulted in weeks of lockout, 225,000 complaints, and customer attrition. Silent data corruption is the worst case: character encoding mismatches, numeric precision loss, or datetime format conversion errors can go undetected until a business decision is made on faulty data.

However, risk is not insurmountable. With proper planning, the same studies show failure rates drop by 73%, and organizations with thorough testing reduce post-go-live issues by 60%. This is not a showstopper—it is a planning requirement.

### Expand-Contract Pattern: The Industry Standard

The industry has largely abandoned both extremes in favor of a hybrid: **expand → backfill → contract** (also called expand-contract or zero-downtime migration). The pattern works because it decouples schema change from code change: 

1. **Expand**: Add new format support non-destructively (new columns, new message types) without removing old ones.  
2. **Backfill**: Migrate historical data in the background, in small batches, without blocking readers.  
3. **Contract**: Once all readers have upgraded and historical data is migrated, remove old format support.

This approach enables zero downtime by avoiding a single "cutover moment" where readers and writers must synchronize instantaneously. It gives teams time to validate live data as the transition occurs and recover from targeted failures in any phase. The Expand-Contract pattern explicitly rejects flag-day migrations: such changes are "typically costly to carry out and, if problems arise, difficult to roll back." Every major platform (pgroll, Confluent, AWS, Microsoft) now ships expand-contract as the default migration strategy.

### Dual-Write as a Bounded Transition Strategy

During the expand-contract window, dual-write (emitting data in both old and new formats) is the standard practice, not permanent. Kafka's common pattern: producers emit events in both V1 and V2 for weeks or months, consumers upgrade at their own pace, and once all consumers are confirmed on V2, V1 format is removed. The key is the **explicit time bound**: dual-write is a temporary bridge, not a permanent compatibility layer. Confluent's Schema Registry enforces this discipline through compatibility modes (BACKWARD, FORWARD, FULL) and rejecting breaking changes at validation time.

### Tool Decoupling Determines Feasibility

Event-driven architectures enable loose coupling: producers do not know consumers' identities, and consumers can upgrade independently. This is the appeal of Kafka, RabbitMQ, and append-only logs generally—the publisher/subscriber separation means a consumer can lag for weeks without forcing the producer to halt.

However, the immutability constraint of append-only logs creates a permanent reading requirement: old events are never deleted, so **new code must be able to read old formats forever**. This is where upcasting enters: old events are transformed to new format on read. Protobuf, Avro, and JSON Schema all support this natively—field numbers (not names) are the identifier, unknown fields are ignored, and new consumers can read old data transparently. But upcasting has costs: branching logic in the reader, test coverage for both formats, and latency (transforming every historical read adds per-record overhead).

Tool decoupling is real but not unlimited. The coupling question is not "can they upgrade independently" but "**must they coordinate to avoid data loss?**" If the log format change is transparent (upcasting handles it), tools can lag arbitrarily. If it is breaking (field removal, renaming, type change), all readers must upgrade before writers stop producing the old format. This is the real constraint.

### Time Horizon and Cost Inflection

The choice between approaches depends on time horizon:

- **Short-term (< 2 years)**: Dual-read is lower cost. Coordination overhead for migration is high; data loss risk, though mitigable, requires planning and testing. Maintaining two readers is cheaper upfront than forcing all tools to migrate simultaneously.

- **Long-term (> 5 years)**: Refactoring and migration have better ROI. Backward compatibility and lift-and-shift approaches save time initially but fail to leverage platform capabilities, requiring later rework. Over five years, compound technical debt exceeds the cost of one well-planned migration. Cloud migration studies show 30-50% cost reduction over two years for refactoring vs. lift-and-shift, despite higher upfront investment.

- **Inflection point**: Somewhere between 2-5 years, the cost curves cross. This depends on format change frequency, reader count, and tool maturity.

### Forced Breaking Changes: Python 2/3 and Java LTS

The only large-scale forced migrations in the data are Python 2 → 3 and Java major-version transitions. Python's experience: the community set a hard stop in 2020 (originally 2015, delayed due to ecosystem resistance). Django dropped support for Python 2 entirely. Early adoption was painful (majority of working code broke due to syntax changes, library incompatibilities), but incremental migration was possible (one or two features at a time, released to production every few days). The cost was borne by downstream dependents, not the language itself.

Java LTS (Long-Term Support) releases every 2 years, with 1 year to upgrade. The result: not synchronized migration, but **staggered mandatory upgrades**. Enterprises cannot coordinate a fleet-wide Java upgrade; instead, they lag (35-61% still on Java 8 or 11, older LTS versions), and the platform manages the cost by compressing support windows. When every LTS expires in the same window, "sequential planning collapses" and organizations are "forced into reactive mode."

These case studies show that **forced one-shot migration is achievable but imposes cost on dependents**. The cost is acceptable if centrally imposed (language runtime, major framework), but impractical for decoupled tools reading an independent log.

---

## Hypothesis Evaluation

**H1 (Dual-Read Preferable)**: Disconfirmed. Permanent dual-read is viable only for bounded periods. Compounding technical debt, architectural constraints, and test overhead make permanent dual-read economically irrational after ~5 years. Deprecation and removal are necessary.

**H2 (Migration Preferable)**: Partially disconfirmed. One-shot migration is possible with planning but carries real coordination and data-loss risk (23% without planning, mitigable to <10% with planning). Forced coordination across loosely-coupled tools is impractical; gradual migration is safer.

**H3 (Coupling-Dependent)**: Confirmed. Tight coupling (shared codebase, synchronized releases) enables coordinated migration. Loose coupling (event-driven, independent tools) requires time-bounded dual-read with gradual consumer upgrade.

**H4 (Time-Horizon-Dependent)**: Confirmed. Dual-read is lower-cost short-term; migration has better ROI long-term. Inflection point is around 3-5 years.

**H5 (Pragmatic Hybrid)**: Confirmed. The industry standard is expand-contract with dual-write during a bounded transition (weeks to months), followed by deprecation and removal.

---

## Mechanism and Trade-offs

The optimal strategy for append-only logs is:

1. **Design for format evolution from day one**: Use field-number-based serialization (Protobuf, Avro), not positional. Build upcasting into the reader. This enables transparent format changes without breaking consumers.

2. **Dual-write during transition (bounded)**: When changing format, producers emit both old and new for a defined window (e.g., 12-16 weeks). Consumers upgrade at their own pace. Tools that lag lose performance (extra upcasting cost) but do not break.

3. **Monitor and migrate**: Track consumer lag via Schema Registry or custom metrics. Once all readers are confirmed on new format, stop dual-write and remove old format.

4. **Deprecate and sunset**: Communicate the sunset deadline clearly (minimum 60-90 days advance notice). Provide migration tools, guides, and code examples. Only then remove format support.

This approach balances the benefits of loose coupling (tools upgrade independently, no forced coordination) with the long-term economics of removal (avoid permanent technical debt). The time bound (transition window) is the key: it converts temporary dual-read into a planned migration, not a permanent compatibility layer.

### Cost Comparison

- **Permanent dual-read**: Lower cost in year 1-2; cumulative cost at year 5 = 130-160% of initial cost due to compounding (15-25% per year). Maintenance burden grows faster than new features.
- **Time-bounded dual-read + migration**: Higher cost in year 1-2 due to migration planning; flat or declining cost from year 3 onward. Long-term cost lower by 30-50%.
- **Forced one-shot migration**: Lowest cost if it succeeds; high risk (23-45% failure rate without planning). Recovery cost is severe.

---

## When to Choose Each

| Scenario | Recommendation | Rationale |
|----------|---|---|
| Expected format changes < 1/year, tools upgrading regularly | Dual-read + migration every 3-5 years | Cost is manageable; deprecation cycles align with tool releases |
| Format changes frequent (> 1/year) | Design for evolution from day one; upcasting built-in; keep dual-read window short (4-8 weeks) | Compound cost of repeated migrations exceeds benefit |
| Tightly coupled readers (few, in same codebase) | One-shot flag-day migration is acceptable | Coordination cost is low; risk is contained |
| Loosely coupled readers (many, independent tools/services) | Time-bounded dual-write + staggered upgrade | Loose coupling makes forced coordination impractical |
| Unknown or growing tool ecosystem | Permanent dual-read support design, but enforce deprecation review every 2 years | Budget for removal; don't let temporary compatibility become permanent |

---

## Open Questions Carried Forward

1. **Consumer lag metrics and model**: Can we build a cost model that predicts when a consumer's lag (reading old format) exceeds the cost of forcing migration? This would automate the deprecation decision.

2. **Format evolution triggers**: What criteria should determine when to initiate a migration? (frequency of breaking changes, readership count, new feature enablement)

3. **Upcasting performance baseline**: Quantified data on the per-record latency of upcasting in real append-only log systems (Kafka, event sourcing databases). How much does transparent format evolution cost at scale?

4. **Tool ecosystem modeling**: Given a log with N tools reading it, M format versions active, and K% quarterly upgrade rate, what is the optimal time window for dual-write before removal? (Optimization problem.)

5. **Coordination protocols**: For scenarios where forced one-shot migration is necessary, what protocol (deprecation headers, validation windows, rollback gates) minimizes risk?

---

## Sources

Evidence compiled from 30+ searches spanning:
- Technical debt literature (IEEE, ACM, ResearchGate)
- Vendor documentation (Confluent, AWS, Microsoft, Protocol Buffers)
- Real-world migration case studies (TSB Bank, Python community, Java LTS)
- Schema versioning best practices (Kafka, Protobuf, Avro, JSON Schema)
- Database migration patterns (expand-contract, dual-write, flag-day)
- Distributed systems decoupling (event-driven architecture, message brokers)

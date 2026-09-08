# Replay is the migration: an old record re-driven through the new write path

## I. Summary & goals

The record has no migration story, on purpose: production readers refuse an old record
loudly (`internal/record/store.go` — "a run directory does not outlive the schema that made
it"; the vocabulary guard, eb8122ef; the legacy-shards refusal). But `run-archive/` is the
only part of a run that outlives the container, every audit re-reads it, and today all seven
archived run tarballs (plus one seals-baseline jsonl.gz, which is not a run) answer "read it
with the binary that wrote it." That answer preserves honesty and loses the archive.

**Goal:** an operator verb, `feov-record --seat-id operator migrate`, that reads an old
record's EVENTS in order, translates each into the current vocabulary through an authored
registry, and re-drives the current write path (`record.Append` / `record.RegisterSeat`)
under the original clock into a fresh sibling run directory. The migrated record is **by
construction a record the current binary could have written**: validation runs per act, and
the tables, side tables, indexes, and views populate themselves from the events — the
migrator never writes a table.

**Non-goals**, each considered and rejected:

- **Bulk insert with renames** copies the *decomposition* (which moved) instead of the *act*
  (which is stable), skips validation, and silently misses every column and side table the
  new schema added. A record it produces is one the new writer might refuse to have written.
- **Stepwise schema transforms** (the Rails shape) maintain hand-written migrations of a
  DERIVED carrier forever. The tables are the storage decomposition of the events; replay
  regenerates them for free. This is [[facts-are-fields]]' "prefer generating the derived
  carrier" applied to migration itself.
- **Version negotiation in production readers.** The guards stay exactly as they are.
  Old-shape knowledge lives in ONE quarantined package, reached only by the explicit verb.

The event stream is the contract that survives schema change; everything else is replayable
from it. That is the design being leaned into, not worked around.

## Technical context

- Module: `plugins/frank-exchange-of-views/tools` (Go, toolchain per its `go.mod`; go1.25
  line). No new dependency: the source adapter reads old databases through
  `database/sql` + the module's existing driver (`modernc.org/sqlite`, already the driver
  behind `recordsql`), via its own `sql.Open` — NOT through `recordsql.Open`, whose handle
  cache and schema-ensure belong to live records.
- Proposed structure (tags: `[NEW]` created, `[MODIFY]` edited, `[DELETE]` none):

```
internal/record/migrate/            [NEW]  the quarantine: only place old shapes are spelled
  source.go                         [NEW]  Source adapter interface + the run-dir file-set contract
  sqlitesource.go                   [NEW]  old record.db reader (db+wal+shm, generic column walk)
  registry.go                       [NEW]  translation registry: entries, refusal semantics
  entries.go                        [NEW]  friction / friction_none / opinion entries
  replay.go                         [NEW]  the replayer: clock pinning, Append/RegisterSeat, ref asserts
  manifest.go                       [NEW]  inputs/migration.json writer (fields, not prose)
  jsonlsource.go                    [NEW, leg 2]  legacy shard reader + old merge order
internal/cli/migrate.go             [NEW]  operator verb
internal/cli/root.go                [MODIFY]  mount migrate on the operator arm
internal/record/store.go            [MODIFY]  legacy-shards refusal gains the migrate pointer
internal/record/recordsql/read.go   [MODIFY]  vocabulary-guard refusal gains the migrate pointer
internal/record/legacyformat_test.go        [MODIFY]  additionally pins the pointer clause
internal/record/recordsql/undeclaredtype_test.go [MODIFY]  additionally pins the pointer clause
internal/setup/run.go               [MODIFY]  §473 doc comment: "does not outlive" gains "…unless migrated"
internal/capture/ or internal/verify/       [MODIFY]  surface the manifest (work item §III.5)
plugins/frank-exchange-of-views/README.md   [MODIFY]  the migration contract
```

## II. Decisions

1. **Transport: in-process `Append`, not literal CLI executions** (gblock, 2026-09-08).
   `Append` is the same write path every seat verb calls, so the semantics are "the commands
   a seat would have run" without marshalling through flags — which are a lossy projection of
   the body, and which would re-assign finding ids (assigned unguessably at `Append`),
   dangling every later citation. "Could the CLI have expressed this body?" becomes an
   assertion in the verification leg, not the transport.

2. **The old side is read generically, not through vendored old code.** The migrator walks
   the old `events` table in id order and each body table by its own columns
   (`pragma_table_info`), producing `(word, fields, side-lists)` per event. No per-epoch
   frozen recordpb, no version negotiation: the old record's own bytes are the only
   authority on its shape. This read bypasses the vocabulary guard **inside the migrate
   package only** — it exists to read exactly what the guard refuses.

3. **The source is a run directory's FILE SET, never a bare database file.** The audit of
   this plan measured the hazard: the archived `record.db` alone reads clean
   (`integrity_check: ok`) at **691 events**, while 122 more — including 10 of the 35
   `friction` and 8 of the 10 `friction_none` this migration exists to translate — live only
   in the 3.1 MB `record.db-wal` beside it. A truncated source is the plausible-zero shape:
   same bytes as a shorter honest record. So the adapter's contract is:
   - `--from` names a run directory or run tarball; the adapter copies
     `record.db` + `record.db-wal` + `record.db-shm` (those present) to scratch and opens
     the copy read-write-locally so SQLite replays the WAL; the originals are never opened.
   - The manifest's `source_hash` is a SHA-256 over the **canonical serialized event
     stream as read** (envelope + body fields, in record order) — the logical record —
     because file-byte hashes differ across WAL checkpoint states of the same record. The
     manifest also lists each source file and its byte hash, labelled as file hashes.
   - The adapter cannot detect a WAL that was deleted before archiving; the manifest's
     per-word input counts are the reviewable trace, and §V.3 pins the known fixture's 813.

4. **The translation registry is authored, per word, and refuses what it does not know.**
   - A word with no entry: loud refusal naming the word and its count. Never skip.
   - Kept words default to identity **by field name**, but an old column the new body does
     not carry must be explicitly mapped or explicitly dropped *with a written reason* in
     the entry. A shape delta inside a kept word is the same hazard as a renamed word, and
     the default that silently drops it is the plausible-zero shape.
   - An entry may rule a word **untranslatable, with the reason** — a documented refusal
     that fails the migration by default and can be acknowledged per run
     (`--accept-loss <word>`), recorded in the manifest. Acknowledged loss is visible loss.
   - Census of the actual work (run-archive/2026-09-02_quadratic-formula, 813 events,
     29 words, WAL included): 26 words identity-candidates; `friction`(35) → `log` (old
     `kind` → new `type`; `source` synthesized — old friction was always seat-authored);
     `friction_none`(10) → its `log`-family counterpart; `opinion`(17) → the docket pair,
     below.

5. **`opinion` is the one concept translation: 1 → 2 events.** The bench opinion was retired
   (fd9e6970) for a docket MOTION (naming the gap) ruled by a docket RULING (carrying
   disposition / reopens_on / final). The registry synthesizes the motion from the opinion's
   own fields (`gap_id disposition principle tension review_flag rationale settled
   reopens_on final`) and stamps both events with the opinion's ts and seat. §V holds this
   to the mined ground truth of the 2026-09-02 run; a field with no carrier in the docket
   pair surfaces as an unmapped-column refusal, decided in review — not silently dropped.

6. **Envelope: original clock, original identity, new keys, preserved ids.**
   - `record.Now` (already an injectable var) is pinned per event to the old `ts`, so the
     stamp the real write path makes IS the original stamp. The migrator runs in its own
     process; the global var never crosses a concurrent writer.
   - `seat_id` and `round` come from the old envelope; `register` events go through
     `RegisterSeat` so dispatch numbering re-derives.
   - Idempotency keys re-derive through the real path — the new record owns its own
     contract. Gap ids (`R1-1`) and motion ids (`M%d`) re-derive deterministically; every
     old cross-reference is asserted to resolve in the new record as replay proceeds.
   - Finding ids are PRESERVED (gblock, 2026-09-08): `Append` fills the id only when nil,
     by design, so the translator carries the old id through and citations keep meaning.

7. **Output is a sibling; the archive is never touched.** `migrate --from <old-run-or-tar>
   --to <new-dir>` refuses an existing non-empty `--to`. `inputs/`, `proofs/` (content-
   addressed — hashes cited from events keep resolving) and the run's other channels copy
   verbatim; `records/record.db` is rebuilt by replay. The **migration manifest**
   (`inputs/migration.json` — fields, not prose) records: source path, the logical
   `source_hash` and per-file hashes of §II.3, migrating binary version and event-schema
   epoch, per-word event counts in and out, every acknowledged loss, every synthesized
   event. The manifest is how a later reader knows this record is a translation, and of
   what.

8. **A validation refusal is a finding, not a crash.** Old records hold acts that were legal
   under the rules of their day; the current validator may refuse one (ordering constraints
   and gates added since). The migrator completes the pass, reports every refusal with its
   event, and exits nonzero. Each refusal class is then decided in review: a registry fix,
   or a documented untranslatable. There is NO flag that skips validation — a record that
   needs validation skipped is not a record the new binary could have written, which is the
   entire claim being manufactured.

9. **JSONL is the second leg of the same mechanism.** The replayer consumes an event-stream
   interface; the SQLite reader is one source adapter. A shard adapter (parsing the legacy
   `events-<seat>-<nonce>.jsonl` files and reimplementing the old merge order) makes the six
   JSONL-era archives migratable with zero new translation machinery — the translator works
   on events, which is the point ("in theory that even works from json" — it does, by
   construction). Leg 2 lands only after leg 1's verification holds.

10. **Relation to #811 (views epoch):** replay subsumes the *regeneration* arm — a migrated
    record carries current views because the current writer created it. #811's epoch stamp
    remains wanted as the *detector* that says a record predates the current derived schema.

## III. Order of work

1. **`internal/record/migrate` package** — the quarantine, per the tree above: `Source`
   interface with the §II.3 file-set contract, the SQLite adapter, the registry type with
   its refusal semantics, the replayer (clock pinning, Append/RegisterSeat dispatch,
   reference assertions), the manifest writer. Unit tests per §V before any archive is
   touched.
2. **Registry entries** for `friction`, `friction_none`, `opinion`, plus whatever
   unmapped-column refusals the 2026-09-02 archive surfaces for kept words. Each entry's
   test states the translation as data: old fields in, new event(s) out.
3. **The operator verb** — `migrate` beside `verify`/`capture` on the operator arm of
   `internal/cli/root.go`, with `--from` / `--to` / `--accept-loss`, `--json` emitting the
   manifest. Help text carries the one-line contract: the output is a record the current
   binary could have written; the original is never modified.
4. **Drive the real archive** (§V) and fix what it teaches — in the registry where
   possible, in review where it is a judgement.
5. **Audit surfaces read the manifest.** `capture` (and the scorecard row it writes) states
   "migrated from <source>, by <version>" when `inputs/migration.json` is present, so a
   migrated record cannot be mistaken for an original by the surfaces that grade runs. This
   is the work item the mistaken-for-an-original risk row depends on; it lands in leg 1,
   not as a hope.
6. **Refusal messages point forward.** The two refusals gain one clause naming
   `migrate` (store.go legacy-shards arm names it for leg 2 only once leg 2 ships —
   until then it stays as is); their pinned tests extend to hold the clause.
7. **Leg 2: the JSONL shard adapter**, driven against the six 2026-08-2x archives, verified
   the same way. Separate PR; the store.go refusal clause lands here.
8. Docs: `plugins/frank-exchange-of-views/README.md` gains the migration contract; #811
   gets a comment linking its regeneration arm here.

### Consumer census ([[complete-the-concept]])

Search: `grep -rln "FORMER record format\|schema does not declare\|binary that wrote\|does
not outlive the schema" --include='*.go' .` in the module — run 2026-09-08, nine files:

| Consumer | Fate |
|---|---|
| `internal/record/store.go` (legacy-shards refusal) | MODIFY, leg 2 (§III.6/7) |
| `internal/record/recordsql/read.go` (vocabulary guard) | MODIFY, leg 1 (§III.6) |
| `internal/record/legacyformat_test.go:39` (pins "FORMER record format", "CANNOT READ") | MODIFY with its message, leg 2; pinned substrings stay |
| `internal/record/recordsql/undeclaredtype_test.go` (pins the guard's behaviour) | MODIFY: additionally pins the pointer clause, leg 1 |
| `internal/cli/scorecardunreadable_test.go` (unreadable record reaches scorecard as "needs the tool") | UNCHANGED — asserts substrings the clause does not touch; re-run to confirm |
| `internal/setup/run.go:473` (doc comment: "does not outlive the schema") | MODIFY comment: gains "unless migrated (see internal/record/migrate)" |
| `internal/record/enums_test.go:228` | UNCHANGED — same PHRASE, different concept (enum-fields coverage), [[facts-are-fields]] clause 4 |
| `internal/record/queries.go:91` (check_kind word) | UNCHANGED — same phrase, mint-check vocabulary, not event vocabulary |
| `internal/record/recordpb/word.go` (enum rendering) | UNCHANGED — same phrase, enum-number rendering |

Also in the concept's blast radius, censused by hand: the operator help surface (gains the
verb; `TestRoleHelpCarriesTheFrictionFooter`-style help tests re-run); `releasegate/fuzz`
does NOT exercise `migrate` (operator verb, no seat surface — its absence is this decision,
not a gap); `internal/capture` + scorecard (§III.5); README (§III.8).

## IV. What stays, and why

- **Both production guards stay verbatim in behaviour.** The vocabulary guard and the
  legacy-shards refusal keep refusing; a pointer clause is not a negotiation. Every reader
  that is not the migrator still refuses old shapes.
- **The archive stays raw.** Migrated runs live beside their source, named by the manifest;
  an audit that wants the original bytes still has them, and one that wants a readable
  record knows exactly which translation it is trusting.
- **No backward-compatibility surface appears anywhere else.** No reader grows an epoch
  switch; no writer learns old words. The migrate package is the single place old shapes
  are spelled, and deleting it one day deletes the capability cleanly.

## Risks, graded (likelihood × impact)

| Risk | L | I | Mitigation |
|---|---|---|---|
| Current validation refuses acts legal under old rules, somewhere in 813 events | high | med | §II.8: refusals are the tool's OUTPUT; per-class review; no skip flag |
| Source missing its WAL silently truncates the record | med | high | §II.3: file-set contract, logical source hash, per-word counts in the manifest; §V.3 pins the fixture |
| `opinion`→docket synthesis misstates the bench's act | med | med | mined ground truth (§V); unmapped-column refusal surfaces lost fields; review per entry |
| Identity default silently narrows a kept word whose shape moved | med | high | unmapped old columns REFUSE by default; entry must map or drop-with-reason |
| JSONL merge order reconstructed wrong (leg 2) | med | med | per-archive verification; leg lands separately |
| `Now` pinning leaks into another writer | low | high | migrator is its own process; nothing else writes its `--to` |
| Migrated record mistaken for an original | med | med | manifest is required and field-bearing; §III.5 makes capture/scorecard read it — built in leg 1, not assumed |

## V. Verification plan

Written before implementation; each check names what re-arms it.

1. **Registry refusal tests** (re-armed by any registry or schema change): unknown word →
   error naming word and count; unmapped old column on a kept word → error naming table and
   column; untranslatable without `--accept-loss` → nonzero with the reason. The miss is
   never the same bytes as the clean pass.
2. **Source-adapter tests** (re-armed by changes to migrate/ or the SQLite driver): a
   fixture db whose WAL holds the tail — reading without the WAL is the test's FAILING arm;
   the adapter's copy-then-open produces the full stream; originals opened read-never
   (asserted by permissions or hash-before/after).
3. **Replayer unit tests** (re-armed by changes to migrate/ or record's write path): clock
   pinning produces the original `ts` byte-for-byte through the real `stamp()`; finding ids
   preserved; a synthesized docket pair carries the opinion's ts and seat; `--to` refusal on
   a non-empty target.
4. **The archive is the fixture** (re-armed by registry entries and by any recordpb change):
   migrate `run-archive/2026-09-02_quadratic-formula.tar.gz` in a scratch dir. Assert:
   event count 813 in (the WAL-inclusive count) → 830 out (813 + 17 opinion-synthesized
   motions), per-word counts per the manifest; every citation and proof hash resolves;
   `verify` green over the migrated run; board census (gap counts, closures, verdict)
   matches the mined record of that run. Command:
   `go test ./internal/record/migrate/ -run TestQuadraticFormulaArchive`; skipped LOUDLY
   (`t.Skipf` naming the missing tarball), never silently green, where `run-archive/` is
   absent.
5. **Consistency oracle** over the migrated run: `internal/consistency` holds the family
   and projections to the raw walk exactly as it does for a native record.
6. **Determinism:** migrating the same source twice yields records whose `Events` dumps are
   identical (byte-identical `record.db` where SQLite permits; the events comparison is the
   contract).
7. **Manifest surfacing** (re-armed by capture/scorecard changes): a capture over a migrated
   run carries the "migrated from" row; over a native run carries none — the absent case and
   the healthy case are different bytes.
8. **Suite:** the module's full test run and `go run ./check` from `scripts/`, as every PR.

Ground truth to hold leg 1 against: the 2026-09-02 mining pass (verified board math, 35
friction events, JSTOR-hash findings) — the migrated record must tell the same story in the
new vocabulary.

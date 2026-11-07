# Blue's repair receipt is deleted

> STATUS: **RULED by gblock 2026-11-06; implemented on `feov/delete-manifest-row` at event-schema epoch 65.**

`F` is `plugins/frank-exchange-of-views` and `T` is `F/tools`.

## I. Summary & Goals

**The ruling (gblock, 2026-11-06).** Delete blue's per-gap repair receipt — the `manifest-row` verb, its event and
every reader of it. A repair's record is already the `BlueEdit`: it holds the old and new text and `answers`, the gap
it answers, and the report renders it under "Edits to the report". The receipt added only blue's prose claim that it
checked its own repair, which nothing can refuse, while the lens's re-audit of the gap is the real check. It also cost
blue turns: its work list offered a receipt for every gap it repaired.

**The objective.** The live tree holds no receipt: no verb, no event type, no table, no work-list item, no scorecard
metric, no report section, no envelope field, no constitution or prompt text asking for one. Blue's pre-flight
self-audit — figures recomputed, universals enumerated, sites swept — stays as how blue repairs, with nothing recorded
for it.

## II. What is removed, per carrier

| Carrier | Removed |
|---|---|
| Record | `EVENT_TYPE_MANIFEST_ROW` (15) and the `ManifestRow` body (48), both reserved; its table and decode; its write-path validation and supersede row; `ManifestOwed` / `ManifestUnreceipted` (`BlueSittings` stays: the work list and blue's answered list read it) |
| CLI | the verb, its help page, the `row` payload key, its once-per-sitting answer |
| Work list | blue's item for a repaired gap with no receipt |
| Scorecard | `manifest_coverage`, and with it `Compute`'s read-time parameter, which nothing else read |
| Report | the board's receipt section; `docket.md`'s and the layout's descriptions of it |
| Engine | `BLUE_ENVELOPE.manifest`, the coverage log, the prompt's receipt duty |
| Constitutions | the receipt half of blue's self-audit fragment; the bench's description of the board |
| Probes and fuzz | seatprobe's expectation, difftest arms and scenarios, the release fuzz's drive, envelope and coverage rows |

`feov-memory/blue-scorecard.md` is the measured history of past runs and keeps its rows as recorded.

## III. Migration and residue

`feov-migrate` translates an archived `manifest_row` to no event (`migrate/entries.go`), counted in `events_in` and
not in `events_out`. `TestARepairsReceiptLeavesAnArchivedRun` drives m19 (6 receipts) and the mersenne smoke (22)
through it: no refusal, every receipt dropped, every edit and every edit's `answers` arriving.

**Residue, accepted:** every archived receipt's prose is lost from the migrated record — 7 to 55 per archived run
from 2026-10-09 on (m37 holds 55). The archive tarballs keep it. Each event after a dropped receipt moves earlier in
the migrated record by the receipts before it; the archived-run tests that name event positions (m19, m20, m30, m32,
b6) carry the migrated positions.

## IV. Verification

`go -C scripts run ./check`; the release fuzz on the pull request's CI.

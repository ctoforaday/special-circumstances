# The record flow — event → record → views → seats

The canonical channel between seats is the **record** (the append-only event log, written
only through `feov-record`). Non-prose actions — findings, citations, gaps, closures,
disputes, opinions — are events; the markdown files are *projections* of the record for a
human reading afterward, never the channel. This diagram is kept current with the code
([[fuzzers-and-diagrams-track-code]]); update it in the same PR as any record/protocol change.

## Findings & citations onto the record (#62 pt1, #70/#71)

```mermaid
flowchart TB
  subgraph lenses["red lens seats (L1–L6)"]
    L["lens finding --key F1<br/>lens cite --claim …"]
  end
  subgraph record["THE RECORD (event log — the only inter-agent channel)"]
    FE["finding events<br/>(TOOL-assigned label L{role}-F{N})"]
    CE["cite events"]
    ME["mint / close / dispose events"]
  end
  subgraph views["projections (show --view … / render)"]
    FV["findings (live JSON)"]
    CL["citation-ledger.md"]
    FR["friction (live JSON / friction.md)"]
    BD["board (live JSON)"]
    DBT["debate.md"]
  end
  merge["red-merge seat"]
  score["scorecards.mjs<br/>(--bin → show --view findings)"]
  bench["lead-judge"]

  L -->|emit| FE
  L -->|emit| CE
  FE --> FV
  CE --> CL
  ME --> BD
  FV -->|coalesce, do not transcribe| merge
  merge -->|mint gap, found_by = finding LABELS| ME
  FV -->|per-role/round yield| score
  BD --> bench
  ME --> DBT

  X["red/candidates/*.md<br/>(RETIRED — no longer written or read)"]
  class X retired
  classDef retired stroke-dasharray:5 5,color:#999,stroke:#999
```

## Invariants the diagram encodes

- **Findings are events, not files.** A lens records each finding through `feov-record lens
  finding --key <local F1>`; the tool assigns the run-unique label `L{role}-F{N}` (role from
  the seat id). `red/candidates/*.md` is retired — nothing writes or reads it.
- **The merge reads the findings VIEW**, structured JSON, and coalesces findings into gaps.
  A gap's `found_by` names finding **labels** (`L1-F1`), which `verify.foundByResolves`
  checks against the recorded findings.
- **Two readers of one replay never drift.** `viewjson.go` (the live views) and `render.go`
  (the markdown projections) both derive from `BoardState`; neither parses the other's output.
- **citations_checked** is the board's `counts.citations` (a tally of `cite` events), not a
  self-report (#70/#71 — the citation half of this same migration).

## Out of scope (separate concepts)

`blue/candidates/` (blue best-of-N lane drafts) is unrelated and untouched. `#62 pt2`
(de-editorialising the merge via `supersedes`/tool-side dedup) is a later change.

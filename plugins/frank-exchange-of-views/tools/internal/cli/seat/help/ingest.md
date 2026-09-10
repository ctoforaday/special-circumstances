## menu

freeze the synthesized report into the record and remove the file — done once, by its author

## detail

The one-time act that turns the report from a file into a record. You author the opening report as a markdown file; `ingest` reads it verbatim into the record, PROVES the record reproduces it byte-for-byte, and then removes the file. After it there is no file: the report is the frozen base plus its append-only diff-stack, read with `show report` and changed only through `blue edit`.

WRITE-ONCE. A report is ingested exactly once. A second ingest is refused and points you at `blue edit` — the base is frozen on the record and cannot be overwritten.

AUTHOR-ONLY. Only the seat that wrote the report freezes it. A response or red seat has no ingest.

VERIFY-BEFORE-DELETE. The file is removed only after the record is proven to render back to exactly its bytes. If it is not, the file is KEPT and you are told to STOP and report it as friction — it is a tooling failure, not something an edit can fix, and there is no diff for you to apply.

WHAT MAY BE IN THE REPORT AT ALL: research prose written for a reader of the SUBJECT, and the markers this tool places. Nothing else — no provenance or attribution tags, no notes to another seat, no argument about the run, no narration of how the report was made. A fact that LIMITS THE CONCLUSION stays, re-voiced as a limit on the answer rather than a story about the attempt. Operational facts go to `log`, arguments to `position` or `closing`. This verb and `ingest` are the only two ways authored text enters the report, so the rule is stated on both.

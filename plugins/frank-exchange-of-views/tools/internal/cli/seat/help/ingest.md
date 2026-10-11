## menu

freeze the synthesized report into the record and remove the file — done once, by its author

## detail

Turns the opening report you authored as a markdown file into a record: it reads the file verbatim, PROVES the record reproduces it byte-for-byte, and only then removes it. After that the report is the frozen base plus its append-only diff-stack — read it with `show report`, change it only with `edit`.

WRITE-ONCE. A second ingest is refused and points you at `edit`; the frozen base cannot be overwritten.

AUTHOR-ONLY. Only the seat that wrote the report freezes it; any other seat is refused.

IF THE PROOF FAILS, the file is KEPT and you are told to STOP and report it with `log` as a defect — a tooling failure no edit can fix, with no diff for you to apply.

NO SECTION THE TOOL COMPOSES. Assembly composes these itself, from the record: {{.RefusedHeadings}}. A report that carries one of them as a `## ` heading, or an `**Outcome:**` line above its first heading, is refused and nothing is frozen: the file stands for you to fix, and the one ingest a run gets is unspent. The refusal names the section, what assembly composes it from, and how content reaches it: the act that puts such content on the record, or that no act of yours does and whose does. {{.SubjectHeadingsNote}} `show report` has an assembled form that gives the same for every section before you write.

WHAT MAY BE IN THE REPORT: research prose for a reader of the SUBJECT, and the markers this tool places — nothing else. No provenance or attribution tags, no notes to another seat, no argument about the run, no narration of how the report was made. A fact that LIMITS THE CONCLUSION stays, re-voiced as a limit on the answer rather than a story about the attempt. Operational facts go to `log`, arguments to `closing`, or to `position` where your seat files one.

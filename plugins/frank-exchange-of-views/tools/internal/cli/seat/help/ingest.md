## menu

record your part of the report — once per lane, before the report's first change

## detail

Turns your part of the report, written as a markdown file, into a record: it reads `blue/candidates/<your seat id>.md` in the run directory verbatim, PROVES the record reproduces it, and only then removes the file. The tool builds that path from your seat; you name no file. After it your part is in the report — read the report with `show report`.

THE JOIN. The report is the lanes' parts in the order they are recorded, with every change after them folded on. The first part is the report as written; each part after it follows the report's text, its trailing blank lines trimmed, after one blank line.

A LANE'S, ONCE, BEFORE THE REPORT CHANGES. Only a lane of this run's cast records a part, and a lane records one: a second is refused. No part is recorded once the report holds a change other than a part — a part then would land after text the synthesis has reorganized — and the refusal tells you to say in your `log` what your part holds that the report lacks.

IF THE PROOF FAILS, the file is KEPT and you are told to STOP and report it with `log` as a defect — a tooling failure no edit can fix, with no diff for you to apply.

THE REPORT'S OWN SECTIONS ONLY. A part holds research sections under headings of its own. These are the synthesizer's to write, and a part that carries one is refused with nothing recorded: {{.PartSurfaces}}. The file stands for you to take that line and what stands under it out.

NO SECTION THE TOOL COMPOSES. Assembly composes these itself, from the record: {{.RefusedHeadings}}. A part that carries one of them as a `## ` heading, or an `**Outcome:**` line above its first heading, is refused and nothing is recorded: the file stands for you to fix. The refusal names the section, what assembly composes it from, and how content reaches it: the act that puts such content on the record, or that no act of yours does and whose does. {{.SubjectHeadingsNote}} `show report` has an assembled form that gives the same for every section before you write.

{{.WovenNotesRule}}

WHAT MAY BE IN THE REPORT: research prose for a reader of the SUBJECT, and the markers this tool places — nothing else. No provenance or attribution tags, no notes to another seat, no argument about the run, no narration of how the report was made. A fact that LIMITS THE CONCLUSION stays, re-voiced as a limit on the answer rather than a story about the attempt. Operational facts go to `log`.

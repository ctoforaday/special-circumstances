<!--
  report.md is ASSEMBLED FROM THE RECORD by the bench seat's assemble verb. Nothing is
  authored at assembly. Two ownership classes, marked per section below:

    [BLUE] — authored by blue INSIDE blue/report.md, audited by red every round (red
             re-reads the full report each round), and LIFTED VERBATIM here. A synthesis
             surface authored at assembly would be authored after red's last audit, so it
             lives in the audited document instead. A missing one is flagged, never filled.

    [RECORD] — composed by the tool from the event log + board. The source of truth is the
               records/ event log; the views are rendered from it on read, never materialized.

  This file is the map, not a form to fill in.
-->

# <Topic> — research report            <!-- [BLUE] the H1 title -->

**Verdict:** VERIFIED | UNVERIFIED | CEILING | HALTED   <!-- [RECORD] the terminal `outcome` event -->

## TL;DR                                <!-- [BLUE] 3–6 sentences: the answer, the confidence, the sharpest caveat -->

## The Catechism                        <!-- [BLUE] the agreed answers, the case against at full strength (see catechism_template.md) -->

## Technical foundations                <!-- [BLUE] what is established, with leaf-node citations -->

## Analysis                             <!-- [BLUE] the argument from foundations to conclusions -->

## Risk matrix                          <!-- [RECORD] every open gap's grades, from the board -->

| Risk | Likelihood | Impact | Complexity to mitigate | Mitigation / disposition |
|---|---|---|---|---|
| <risk> | low/med/high | low/med/high | low/med/high | <mitigation, or `defect_accepted` + rationale> |

## The expansions                       <!-- [RECORD] lines of inquiry PURSUED — a research topic followed, and what it yielded -->

## Deferred — for a later run or a deeper context
                                        <!-- [RECORD] lines of inquiry DEFERRED — kept, not rejected: worth taking and not by THIS
                                              run. --reason says what a later run should pick it up FOR. It reaches the
                                              report as a proposal a human selects, never an automatic seed. This is its
                                              OWN section because filing it under "Alternatives considered" said the
                                              opposite of its fate. -->

## Still undecided — proposed and never resolved
                                        <!-- [RECORD] lines of inquiry still PROPOSED — put forward and never resolved. Phrased to
                                              state the absence of a decision rather than imply one. A run that ends with
                                              topics undecided should SAY so: a report with no roads-not-taken is
                                              indistinguishable from one that never looked. -->

## Alternatives considered              <!-- [RECORD] lines of inquiry DECLINED (weighed, not taken) and ABANDONED (tried, died) —
                                              each with its reason (the counter), the history that produced its status,
                                              RED'S RULING on the direction, and, where blue took a line red ruled
                                              out-of-scope or too-thin, the fact that it did so against that ruling. A
                                              ruling is an argument, not a command; the disagreement is the substance.
                                              A line of inquiry still merely PROPOSED has its own section above; the
                                              stale-line duty also asks blue to decide it. -->

## Open questions                       <!-- [BLUE] what the debate could not resolve; a question nobody could answer is a finding -->

## Blue team report (in full)           <!-- [BLUE] blue/report.md embedded verbatim — union, not summary -->

## Red team findings (in full)          <!-- [RECORD] open gaps with grades + the closure index, from the board; then the
                                              lens findings credited by no gap's found_by — red's leaf audit the merge
                                              weighed and did not mint. A finding is addressed by COALESCENCE and nothing
                                              else since #327 retired observe/dispose. Then red's archive spot-checks and
                                              blue's correctness manifest. -->

## The debate                           <!-- [RECORD] ONE transcript: per round the parties' positions, closings and grade
                                              disputes, the PETITIONS — both the filing (class, basis, relief sought) and
                                              its ruling, never the ruling alone — and the bench's opinions; then the
                                              terminal bench disposition (halt / certification), which also states plainly
                                              if any petition went unruled. -->

## Claims withdrawn                     <!-- [RECORD] from the retire events: the claim as it stood, why it went, and what
                                              replaced it. A claim argued and then withdrawn is part of what the debate
                                              decided; omitting it makes this report identical to one where the claim was
                                              never made. Omitted entirely when nothing was retired. -->

<!-- [RECORD] ## Bibliography is COMPOSED AT ASSEMBLY from the cite events — do NOT author it, and do
     not author a "## Footnotes" section either: assembly drops a blue-authored one and weaves every
     invisible <!--cite:--> anchor into the visible [^N] refs plus the bibliography. Cite with the tool. -->

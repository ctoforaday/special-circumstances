<!--
  report.md is ASSEMBLED FROM THE RECORD by `feov-record bench assemble`. Nothing is
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
| <risk> | low/med/high | low/med/high | low/med/high | <mitigation, or `risk_accepted` + rationale> |

## The expansions                       <!-- [RECORD] avenues PURSUED — a concept expansion accepted -->

## Alternatives considered              <!-- [RECORD] every avenue NOT pursued — abandoned, declined, deferred, or still
                                              merely proposed — each with its reason (the counter), the history that
                                              produced its status, RED'S RULING on the direction, and, where blue took a
                                              line red ruled out-of-scope or too-thin, the fact that it did so against
                                              that ruling. A ruling is an argument, not a command; the disagreement is
                                              the substance. -->

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

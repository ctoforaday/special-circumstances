<!--
  THE DELIVERABLE IS A SET OF DOCUMENTS, ASSEMBLED FROM THE RECORD by the bench seat's assemble
  verb. Nothing is authored at assembly. Two ownership classes, marked per section below:

    [BLUE] — authored by blue INSIDE blue/report.md, audited by red every round (red
             re-reads the full report each round), and LIFTED VERBATIM here. A synthesis
             surface authored at assembly would be authored after red's last audit, so it
             lives in the audited document instead. A missing one is flagged, never filled.

    [RECORD] — composed by the tool from the event log + board. The source of truth is the
               records/ event log; the views are rendered from it on read, never materialized.

  WHY A SET. Measured on the archived runs, 70–76% of the single report.md was process record —
  the transcript, the board in full, an operator log LARGER THAN THE ENTIRE RESEARCH
  ARGUMENT — and the research the run was commissioned for was a quarter of its own deliverable.
  Six audiences were unioned into one artifact, so none could be addressed, revised or linked
  without the other five. NOTHING IS DROPPED BY THE SPLIT: every section still ships, in exactly
  one document, and the union is the directory. Empty documents are not written at all.

  Each document opens with the run's title, a link bar over the set, and a rule. report.html
  renders the same set with real tabs and cross-document identity links.

  This file is the map, not a form to fill in.
-->

# README.md — the run's front door                <!-- [RECORD] verdict, rounds, gaps, and what each document holds -->

# report.md — THE RESEARCH

# <Topic>                                <!-- [BLUE] blue's H1, cut at the author's own punctuation boundary -->
**Question:** <the full brief>           <!-- [RECORD] the rest of blue's H1, as a field rather than a heading -->

**Verdict:** VERIFIED | UNVERIFIED | CEILING | HALTED   <!-- [RECORD] the terminal `outcome` event. THE WORD ALONE:
                                              a field a reader can skim, badge or grep is one token. Its argument —
                                              what the verdict means, the basis it rests on, the bench's words on a
                                              deadlock — opens "Read this first" instead. -->

## How this run was conducted            <!-- [RECORD] what ANSWERED each seat, measured; never the configuration -->

## Read this first                       <!-- [RECORD] the verdict's argument, then the bench's TERMINAL ask (one, never
                                              one per certify event — the superseded ones are in CHANGELOG.md), then the
                                              open gaps ranked most-severe first -->

## TL;DR                                 <!-- [BLUE] 3–6 sentences: the answer, the confidence, the sharpest caveat -->

## The Catechism                         <!-- [BLUE] the agreed answers, the case against at full strength (see catechism_template.md) -->

## Technical foundations                 <!-- [BLUE] what is established, with leaf-node citations -->

## Analysis                              <!-- [BLUE] the argument from foundations to conclusions -->

## Risk matrix                           <!-- [RECORD] every open gap's grades, from the board -->

| Risk | Likelihood | Impact | Complexity to mitigate | Mitigation / disposition |
|---|---|---|---|---|
| <risk> | low/med/high | low/med/high | low/med/high | <mitigation, or `defect_accepted` + rationale> |

## Research areas                        <!-- [RECORD] lines of inquiry PURSUED — a research topic followed, and what it yielded -->

## Future research directions            <!-- [RECORD] lines of inquiry DEFERRED — kept, not rejected: worth taking and not by THIS
                                              run. --reason says what a later run should pick it up FOR. It reaches the
                                              report as a proposal a human selects, never an automatic seed. This is its
                                              OWN section because filing it under "Alternatives considered" said the
                                              opposite of its fate. -->

## Alternatives considered               <!-- [RECORD] lines of inquiry DECLINED (weighed, not taken) and ABANDONED (tried, died) —
                                              each with its reason (the counter), the history that produced its status,
                                              RED'S RULING on the direction, and, where blue took a line red ruled
                                              out-of-scope or too-thin, the fact that it did so against that ruling. A
                                              ruling is an argument, not a command; the disagreement is the substance. -->

## Open questions                        <!-- [BLUE] what the debate could not resolve; a question nobody could answer is a finding -->

## Blue team report (sections not composed above)
                                         <!-- [BLUE] whatever blue authored that is genuinely ADDITIONAL — union, not summary.
                                              Its lifted surfaces and any tool-owned section it wrongly authored are dropped,
                                              and the section is omitted entirely when nothing survives. -->

## Bibliography                          <!-- [RECORD] COMPOSED AT ASSEMBLY from the cite events — do NOT author it, and do
                                              not author a "## Footnotes" section either: assembly drops a blue-authored one
                                              and weaves every invisible <!--cite:--> anchor into the visible [^N] refs.
                                              Cite with the tool. PER DOCUMENT: a footnote definition cannot cross a file
                                              boundary, so each document defines the references it carries. -->

# docket.md — the board

## The board                             <!-- [RECORD] open gaps with grades + the closure index; then the lens findings
                                              credited by no gap's found_by — red's leaf audit the merge weighed and did
                                              not mint. A finding is addressed by COALESCENCE and nothing else since #327
                                              retired observe/dispose. Then red's archive spot-checks and blue's
                                              correctness manifest. -->

# debate.md — the adversarial record

## The debate                            <!-- [RECORD] ONE transcript: per round the parties' positions, closings and grade
                                              disputes, and the bench's opinions; then the terminal bench disposition
                                              (halt / certification), which also states plainly if any petition went unruled. -->

# judgments.md — the judicial record

## Motions                               <!-- [RECORD] every adjudicated exchange, joined on its id: the FILING (class, basis,
                                              relief sought) and its ruling, never the ruling alone. -->

# evidence.md — the computations

## Proofs                                <!-- [RECORD] every proof on the record, with its script, output, sha256 and red's
                                              independent re-run. Numbered RUN-WIDE, so P3 is the same computation in every
                                              document that cites it. A proof no document references is still shown, and
                                              said to be unreferenced. -->

# run.md — how the machinery behaved

## Log (what the run told the operator)   <!-- [RECORD] the log entries, each with what it asserts; nominal entries render apart -->

## Record verification                   <!-- [RECORD] the record's own invariant check — a section, never a gate -->

## Cost                                  <!-- [RECORD] folded in by capture, which is the only stage with transcript access -->

# CHANGELOG.md — this report's own provenance

## Report revision history               <!-- [RECORD] every recorded edit to blue's report, in round order -->

## Claims withdrawn                      <!-- [RECORD] from the retire events: the claim as it stood, why it went, and what
                                              replaced it. A claim argued and then withdrawn is part of what the debate
                                              decided; omitting it makes this report identical to one where the claim was
                                              never made. Omitted entirely when nothing was retired. -->

## Superseded bench statements           <!-- [RECORD] the certifications the terminal one replaced. A bench that certifies
                                              twice has CHANGED its statement, not made two asks. -->

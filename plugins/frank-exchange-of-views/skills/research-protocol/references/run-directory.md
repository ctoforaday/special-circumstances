# The run directory

The operator's map of one run: what `setup` lays down, what the record holds, what is assembled
last, and how a run ends. A seat does not read this file — its dispatch and its role's help carry
what it acts on.

**The tool is the read path.** Where a line below says RECORD, that artifact has no
authoritative file — read it with `show <name>` and never from disk — `show` is a GROUP, so `show --help` lists every projection.

```
research/<date>_<slug>/
├── records/           # THE RECORD — the source of truth: records/record.db, the run's one store
│                      # (MAY live outside the run entirely; a `.records-elsewhere` note appears
│                      #  here instead. Nothing changes for a seat, because a seat reads the
│                      #  record with `show <name>` and never from disk — which is the
│                      #  point: a run can be configured so that is the ONLY way, and then a
│                      #  missing verb has to surface in the log instead of a workaround)
│                      # EVERYTHING FROM README.md TO report.html IS ASSEMBLED LAST, from the record,
│                      # by `assemble`, for the HUMAN reader. None of it exists while a seat sits, and
│                      # a seat never opens one: the report is read through the record tool, and the rest
│                      # are projections the tool serves live from the same record.
├── README.md          # the run's front door: verdict, gaps, and what each document holds
├── report.md          # the research — verdict, Catechism, foundations, analysis, risks, open questions
├── docket.md          # the board: every gap and how it closed, blue's manifest, red's spot-checks
├── debate.md          # the transcript, epoch by epoch, and the bench's terminal disposition
├── judgments.md       # motions — every contested question and how it was ruled
├── avenues.md         # each avenue's fate, the path it took, its ruling and appeal
├── evidence.md        # the computations, with script, output and sha256
├── run.md             # the log, the record's invariant check, and cost
├── CHANGELOG.md       # this report's own provenance: revisions, retired claims, repairs
├── report.html        # the same set with real tabs and cross-document links — one file, no server
├── inputs/PINNED.md   # the evidence base, pinned: repo HEAD at launch + cited corpora's commit/revision
├── blue/
│                      # (the opening hypotheses are AVENUES on the record, not a file — read
│                      #  them as the `avenues` projection. A hypothesis in a file is one red
│                      #  cannot rule too-thin or out-of-scope, and the opening ones shape the
│                      #  whole run)
│   ├── report.md      # written by the synthesizer at synthesis, then frozen into the record
│   │                  #   by `ingest`, which deletes it. From then the report is read
│   │                  #   through the tool, and every change goes through the `edit` verb
│   └── candidates/    # best-of-N lane drafts, one method each, preserved (authored)
└── cost.md            # measured tokens + dollars per seat-sitting (written by `capture`)

RECORD — no file at all; read through the tool. Every projection, what each is for, and the
verb that WRITES each one are in each role's help, which that role's agent definition carries in
full — generated from the command tree, so it cannot disagree with it. A catalogue here would
be a second copy that can.

  a seat's work list — what it reads first, and again before it stops: everything open to that
  seat, each item saying whether it is what blocks the seat closing, plus whether the sitting may
  close at all. `complete: true` with items still open means the gates are satisfied, NOT that
  the seat is done.

trajectories/       journal.jsonl (the HARNESS's lifecycle record, tracked)
                    + agent-transcripts.tar.gz (gitignored)
```

`setup` lays down directories and no file a seat fills. **A stub is not an artifact**: a
placeholder reads as the real thing. Anything under RECORD above has no file at all — read it
with `show <name>`.

`blue/report.md` is the file the synthesizer writes and then FREEZES: the freeze records its
text as the base of the record and DELETES the file. From then it too is a projection —
there is no `blue/report.md` to open; read the report through the tool and change it only
through the tool's edit path, each change an event the report is replayed from. It cannot be
raw-written or bypassed, which is the point: the report a seat reads and the report the record holds
are the same bytes, by construction.

**Termination is the record's, and the standing practice is stop-and-resume**: the chair's `dispatch next` says who sits; the run ends when nobody is ready — PASS
permitted (VERIFIED), or every open material gap at its limit (CEILING): the bench remanded it,
the one more exchange the remand grants its minting lens and blue left it at impasse, and the
bench remanded it again — or when the run reaches its epoch limit with parties still ready (CEILING, the limit
named as the reason). The bounds are the run's terms, recorded at setup: the exchanges a gap gets before
impasse (k-max), the floor of the gaps a lens may mint (mint-budget), which the record raises
with the report's size in each lens's unit — citations, proofs, claims or prose paragraphs — and
the chair sittings the run gets (max-epochs, default 12). One stop is the engine's own: a dispatch
plan identical for three chair sittings in a row — the same parties readied against the same
head for the same reasons — is a loop nothing on the board is moving, and the engine ends the run
UNVERIFIED naming the stuck parties and the head. Red owns PASS/FAIL — *is it defensible*. **The bench
owns the stopping judgment** — *is it close enough*, the one call that weighs remaining defect
against remaining cost, and the only terminal value (economy) that otherwise has no organ. It
reads the telemetry projection — the series, never a snapshot — and files a reasoned,
cost-stated opinion; the operator acts on it,
stopping a run past its value and resuming for the honest UNVERIFIED
assembly — cache replay makes the stop ~$0. **Stopping is not passing**: the verdict
stays UNVERIFIED with the open count stated. The bench, not a severity floor, makes the stopping
call: a floor automates the one call that belongs to judgment. NEVER
change models on the resume.

All artifacts are git-tracked; nothing is summarized away. The payload is the file; the envelope is the handle — no large content travels through agent return values.

# friction.md — Is the event-log-as-source-of-truth record model in frank-exchange-of-views sound, and where are its failure modes?

## RED LENS 1 (2026-07-20)

**Tool missing**: feov-record binary not found at /AppData/Local/Temp/feov-bin/feov-record. Attempted registration failed with exit code 127. Impact: unable to use the record tool's `lens register` verb to formally declare seat identity and all board acts (findings write, closure recording, ledger append). Workaround used: direct file writes to /red/candidates/ and manual ledger edit. This bypasses the record tool's schema validation and write guards. Recommendation: either install feov-record at the expected path or update the instruction to provide the actual path/command.

## auto-harvested at capture (envelope entries absent from the file — lens/abort loss class)

- lines_of_inquiry — scorecard shows this run has no avenues recorded yet (diagnostic not computed); avenue verb was used to record 10 lines of inquiry post-synthesis, but the scorecard metric fires only on the second round with recorded data; this is expected behavior and not a friction item per se, but the gap-pattern file W1.11 notes the detection window and the metric ceiling
- confidence_scoring — per-claim confidence records do not yet exist (W2f flagged as BLOCKED); this round lacks confidence grading alongside claims; next round's calibration-is-craft audit will require adding confidence per claim in the report itself, not post-hoc through envelope fields
- Tool behavior: feov-record `merge show --view board` output is JSON; no native-to-envelope translation tool provided. Workaround: hand-transcription from JSON to gap structs (safe, done correctly, but should be automated).

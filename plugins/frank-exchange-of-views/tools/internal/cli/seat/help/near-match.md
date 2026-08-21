## menu

screen a candidate against the board BEFORE minting, so a reopen does not arrive as a fresh gap

## detail

screen a candidate against the board before minting: --problem "<what is wrong>" [--quote "<the sentence it lives at>"] — returns the top {{.NearMatchTopN}} gaps (open AND closed) by lexical overlap, so a near-duplicate surfaces as a reopen (mint --supersedes <id>) rather than a fresh gap. The tool SCREENS and ranks; you decide reopen-or-new. Read-only: it records nothing.

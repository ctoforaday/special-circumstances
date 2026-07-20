# Red verifies verbatim — WebFetch is prohibited for citation verification

Ruling 2026-07-19 (during smoke run b). **APPLY TO debate.js ONLY AFTER run b lands** —
editing the script mid-run busts the workflow's replay cache for any agent whose prompt
changed, so it cannot be a faithful resume.

## The finding

WebFetch's own contract: *"Fetches a URL, converts the page to markdown, and answers
`prompt` against it using a small fast model."* It returns a SMALL MODEL's summary of the
source, never the source. Leaf-node citation verification — the red-citation-lens's whole
job — cannot be done against a summary: it drops exact digits, quoted sentences, and which
experimental arm carries a result. Grading a citation against a WebFetch answer grades the
summarizer, not the leaf. This is `pattern_verification_probe_layer_masking` — the verdict
produced by a different layer than the one under test.

## The rule (stronger than "prefer verbatim")

- **Blue MAY use WebFetch** for breadth. **Red MAY NOT use it to verify.**
- Red reads sources VERBATIM with Bash and reads them ITSELF: `curl -sL <url>`,
  `gh issue view <n> --comments`, `pdftotext`/pandoc for PDFs.
- If a source is too large for red's context, red spawns its OWN verification sub-agent —
  a full-capability agent it instructs to check the specific claim and return the finding
  WITH quoted evidence. That agent is a real verifier red chose and briefed. **WebFetch's
  small model is not that entity** — the distinction is not "delegation is bad" but "the
  delegate must be a verifier, not a summarizer."
- A truncated read is not a read; state truncation; never grade a body you could not fully
  read.

## Where to apply

`ledgerClause` (~line 573 in debate.js) — replace the "LARGE SOURCES (W1.12)" sentence
(currently "WebFetch is lossy on threads, use gh for those") with the hard prohibition
above. Consider also strengthening RED_LENSES[0] ("leaf-node citation verification") to name
the verbatim requirement.

Bundle with the merge-seat migration edits (also uncommitted) when the loop proves out.

## Measured mandate (2026-07-18 keeper run transcripts)

Not hypothetical. Fetch tools by seat role in the keeper run:

| seat | WebFetch (summary) | curl+gh (verbatim) | verbatim % |
|---|---|---|---|
| red-citation-lens | 45 | 28 | **38%** |
| red logic/darkside | 8 | 0 | 0% |
| red-merge | 0 | 4 | 100% |
| bench | 15 | 3 | 17% |
| **total** | **68** | **35** | **33%** |

The citation lens — whose whole job is leaf verification — did **62% of its reads through a
summarizer**. The bench did **15** WebFetch reads on its probe/staleness re-checks. So the
prohibition applies to **red AND the bench** (every verifying seat), not the citation lens
alone. (Trajectory tarball is a round-2+ subset, so blue's building is under-sampled — but
blue using summaries is permitted; the hole is that red, the verbatim backstop, wasn't one.)

Not every WebFetch grade is wrong — the small model did read the page — but every EXACT
claim (digit, quote, experimental arm) verified this way is unreliable, and that is most of
what citation verification checks.

## Follow-up: the verifier sub-agent needs the Task tool

The "red spawns its own full-model verifier to protect its context" idea (a good one) is
NOT available yet: red-auditor's tool allowlist has no Task/Agent tool, so a seat cannot
spawn a sub-agent. Enabling it means (a) adding Task to red-auditor, and (b) verifying a
WORKFLOW-spawned seat can nest-spawn an agent at all — unconfirmed. Until then the prompt
says: read large sources in sections (curl ranges / pdftotext) and name what you read.

## Follow-up: red needs a VERBATIM JS renderer (smoke-c finding)

Removing WebFetch killed summaries AND removed the only tool that renders JavaScript. `curl`
returns the SPA shell for JS-rendered pages — and `platform.claude.com` (the Claude docs, a
PRIMARY source for the telemetry topic) is JS-rendered. In smoke-c red hit this on 4 sources,
did the RIGHT thing (graded them LOW confidence, filed a capability-gap friction, did not
fake a verified grade), and named the fix itself.

The fix is NOT WebFetch back — it is a verbatim JS-render path:
- Add a browser tool to ALL THREE allowlists that lost WebFetch — red-auditor,
  blue-researcher, AND lead-judge. Blue is not exempt: it BUILDS claims from these
  sources (platform.claude.com is JS-rendered and primary for this topic), so without a
  JS renderer blue's own sourcing is crippled the same way red's verification is.
  `claude-in-chrome` `read_page` /
  `get_page_text` renders the DOM and returns the actual page text (not a small-model
  summary), which is verbatim-for-JS. Verify it returns raw text, not a processed answer.
- VERIFY reachability: MCP servers may be absent in headless/workflow seats, and
  claude-in-chrome needs the extension + a running browser. If unreachable in a workflow
  seat, red's honest "verification impossible for JS-rendered sources" friction is the
  correct floor — flagged, never faked.

The friction channel worked exactly as designed: silent hearsay became a visible, honest,
actionable capability gap. That is the system succeeding, not failing.

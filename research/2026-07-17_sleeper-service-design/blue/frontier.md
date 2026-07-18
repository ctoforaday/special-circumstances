# blue frontier — How should sleeper-service, the autonomous learning loop (Phase 4), be designed?

Formulated before searching, per research protocol. Each states what would be TRUE if the
candidate answer were right, so searches test rather than wander. Evidence base pinned at
`7bc501e` (see `inputs/PINNED.md`); port-plan §3c read at `AgentOrange/docs/claude-port-plan.md`
(the pin's `plans/claude-port-plan.md` path does not exist in the special-circumstances tree —
see friction).

## H1 — Artifact-mining beats introspection (what the loop CONSUMES)

If the right input design is a pipeline over durable run artifacts — `friction.md` harvests,
`cost.md`, `trajectories/board-telemetry.jsonl`, run records, red-auditor gap-pattern memory,
`ideas/backlog.md` + `doubts.md` — then: (a) the pinned corpus already demonstrates sufficiency
(the friction harvests read as pre-ranked improvement backlogs: PDF extraction ranked #1 across
two runs by three seats; write-guard and Read-cap classes recur with counts attached); (b)
external self-improvement literature (Reflexion-class loops, postmortem/telemetry mining in
AIOps/SRE practice) shows loops fed by recorded execution traces outperform loops that
introspect prompts/rules without execution evidence; (c) no proposal-quality signal exists that
is NOT already captured in a durable artifact — if we find critical signal that evaporates
(session-local state, un-persisted log()), the input design must add capture, not clever recall.
Falsifier: friction/telemetry proves too noisy or too lagging to rank proposals, forcing a
curation stage that reintroduces judgment cost.

## H2 — /self-improve is a thin driver over FEOV, not a second research engine

If /self-improve's right shape is enumerate → pick ONE → delegate "how should X evolve?" to the
existing /research machinery → emit an idea stub with alternatives, then: (a) the picker is
mechanics (cheap tier, deterministic scoring over harvested signal: recurrence count × severity ×
staleness), and the research is judgment (existing FEOV, unchanged); (b) a full debate run
($80–120 at list rates, measured run 3) is unaffordable daily, so the default delegation target
must be a bounded mode (smoke-scale lanes/rounds or a single-researcher pass), with full-debate
promotion reserved for /graduate; (c) the idea stub's contract (alternatives enumerated,
provenance to the harvested signal, acceptance shape stated) is what makes graduation auditable
later. Falsifier: bounded research modes produce stubs too shallow to survive red at graduation
time — then the loop should run LESS often at full strength rather than daily at smoke strength.

## H3 — Headless viability is a harness-facts question, not a design preference

If daily unattended `claude -p "/self-improve"` is viable, then the harness must verifiably
support, non-interactively: plugin slash-command invocation; subagent/workflow spawning; MCP
availability headless (stdio servers, and the qmd HTTP daemon — verified live 2026-07-14:
`qmd mcp --http --daemon`, PID file, `/health`, MCP Streamable HTTP at :8181/mcp — as the
recall path that avoids per-invocation model loads); permission pre-configuration
(allowed-tools / permission modes) without an interactive consent prompt; and machine-readable
exit status for the scheduler. Each is checkable against primary Claude Code documentation; any
one failing forces a fallback (scheduled cloud agents, or a "prepared docket awaiting an
interactive session" design where the schedule only stages work). Falsifier: headless mode
cannot load plugin commands or MCP servers — then scheduling docs must lead with the
semi-attended recipe, not cron.

## H4 — Consent gates must be structural, or autonomy erodes them

If the semantic-consent doctrine is to survive an autonomous loop, the write boundary
(research/ + ideas/ only; NEVER rules/skills/commands/hooks/settings/plugin manifests — and
NEVER the loop's own guardrail definitions) must be enforced by mechanism outside the model's
reasoning: permission deny-rules, path-scoped hooks (the same class as sc-quality-gate, already
verified to fire on subagent writes), and a promotion step that structurally requires a human
action (pull-request review / explicit command) rather than prompt text saying "please don't."
Then: (a) unattended-agent safety practice (least privilege, proposer/approver separation,
defense against the agent editing its own constraints) corroborates structural over
instructional gates; (b) the corpus's own evidence — prompt-level MUSTs violated live (the lane
that skipped the PDF MUST-try clause; batching prose ignored at haiku, the script-vs-prose
doctrine) — shows instructions alone degrade under load; (c) every gate should carry a stated
answer to "how would the loop, optimizing its objective, route around this?" Falsifier:
hook/permission machinery cannot express the deny-set headless (e.g., settings not honored
under `-p`), forcing the gate up a level (sandbox/worktree with restricted checkout, or
cloud-agent isolation).

## H5 — Cost discipline: ceilings + honest partials; cheapen mechanics, never judgment

If unattended spend is disciplinable under the efficiency doctrine, then: (a) every unattended
run carries hard ceilings (turn caps, round caps, model pinned to a bulk tier for enumeration
and mechanics; judgment seats inherit stronger models only in gated or graduation runs); (b)
stop conditions must yield a durable honest partial — run 4's death at the monthly spend limit
proved the pattern: null-guard abort, resumable cached state, honest UNVERIFIED assembly beats
losing paid-for work — so every loop stage writes to the blackboard as it goes and an aborted
nightly run leaves a resumable stub, not nothing; (c) a per-run budget plus a standing monthly
cap for the loop is checked BEFORE launch (quota guard), with cost telemetry (cost.md class)
feeding back into the loop's own consumption. Falsifier: the harness offers no pre-launch quota
introspection — then the guard degrades to conservative static ceilings plus post-hoc audit,
and the design must say so honestly.

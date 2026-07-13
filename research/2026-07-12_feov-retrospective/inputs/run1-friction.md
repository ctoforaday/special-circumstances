# Run 1 friction reports (verbatim, from the workflow envelope)

- red-merge-r1: Workflow script passed a literal 'undefined' run directory (uninitialized variable in the caller) — recurrence recorded in agent memory 'workflow-undefined-rundir'. Caller should validate the run-dir argument before dispatching red.
- blue-respond-r1: Uninitialized invocation: run-directory and topic arrived as literal 'undefined'. Fix: caller-side variable binding plus a dispatch-time guard.
- blue-respond-r1: Write to undefined/blue/report.md was blocked by the environment ('Subagents should return findings as text, not write report files'), but the debate protocol requires a LIVING blue/report.md on the blackboard. I returned the content in the envelope instead; the lead must materialize it or exempt blackboard artifacts from the restriction.
- red-merge-r2: literal "undefined" run-directory paths for the third consecutive dispatch; requested caller-side preflight guard.
- red-merge-r2: No round-2 red candidate lens passes were ever produced, so the merge had no inputs by construction.
- judge-r2: Task instructed reading undefined/red/findings.md — file does not exist; adjudicated from debate.md and direct filesystem verification instead.
- judge-r2: The requested preflight guard is still absent; I cannot patch the caller's workflow script from the judge role.
- judge-r2: Final assembly by union impossible on this deadlock: no blue report and no red findings were ever produced.
- judge-r2: Blue's round-1 report write was blocked ("subagents return findings as text"), so the blue envelope is the only carrier of its full report.

# Known outcomes (verified by the operator pair after the run)
- Plugin PostToolUse hooks FIRED on workflow-agent writes (hook log evidence) — the quality gate holds across the swarm.
- red-auditor's `memory: project` WORKED — it wrote agent-memory/frank-exchange-of-views-red-auditor/workflow_undefined_rundir.md.
- The engine never fabricated: FAIL rounds, honest deadlock, UNVERIFIED verdict. The gate held under garbage input.
- Root cause was the caller: args arrived JSON-stringified on a resume, destructured to undefined. A parse+guard has since been added to the debate script.
- Cost of the null run: 16 agents, 252.9k tokens, 11m48s.

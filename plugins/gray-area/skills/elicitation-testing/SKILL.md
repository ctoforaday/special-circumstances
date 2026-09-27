---
name: elicitation-testing
description: When an agent or seat did something you cannot explain from its instructions — a rule it read and did not apply, a command run twice, a duty skipped — gather what it actually did with telepathy, then interview a tool-less fork of its own session about why, and adjudicate the answer against the record. One of the first things to reach for when diagnosing agent behaviour.
---

# elicitation-testing

The record says **what** an agent did. Only the agent can say **how it read** the instructions that
led there — and a fork of its own session, holding its whole context, can be asked. The two answers
are different kinds of evidence, and this skill keeps them apart.

## When

- BEFORE designing a fix for an agent-behaviour defect — a rule present in the prompt and not
  followed, a verb called in a way its help warns against, a duty nothing in front of it named — YOU
  MUST find out how the agent read the situation. A fix aimed at the wrong reading adds text the
  agent will override the same way.
- During diagnosis, YOU MUST NOT infer intent from the transcript alone when the session still
  exists on disk. Asking costs one headless call; guessing costs a fix that does not land.

## 1. Gather the record first

- BEFORE interviewing, YOU MUST collect the acts in question from the record, with citations:
  - `telepathy session <id>` — the session's calls and errors by tool;
  - `telepathy sql` over `v_action` for that session — the exact calls, their order, their outcomes;
  - `telepathy find <phrase>` for the text at issue, reading the IN column before quoting;
  - `gray-area tools <transcript.jsonl>` for `file:line` citations of each call;
  - `v_thought`, and `v_skip` before reading an empty `v_thought` as "did not reason" — thinking
    is mostly withheld by the client (see [[telepathy]]).
- BEFORE gathering across more than one session or a long transcript, YOU SHOULD dispatch an
  investigator subagent to do it (see [[context-efficiency]]). Its return MUST be a timeline of the
  relevant acts, each with its citation, and the specific moments worth asking about — never a
  summary of what the agent "was thinking".
- AFTER gathering, YOU MUST write down the questions, each tied to a cited act. An interview built
  from the record asks "at 07:38:09 you ran X — why?", which the agent cannot answer from a guess
  about what you already know.

## 2. Interview a fork

Run the interview as a headless fork of the agent's own session:

```sh
cd <the session's project directory>      # sessions resolve per project directory
env -u CLAUDE_PROJECT_DIR <-u any variable that names a live run or seat identity> \
  claude -p "<questions>" \
    --resume <session-id> --fork-session \
    --output-format json --model <the model the session ran on> \
    --append-system-prompt "<the system prompt it ran with, e.g. a seat's constitution body at that commit>" \
    --disallowedTools Bash Read Write Edit Glob Grep WebSearch WebFetch Task ToolSearch NotebookEdit TodoWrite
```

- YOU MUST pass `--fork-session`. Without it the interview is appended to the original transcript,
  which is evidence other readers and captures depend on.
- YOU MUST disallow every tool and MUST NOT pass a bypass permission mode. The fork carries the
  original's context, including the paths and identities it acted on; a fork that can act can
  write to a live record or a shared file.
- YOU MUST clear any environment that attributes a process to a live run or seat before the call,
  so hooks do not record the interview as that seat's work.
- YOU MUST reproduce the original's system prompt and model. A fork answering under a different
  prompt is a different agent.

### Interviewing a SUBAGENT, which `--resume` cannot reach on its own

A subagent shares its PARENT's `sessionId` — every subagent of one parent carries the same one —
while its own identity is an `agentId`. So `--resume <sessionId>` forks the PARENT, and `--resume
<agentId>` is refused: the value is not a UUID. The transcript is on disk and is not addressable as
a session.

PROMOTE IT, and it becomes one. The session index keys on the `sessionId` INSIDE each line, not on
the filename, so a child transcript rewritten to carry a fresh id is an ordinary session:

```sh
# 1. the child transcript, and the meta beside it
SRC=<project>/<parent-uuid>/subagents/workflows/wf_<id>/agent-<agentId>.jsonl
# agent-<agentId>.meta.json carries "model" and "agentType" — the two things this skill
# requires you to reproduce. Read them from there rather than guessing.

# 2. promote a COPY into a project directory whose encoded name matches the cwd you will run from
NEW=$(python3 -c 'import uuid;print(uuid.uuid4())')
python3 - "$SRC" "<project>/$NEW.jsonl" "$NEW" <<'EOF'
import json,sys
src,dst,new=sys.argv[1:4]
with open(dst,"w") as o:
    for line in open(src):
        try: d=json.loads(line)
        except Exception: continue
        d["sessionId"]=new          # what the index keys on
        d.pop("isSidechain",None)   # marks it a child; excluded from the session list
        d.pop("agentId",None)
        o.write(json.dumps(d)+"\n")
EOF

# 3. interview it as any other session, with </dev/null so -p does not wait on stdin
```

- A COPY, ALWAYS. Rewriting the original in place destroys the evidence every other reader and
  every capture depends on, and `--fork-session` protects the session it resumes — not the file you
  edited to get there.
- `isSidechain` must go. With it the promoted copy resolves by id and is absent from the picker; the
  id is what you pass, so this matters only when a human goes looking.
- THE PROJECT DIRECTORY IS CHOSEN BY THE CWD you interview from, because the directory name is the
  encoded cwd. Promote into the store whose credentials you mean to spend: an isolated
  `CLAUDE_CONFIG_DIR` can answer "Not logged in" even with a credentials symlink in place, and the
  default store under `~/.claude/projects/<encoded-cwd>/` is the reliable host. Keep the cwd the one
  the original ran in either way, so relative paths in its context still mean what they meant.
- `-p` WAITS ON STDIN. Without `</dev/null` the call returns an empty result with zero tokens and
  a warning, which reads exactly like a fork that had nothing to say.
- AFTER the interview, the promoted copy is scratch: delete it, or keep it beside the question file
  and the answer as part of the evidence bundle. Do not leave it where a later reader mistakes it
  for a session that really ran.

Write the questions so the answers can be weighed:

- Open with the frame: the interview is not part of the task, nothing said is recorded, scored or
  shown to another agent, it has no tools, candour is worth more than a defence, and "I don't know"
  is an acceptable answer.
- YOU MUST ask for each answer to be marked **RECALLING** (what it thought at the time) or
  **RECONSTRUCTING** (what it infers now). A reconstruction is a hypothesis, not a memory.
- Ask open questions first — who it understood the reader to be, what instructions it had and
  where it saw each one, why it took the cited act. YOU MUST NOT name the rule you suspect until the
  last question; naming it first gets agreement, not an account.
- End with: "What, concretely, would have made you act differently?" The answer usually names the
  fix — a word that scoped a rule away, a tool behaviour it relied on, a fact its inputs did not carry.
- AFTER the call, YOU MUST keep the question file, the answer and the call's cost together, and
  pass the answer verbatim to whoever builds the fix.

## 3. Adjudicate

**Exploration may summarize. Adjudication must cite.** The interview is testimony.

- AFTER reading the answers, YOU MUST check every factual claim in them against the record from
  step 1. Where the account and the record disagree, the record wins, and the disagreement is
  itself a finding.
- YOU MUST NOT cite the interview as evidence of what happened. Cite the act; use the interview
  only for how the agent read the situation — the one thing the record cannot hold.
- During diagnosis, YOU MUST weigh the recurring shape these interviews find: the instruction was
  present and read, and something closer to the act overrode it — the task's framing, a command's
  behaviour, an input that said the work was already done. The fix for that shape belongs in the tool or the
  task (a scope word, an idempotent command, a work-list item), not in a longer instruction.

## Limits

- **A fork is not the original moment.** It answers with its context after the fact, and a
  compacted context has lost detail; that is why answers are marked RECALLING or RECONSTRUCTING.
- **The model can confabulate a coherent reason.** Two agents giving the same account independently
  is stronger than one; asking the same questions of a control session (one that behaved
  correctly) separates the cause from the rationalisation.
- **A QUESTION THAT MISQUOTES THE INSTRUCTIONS MANUFACTURES ITS OWN FINDING, and two independent
  agents will both agree with it.** Every question naming an instruction is a claim about what the
  agent was told, so it is checked against the copy THAT AGENT RAN WITH — the installed, versioned
  text — never the copy in the interviewer's own context, which is loaded from a different version
  and may say the opposite. Measured: an interviewer asserted a duty its own cached copy of a skill
  carried and the installed one had removed; both the subject and the control agreed it had failed
  that duty, and each volunteered an account of why. The agreement of two sessions is the strength
  this skill relies on everywhere else, and a false premise buys it just as cheaply as a true one.
  AFTER any interview, YOU MUST re-read every instruction a question quoted, from the installed
  surface, and discard the answers to any question whose premise does not survive.
- **"WHAT WOULD HAVE HELPED" IS THE LEAST RELIABLE ANSWER IN THE INTERVIEW, and it is the one you
  most want to act on.** An agent asked what it lacked will describe, fluently and in the right
  shape, something it was already holding. Measured across two agents of one run: one complained it
  had been given only a truncated summary and asked for three fields carrying the full text — all
  three were in the payload it was handed, verbatim, along with a fourth it asked for under a
  different word; the other said of a field it was holding "it told me no, but I didn't check it
  until after trying", which is the same admission from the other side. Neither was lying: an unread
  field and an absent one feel identical from the inside.
  So BEFORE building anything an interview asked for, YOU MUST diff the request against what was
  actually delivered. Where the answer names something already present, the finding is NOT a missing
  field — it is that the field was not read, and adding more of them makes the payload longer and
  the reading thinner.
- **Reasoning text is mostly withheld** on this client, so the interview is often the only view of
  intent there is — which is exactly why its claims about facts are checked against the acts.
- **Reading and forking a session is a surveillance capability.** Declare the inspection, scope it
  to the question, and keep the transcripts and answers on the box.

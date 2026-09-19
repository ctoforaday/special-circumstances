// Package machinereader owns one published fact about a session: WHO READS ITS FINAL MESSAGE.
//
// Every hook in this suite writes for an agent that is talking to a person. A session launched
// by a headless launcher may not be: its final message can be an envelope a caller parses —
// "your FINAL message must be exactly ONE JSON object and nothing else" — and such a session is
// still a MAIN session, so `Stop` fires in it exactly as it does in a dev session at a terminal.
//
// MEASURED, on the smoke run of 2026-09-17 (#1025): four of 35 sittings answered the Stop hook's
// checkpoint nudge in prose instead of sending their envelope, and cost 34.6 of 147.4
// sitting-minutes — 23% of the run's wall clock — in corrective turns. Two of those seats,
// interviewed tool-less and independently, each said the injection read as a fresh human-directed
// instruction that superseded the return-format contract.
//
// NOTHING IN THE PAYLOAD OR THE ENVIRONMENT ANSWERS THE QUESTION, which is why the answer is
// declared rather than detected. `Stop` carries session_id, transcript_path, cwd, prompt_id,
// permission_mode, hook_event_name, stop_hook_active, last_assistant_message, background_tasks and
// session_crons — no agent id, no mode, no entrypoint. The client's own undocumented
// CLAUDE_CODE_SESSION_ATTENDED is 1 for a terminal and 0 for `claude -p`, and ALSO 0 for a Remote
// Control session, which is a human-read session and the primary dev surface here: gating on it
// would silence the nudge exactly where it is most wanted. `scratchpad_dir` draws the same wrong
// line.
//
// SO IT IS COOPERATIVE, AND THAT IS THE COST. A launcher that does not set the variable gets
// today's behaviour, injection and all. The suite cannot detect this condition; it can only
// publish a way to declare it, and the dependency then points from the launcher to
// prosthetic-conscience and never the other way.
package machinereader

// Var is the published spelling, and this constant is its only one in the tree. It is set by the
// LAUNCHER on the session's process — never by this plugin, and never by the agent inside the
// session, which cannot know what its caller will do with what it sends.
const Var = "SC_FINAL_MESSAGE_CONTRACTED"

// Contracted reports whether the variable asserts a machine reader.
//
// PRESENCE, not truthiness: any non-empty value asserts it, "0" included. A launcher that sets it
// to a computed flag must set it to the EMPTY string to mean "no", because a variable whose value
// is parsed has a third outcome — a misspelled value — and that outcome would be silently
// indistinguishable from an honest "no". Presence has two states and both are visible in `env`.
func Contracted(v string) bool { return v != "" }

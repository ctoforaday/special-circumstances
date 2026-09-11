package telecli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A CORPUS THE TESTS OWN, not a copy of the developer's own transcripts.
//
// Every golden below is produced from these bytes, which means a golden that changes is either a
// deliberate change to the output or a regression — never yesterday's real sessions drifting. The
// corpus deliberately covers the awkward cases rather than the tidy one:
//
//   - all three transcript tiers, including a seat inside a WORKFLOW (the tier that left 56 files
//     unattributed when it was missed),
//   - a tool call that FAILED and one whose result never arrived (outcome `unresolved`),
//   - all THREE reasoning states, which must read differently: text stored, a block that arrived
//     EMPTY (the client withholding it — 15,503 such blocks on the measured corpus against 248
//     stored), and no thinking block at all,
//   - a line that is not JSON, which must be counted rather than skipped in silence,
//   - EVERY SPEAKER a user-role record can have — the human with and without an origin, a peer turn
//     that is also isMeta, a task notification, harness text (isMeta and a compaction summary), a
//     seat prompt with no origin — and every mid-turn delivery shape: a peer's, the human's, a
//     notification known only by its commandMode, and a coordinator's inside a seat's file,
//   - a term the human said FIRST and an agent repeated later, which a filter on the most recent
//     hit alone drops.

// frozen is the clock every golden is rendered against. All fixture timestamps are relative to it,
// so `2h ago` means the same thing in 2026 as it will in 2030.
var frozen = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

func at(d time.Duration) string { return frozen.Add(-d).UTC().Format(time.RFC3339) }

// corpus is one fixture transcript file: where it goes, and what is in it.
type corpusFile struct {
	rel   string // relative to the projects root
	lines []string
}

func userLine(uuid, promptID, sid, cwd, text string, ago time.Duration) string {
	return jsonLine(map[string]any{
		"uuid": uuid, "promptId": promptID, "sessionId": sid, "cwd": cwd,
		"type": "user", "timestamp": at(ago),
		"message": map[string]any{"role": "user", "content": []any{
			map[string]any{"type": "text", "text": text},
		}},
	})
}

func assistantLine(uuid, parent, sid, cwd string, ago time.Duration, blocks ...map[string]any) string {
	bs := make([]any, len(blocks))
	for i, b := range blocks {
		bs[i] = b
	}
	return jsonLine(map[string]any{
		"uuid": uuid, "parentUuid": parent, "sessionId": sid, "cwd": cwd,
		"type": "assistant", "timestamp": at(ago),
		"message": map[string]any{"role": "assistant", "content": bs},
	})
}

func resultLine(uuid, parent, sid string, ago time.Duration, toolUseID string, isError *bool) string {
	blk := map[string]any{"type": "tool_result", "tool_use_id": toolUseID}
	if isError != nil {
		blk["is_error"] = *isError
	}
	return jsonLine(map[string]any{
		"uuid": uuid, "parentUuid": parent, "sessionId": sid,
		"type": "user", "timestamp": at(ago),
		"message": map[string]any{"role": "user", "content": []any{blk}},
	})
}

func toolUse(id, name string, input map[string]any) map[string]any {
	return map[string]any{"type": "tool_use", "id": id, "name": name, "input": input}
}

// turnLine is a user-role record carrying the fields SpeakerOf reads. content may be a string —
// the shape peer turns and notifications arrive in — or a block list.
func turnLine(uuid, sid, cwd string, ago time.Duration, extra map[string]any, content any) string {
	rec := map[string]any{
		"uuid": uuid, "sessionId": sid, "cwd": cwd, "type": "user", "timestamp": at(ago),
		"message": map[string]any{"role": "user", "content": content},
	}
	for k, v := range extra {
		rec[k] = v
	}
	return jsonLine(rec)
}

// queuedLine is a MID-TURN DELIVERY: an attachment record with no message. kind "" and mode ""
// are left out of the record, as the client leaves them out.
func queuedLine(uuid, sid, cwd string, ago time.Duration, kind, mode, prompt string) string {
	a := map[string]any{"type": "queued_command", "prompt": prompt}
	if kind != "" {
		a["origin"] = map[string]any{"kind": kind}
	}
	if mode != "" {
		a["commandMode"] = mode
	}
	return jsonLine(map[string]any{
		"uuid": uuid, "sessionId": sid, "cwd": cwd, "type": "attachment", "timestamp": at(ago),
		"attachment": a,
	})
}

func origin(kind string) map[string]any {
	return map[string]any{"origin": map[string]any{"kind": kind}}
}

func jsonLine(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func ptrBool(b bool) *bool { return &b }

// sessionAlpha is the rich one: two tiers of seat, a failure, an unresolved call, and reasoning.
const (
	alphaID     = "aaaaaaaa-1111-4111-8111-111111111111"
	betaID      = "bbbbbbbb-2222-4222-8222-222222222222"
	alphaCWD    = "/work/alpha"
	betaCWD     = "/work/beta"
	projectSlug = "-work-alpha"
	// gamma holds a speaker of every kind; delta is the human-first, agent-later transcript.
	gammaID  = "cccccccc-3333-4333-8333-333333333333"
	deltaID  = "dddddddd-4444-4444-8444-444444444444"
	gammaCWD = "/work/gamma"
	deltaCWD = "/work/delta"
)

// nightlyBuild is one notification, delivered as a TURN at top level and MID-TURN inside a seat.
const nightlyBuild = "<task-notification>\n<status>completed</status>\n<summary>Background command \"nightly build\" completed</summary>\n</task-notification>"

// harnessReminder is text the client injects, isMeta and with no origin, at both tiers.
const harnessReminder = "<system-reminder>The task tools have not been used recently.</system-reminder>"

func corpus() []corpusFile {
	return []corpusFile{
		{
			rel: filepath.Join(projectSlug, alphaID+".jsonl"),
			lines: []string{
				userLine("u1", "p1", alphaID, alphaCWD, "widen the gap view", 4*time.Hour),
				assistantLine("a1", "u1", alphaID, alphaCWD, 3*time.Hour+50*time.Minute,
					map[string]any{"type": "thinking", "thinking": "four sites carry this field, not one"},
					map[string]any{"type": "text", "text": "Widening the four carriers."},
					toolUse("t1", "Read", map[string]any{"file_path": "/work/alpha/mint.go"}),
				),
				resultLine("r1", "a1", alphaID, 3*time.Hour+49*time.Minute, "t1", nil),
				assistantLine("a2", "u1", alphaID, alphaCWD, 3*time.Hour+40*time.Minute,
					toolUse("t2", "Edit", map[string]any{"file_path": "/work/alpha/mint.go"}),
					toolUse("t3", "Bash", map[string]any{"command": "go test ./..."}),
					// AN EDIT MADE THROUGH THE SHELL. Its target is the whole command, so the
					// path sits in the MIDDLE of the string — invisible to a suffix match, and
					// the reason `touched` matches on a substring instead.
					toolUse("t6", "Bash", map[string]any{"command": "sed -i s/a/b/ /work/alpha/view.go && gofmt -w ."}),
				),
				resultLine("r2", "a2", alphaID, 3*time.Hour+39*time.Minute, "t2", ptrBool(false)),
				resultLine("r3", "a2", alphaID, 3*time.Hour+38*time.Minute, "t3", ptrBool(true)),
				resultLine("r6", "a2", alphaID, 3*time.Hour+37*time.Minute, "t6", ptrBool(false)),
				// A call whose result never arrived: the session was interrupted. This must read
				// as `unresolved`, which is neither `ok` nor `error`.
				assistantLine("a3", "u1", alphaID, alphaCWD, 3*time.Hour+30*time.Minute,
					toolUse("t4", "Write", map[string]any{"file_path": "/work/alpha/view.go"}),
				),
				// REASONING WITHHELD. A signature and no text is what the client emits, and
				// dropping it silently made this indistinguishable from an agent that did not
				// think. Two of them, so the count in the output cannot be confused with a
				// boolean.
				assistantLine("a4", "u1", alphaID, alphaCWD, 3*time.Hour+20*time.Minute,
					map[string]any{"type": "thinking", "thinking": "", "signature": "abc123"},
				),
				assistantLine("a5", "u1", alphaID, alphaCWD, 3*time.Hour+10*time.Minute,
					map[string]any{"type": "thinking", "thinking": "   ", "signature": "def456"},
				),
				// Not JSON. Counted as unparsed, never silently dropped.
				`{"uuid":"torn","sessionId":`,
			},
		},
		{
			rel: filepath.Join(projectSlug, alphaID, "subagents", "agent-explore-01.jsonl"),
			lines: []string{
				userLine("su1", "p9", alphaID, alphaCWD, "find every caller", 3*time.Hour),
				assistantLine("sa1", "su1", alphaID, alphaCWD, 2*time.Hour+55*time.Minute,
					map[string]any{"type": "text", "text": "Three callers."},
					toolUse("st1", "Grep", map[string]any{"pattern": "Mint("}),
				),
				resultLine("sr1", "sa1", alphaID, 2*time.Hour+54*time.Minute, "st1", nil),
			},
		},
		{
			// THE THIRD TIER. A rule naming only <sid>.jsonl and <sid>/subagents/ misses this
			// path entirely and reports a clean, complete-looking corpus without it.
			rel: filepath.Join(projectSlug, alphaID, "subagents", "workflows", "wf-7", "agent-review-02.jsonl"),
			lines: []string{
				userLine("wu1", "p9", alphaID, alphaCWD, "review the diff", 2*time.Hour),
				assistantLine("wa1", "wu1", alphaID, alphaCWD, time.Hour+55*time.Minute,
					toolUse("wt1", "Read", map[string]any{"file_path": "/work/alpha/mint.go"}),
				),
				resultLine("wr1", "wa1", alphaID, time.Hour+54*time.Minute, "wt1", nil),
			},
		},
		{
			// NO THINKING BLOCK AT ALL, which is the third state and not the same as the two
			// empty ones above: nothing was withheld here, the transcript simply has none.
			rel: filepath.Join("-work-beta", betaID+".jsonl"),
			lines: []string{
				userLine("bu1", "q1", betaID, betaCWD, "bump the pin", 30*time.Minute),
				assistantLine("ba1", "bu1", betaID, betaCWD, 29*time.Minute,
					map[string]any{"type": "text", "text": "Pinned."},
					toolUse("bt1", "Edit", map[string]any{"file_path": "/work/beta/go.mod"}),
				),
				resultLine("br1", "ba1", betaID, 28*time.Minute, "bt1", nil),
			},
		},
		{
			// EVERY SPEAKER AT TOP LEVEL. The peer turn is isMeta as well, as every measured one is:
			// a classifier that read isMeta before origin would file it as harness text.
			rel: filepath.Join("-work-gamma", gammaID+".jsonl"),
			lines: []string{
				turnLine("gu1", gammaID, gammaCWD, 100*time.Minute, origin("human"), "start the audit sweep"),
				assistantLine("ga1", "gu1", gammaID, gammaCWD, 99*time.Minute,
					map[string]any{"type": "text", "text": "Starting the sweep."}),
				turnLine("gpeer", gammaID, gammaCWD, 95*time.Minute,
					map[string]any{"isMeta": true, "origin": map[string]any{"kind": "peer", "name": "session-x"}},
					"Another Claude session sent a message:\n<cross-session-message from=\"session-x\">is the conduct section settled?</cross-session-message>"),
				turnLine("gnote", gammaID, gammaCWD, 90*time.Minute, origin("task-notification"), nightlyBuild),
				turnLine("gmeta", gammaID, gammaCWD, 85*time.Minute, map[string]any{"isMeta": true},
					[]any{map[string]any{"type": "text", "text": harnessReminder}}),
				turnLine("gcompact", gammaID, gammaCWD, 80*time.Minute, map[string]any{"isCompactSummary": true},
					"This session is being continued from a previous conversation. The task tools were idle."),
			},
		},
		{
			// A SEAT'S FILE: its prompt carries no origin and is the lead's, and a coordinator and a
			// notification arrive mid-turn — the notification known by its commandMode alone.
			rel: filepath.Join("-work-gamma", gammaID, "subagents", "agent-seat-03.jsonl"),
			lines: []string{
				turnLine("gsu1", gammaID, gammaCWD, 70*time.Minute, nil, "Audit the ledger migration for drift."),
				assistantLine("gsa1", "gsu1", gammaID, gammaCWD, 69*time.Minute,
					map[string]any{"type": "text", "text": "No drift found."}),
				turnLine("gsmeta", gammaID, gammaCWD, 67*time.Minute, map[string]any{"isMeta": true},
					[]any{map[string]any{"type": "text", "text": harnessReminder}}),
				queuedLine("gscoord", gammaID, gammaCWD, 65*time.Minute, "coordinator", "",
					"The coordinator sent a message: list every caller, then stop"),
				queuedLine("gsnote", gammaID, gammaCWD, 60*time.Minute, "", "task-notification", nightlyBuild),
			},
		},
		{
			// THE HUMAN FIRST, AN AGENT LATER. `--in user` on "ledger column" must keep this row and
			// show the human's words, though the newest hit is the assistant's.
			rel: filepath.Join("-work-delta", deltaID+".jsonl"),
			lines: []string{
				turnLine("du1", deltaID, deltaCWD, 50*time.Minute, origin("human"), "rename the ledger column to amount_cents"),
				assistantLine("da1", "du1", deltaID, deltaCWD, 45*time.Minute,
					map[string]any{"type": "text", "text": "Renaming the ledger column now."}),
				queuedLine("dpeer", deltaID, deltaCWD, 40*time.Minute, "peer", "prompt",
					"<cross-session-message from=\"session-y\">the conduct section needs one more pass</cross-session-message>"),
				queuedLine("dhuman", deltaID, deltaCWD, 35*time.Minute, "human", "prompt", "and keep it nullable"),
			},
		},
	}
}

// writeCorpus materialises the fixture and returns the projects root.
func writeCorpus(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range corpus() {
		p := filepath.Join(root, f.rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(strings.Join(f.lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// sessionFile writes one of the client's liveness advertisements.
//
// pidDomain is deliberately foreign, which pins liveness at `unknown` on every platform and in
// every container this suite runs in. A fixture claiming the LOCAL domain would report `live` or
// `ended` depending on whether some pid happened to exist — a golden that passes on the author's
// box and fails in CI, and worse, one whose failure would look like a liveness bug.
func sessionFile(t *testing.T, dir string, pid int, sid, cwd string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := jsonLine(map[string]any{
		"pid": pid, "sessionId": sid, "cwd": cwd,
		"procStart": "0", "pidDomain": "linux:fixture-machine:pid:0",
	})
	if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d.json", pid)), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

package record

import (
	"fmt"
	"os"
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WHERE A SEAT'S TRANSCRIPT IS, read off the record.
//
// The SubagentStop hook is the one party TOLD where a seat's conversation lives — its payload's
// agent_transcript_path — and sittingwrite stores it on the agent's sitting_close. Every reader that
// opens a seat's transcript asks here, so none of them searches a configuration directory for a file
// named after the agent, or recovers an agent id from a file's name. The harness writes two layouts
// (`subagents/` for a plain subagent, `subagents/workflows/wf_<id>/` for a Workflow seat); a search
// that guessed one found no Workflow seat at all for two runs (#1438), and its miss read exactly
// like a seat that never ran.
//
// THREE MISSES, AND EACH SAYS WHICH. The record names no transcript for an agent whose sitting has
// not stopped — it is still sitting, or it ran as a `claude -p` main session, which fires no
// SubagentStop; it names none where the stop carried no path; and a path it names may be gone,
// because run-archive/ keeps the record and not the transcripts, so a path is readable only on the
// box the run ran on. None of the three is a seat that took no turns, and none is answered from
// anywhere else.

// Transcript is one agent's transcript as its stop on the record names it.
type Transcript struct {
	AgentID string
	Path    string
}

// Missing is nil when the file the record names is readable here, and otherwise says why not.
func (t Transcript) Missing() error {
	st, err := os.Stat(t.Path)
	if err == nil && !st.IsDir() {
		return nil
	}
	why := "it is a directory"
	if err != nil {
		why = err.Error()
	}
	return fmt.Errorf("the record names %s as agent %s's transcript and it is not readable here (%s): transcripts are not archived with the record, so one is readable only on the box its run ran on", t.Path, t.AgentID, why)
}

// TranscriptCensus is every agent the record saw sit, by what the record can say of its transcript.
type TranscriptCensus struct {
	// Named is each agent whose stop names its transcript, in agent-id order. The path is the
	// latest stop's.
	Named []Transcript
	// Unnamed is each agent whose stop carries no path.
	Unnamed []string
	// Sitting is each agent that opened a sitting — a register or the hook's bracket — and has no
	// stop on the record.
	Sitting []string
}

// TranscriptsOf reads the census off the family's events.
func TranscriptsOf(f Family) TranscriptCensus {
	path := map[string]string{}
	stopped := map[string]bool{}
	opened := map[string]bool{}
	for _, e := range f.Events {
		if a := recordpb.AgentOpening(e); a != "" {
			opened[a] = true
		}
		if b, ok := recordpb.BodyAs[*recordpb.SittingClose](e); ok && b.GetAgentId() != "" {
			stopped[b.GetAgentId()] = true
			if p := b.GetAgentTranscriptPath(); p != "" {
				path[b.GetAgentId()] = p
			}
		}
	}
	var c TranscriptCensus
	for a := range stopped {
		if p, ok := path[a]; ok {
			c.Named = append(c.Named, Transcript{AgentID: a, Path: p})
		} else {
			c.Unnamed = append(c.Unnamed, a)
		}
	}
	for a := range opened {
		if !stopped[a] {
			c.Sitting = append(c.Sitting, a)
		}
	}
	sort.Slice(c.Named, func(i, j int) bool { return c.Named[i].AgentID < c.Named[j].AgentID })
	sort.Strings(c.Unnamed)
	sort.Strings(c.Sitting)
	return c
}

// TranscriptOf is the path of agentID's transcript, readable here, or the reason the record cannot
// give one.
func TranscriptOf(f Family, agentID string) (string, error) {
	stopped, path := false, ""
	for _, e := range f.Events {
		if b, ok := recordpb.BodyAs[*recordpb.SittingClose](e); ok && b.GetAgentId() == agentID {
			stopped = true
			if p := b.GetAgentTranscriptPath(); p != "" {
				path = p
			}
		}
	}
	switch {
	case !stopped:
		return "", fmt.Errorf("agent %s has no sitting_close on the record: its sitting has not stopped (it is still sitting, or it ran as a main session, which fires no SubagentStop), so the record names no transcript for it yet", agentID)
	case path == "":
		return "", fmt.Errorf("agent %s's sitting_close names no transcript path, so the record cannot say where its transcript is", agentID)
	}
	t := Transcript{AgentID: agentID, Path: path}
	if err := t.Missing(); err != nil {
		return "", err
	}
	return path, nil
}

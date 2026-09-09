package catalogue

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Liveness is the three-valued answer. `Unknown` is a real answer, not a failure: a board that
// renders an unmeasurable agent as Ended is the confident-wrong output this plugin exists to
// refuse, and the same shape as `capture_category`'s populations and feov's three-state
// TextExtracted.
type Liveness string

const (
	Live    Liveness = "live"
	Ended   Liveness = "ended"
	Unknown Liveness = "unknown"
)

// SessionFile is what the client writes for every live session, at
// ~/.claude/sessions/<pid>.json. We read it; we never write it.
type SessionFile struct {
	PID       int    `json:"pid"`
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	StartedAt int64  `json:"startedAt"`
	ProcStart string `json:"procStart"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	// PidDomain is the vendor's own guard that a pid only means something inside one namespace on
	// one machine: `linux:<machine-id>:<pid namespace>`. Comparing a pid against local /proc
	// without checking it would report a session from a container, or another host sharing $HOME,
	// as Ended — which is exactly the confident-wrong answer Unknown exists to refuse.
	PidDomain string `json:"pidDomain"`
}

// ReadSessionFiles returns every session the client currently advertises.
func ReadSessionFiles(dir string) ([]SessionFile, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	var out []SessionFile
	for _, p := range paths {
		body, err := os.ReadFile(p)
		if err != nil {
			continue // a session file racing us out of existence is not an error
		}
		var sf SessionFile
		if json.Unmarshal(body, &sf) != nil || sf.SessionID == "" {
			continue
		}
		out = append(out, sf)
	}
	return out, nil
}

// Of resolves one session's liveness.
//
// Live requires all three: the file exists, the pid exists, and its start time matches. A pid
// alone is unsafe because pids are reused, and a recycled one would report a dead agent as alive
// — a wrong answer indistinguishable from a right one. `procStart` makes the pair a durable
// process identity.
func Of(sf SessionFile, localDomain string, start func(int) (string, bool)) Liveness {
	// A foreign pidDomain means the pid is not ours to interpret. Not Ended: unmeasurable.
	if sf.PidDomain != "" && localDomain != "" && sf.PidDomain != localDomain {
		return Unknown
	}
	if localDomain == "" {
		return Unknown // no way to establish our own domain: say so
	}
	got, ok := start(sf.PID)
	if !ok {
		return Ended
	}
	if got != sf.ProcStart {
		return Ended // pid reused by an unrelated process
	}
	return Live
}

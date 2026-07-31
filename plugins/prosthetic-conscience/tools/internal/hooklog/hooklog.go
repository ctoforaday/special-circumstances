// Package hooklog is the suite's one hook-firing log, and the one place that knows how to
// write it.
//
// WHY IT EXISTS. sc-quality-gate and sc-recall-index carried byte-identical copies of this
// function, and both append to the SAME file on the same PostToolUse event. Measured
// (hook-surface-spike.md §4b): the client runs an event's hooks in PARALLEL, so those two
// writes were CONCURRENT BY DESIGN — not a possible interleaving, a certain one.
//
// The real fix is one writer, which the merged PostToolUse binary provides: units RETURN
// their log line and the binary appends them in order. This package is where that append
// lives, and the standalone entry points share it rather than keeping a copy each.
//
// Instrumentation is best-effort by design: a log that cannot be written must never cost
// the tool call, so every error here is deliberately swallowed. An empty projectDir
// disables it rather than writing a relative .claude/ into whatever directory the process
// happened to start in — the defect internal/hookenv exists to close.
package hooklog

import (
	"os"
	"path/filepath"
)

// Name is the log's filename under .claude/.
const Name = "prosthetic-conscience-hook.log"

// Path is where the log lives for a project.
func Path(projectDir string) string { return filepath.Join(projectDir, ".claude", Name) }

// Append records one or more hook firings. Each line must carry its own trailing newline.
// Passing every line from one call is what keeps a single writer single.
func Append(projectDir string, lines ...string) {
	if projectDir == "" {
		return
	}
	body := ""
	for _, l := range lines {
		body += l
	}
	if body == "" {
		return
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(Path(projectDir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(body)
	_ = f.Close()
}

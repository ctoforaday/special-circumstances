// Package trajectory reads Claude Code and Antigravity session transcripts into citable events.
//
// Contract (Design by Contract):
//
//	Every event MUST carry its provenance — file, 1-based line, and uuid — because
//	a finding cites the trajectory, never the reader's opinion of it.
//	A line that does not parse MUST be counted and reported, never silently
//	skipped: a reader that quietly drops input answers confidently from a subset.
package trajectory

import (
	"io"

	sharedtraj "github.com/ctoforaday/special-circumstances/shared/trajectory"
)

// Event is one citable act. Provenance is not optional: File, Line and UUID are
// what a consumer quotes when a finding rests on this event.
type Event = sharedtraj.Event

// Result is the parse outcome across a whole transcript.
type Result = sharedtraj.Result

// Parse reads one transcript into citable acts (tool_use and tool_result).
func Parse(r io.Reader, file string) (Result, error) {
	return sharedtraj.Parse(r, file)
}

// ParseAll reads one transcript into all citable events including user and assistant text.
func ParseAll(r io.Reader, file string) ([]Event, error) {
	return sharedtraj.ParseAll(r, file)
}

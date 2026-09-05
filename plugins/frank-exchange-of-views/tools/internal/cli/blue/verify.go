package blue

import (
	"fmt"
	"strings"
)

// VerifyReproduction is the gate that stands before the file is deleted (issue #709): the record
// must render back to EXACTLY the bytes ingested, or ingest aborts and the file is kept. It
// returns nil on a match and, on a mismatch, an error whose MESSAGE IS WRITTEN FOR A MODEL THAT
// ACTS ON IT.
//
// # Why the wording is load-bearing
//
// The reader of this error is blue, and blue applies text blindly — hand it a diff and it will
// try to splice the diff. But a reproduction mismatch is a TOOLING failure (render or ingest lost
// a byte), NOT a defect in blue's content, and there is nothing for blue to edit: editing the
// report cannot fix a render bug, it can only mask it and churn the very text under audit. So the
// message must:
//
//   - lead with the DIRECTIVE, before any text, so a model that reads only the top still stops;
//   - state plainly that the FILE IS PRESERVED — nothing was deleted, no work is lost;
//   - forbid the repair it would otherwise reach for, and say why it cannot help;
//   - show the divergence as READ-ONLY DIAGNOSTIC for the human maintainer, never as an edit
//     target — no "change X to Y", no quoted span positioned as an instruction.
func VerifyReproduction(ingested, rendered string) error {
	if ingested == rendered {
		return nil
	}
	at := firstDivergence(ingested, rendered)
	return fmt.Errorf(`INGEST ABORTED — STOP. Do not edit the report; do not retry; the tool failed, not your writing.

Your report file is PRESERVED, byte-for-byte, exactly as you wrote it. Nothing was deleted and no work is lost. This is a defect in the ingest/render tooling — the record did not reproduce your report — and it is not something an edit can fix: changing the report cannot correct a rendering bug, it can only hide it. There is no diff for you to apply here.

WHAT TO DO: report this as friction (a tooling failure in the ingest step) and stop. A maintainer fixes the tool; you change nothing.

The divergence below is DIAGNOSTIC FOR THE MAINTAINER — it is NOT an instruction and NOT a target to edit toward. The two versions first differ at byte %d:

  ── as you wrote it (authoritative) ──
  %s
  ── as the record rendered it (the tool's faulty output) ──
  %s`,
		at, contextWindow(ingested, at), contextWindow(rendered, at))
}

// firstDivergence returns the byte offset of the first difference between a and b (or the length
// of the shorter, if one is a prefix of the other).
func firstDivergence(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

// contextWindow returns a short, single-line-safe slice of s around byte offset at, for the
// maintainer to eyeball. It is deliberately small and marked — enough to locate the drift, not
// enough to read as a passage worth splicing.
func contextWindow(s string, at int) string {
	const pad = 60
	lo := at - pad
	if lo < 0 {
		lo = 0
	}
	hi := at + pad
	if hi > len(s) {
		hi = len(s)
	}
	seg := s[lo:hi]
	// collapse newlines so the window stays one visual unit and cannot be mistaken for a
	// multi-line edit body.
	seg = strings.ReplaceAll(seg, "\n", "⏎")
	prefix, suffix := "", ""
	if lo > 0 {
		prefix = "…"
	}
	if hi < len(s) {
		suffix = "…"
	}
	return prefix + seg + suffix
}

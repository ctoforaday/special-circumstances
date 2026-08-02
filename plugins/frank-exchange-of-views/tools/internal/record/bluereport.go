package record

import (
	crand "crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

// MutateBlueReport is the ONE way a feov-record process may mutate blue/report.md
// (the finding-marker layer). It is the exported, ERROR-PROPAGATING counterpart to
// writeAtomic: unlike a projection, a marker write is NOT re-derivable from the event
// log, so a lost write must surface as an error, never be swallowed.
//
// It holds the "blue-report" flock across the whole read-modify-write and lands the
// new bytes via durableWrite (temp + fsync + rename), so a concurrent reader never
// sees a torn file and a crash mid-write leaves the prior file intact. The flock
// serializes the CONCURRENT lens-seat inserts (L1-L6 file findings in parallel);
// blue's own Edit-tool writes are not feov-record processes and are guarded instead
// by turn-boundary sequencing (findings land in red-lens, blue edits in blue-respond).
//
// transform receives the current bytes (nil if the file is absent) and returns the
// new bytes. To REJECT (write nothing — e.g. a finding whose quote is not in the
// report), transform returns a non-nil error; MutateBlueReport propagates it and does
// not write. On success it returns the write error, or nil.
func MutateBlueReport(runDir string, transform func(old []byte) ([]byte, error)) error {
	path := filepath.Join(runDir, "blue", "report.md")
	var retErr error
	withLock(runDir, "blue-report", func() {
		old, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			retErr = err
			return
		}
		next, err := transform(old)
		if err != nil {
			retErr = err // a REJECT (or a transform failure): write nothing, surface it.
			return
		}
		b := make([]byte, 3)
		if _, err := crand.Read(b); err != nil {
			retErr = err
			return
		}
		tmp := path + ".tmp-" + hex.EncodeToString(b)
		retErr = durableWrite(path, next, tmp)
	})
	return retErr
}

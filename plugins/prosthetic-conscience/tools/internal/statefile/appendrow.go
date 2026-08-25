package statefile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppendRow adds one JSON object as one line to a JSONL record, creating the file and its
// directory if needed.
//
// It is here rather than in a package of its own because it is the same subject: a small
// record several processes write. It is deliberately NOT the same function as Write — that
// one REPLACES a state atomically and last-writer-wins is acceptable; this one ACCUMULATES,
// and a lost row is a lost measurement that nothing can reconstruct.
//
// # What this does and does not promise about concurrency
//
// O_APPEND makes each write land at the current end of file, so two writers cannot
// overwrite each other's rows. On POSIX a single write under PIPE_BUF (4096 bytes) is also
// atomic, so rows do not interleave; seal rows are ~200 bytes, well inside it. That is a
// platform property, not a guarantee this function makes, and a caller writing rows near
// 4 KB from parallel hooks needs a lock rather than this.
//
// It returns the error instead of choosing what to do about it, because the two callers
// legitimately differ: postcompactobserve reports the failure to stderr, and the sealer
// reports it with its own binary's name in front. Neither may fail the hook.
func AppendRow[T any](path string, v T) error {
	line, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

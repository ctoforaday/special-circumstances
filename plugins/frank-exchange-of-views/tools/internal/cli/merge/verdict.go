package merge

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// verdict: the merge seat's terminal act.
//
// It CHECKPOINTS the whole records/ directory (the append-only event log) to a
// mirror outside the run. The run directory is untracked-by-design until
// capture, and the 2026-07-17 incident showed how a stray git operation can
// delete a live blackboard mid-round; a mirror keyed by the run path means the
// events survive the working tree. Projections are regenerated on read from the
// mirror, so the frozen snapshot is the source, not a materialized cache.
func newVerdict() *cobra.Command {
	c := seat.New("verdict", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		v, ok := record.VerdictOf(seat.Str(cmd, flags.As))
		if !ok {
			// THE REFUSAL NAMES WHAT WOULD HAVE WORKED. `%q is not a verdict` tells a seat it
			// was wrong and not what to type — and this is the terminal act of the round, so
			// a seat that cannot get past it has nowhere to go. The set is rendered from the
			// same declaration the help renders, so the two cannot say different things.
			return nil, feov.Errorf(feov.Validation,
				"merge verdict: --%s must be one of %s (got %q) — the verdict is the round's terminal act and every later reader switches on it",
				flags.As, record.MustEnum("verdict", "verdict").Spelling(), seat.Str(cmd, flags.As))
		}
		if _, err := record.Append(s.Identity(), &recordpb.RoundVerdict{Verdict: &v}); err != nil {
			return nil, err
		}
		open, closed, err := view.Counts(run)
		if err != nil {
			return nil, err
		}
		mirror, err := checkpoint(run)
		if err != nil {
			return nil, err
		}
		return verdictResult{Verdict: seat.Str(cmd, flags.As), Open: open, Closed: closed, Checkpoint: mirror}, nil
	})

	enumhelp.Flag(c, flags.As, record.MustEnum("verdict", "verdict"), ("the seat's terminal act"))
	return c
}

// checkpoint copies records/ to a per-run directory under the user cache, keyed
// by a hash of the run dir so two runs never collide.
func checkpoint(run record.Run) (string, error) {
	sum := sha1.Sum([]byte(run.Dir()))
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	mirror := filepath.Join(home, ".cache", "feov", "run-mirror", hex.EncodeToString(sum[:])[:12])
	if err := os.MkdirAll(mirror, 0o755); err != nil {
		return "", err
	}
	src := run.Records()
	if _, err := os.Stat(src); err == nil {
		if err := copyDir(src, mirror); err != nil {
			return "", err
		}
	}
	return mirror, nil
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		s, d := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(s, d); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

type verdictResult struct {
	Verdict    string `json:"verdict"`
	Open       int    `json:"open"`
	Closed     int    `json:"closed"`
	Checkpoint string `json:"checkpoint"`
}

func (r verdictResult) Human() string {
	return fmt.Sprintf("verdict %s (%d open, %d closed) and checkpointed to %s", r.Verdict, r.Open, r.Closed, r.Checkpoint)
}

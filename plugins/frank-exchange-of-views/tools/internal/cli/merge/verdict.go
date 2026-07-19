package merge

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// verdict: the merge seat's terminal act.
//
// It renders and then CHECKPOINTS the whole records/ directory to a mirror
// outside the run. The run directory is untracked-by-design until capture, and
// the 2026-07-17 incident showed how a stray git operation can delete a live
// blackboard mid-round; a mirror keyed by the run path means the events survive
// the working tree.
func newVerdict() *cobra.Command {
	c := seat.New(role, "verdict",
		"the seat's terminal act: --as PASS|FAIL — renders all projections and checkpoints records/ to the recovery mirror",
		func(s seat.Context, cmd *cobra.Command) (string, error) {
			p := seat.Set(cmd, record.NewPayload(), "verdict", flags.As)
			if _, err := record.Append(s.RunDir, s.SeatID, "verdict", p); err != nil {
				return "", err
			}
			r, err := record.Render(s.RunDir, "")
			if err != nil {
				return "", err
			}
			mirror, err := checkpoint(s.RunDir)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("verdict %s — rendered (%d open, %d closed) and checkpointed to %s",
				seat.Str(cmd, flags.As), r.Open, r.Closed, mirror), nil
		})

	c.Flags().String(flags.As, "", "PASS | FAIL — the seat's terminal act")
	return c
}

// checkpoint copies records/ to a per-run directory under the user cache, keyed
// by a hash of the run dir so two runs never collide.
func checkpoint(runDir string) (string, error) {
	sum := sha1.Sum([]byte(runDir))
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	mirror := filepath.Join(home, ".cache", "feov", "run-mirror", hex.EncodeToString(sum[:])[:12])
	if err := os.MkdirAll(mirror, 0o755); err != nil {
		return "", err
	}
	src := filepath.Join(runDir, "records")
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

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// PROSE IS READ THROUGH THE RESOLVER, NEVER OFF THE FLAG.
//
// # Why a source scan and not another end-to-end case
//
// The behavioural comparison drives both spellings through a verb and compares the record. It is
// the right shape and it does not scale: 51 commands take prose, most need board state to run at
// all, and the version of it that existed covered eight — with both offenders outside the eight.
// A case table is a list of verbs someone remembered.
//
// The invariant underneath is one line of code, at every site: prose comes from seat.Reason /
// flags.ReadPayload, which resolves --reason, --reason-file and `--reason-file -` into one string.
// A site that reads the --reason FLAG instead gets the inline spelling only, and the file and
// stdin forms silently vanish into a field nobody set.
//
// # Both shapes this has taken
//
//	line-of-inquiry propose  registered the pair correctly and filled `line` from the raw flag,
//	                         so --reason worked and --reason-file was refused for a missing field
//	                         the seat had supplied.
//	spot-check, outcome      registered --reason by hand, skipping seat.Prose entirely, so neither
//	                         had a file form at all. `outcome --reason` is, by its own help, "the
//	                         only evidence the determination ever had" on a judged deadlock.
//
// The first was found by a seat filing friction after three refusals. The second was found by
// grepping for the first. Nothing found either one before that.
func TestNoVerbReadsTheProseFlagDirectly(t *testing.T) {
	// The resolver itself, and the one helper that hands it the flag names.
	allowed := map[string]bool{
		filepath.Join("seat", "seat.go"):    true, // Reason() — the resolver's own call
		filepath.Join("..", "flags"):        true,
		filepath.Join("seat", "shapes.go"):  true,
		filepath.Join("seat", "require.go"): true,
	}

	// The reads that take the FLAG rather than the channel. Registering it is fine — that is what
	// seat.Prose does — so this looks for the accessors, not the declaration.
	banned := []string{
		"Str(cmd, flags.Reason)",
		"SetSame(cmd, p, flags.Reason)",
		"GetString(flags.Reason)",
		"GetString(Reason)",
	}

	var scanned, hits int
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		if allowed[strings.TrimPrefix(path, "./")] {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		scanned++
		for _, bad := range banned {
			if strings.Contains(string(b), bad) {
				hits++
				t.Errorf("%s reads the prose FLAG directly (%s).\n\n"+
					"Use seat.Reason(cmd) or seat.SetReason(cmd, p, key), which resolves --reason, "+
					"--reason-file and `--reason-file -` into one string. Reading the flag gets the inline "+
					"spelling only: a seat passing a heredoc has its prose silently dropped, and the write "+
					"is then refused for a field it did supply.", path, bad)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if scanned == 0 {
		t.Fatal("scanned no source at all — a walk that reads nothing reports every file clean")
	}
	t.Logf("%d file(s) scanned, %d direct read(s)", scanned, hits)
}

package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A PROGRAM'S BYTES REACH A RECORD STRING FIELD ONLY AS VALID UTF-8. A drift carries what a program
// wrote — its output, the names of the files it left under out/ — and a program may write any byte.
// A record string field holding invalid UTF-8 is written without complaint and then refused by every
// read that projects the body (protojson), so the act is on the record and unreadable. Each case is
// driven through the real write: `compute` records the drift of two runs that moved, and `reproduce`
// the drift of a re-run that differed.

// driftProgram prints its first argument's file, and leaves an out/ file named by its second's bytes
// when that file is not empty; with a third argument it also leaves one named by an invalid byte and
// the length of its working directory, which moves between the two copies.
const driftProgram = `cat "$1"
if [ -s "$2" ]; then : > "out/$(cat "$2")"; fi
if [ -n "$3" ]; then : > "out/$(printf '\377')$(pwd | wc -c | tr -d ' ')"; fi
`

// A NAME THAT IS NOT UTF-8 EXISTS ONLY WHERE NAMES ARE BYTES. Windows stores a file name as UTF-16,
// so the program's 0xff byte reaches the directory listing as U+00FF and the drift names it as text;
// elsewhere the byte survives and the drift Go-quotes the name. nameOpen and nameClose are what the
// drift puts around a name that began with that byte.
func nameOpen() string {
	if runtime.GOOS == "windows" {
		return "\u00ff"
	}
	return `"\xff`
}

func nameClose() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	return `"`
}

// validDrift fails the test unless the body's drift is valid UTF-8 and the body projects.
func validDrift(t *testing.T, body proto.Message, drift string) {
	t.Helper()
	if !utf8.ValidString(drift) {
		t.Errorf("the recorded drift is not valid UTF-8: %q", drift)
	}
	if _, err := protojson.Marshal(body); err != nil {
		t.Errorf("the recorded body does not project: %v", err)
	}
}

func TestAReRunsDriftIsValidUTF8(t *testing.T) {
	for _, c := range []struct {
		name       string
		ctl, outNm string
		want       string
	}{
		// The program's output is not UTF-8: the drift renders it Go-quoted, every byte kept.
		{"binary output", "\xff\xfe\x00bin\n", "", `"\xff\xfe\x00bin\n"`},
		// The cut at the drift's limit falls inside a rune: the drift ends on the rune before it.
		{"a cut inside a rune", strings.Repeat("a", 399) + "é tail", "", strings.Repeat("a", 399) + "…"},
		// The program leaves an out/ file whose name is not UTF-8.
		{"an out/ name", "a\n", "\xff", "out/" + nameOpen() + nameClose() + " written by the re-run only"},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := newRun(t)
			computeSeat(t, runDir, "# H\n\nThe program prints a.\n")
			draft := script(t, runDir, "p.sh", driftProgram)
			ctl, nm := filepath.Join(runDir, "ctl"), filepath.Join(runDir, "nm")
			for p, b := range map[string]string{ctl: "a\n", nm: ""} {
				if err := os.WriteFile(p, []byte(b), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if out, err := run(t, "compute", "--run", runDir, "--seat-id", "blue-respond", "--quote", "The program prints a.",
				"--dir", draft, "--reason", "r", "--", "bash", "p.sh", ctl, nm); err != nil {
				t.Fatalf("compute: %v\n%s", err, out)
			}
			k := lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
			if err := os.WriteFile(ctl, []byte(c.ctl), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(nm, []byte(c.outNm), 0o644); err != nil {
				t.Fatal(err)
			}
			if out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-logic", "--anchor", k, "--as", "unsound",
				"--reason", "it prints what its input says"); err != nil {
				t.Fatalf("reproduce: %v\n%s", err, out)
			}
			r := lastBody(t, runDir, &recordpb.Reproduce{})
			if r.GetRerun() != recordpb.Rerun_RERUN_DIFFERS {
				t.Fatalf("rerun %s, want differs", r.GetRerun())
			}
			validDrift(t, r, r.GetDrift())
			if !strings.Contains(r.GetDrift(), c.want) {
				t.Errorf("drift %q, want it to carry %q", r.GetDrift(), c.want)
			}
		})
	}
}

func TestAComputationsDriftIsValidUTF8(t *testing.T) {
	runDir := newRun(t)
	computeSeat(t, runDir, "# H\n\nThe program prints a.\n")
	draft := script(t, runDir, "p.sh", driftProgram)
	ctl, nm := filepath.Join(runDir, "ctl"), filepath.Join(runDir, "nm")
	for p, b := range map[string]string{ctl: "a\n", nm: ""} {
		if err := os.WriteFile(p, []byte(b), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if out, err := run(t, "compute", "--run", runDir, "--seat-id", "blue-respond", "--quote", "The program prints a.",
		"--dir", draft, "--reason", "r", "--variance", "it prints a", "--", "bash", "p.sh", ctl, nm, "moving"); err != nil {
		t.Fatalf("compute: %v\n%s", err, out)
	}
	k := lastBody(t, runDir, &recordpb.Computation{})
	if k.GetBasis() != recordpb.Basis_BASIS_VARIANT {
		t.Fatalf("basis %s, want variant — the out/ name was meant to move between the copies", k.GetBasis())
	}
	validDrift(t, k, k.GetDrift())
	if !strings.Contains(k.GetDrift(), "out/"+nameOpen()) {
		t.Errorf("drift %q, want it to name the out/ file as the platform lists it", k.GetDrift())
	}
}

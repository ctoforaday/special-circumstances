package migratecli

import (
	"bytes"
	"strings"
	"testing"
)

// feov-migrate DOES NOT KEEP THE LAST OF A FLAG STATED TWICE EITHER: a second --from or --to would
// migrate a directory other than the one a first value names. --accept-loss repeats by declaration.
func TestAFlagStatedTwiceIsRefusedAndAListTypeRepeats(t *testing.T) {
	refused := func(args ...string) (string, int) {
		var stdout, stderr bytes.Buffer
		code := Main(args, &stdout, &stderr)
		return stderr.String(), code
	}
	said, code := refused("--from", "a-dir", "--from", "b-dir", "--to", "c-dir")
	if code != 2 || !strings.Contains(said, "--from was given twice") || !strings.Contains(said, `"a-dir"`) || !strings.Contains(said, `"b-dir"`) {
		t.Errorf("exit %d, said: %s", code, said)
	}
	said, _ = refused("--from", "no-such-dir", "--to", "c-dir", "--accept-loss", "x", "--accept-loss", "y")
	if strings.Contains(said, "given twice") {
		t.Errorf("--accept-loss repeats and was refused for it: %s", said)
	}
}

// A REFUSED REPEAT UNDER --json IS THE ENVELOPE, as every refusal of this tool is.
func TestARepeatedFlagUnderJSONIsTheEnvelope(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Main([]string{"--json", "--from", "a", "--from", "b"}, &stdout, &stderr); code != 2 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout.String(), `"code":"validation"`) || !strings.Contains(stdout.String(), "given twice") {
		t.Errorf("stdout: %s", stdout.String())
	}
}

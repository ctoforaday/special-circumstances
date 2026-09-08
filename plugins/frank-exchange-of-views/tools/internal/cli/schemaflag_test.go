package cli

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// `--schema` PRINTS THE EPOCH SETUP COMPARES, and nothing asserted it until now.
//
// setup.PreflightRecordBinary shells `<bin> --schema`, Atoi's stdout, and refuses the run when it
// disagrees with the plugin manifest's `eventSchema`. That is the guard which produces the "FORMER
// record format … THIS BINARY CANNOT READ" refusal on an archived run — so the whole epoch check
// rests on this flag printing one parseable integer equal to record.EventSchema.
//
// IT WAS DRIVEN NOWHERE. Not by the release sweep, not by the surface tests, not here. It was
// invisible to the coverage gate for a structural reason rather than an oversight: CommandFlags
// reports `Flags()`, which is local-only, so the four root persistent flags were absent from the
// gate's denominator and could not be reported undriven however long they went undriven.
//
// THE ASSERTIONS ARE setup's OWN READS, in order, because a test that merely checked for
// non-empty output would pass on prose the Atoi rejects.
func TestSchemaFlagPrintsTheEpochSetupCompares(t *testing.T) {
	// THROUGH ExecuteRoot AND os.Args, because that is where the flag lives. It is answered by a
	// scan of os.Args BEFORE cobra dispatches — deliberately, so a binary setup does not yet
	// trust still answers even when the argv would otherwise be refused. Driving cobra instead
	// reaches the "you named no command" refusal and proves nothing about the real path.
	var out, errOut bytes.Buffer
	root := NewRootFor("red-merge-r1")
	root.SetOut(&out)
	root.SetErr(&errOut)
	saved := os.Args
	os.Args = []string{"feov-record", "--schema"}
	defer func() { os.Args = saved }()
	if err := ExecuteRoot(root); err != nil {
		t.Fatalf("--schema returned an error; setup treats a non-zero exit as an unrunnable binary: %v", err)
	}

	// 1. Parseable as an integer — setup does exactly this and refuses on failure.
	text := strings.TrimSpace(out.String())
	got, err := strconv.Atoi(text)
	if err != nil {
		t.Fatalf("--schema printed %q, which setup's Atoi rejects — it would report the binary as "+
			"predating the check or not being feov-record at all", text)
	}
	// 2. Equal to the epoch this binary actually writes. If these drift, setup compares the
	// manifest against a number no record carries and the guard passes or fails on a fiction.
	if got != record.EventSchema {
		t.Errorf("--schema printed %d and this binary writes epoch %d — setup compares the plugin "+
			"manifest against the printed value, so a mismatch here makes the epoch guard answer "+
			"about the wrong number", got, record.EventSchema)
	}
	// 3. ONE integer and nothing else. setup Atoi's the WHOLE of stdout, so a banner, a warning
	// or a trailing note turns a working binary into "predates the check".
	if strings.ContainsAny(text, " \t\n") {
		t.Errorf("--schema printed %q — setup parses the whole of stdout, so anything beside the "+
			"integer is read as a malformed epoch", text)
	}
}

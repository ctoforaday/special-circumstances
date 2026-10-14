package refusalcheck

import "testing"

var roles = []string{"bench", "blue", "chair", "lens", "operator"}

// The reading tells a refusal that names its command again from one that starts at what is wrong.
func TestAgainTellsARepeatedNameFromADiagnosis(t *testing.T) {
	for _, c := range []struct {
		path, diagnosis string
		again           bool
	}{
		{"finding", "requires --reason: the explanation red re-audits the repair against", false},
		{"finding", "lens finding requires --reason: the explanation", true},
		{"finding", "lens: requires --reason", true},
		{"finding", "finding: --quote was not found", true},
		{"finding", "finding requires --reason", true},
		{"finding", "findings are raised against the report as it stands", false},
		{"finding", "record: --reason is empty", true},
		{"close", "close --id names gap G-1, which is already CLOSED", true},
		{"close", "--id names gap G-1, which is already CLOSED — closing it twice double-counts; the chair's `carry --carried-from 1` restates it", false},
		{"carry", "red-chair has already recorded a close on \"G-1\" this sitting", false},
		{"mint", "mint refused — the first sentence prints in the report", true},
		{"mint", "--supersedes names G-1, which no mint event has created", false},
		{"prove", "blue's `prove` --answers names gap G-1, which no mint event created", true},
		{"reproduce", "the proof is blue's `prove` and its script is gone", false},
		{"edit", "No act of blue's adds a row; blue's `edit` answering the gap is what takes it out", false},
		{"regrade", "red-lens-evidence has already recorded a regrade on \"G-1\" this sitting", false},
		{"edit", "your old span carries finding anchor <!--fx:F-1--> (its claim leaves by `edit` down to the bare anchor, then `retire`)", false},
		{"motion docket file", "motion docket file: no gap G-1", true},
		{"motion docket file", "lens motion docket file requires --reason", true},
		{"motion docket file", "file: no gap", true},
		{"motion docket file", "a docket motion names the gap the bench is to decide", false},
		{"motion docket rule", "motion rule requires --tension (the values that pulled against each other)", true},
		{"motion docket rule", "requires --principle (the rule the bench applied)", false},
		{"motion grade", "a grade motion names no verb — it has file: see `motion grade --help`", false},
		{"motion", "`motion` has no `nonsense` act on any seat's surface", true},
		{"motion", "no `nonsense` act exists here, on any seat's surface — yours has docket, grade, petition.", false},
		{"avenue", "This act is the chair's `avenue review`, and it is not on your surface", false},
		{"count-claims", `unknown command "extra" for "feov-record count-claims"`, true},
		{"count-claims", `takes no arguments; got "extra"`, false},
		{"close", "requires --reason\n\n      --reason string   the closure's argument\n\nRun `feov-record close --help` for what it takes, flag by flag.", false},
		{"close", "see `feov-record closing --help`", false},
		{"avenue", "see `feov-record avenue review`", false},
		{"scorecard", `--card "x" is not a scorecard. usage: feov-record scorecard --run <dir>`, true},
		{"fetch", "fetch: https://example.org returned HTTP 404", true},
		{"show report", "render: no base has been ingested for this run — there is nothing to render", true},
		{"show report", "no base has been ingested for this run — there is nothing to render", false},
		{"show report", "replaying edit 3 of 5: the quote is gone", false},
		{"show board", "connecting: still busy after 5s", true},
		{"carry", "asking the record for its open gaps: no such table", false},
		{"carry", "record: asking the record for its open gaps: no such table", true},
		{"show report", "recordsql: opening the record: no such file", true},
		{"finding", "--about-kind takes one of section|avenue|gap and was given nothing", false},
		{"", `no command named "x" exists, on your surface or on any other seat's`, false},
		{"", "lens: you named no command", true},
	} {
		if got := Again(c.path, roles, c.diagnosis); (len(got) > 0) != c.again {
			t.Errorf("`%s` refused with %q: named again = %v (%v), want %v", c.path, c.diagnosis, len(got) > 0, got, c.again)
		}
	}
}

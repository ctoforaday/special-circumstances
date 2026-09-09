package seat

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// The closure payload is shared by two verbs in two trees: `lens close` — the originator closes
// its gap (plans/roundless.md §III.B.3) — and `merge carry`, the chair restating a closure the
// archive already holds. One payload, one flag set, one result shape, so the two cannot drift.

// ClosureFlags registers --id, --as and --superseded-by.
func ClosureFlags(c *cobra.Command) {
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap id")
	_ = c.MarkFlagRequired(flags.ID)
	enumhelp.Flag(c, flags.As, record.MustEnum("close", "closure_class"), ("HOW the gap ended. One vocabulary with the bench's dispositions since #342 — a reader no longer has to know which verb produced a closure before it can interpret the word"))
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.SupersededBy, "the gap id carrying the unresolved remainder forward")
}

// ClosurePayload reads the closure a verb carries; the disposition defaults to repaired.
func ClosurePayload(cmd *cobra.Command) (*recordpb.Close, error) {
	word := Str(cmd, flags.As)
	if word == "" {
		// THE DEFAULT IS THE PLAIN REPAIR, and it is spelled from the schema rather than typed.
		// It was the literal "closed", which stopped being a word this vocabulary carries when the
		// values were renamed to name the DEFECT's state rather than the paperwork — so every
		// close that omitted --as was refused by its own default.
		word = recordpb.Word(recordpb.Disposition_DISPOSITION_REPAIRED)
	}
	class, ok := record.DispositionOf(word)
	if !ok {
		return nil, feov.Errorf(feov.Validation,
			"lens close: %q is not a disposition — an unrecognized word lands in no bucket and the gap reads as closed for no stated reason", word)
	}
	// A MERGE MAY CLOSE AND MAY NOT CARRY, and the subset comes off the vocabulary rather than from
	// a word typed here. The database refuses the same value through the CHECK the schema generates
	// from `subset: "closes"`, so this refusal is the teaching copy of a constraint that holds even
	// against SQL written straight at the file.
	if !recordpb.Closes(class) {
		return nil, feov.Errorf(feov.Validation,
			"lens close: %q defers the gap instead of closing it, and deferring is the BENCH decision — a close asserts a verified repair. Put it before the bench with `motion docket file`, which the bench then rules, or close it with a class that states what the repair was", word)
	}
	// The SHARED prose channel, not a private one. close hand-rolled its own --file read and so
	// was the only prose-bearing verb with no --text at all — a verb that opts out of the shared
	// helper drifts from it by construction.
	prose, err := Reason(cmd)
	if err != nil {
		return nil, err
	}
	return &recordpb.Close{
		GapId:        proto.String(Str(cmd, flags.ID)),
		ClosureClass: &class,
		Successor:    OptStr(cmd, flags.SupersededBy),
		Prose:        proto.String(prose),
	}, nil
}

// CloseResult is what close and carry answer with.
type CloseResult struct {
	GapID   string `json:"gap_id"`
	Class   string `json:"class"`
	Carried bool   `json:"carried,omitempty"`
}

func (r CloseResult) Human() string {
	if r.Carried {
		return "carried the closure of " + r.GapID + " (" + r.Class + ")"
	}
	return "closed " + r.GapID + " (" + r.Class + ")"
}

package record

import (
	"fmt"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// ResolveAbout turns the --about-kind/--about pair into a checked anchor, for EVERY verb that
// takes one. admitted is the set the verb declares for --about-kind: a gap is never about a
// contradiction, and a finding may be.
//
// IT LIVES HERE RATHER THAN BESIDE EITHER CALLER BECAUSE THERE ARE TWO. `lens finding` files the
// finding and `lens mint` files the gap that finding becomes; a seat that anchors an omission at
// the first step and then has to anchor it differently at the second is being asked to hold two
// vocabularies for one act. Two copies of this would also be two copies of the checks, free to
// drift on which references the record verifies — and the checks are the entire reason the pair
// beats a borrowed quote.
//
// The refusal names the flags and not the command: whichever verb the seat typed, the renderer
// prints it.
func ResolveAbout(run Run, admitted EnumField, kind, ref string) (*recordpb.AboutKind, *string, error) {
	if kind == "" && ref == "" {
		return nil, nil, nil
	}
	// AS TYPED. A value with space around it — or of space alone — is refused rather than trimmed:
	// trimming turned a blank pair into no anchor at all, recorded as an act about nothing.
	for _, f := range []struct{ flag, v string }{{"--about-kind", kind}, {"--about", ref}} {
		if f.v != strings.TrimSpace(f.v) {
			return nil, nil, feov.Errorf(feov.Validation, "%s %q has space around it: pass the value exactly, or leave the flag off", f.flag, f.v)
		}
	}
	// HALF AN ANCHOR IS NOT AN ANCHOR. Accepting a kind with no reference would record what sort of
	// thing this is about without saying which one, and a reader cannot follow that anywhere.
	if kind == "" || ref == "" {
		return nil, nil, fmt.Errorf("--about-kind and --about are one anchor and are given together — " +
			"the kind says what sort of thing you are naming, the reference says which one")
	}
	// THE VERB'S OWN DECLARED SET, so the refusal lists what this verb takes.
	names := Names(admitted.Values)
	k, ok := AboutKindOf(kind)
	if !ok || k == recordpb.AboutKind_ABOUT_KIND_UNSPECIFIED || !slices.Contains(names, kind) {
		if k == recordpb.AboutKind_ABOUT_KIND_CONTRADICTION {
			return nil, nil, fmt.Errorf("a gap is not about a contradiction: raise the contradiction with a finding (`--about-kind contradiction --about <its id>`), and mint the gap from that finding with --found-by")
		}
		return nil, nil, fmt.Errorf("%q is not a thing this can be about (%s)", kind, strings.Join(names, " | "))
	}
	// THE REFERENCE IS CHECKED AGAINST THE RECORD, which is the whole advantage over a quote
	// borrowed from a nearby paragraph: an inquiry id either names an inquiry this run proposed or it
	// does not. A section name is not checked — the report's headings are blue's to change mid-run,
	// and refusing on a stale one would refuse the finding rather than the staleness.
	switch k {
	case recordpb.AboutKind_ABOUT_KIND_INQUIRY:
		if err := RequireInquiryRef(run, ref); err != nil {
			return nil, nil, err
		}
	case recordpb.AboutKind_ABOUT_KIND_GAP:
		db, err := openRunForRead(run)
		if err != nil {
			return nil, nil, err
		}
		if err := requireGap(db, ref, "--about"); err != nil {
			return nil, nil, err
		}
	case recordpb.AboutKind_ABOUT_KIND_CONTRADICTION:
		db, err := openRunForRead(run)
		if err != nil {
			return nil, nil, err
		}
		if err := requireOpenContradiction(db, ref); err != nil {
			return nil, nil, err
		}
	}
	return &k, proto.String(ref), nil
}

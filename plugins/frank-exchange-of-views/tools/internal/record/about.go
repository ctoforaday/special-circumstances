package record

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// ResolveAbout turns the --about-kind/--about pair into a checked anchor, for EVERY verb that
// takes one.
//
// IT LIVES HERE RATHER THAN BESIDE EITHER CALLER BECAUSE THERE ARE TWO. `lens finding` files the
// finding and `merge mint` files the gap that finding becomes; a seat that anchors an omission at
// the first step and then has to anchor it differently at the second is being asked to hold two
// vocabularies for one act. Two copies of this would also be two copies of the checks, free to
// drift on which references the record verifies — and the checks are the entire reason the pair
// beats a borrowed quote.
//
// `verb` is the caller's own name, so the refusal names the command the seat actually typed. It is
// the only thing that differs between them.
func ResolveAbout(verb string, run Run, kind, ref string) (*recordpb.AboutKind, *string, error) {
	kind, ref = strings.TrimSpace(kind), strings.TrimSpace(ref)
	if kind == "" && ref == "" {
		return nil, nil, nil
	}
	// HALF AN ANCHOR IS NOT AN ANCHOR. Accepting a kind with no reference would record what sort of
	// thing this is about without saying which one, and a reader cannot follow that anywhere.
	if kind == "" || ref == "" {
		return nil, nil, fmt.Errorf("%s: --about-kind and --about are one anchor and are given together — "+
			"the kind says what sort of thing you are naming, the reference says which one", verb)
	}
	k, ok := AboutKindOf(kind)
	if !ok || k == recordpb.AboutKind_ABOUT_KIND_UNSPECIFIED {
		return nil, nil, fmt.Errorf("%s: %q is not a thing this can be about (section | inquiry | gap)", verb, kind)
	}
	// THE REFERENCE IS CHECKED AGAINST THE RECORD, which is the whole advantage over a quote
	// borrowed from a nearby paragraph: an avenue id either names a line this run proposed or it
	// does not. A section name is not checked — the report's headings are blue's to change mid-run,
	// and refusing on a stale one would refuse the finding rather than the staleness.
	switch k {
	case recordpb.AboutKind_ABOUT_KIND_INQUIRY:
		if err := RequireInquiryRef(run, ref); err != nil {
			return nil, nil, err
		}
	case recordpb.AboutKind_ABOUT_KIND_GAP:
		if err := GapExists(run.Dir(), ref); err != nil {
			return nil, nil, err
		}
	}
	return &k, proto.String(ref), nil
}

package record

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportheading"
)

// aboutShapes are the --about-kinds whose reference is an id of a known shape.
var aboutShapes = map[recordpb.AboutKind]func() *flags.ShapedValue{
	recordpb.AboutKind_ABOUT_KIND_INQUIRY:  flags.InquiryID,
	recordpb.AboutKind_ABOUT_KIND_QUESTION: flags.QuestionPartID,
	recordpb.AboutKind_ABOUT_KIND_ANCHOR:   flags.AboutAnchor,
	recordpb.AboutKind_ABOUT_KIND_GAP:      flags.GapID,
}

// ResolveAbout turns the --about-kind/--about pair into a checked anchor, for EVERY verb that
// takes one. admitted is the set the verb declares for --about-kind. The checks are the entire
// reason the pair beats a borrowed quote, so they live in this one place.
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
	// thing this is about without saying which one, and a reader cannot follow that anywhere. THE
	// QUESTION IS THE ONE KIND WHOSE REFERENCE MAY BE ABSENT: there is one question, and a part of it
	// no question part covers has no id to name — the gap is about the question itself.
	questionWhole := kind == recordpb.Word(recordpb.AboutKind_ABOUT_KIND_QUESTION) && ref == ""
	if (kind == "" || ref == "") && !questionWhole {
		return nil, nil, fmt.Errorf("--about-kind and --about are one anchor and are given together — " +
			"the kind says what sort of thing you are naming, the reference says which one")
	}
	// THE VERB'S OWN DECLARED SET, so the refusal lists what this verb takes.
	names := Names(admitted.Values)
	k, ok := AboutKindOf(kind)
	if !ok || k == recordpb.AboutKind_ABOUT_KIND_UNSPECIFIED || !slices.Contains(names, kind) {
		return nil, nil, fmt.Errorf("%q is not a thing this can be about (%s)", kind, strings.Join(names, " | "))
	}
	// THE REFERENCE IS CHECKED AGAINST THE RECORD, which is the whole advantage over a quote
	// borrowed from a nearby paragraph: an inquiry id either names an inquiry this run proposed or it
	// does not, and a section names one heading of the report as it stands (§III.B0 "Sections").
	//
	// THE SHAPE FIRST, or an existence check answers for it: an id spelled another way reads as an id
	// the record lacks, and `--about '<!--computation:K-050249f6-->'` was told no computation on the
	// record carries an id one did (universe-m37). An anchor's token is taken whole, as the other
	// anchor flags take it, and the bare id is what is stored.
	if shape := aboutShapes[k]; shape != nil && ref != "" {
		v := shape()
		if err := v.Set(ref); err != nil {
			return nil, nil, feov.Errorf(feov.Validation, "--about with --about-kind %s: %v", kind, err)
		}
		ref = v.String()
	}
	switch k {
	case recordpb.AboutKind_ABOUT_KIND_INQUIRY:
		if err := RequireInquiryRef(run, ref, "--about"); err != nil {
			return nil, nil, err
		}
	case recordpb.AboutKind_ABOUT_KIND_QUESTION:
		if ref != "" {
			db, err := openRunForRead(run)
			if err != nil {
				return nil, nil, err
			}
			if err := requireQuestionPartRef(db, ref, "--about"); err != nil {
				return nil, nil, err
			}
		}
	case recordpb.AboutKind_ABOUT_KIND_SECTION:
		if err := requireSectionRef(run, ref); err != nil {
			return nil, nil, err
		}
	case recordpb.AboutKind_ABOUT_KIND_ANCHOR:
		db, err := openRunForRead(run)
		if err != nil {
			return nil, nil, err
		}
		if err := requireAboutAnchor(db, k, ref); err != nil {
			return nil, nil, err
		}
	}
	return &k, proto.String(ref), nil
}

// requireSectionRef is a section reference checked against the report (§III.B0 "Sections"): it keys,
// by reportheading.HeadingKey, exactly one heading of blue's text as it stands, or one of the sections
// assembly composes from the inquiries (reportheading.InquirySections). A reference keying none or two
// is refused, naming the headings the text holds.
func requireSectionRef(run Run, ref string) error {
	if composedSection(ref) {
		return nil
	}
	report, _, err := ReportAtHead(run)
	if err != nil {
		return fmt.Errorf("the report could not be rendered, so whether a heading keys to --about %q cannot be told: %w", ref, err)
	}
	_, _, err = anchortext.HeadingLocation(report, ref)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, anchortext.ErrNoHeading):
		return feov.Errorf(feov.Validation, "--about %q with --about-kind section keys no heading of the report. A section gap names a section the report holds, by its heading; the report's headings are: %s. Nothing was recorded", ref, sectionHeadingList(report))
	case errors.Is(err, anchortext.ErrHeadingTwice):
		return feov.Errorf(feov.Validation, "--about %q with --about-kind section keys two headings of the report, so which section the gap is about cannot be told; the report's headings are: %s. Nothing was recorded", ref, sectionHeadingList(report))
	}
	return feov.Errorf(feov.Validation, "--about %q with --about-kind section: %v. Nothing was recorded", ref, err)
}

// sectionHeadingList is the headings of report, quoted, and the composed inquiry sections, for a
// refusal that names what a section reference may key.
func sectionHeadingList(report string) string {
	var out []string
	for _, h := range anchortext.Headings(report) {
		out = append(out, strconv.Quote(reportheading.HeadingKey(h.Text)))
	}
	for _, s := range reportheading.InquirySections {
		out = append(out, strconv.Quote(reportheading.HeadingKey(s.Heading)))
	}
	return strings.Join(out, ", ")
}

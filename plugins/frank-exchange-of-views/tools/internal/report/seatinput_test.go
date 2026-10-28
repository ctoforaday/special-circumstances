package report

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// NO SEAT-WRITTEN INPUT IS MARKUP ON THE PAGE, WHEREVER IT LANDS (#1569). The renderer builds HTML
// by hand, so each place it writes a value is a place a value can break out of: a footnote label in
// an href, a title attribute, a destination whose scheme is spelled with character references. The
// class is tested over every input a seat writes that reaches report.html, each with every hostile
// string: what the renderer writes, before the wall, holds no element, attribute, comment or link
// outside the allowlist; and the wall (bodyPolicy) changes nothing a browser reads of it.

// hostileText is what a seat could type to make the page run, load or follow something: a tag; an
// attribute broken out of with either quotation mark; a scheme in plain, mixed case, with character
// references, with a reference that is one only once the attribute is read, and with control
// characters; a data URL; a comment's close.
var hostileText = []string{
	`<img src=x onerror=alert(1)>`,
	`"><img/src=x/onerror=alert(1)>`,
	`'><svg/onload=alert(1)>`,
	`a"onmouseover="alert(1)`,
	`a'onmouseover='alert(1)`,
	`javascript:alert(1)`,
	`JaVaScRiPt:alert(1)`,
	`javascript&colon;alert(1)`,
	`javascript&amp;colon;alert(1)`,
	`&#106;avascript:alert(1)`,
	`&#x6A;avascript&#x3A;alert(1)`,
	`java&Tab;script:alert(1)`,
	"java\tscript:alert(1)",
	"\x01java\x02script:\x00alert(1)\x1e",
	`data:text/html,<script>alert(1)</script>`,
	`--><script>alert(1)</script><!--`,
}

// textInputs is every place the report's text puts a seat's words on the page, each a document with
// the words at {h}, and the title the set ships under, which may hold them too. The record's text reaches the page through
// these and through recordInputs.
var textInputs = []struct{ name, title, doc string }{
	{"the report's title", "# On {h} — research report", "Body.\n"},
	{"a footnote reference in the title", "# On 91[^{h}]", "Body.\n"},
	{"a section heading", "# T", "## {h}\n\nBody.\n"},
	{"a deeper heading", "# T", "### {h}\n"},
	{"body text", "# T", "Before {h} after.\n"},
	{"emphasis", "# T", "*{h}* and **{h}**\n"},
	{"a code span", "# T", "`{h}`\n"},
	{"a list item", "# T", "- {h}\n- G-00000001 {h}\n"},
	{"an ordered list item", "# T", "3. {h}\n"},
	{"a table cell", "# T", "| h |\n|---|\n| {h} |\n"},
	{"a blockquote", "# T", "> {h}\n"},
	{"a fence's info string", "# T", "```{h}\ncode\n```\n"},
	{"a fence's code", "# T", "```\n{h}\n```\n"},
	{"a link's destination", "# T", "[x]({h})\n"},
	{"a link's destination in angle brackets", "# T", "[x](<{h}>)\n"},
	{"a link to a sibling document's fragment", "# T", "[x](docket.md#{h})\n"},
	{"a link's title", "# T", "[x](https://example.invalid \"{h}\")\n"},
	{"a link's text", "# T", "[{h}](https://example.invalid)\n"},
	{"an image's destination", "# T", "![a]({h})\n"},
	{"an image's text", "# T", "![{h}](https://example.invalid/i.png)\n"},
	{"an autolink", "# T", "<{h}>\n"},
	{"a footnote label", "# T", "A claim[^{h}].\n\n[^{h}]: the note.\n"},
	{"a footnote's text", "# T", "A claim[^n].\n\n[^n]: {h}\n"},
	{"a footnote's text cited from another", "# T", "A claim[^n].\n\n[^n]: see[^m].\n[^m]: {h}\n"},
}

// recordInputs is every field of the record a seat writes that assembly places on the page, each
// set to the hostile words by assembleHostile (assembled through the production write path), and read back from
// the document it lands in: a citation's title and URL, a proof's reason, an inquiry's question,
// method and reason, a gap's problem, location, fix and check, a closing, a motion's basis, the
// bench's opinion, a position, the outcome's prose. Each is written behind a word of its own, which
// the test finds on the page, so a field that stops reaching it fails here rather than passing.
var recordInputs = map[string]string{
	"citation title": "Title ", "citation URL": "example.invalid/", "proof reason": "reason ", "inquiry question": "question ",
	"inquiry method": "method ", "inquiry reason": "why ", "gap problem": "problem ", "gap location": "location ",
	"gap fix": "fix ", "gap check": "check ", "closing": "closing ", "motion basis": "basis ", "opinion": "opinion ",
	"position": "position ", "outcome prose": "prose ",
}

// assembleHostile records every recordInput with the words h, and returns the set as it ships.
func assembleHostile(t *testing.T, h string) Shipped {
	t.Helper()
	runDir := newRun(t)
	const sha = "feed0000face1111"
	seedProof(t, runDir, sha, "console.log(1);", "1\n")
	blue := "# On the cache — research report\n\n## TL;DR\nCoherent<!--cite:C-00000001-->, and checked<!--proof:P-00000001-->.\n"
	seen := map[string]bool{}
	add := func(seatID string, body proto.Message) {
		t.Helper()
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: seatID}
		if !seen[seatID] {
			occasion := ""
			if record.SeatOwesOccasion(seatID) {
				occasion = "assemble"
			}
			if _, _, err := record.RegisterSeat(id, "", occasion); err != nil {
				t.Fatalf("register %s: %v", seatID, err)
			}
			seen[seatID] = true
		}
		if _, err := record.Append(id, body); err != nil {
			t.Fatalf("append %s/%T: %v", seatID, body, err)
		}
	}
	add("blue-synthesize", &recordpb.BaseIngest{Text: proto.String(blue)})
	add("blue-synthesize", &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), SourceSpan: proto.String("s"), SpanStart: proto.Int32(0), SpanEnd: proto.Int32(1), SpanTextSha: proto.String("t"), SourceMatched: proto.String("s"), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(), CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(),
		Label: proto.String("C-00000001"), Url: proto.String("https://example.invalid/" + h),
		Sha256: proto.String("deadbeef"), Title: proto.String("Title " + h), AccessDate: proto.String("2026-08-03"),
	})
	add("blue-synthesize", &recordpb.Proof{
		ProofId: proto.String("P-00000001"), ProofSha: proto.String(sha),
		ProofBasis: proto.String("reproducible"), Script: proto.String("check.js"), Text: proto.String("reason " + h),
	})
	add("red-chair", &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		GapId: proto.String("G1"), Problem: proto.String("problem " + h),
		Location: proto.String("location " + h), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH),
		AcceptanceCheck: proto.String("check " + h),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		RequiredFix:     proto.String("fix " + h),
	})
	add("red-chair", &recordpb.Position{Text: proto.String("position " + h)})
	add("blue-respond", &recordpb.Inquiry{
		InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED),
		Question: proto.String("question " + h),
	})
	add("blue-respond", &recordpb.Inquiry{
		InquiryId: proto.String("Q1"), SupersedesStatus: proto.String("1"),
		Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Method: proto.String("method " + h),
	})
	add("blue-respond", &recordpb.Inquiry{
		InquiryId: proto.String("Q2"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED),
		Question: proto.String("other " + h), Reason: proto.String("why " + h),
	})
	add("red-chair", &recordpb.Motion{MotionId: proto.String("M1"), Basis: proto.String("basis " + h),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}})
	add("blue-respond", &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("closing " + h)})
	add("judge", &recordpb.MotionRule{MotionId: proto.String("M1"), Opinion: proto.String("opinion " + h),
		Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REMANDED),
			Principle:   proto.String("correctness"), Tension: proto.String("cost vs certainty"), ReviewFlag: proto.String("false"),
			Settled:   proto.String("settled " + h),
			Direction: proto.String("direction " + h),
		}}})
	add("judge", &recordpb.Outcome{
		Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING),
		Prose:   proto.String("prose " + h),
	})
	set, err := AssembleShipped(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	return set
}

// honestOnThePage holds set to the class: each fragment the renderer writes (renderSet) holds nothing
// outside the allowlist, and the wall changes nothing in it; and the page's header and bodies, after
// the wall, hold nothing outside it either.
func honestOnThePage(t *testing.T, set Shipped) {
	t.Helper()
	title, bodies := renderSet(set, stdRenderer())
	for i, frag := range append([]string{title}, bodies...) {
		where := "the title"
		if i > 0 {
			where = set.docs[i-1].File
		}
		if bad := outsideAllowlist(frag); len(bad) > 0 {
			t.Errorf("the renderer wrote markup into %s: %q\n%s", where, bad, frag)
		}
		got, want := tokens(bodyPolicy.Sanitize(frag)), tokens(frag)
		if !slices.EqualFunc(got, want, func(a, b html.Token) bool {
			return a.Type == b.Type && a.Data == b.Data && slices.Equal(a.Attr, b.Attr)
		}) {
			t.Errorf("the wall changed what the renderer wrote into %s: it let through what the allowlist drops\n  renderer %q\n  wall     %q", where, frag, bodyPolicy.Sanitize(frag))
		}
	}
	page := RenderSite(set, record.Family{})
	if bad := outsideAllowlist(mastheadOf(page) + mainOf(page)); len(bad) > 0 {
		t.Errorf("report.html carries markup outside the allowlist: %q", bad)
	}
	// The <title> is the template's to escape: a browser shows the title as typed, a NUL as the
	// replacement character.
	if got, want := titleElement(t, page), strings.ReplaceAll(shortTitle(set), "\x00", "\uFFFD"); got != want {
		t.Errorf("the page's <title> reads %q, want the title as typed %q", got, want)
	}
}

// mastheadOf is what the renderer wrote into the page's header: the title's inside of its <h1>.
func mastheadOf(page string) string {
	h := page[strings.Index(page, `<header class="masthead">`):]
	h = h[strings.Index(h, "<h1>")+len("<h1>"):]
	return h[:strings.Index(h, "</h1>")]
}

// titleElement is the text a browser shows as the page's title.
func titleElement(t *testing.T, page string) string {
	t.Helper()
	z := html.NewTokenizer(strings.NewReader(page))
	for {
		switch z.Next() {
		case html.ErrorToken:
			t.Fatal("the page has no <title>")
		case html.StartTagToken:
			if tok := z.Token(); tok.Data == "title" {
				z.Next()
				return z.Token().Data
			}
		}
	}
}

func TestNoSeatInputIsMarkupOnThePage(t *testing.T) {
	for _, in := range textInputs {
		for _, h := range hostileText {
			t.Run(in.name+"/"+fmt.Sprintf("%q", h), func(t *testing.T) {
				docs := []Doc{
					{File: FileReport, Nav: "Report", Body: strings.ReplaceAll(in.doc, "{h}", h)},
					{File: FileDocket, Nav: "Board", Body: "## The board\n"},
				}
				honestOnThePage(t, stdReaders().shipSet(strings.ReplaceAll(in.title, "{h}", h), docs))
			})
		}
	}
	for _, h := range hostileText {
		t.Run(fmt.Sprintf("the record's fields/%q", h), func(t *testing.T) {
			set := assembleHostile(t, h)
			honestOnThePage(t, set)
			// Each field reached the page: a field no document carries tests nothing here.
			_, bodies := renderSet(set, stdRenderer())
			shown := textOf(strings.Join(bodies, ""))
			for field, w := range recordInputs {
				if !strings.Contains(shown, w) {
					t.Errorf("no document shows the %s (written %q…): the class test does not reach it", field, w)
				}
			}
		})
	}
}

// THE WALL STANDS OVER THE TITLE AS OVER EACH BODY. A title renderer that wrote a seat's markup ships
// none of it: the masthead holds only what the allowlist holds.
func TestTheAllowlistHoldsTheTitle(t *testing.T) {
	rr := stdRenderer()
	rr.title = func(string, map[string]string, string) string {
		return `kept<script>alert(1)</script><img src="x" onerror="y()"><a href="javascript:z()" onclick="w()">l</a>`
	}
	page := renderSite(stdReaders().shipSet("# T", []Doc{{File: FileReport, Nav: "Report", Body: "x\n"}}), record.Family{}, rr)
	if bad := outsideAllowlist(mastheadOf(page)); len(bad) > 0 {
		t.Errorf("the masthead carries what the allowlist exists to drop: %q\n%s", bad, mastheadOf(page))
	}
	if !strings.HasPrefix(mastheadOf(page), "kept") {
		t.Errorf("the allowlist dropped the title's text:\n%s", mastheadOf(page))
	}
}

// NO TOKEN SHIPS FROM TEXT COMPOSED AFTER THE STRIP. A proof's reason and a citation's title reach the
// documents after the strip that runs before the weave — the weave's notes, evidence.md's heading for
// the proof — and a token a seat copied into either shipped there as typed. Through real assembly. A
// proof's script is code, and evidence.md shows it as the bytes that ran, a token's shape and all. A
// field's lines are flattened before its tokens are taken: a token on an indented line is code in the
// field and prose on the one line the tool writes, and stripped in the field's reading it shipped.
func TestATokenInAProofReasonOrCitationTitleDoesNotShip(t *testing.T) {
	runDir := newRun(t)
	const sha = "feed0000face1111"
	const script = "console.log('<!--fx:F-0000beef-->'.length);"
	seedProof(t, runDir, sha, script, "26\n")
	blue := "# On the cache — research report\n\n## TL;DR\nCoherent<!--cite:C-00000001-->, and checked<!--proof:P-00000001-->.\n"
	id := record.Identity{Run: runtest.Open(t, runDir), SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	for _, body := range []proto.Message{
		&recordpb.BaseIngest{Text: proto.String(blue)},
		&recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), SourceSpan: proto.String("s"), SpanStart: proto.Int32(0), SpanEnd: proto.Int32(1), SpanTextSha: proto.String("t"), SourceMatched: proto.String("s"), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(), CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(),
			Label: proto.String("C-00000001"), Url: proto.String("https://example.invalid/a"),
			Sha256: proto.String("deadbeef"), Title: proto.String("Coherence <!--gap:G-0000abcd-->Proof <!--cite:C-…-->\n\n    <!--gap:G-00000001-->"), AccessDate: proto.String("2026-08-03"),
		},
		&recordpb.Proof{
			ProofId: proto.String("P-00000001"), ProofSha: proto.String(sha),
			ProofBasis: proto.String("reproducible"), Script: proto.String("check.js"),
			Text: proto.String("trial division <!--fx:F-0000beef-->settles it <!--proof:P-00000002-->\n\n    <!--gap:G-00000001-->"),
		},
	} {
		if _, err := record.Append(id, body); err != nil {
			t.Fatalf("append %T: %v", body, err)
		}
	}
	if _, err := Assemble(runtest.Open(t, runDir)); err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	page, err := os.ReadFile(filepath.Join(runDir, "report.html"))
	if err != nil {
		t.Fatal(err)
	}
	shown := map[string]string{"report.html": textOf(mainOf(string(page)))}
	for _, f := range []string{FileReport, FileEvidence} {
		md, err := os.ReadFile(filepath.Join(runDir, f))
		if err != nil {
			t.Fatal(err)
		}
		shown[f] = string(md)
	}
	for where, text := range shown {
		if strings.Contains(strings.ReplaceAll(text, script, ""), "<!--") {
			t.Errorf("%s shows a token as typed:\n%s", where, text)
		}
		for _, want := range []string{"Coherence Proof", "trial division settles it"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s does not show %q:\n%s", where, want, text)
			}
		}
	}
	for where, text := range map[string]string{FileEvidence: shown[FileEvidence], "report.html": textOf(string(page))} {
		if !strings.Contains(text, script) {
			t.Errorf("%s does not show the proof's script as it ran:\n%s", where, text)
		}
	}
}

// A FIELD'S CODE IS READ ON THE LINE IT SHIPS ON. plain writes a field after the tool's own text, so
// the field is inline text of that line: a fence it opens alone is prose there, and a code span it
// holds whole is code. A token keeps its place only in the code of that line.
func TestAFieldsTokensAreReadOnItsLine(t *testing.T) {
	for in, want := range map[string]string{
		"```x <!--gap:G-00000001--> y":      "```x y",
		"a\n\n    <!--gap:G-00000001--> b":  "a b",
		"a `<!--gap:G-00000001-->` b":       "a `<!--gap:G-00000001-->` b",
		"a ``\n<!--gap:G-00000001-->\n`` b": "a `` <!--gap:G-00000001--> `` b",
	} {
		if got := plain(in); got != want {
			t.Errorf("plain(%q) = %q, want %q", in, got, want)
		}
	}
}

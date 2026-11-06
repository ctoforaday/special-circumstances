package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN ACT ON A RETIRED ANCHOR IS THE ACT'S TO ANSWER, NOT THE RECORD'S SHAPE (judgeAct). A citation is
// retired; then a mint and a finding about it and a computation citing it are each refused live, naming
// the retire, and each written by the replay — an archived act legal when it was made migrates as it
// was made.
func TestAnActOnARetiredAnchorIsRefusedLiveAndReplayed(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	lens := sit(t, run, "red-lens-evidence")
	const c = "C-0000c0de"
	if _, err := replayed(t, blue, &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(),
		SourceSpan: proto.String("s"), SpanStart: proto.Int32(0), SpanEnd: proto.Int32(1), SpanTextSha: proto.String("t"), SourceMatched: proto.String("s"),
		WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(),
		CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(), Label: proto.String(c), Url: proto.String("https://example.org/c"),
		Title: proto.String("t"), CiteKey: proto.String("k-c"), Location: proto.String("The sky is blue.")}); err != nil {
		t.Fatalf("the cite: %v", err)
	}
	if _, err := replayed(t, blue, &recordpb.Retire{Claim: proto.String("The sky is blue."), Reason: proto.String("it does not hold"),
		Anchors: []string{c}, ClaimKept: proto.Bool(true), RemovalBasis: proto.String(RemovalAsserted)}); err != nil {
		t.Fatalf("the retire: %v", err)
	}
	for _, act := range []struct {
		name string
		who  Identity
		body func() proto.Message
		says string
	}{
		{"a mint about it", lens, func() proto.Message {
			return &recordpb.Mint{GapId: proto.String("G-0000a001"), AboutKind: recordpb.AboutKind_ABOUT_KIND_ANCHOR.Enum(), AboutRef: proto.String(c),
				Class: recordpb.GapClass_GAP_CLASS_VERIFICATION_SCOPE_BLINDSPOT.Enum(), Problem: proto.String("p"), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("c"),
				CheckKind: recordpb.CheckKind_CHECK_KIND_SOURCE.Enum(), Severity: recordpb.Grade_GRADE_HIGH.Enum(), Likelihood: recordpb.Grade_GRADE_HIGH.Enum(),
				Impact: recordpb.Grade_GRADE_HIGH.Enum()}
		}, "no gap about it"},
		{"a finding about it", lens, func() proto.Message {
			return &recordpb.Finding{Id: proto.String("F-0000f001"), AboutKind: recordpb.AboutKind_ABOUT_KIND_ANCHOR.Enum(), AboutRef: proto.String(c),
				Text: proto.String("the anchor is wrong"), Severity: recordpb.Grade_GRADE_LOW.Enum(), Likelihood: recordpb.Grade_GRADE_LOW.Enum(), Impact: recordpb.Grade_GRADE_LOW.Enum()}
		}, "no finding about it"},
		{"a computation citing it", blue, func() proto.Message {
			return &recordpb.Computation{Argv: []string{"python3", "s.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), ComputationId: proto.String("K-0000b001"),
				BundleSha: proto.String("sha-s"), Cites: proto.String(c)}
		}, "names a retired citation"},
	} {
		t.Run(act.name, func(t *testing.T) {
			if _, err := Append(act.who, act.body()); err == nil || !strings.Contains(err.Error(), act.says) || !strings.Contains(err.Error(), "blue retired it at event") {
				t.Errorf("live, %s = %v, want refused naming the retire (%q)", act.name, err, act.says)
			}
			if _, err := replayed(t, act.who, act.body()); err != nil {
				t.Errorf("the replay refused %s: %v", act.name, err)
			}
		})
	}
}

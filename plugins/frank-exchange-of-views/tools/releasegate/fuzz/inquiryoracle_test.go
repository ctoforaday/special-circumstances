package fuzz

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/proto"
)

// The acts an inquiry's judgement is read from, as the record holds them.
func oracleProposal(t *testing.T) *recordpb.Event {
	return recordtest.Event(t, "blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Question: proto.String("does it settle"),
		Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()})
}

func oracleFate(t *testing.T, st recordpb.InquiryStatus) *recordpb.Event {
	return recordtest.Event(t, "blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: st.Enum(),
		SupersedesStatus: proto.String("proposed"), Reason: proto.String("what the approach found")})
}

func oracleCheck(t *testing.T) *recordpb.Event {
	return recordtest.Event(t, record.InquiryOwner(), &recordpb.InquiryCheck{InquiryId: proto.String("Q1"), Reason: proto.String("its fate stands")})
}

// oracleGap is a gap about Q1 of class materiality cm, graded sev at its mint.
func oracleGap(t *testing.T, cm recordpb.ClassMaterial, sev recordpb.Grade) *recordpb.Event {
	return recordtest.Event(t, "red-lens-evidence", &recordpb.Mint{GapId: proto.String("G1"),
		AboutKind: recordpb.AboutKind_ABOUT_KIND_INQUIRY.Enum(), AboutRef: proto.String("Q1"),
		Class: recordtest.ClassFor(cm).Enum(), ClassMaterial: cm.Enum(), Problem: proto.String("the fate reads past what ran"),
		AcceptanceCheck: proto.String("c"), CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Severity: sev.Enum(), Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()})
}

// oracleRun seeds evs into a fresh run and opens it.
func oracleRun(t *testing.T, evs ...*recordpb.Event) record.Run {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	return runtest.Open(t, dir)
}

// THE INQUIRY ORACLE HOLDS A VERIFIED RUN TO EVERY INQUIRY BEING SOUND, and each way an inquiry
// falls short of it is a fixture: open, settled and never checked, checked before its latest fate,
// a material gap about it open, checked before a gap about it that was material at its mint and has
// since been graded down, and checked before a gap that was never material. Two fixtures are sound:
// checked after its fate, and checked after a gap that was never material. Any other verdict holds no
// inquiry to anything.
func TestTheInquiryOracleHoldsAVerifiedRunToSoundInquiries(t *testing.T) {
	settled := recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED
	for _, tc := range []struct {
		name   string
		evs    func(t *testing.T) []*recordpb.Event
		breach string // "" = sound
	}{
		{"open", func(t *testing.T) []*recordpb.Event { return []*recordpb.Event{oracleProposal(t)} }, "a fate that settles nothing"},
		{"settled and never checked", func(t *testing.T) []*recordpb.Event {
			return []*recordpb.Event{oracleProposal(t), oracleFate(t, settled)}
		}, "never checked"},
		{"checked after its fate", func(t *testing.T) []*recordpb.Event {
			return []*recordpb.Event{oracleProposal(t), oracleFate(t, settled), oracleCheck(t)}
		}, ""},
		{"a new fate after its check", func(t *testing.T) []*recordpb.Event {
			return []*recordpb.Event{oracleProposal(t), oracleFate(t, settled), oracleCheck(t), oracleFate(t, recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED)}
		}, "before its latest fate"},
		{"a material gap about it open", func(t *testing.T) []*recordpb.Event {
			return []*recordpb.Event{oracleProposal(t), oracleFate(t, settled), oracleCheck(t),
				oracleGap(t, recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, recordpb.Grade_GRADE_MEDIUM)}
		}, "material gap(s) about inquiry Q1 open"},
		{"a gap material at its mint, since graded down", func(t *testing.T) []*recordpb.Event {
			return []*recordpb.Event{oracleProposal(t), oracleFate(t, settled), oracleCheck(t),
				oracleGap(t, recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, recordpb.Grade_GRADE_HIGH),
				recordtest.Event(t, "red-lens-evidence", &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordpb.Grade_GRADE_LOW.Enum(),
					Basis: proto.String("the consequence is bounded")})}
		}, "before a gap minted about it"},
		{"a gap never material, after its check", func(t *testing.T) []*recordpb.Event {
			return []*recordpb.Event{oracleProposal(t), oracleFate(t, settled), oracleCheck(t),
				oracleGap(t, recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, recordpb.Grade_GRADE_LOW)}
		}, "before a gap minted about it"},
		{"a gap never material, before its check", func(t *testing.T) []*recordpb.Event {
			return []*recordpb.Event{oracleProposal(t), oracleFate(t, settled),
				oracleGap(t, recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, recordpb.Grade_GRADE_LOW), oracleCheck(t)}
		}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := oracleRun(t, tc.evs(t)...)
			got := inquirySoundOracle(run, "VERIFIED")
			if (tc.breach == "") != (got == "") || !strings.Contains(got, tc.breach) {
				t.Fatalf("the oracle over a VERIFIED run = %q, want %q", got, tc.breach)
			}
			if got := inquirySoundOracle(run, "CEILING"); got != "" {
				t.Errorf("the oracle over a CEILING run = %q, want \"\": only a PASS holds an inquiry to soundness", got)
			}
		})
	}
}

// THE INQUIRY ORACLE FAILS ON A RECORD IT CANNOT READ. "" from it means every inquiry is sound, so a
// failed read answering "" is an unchecked inquiry reported sound. The view is dropped through the
// run's cached handle, so the oracle's read fails as a busy or malformed database would.
func TestTheInquiryOracleFailsOnARecordItCannotRead(t *testing.T) {
	run := oracleRun(t, oracleProposal(t), oracleFate(t, recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED), oracleCheck(t))
	if got := inquirySoundOracle(run, "VERIFIED"); got != "" {
		t.Fatalf("the oracle over a sound record = %q, want \"\"", got)
	}
	db, err := recordsql.Open(filepath.Join(run.Records(), "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP VIEW "inquiry_state"`); err != nil {
		t.Fatal(err)
	}
	got := inquirySoundOracle(run, "VERIFIED")
	if !strings.Contains(got, "could not be read") || !strings.Contains(got, "inquiry_state") {
		t.Errorf("the oracle over a record it cannot read = %q, want a failure naming the read (inquiry_state) — \"\" here is an inquiry passed without being checked", got)
	}
}

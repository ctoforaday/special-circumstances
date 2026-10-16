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

// inquiryOracleRun seeds one inquiry proposed with the question line, ruled as ruling, and then recorded underway by blue
// against the ruling, and returns the run and its board.
func inquiryOracleRun(t *testing.T, line string, ruling recordpb.InquiryRuling) (record.Run, record.Family) {
	t.Helper()
	run := runtest.New(t, recordtest.TmpRun(t))
	blue := record.Identity{Run: run, SeatID: "blue-respond"}
	red := record.Identity{Run: run, SeatID: "red-chair"}
	for _, step := range []struct {
		id   record.Identity
		body proto.Message
	}{
		{blue, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Question: proto.String(line), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()}},
		{red, &recordpb.MotionRule{
			MotionId: proto.String("Q1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY),
			Opinion:  proto.String("ruled as the line was proposed"),
			Ruling:   &recordpb.MotionRule_Inquiry{Inquiry: ruling},
		}},
		{blue, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY.Enum(),
			SupersedesStatus: proto.String("proposed"), Reason: proto.String("taking it up")}},
	} {
		if _, err := record.Append(step.id, step.body); err != nil {
			t.Fatal(err)
		}
	}
	board, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	return run, board
}

// THE INQUIRY ORACLE FAILS ON A RULING IT CANNOT READ. The read is record.InquiryRuling, whose ""
// means "red has not ruled" and sends the oracle past the inquiry; a failed read answering "" is
// therefore an unaudited inquiry reported clean.
//
// Two fixtures, because the failure has to be told apart from BOTH healthy answers: a record with
// a breach on it (ruled out of scope and taken up `underway`, on an inquiry the drive declines) and a record with
// none (endorsed, `underway`). The view is dropped through the run's cached handle, so the oracle's read fails as a
// busy or malformed database would.
func TestTheInquiryOracleFailsOnARulingItCannotRead(t *testing.T) {
	for _, tc := range []struct {
		name, line string
		ruling     recordpb.InquiryRuling
		healthy    string // "" = the oracle passes the healthy record
	}{
		{"a clean record", avEndorse, recordpb.InquiryRuling_INQUIRY_RULING_ENDORSED, ""},
		{"a record with a breach", avScope, recordpb.InquiryRuling_INQUIRY_RULING_OUT_OF_SCOPE, "ruled out_of_scope and underway anyway"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run, board := inquiryOracleRun(t, tc.line, tc.ruling)
			got := inquiryRulingOracle(run, board)
			if (tc.healthy == "") != (got == "") || !strings.Contains(got, tc.healthy) {
				t.Fatalf("the oracle over the healthy record = %q, want %q", got, tc.healthy)
			}

			db, err := recordsql.Open(filepath.Join(run.Records(), "record.db"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`DROP VIEW "inquiry_state"`); err != nil {
				t.Fatal(err)
			}
			got = inquiryRulingOracle(run, board)
			if !strings.Contains(got, "could not be read") || !strings.Contains(got, "inquiry_state") {
				t.Errorf("the oracle over a ruling it cannot read = %q, want a failure naming the read (inquiry_state) — \"\" here is an inquiry passed without being checked", got)
			}
		})
	}
}

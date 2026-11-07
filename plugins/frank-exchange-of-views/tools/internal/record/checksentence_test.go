package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE LEDGER CARRIES THE SENTENCE EACH CHECK JUDGED (ruling L5), for both kinds: a verify's claim, a
// reproduce's claim, and "" for a reproduce that recorded none, as a migrated one is. The sentence is
// how a check of an earlier wording surfaces; it decides no standing.
func TestTheLedgerCarriesTheSentenceEachCheckJudged(t *testing.T) {
	b := reopenedBoard(t).computation("K-1", "s1").computation("K-2", "s2")
	b.add(evLens, &recordpb.Reproduce{ComputationId: proto.String("K-1"), BundleSha: proto.String("s1"), Ran: proto.Bool(true),
		Rerun: recordpb.Rerun_RERUN_REPRODUCED.Enum(), Soundness: recordpb.Soundness_SOUNDNESS_SOUND.Enum(), Note: proto.String("read"),
		RerunSha: proto.String(rerunOf("s1")), Claim: proto.String("the sentence K-1 backs")})
	b.reproduce(evLens, "K-2", "s2", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_SOUND)
	m := loadedT(t, b.evs...)
	l := anchorChecksOf(m.Events, m.At)
	for anchor, want := range map[string]string{"C-1": "the claim C-1 backs", "K-1": "the sentence K-1 backs", "K-2": ""} {
		s := l.at(anchor)
		if s == nil || len(s.checks) != 1 {
			t.Fatalf("the ledger holds no single check of %s: %+v", anchor, s)
		}
		if got := s.checks[0].sentence; got != want {
			t.Errorf("%s's check carries the sentence %q, want %q", anchor, got, want)
		}
	}
}

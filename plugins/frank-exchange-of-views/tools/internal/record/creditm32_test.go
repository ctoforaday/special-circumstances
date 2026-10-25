package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// M32'S ANSWERING MINTS WERE DUPLICATES, AND THE OPEN GAP ANSWERS THEM. On universe-m32 the evidence
// lens raised contradiction 80 on the prime-element definition (F-d8329c9a) and answered it with
// G-f417112d. The architecture lens then read the same sentence against the same source, opened
// contradiction 100, raised it (F-1c990edb), and its answering mint was refused as a duplicate of
// G-f417112d — the first of seven such refusals across three raises, with six sittings ending owing
// the answer. Replayed in that shape, with m32's own gap text: the screen still refuses the mint and
// now names G-f417112d and the credit; the credit lands, contradiction 100 reads answered, and no
// contradiction holds the PASS.
func TestM32TheOpenGapAnswersARefusedAnsweringMint(t *testing.T) {
	const arch = "red-lens-architecture"
	const claim = "A prime element is an element that is not zero and not a unit."
	b := reopenedBoard(t, arch).dispatch(2, arch).register(arch).
		add(evLens, contradictingVerify(claim))
	run := b.seed()
	md := "# Report\n\nA paragraph that cites nothing."
	withRenderer(t, renders(&md))
	ev, ar := Identity{Run: run, SeatID: evLens}, Identity{Run: run, SeatID: arch}

	first := unansweredOf(t, run)
	if len(first) != 1 {
		t.Fatalf("want the evidence lens's one contradiction, got %+v", first)
	}
	mustAppend(t, ev, referenceFinding("F-d8329c9a", first[0].ID))
	held := raisingGap("G-f417112d", "F-d8329c9a")
	held.Problem = proto.String(`The report gives the Wikipedia definition of a prime element as only "not zero and not a unit", ` +
		"dropping the divisibility clause the source states, so read as written it makes every non-zero non-unit prime.")
	mustAppend(t, ev, held)

	reading := contradictingVerify(claim)
	reading.CopySubstitution = recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum()
	mustAppend(t, ar, reading)
	var second string
	for _, c := range unansweredOf(t, run) {
		if c.Reader == arch {
			second = c.ID
		}
	}
	if second == "" || second == first[0].ID {
		t.Fatalf("the architecture lens's reading opened no new contradiction: %+v", unansweredOf(t, run))
	}
	mustAppend(t, ar, referenceFinding("F-1c990edb", second))

	dup := raisingGap("G-0000d0b1", "F-1c990edb")
	dup.Problem = proto.String(`The report's definition of a prime element as "not zero and not a unit" drops the divisibility ` +
		"clause its Wikipedia source states, so as written every non-zero non-unit is prime.")
	_, err := Append(ar, dup)
	mustRefuse(t, err, "would answer contradiction "+second, "your finding F-1c990edb raised",
		"G-f417112d (open, minted by "+evLens, "`credit --id G-f417112d --found-by F-1c990edb`")

	mustAppend(t, ar, creditOf("G-f417112d", "F-1c990edb"))
	if left := unansweredOf(t, run); len(left) != 0 {
		t.Errorf("contradictions stand after the credit: %+v", left)
	}
	blockers, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, bl := range blockers {
		if bl.Kind == BlockerContradiction {
			t.Errorf("a contradiction holds the PASS after the credit: %+v", bl)
		}
		if strings.Contains(bl.Subject, second) {
			t.Errorf("contradiction %s still blocks: %+v", second, bl)
		}
	}
}

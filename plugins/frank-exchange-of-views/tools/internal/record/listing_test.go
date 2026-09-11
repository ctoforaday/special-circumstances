package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A LISTING SHOWS EVERY ACT, THE STRUCK ONES MARKED, IN THE PLACE THE SEAT MADE THEM: a chain of two
// corrections lists the first wording, the second, then the one that stands — before a later act
// the seat made in between, because the correction took the first act's place.
func TestListingMarksAStruckChainInItsPlace(t *testing.T) {
	corr := func(k, r string) *Event {
		return recordtest.At(t, "s", "s:correction:"+k, &recordpb.Correction{
			Corrects: proto.String(k), Replacement: proto.String(r), Why: proto.String("a word was lost")})
	}
	evs := []*Event{
		recordtest.At(t, "s", "s:log:#1", seatLog("one")),
		recordtest.At(t, "s", "s:log:#2", seatLog("later act")),
		recordtest.At(t, "s", "s:log:#1~1", seatLog("one, corrected")),
		corr("s:log:#1", "s:log:#1~1"),
		recordtest.At(t, "s", "s:log:#1~2", seatLog("one, corrected again")),
		corr("s:log:#1~1", "s:log:#1~2"),
	}
	var got []string
	for _, l := range Listing(evs) {
		if lg, ok := recordpb.BodyAs[*recordpb.Log](l.Event); ok {
			got = append(got, l.Markdown(lg.GetText()))
		}
	}
	want := []string{
		"~~one~~ (struck by s: a word was lost)",
		"~~one, corrected~~ (struck by s: a word was lost)",
		"one, corrected again",
		"later act",
	}
	if strings.Join(got, " | ") != strings.Join(want, " | ") {
		t.Errorf("Listing = %q\nwant      %q", got, want)
	}
	// With no correction on the stream, a listing is the stream.
	if n := len(Listing(evs[:2])); n != 2 {
		t.Errorf("a stream with no correction listed %d acts, want 2", n)
	}
}

// A NARROWED READ KEEPS ITS CORRECTIONS, or Live and Listing have nothing to strike with.
func TestEventsOfACorrectableTypeCarriesItsCorrections(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	k := mustAppend(t, blue, &recordpb.Position{Text: proto.String("the board is  going in")}).GetKey()
	if _, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_POSITION, k, "a word was lost"),
		&recordpb.Position{Text: proto.String("the board is clean going in")}); err != nil {
		t.Fatal(err)
	}
	evs, err := EventsOf(run, recordpb.EventType_EVENT_TYPE_POSITION)
	if err != nil {
		t.Fatal(err)
	}
	live := Live(evs)
	var texts []string
	for _, e := range live {
		if p, ok := recordpb.BodyAs[*recordpb.Position](e); ok {
			texts = append(texts, p.GetText())
		}
	}
	if len(texts) != 1 || texts[0] != "the board is clean going in" {
		t.Errorf("the positions standing on a narrowed read = %q, want only the corrected one", texts)
	}
}

package modeltier

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

func TestOfPicksTheTierBySubstring(t *testing.T) {
	for model, want := range map[string]string{
		"claude-fable-5":       "fable",
		"claude-opus-4-8":      "opus",
		"claude-sonnet-5":      "sonnet",
		"claude-haiku-4-5":     "haiku",
		"CLAUDE-OPUS-4-8":      "opus",
		"something-unheard-of": "fable",
		"":                     "fable",
	} {
		if got := Of(model); got != want {
			t.Errorf("Of(%q) = %q, want %q", model, got, want)
		}
	}
}

// Of flattens two different states onto `fable` — an unknown model and no model at all — and
// Recognized is the only thing that tells them from a real fable.
func TestRecognizedSeparatesAMeasurementFromTheFallback(t *testing.T) {
	if !Recognized("claude-fable-5") {
		t.Error("a real fable is recognized")
	}
	for _, m := range []string{"", "gpt-9", "claude-next"} {
		if Recognized(m) {
			t.Errorf("%q must not read as a recognized tier even though Of() answers fable", m)
		}
	}
}

// THE LADDER AND THE RANK ARE THE SAME LIST, which is the property that used to be an unstated
// coincidence between Order and the price table.
func TestDearerFollowsTheLadder(t *testing.T) {
	for i := 1; i < len(Order); i++ {
		if !Dearer(Order[i], Order[i-1]) {
			t.Errorf("%s must be dearer than %s", Order[i], Order[i-1])
		}
	}
	if Dearer("opus", "opus") {
		t.Error("a tier is not dearer than itself")
	}
	// A tier off the ladder ranks dearest, matching Of's over-report-never-under-report fallback.
	if !Dearer("unheard-of", "fable") {
		t.Error("an unknown tier must rank dearest")
	}
}

// THE LOOKUP RESOLVES A NUMBERED LANE THROUGH THE LANE BASE, and a seat the map does not name is not
// configured rather than given another seat's spec.
func TestLookupSeatResolvesALaneThroughItsBase(t *testing.T) {
	seats := map[string]SeatSpec{
		"blue-lane":          {Harness: "claude", Model: "claude-sonnet-5"},
		"red-lens-dark-side": {Harness: "antigravity", Model: "gpt-oss-120b-medium"},
	}
	lane, ok := LookupSeat(seats, "blue-lane-2")
	if !ok || lane.Harness != "claude" || lane.Model != "claude-sonnet-5" {
		t.Fatalf("lane lookup got (%+v, %v)", lane, ok)
	}
	ds, ok := LookupSeat(seats, "red-lens-dark-side")
	if !ok || ds.Harness != "antigravity" || ds.Model != "gpt-oss-120b-medium" {
		t.Fatalf("dark-side lookup got (%+v, %v)", ds, ok)
	}
	if got, ok := LookupSeat(seats, "red-chair"); ok {
		t.Errorf("an unnamed seat resolved to %+v", got)
	}
}

// THE TWO CLASS SPECS EXPAND TO EVERY CONFIGURABLE SEAT, each on its class's spec, and a zero spec
// configures no seat of its class.
func TestTieredSeatsPutsEachSeatOnItsClassSpec(t *testing.T) {
	bulk, judgment := SeatSpec{Harness: "claude", Model: "sonnet"}, SeatSpec{Harness: "claude", Model: "opus"}
	seats := TieredSeats(bulk, judgment)
	if len(seats) != len(Seats()) {
		t.Fatalf("expanded %d seats, want every configurable seat (%d)", len(seats), len(Seats()))
	}
	for id, spec := range seats {
		want := bulk
		if seatclass.ClassOf(baseOf(id)) == "judgment" {
			want = judgment
		}
		if spec != want {
			t.Errorf("%s: %+v, want %+v", id, spec, want)
		}
	}
	if seats["red-chair"] != judgment || seats["blue-lane"] != bulk {
		t.Errorf("red-chair %+v, blue-lane %+v", seats["red-chair"], seats["blue-lane"])
	}
	for id := range TieredSeats(bulk, SeatSpec{}) {
		if seatclass.ClassOf(baseOf(id)) == "judgment" {
			t.Errorf("a zero judgment spec configured %s", id)
		}
	}
}

func TestSameModel(t *testing.T) {
	if !SameModel("sonnet", "claude-sonnet-5-5") {
		t.Error("recognized Anthropic tier should match")
	}
	if SameModel("sonnet", "claude-opus-5-5") {
		t.Error("different Anthropic tiers must not match")
	}
	if !SameModel("gpt-oss-120b-medium", "GPT-OSS-120B-MEDIUM") {
		t.Error("non-Anthropic models should match case-insensitively")
	}
	if SameModel("gpt-oss-120b-medium", "gemini-4-argon-high") {
		t.Error("different non-Anthropic models must not match")
	}
	if SameModel("gpt-oss-120b-medium", "claude-fable-5") {
		t.Error("non-Anthropic model must not match fable tier")
	}
}

package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
)

func TestRoleOf(t *testing.T) {
	cases := map[string]string{
		// A live seat id carries its AREA, which is the lens's identity.
		"red-lens-adversary": "adversary",
		"red-lens-evidence":  "evidence",
		"red-lens-dark-side": "dark-side",
		// THE NUMERIC FORM IS NOT READ LIVE. Records that carry it are migrated to the area form
		// (migrate/remap.go, plans/roundless.md §III.A.5); a live seat id shaped that way is no lens.
		"red-lens-L1": "",
		"red-lens-L2": "",
		"red-chair":   "", // no lens role
		"blue-lane-1": "",
		"judge":       "",
	}
	for seat, want := range cases {
		if got := AreaOf(seat); got != want {
			t.Errorf("AreaOf(%q) = %q, want %q", seat, got, want)
		}
	}
}

// AN ANCHOR IS PLACED FOR AN ID AN ANCHOR KIND CLAIMS, AND FOR NO OTHER. An id of no kind spells a
// token no reader recognises, so its Anchor would put a stray comment in the report while its gap
// read `gone`; the write refuses it, and the empty id with it.
func TestAnAnchorForAnIDOfNoKindIsRefusedAtTheWrite(t *testing.T) {
	run := mustRun(t, newRun(t))
	for _, id := range []string{"", "G1", "f-00a1b2c3", "Q-00a1b2c3"} {
		_, err := Append(Identity{Run: run, SeatID: "red-lens-logic"}, &recordpb.Anchor{Id: proto.String(id), Location: proto.String("Water is wet.")})
		if err == nil || !strings.Contains(err.Error(), "is none of them") {
			t.Errorf("Anchor{id: %q} = %v, want the refusal naming the kinds an anchor stands for", id, err)
		}
	}
}

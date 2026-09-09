package catalogue

import (
	"strings"
	"testing"
)

// SEQ IS THE ORDER THE AGENT DID THINGS IN, and it must be the SAME order every time the same
// bytes are projected.
//
// Acts were assembled by ranging the tool_use map, and a Go map has no order — so `seq` held a
// fresh shuffle on every projection. Nothing about the result looked wrong: the values were dense,
// unique, and every count over them was correct. The only visible symptom was that projecting one
// file twice produced two different stores, which no count can see.
//
// Two assertions, because either alone is too weak. Stability alone would pass a projection that
// is consistently in the wrong order; matching the file's order alone would pass by luck on a run
// where the map happened to agree.
func TestActSeqFollowsTheFileAndIsStable(t *testing.T) {
	transcript := strings.Join([]string{
		line("u1", "", "user", `[{"type":"text","text":"go"}]`),
		line("a1", "u1", "assistant", `[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/a"}},`+
			`{"type":"tool_use","id":"t2","name":"Edit","input":{"file_path":"/b"}}]`),
		line("a2", "u1", "assistant", `[{"type":"tool_use","id":"t3","name":"Bash","input":{"command":"go test"}}]`),
		line("a3", "u1", "assistant", `[{"type":"tool_use","id":"t4","name":"Write","input":{"file_path":"/c"}},`+
			`{"type":"tool_use","id":"t5","name":"Grep","input":{"pattern":"x"}}]`),
	}, "\n")

	want := []string{"Read", "Edit", "Bash", "Write", "Grep"}
	// Repeated, because a map-order bug reproduces on most runs but not all: one pass could agree
	// with the file by chance. Go re-seeds map iteration per range, so a handful of passes makes
	// an accidental agreement vanishingly unlikely.
	for pass := 0; pass < 20; pass++ {
		p := Project(strings.NewReader(transcript), 0)
		if len(p.Acts) != len(want) {
			t.Fatalf("pass %d: %d acts, want %d", pass, len(p.Acts), len(want))
		}
		for i, a := range p.Acts {
			if a.Seq != i {
				t.Fatalf("pass %d: act %d carries seq %d — seq must be the position", pass, i, a.Seq)
			}
			if a.Tool != want[i] {
				t.Fatalf("pass %d: seq %d is %s, want %s (acts are in the file's order, not a map's)",
					pass, i, a.Tool, want[i])
			}
		}
	}
}

// startSeq continues the numbering, which is what makes an appended file resumable: the acts read
// on the second pass must not reuse the numbers the first pass handed out.
func TestActSeqContinuesFromStartSeq(t *testing.T) {
	transcript := line("a1", "", "assistant",
		`[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/a"}}]`)
	p := Project(strings.NewReader(transcript), 41)
	if len(p.Acts) != 1 || p.Acts[0].Seq != 41 {
		t.Fatalf("acts = %+v, want one act at seq 41", p.Acts)
	}
}

func line(uuid, parent, typ, content string) string {
	return `{"uuid":"` + uuid + `","parentUuid":"` + parent + `","sessionId":"S",` +
		`"type":"` + typ + `","timestamp":"2026-09-09T12:00:00Z",` +
		`"message":{"role":"` + typ + `","content":` + content + `}}`
}

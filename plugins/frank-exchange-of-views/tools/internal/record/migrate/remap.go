package migrate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// remap is the STATEFUL half of the translation (plans/roundless.md §III.A.5). The registry
// translates one event at a time and cannot know what the run has already minted; this can, and
// it is what turns an archived run's identifiers into the ones a live reader speaks:
//
//   - a seat id loses its epoch and a numbered lens becomes its area: `red-merge-r2` is
//     `red-chair` (its second register is its second sitting, which opens epoch 2);
//     `red-lens-r1-L5` is `red-lens-logic`. Two citation instances in one epoch (`-L1`, `-L2`)
//     were two sittings of the evidence lens, and that is what they become.
//   - a gap id `R<r>-<n>` becomes `G<k>`, k its position in the run's mint order, and every
//     later reference to it — close, closing, regrade, motion, answers, successor, supersedes,
//     spot-check ids, manifest rows — says the new name.
//   - a finding label `L<k>-F<n>` becomes `<area>-F<m>`, m the area's own count so far, because
//     collapsing L1 and L2 into one area would otherwise mint two `evidence-F1`s.
//
// Nothing here reads the record back: the maps are the state, filled in replay order, so two
// migrations of one source produce byte-identical records (the archive test holds that).
type remap struct {
	gaps    map[string]string // R<r>-<n> -> G<k>
	labels  map[string]string // L<k>-F<n> -> <area>-F<m>
	perArea map[string]int    // findings labelled so far, per area
}

func newRemap() *remap {
	return &remap{gaps: map[string]string{}, labels: map[string]string{}, perArea: map[string]int{}}
}

var (
	numberedLens = regexp.MustCompile(`^red-lens-r\d+-L(\d+)$`)
	namedLens    = regexp.MustCompile(`^red-lens-r\d+-([a-z]+(?:-[a-z]+)*)$`)
	roundedSeat  = regexp.MustCompile(`^(red-merge|red-chair|blue-respond|judge)-r\d+$`)
	oldGapID     = regexp.MustCompile(`^R\d+-\d+$`)
	oldLabel     = regexp.MustCompile(`^L\d+-F\d+$`)
)

// lensAreaOfNumber is the roster of the runs in run-archive/ (2026-08-22 .. 2026-09-02, the W2i
// era): roles 1-4 were instances of the citation lens, 5 logic, 6 dark-side, 7 report voice —
// debate.js at 673b7b2b^ line 858-860. A number outside it is a refusal, not a guess.
func lensAreaOfNumber(n int) (string, bool) {
	switch {
	case n >= 1 && n <= 4:
		return "evidence", true
	case n == 5:
		return "logic", true
	case n == 6:
		return "dark-side", true
	case n == 7:
		return "voice", true
	}
	return "", false
}

// seat translates an archived seat id to its roundless form; an id already in that form passes.
func (r *remap) seat(old string) (string, error) {
	if m := numberedLens.FindStringSubmatch(old); m != nil {
		n, _ := strconv.Atoi(m[1])
		area, ok := lensAreaOfNumber(n)
		if !ok {
			return "", fmt.Errorf("migrate: seat %q names lens number %d, which no archived roster had — the number-to-area table in remap.go stops at 7", old, n)
		}
		return "red-lens-" + area, nil
	}
	if m := namedLens.FindStringSubmatch(old); m != nil {
		return "red-lens-" + m[1], nil
	}
	if m := roundedSeat.FindStringSubmatch(old); m != nil {
		if m[1] == "red-merge" {
			return "red-chair", nil
		}
		return m[1], nil
	}
	if petitioner, ok := strings.CutPrefix(old, "judge-petition-"); ok {
		p, err := r.seat(petitioner)
		if err != nil {
			return "", err
		}
		return "judge-petition-" + p, nil
	}
	return old, nil
}

// apply rewrites the id-bearing fields of one translated body, in place. seatID is the body's
// (already translated) seat, which is where a finding's area comes from.
func (r *remap) apply(body proto.Message, seatID string) {
	r.walk(body.ProtoReflect(), seatID)
}

func (r *remap) walk(m protoreflect.Message, seatID string) {
	isMint := m.Descriptor().Name() == "Mint"
	isFinding := m.Descriptor().Name() == "Finding"
	fields := m.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if !m.Has(fd) {
			continue
		}
		switch {
		case fd.Kind() == protoreflect.MessageKind && !fd.IsList() && !fd.IsMap():
			r.walk(m.Mutable(fd).Message(), seatID)
		case fd.Kind() == protoreflect.StringKind && fd.IsList():
			list := m.Mutable(fd).List()
			for j := 0; j < list.Len(); j++ {
				list.Set(j, protoreflect.ValueOfString(r.ref(string(fd.Name()), list.Get(j).String())))
			}
		case fd.Kind() == protoreflect.StringKind:
			name, v := string(fd.Name()), m.Get(fd).String()
			switch {
			case isMint && name == "gap_id" && oldGapID.MatchString(v):
				nv := fmt.Sprintf("G%d", len(r.gaps)+1)
				r.gaps[v] = nv
				m.Set(fd, protoreflect.ValueOfString(nv))
			case isFinding && name == "label" && oldLabel.MatchString(v):
				area := strings.TrimPrefix(seatID, "red-lens-")
				r.perArea[area]++
				nv := fmt.Sprintf("%s-F%d", area, r.perArea[area])
				r.labels[v] = nv
				m.Set(fd, protoreflect.ValueOfString(nv))
			default:
				if nv := r.ref(name, v); nv != v {
					m.Set(fd, protoreflect.ValueOfString(nv))
				}
			}
		}
	}
}

// ref maps a REFERENCE by the field it sits in. A reference to an id the run never minted is
// left as it was: the write path decides what a dangling reference is worth, and it says so.
func (r *remap) ref(field, v string) string {
	switch field {
	case "gap_id", "answers", "successor", "superseded_by", "supersedes", "ids":
		if nv, ok := r.gaps[v]; ok {
			return nv
		}
	case "found_by", "label":
		if nv, ok := r.labels[v]; ok {
			return nv
		}
	}
	return v
}

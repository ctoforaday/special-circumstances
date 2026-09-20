package record

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN OCCASION IS WHAT A SITTING WAS CONVENED TO DO, and it is recorded exactly where the seat id
// does not already say.
//
// For every seat but the bench, the id determines the answer: a lens audits, a lane researches, the
// chair runs the epoch. Recording it there would store a derivation the record already holds. The
// bench is the one seat whose four sittings ask four different questions under one id, so the bench
// is the one seat that owes this — see the Occasion enum in record.proto for what the collapse
// removed and why.
//
// THE REFUSAL RUNS BOTH WAYS, and the second half is what makes the field a fact rather than a
// hint. A bench register with no occasion is refused; any other seat supplying one is refused. With
// only the first half an occasion could be asserted anywhere and would mean nothing, and a reader
// could not tell a seat that answered the question from one that was never asked it.

// benchRole is the role whose seat id does not determine its occasion. It keys roleSeats, so the
// word has one spelling: a role renamed there and not here would silently stop owing an occasion,
// which is precisely the miss this field exists to make impossible.
const benchRole = "bench"

// SeatOwesOccasion reports whether a seat must say what its sitting was convened to do.
func SeatOwesOccasion(seatID string) bool { return roleOfSeat(seatID) == benchRole }

// OccasionOf resolves the word a seat types into the schema's value. It refuses rather than
// returning the zero, for the reason every resolver in enumof.go refuses: UNSPECIFIED is reserved
// for "the seat never said", so recording it for a typo would turn a misspelling into a field that
// reads downstream exactly like a question nobody asked.
func OccasionOf(word string) (recordpb.Occasion, bool) {
	return enumOf[recordpb.Occasion](recordpb.Occasion(0).Descriptor(), word)
}

// OccasionWords is the vocabulary in schema order, skipping the UNSPECIFIED zero — derived from the
// descriptor, so the words a refusal offers are the words the write path accepts.
func OccasionWords() []string {
	d := recordpb.Occasion(0).Descriptor().Values()
	out := make([]string, 0, d.Len())
	for i := 0; i < d.Len(); i++ {
		v := d.Get(i)
		if v.Number() == 0 {
			continue
		}
		out = append(out, recordpb.Spelling(v))
	}
	return out
}

// checkOccasion holds a register's occasion against the seat making it, and returns the resolved
// value to write. An empty word from a seat that owes nothing is the ordinary case and writes
// nothing.
//
// The messages name the seat and the whole vocabulary, because the seat reading them is mid-sitting
// and the alternative is another round trip to `--help` for a four-word list.
func checkOccasion(seatID, word string) (*recordpb.Occasion, error) {
	word = strings.TrimSpace(word)
	owes := SeatOwesOccasion(seatID)
	if word == "" {
		if owes {
			return nil, feov.Errorf(feov.MissingField,
				"record: %s registers with no --occasion. The bench is one seat asked four different questions, so its id cannot say which sitting this is — name it: %s",
				seatID, strings.Join(OccasionWords(), " | "))
		}
		return nil, nil
	}
	if !owes {
		return nil, feov.Errorf(feov.RoleViolation,
			"record: %s passed --occasion, and only the bench carries one. Every other seat's id already says what its sitting is for, so an occasion here would be a second answer to a question the record has already settled",
			seatID)
	}
	occ, ok := OccasionOf(word)
	if !ok {
		return nil, feov.Errorf(feov.Validation,
			"record: %q is not an occasion. The bench sits for one of: %s",
			word, strings.Join(OccasionWords(), " | "))
	}
	return &occ, nil
}

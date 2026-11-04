package terms

// THE DEFINITIONS, AS GO LITERALS. Two readers need only what a term MEANS and which seats meet
// it: an enum value's `(defined_term)` (recordpb.EnumValueDoc, reached from the record package's
// own initialisation in every process) and the words `manual` prints. Neither needs the bans, so
// neither pays for decoding terms.json, validating it and compiling its patterns; that is the
// vocabulary gate's work, and Load stays its entry.
//
// definitions_gen.go is written from Load by
// `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen`; vocabgen -check and
// TestTheDefinitionTableIsTheRegistry fail a stale one.

// Definition is one entry's term, the sentence that defines it, and the seats that meet it.
type Definition struct {
	Term       string
	Definition string
	Seats      []string
}

// Defined returns the registry's definition of a term.
func Defined(term string) (string, bool) {
	for i := range definitions {
		if definitions[i].Term == term {
			return definitions[i].Definition, true
		}
	}
	return "", false
}

// DefinitionsFor returns the definitions a seat's surface delivers, in registry order.
func DefinitionsFor(role string) []Definition {
	var out []Definition
	for _, d := range definitions {
		for _, s := range d.Seats {
			if s == role {
				out = append(out, d)
				break
			}
		}
	}
	return out
}

// DefinitionOf is an entry's row of the generated table.
func DefinitionOf(e Entry) Definition {
	return Definition{Term: e.Term, Definition: e.Definition, Seats: e.Seats}
}

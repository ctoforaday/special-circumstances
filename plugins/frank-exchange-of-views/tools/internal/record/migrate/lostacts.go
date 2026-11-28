package migrate

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// AN ACT REFUSED ONLY BECAUSE IT DEPENDS ON A LOST ACT IS THAT SAME LOSS (gblock 2026-11-14). A close
// the delivered-text rule drops leaves its gap open, so a later act that stood on the gap closing —
// a mint superseding it, an inquiry's check waiting on it — is refused by the write, and an act on a
// gap whose mint was lost names a gap the record never created. Such an act is written as no event
// and listed with the act it depended on. The dependency is read off the ids the act's body names
// and the gaps about them, never off the refusal's wording; an act refused that names no lost gap
// stays a refusal.

// lostGaps is each gap a loss touched — its close dropped, its mint lost, or the ruling on the
// docket motion about it dropped — and the source act that was lost.
type lostGaps map[string]lostAct

// lostAct is the lost source act a gap's loss came from.
type lostAct struct {
	oldID int64
	word  string
}

// noteLoss records what a lost act takes with it: the gap a close closed, the gap a mint created,
// or the gaps a ruling's docket motion is about — a lost ruling leaves them open, as a lost close
// does.
func (l lostGaps) noteLoss(q recordsql.Querier, body proto.Message, oldID int64, word string) error {
	switch b := body.(type) {
	case *recordpb.Close:
		l[b.GetGapId()] = lostAct{oldID, word}
	case *recordpb.Mint:
		l[b.GetGapId()] = lostAct{oldID, word}
	case *recordpb.MotionRule:
		rows, err := q.Query(`SELECT d."gap_id" FROM "motion" m JOIN "motion_docket" d ON d."event_id" = m."event_id" WHERE m."motion_id" = ?`, b.GetMotionId())
		if err != nil {
			return fmt.Errorf("asking which gaps motion %s docketed: %w", b.GetMotionId(), err)
		}
		defer rows.Close()
		for rows.Next() {
			var g string
			if err := rows.Scan(&g); err != nil {
				return err
			}
			l[g] = lostAct{oldID, word}
		}
		return rows.Err()
	}
	return nil
}

// dependsOn is the lost act a refused body depends on: a lost gap it names, a lost gap about an id it
// names (an inquiry's gaps), or the lost gap a docketing it names put before the bench (a ruling of a
// docketing whose ending close was lost). ok is false where it names none.
func (l lostGaps) dependsOn(q recordsql.Querier, body proto.Message) (string, lostAct, bool, error) {
	if len(l) == 0 {
		return "", lostAct{}, false, nil
	}
	for _, id := range namedIDs(body) {
		if a, lost := l[id]; lost {
			return id, a, true, nil
		}
		rows, err := q.Query(`SELECT "gap_id" FROM "mint" WHERE "about_ref" = ?1 UNION SELECT "gap_id" FROM "docketing" WHERE "docketing_id" = ?1`, id)
		if err != nil {
			return "", lostAct{}, false, fmt.Errorf("asking which gaps are about %s or put before the bench by it: %w", id, err)
		}
		var about []string
		for rows.Next() {
			var g string
			if err := rows.Scan(&g); err != nil {
				rows.Close()
				return "", lostAct{}, false, err
			}
			about = append(about, g)
		}
		rows.Close()
		for _, g := range about {
			if a, lost := l[g]; lost {
				return g, a, true, nil
			}
		}
	}
	return "", lostAct{}, false, nil
}

// namedIDs is every string a body's fields hold, singular or repeated, nested messages included:
// the ids it names are among them, and a string that is no id matches no lost gap.
func namedIDs(body proto.Message) []string {
	var out []string
	var walk func(m protoreflect.Message)
	walk = func(m protoreflect.Message) {
		m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			switch {
			case fd.IsList():
				l := v.List()
				for i := 0; i < l.Len(); i++ {
					switch fd.Kind() {
					case protoreflect.StringKind:
						out = append(out, l.Get(i).String())
					case protoreflect.MessageKind:
						walk(l.Get(i).Message())
					}
				}
			case fd.IsMap():
			case fd.Kind() == protoreflect.StringKind:
				out = append(out, v.String())
			case fd.Kind() == protoreflect.MessageKind:
				walk(v.Message())
			}
			return true
		})
	}
	walk(body.ProtoReflect())
	return out
}

// dependentLoss is the reason a refused act that depends on a lost act is listed as lost.
func dependentLoss(gap string, a lostAct, refusal error) string {
	return fmt.Sprintf("this act depends on gap %s, which the lost %s (source event %d) took with it, and the write refuses it: %v", gap, a.word, a.oldID, refusal)
}

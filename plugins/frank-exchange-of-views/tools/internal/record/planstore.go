package record

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WithKey is p carrying its key: the first sixteen hex of the sha256 of p as printed, with the key
// itself empty and every field the declaration leaves out of the key (PlanField.Unkeyed) at its
// zero. Those fields say when the plan was asked, or what the sitting asking did — the epoch moves at every chair sitting — so a
// key over them would never repeat, and the engine's no-progress valve, which compares keys, would
// never stop a run whose plan is otherwise the same sitting after sitting.
func (p Plan) WithKey() Plan {
	q := p
	q.Key = ""
	v := reflect.ValueOf(&q).Elem()
	for _, f := range PlanFields() {
		if f.Unkeyed {
			fv := v.FieldByName(f.Name)
			fv.Set(reflect.Zero(fv.Type()))
		}
	}
	b, err := json.Marshal(q)
	if err != nil {
		// Unreachable for a struct of strings, numbers, booleans and lists of them.
		panic(fmt.Sprintf("the plan does not marshal: %v", err))
	}
	sum := sha256.Sum256(b)
	p.Key = hex.EncodeToString(sum[:8])
	return p
}

// Printed is the plan as `dispatch next` prints it and the record stores it.
func (p Plan) Printed() string {
	b, err := json.Marshal(p)
	if err != nil {
		panic(fmt.Sprintf("the plan does not marshal: %v", err)) // as WithKey
	}
	return string(b)
}

// ReadPlan reads raw as a plan — every key one record.Plan declares, every value of its declared
// type — and returns it with its printing. A relayed plan read this way prints as the stored one
// exactly when it is the same plan: the key order, the spacing and a lens's `occasions` (absent and
// empty are one state) are the encoding's, never the plan's.
func ReadPlan(raw []byte) (Plan, string, error) {
	var p Plan
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return Plan{}, "", err
	}
	if dec.More() {
		return Plan{}, "", fmt.Errorf("more than one value")
	}
	return p, p.Printed(), nil
}

// requirePrintedPlan refuses, at the write, a stored plan that is not a plan as the verb prints it:
// a key record.Plan does not declare, any byte its own printing would not write, a key that is not
// the plan's, or a plan its declaration's rules refuse (relayable). The relay is compared with these
// bytes, so bytes no printing writes would fail every faithful relay.
func requirePrintedPlan(printed string) error {
	p, again, err := ReadPlan([]byte(printed))
	if err != nil {
		return fmt.Errorf("the stored plan is not a plan: %v", err)
	}
	if again != printed {
		return fmt.Errorf("the stored plan is not as the verb prints it — it is stored as printed, so a relay of it is compared byte for byte")
	}
	if want := p.WithKey().Key; p.Key != want {
		return fmt.Errorf("the stored plan's key is %q and the plan's own is %q", p.Key, want)
	}
	return relayable(p)
}

// StoredPlan is one plan the record stored: its place in the stream, the plan, and its printing.
type StoredPlan struct {
	At      int
	Plan    Plan
	Printed string
}

// StoredPlans is every plan the record stored, in stream order. The write refused every one that
// does not read as a plan, so a refusal here is a record written by something other than the write.
func StoredPlans(evs []*Event) ([]StoredPlan, error) {
	var out []StoredPlan
	for i, e := range evs {
		b, ok := recordpb.BodyAs[*recordpb.Plan](e)
		if !ok {
			continue
		}
		p, _, err := ReadPlan([]byte(b.GetPrinted()))
		if err != nil {
			return nil, fmt.Errorf("the plan stored at position %d of the stream does not read as a plan: %w", i, err)
		}
		out = append(out, StoredPlan{At: i, Plan: p, Printed: b.GetPrinted()})
	}
	return out, nil
}

// standingPlan is the record's standing plan, as printed: the plan stored last, with no sitting
// opened since — no bracket, no register opening one, by any seat. None stands from the moment a
// sitting opens until that sitting's chair asks.
func standingPlan(evs []*Event, win WindowIndex) (printed string, stands bool) {
	for i := len(evs) - 1; i >= 0; i-- {
		if _, opens := win.Opens(evs[i]); opens {
			return "", false
		}
		if b, ok := recordpb.BodyAs[*recordpb.Plan](evs[i]); ok {
			return b.GetPrinted(), true
		}
	}
	return "", false
}

// planStands reports whether printed is the record's standing plan. Asked twice in one sitting, for
// the prose and then for --json, it is one decision, and the B5 and B6 chairs each wrote it twice.
func planStands(evs []*Event, win WindowIndex, printed string) bool {
	standing, stands := standingPlan(evs, win)
	return stands && standing == printed
}

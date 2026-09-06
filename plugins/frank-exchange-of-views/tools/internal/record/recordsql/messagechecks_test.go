package recordsql

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EVERY TABLE-LEVEL RULE ANYONE DECLARES REACHES THE DATABASE, wherever the message sits.
//
// A message-level `option (check)` is how this schema states a rule spanning two fields —
// "repaired_with_regression requires a successor" — which no annotation on either field can say.
// It reached the DDL through tableFor, which runs on the Event body messages. It did NOT reach it
// through armTable, which builds the table for a oneof's MESSAGE arm: that path emitted columns,
// field checks and foreign keys and never asked for the message's own rules.
//
// SO AN AUTHORED CONSTRAINT WOULD HAVE GENERATED NOTHING. The DDL applies cleanly, the schema is
// valid, and the rule is simply absent — no error, no warning, nothing to grep for. It was latent
// when it was found (no arm had declared one yet), and the first arm to need a rule would have
// been the first to silently not have it.
//
// THIS TEST IS WRITTEN OVER THE CONCEPT, NOT THE SHAPE, so it does not go vacuous. It walks every
// message the schema reaches — bodies and arms alike — collects what each DECLARES, and asserts
// the DDL contains it. `DocketRuling` — an arm of `MotionRule.ruling` — declares TWO (the
// reopens_on/final pair), so the arm half is no longer hypothetical and this measures the exact
// case the bug dropped. It was written while the arms declared none, which is the point: the
// concept-level walk covered the first arm to need a rule without anybody remembering this file
// exists. The floor below is what makes "found nothing" a failure rather than a pass.
func TestEveryDeclaredMessageCheckReachesTheDDL(t *testing.T) {
	ddl, err := Schema()
	if err != nil {
		t.Fatal(err)
	}

	bodies, err := Bodies()
	if err != nil {
		t.Fatal(err)
	}

	type declared struct{ owner, expr string }
	var all []declared
	for _, md := range bodies {
		for _, c := range MessageChecks(md) {
			all = append(all, declared{string(md.Name()), c.GetExpr()})
		}
		for _, arm := range armMessages(md) {
			for _, c := range MessageChecks(arm) {
				all = append(all, declared{string(md.Name()) + "/" + string(arm.Name()), c.GetExpr()})
			}
		}
	}

	// THE FLOOR. An empty walk and a fully-satisfied one print the same result, so the count is
	// asserted rather than assumed — the failure this whole test is about is a rule that generates
	// nothing while everything looks green.
	if len(all) == 0 {
		t.Fatal("walked the schema and found no message-level checks at all; this test would pass " +
			"vacuously, which is the exact shape it exists to catch")
	}
	t.Logf("%d declared message-level check(s) across bodies and arms", len(all))

	for _, d := range all {
		if !strings.Contains(ddl, d.expr) {
			t.Errorf("%s declares the table-level rule %q and the generated DDL does not carry it — "+
				"the schema applies cleanly and the constraint is absent, which is indistinguishable "+
				"from a schema that never needed it", d.owner, d.expr)
		}
	}
}

// armMessages lists the message arms of a message's real oneofs — the ones that get a table of
// their own through armTable. Synthetic oneofs are proto3 `optional` under the hood and are not
// groups; conflating them would treat every optional field as a mutually-exclusive arm.
func armMessages(md protoreflect.MessageDescriptor) []protoreflect.MessageDescriptor {
	var out []protoreflect.MessageDescriptor
	for i := 0; i < md.Oneofs().Len(); i++ {
		od := md.Oneofs().Get(i)
		if od.IsSynthetic() {
			continue
		}
		for j := 0; j < od.Fields().Len(); j++ {
			if f := od.Fields().Get(j); f.Kind() == protoreflect.MessageKind {
				out = append(out, f.Message())
			}
		}
	}
	return out
}

// AND armTable ITSELF EMITS THEM, proven against a descriptor that really declares two.
//
// The test above walks what the schema declares today. When both tests were written no oneof arm
// declared a message-level rule, so that one would have passed with armTable still dropping them
// — this one closes the gap by driving armTable DIRECTLY with a message that does declare them:
// `Close`, an arm of Event's `body` oneof, carrying the regression-successor rule and the
// closure-argument rule.
//
// `DocketRuling` has since given the walk a real arm to find, so the two now overlap. They are
// still not the same test: this one holds when the schema happens to have no such arm, which is
// the state the defect was found in and the state a later deletion could restore.
//
// It is not a shape assertion. Delete the MessageChecks call from armTable and this fails; that
// is the whole contract, and it is the reason the fix is not merely "one more line in a builder".
func TestArmTableCarriesTheArmsOwnMessageChecks(t *testing.T) {
	ev := (&recordpb.Event{}).ProtoReflect().Descriptor()
	body := ev.Oneofs().ByName("body")
	if body == nil {
		t.Fatal("Event has no `body` oneof; the schema's whole derivation rests on it")
	}

	var withChecks protoreflect.FieldDescriptor
	var want []string
	for i := 0; i < body.Fields().Len(); i++ {
		f := body.Fields().Get(i)
		if cs := MessageChecks(f.Message()); len(cs) > 0 {
			withChecks = f
			for _, c := range cs {
				want = append(want, c.GetExpr())
			}
			break
		}
	}
	if withChecks == nil {
		t.Skip("no body arm declares a message-level check; nothing to prove armTable against")
	}

	ddl, err := armTable(ev, withChecks)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("armTable(%s) — asserting %d check(s)", withChecks.Name(), len(want))
	for _, expr := range want {
		if !strings.Contains(ddl, "CHECK ("+expr+")") {
			t.Errorf("armTable dropped %s's table-level rule %q:\n%s", withChecks.Message().Name(), expr, ddl)
		}
	}
}

// A MIXED ONEOF IS STILL EXCLUSIVE, AND THE DDL HAS TO SAY SO.
//
// The generated exclusivity check summed the SCALAR arms alone. While every arm of every oneof was
// a scalar that was the whole statement; `MotionRule.ruling` gained a MESSAGE arm and the sentence
// silently narrowed to three arms of four — a row naming a scalar ruling AND carrying a docket
// child satisfied it. Nothing broke, and nothing could: the constraint went on reading like a
// complete statement of exclusivity.
//
// This drives the DDL rather than the generator, because the generator's output is only a claim
// until SQLite refuses the row.
func TestAMixedOneofRefusesTwoArmsAtOnce(t *testing.T) {
	ddl, err := Schema()
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("the schema does not apply: %v", err)
	}
	// NO PARENT ROWS AND NO PRAGMA. SQLite leaves foreign keys OFF unless asked, so `motion_rule`
	// accepts an event_id nothing backs — which is what this test wants: the only thing that can
	// refuse either row below is the CHECK under test, not a reference that happened to be missing.

	// One scalar arm alone is legal — the anti-vacuity half. Without it a constraint that refused
	// EVERYTHING would satisfy the assertion below.
	if _, err := db.Exec(`INSERT INTO "motion_rule" ("event_id","motion_id","subject","opinion","grade")
	  VALUES (1,'M1','grade','because','accepted')`); err != nil {
		t.Fatalf("a single scalar arm must be accepted: %v", err)
	}
	// The message arm alone is legal too, and it is the arm this change added.
	if _, err := db.Exec(`INSERT INTO "motion_rule" ("event_id","motion_id","subject","opinion","ruling_case")
	  VALUES (2,'M2','docket','because','docket')`); err != nil {
		t.Fatalf("the message arm alone must be accepted: %v", err)
	}

	// A scalar arm AND the message arm's discriminator, which is two arms of one oneof.
	_, err = db.Exec(`INSERT INTO "motion_rule" ("event_id","motion_id","subject","opinion","grade","ruling_case")
	  VALUES (3,'M3','grade','because','accepted','docket')`)
	if err == nil {
		t.Fatal("a motion_rule row set BOTH a grade ruling and a docket ruling was accepted — " +
			"the oneof's exclusivity check covers the scalar arms only, so the message arm is outside it")
	}
	if !strings.Contains(err.Error(), "CHECK") && !strings.Contains(err.Error(), "constraint") {
		t.Errorf("refused, but not by the constraint under test: %v", err)
	}
}

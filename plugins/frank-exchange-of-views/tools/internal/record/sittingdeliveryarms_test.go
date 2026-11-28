package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// SeededSittingDeliveryArm is one seeded record that holds an arm of the sitting and delivery
// questions, named for the arm.
type SeededSittingDeliveryArm struct {
	Name string
	Run  Run
}

// SeededSittingDeliveryArms is a record for each arm of the sitting and delivery questions, the ones
// no archived run reaches among them: a failed sitting, a sitting that never closed, a stop that
// lands late, a delivery in each, a delivery in no sitting, an upheld position naming a lens, a
// closing duty a sitting's end decides, and a spot-check's stamp. Each is a few acts on a small
// board, so every line its answers render is checkable against the acts by hand
// (TestEverySeededSittingAndDeliveryArmAnswersAsItsGoldenSays, in the external test package, renders
// them through the calls that take a run).
//
// An id below is the act's events.id: a stage's acts are seeded in order from 1, and the settled
// question's three events follow them; the golden lists each record's acts with their ids.
func SeededSittingDeliveryArms(t *testing.T) []SeededSittingDeliveryArm {
	t.Helper()
	var out []SeededSittingDeliveryArm
	arm := func(name string, run Run) { out = append(out, SeededSittingDeliveryArm{Name: name, Run: run}) }

	// The lens mints G1 in its first sitting (5-8, the mint 9). The chair engages it and blue on G1
	// (11, 12); blue's sitting (13) closes with no envelope admitted, or with one; the lens then sits
	// through its own. A second lens is cast and never dispatched.
	exchange := func(blueAdmitted bool) Run {
		b := newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high").
			register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").
			registerAs("blue-respond", "blue-agent")
		if blueAdmitted {
			b.stop("blue-agent")
		} else {
			b.fails("blue-agent")
		}
		return b.sitClosed(evLens).register("red-chair").seed()
	}
	arm("a failed sitting: blue's sitting on G1 closes with no envelope admitted", exchange(false))
	arm("the same exchange, blue's envelope admitted", exchange(true))

	// The lens sits through a sitting that has the report at head 2 (5-8). The report moves (9). Its
	// next sitting (11) has the report at head 9 (12) and never closes: the lens opens a third (14),
	// which is open.
	{
		b := retBoard(t).sitHooked(2, evLens, "a1", true).ingest()
		b.sitHooked(b.at(), evLens, "a2", false).dispatch(b.head(), evLens).bracket(evLens, "a3")
		arm("a delivery in a sitting that never closed, and the sitting after it still open", b.register("red-chair").seed())
	}

	// The agent of the lens's first sitting (5, its delivery 6) stops (9, 10) after the lens has
	// opened its second (8).
	{
		b := retBoard(t).sitHooked(2, evLens, "a1", false).dispatch(2, evLens).bracket(evLens, "a2").stop("a1")
		arm("a stop that lands after the seat's next opening", b.register("red-chair").seed())
	}

	// The lens mints G1, never material (4), and sits three barren sittings. Blue answers G1 in a
	// sitting of its own. The lens's sitting for that answer (its fourth) never closes: it opens a
	// fifth, which is open. The report moves.
	{
		b := retBoard(t)
		b.add(evLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high"))
		for range retiresAfter {
			b.sit(2, evLens)
		}
		blueAnswersGap(b, "G1")
		b.dispatch(2, evLens, "G1").bracket(evLens, "a9").bracket(evLens, "a10").ingest()
		arm("a debt outlives a sitting that never closed", b.register("red-chair").seed())
	}

	// The lens has both whole deliveries at head 2. The chair withholds the PASS its plan permits (11,
	// 12) and the bench upholds the position, naming the lens (14). The lens's sitting after the
	// ruling (17) opens; the whole report is delivered in it (18); it sits through it (19, 20).
	ruled := func() *stage {
		b := permittingBoard(t).withhold(heardID).benchRegister(recordpb.Occasion_OCCASION_POSITION).
			add(benchSeat, positionRuling(heardID, upheld, evLens, nil, direction)).register("red-chair")
		return b.dispatch(b.head(), evLens).registerAs(evLens, "lens-audits")
	}
	arm("an upheld position naming the lens: its sitting after the ruling, the report not yet delivered", ruled().seed())
	{
		b := ruled()
		arm("an upheld position naming the lens: the whole report delivered in a sitting still open", b.served(evLens, b.head()).seed())
	}
	{
		b := ruled()
		arm("an upheld position naming the lens: the sitting that had the whole report sat through",
			b.served(evLens, b.head()).stop("lens-audits").register("red-chair").seed())
	}

	// A SITTING IS PLACED BY ITS OPENING, NOT BY ITS DELIVERY'S ROW. The lens's sitting opens (14)
	// before the bench upholds the position (16); the whole report is delivered in it after the ruling
	// (17) and the lens sits through it (18, 19). The delivery's row follows the ruling and its
	// sitting's opening precedes it.
	{
		b := permittingBoard(t).withhold(heardID)
		b.dispatch(b.head(), evLens).bracket(evLens, "opened-before").benchRegister(recordpb.Occasion_OCCASION_POSITION).
			add(benchSeat, positionRuling(heardID, upheld, evLens, nil, direction))
		arm("an upheld position naming the lens: its sitting opened before the ruling, its delivery after it",
			b.served(evLens, b.head()).stop("opened-before").register("red-chair").seed())
	}

	// The same of a dispatch: the lens's sitting opens (5), the chair dispatches it again (6), and the
	// report is delivered in the open sitting after that dispatch (7), which the lens sits through (8,
	// 9). It has opened no sitting for the second dispatch.
	{
		b := retBoard(t).dispatch(2, evLens).bracket(evLens, "a1").dispatch(2, evLens).served(evLens, 2).stop("a1")
		arm("a delivery that lands after the next dispatch, in the sitting opened before it", b.register("red-chair").seed())
	}

	// The report is at head 4 when it is delivered to the lens (8), which has opened no sitting: the
	// write path's own report_read, which it does not refuse. The lens then sits (9, 10) with no
	// delivery.
	{
		b := retBoard(t).ingest()
		run := b.seed()
		if err := RecordDelivery(run, evLens, b.head()); err != nil {
			t.Fatal(err)
		}
		arm("a delivery in no sitting", b.more().dispatch(b.head(), evLens).register(evLens).register("red-chair").seed())
	}

	// G1 is docketed by blue's motion M1 (18). The lens's sitting after it never closes (a1, then a2
	// open); blue sits through one and files no closing. Then the bench remands G1 (27, 28), and the
	// lens sits through a sitting opened after that ruling.
	docket := func() *stage {
		return docketed(t).dispatch(2, evLens).bracket(evLens, "a1").dispatch(2, evLens).bracket(evLens, "a2").sitClosed("blue-respond").register("red-chair")
	}
	arm("a docketed gap: the lens's sitting since never closed, blue sat through one and filed no closing", docket().seed())
	arm("the same docketing ruled: the lens sits through a sitting opened after the ruling",
		docket().benchRegister(recordpb.Occasion_OCCASION_DOCKET).add(benchSeat, remandRuling("M1", "state the source's own figure")).
			register("red-chair").dispatch(2, evLens).sitClosed(evLens).register("red-chair").seed())

	// The lens retires at pin 10 and the report moves to 14. The chair's sitting (15) is delivered the
	// change behind the pin and spot-checks the area, both through the write path: the write stamps
	// the pin and the change.
	{
		b, _ := retiredBehindHead(t)
		arm("a stale area: a retired lens behind the head, before the chair's spot-check", b.register("red-chair").seed())
	}
	{
		b, _ := retiredBehindHead(t)
		run := b.register("red-chair").seed()
		deliver(t, run, "red-chair")
		if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, &recordpb.SpotCheck{Areas: []string{evLens},
			Reason: proto.String("the change holds against the area's duties")}); err != nil {
			t.Fatal(err)
		}
		arm("the chair's spot-check of that area, stamped by the write", run)
	}
	return out
}

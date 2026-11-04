package record

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A RENDER OF THE REPORT AS IT STANDS IS READ FROM ONE SNAPSHOT while other seats write. RenderRun
// reads the newest checkpoint and then the ops after it; read as two statements on the run's handle,
// a write that commits between them lands the eighth op and its checkpoint after the checkpoint the
// render read, and the render refuses a record its write path wrote as one "not written by this
// binary's write path" (release fuzz seeds 5, 37, 40: a lens's `verify --url` rendered while the other
// lenses placed anchors). Each anchor here is committed through the write path that stamps the
// checkpoints, so every render must succeed.
func TestARenderOfTheReportAsItStandsReadsOneSnapshot(t *testing.T) {
	run := newStage(t).cast(evLens).ingest().seed()
	const anchors = 400
	evs := make([]*recordpb.Event, anchors)
	for k := range evs {
		evs[k] = recordtest.Event(t, evLens, &recordpb.Anchor{Id: proto.String(fmt.Sprintf("F-%d", k)), Location: proto.String("# report")})
		evs[k].Ts = proto.String("2026-01-01T00:00:00Z")
	}
	db, err := recordsql.Open(filepath.Join(run.Dir(), "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	var done atomic.Bool
	werr := make(chan error, 1)
	go func() {
		defer done.Store(true)
		for _, ev := range evs {
			if _, err := recordsql.Insert(db, ev); err != nil {
				werr <- err
				return
			}
		}
		werr <- nil
	}()
	renders := 0
	var first error
	for !done.Load() {
		renders++
		if _, err := RenderRun(run); err != nil && first == nil {
			first = err
		}
	}
	if err := <-werr; err != nil {
		t.Fatalf("writing the anchors: %v", err)
	}
	// A WRITER THAT FINISHED BEFORE THE RENDERS BEGAN MEASURED NOTHING, and says so.
	if renders < 20 {
		t.Fatalf("%d render(s) overlapped the %d writes — too few to say anything", renders, anchors)
	}
	if first != nil {
		t.Fatalf("a render made while %d anchors were written (%d renders) failed: %v", anchors, renders, first)
	}
	if _, err := RenderRun(run); err != nil {
		t.Fatalf("the report after the writes: %v", err)
	}
}

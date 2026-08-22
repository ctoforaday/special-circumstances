package record

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// FIVE TESTS ARE GONE FROM THIS FILE, and each named a mechanism the record no longer has.
//
//   - TestAppendEntersACriticalSection — the append took a critical section around a FILE write.
//   - TestAppendHealsATornFinalLine — a crash mid-append left an unterminated fragment, and
//     appending onto it would destroy the next event too. A transaction commits or it does not.
//   - TestAppendAssignsGapFreePerShardSequence — `seq` was the seat's position in its own shard;
//     nothing read it, and numbering it cost every write a query.
//   - TestRegisterRotatesTheNonceAndRepointsTheSeat — the nonce named a shard file and the pointer
//     said which was live. Both are gone; the sitting is the count of a seat's registers.
//   - TestAppendImplicitlyRegistersWhenThePointerIsAbsent — there is no pointer to be absent.
//
// What they protected is not unprotected: atomicity is the transaction (recordsql.InsertTx),
// concurrent writers are covered by TestConcurrentSeatsDoNotLoseEvents in recordsql — which spawns
// real processes, because goroutines did not reproduce the hazard — and the write path's
// registration and seat-id rules are still tested below.

// enterCritical must BLOCK once shutdown has begun: starting a write we may not
// be allowed to finish is the truncation the guard exists to prevent.
func TestEnterCriticalBlocksOnceShuttingDown(t *testing.T) {
	guardMu.Lock()
	shuttingDown = true
	guardMu.Unlock()
	t.Cleanup(func() {
		guardMu.Lock()
		shuttingDown = false
		guardCond.Broadcast()
		guardMu.Unlock()
	})

	entered := make(chan struct{})
	go func() {
		enterCritical()
		exitCritical()
		close(entered)
	}()

	select {
	case <-entered:
		t.Fatal("enterCritical proceeded during shutdown — a write could start that the exit will not wait for")
	default:
	}

	// Releasing the shutdown flag lets the blocked writer through.
	guardMu.Lock()
	shuttingDown = false
	guardCond.Broadcast()
	guardMu.Unlock()
	<-entered
}

// A blocked caller must not increment the depth, or it would delay the very exit
// it is waiting on — a deadlock at shutdown.
func TestBlockedCallerDoesNotHoldTheDepth(t *testing.T) {
	guardMu.Lock()
	shuttingDown = true
	guardMu.Unlock()
	t.Cleanup(func() {
		guardMu.Lock()
		shuttingDown = false
		guardCond.Broadcast()
		guardMu.Unlock()
	})

	started := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		close(started)
		enterCritical()
		exitCritical()
	}()
	<-started

	guardMu.Lock()
	d := guardDepth
	guardMu.Unlock()
	if d != 0 {
		t.Errorf("guardDepth = %d while a caller is BLOCKED on shutdown, want 0", d)
	}

	// Release and JOIN before returning. The depth is process-global, so a
	// goroutine still blocked in enterCritical when this test ends would enter it
	// during a later test and corrupt that test's reading.
	guardMu.Lock()
	shuttingDown = false
	guardCond.Broadcast()
	guardMu.Unlock()
	<-finished
}

// Nested critical sections must both be counted, so an outer section is not
// released by an inner one finishing.
func TestCriticalSectionsNest(t *testing.T) {
	// The depth is process-global, so this asserts on the DELTA rather than an
	// absolute value: what nesting has to guarantee is that each enter adds
	// exactly one and each exit removes exactly one, whatever the starting point.
	depth := func() int {
		guardMu.Lock()
		defer guardMu.Unlock()
		return guardDepth
	}
	base := depth()

	enterCritical()
	enterCritical()
	if got := depth() - base; got != 2 {
		t.Errorf("depth rose by %d after two enters, want 2", got)
	}
	exitCritical()
	if got := depth() - base; got != 1 {
		t.Errorf("depth is +%d after one exit of two, want 1 — the outer section was released early", got)
	}
	exitCritical()
	if got := depth() - base; got != 0 {
		t.Errorf("depth is +%d after balanced exits, want 0", got)
	}
}

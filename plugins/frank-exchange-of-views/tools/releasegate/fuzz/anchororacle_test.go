package fuzz

import (
	"database/sql"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// THE ANCHOR ORACLES RUN AFTER EVERY ACT THE RECORD ADMITTED (execCall), on the record as it then
// stands, and fail the run at the act (noteActErr):
//
//   - no bare anchor: claimcount.BareAnchorIDs over the replayed report is empty (bareAnchors);
//   - anchor_removal agrees with the text: its rows are the placed anchors absent from the replayed
//     report, each keyed on the act after which its token is first absent (removalsDisagree);
//   - the stored gap_state is a fresh gap_refresh over every gap (recordsql.GapStateRefreshDiff);
//   - no computation in the report names a citation absent from it (citesAbsent), read off the
//     replayed text and the computation table, never off the refusal's own query
//     (requireCitationUncited), so the rule cannot vouch for itself.
//
// THE REPORT THEY JUDGE IS REPLAYED HERE, op by op, from the record's whole op stream
// (record.ReportProjectionAt), carried across acts so each check folds only the ops since the last:
// that is what keys a token's absence to the event whose op took it. The replay is also held to the
// report the record renders from its checkpoints (record.RenderAt): the two are one text.

// anchorReplay is the per-act oracles' replay of the report, carried across the run's acts. mu holds
// one oracle check at a time: seats act on goroutines of their own, and the fold is one sequence.
type anchorReplay struct {
	mu     sync.Mutex
	text   string
	parted bool
	// folded is how many ops of the stream the text holds, lastOp the event of the last of them, and
	// head the record's newest event at the last check.
	folded       int
	lastOp, head int64
	// present is the anchors the text holds; firstAbsent each anchor the replay placed and no longer
	// holds, with the event whose op it was first absent after.
	present     map[string]bool
	firstAbsent map[string]int64
}

// advance folds the ops of the stream the replay has not yet folded, one at a time, keying each
// anchor that leaves the text to the event of the op it left at. The stream's prefix is the one the
// replay already folded: an event appended is a later id, so a prefix that moved is the record's
// defect, said so.
func (a *anchorReplay) advance(ops []reportproj.ReportOp) error {
	if len(ops) < a.folded || (a.folded > 0 && ops[a.folded-1].EventID != a.lastOp) {
		return fmt.Errorf("the record's report ops are no longer the %d the replay folded through event %d: the stream holds %d", a.folded, a.lastOp, len(ops))
	}
	if a.present == nil {
		a.present, a.firstAbsent = map[string]bool{}, map[string]int64{}
	}
	for _, op := range ops[a.folded:] {
		text, err := reportproj.Replay(a.text, a.parted, []reportproj.ReportOp{op})
		if err != nil {
			return err
		}
		a.text, a.parted = text, true
		now := map[string]bool{}
		for _, id := range anchor.IDs(text) {
			now[id] = true
			if at, gone := a.firstAbsent[id]; gone {
				return fmt.Errorf("anchor %s, out of the report after event %d, stands in it again after event %d", id, at, op.EventID)
			}
		}
		for id := range a.present {
			if !now[id] {
				a.firstAbsent[id] = op.EventID
			}
		}
		a.present = now
		a.folded++
		a.lastOp = op.EventID
	}
	return nil
}

// anchorRemoval is one row of anchor_removal: the anchor, and the act that took it.
type anchorRemoval struct {
	id    string
	event int64
}

// bareAnchors is why text holds an anchor standing on no words, or "".
func bareAnchors(text string) string {
	if bare := claimcount.BareAnchorIDs(text); len(bare) > 0 {
		return fmt.Sprintf("the replayed report holds anchor(s) %v standing on no words", bare)
	}
	return ""
}

// removalsDisagree is why the anchor_removal rows are not the anchors a's replay placed and no longer
// holds, each keyed on the event after which it is first absent, or "".
func removalsDisagree(a *anchorReplay, rows []anchorRemoval) string {
	var out []string
	by := map[string][]int64{}
	for _, r := range rows {
		by[r.id] = append(by[r.id], r.event)
	}
	for _, id := range slices.Sorted(maps.Keys(by)) {
		events := by[id]
		at, gone := a.firstAbsent[id]
		switch {
		case len(events) > 1:
			out = append(out, fmt.Sprintf("anchor %s has %d anchor_removal rows (events %v), one act takes an anchor", id, len(events), events))
		case a.present[id]:
			out = append(out, fmt.Sprintf("anchor_removal holds %s (event %d) and the replayed report still holds its token", id, events[0]))
		case !gone:
			out = append(out, fmt.Sprintf("anchor_removal holds %s (event %d) and the replay never placed it", id, events[0]))
		case events[0] != at:
			out = append(out, fmt.Sprintf("anchor_removal keys %s on event %d and its token is first absent after event %d", id, events[0], at))
		}
	}
	for _, id := range slices.Sorted(maps.Keys(a.firstAbsent)) {
		if _, ok := by[id]; !ok {
			out = append(out, fmt.Sprintf("anchor %s is out of the replayed report after event %d and anchor_removal holds no row for it", id, a.firstAbsent[id]))
		}
	}
	return strings.Join(out, "; ")
}

// citesAbsent is why a computation whose token text holds names, as the method it applies, a citation
// whose token text does not hold, or "". cites is each computation's citation, off the computation
// table.
func citesAbsent(text string, cites map[string]string) string {
	var out []string
	for _, k := range anchor.IDs(text) {
		if c := cites[k]; anchor.Kind(k) == "computation" && c != "" && !strings.Contains(text, anchor.Token(c)) {
			out = append(out, fmt.Sprintf("computation %s stands in the report applying citation %s, which is out of it", k, c))
		}
	}
	return strings.Join(out, "; ")
}

// anchorOracles runs the per-act oracles on the run's record as it stands, and fails the run at the
// act on the first that does not hold. A record that has not moved since the last check is not
// judged again.
func (r *runner) anchorOracles() {
	if msg := r.anchorOracleMsg(); msg != "" {
		r.noteActErr("anchor oracle: " + msg)
	}
}

// anchorOracleMsg is the per-act oracles' failure, or "".
func (r *runner) anchorOracleMsg() string {
	a := &r.anchorOracle
	a.mu.Lock()
	defer a.mu.Unlock()
	path := filepath.Join(r.run().Records(), "record.db")
	if _, err := os.Stat(path); err != nil {
		return "" // no record yet
	}
	db, err := recordsql.Open(path) // the process's cached handle: never closed here
	if err != nil {
		return "the record could not be opened: " + err.Error()
	}
	msg, moved, err := a.judgeText(db)
	if err != nil {
		return err.Error()
	}
	if msg != "" || !moved {
		return msg
	}
	d, err := recordsql.GapStateRefreshDiff(db)
	if err != nil {
		return "the stored gap state could not be held to the refresh of every gap: " + err.Error()
	}
	if d != "" {
		return "the stored gap state is not the refresh of every gap: " + d
	}
	return ""
}

// judgeText runs the three oracles over the replayed text, off one snapshot of db: what fails, and
// whether the record moved since the last check.
func (a *anchorReplay) judgeText(db *sql.DB) (msg string, moved bool, err error) {
	tx, err := db.Begin()
	if err != nil {
		return "", false, err
	}
	defer func() { _ = tx.Rollback() }()
	var head int64
	if err := tx.QueryRow(`SELECT COALESCE(MAX("id"), 0) FROM "events"`).Scan(&head); err != nil {
		return "", false, err
	}
	if head == a.head {
		return "", false, nil
	}
	a.head = head
	ops, err := record.ReportProjectionAt(tx)
	if err != nil {
		return "", false, fmt.Errorf("the report ops could not be read: %w", err)
	}
	if len(ops) == 0 {
		return "", true, nil // no part yet: no report
	}
	if err := a.advance(ops); err != nil {
		return "the report could not be replayed: " + err.Error(), true, nil
	}
	written, err := record.RenderAt(tx)
	if err != nil {
		return "", false, fmt.Errorf("the report could not be rendered: %w", err)
	}
	if written != a.text {
		return "the report the record renders from its checkpoints is not the replay of its ops", true, nil
	}
	rows, err := queryRemovals(tx)
	if err != nil {
		return "", false, err
	}
	cites, err := queryCites(tx)
	if err != nil {
		return "", false, err
	}
	var out []string
	for _, m := range []string{bareAnchors(a.text), removalsDisagree(a, rows), citesAbsent(a.text, cites)} {
		if m != "" {
			out = append(out, m)
		}
	}
	return strings.Join(out, "; "), true, nil
}

// queryRemovals is every anchor_removal row, in event order.
func queryRemovals(tx *sql.Tx) ([]anchorRemoval, error) {
	rs, err := tx.Query(`SELECT "anchor_id", "event_id" FROM "anchor_removal" ORDER BY "event_id", "ord"`)
	if err != nil {
		return nil, fmt.Errorf("anchor_removal could not be read: %w", err)
	}
	defer rs.Close()
	var out []anchorRemoval
	for rs.Next() {
		var r anchorRemoval
		if err := rs.Scan(&r.id, &r.event); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rs.Err()
}

// queryCites is each computation's citation, the method it applies (computation.cites).
func queryCites(tx *sql.Tx) (map[string]string, error) {
	rs, err := tx.Query(`SELECT "computation_id", COALESCE("cites", '') FROM "computation" WHERE COALESCE("computation_id", '') != ''`)
	if err != nil {
		return nil, fmt.Errorf("the computation table could not be read: %w", err)
	}
	defer rs.Close()
	out := map[string]string{}
	for rs.Next() {
		var k, c string
		if err := rs.Scan(&k, &c); err != nil {
			return nil, err
		}
		out[k] = c
	}
	return out, rs.Err()
}

// THE BARE-ANCHOR ORACLE fails a report holding an anchor on no words, and passes one whose every
// anchor stands on its sentence.
func TestTheBareAnchorOracle(t *testing.T) {
	c := anchor.Token("C-1a2b3c4d")
	if msg := bareAnchors("# Findings\n\nA claim stands" + c + ". Another follows.\n"); msg != "" {
		t.Errorf("a report whose anchor stands on its sentence: %s", msg)
	}
	if msg := bareAnchors("# Findings\n\nA claim stands.\n\n" + c + "\n"); !strings.Contains(msg, "C-1a2b3c4d") {
		t.Errorf("a report holding %s on no words passed: %q", c, msg)
	}
}

// THE REMOVAL ORACLE holds anchor_removal to the replayed text: it passes the rows of the anchors the
// replay placed and no longer holds, each on the event after which its token is first absent, and fails
// a missing row, a row for a token still standing, a row keyed on another event, a row for an anchor
// never placed, and two rows for one anchor.
func TestTheRemovalOracle(t *testing.T) {
	const c1, c2, c9 = "C-1a2b3c4d", "C-5e6f7a8b", "C-99999999"
	var a anchorReplay
	if err := a.advance([]reportproj.ReportOp{
		{EventID: 3, Kind: "append", A: "# Findings\n\nThe first claim holds" + anchor.Token(c1) + ". The second claim holds" + anchor.Token(c2) + ".\n"},
		{EventID: 5, Kind: "remove", A: c1},
	}); err != nil {
		t.Fatal(err)
	}
	if msg := removalsDisagree(&a, []anchorRemoval{{c1, 5}}); msg != "" {
		t.Errorf("the rows of the replay: %s", msg)
	}
	for name, rows := range map[string][]anchorRemoval{
		"no row for an absent anchor":    nil,
		"a row for a standing token":     {{c1, 5}, {c2, 5}},
		"a row keyed on another event":   {{c1, 4}},
		"a row for an anchor not placed": {{c1, 5}, {c9, 5}},
		"two rows for one anchor":        {{c1, 5}, {c1, 6}},
	} {
		if msg := removalsDisagree(&a, rows); msg == "" {
			t.Errorf("%s passed", name)
		}
	}
	// AN ANCHOR BACK IN THE TEXT AFTER IT LEFT is the replay's own failure.
	if err := a.advance([]reportproj.ReportOp{
		{EventID: 3, Kind: "append"}, {EventID: 5, Kind: "remove"},
		{EventID: 7, Kind: "insert", A: "The second claim holds", B: c1},
	}); err == nil {
		t.Error("an anchor placed again after it left replayed without failing")
	}
}

// THE CITES ORACLE fails a computation standing in the report that applies a citation out of it, read
// off the text and the computation table, and passes one whose citation stands.
func TestTheCitesOracle(t *testing.T) {
	const c, k = "C-1a2b3c4d", "K-5e6f7a8b"
	cites := map[string]string{k: c}
	text := "The method holds" + anchor.Token(c) + ". The figure holds" + anchor.Token(k) + "."
	if msg := citesAbsent(text, cites); msg != "" {
		t.Errorf("a computation whose citation stands: %s", msg)
	}
	if msg := citesAbsent(strings.Replace(text, anchor.Token(c), "", 1), cites); !strings.Contains(msg, k) || !strings.Contains(msg, c) {
		t.Errorf("a computation applying a citation out of the report passed: %q", msg)
	}
}

// THE ORACLES PASS A CLEAN RECORD AND FAIL A DOCTORED ONE AT THE NEXT ACT: the staged run after the
// lifecycle arms' setup holds every oracle; with one gap's stored state doctored behind the triggers,
// the next act the record admits fails the run naming the stored gap state.
func TestTheAnchorOraclesFailADoctoredRecordAtTheNextAct(t *testing.T) {
	r := stagedLifecycleArms(t)
	g, why := r.lifecycleMint(citationOwner, "document", lifecycleClaims[0], "", "low")
	if g == "" {
		t.Fatalf("the mint: %s", why)
	}
	r.anchorOracle.head = 0
	if msg := r.anchorOracleMsg(); msg != "" || r.actErr != "" {
		t.Fatalf("a clean record fails the oracles: %s (%s)", msg, r.actErr)
	}
	db, err := recordsql.Open(filepath.Join(r.run().Records(), "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE "gap_state" SET "open" = NOT "open" WHERE "gap_id" = ?`, g); err != nil {
		t.Fatalf("doctor the stored gap state: %v", err)
	}
	r.lifecycleCite("blue-respond", lifecycleClaims[1], "/doctored", "")
	if !strings.Contains(r.actErr, "anchor oracle: the stored gap state is not the refresh of every gap") {
		t.Errorf("the act after the doctoring left the run reading %q", r.actErr)
	}
}

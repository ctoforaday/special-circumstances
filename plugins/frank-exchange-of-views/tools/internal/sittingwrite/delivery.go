package sittingwrite

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A LENS IS HANDED ITS WORK ONE ITEM AT A TIME (gblock 2026-11-06).
//
// The whole work list was the lens's delivery, and most of it was not the lens's work: across 248
// archived lens sittings over nine runs, 217 held no blocking or readied item of the lens's own,
// and the bulk of each delivery was other lenses' open gaps — every one with its passage, backing
// and problem. A delivery over the harness's 10,000-character cap is not delivered at all: it is
// spilled to a file in the PARENT session, and the seat sees a two-kilobyte preview (measured on
// claude 2.1.318). So the lens is handed a header, the head of its queue (record.QueueOf) in full,
// and one line for everything else, and the next item arrives after the call that discharged this
// one (WriteNextItem).

// DeliveryBudget is the most a lens delivery may hold, in bytes of additionalContext, below the
// harness's 10,000-character cap with room for the document that carries it. A delivery over it
// gives way in a fixed order until it fits (fitted), and says what it cut in a field.
const DeliveryBudget = 9000

// clipContext is how much of a clipped passage is kept on each side of the gap's location.
const clipContext = 400

const (
	// lensArrived is the preamble on a lens's delivery at dispatch. It names no command: the surface
	// owns the verbs, and the whole list is where the seat finds everything else.
	lensArrived = "YOUR WORK, read from this run's record at the moment you were dispatched, and handed to you " +
		"one item at a time. `current` is the item to work now, in full, with the gap it is about; `index` " +
		"is one line for each other item queued for you and each other open gap on the board, and " +
		"`estopped` one line for each bench ruling you may not re-raise. `sitting` is your standing: " +
		"whether anything blocks you finishing, your read of the report, the items " +
		"that are not queued (`notes`), the bench's holdings (`declarations`), each gap one names beside " +
		"where the board says it stands now, and granted relief (`relief`). A count in a field ending " +
		"`_omitted` is that many entries left out to keep this delivery whole; your work list holds them. A field " +
		"ending `_clipped` marks a part of the current gap cut short for the same reason: a clipped `backing` is its anchors " +
		"without their readings, which the evidence view gives an anchor at a time, and a clipped `edited_since` is each " +
		"edit's epoch and editor without its text, which the changes view gives for the gap; a clipped passage, problem, " +
		"fix or check is whole on your work list. The next item arrives after the call of yours that the " +
		"record shows discharged this one. Your whole work list is on your own surface when you want it."

	// nextArrived is the preamble on the item handed after a call that discharged the one before it.
	nextArrived = "YOUR NEXT ITEM: the record shows the one you were working discharged, or gone from your list. " +
		"`current` is the item to work now, in full; `index` is the rest of your queue and the other open gaps."

	// changedArrived is the preamble on the current item handed again because the record moved it.
	changedArrived = "YOUR CURRENT ITEM CHANGED on the record since it was handed to you " +
		"(`changed_since_delivered`): it is still the item to work, and this is it as it stands now."
)

// lensDelivery is what a lens is handed: its standing, the current item, and the index.
type lensDelivery struct {
	Sitting  *lensStanding  `json:"sitting,omitempty"`
	Current  *deliveredItem `json:"current"`
	Index    []string       `json:"index"`
	Estopped []string       `json:"estopped,omitempty"`
	// IndexOmitted and EstoppedOmitted count the lines cut from the end of Index and Estopped to keep
	// the delivery within DeliveryBudget (fitted); the work list on the seat's surface holds them all.
	IndexOmitted    int `json:"index_omitted,omitempty"`
	EstoppedOmitted int `json:"estopped_omitted,omitempty"`
}

// lensStanding is the sitting's header: the work list's sitting without its items, the counts, the
// items that are not queued, and the bench's holdings and granted relief — inline where they fit,
// one line each where they do not.
type lensStanding struct {
	Seat     string               `json:"seat"`
	Role     string               `json:"role"`
	Complete bool                 `json:"complete"`
	LastRead *record.LastReadJSON `json:"last_read,omitempty"`
	Counts   struct {
		Open     int `json:"open"`
		Estopped int `json:"estopped"`
		Queued   int `json:"queued"`
	} `json:"counts"`
	// Notes are the list's items that are not queued: the log channel, news of a closure,
	// an anchor no check of this sitting can reach, the statement that nothing is owed.
	// Each is whole where the delivery fits, and one line an entry where it would not (fitted).
	Notes        any `json:"notes"`
	Declarations any `json:"declarations"`
	Relief       any `json:"relief"`
	// NotesOmitted, DeclarationsOmitted and ReliefOmitted count the lines cut from the end of each to
	// keep the delivery within DeliveryBudget; the work list on the seat's surface holds them all.
	NotesOmitted        int `json:"notes_omitted,omitempty"`
	DeclarationsOmitted int `json:"declarations_omitted,omitempty"`
	ReliefOmitted       int `json:"relief_omitted,omitempty"`
	// w is the list the standing was read from, for the one-line forms; never rendered.
	w record.WorkJSON
}

// deliveredItem is a queue entry as the seat is handed it.
type deliveredItem struct {
	Key     string              `json:"key"`
	Kind    record.ItemKind     `json:"kind"`
	Subject *record.ItemSubject `json:"subject,omitempty"`
	Blocks  bool                `json:"blocks"`
	// ChangedSinceDelivered marks an item handed again because the record moved it under the seat.
	ChangedSinceDelivered bool          `json:"changed_since_delivered,omitempty"`
	Items                 []record.Item `json:"items"`
	// ItemsOmitted counts the items cut from the end of Items to keep the delivery within
	// DeliveryBudget (fitted); the work list holds them all.
	ItemsOmitted int           `json:"items_omitted,omitempty"`
	Gap          *deliveredGap `json:"gap,omitempty"`
}

// deliveredGap is the work list's row for the current item's gap, without the synopsis of a problem
// it carries whole.
type deliveredGap struct {
	record.WorkGapJSON
	// ProblemSynopsis shadows the row's synopsis, and is never set: the shallower field wins the
	// name, and a nil one is omitted.
	ProblemSynopsis *struct{} `json:"problem_synopsis,omitempty"`
	// PassageClipped marks a passage cut to its anchor's sentence and the context around it, to keep
	// the delivery within DeliveryBudget; the whole passage is on the work list.
	PassageClipped bool `json:"passage_clipped,omitempty"`
	// Backing and EditedSince shadow the row's own: the row's lists where the delivery fits, and the
	// short forms (anchorIDs, editMarks) where it does not — marked BackingClipped and
	// EditedSinceClipped, with entries cut from the end counted in BackingOmitted and
	// EditedSinceOmitted. Never omitted, as on the row: an empty list is an answer.
	Backing            any  `json:"backing"`
	EditedSince        any  `json:"edited_since"`
	BackingClipped     bool `json:"backing_clipped,omitempty"`
	EditedSinceClipped bool `json:"edited_since_clipped,omitempty"`
	BackingOmitted     int  `json:"backing_omitted,omitempty"`
	EditedSinceOmitted int  `json:"edited_since_omitted,omitempty"`
	// ProblemClipped, RequiredFixClipped, AcceptanceCheckClipped, FixOldClipped and FixNewClipped mark
	// that text cut short, the last thing given way; the work list holds it whole.
	ProblemClipped         bool `json:"problem_clipped,omitempty"`
	RequiredFixClipped     bool `json:"required_fix_clipped,omitempty"`
	AcceptanceCheckClipped bool `json:"acceptance_check_clipped,omitempty"`
	FixOldClipped          bool `json:"fix_old_clipped,omitempty"`
	FixNewClipped          bool `json:"fix_new_clipped,omitempty"`
}

// editMark is an edit of the gap's sentence without its text: the changes view holds the pair.
type editMark struct {
	Epoch    int    `json:"epoch"`
	EditedBy string `json:"edited_by"`
}

// inlineHoldings is the most the declarations and relief may take inline before each is one line.
const inlineHoldings = 1500

// RenderLensDelivery is the lens's delivery at dispatch, over its work list w: its standing, the head
// of its queue (record.QueueOf) in full, and the index.
func RenderLensDelivery(w record.WorkJSON) string { return renderLens(w, record.QueueOf(w)) }

// renderLens is RenderLensDelivery over the queue already read off w.
func renderLens(w record.WorkJSON, q []record.QueueEntry) string {
	d := lensDelivery{Sitting: standingOf(w, q), Index: indexOf(w, q, 1), Estopped: estoppedLines(w)}
	if len(q) > 0 {
		d.Current = itemOf(q[0], false)
	}
	return fitted(lensArrived, d)
}

// renderNext is the item handed after a call: the current entry, and the index of the rest.
func renderNext(w record.WorkJSON, q []record.QueueEntry, at int, changed bool) string {
	d := lensDelivery{Current: itemOf(q[at], changed), Index: indexOf(w, q, -1, at)}
	preamble := nextArrived
	if changed {
		preamble = changedArrived
	}
	return fitted(preamble, d)
}

// fitted renders d under its preamble within DeliveryBudget, measured in BYTES of the rendered
// text — the harness's cap counts characters, and a byte count is never below it. The CURRENT ITEM IS
// ALWAYS DELIVERED, whole but for its gap's passage; everything else gives way, in a fixed order,
// until the delivery fits:
//
//  1. the current gap's passage is cut to its anchor's sentence and the context around it;
//  2. the header's notes, declarations and relief become one line an entry;
//  3. lines are cut from the end of the index, then of the estopped lines, then of the notes,
//     the declarations and the relief, each cut counted in its *_omitted field;
//  4. the current gap's passage is cut to its location alone;
//  5. its backing becomes its anchor ids (backing_clipped), and its edited_since each edit's epoch and
//     editor (edited_since_clipped) — what a seat needs to ask the evidence and changes views;
//  6. anchor ids, then edit marks, are cut from the end, counted in backing_omitted and
//     edited_since_omitted — each is a pointer the views hold, where the prose below is the work;
//  7. its problem, required fix, acceptance check and fix pair are cut, each to a length halved until
//     the delivery fits, down to nothing, each cut marked in its *_clipped field.
//
// THE RECORD HOLDS NO LIMIT on the length of any of these — a mint's problem, fix and check are the
// lens's prose, a sentence may carry any number of anchors — so no maximum size of the current item
// can be computed from the record, and the budget holds because steps 6 and 7 can cut that prose and
// those lists to nothing. What is never cut is the item's key, kind and items, and the gap's id,
// grades, location and short fields, beside the fixed header (the seat, its standing, its read of
// the report and its budget).
func fitted(preamble string, d lensDelivery) string {
	text := render(preamble, d)
	fits := func() bool { text = render(preamble, d); return len(text) <= DeliveryBudget }
	if len(text) <= DeliveryBudget {
		return text
	}
	var g *deliveredGap
	if d.Current != nil {
		g = d.Current.Gap
	}
	if g != nil {
		if clipped, ok := clipPassage(g.Passage, g.Location); ok {
			g.Passage, g.PassageClipped = clipped, true
			if fits() {
				return text
			}
		}
	}
	if s := d.Sitting; s != nil {
		s.Notes, s.Declarations, s.Relief = noteLines(s.Notes), oneLine(holdingLines(s.w)), oneLine(reliefLines(s.w))
		if fits() {
			return text
		}
	}
	cuts := []struct {
		lines   *[]string
		omitted *int
	}{{&d.Index, &d.IndexOmitted}, {&d.Estopped, &d.EstoppedOmitted}}
	if s := d.Sitting; s != nil {
		notes, decls, relief := s.Notes.([]string), s.Declarations.([]string), s.Relief.([]string)
		s.Notes, s.Declarations, s.Relief = &notes, &decls, &relief
		cuts = append(cuts, struct {
			lines   *[]string
			omitted *int
		}{&notes, &s.NotesOmitted}, struct {
			lines   *[]string
			omitted *int
		}{&decls, &s.DeclarationsOmitted}, struct {
			lines   *[]string
			omitted *int
		}{&relief, &s.ReliefOmitted})
	}
	for _, c := range cuts {
		for len(*c.lines) > 0 {
			*c.lines = (*c.lines)[:len(*c.lines)-1]
			*c.omitted++
			if fits() {
				return text
			}
		}
	}
	if g == nil {
		// AN ENTRY OF MANY ITEMS AND NO GAP — the voice lens's composed sources (record.QueueOf) — gives
		// way item by item from its end, counted, as the index does: each one is whole on the work list.
		if c := d.Current; c != nil {
			for len(c.Items) > 1 {
				c.Items = c.Items[:len(c.Items)-1]
				c.ItemsOmitted++
				if fits() {
					return text
				}
			}
		}
		return text
	}
	if g.Passage != g.Location {
		g.Passage, g.PassageClipped = g.Location, true
		if fits() {
			return text
		}
	}
	ids, marks := anchorIDs(g.WorkGapJSON.Backing), editMarks(g.WorkGapJSON.EditedSince)
	g.Backing, g.BackingClipped = ids, len(ids) > 0
	g.EditedSince, g.EditedSinceClipped = marks, len(marks) > 0
	if fits() {
		return text
	}
	for len(ids) > 0 {
		ids = ids[:len(ids)-1]
		g.Backing, g.BackingOmitted = ids, g.BackingOmitted+1
		if fits() {
			return text
		}
	}
	for len(marks) > 0 {
		marks = marks[:len(marks)-1]
		g.EditedSince, g.EditedSinceOmitted = marks, g.EditedSinceOmitted+1
		if fits() {
			return text
		}
	}
	prose := []struct {
		whole   string
		text    *string
		clipped *bool
	}{
		{g.WorkGapJSON.Problem, &g.Problem, &g.ProblemClipped},
		{g.WorkGapJSON.RequiredFix, &g.RequiredFix, &g.RequiredFixClipped},
		{g.WorkGapJSON.AcceptanceCheck, &g.AcceptanceCheck, &g.AcceptanceCheckClipped},
		{g.WorkGapJSON.FixOld, &g.FixOld, &g.FixOldClipped},
		{g.WorkGapJSON.FixNew, &g.FixNew, &g.FixNewClipped},
	}
	longest := 0
	for _, p := range prose {
		longest = max(longest, len([]rune(p.whole)))
	}
	for n := longest / 2; ; n /= 2 {
		for _, p := range prose {
			if len([]rune(p.whole)) > n {
				*p.text, *p.clipped = clip(p.whole, n), true
			}
		}
		if fits() || n == 0 {
			return text
		}
	}
}

// anchorIDs is a gap's backing as its anchor ids alone; the evidence view holds each one's reading.
func anchorIDs(backing []record.GapBackingJSON) []string {
	out := []string{}
	for _, b := range backing {
		out = append(out, b.Anchor)
	}
	return out
}

// editMarks is a gap's edits without their text; the changes view holds each pair.
func editMarks(edits []record.GapEdit) []editMark {
	out := []editMark{}
	for _, e := range edits {
		out = append(out, editMark{Epoch: e.Epoch, EditedBy: e.EditedBy})
	}
	return out
}

// noteLines is the notes, one line each: the item's kind, what it is about, and its sentence cut
// short. Already lines, they are returned as they are.
func noteLines(notes any) []string {
	items, ok := notes.([]record.Item)
	if !ok {
		lines, _ := notes.([]string)
		return lines
	}
	out := []string{}
	for _, it := range items {
		line := string(it.Kind)
		if it.Subject != nil {
			line += " · " + string(it.Subject.Kind) + " " + it.Subject.ID
		}
		out = append(out, line+" · "+clip(it.What, 160))
	}
	return out
}

func oneLine(lines []string) []string {
	if lines == nil {
		return []string{}
	}
	return lines
}

func render(preamble string, d lensDelivery) string {
	b, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		return fmt.Sprintf("%s\n\nthe delivery could not be rendered: %v\n", unreadable, err)
	}
	return preamble + "\n\n" + string(b) + "\n"
}

// standingOf is the delivery's header.
func standingOf(w record.WorkJSON, q []record.QueueEntry) *lensStanding {
	s := &lensStanding{Seat: w.Sitting.Seat, Role: w.Sitting.Role, Complete: w.Sitting.Complete,
		LastRead: w.Sitting.LastRead, w: w}
	notes := []record.Item{}
	s.Counts.Open, s.Counts.Estopped, s.Counts.Queued = w.Counts.Open, w.Counts.Estopped, len(q)
	queued := map[string]bool{}
	for _, e := range q {
		for _, it := range e.Items {
			queued[itemID(it)] = true
		}
	}
	for _, it := range w.Sitting.Open {
		if !queued[itemID(it)] {
			notes = append(notes, it)
		}
	}
	s.Notes = notes
	s.Declarations, s.Relief = holdingsOf(w)
	return s
}

// itemID is an item's identity within one read of the list: its fields, the sentence included.
func itemID(it record.Item) string {
	b, _ := json.Marshal(it)
	return string(b)
}

// holdingsOf is the bench's declarations and granted relief, inline where together they fit within
// inlineHoldings, and otherwise one line each.
func holdingsOf(w record.WorkJSON) (declarations, relief any) {
	b, _ := json.Marshal([]any{w.Declarations, w.Relief})
	if len(b) <= inlineHoldings {
		return w.Declarations, w.Relief
	}
	return holdingLines(w), reliefLines(w)
}

// holdingLines is each declaration in one line: its epoch, its holding cut short, and the gaps it
// names with where each stands.
func holdingLines(w record.WorkJSON) []string {
	out := []string{}
	for _, d := range w.Declarations {
		var ids []string
		for _, g := range d.Gaps {
			ids = append(ids, g.ID+" "+g.Status)
		}
		out = append(out, fmt.Sprintf("epoch %d: %s · gaps: %s", d.Epoch, clip(d.Holding, 240), strings.Join(ids, ", ")))
	}
	return out
}

// reliefLines is each granted petition's order in one line.
func reliefLines(w record.WorkJSON) []string {
	out := []string{}
	for _, r := range w.Relief {
		out = append(out, fmt.Sprintf("%s (%s, binds %s, epoch %d): %s", r.Motion, r.Petitioner, r.Binds, r.Epoch, clip(r.Relief, 240)))
	}
	return out
}

// itemOf is a queue entry as the seat is handed it.
func itemOf(e record.QueueEntry, changed bool) *deliveredItem {
	d := &deliveredItem{Key: e.Key, Kind: e.Kind, Subject: e.Subject, Blocks: e.Blocks, ChangedSinceDelivered: changed, Items: e.Items}
	if e.Gap != nil {
		d.Gap = &deliveredGap{WorkGapJSON: *e.Gap, Backing: e.Gap.Backing, EditedSince: e.Gap.EditedSince}
	}
	return d
}

// indexOf is one line per queue entry other than the ones at skip, then one per open gap no entry is
// about. from is the first entry listed (1 at dispatch, where entry 0 is current), or -1 for every
// entry not skipped.
func indexOf(w record.WorkJSON, q []record.QueueEntry, from int, skip ...int) []string {
	out := []string{}
	about := map[string]bool{}
	for i, e := range q {
		if e.Gap != nil {
			about[e.Gap.ID] = true
		}
		if (from >= 0 && i < from) || containsInt(skip, i) {
			continue
		}
		state := "open"
		switch {
		case e.Blocks:
			state = "blocks"
		case e.Kind == record.ItemReaudit && e.Gap != nil && e.Gap.YourTurn:
			state = "readied"
		}
		if e.Gap != nil {
			out = append(out, fmt.Sprintf("%s · %s/%s · %s · %s", e.Gap.ID, e.Kind, e.Gap.Class, e.Gap.MintedBy, state))
			continue
		}
		out = append(out, fmt.Sprintf("%s · %s", e.Key, state))
	}
	for _, list := range [][]record.WorkGapJSON{w.Open, w.ClosingsOwed} {
		for _, g := range list {
			if !about[g.ID] {
				about[g.ID] = true
				out = append(out, fmt.Sprintf("%s · %s · %s · open", g.ID, g.Class, g.MintedBy))
			}
		}
	}
	return out
}

// estoppedLines is one line per bench ruling the seat may not re-raise: the gap, its class, the fate,
// and what the ruling settled.
func estoppedLines(w record.WorkJSON) []string {
	out := []string{}
	for _, e := range w.Estopped {
		line := fmt.Sprintf("%s · %s · %s", e.ID, e.Class, e.Fate)
		if e.Settled != "" {
			line += " · settled: " + clip(e.Settled, 200)
		}
		out = append(out, line)
	}
	return out
}

// clipPassage cuts a passage to the sentence holding location and clipContext bytes either side,
// widened to sentence ends; ok is false where it is already that short.
func clipPassage(passage, location string) (string, bool) {
	if len(passage) <= 2*clipContext+len(location) {
		return passage, false
	}
	at := -1
	if location != "" {
		at = strings.Index(passage, location)
	}
	if at < 0 {
		at = 0
	}
	start, end := at-clipContext, at+len(location)+clipContext
	if start < 0 {
		start = 0
	}
	if end > len(passage) {
		end = len(passage)
	}
	if i := strings.LastIndex(passage[:start], ". "); start > 0 && i >= 0 && start-i < clipContext/2 {
		start = i + 2
	}
	if i := strings.Index(passage[end:], ". "); end < len(passage) && i >= 0 && i < clipContext/2 {
		end += i + 1
	}
	start, end = runeStart(passage, start), runeStart(passage, end)
	out := passage[start:end]
	if start > 0 {
		out = "…" + out
	}
	if end < len(passage) {
		out += "…"
	}
	return out, true
}

// runeStart moves i back to the start of the rune it falls in.
func runeStart(s string, i int) int {
	for i > 0 && i < len(s) && s[i]&0xC0 == 0x80 {
		i--
	}
	return i
}

func clip(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

func containsInt(xs []int, x int) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

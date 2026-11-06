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
// clips the current gap's passage to the sentence its anchor sits in and the context around it,
// and says so (passage_clipped).
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
		"whether anything blocks you finishing, your read of the report, your mint budget, the items " +
		"that are not work to queue (`notes`), the bench's holdings (`declarations`), each gap one names beside " +
		"where the board says it stands now, and granted relief (`relief`). The next item arrives after the call of yours that the " +
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
}

// lensStanding is the sitting's header: the work list's sitting without its items, the counts, the
// items that are not queued, and the bench's holdings and granted relief — inline where they fit,
// one line each where they do not.
type lensStanding struct {
	Seat       string                 `json:"seat"`
	Role       string                 `json:"role"`
	Complete   bool                   `json:"complete"`
	LastRead   *record.LastReadJSON   `json:"last_read,omitempty"`
	MintBudget *record.MintBudgetJSON `json:"mint_budget,omitempty"`
	Counts     struct {
		Open     int `json:"open"`
		Estopped int `json:"estopped"`
		Queued   int `json:"queued"`
	} `json:"counts"`
	// Notes are the list's items that are not queued: the log channel, the budget, news of a closure,
	// an anchor no check of this sitting can reach, the statement that nothing is owed.
	Notes        []record.Item `json:"notes"`
	Declarations any           `json:"declarations"`
	Relief       any           `json:"relief"`
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
	Gap                   *deliveredGap `json:"gap,omitempty"`
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

// fitted renders d under its preamble, clipping the current gap's passage when the whole is over
// DeliveryBudget.
func fitted(preamble string, d lensDelivery) string {
	text := render(preamble, d)
	if len(text) <= DeliveryBudget || d.Current == nil || d.Current.Gap == nil {
		return text
	}
	g := d.Current.Gap
	if clipped, ok := clipPassage(g.Passage, g.Location); ok {
		g.Passage, g.PassageClipped = clipped, true
		text = render(preamble, d)
	}
	return text
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
		LastRead: w.Sitting.LastRead, MintBudget: w.Sitting.MintBudget, Notes: []record.Item{}}
	s.Counts.Open, s.Counts.Estopped, s.Counts.Queued = w.Counts.Open, w.Counts.Estopped, len(q)
	queued := map[string]bool{}
	for _, e := range q {
		for _, it := range e.Items {
			queued[itemID(it)] = true
		}
	}
	for _, it := range w.Sitting.Open {
		if !queued[itemID(it)] {
			s.Notes = append(s.Notes, it)
		}
	}
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
	dl, rl := []string{}, []string{}
	for _, d := range w.Declarations {
		var ids []string
		for _, g := range d.Gaps {
			ids = append(ids, g.ID+" "+g.Status)
		}
		dl = append(dl, fmt.Sprintf("epoch %d: %s · gaps: %s", d.Epoch, clip(d.Holding, 240), strings.Join(ids, ", ")))
	}
	for _, r := range w.Relief {
		rl = append(rl, fmt.Sprintf("%s (%s, binds %s, epoch %d): %s", r.Motion, r.Petitioner, r.Binds, r.Epoch, clip(r.Relief, 240)))
	}
	return dl, rl
}

// itemOf is a queue entry as the seat is handed it.
func itemOf(e record.QueueEntry, changed bool) *deliveredItem {
	d := &deliveredItem{Key: e.Key, Kind: e.Kind, Subject: e.Subject, Blocks: e.Blocks, ChangedSinceDelivered: changed, Items: e.Items}
	if e.Gap != nil {
		d.Gap = &deliveredGap{WorkGapJSON: *e.Gap}
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

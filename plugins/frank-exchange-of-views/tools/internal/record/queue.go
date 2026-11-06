package record

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
)

// A LENS'S WORK, ONE ITEM AT A TIME (gblock 2026-11-06).
//
// A lens was handed its whole work list at dispatch and told to read it again after every act. The
// list is mostly not its work: across 248 archived lens sittings, 217 held no blocking or readied
// item of the lens's own, and the bulk of each delivery was other lenses' gaps, every one with its
// passage, backing and problem. So the list is ordered into a queue, the seat is handed the head of
// it in full and a one-line index of the rest, and the next item arrives when the record says the
// current one is discharged.
//
// THE QUEUE IS READ OFF THE LIST'S FIELDS AND NOTHING ELSE. An item is keyed by its kind and subject
// (itemkind.go), a gap by its row, and "discharged" is "gone from the next read of the same list" —
// the write path's own reading of the act, never a parse of the sentence that described it.

// QueueEntry is one unit of a seat's work: an item the seat owes, or a gap of its own the dispatch
// readies it on with the items about that gap folded in.
type QueueEntry struct {
	// Key is the entry's identity across reads of the list: its kind and its subject. Two entries of
	// one read never share a key.
	Key     string       `json:"key"`
	Kind    ItemKind     `json:"kind"`
	Subject *ItemSubject `json:"subject,omitempty"`
	// Blocks is whether any item the entry carries blocks the seat's sitting.
	Blocks bool `json:"blocks"`
	// Items are the list's items the entry carries, in list order: the one owed item, or, for a
	// readied gap, the items about it. Never omitted; a readied gap no item names carries none.
	Items []Item `json:"items"`
	// Gap is the list's row for the gap the entry is about, where it is about one the list holds.
	Gap *WorkGapJSON `json:"gap,omitempty"`
	// Digest is a hash of Items and Gap: the same key with a different digest is the same work
	// changed under the seat.
	Digest string `json:"digest"`
}

// QueueOf orders a seat's work list into the queue its work arrives in, from the list's fields:
//
//  1. the audit item, while the lens has not been served the report whole since it was last behind
//     (last_read first or behind);
//  2. every item the seat's own act clears (SittingJSON.Owed), in list order;
//  3. each open gap the dispatch readies the seat on (your_turn) — material first, then by mass, then
//     in mint order — with the non-blocking items about it folded in;
//  4. each accepted grade motion owed a regrade, and each finding about the seat's own gap, not
//     already folded into a readied gap.
//
// Everything else on the list is not queued: the log channel, the budget, the statement that nothing
// is owed, news of a closure, an anchor no check can reach, and every gap that is not the seat's to
// act on. A delivery states those in its header and index.
func QueueOf(w WorkJSON) []QueueEntry {
	rows := map[string]*WorkGapJSON{}
	for _, list := range [][]WorkGapJSON{w.Open, w.ClosingsOwed} {
		for i := range list {
			rows[list[i].ID] = &list[i]
		}
	}
	rowOf := func(s *ItemSubject) *WorkGapJSON {
		if s == nil || s.Kind != SubjectGap {
			return nil
		}
		return rows[s.ID]
	}
	var out []QueueEntry
	used := make([]bool, len(w.Sitting.Open))
	add := func(e QueueEntry) {
		e.Key = queueKey(e.Kind, e.Subject)
		for n := 2; slices.ContainsFunc(out, func(o QueueEntry) bool { return o.Key == e.Key }); n++ {
			e.Key = fmt.Sprintf("%s#%d", queueKey(e.Kind, e.Subject), n)
		}
		if e.Items == nil {
			e.Items = []Item{}
		}
		for _, it := range e.Items {
			e.Blocks = e.Blocks || it.Blocks
		}
		e.Digest = digestOf(e.Items, e.Gap)
		out = append(out, e)
	}
	single := func(i int) {
		it := w.Sitting.Open[i]
		used[i] = true
		add(QueueEntry{Kind: it.Kind, Subject: it.Subject, Items: []Item{it}, Gap: rowOf(it.Subject)})
	}

	if lr := w.Sitting.LastRead; lr != nil && (lr.Kind == "first" || lr.Kind == "behind") {
		for i, it := range w.Sitting.Open {
			if it.Kind == ItemAudit && !used[i] {
				single(i)
			}
		}
	}
	for i, it := range w.Sitting.Open {
		if it.Blocks && !it.AnotherSeatClears && !used[i] {
			single(i)
		}
	}
	var readied []*WorkGapJSON
	for i := range w.Open {
		if w.Open[i].YourTurn {
			readied = append(readied, &w.Open[i])
		}
	}
	slices.SortStableFunc(readied, func(a, b *WorkGapJSON) int {
		if a.Material != b.Material {
			if a.Material {
				return -1
			}
			return 1
		}
		return cmp.Compare(massOf(b), massOf(a))
	})
	for _, g := range readied {
		e := QueueEntry{Kind: ItemReaudit, Subject: about(SubjectGap, g.ID), Gap: g}
		for i, it := range w.Sitting.Open {
			if used[i] || it.Blocks || it.Subject == nil || *it.Subject != *e.Subject {
				continue
			}
			used[i] = true
			e.Items = append(e.Items, it)
		}
		add(e)
	}
	for i, it := range w.Sitting.Open {
		if !used[i] && (it.Kind == ItemRegradeOwed || it.Kind == ItemFindingOnGap) {
			single(i)
		}
	}
	return out
}

// queueKey is an entry's key: its kind, and the kind and id of what it is about.
func queueKey(k ItemKind, s *ItemSubject) string {
	if s == nil {
		return string(k)
	}
	return string(k) + ":" + string(s.Kind) + ":" + s.ID
}

// massOf is a gap row's severity mass (GapMass), 0 where its grades are not words the mass table holds.
func massOf(g *WorkGapJSON) float64 {
	l, _ := g.Likelihood.(string)
	i, _ := g.Impact.(string)
	return GapMass(l, i)
}

// digestOf hashes what an entry hands the seat, so a re-read can tell the same work from changed work.
func digestOf(items []Item, gap *WorkGapJSON) string {
	b, err := json.Marshal(struct {
		Items []Item       `json:"items"`
		Gap   *WorkGapJSON `json:"gap"`
	}{items, gap})
	if err != nil {
		// Every field is a string, number, bool or a list of them; a marshal error is a programming
		// fault, and the digest still differs from every real one.
		return "unmarshalable: " + err.Error()
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}

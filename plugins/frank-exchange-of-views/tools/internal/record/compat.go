package record

// THE DUAL-READ, AND IT IS PERMANENT (#344).
//
// The record is append-only and stored runs are re-read — `bench assemble`, every `show`,
// `capture`, the dashboard. Before the motion collapse, the three exchanges were written as
// `dispute`/`dispute-respond`, `petition`/`petition-rule` and `avenue`/`avenue-rule`. Every
// consumer of those types is a bare `switch e.Type` with no default, so under a `motion` type a
// pre-collapse record would render AN EMPTY SECTION AND `0 filed / 0 ruled` — indistinguishable
// from a run that genuinely had no disputes and no petitions.
//
// That is `facts-are-fields`' plausible zero, produced by the change whose stated purpose is
// removing plausible zeros. So the retired types are READ, not dropped.
//
// WHY THERE IS NO SUNSET. An earlier draft of the plan had the dual-read retire "when no run
// directory in research/ still carries the retired types". That test is unobservable: this plugin
// ships to installing projects, whose run records are invisible to it — it could read clear here
// while consumers' records still carry the old vocabulary. And a record is permanent: a reader of
// a 2026 record in 2030 must still work. The dual-read is the honest cost of the collapse, not a
// defect awaiting cleanup.
//
// WHY IT IS A SEPARATE FILE. Bending `Motions` to accept three legacy shapes would put the
// compatibility concern inside the shape it compromises. Here the translation is one place, named
// for what it is, and deletable in one piece if a generation is ever genuinely retired.

// legacyMotions reconstructs Motions from a PRE-COLLAPSE record.
//
// The ids are SYNTHESISED, and they have to be: the whole reason for the collapse is that the old
// exchanges had no id, so a legacy motion's identity cannot be recovered — only assigned. They
// are assigned deterministically from the filing's position so two replays of the same record
// agree, and marked so no reader mistakes a synthesised id for one the record carried.
func legacyMotions(b *Board) []*Motion {
	// TWO PASSES, AND THE FIRST DRAFT'S SINGLE PASS WAS WRONG.
	//
	// Each seat writes its own shard and replay orders by TIMESTAMP across all of them, so the
	// shards interleave and a RULING CAN REPLAY BEFORE ITS FILING. Measured on the compat
	// fixture the moment it existed: `petition-rule` (judge-r1) replays before `petition`
	// (red-merge-r1), and `avenue-rule` replays after both `avenue` events. A single pass that
	// files-then-rules silently dropped the petition's ruling and the direction's appeal — two
	// of the three exchanges, rendered as asks nobody answered.
	//
	// Pass 1 creates every motion. Pass 2 attaches every answer. Order-independent by
	// construction, which is the only safe assumption about a multi-shard log.
	var out []*Motion
	byGrade := map[string]*Motion{}    // (gap_id, dimension) — blue may contest two grades on one gap
	byPetition := map[string]*Motion{} // (petitioner, class) — #312: cannot split two filings by one seat in one round
	byAvenue := map[string]*Motion{}

	next := func() string { return legacyID(len(out) + 1) }

	for _, e := range b.Events {
		switch e.Type {
		case "dispute":
			m := &Motion{
				ID: next(), Subject: "grade", Filer: e.SeatID, Round: e.Round,
				Basis: e.Payload.Str("evidence"),
				Fields: map[string]string{
					"gap_id": e.Payload.Str("gap_id"), "dimension": e.Payload.Str("dimension"),
					"proposed": e.Payload.Str("proposed"), "legacy": "true",
				},
			}
			byGrade[e.Payload.Str("gap_id")+"/"+e.Payload.Str("dimension")] = m
			out = append(out, m)
		case "petition":
			m := &Motion{
				ID: next(), Subject: "petition", Filer: e.SeatID, Round: e.Round,
				Basis: e.Payload.Str("basis"), Relief: e.Payload.Str("relief"),
				Fields: map[string]string{"class": e.Payload.Str("class"), "legacy": "true"},
			}
			byPetition[e.SeatID+"/"+e.Payload.Str("class")] = m
			out = append(out, m)
		case "avenue-rule":
			// A direction's FILING is the proposal (an `avenue` event with its own lifecycle),
			// so the legacy motion is created by the RULING and back-references it.
			id := e.Payload.Str("avenue_id")
			m := &Motion{
				ID: next(), Subject: "direction", Round: e.Round,
				Fields: map[string]string{"avenue_id": id, "legacy": "true"},
			}
			byAvenue[id] = m
			out = append(out, m)
		}
	}

	for _, e := range b.Events {
		switch e.Type {
		case "dispute-respond":
			if m := byGrade[e.Payload.Str("gap_id")+"/"+e.Payload.Str("dimension")]; m != nil {
				m.Ruling, m.RulingBy, m.RulingRound = e.Payload.Str("response"), e.SeatID, e.Round
				m.Opinion = e.Payload.Str("rationale")
			}
		case "petition-rule":
			if m := byPetition[e.Payload.Str("petitioner")+"/"+e.Payload.Str("class")]; m != nil {
				m.Ruling, m.RulingBy, m.RulingRound = e.Payload.Str("ruling"), e.SeatID, e.Round
				m.Opinion = e.Payload.Str("opinion")
			}
		case "avenue-rule":
			if m := byAvenue[e.Payload.Str("avenue_id")]; m != nil {
				m.Ruling, m.RulingBy, m.RulingRound = e.Payload.Str("ruling"), e.SeatID, e.Round
				m.Opinion = e.Payload.Str("reason")
			}
		case "avenue":
			// `contests_ruling` is blue pursuing against a ruling — the legacy spelling of an
			// appeal, and the one field on the old shape with no counterpart at all.
			if c := e.Payload.Str("contests_ruling"); c != "" {
				if m := byAvenue[e.Payload.Str("avenue_id")]; m != nil {
					m.Appealed, m.AppealReason = true, e.Payload.Str("reason")
					m.Filer = e.SeatID
				}
			}
		}
	}
	return out
}

// legacyID marks a synthesised id so nothing mistakes it for one the record carried. A reader
// that sees `L1` knows the identity was assigned at read time and that a legacy petition join may
// be ambiguous (#312); a reader that sees `M1` knows the record said so.
func legacyID(n int) string {
	return "L" + itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// AllMotions returns every motion on the record, from whichever vocabulary wrote it.
//
// THIS IS THE ONE READER EVERY CONSUMER SHOULD USE. A consumer that calls Motions directly gets
// the post-collapse events only and silently renders nothing for a pre-collapse record — the
// exact failure the dual-read exists to prevent, reintroduced one call site at a time.
func AllMotions(b *Board) []*Motion {
	// CONCATENATE, NEVER CHOOSE. The first draft returned the legacy set only when the motion set
	// was EMPTY, and that is wrong for the whole additive stage, when both vocabularies are live:
	// one `motion grade file` anywhere in a run would have made the motion set non-empty and
	// dropped every `dispute`, `petition` and `avenue-rule` in the same record — silently, into a
	// section that still rendered and still looked complete. The fuzz drives both by design, which
	// is what put the mixed record in front of me.
	//
	// Nothing writes an exchange in both vocabularies, so there is no double count to guard
	// against: a motion is filed one way or the other, and its id namespace says which (M-ids and
	// avenue A-ids from the record, L-ids synthesised at read time).
	return append(Motions(b), legacyMotions(b)...)
}

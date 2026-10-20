package record

// ArchiveMigrator is where the external test package installs migrate (archivemigrate_test.go),
// for archivedEvents.
var ArchiveMigrator = &archiveMigrate

// The plan's `why` producers that print an events.id, for the external test that renders each with
// sentinel ids and hands the lines to migrate's translation (planidphrases_test.go).

// LensReadiness is the lens fold's reason for a lens in the given state (lensFold.readiness).
func LensReadiness(seat string, head, pin int64, sittings, barren int, owes []string, short string) string {
	_, why := lensFold{seat: seat, pin: pin, sittings: sittings, barren: barren, owes: owes}.readiness(head, short)
	return why
}

// UnsatSitting is the fold's clause on a sitting the lens did not sit through.
var UnsatSitting = unsatSitting

// SittingShortReasons is every reason satThrough gives for a sitting its seat did not sit through.
var SittingShortReasons = []string{shortCallLimit, shortNeverClosed, shortNotClosed}

// RetiresAfter is the count of barren sittings that retires a lens.
const RetiresAfter = retiresAfter

// OwedBlocker is a blocker the lens owns, as its debts name it.
func OwedBlocker(kind BlockerKind, subject string) string {
	return owedBlocker(Blocker{Kind: kind, Subject: subject})
}

// ContradictionReadiness is the plan's reason on an unanswered contradiction a lens read.
func ContradictionReadiness(owner, claim, subject string, satSince bool) string {
	return contradictionReadiness(Blocker{Kind: BlockerContradiction, Owner: owner, Detail: claim, Subject: subject}, satSince)
}

// BlockerWhy is the plan's reason for one blocker of each kind that states one, keyed by kind;
// subject names the blocker's subject for the kind, and since and lastCheck are its events.ids.
func BlockerWhy(subject func(BlockerKind) string, owner, detail string, since, lastCheck func() int64) map[BlockerKind]string {
	out := map[BlockerKind]string{}
	for _, k := range blockerKinds {
		if k.why == nil {
			continue
		}
		out[k.kind] = k.why(Blocker{Kind: k.kind, Subject: subject(k.kind), Owner: owner, Detail: detail, since: since(), lastCheck: lastCheck()})
	}
	return out
}

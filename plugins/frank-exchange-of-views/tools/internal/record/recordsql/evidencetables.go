package recordsql

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EVIDENCE TABLES: WHAT AN EVIDENCE ANCHOR'S KIND IS ON THE RECORD.
//
// An evidence anchor is an anchor whose row of the anchor kinds table (internal/anchor) says it
// `backs` its sentence: a citation or a computation. Every such anchor lives one way — placed at the
// end of its quote, carried by every edit, checked by one lens verb, retired by `retire` — and every
// reader of that lifecycle asks the kind's row here for what differs, never `anchor.Kind(id) == "…"`.
//
// It lives in recordsql, beside the DDL, and not in record: the views are written from it, record
// imports recordsql, and recordsql cannot import record. The anchor kinds table stays the vocabulary
// leaf (label, token, assembly, backs); this table holds the record's facts about each backing kind.
// TestEveryBackingKindHasAnEvidenceRow (record) holds the two level, and each column to the DDL.

// EvidenceTable is one backing kind's row.
type EvidenceTable struct {
	// Kind is the anchor kinds table's name for the kind (anchor.Kind).
	Kind string
	// Table and IDColumn are the event that places an anchor of the kind and the column holding
	// its id: ONE placing table per kind.
	Table, IDColumn string
	// Check is the check kind a live anchor of the kind answers: the enum_check_kind word a gap's
	// `check_kind` is compared with.
	Check string
	// CheckTable and CheckIDColumn are the event that checks an anchor of the kind and the column
	// naming the anchor it checks.
	CheckTable, CheckIDColumn string
	// Reads is the SQL predicate over a CheckTable row, aliased as given, under which that check
	// READ its sentence and so replaces its seat's earlier check: the anchor ledger's `reads`
	// (record.anchorChecksOf), stated in SQL.
	Reads func(alias string) string
	// Holds is the SQL predicate over a CheckTable row, aliased as given, under which that check
	// backs its sentence: the anchor ledger's `holds` (record.anchorChecksOf), stated in SQL.
	Holds func(alias string) string
	// Noun is the word every refusal and result names an anchor of the kind by.
	Noun string
	// Verb is blue's verb that places an anchor of the kind, named by every refusal that asks for
	// one placed as a gap's answer (`<verb> --answers G`).
	Verb string
	// Failed is the failed-check state word (record.CheckState) in which a check of this kind
	// finds against its sentence.
	Failed string
}

// EvidenceTables is every backing kind's row, in the anchor kinds table's order.
var EvidenceTables = []EvidenceTable{
	{
		Kind: "citation", Table: "cite", IDColumn: "label", Check: recordpb.Word(recordpb.CheckKind_CHECK_KIND_SOURCE),
		CheckTable: "verify", CheckIDColumn: "anchor", Reads: verifyReads, Holds: verifyHolds, Noun: "citation", Verb: "cite", Failed: "found_against",
	},
	{
		Kind: "computation", Table: "computation", IDColumn: "computation_id", Check: recordpb.Word(recordpb.CheckKind_CHECK_KIND_COMPUTATION),
		CheckTable: "reproduce", CheckIDColumn: "computation_id", Reads: reproduceReads, Holds: reproduceHolds, Noun: "computation", Verb: "compute", Failed: "unsound",
	},
}

// EvidenceTableOf is the row of the backing kind named kind, and false for a kind that backs nothing.
func EvidenceTableOf(kind string) (EvidenceTable, bool) {
	for _, t := range EvidenceTables {
		if t.Kind == kind {
			return t, true
		}
	}
	return EvidenceTable{}, false
}

// verifyHolds is a reading that backs its sentence: it reads the work (record.VerifyReads — an
// outcome that read it, of a copy that is the work) and its outcome does not find against
// (record.FindsAgainst reads the same `finds_against` option), so `weak` backs and `refutes` and
// `absent` do not. The outcome list is read off the enum's annotations, so a value added to
// SourceOutcome lands here as the ledger reads it.
func verifyHolds(alias string) string {
	var backing []string
	vals := recordpb.SourceOutcome(0).Descriptor().Values()
	for i := 0; i < vals.Len(); i++ {
		o := recordpb.SourceOutcome(vals.Get(i).Number())
		if o == recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED || o == recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE || recordpb.FindsAgainstOf(o) {
			continue
		}
		backing = append(backing, "'"+recordpb.Word(o)+"'")
	}
	return fmt.Sprintf(`(%[1]s."outcome" IN (%[2]s) AND COALESCE(%[1]s."source_completeness", '') <> '%[3]s')`,
		alias, strings.Join(backing, ", "), recordpb.Word(recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK))
}

// verifyReads is a reading that read the work (record.VerifyReads): an outcome other than
// `unreachable`, of a copy the tool did not stamp not_the_work.
func verifyReads(alias string) string {
	return fmt.Sprintf(`(%[1]s."outcome" IS NOT NULL AND %[1]s."outcome" <> '%[2]s' AND COALESCE(%[1]s."source_completeness", '') <> '%[3]s')`,
		alias, recordpb.Word(recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE), recordpb.Word(recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK))
}

// reproduceReads is every judgement: a stored bundle always runs, so a reproduce always reads.
func reproduceReads(string) string { return "1" }

// reproduceHolds is a judgement that the computation establishes its sentence: `sound`.
func reproduceHolds(alias string) string {
	return fmt.Sprintf(`%s."soundness" = '%s'`, alias, recordpb.Word(recordpb.Soundness_SOUNDNESS_SOUND))
}

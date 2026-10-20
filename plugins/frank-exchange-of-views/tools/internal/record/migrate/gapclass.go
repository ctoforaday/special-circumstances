package migrate

// archivedGapClasses maps an archived gap-class slug that is not a current GapClass word under
// recordpb.SameWord to the value it migrates as. A kept registry slug (`report-voice`) resolves by
// the fold and has no row; a merged or dropped registry slug and every slug a run coined has one.
// Registered in valueRenames under feov.record.v1.GapClass, so a mint's class resolves through it
// like every other archived enum word (plans/gap-class-enum.md §III.2 Tables B and C).
var archivedGapClasses = map[string]string{
	// Registry classes merged into a kept value (Table B).
	"enumeration-non-exhaustive":          "exhaustive_sweep_omits_case",
	"figure-miscomposition":               "figure_recount_fails",
	"causal-narrative-fails-reproduction": "figure_recount_fails",
	"pricing-basis-drift":                 "metric_conflation",
	"measurement-methodology-drift":       "metric_conflation",
	"citation-status-drift":               "live_source_drift",
	"audited-artifact-sibling-halo":       "verification_scope_blindspot",
	"undecided-disjunction":               "spec_underspecification",
	"negative-definition":                 "spec_underspecification",
	// Registry classes dropped onto the catch-all (Table B).
	"recurrence-detector-keying": "unclassified",
	"name-keying-vs-marker":      "unclassified",
	"cross-corpus-id-collision":  "unclassified",
	"config-semantics-error":     "unclassified",
	"reflexivity-blindspot":      "unclassified",
	// Slugs a run coined (Table C). `structure-noncompliance` and `invalid-inference` are current
	// words under the fold and have no row.
	"research-protocol-violation":       "self_attestation",
	"uncited-claims":                    "unsourced_claim",
	"unproven-computation":              "unsourced_claim",
	"unsettled-inquiry":                 "claim_contradicts_own_record",
	"contradictory-framing":             "cross_section_contradiction",
	"hypothesis-framing-incomplete":     "unclassified",
	"silent-no-match-probe":             "verification_scope_blindspot",
	"risk-taxonomy-conflation":          "metric_conflation",
	"guardrail-scope-overclaim":         "partial_control_coverage",
	"unflagged-inferential-transfer":    "within_source_condition_misattribution",
	"confinement-scope-gap":             "partial_control_coverage",
	"mitigation-coverage-mismatch":      "partial_control_coverage",
	"mitigation-self-contradiction":     "self_defeating_mitigation",
	"risk-cross-term-ungraded":          "risk_coverage_omission",
	"model-scope-overclaim":             "within_source_condition_misattribution",
	"assertion-without-inquiry":         "unsourced_claim",
	"citation-support-mismatch":         "citation_figure_misattribution",
	"unsourced-quotation-or-appeal":     "unsourced_claim",
	"graded-warrant-contradicts-record": "cross_section_contradiction",
	"unrecorded-report-edit":            "claim_contradicts_own_record",
	"misplaced-citation-anchor":         "citation_figure_misattribution",
	"unsurfaced-inquiry-finding":        "reader_channel_mismatch",
	"nonreproducible-cited-proof":       "artifact_preservation",
	"process-voice-leak":                "report_voice",
	"orphaned-edit-artifact":            "unclassified",
	"inquiry-fate-mismatch":             "claim_contradicts_own_record",
	"record-status-erasure":             "claim_contradicts_own_record",
	"inquiry-reason-record-mismatch":    "claim_contradicts_own_record",
	"dangling-citation-anchor":          "unsourced_claim",
	"citation-identity-mismatch":        "citation_figure_misattribution",
	"non-discriminating-check":          "invalid_inference",
	"summary-scope-omission":            "reader_channel_mismatch",
}

package report

// The external test package (migrateddrift_test.go, which needs migrate, an importer of report)
// reaches the evidence document's renderer and the shape census through these.
var (
	EvidenceDoc  = evidenceDoc
	ResolveMarks = resolveMarks
	Shapes       = shapes
	SameShapes   = sameShapes
)

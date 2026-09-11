package reportproj

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"

// The record's write path sizes a lens's mint budget by the CURRENT report, and cannot import
// this package to render it (this package imports record). So the renderer is handed over here,
// in every binary that links the projection; a binary that does not gets a stated refusal from
// the mint, never a budget read as the floor.
func init() { record.RegisterReportRenderer(RenderFromRecord) }

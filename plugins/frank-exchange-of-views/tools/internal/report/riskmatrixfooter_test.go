package report

import (
	"strings"
	"testing"
)

// The footer STATES the limit, so a reader meets the rule at the place the rule applies
// rather than inferring it from a row that looks unfinished.
func TestRiskMatrixFooterStatesTheLimit(t *testing.T) {
	if !strings.Contains(riskMatrixFooter(), "100") {
		t.Errorf("the matrix footer does not state its truncation limit:\n%s", riskMatrixFooter())
	}
}

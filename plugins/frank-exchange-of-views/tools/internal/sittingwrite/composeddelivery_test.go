package sittingwrite

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE VOICE LENS'S COMPOSED SOURCES ARE ONE ENTRY WITH NO GAP, and one over budget gives way item by
// item from its end, counted: the delivery stays within the harness's cap and says how many it left to
// the work list.
func TestAComposedEntryOverBudgetCutsItemsAndCountsThem(t *testing.T) {
	var items []record.Item
	for i := range 60 {
		items = append(items, record.Item{Kind: record.ItemComposedSource, Subject: &record.ItemSubject{Kind: record.SubjectInquiry, ID: fmt.Sprintf("Q-%08d", i)},
			What: fmt.Sprintf("inquiry Q-%08d prints under Research areas with its fate confirmed, and the reason and the method its record gave beside it are report text — %s", i, strings.Repeat("x", 120))})
	}
	w := record.WorkJSON{Sitting: record.SittingJSON{Seat: "red-lens-voice", Role: "lens", Open: items}}
	text := RenderLensDelivery(w)
	if len(text) > DeliveryBudget {
		t.Fatalf("delivery is %d bytes, over the budget of %d", len(text), DeliveryBudget)
	}
	var d struct {
		Current struct {
			Kind         record.ItemKind `json:"kind"`
			Items        []record.Item   `json:"items"`
			ItemsOmitted int             `json:"items_omitted"`
		} `json:"current"`
	}
	if err := json.Unmarshal([]byte(text[strings.Index(text, "{"):]), &d); err != nil {
		t.Fatal(err)
	}
	if d.Current.Kind != record.ItemComposedSource || len(d.Current.Items) == 0 {
		t.Fatalf("the current entry is not the composed sources: %+v", d.Current)
	}
	if got := len(d.Current.Items) + d.Current.ItemsOmitted; got != len(items) || d.Current.ItemsOmitted == 0 {
		t.Errorf("handed %d items and counted %d omitted, want %d in all with some omitted", len(d.Current.Items), d.Current.ItemsOmitted, len(items))
	}
}

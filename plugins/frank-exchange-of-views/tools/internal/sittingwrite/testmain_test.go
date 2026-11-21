package sittingwrite

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This package's tests open records, so the process must not exit while a cached database handle
// outlives the directory it lived in — invisible on Linux, a Windows-leg failure. See
// recordtest.Main (#666).
//
// With deliveryHolderEnv set, the test binary is instead a delivery-lock holder: it takes the lock
// on the path the variable names, says "held" on stdout, and waits to be killed.
func TestMain(m *testing.M) {
	if path := os.Getenv(deliveryHolderEnv); path != "" {
		if _, err := lockDelivery(path); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Println("held")
		time.Sleep(time.Hour)
		os.Exit(3)
	}
	recordtest.Main(m)
}

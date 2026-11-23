package tessocr

import (
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativehelpertest"
)

// The fakes nativehelpertest builds are removed at exit; see its package comment.
func TestMain(m *testing.M) { os.Exit(nativehelpertest.Run(m)) }

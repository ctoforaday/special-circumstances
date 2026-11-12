// Package vecembedtest points a test at a fake feov-embed: nativehelpertest's fake, built with the
// pins this binary carries, so a ranked search runs end to end without the native build.
package vecembedtest

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativehelpertest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/vecembed"
)

// Spec is the fake feov-embed that passes this binary's handshake: its digest and its asset pins
// for runtime.GOARCH.
func Spec() nativehelpertest.Spec {
	return nativehelpertest.Spec{
		Digest: nativehelper.HelperDigest(vecembed.HelperName),
		Assets: nativehelper.HelperAssets(vecembed.HelperName),
	}
}

// UseFake installs the fake feov-embed in a fresh directory and points NATIVE_HELPER_DIR at it for
// the rest of the test. A test that needs the embedder and cannot build it fails: a skip would
// read as a pass.
func UseFake(t testing.TB) {
	t.Helper()
	dir := t.TempDir()
	nativehelpertest.Install(t, dir, vecembed.HelperName, Spec())
	t.Setenv(nativehelper.EnvDir, dir)
}

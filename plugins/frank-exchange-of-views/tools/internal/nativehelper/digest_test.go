package nativehelper

import (
	"os"
	"reflect"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativesrc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// TestNativeDigestMatchesSources recomputes digest_gen.go from native/ and fails on drift: a
// binary carrying a stale digest refuses the helper built from its own checkout.
func TestNativeDigestMatchesSources(t *testing.T) {
	native, err := repotree.Plugin("native")
	if err != nil {
		t.Fatal(err)
	}
	want, err := nativesrc.Render(native)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile("digest_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(normalizeLF(got)) != string(want) {
		t.Fatalf("digest_gen.go is stale; run `go run ./devcmd/nativedigest` in plugins/frank-exchange-of-views/tools. It should read:\n%s", want)
	}
	hs, err := nativesrc.Read(native)
	if err != nil {
		t.Fatal(err)
	}
	compiled := map[string]Pins{}
	for _, h := range hs {
		p := Pins{Digest: h.Digest}
		for _, a := range h.Assets {
			p.Assets = append(p.Assets, AssetPin{File: a.File, SHA256: a.SHA256})
		}
		compiled[h.Name] = p
	}
	if !reflect.DeepEqual(compiled, helpers) {
		t.Fatalf("the compiled pins %+v are not native/'s %+v", helpers, compiled)
	}
}

func normalizeLF(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] == '\r' && i+1 < len(b) && b[i+1] == '\n' {
			continue
		}
		out = append(out, b[i])
	}
	return out
}

package fetchcache

import "testing"

// A REFUSED FETCH LEAVES A RECORD. The whole of #736's fetch half: before this, a 403 returned an
// error and the index recorded nothing, so "unreachable from this container" survived only as
// prose in a report — which a run then counted by grep and got wrong twice.
func TestARefusedFetchIsRecordedWithWhoRefused(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://proxy.invalid:8080")
	if got := refusalClass(403); got != "unknown" {
		t.Errorf("with a proxy configured a 403 is genuinely ambiguous; want unknown, got %q", got)
	}
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("HTTP_PROXY", "")
	if got := refusalClass(403); got != "origin" {
		t.Errorf("with no proxy the refusal is the source's own; want origin, got %q", got)
	}
}

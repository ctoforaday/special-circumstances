package catalogue

import (
	"os"
	"path/filepath"
	"testing"
)

// THE BOOT IS READ OR IT IS NOT MEASURED — never a zero. Each failure shape must come back as
// ok=false, because a boot of 0 would make every session read as running since the boot and the
// restart would appear to have cut off nothing.
func TestBootIsReadFromTheBtimeLineOrNotMeasured(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	for _, tc := range []struct {
		name string
		path string
		want int64
		ok   bool
	}{
		{"a btime line", write("good", "cpu  1 2 3 4\nintr 5 6\nctxt 99\nbtime 1789058032\nprocesses 7\n"), 1789058032, true},
		{"no btime line", write("none", "cpu  1 2 3 4\nctxt 99\n"), 0, false},
		{"a non-numeric btime", write("nan", "btime soon\n"), 0, false},
		{"a zero btime", write("zero", "btime 0\n"), 0, false},
		{"an unreadable file", filepath.Join(dir, "absent"), 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := bootFrom(tc.path)
			if got != tc.want || ok != tc.ok {
				t.Errorf("bootFrom = (%d, %v), want (%d, %v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}

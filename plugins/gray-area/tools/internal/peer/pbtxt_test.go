package peer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePbtxtTitle(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "real agy format",
			src:  `title:"Finish Stuck Pull Request"  last_user_view_time:{seconds:1794556395  nanos:62000000}`,
			want: "Finish Stuck Pull Request",
		},
		{
			name: "nested title ignored before top-level title",
			src:  "meta { title: \"nested\" }\n# comment title: \"ignored\"\ntitle: \"Top \\\"Level\\\"\" ' Title'\narchived: true\n",
			want: `Top "Level" Title`,
		},
		{
			name: "unicode and octal escapes",
			src:  `title: "\101\x42\u0043"`,
			want: "ABC",
		},
		{
			name: "missing title",
			src:  `archived: false  last_user_view_time: { seconds: 1 }`,
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parsePbtxtTitle(tc.src); got != tc.want {
				t.Errorf("parsePbtxtTitle() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReadAgyTitle(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "0eea6188.pbtxt"), []byte(`title:"Cross Platform Plugin Configuration"  archived:false`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := ReadAgyTitle([]string{dir}, "0eea6188"); got != "Cross Platform Plugin Configuration" {
		t.Errorf("got %q", got)
	}
	if got := ReadAgyTitle([]string{dir}, "missing"); got != "" {
		t.Errorf("missing file returned %q, want empty", got)
	}
	if got := ReadAgyTitle([]string{dir}, "../escape"); got != "" {
		t.Errorf("traversal returned %q, want empty", got)
	}
}

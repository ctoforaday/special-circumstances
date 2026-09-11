package hookgate

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

// LimitReason is what a seat reads when a tool call is refused because its sitting is past the
// run's call limit. The first clause is the instruction; the rest is the fact behind it.
func LimitReason(count, limit int) string {
	return fmt.Sprintf("turn limit reached — return your envelope now. This sitting has made %d tool "+
		"calls against the run's limit of %d, and no further tool call runs in it.", count, limit)
}

// OpensASitting reports whether a Bash call invokes the record tool's `register` verb.
//
// IT IS THE ONE CALL A LIMITED SITTING STILL LETS THROUGH, because register is the sitting
// boundary: it is what starts the next sitting's count. A warm session resumed for a new sitting
// arrives with the old sitting's count still open, and refusing its register would refuse it every
// sitting after the first one it spent.
//
// Recognised from the command text, which is a pattern and says so: a token whose base name is the
// record binary, followed by a `register` token before any shell separator. A miss refuses a real
// register with the limit message; a false match lets one call through that names both.
func OpensASitting(in Input) bool {
	if in.ToolName != "Bash" {
		return false
	}
	var ti toolInput
	if json.Unmarshal(in.ToolInput, &ti) != nil {
		return false
	}
	fields := strings.Fields(ti.Command)
	for i, f := range fields {
		base := path.Base(strings.ReplaceAll(strings.Trim(f, `"'`), `\`, "/"))
		if base != "feov-record" && base != "feov-record.exe" {
			continue
		}
		for _, g := range fields[i+1:] {
			if g == "register" {
				return true
			}
			if strings.ContainsAny(g, ";&|") {
				break
			}
		}
	}
	return false
}

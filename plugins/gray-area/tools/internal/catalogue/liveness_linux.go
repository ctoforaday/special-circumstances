//go:build linux

package catalogue

import (
	"os"
	"strings"
)

// LocalPidDomain composes the same string the client writes: linux:<machine-id>:<pid namespace>.
func LocalPidDomain() string {
	id, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return ""
	}
	ns, err := os.Readlink("/proc/self/ns/pid")
	if err != nil {
		return ""
	}
	return "linux:" + strings.TrimSpace(string(id)) + ":" + ns
}

// ProcStart reads field 22 of /proc/<pid>/stat, the process start time in clock ticks since
// boot. Parsing is by the LAST ')' rather than by splitting on spaces, because field 2 is the
// executable name in parentheses and may itself contain spaces.
func ProcStart(pid int) (string, bool) {
	body, err := os.ReadFile("/proc/" + itoa(pid) + "/stat")
	if err != nil {
		return "", false
	}
	s := string(body)
	i := strings.LastIndex(s, ")")
	if i < 0 {
		return "", false
	}
	fields := strings.Fields(s[i+1:])
	// After the ')' the next field is state (field 3), so field 22 is index 19 here.
	if len(fields) < 20 {
		return "", false
	}
	return fields[19], true
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	p := len(b)
	for i > 0 {
		p--
		b[p] = byte('0' + i%10)
		i /= 10
	}
	return string(b[p:])
}

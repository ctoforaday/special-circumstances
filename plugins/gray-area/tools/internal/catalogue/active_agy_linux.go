//go:build linux

package catalogue

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ActiveAntigravitySessions returns map[conversationID]pid of running Antigravity processes
// that hold an open SQLite database connection to their conversation.
func ActiveAntigravitySessions() map[string]int {
	active := map[string]int{}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return active
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		comm, err := os.ReadFile(filepath.Join("/proc", e.Name(), "comm"))
		if err != nil {
			continue
		}
		commStr := strings.TrimSpace(string(comm))
		if !strings.Contains(commStr, "agy") && !strings.Contains(commStr, "antigravity") {
			continue
		}
		fdDir := filepath.Join("/proc", e.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, f := range fds {
			target, err := os.Readlink(filepath.Join(fdDir, f.Name()))
			if err != nil {
				continue
			}
			if strings.Contains(target, "/conversations/") && (strings.HasSuffix(target, ".db") || strings.HasSuffix(target, ".db-wal")) {
				base := filepath.Base(target)
				for _, ext := range []string{".db-wal", ".db-shm", ".db"} {
					if strings.HasSuffix(base, ext) {
						base = strings.TrimSuffix(base, ext)
						break
					}
				}
				if base != "" && base != "conversation_summaries" {
					active[base] = pid
				}
			}
		}
	}
	return active
}

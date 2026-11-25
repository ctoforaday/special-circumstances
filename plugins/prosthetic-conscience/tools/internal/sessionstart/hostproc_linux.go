//go:build linux

package sessionstart

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// defaultHostProcessID walks /proc ancestors to identify the hosting Antigravity process by
// PID and kernel start time (matching gray-area's peer.ServerIdentity). When no agy ancestor
// is present (direct invocation or tests), it falls back to the immediate parent's PID + start time.
func defaultHostProcessID() string {
	pid := os.Getppid()
	firstPID, firstStart := pid, int64(0)
	for depth := 0; depth < 32 && pid > 1; depth++ {
		ppid, comm, start, ok := readProcStat(pid)
		if !ok {
			break
		}
		if depth == 0 {
			firstStart = start
		}
		if isAgyProcess(pid, comm) {
			return fmt.Sprintf("%d:%d", pid, start)
		}
		if ppid == pid {
			break
		}
		pid = ppid
	}
	return fmt.Sprintf("ppid:%d:%d", firstPID, firstStart)
}

func readProcStat(pid int) (ppid int, comm string, startTime int64, ok bool) {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, "", 0, false
	}
	s := string(b)
	open := strings.IndexByte(s, '(')
	closeIdx := strings.LastIndexByte(s, ')')
	if open < 0 || closeIdx <= open || closeIdx+2 >= len(s) {
		return 0, "", 0, false
	}
	comm = s[open+1 : closeIdx]
	fields := strings.Fields(s[closeIdx+2:])
	if len(fields) <= 19 {
		return 0, "", 0, false
	}
	ppid, err1 := strconv.Atoi(fields[1])
	startTime, err2 := strconv.ParseInt(fields[19], 10, 64)
	if err1 != nil || err2 != nil {
		return 0, "", 0, false
	}
	return ppid, comm, startTime, true
}

func isAgyProcess(pid int, comm string) bool {
	if strings.Contains(comm, "agy") || strings.Contains(comm, "antigravity") {
		return true
	}
	raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil || len(raw) == 0 {
		return false
	}
	arg0, _, _ := bytes.Cut(raw, []byte{0})
	base := filepath.Base(string(arg0))
	return base == "agy" || strings.Contains(base, "antigravity")
}

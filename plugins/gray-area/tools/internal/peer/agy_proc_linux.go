//go:build linux

package peer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const daemonUnit = "antigravity-cli-daemon.service"

// HostServer walks the current process's ancestor chain in /proc to identify the Antigravity
// language-server process hosting this shell or hook.
func HostServer() ServerIdentity {
	pid := os.Getppid()
	for depth := 0; depth < 32 && pid > 1; depth++ {
		ppid, comm, start, ok := readProcStat(pid)
		if !ok {
			break
		}
		if isAgyProcess(pid, comm) {
			return ServerIdentity{
				PID:       pid,
				StartTime: start,
				Daemon:    isDaemonCmdline(pid),
			}
		}
		if ppid == pid {
			break
		}
		pid = ppid
	}
	return ServerIdentity{}
}

// CurrentDaemon returns the identity (MainPID and start time) of the running Antigravity
// remote-control daemon, if one is active. It queries systemd first, then falls back to scanning
// /proc for an `agy remote-control serve` process.
func CurrentDaemon() (ServerIdentity, bool) {
	if pid, ok := systemdDaemonPID(); ok {
		if _, _, start, ok := readProcStat(pid); ok && start > 0 {
			return ServerIdentity{PID: pid, StartTime: start, Daemon: true}, true
		}
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return ServerIdentity{}, false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid <= 1 {
			continue
		}
		if !isDaemonCmdline(pid) {
			continue
		}
		if _, _, start, ok := readProcStat(pid); ok && start > 0 {
			return ServerIdentity{PID: pid, StartTime: start, Daemon: true}, true
		}
	}
	return ServerIdentity{}, false
}

// ProcessStartTime returns the kernel start time (clock ticks since boot) of pid from /proc/<pid>/stat.
func ProcessStartTime(pid int) (int64, bool) {
	if pid <= 0 {
		return 0, false
	}
	_, _, start, ok := readProcStat(pid)
	return start, ok && start > 0
}

// otherDBHolder checks whether ANY live process (regardless of comm) other than excludePID holds
// <agyRoot>/conversations/<conversation>.db (or its -wal/-shm) open.
// Fails CLOSED: if /proc cannot be read, agyRoot cannot be canonicalised, or a same-UID
// process's /proc/<pid>/fd cannot be read (for a reason other than the process exiting), it
// reports busy with the reason.
func otherDBHolder(agyRoot, conversation string, excludePID int) (bool, string) {
	if agyRoot == "" || conversation == "" {
		return true, "conversation root is empty, so open handles cannot be verified"
	}
	canonRoot, err := filepath.EvalSymlinks(agyRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			canonRoot = filepath.Clean(agyRoot)
		} else {
			return true, fmt.Sprintf("cannot resolve conversation root %s: %v", agyRoot, err)
		}
	}
	wantPrefix := filepath.Join(canonRoot, "conversations", conversation+".db")
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return true, fmt.Sprintf("cannot read /proc to verify no second writer: %v", err)
	}
	myUID := uint32(os.Getuid())
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid <= 1 || pid == excludePID {
			continue
		}
		fdDir := filepath.Join("/proc", e.Name(), "fd")
		// Skip fd dirs owned by another user (including root-owned non-dumpable helpers like
		// systemd's (sd-pam), whose /proc/<pid> is ours while /proc/<pid>/fd is root:root 0500).
		info, err := os.Stat(fdDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // process exited
			}
			return true, fmt.Sprintf("cannot stat %s to verify no second writer: %v", fdDir, err)
		}
		if st, ok := info.Sys().(*syscall.Stat_t); ok && st.Uid != myUID {
			continue
		}
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // process exited between stat and ReadDir
			}
			return true, fmt.Sprintf("cannot inspect /proc/%d/fd to verify no second writer: %v", pid, err)
		}
		for _, f := range fds {
			target, err := os.Readlink(filepath.Join(fdDir, f.Name()))
			if err != nil {
				continue // fd closed or process exited
			}
			if target == wantPrefix || target == wantPrefix+"-wal" || target == wantPrefix+"-shm" {
				return true, fmt.Sprintf("process pid %d already has conversations/%s.db open", pid, conversation)
			}
		}
	}
	return false, ""
}

// flockHeldOnlyBy reports whether /proc/locks shows FLOCK holders for path's inode and ALL of them
// are excludePID. Returns false when any other PID holds the lock or /proc/locks cannot confirm.
func flockHeldOnlyBy(path string, excludePID int) bool {
	if excludePID <= 0 {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	raw, err := os.ReadFile("/proc/locks")
	if err != nil {
		return false
	}
	wantIno := strconv.FormatUint(st.Ino, 10)
	sawExclude := false
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		// Format: N: FLOCK ADVISORY WRITE <pid> <maj>:<min>:<ino> 0 EOF
		if len(fields) < 6 || fields[1] != "FLOCK" {
			continue
		}
		devIno := strings.Split(fields[5], ":")
		if len(devIno) != 3 || devIno[2] != wantIno {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err != nil {
			return false
		}
		if pid != excludePID {
			return false
		}
		sawExclude = true
	}
	return sawExclude
}

func systemdDaemonPID() (int, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "systemctl", "--user", "show", "-p", "MainPID", "--value", daemonUnit).Output()
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

// readProcStat parses ppid, comm, and starttime (field 22) from /proc/<pid>/stat.
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
	// fields[0] is state (field 3), fields[1] is ppid (field 4), fields[19] is starttime (field 22).
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

func isDaemonCmdline(pid int) bool {
	raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil || len(raw) == 0 {
		return false
	}
	parts := bytes.Split(raw, []byte{0})
	if len(parts) < 3 {
		return false
	}
	base := filepath.Base(string(parts[0]))
	if base != "agy" && !strings.Contains(base, "antigravity") {
		return false
	}
	hasRC, hasServe := false, false
	for _, p := range parts[1:] {
		switch string(p) {
		case "remote-control":
			hasRC = true
		case "serve":
			hasServe = true
		}
	}
	return hasRC && hasServe
}

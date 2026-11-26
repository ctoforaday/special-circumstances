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
		ppid, comm, start, err := readProcStat(pid)
		if err != nil {
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
		if _, _, start, err := readProcStat(pid); err == nil && start > 0 {
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
		if _, _, start, err := readProcStat(pid); err == nil && start > 0 {
			return ServerIdentity{PID: pid, StartTime: start, Daemon: true}, true
		}
	}
	return ServerIdentity{}, false
}

// ProcessStartTime returns the kernel start time (clock ticks since boot) of pid from /proc/<pid>/stat.
func ProcessStartTime(pid int) (int64, bool) {
	start, err := procStartTime(pid)
	return start, err == nil && start > 0
}

// procStartTime is pid's start time in clock ticks since boot. Its error wraps fs.ErrNotExist only
// when there is no such process; any other failure to read it is a different error, so a caller
// can tell a process that is gone from one it could not look at.
func procStartTime(pid int) (int64, error) {
	if pid <= 0 {
		return 0, fmt.Errorf("pid %d is not a process id", pid)
	}
	_, _, start, err := readProcStat(pid)
	return start, err
}

// otherDBHolder finds a live process (regardless of comm) other than excludePID that holds
// <agyRoot>/conversations/<conversation>.db (or its -wal/-shm) open.
// Fails CLOSED: if /proc cannot be read, agyRoot cannot be canonicalised, or a same-UID
// process's /proc/<pid>/fd cannot be read (for a reason other than the process exiting), it
// reports HolderUninspectable with the reason.
func otherDBHolder(agyRoot, conversation string, excludePID int) Holder {
	uninspectable := func(format string, a ...any) Holder {
		return Holder{Kind: HolderUninspectable, Detail: fmt.Sprintf(format, a...)}
	}
	if agyRoot == "" || conversation == "" {
		return uninspectable("conversation root is empty, so open handles cannot be verified")
	}
	canonRoot, err := filepath.EvalSymlinks(agyRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			canonRoot = filepath.Clean(agyRoot)
		} else {
			return uninspectable("cannot resolve conversation root %s: %v", agyRoot, err)
		}
	}
	convDir := filepath.Join(canonRoot, "conversations")
	canonConvDir, err := filepath.EvalSymlinks(convDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			canonConvDir = convDir
		} else {
			return uninspectable("cannot resolve conversations dir %s: %v", convDir, err)
		}
	}
	wantPrefix := filepath.Join(canonConvDir, conversation+".db")
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return uninspectable("cannot read /proc to verify no second writer: %v", err)
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
		// Skip fd directories not owned by our UID. Because non-dumpable same-UID processes
		// (e.g. systemd's (sd-pam), gpg-agent) have root:root 0500 /proc/<pid>/fd, this skips
		// them as well as other users' processes. Residual: a non-dumpable agy process holding
		// the conversation DB would be missed; agy is dumpable, so the risk is low.
		info, err := os.Stat(fdDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // process exited
			}
			return uninspectable("cannot stat %s to verify no second writer: %v", fdDir, err)
		}
		if st, ok := info.Sys().(*syscall.Stat_t); ok && st.Uid != myUID {
			continue
		}
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // process exited between stat and ReadDir
			}
			return uninspectable("cannot inspect /proc/%d/fd to verify no second writer: %v", pid, err)
		}
		for _, f := range fds {
			target, err := os.Readlink(filepath.Join(fdDir, f.Name()))
			if err != nil {
				continue // fd closed or process exited
			}
			if target == wantPrefix || target == wantPrefix+"-wal" || target == wantPrefix+"-shm" {
				return Holder{PID: pid, Kind: HolderDBOpen, Detail: fmt.Sprintf("pid %d has conversations/%s.db open", pid, conversation)}
			}
		}
	}
	return Holder{Kind: HolderNone}
}

// flockHolders lists the pids /proc/locks shows holding a FLOCK on path's inode, without opening
// path. Its error wraps fs.ErrNotExist when path does not exist.
//
// It refuses, with an error, where /proc/locks may not show the process holding the lock:
//   - a pid it names has exited. /proc/locks names the process that TOOK a flock, and the lock
//     lives as long as any descriptor sharing that open file, so a locker that passed its
//     descriptor on and exited is named while the process now holding the lock is not;
//   - it names no holder while this process runs in a nested pid namespace, where a lock whose
//     owner is outside the namespace is left out of /proc/locks rather than listed;
//   - it names no holder while the file is on a network filesystem, where a lock taken on
//     another host is not in this host's /proc/locks.
func flockHolders(path string) ([]int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("no inode for %s", path)
	}
	raw, err := os.ReadFile("/proc/locks")
	if err != nil {
		return nil, err
	}
	pids, err := lockHolders(string(raw), strconv.FormatUint(st.Ino, 10))
	if err != nil {
		return nil, err
	}
	for _, pid := range pids {
		if _, err := os.Stat(fmt.Sprintf("/proc/%d", pid)); errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("/proc/locks names pid %d, which has exited, so the process now holding the lock is not named", pid)
		}
	}
	if len(pids) == 0 {
		if why := locksBlindSpot(path); why != "" {
			return nil, errors.New(why)
		}
	}
	return pids, nil
}

// locksBlindSpot says why /proc/locks may leave out a lock on path, or "" when it lists every one.
func locksBlindSpot(path string) string {
	status, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return "cannot read /proc/self/status to learn whether /proc/locks lists every holder: " + err.Error()
	}
	nested, known := nestedPidNamespace(string(status))
	switch {
	case !known:
		return "/proc/self/status has no NSpid line, so whether /proc/locks lists every holder is unknown"
	case nested:
		return "this process runs in a nested pid namespace, whose /proc/locks leaves out a holder outside it"
	}
	var sfs syscall.Statfs_t
	if err := syscall.Statfs(filepath.Dir(path), &sfs); err != nil {
		return "cannot learn the filesystem of " + filepath.Dir(path) + ": " + err.Error()
	}
	if name := networkFS(uint32(sfs.Type)); name != "" {
		return "the lock is on " + name + ", where a lock taken on another host is not in this host's /proc/locks"
	}
	return ""
}

// nestedPidNamespace reads /proc/self/status: its NSpid line lists this process's pid in each pid
// namespace from the outermost in, so more than one means a nested namespace. known is false when
// there is no such line.
func nestedPidNamespace(status string) (nested, known bool) {
	for _, line := range strings.Split(status, "\n") {
		if rest, ok := strings.CutPrefix(line, "NSpid:"); ok {
			return len(strings.Fields(rest)) > 1, true
		}
	}
	return false, false
}

// networkFS names the network filesystem with superblock magic m, or "". The magics are
// include/uapi/linux/magic.h's and the filesystems' own headers'.
func networkFS(m uint32) string {
	switch m {
	case 0x6969:
		return "NFS"
	case 0x517B, 0xFF534D42, 0xFE534D42:
		return "SMB"
	case 0x00C36400:
		return "Ceph"
	case 0x5346414F, 0x6B414653:
		return "AFS"
	case 0x01021997:
		return "9P"
	case 0x01161970:
		return "GFS2"
	case 0x7461636F:
		return "OCFS2"
	case 0x0BD00BD0:
		return "Lustre"
	case 0x73757245:
		return "Coda"
	}
	return ""
}

// lockHolders reads /proc/locks text for the FLOCK holders of inode ino.
func lockHolders(locks, ino string) ([]int, error) {
	var pids []int
	for _, line := range strings.Split(locks, "\n") {
		fields := strings.Fields(line)
		// Format: N: FLOCK ADVISORY WRITE <pid> <maj>:<min>:<ino> 0 EOF. A waiter's line reads
		// "N: -> FLOCK ...", so its second field is not FLOCK and it is not a holder.
		if len(fields) < 6 || fields[1] != "FLOCK" {
			continue
		}
		devIno := strings.Split(fields[5], ":")
		if len(devIno) != 3 || devIno[2] != ino {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err != nil {
			return nil, fmt.Errorf("/proc/locks names no pid for the lock: %q", line)
		}
		pids = append(pids, pid)
	}
	return pids, nil
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

// readProcStat parses ppid, comm, and starttime (field 22) from /proc/<pid>/stat. Its error wraps
// fs.ErrNotExist when there is no such process, and is some other error when the file could not be
// read or parsed.
func readProcStat(pid int) (ppid int, comm string, startTime int64, err error) {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, "", 0, err
	}
	malformed := fmt.Errorf("/proc/%d/stat is malformed", pid)
	s := string(b)
	open := strings.IndexByte(s, '(')
	closeIdx := strings.LastIndexByte(s, ')')
	if open < 0 || closeIdx <= open || closeIdx+2 >= len(s) {
		return 0, "", 0, malformed
	}
	comm = s[open+1 : closeIdx]
	fields := strings.Fields(s[closeIdx+2:])
	// fields[0] is state (field 3), fields[1] is ppid (field 4), fields[19] is starttime (field 22).
	if len(fields) <= 19 {
		return 0, "", 0, malformed
	}
	ppid, err1 := strconv.Atoi(fields[1])
	startTime, err2 := strconv.ParseInt(fields[19], 10, 64)
	if err1 != nil || err2 != nil {
		return 0, "", 0, malformed
	}
	return ppid, comm, startTime, nil
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

package main

import "syscall"

// freeBytes is the space available to an unprivileged writer on the filesystem holding path,
// or 0 when it cannot be read — which keeps the gates on disk.
func freeBytes(path string) uint64 {
	var st syscall.Statfs_t
	if syscall.Statfs(path, &st) != nil {
		return 0
	}
	return st.Bavail * uint64(st.Bsize)
}

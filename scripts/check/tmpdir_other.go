//go:build !linux

package main

// freeBytes has nothing to read where there is no /dev/shm; chooseTmp has already declined.
func freeBytes(string) uint64 { return 0 }

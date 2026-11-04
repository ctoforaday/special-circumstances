//go:build !noort && !windows

package main

// #include <stdlib.h>
import "C"

func modelPath(p string) *C.char { return C.CString(p) }

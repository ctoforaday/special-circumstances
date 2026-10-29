//go:build tessocr && cgo && linux

// Package libcprobe probes, at a page boundary, the bounded string scans of the libc the OCR
// engine is statically linked against (the C stack's zig toolchain, third_party/pins). Zig
// 0.16's libc scanned 16 bytes at a time over the caller's whole bound, so a short string in
// the last bytes of a page faulted when the next page was unmapped — leptonica's
// pixMorphSequence reaches it through sscanf on every page (#1598).
//
// It is its own package, not a file in internal/tessocr, because tessocr's Identity() hashes
// every non-test file there and a probe changes no byte a reading produces. The probe is in a
// non-test file because a _test.go file cannot use cgo; nothing outside the test calls it.
package libcprobe

/*
#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>

// Places "d7.1" in the last bytes of a page whose successor is PROT_NONE and runs one
// bounded scan over it, with the bounds musl's sscanf uses. A read past the terminator
// faults; returning is the pass.
static int tessocr_libc_probe(const char *which) {
	long pg = sysconf(_SC_PAGESIZE);
	char *m = mmap(NULL, 2 * pg, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
	if (m == MAP_FAILED) return -1;
	if (mprotect(m + pg, pg, PROT_NONE) != 0) return -1;
	char *op = m + pg - 5;
	memcpy(op, "d7.1", 5);
	volatile size_t sink = 0;
	if (strcmp(which, "sscanf") == 0) {
		int a = 0, b = 0;
		sink = (size_t)sscanf(op + 1, "%d.%d", &a, &b);
	} else if (strcmp(which, "memchr") == 0) {
		sink = (size_t)memchr(op + 1, 0, 257);
	} else if (strcmp(which, "strnlen") == 0) {
		sink = strnlen(op + 1, 257);
	} else if (strcmp(which, "strncpy") == 0) {
		char dst[257];
		strncpy(dst, op + 1, sizeof dst);
		sink = (size_t)dst[0];
	} else {
		return -2;
	}
	(void)sink;
	munmap(m, 2 * pg);
	return 0;
}
*/
import "C"

import "unsafe"

// libcProbeCalls names the scans libcProbe exercises. memccpy is absent: Zig 0.17's still
// scans its whole bound (string.zig:288), and nothing in this C stack calls it.
var libcProbeCalls = []string{"sscanf", "memchr", "strnlen", "strncpy"}

// libcProbe runs one bounded scan over a string that ends at an unmapped page. It returns
// only if the scan stayed inside the string; an overread kills the process.
func libcProbe(which string) int {
	cs := C.CString(which)
	defer C.free(unsafe.Pointer(cs))
	return int(C.tessocr_libc_probe(cs))
}

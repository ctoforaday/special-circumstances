//go:build !noort && windows

package main

/*
#include <stdlib.h>
#include <string.h>
#include <wchar.h>
static wchar_t* widen(const unsigned short* s, int n) {
	wchar_t* w = (wchar_t*)malloc((size_t)(n + 1) * sizeof(wchar_t));
	for (int i = 0; i < n; i++) w[i] = (wchar_t)s[i];
	w[n] = 0;
	return w;
}
*/
import "C"

import (
	"unicode/utf16"
	"unsafe"
)

func modelPath(p string) *C.wchar_t {
	u := utf16.Encode([]rune(p))
	u = append(u, 0)
	return C.widen((*C.ushort)(unsafe.Pointer(&u[0])), C.int(len(u)-1))
}

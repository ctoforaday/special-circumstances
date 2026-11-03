package gte

import "unsafe"

func unsafePtrUint32(p *uint32) unsafe.Pointer { return unsafe.Pointer(p) }
func unsafePtrFloat32(p *float32) unsafe.Pointer { return unsafe.Pointer(p) }

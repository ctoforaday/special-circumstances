//go:build !linux

package catalogue

// BootTime is not measured off Linux, and says so: ok is false, and `telepathy agents --lost`
// then refuses without an explicit --boot rather than guess.
//
// SCOPED DELIBERATELY, as liveness is (liveness_other.go): the full concept additionally needs a
// darwin reader (sysctl kern.boottime) and a windows one (GetTickCount64), named here so the
// omission is tracked rather than discovered.
func BootTime() (int64, bool) { return 0, false }

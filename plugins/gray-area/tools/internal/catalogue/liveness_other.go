//go:build !linux

package catalogue

// LocalPidDomain returns "" off Linux, which makes Of report Unknown for every session.
//
// SCOPED DELIBERATELY. Establishing a durable process identity needs a per-OS probe, and this
// repository has none; the release ships six GOOS/GOARCH pairs. Rather than hand-roll three,
// liveness is exact on Linux and honestly unmeasurable elsewhere. The full concept additionally
// needs liveness_darwin.go and liveness_windows.go, named here so the omission is tracked rather
// than discovered.
func LocalPidDomain() string { return "" }

// ProcStart is never consulted off Linux, because Of short-circuits on the empty domain.
func ProcStart(int) (string, bool) { return "", false }

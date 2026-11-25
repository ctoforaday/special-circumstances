//go:build race

package installshape

// raceBuild is true under -race. The contract test drives subprocesses only, so the detector has
// nothing to instrument, and running it twice per job doubles the client and release-build cost.
const raceBuild = true

package nativehelper

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// The five ways a helper fails, each a distinct error a caller matches with errors.Is. None of
// them ever comes back as an empty result: a missing helper is not a page with no text.
var (
	// ErrAbsent: no executable at the looked-up path, or NATIVE_HELPER_DIR set to none or empty.
	ErrAbsent = errors.New("helper-absent")
	// ErrDigestMismatch: the handshake reported another source digest, or another asset sha256.
	ErrDigestMismatch = errors.New("helper-digest-mismatch")
	// ErrCrashed: the helper's stream ended or the process exited during a request.
	ErrCrashed = errors.New("helper-crashed")
	// ErrDeadline: a request overran its deadline; the helper was killed.
	ErrDeadline = errors.New("helper-deadline")
	// ErrProtocol: a frame the client refuses, or a protocol version other than ProtocolVersion.
	ErrProtocol = errors.New("helper-protocol")
)

// Error carries the details each failure names. Kind is one of the five sentinels above.
type Error struct {
	Kind   error
	Helper string
	// Path is the executable looked up (ErrAbsent) or run.
	Path string
	// Want and Got are the compiled and reported digest or asset sha256 (ErrDigestMismatch).
	Want, Got string
	// What names the compared value: "source digest", or an asset's file name.
	What string
	// Exit is the process's exit status (ErrCrashed).
	Exit string
	// Stderr is the helper's stderr (ErrCrashed): the capture of the request it died in, then what
	// it wrote to the stderr pipe, each its last StderrKeep bytes.
	Stderr []byte
	// Op and Elapsed name the request (ErrDeadline, ErrCrashed, ErrProtocol).
	Op      Op
	Elapsed time.Duration
	// Detail is the rest: the lookup rule, the decoder's complaint.
	Detail string
	// Handshake is set when the helper never came up: the failure came before its handshake was
	// accepted, and the client has latched it.
	Handshake bool
}

func (e *Error) Is(target error) bool { return target == e.Kind }

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s", e.Kind, e.Helper)
	switch e.Kind {
	case ErrAbsent:
		if e.Path != "" {
			fmt.Fprintf(&b, ": no executable at %s", e.Path)
		}
	case ErrDigestMismatch:
		fmt.Fprintf(&b, " at %s: %s %s, this binary was built for %s", e.Path, e.What, e.Got, e.Want)
	case ErrCrashed:
		fmt.Fprintf(&b, " at %s: %s during %s (%s)", e.Path, e.Exit, e.Op, e.Elapsed.Round(time.Millisecond))
	case ErrDeadline:
		fmt.Fprintf(&b, " at %s: %s ran %s, over the deadline; killed", e.Path, e.Op, e.Elapsed.Round(time.Millisecond))
	case ErrProtocol:
		fmt.Fprintf(&b, " at %s: %s", e.Path, e.Op)
	}
	if e.Detail != "" {
		fmt.Fprintf(&b, " — %s", e.Detail)
	}
	if len(e.Stderr) > 0 {
		fmt.Fprintf(&b, "\nhelper stderr:\n%s", strings.TrimRight(string(e.Stderr), "\n"))
	}
	return b.String()
}

// OpError is an op that ran and refused (StatusOpError, StatusUnknownOp). The helper is healthy
// and stays up; it is none of the five failures.
type OpError struct {
	Helper  string
	Op      Op
	Status  Status
	Message string
	Stderr  []byte
}

func (e *OpError) Error() string {
	return fmt.Sprintf("%s: %s: %s: %s", e.Helper, e.Op, e.Status, e.Message)
}

package nativehelper

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// THE WIRE FORMAT, the same bytes native/frame/frame.cpp reads and writes. A change here is a
// change there, and a change to ProtocolVersion in both.
//
// Header, 16 bytes, little-endian:
//
//	offset 0  uint32  Magic ("FVNH" as bytes)
//	offset 4  uint16  protocol version
//	offset 6  uint16  op
//	offset 8  uint32  request id (a response echoes its request's)
//	offset 12 uint32  payload length
//
// Payload: a run of fields, each a uint32 little-endian length followed by that many bytes. The
// fields fill the payload exactly; a short or overlong field is a protocol error.
//
// A response's payload is: field 0, one byte of Status; field 1, the helper's stderr captured
// during the request, bounded by BoundStderr; fields 2 on, the op's result. A status other than
// the three below is a protocol error. A handshake response's result is the helper's
// source digest, then the sha256 of each extra asset it loaded, in its ASSETS file's order.
const (
	Magic           uint32 = 'F' | 'V'<<8 | 'N'<<16 | 'H'<<24
	ProtocolVersion uint16 = 1
	HeaderSize             = 16
	// MaxPayload bounds one frame. The largest request the plan sizes is one 600-DPI page as PNG,
	// about 1.4 MB; the bound refuses a corrupt length before it allocates.
	MaxPayload = 64 << 20
	// MaxFields bounds the fields of one frame, for the same reason.
	MaxFields = 1 << 16
	// StderrKeep bounds a run of helper stderr: the last StderrKeep bytes are kept. native/frame's
	// kStderrKeep is the same number.
	StderrKeep = 64 << 10
)

// BoundStderr is b's last StderrKeep bytes. When anything was cut — the bytes of b beyond the
// bound, plus dropped bytes already cut before b — they are preceded by one line saying how
// many, the same line native/frame writes.
func BoundStderr(b []byte, dropped int64) []byte {
	if len(b) > StderrKeep {
		dropped += int64(len(b) - StderrKeep)
		b = b[len(b)-StderrKeep:]
	}
	if dropped == 0 {
		return b
	}
	return append([]byte(fmt.Sprintf("frame: %d earlier bytes of stderr dropped\n", dropped)), b...)
}

// Op names one request. Ops 0-15 are the protocol's own, answered by every helper; a helper's
// own ops start at FirstHelperOp.
type Op uint16

const (
	OpHandshake      Op = 0
	OpSelftestStdout Op = 1
	OpSelftestSleep  Op = 2
	FirstHelperOp    Op = 16
)

var opNames = map[Op]string{
	OpHandshake:      "handshake",
	OpSelftestStdout: "selftest-stdout",
	OpSelftestSleep:  "selftest-sleep",
}

func (o Op) String() string {
	if n, ok := opNames[o]; ok {
		return n
	}
	return fmt.Sprintf("op-%d", uint16(o))
}

// Status is a response's first field.
type Status byte

const (
	StatusOK Status = 0
	// StatusOpError: the op ran and refused; the result's first field is its message.
	StatusOpError Status = 1
	// StatusUnknownOp: the helper has no such op.
	StatusUnknownOp Status = 2
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusOpError:
		return "op-error"
	case StatusUnknownOp:
		return "unknown-op"
	}
	return fmt.Sprintf("status-%d", byte(s))
}

// Frame is one decoded frame.
type Frame struct {
	Version uint16
	Op      Op
	ID      uint32
	Fields  [][]byte
}

// errFrame marks a malformed frame: the stream is no longer trustworthy.
var errFrame = errors.New("malformed frame")

// EncodeFrame renders f, refusing a frame the decoder would refuse.
func EncodeFrame(f Frame) ([]byte, error) {
	if len(f.Fields) > MaxFields {
		return nil, fmt.Errorf("%w: %d fields, over %d", errFrame, len(f.Fields), MaxFields)
	}
	n := 0
	for _, fl := range f.Fields {
		n += 4 + len(fl)
		if n > MaxPayload {
			return nil, fmt.Errorf("%w: payload over %d bytes", errFrame, MaxPayload)
		}
	}
	out := make([]byte, HeaderSize, HeaderSize+n)
	binary.LittleEndian.PutUint32(out[0:], Magic)
	binary.LittleEndian.PutUint16(out[4:], f.Version)
	binary.LittleEndian.PutUint16(out[6:], uint16(f.Op))
	binary.LittleEndian.PutUint32(out[8:], f.ID)
	binary.LittleEndian.PutUint32(out[12:], uint32(n))
	for _, fl := range f.Fields {
		out = binary.LittleEndian.AppendUint32(out, uint32(len(fl)))
		out = append(out, fl...)
	}
	return out, nil
}

// WriteFrame encodes f and writes it whole.
func WriteFrame(w io.Writer, f Frame) error {
	b, err := EncodeFrame(f)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// ReadFrame reads one frame. It returns io.EOF only when the stream ends cleanly before a frame
// begins; a stream that ends inside a frame is io.ErrUnexpectedEOF; a frame the decoder refuses
// wraps errFrame.
func ReadFrame(r io.Reader) (Frame, error) {
	var h [HeaderSize]byte
	if _, err := io.ReadFull(r, h[:]); err != nil {
		return Frame{}, err
	}
	f, n, err := decodeHeader(h[:])
	if err != nil {
		return Frame{}, err
	}
	payload := make([]byte, n)
	if _, err := io.ReadFull(r, payload); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return Frame{}, err
	}
	f.Fields, err = DecodeFields(payload)
	return f, err
}

// DecodeFrame decodes one whole frame from b, which must hold exactly that frame.
func DecodeFrame(b []byte) (Frame, error) {
	if len(b) < HeaderSize {
		return Frame{}, fmt.Errorf("%w: %d bytes, under the %d-byte header", errFrame, len(b), HeaderSize)
	}
	f, n, err := decodeHeader(b[:HeaderSize])
	if err != nil {
		return Frame{}, err
	}
	if len(b)-HeaderSize != n {
		return Frame{}, fmt.Errorf("%w: header says %d payload bytes, frame holds %d", errFrame, n, len(b)-HeaderSize)
	}
	f.Fields, err = DecodeFields(b[HeaderSize:])
	return f, err
}

func decodeHeader(h []byte) (Frame, int, error) {
	if m := binary.LittleEndian.Uint32(h[0:]); m != Magic {
		return Frame{}, 0, fmt.Errorf("%w: magic %#08x, want %#08x", errFrame, m, Magic)
	}
	n := binary.LittleEndian.Uint32(h[12:])
	if n > MaxPayload {
		return Frame{}, 0, fmt.Errorf("%w: payload length %d, over %d", errFrame, n, MaxPayload)
	}
	return Frame{
		Version: binary.LittleEndian.Uint16(h[4:]),
		Op:      Op(binary.LittleEndian.Uint16(h[6:])),
		ID:      binary.LittleEndian.Uint32(h[8:]),
	}, int(n), nil
}

// DecodeFields splits a payload into its fields; they must fill it exactly.
func DecodeFields(p []byte) ([][]byte, error) {
	fields := [][]byte{}
	for len(p) > 0 {
		if len(fields) == MaxFields {
			return nil, fmt.Errorf("%w: over %d fields", errFrame, MaxFields)
		}
		if len(p) < 4 {
			return nil, fmt.Errorf("%w: %d trailing bytes, under a field's length prefix", errFrame, len(p))
		}
		n := binary.LittleEndian.Uint32(p)
		p = p[4:]
		if uint64(n) > uint64(len(p)) {
			return nil, fmt.Errorf("%w: field of %d bytes, %d remain", errFrame, n, len(p))
		}
		fields = append(fields, p[:n:n])
		p = p[n:]
	}
	return fields, nil
}

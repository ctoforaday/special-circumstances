package nativehelper

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func mustEncode(t testing.TB, f Frame) []byte {
	t.Helper()
	b, err := EncodeFrame(f)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAFrameRoundTrips(t *testing.T) {
	all := make([]byte, 256)
	for i := range all {
		all[i] = byte(i)
	}
	f := Frame{Version: ProtocolVersion, Op: OpSelftestStdout, ID: 7, Fields: [][]byte{{0}, {}, all}}
	b := mustEncode(t, f)
	var w bytes.Buffer
	if err := WriteFrame(&w, f); err != nil || !bytes.Equal(w.Bytes(), b) {
		t.Fatalf("WriteFrame wrote %x, %v; EncodeFrame %x", w.Bytes(), err, b)
	}
	got, err := ReadFrame(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != f.Version || got.Op != f.Op || got.ID != f.ID || len(got.Fields) != 3 || !bytes.Equal(got.Fields[2], all) {
		t.Fatalf("round trip: got %+v", got)
	}
	if _, err := ReadFrame(bytes.NewReader(nil)); err != io.EOF {
		t.Errorf("an empty stream reads %v, want io.EOF", err)
	}
	if _, err := ReadFrame(bytes.NewReader(b[:len(b)-1])); err != io.ErrUnexpectedEOF {
		t.Errorf("a cut frame reads %v, want io.ErrUnexpectedEOF", err)
	}
	bad := append([]byte(nil), b...)
	bad[0] ^= 1
	if _, err := ReadFrame(bytes.NewReader(bad)); !errors.Is(err, errFrame) {
		t.Errorf("a bad magic reads %v, want a malformed frame", err)
	}
	long := append([]byte(nil), b...)
	binary.LittleEndian.PutUint32(long[12:], MaxPayload+1)
	if _, err := ReadFrame(bytes.NewReader(long)); !errors.Is(err, errFrame) {
		t.Errorf("an oversized length reads %v, want a malformed frame", err)
	}
}

// FuzzFrame: whatever bytes arrive, the decoder returns a frame or an error and never panics,
// and a frame it accepts re-encodes to exactly its input — the format has one encoding per frame.
// ReadFrame over the same bytes agrees with DecodeFrame.
func FuzzFrame(f *testing.F) {
	f.Add(mustEncode(f, Frame{Version: ProtocolVersion, Op: OpHandshake, ID: 1}))
	f.Add(mustEncode(f, Frame{Version: ProtocolVersion, Op: OpSelftestSleep, ID: 2, Fields: [][]byte{{10, 0, 0, 0}}}))
	f.Add(mustEncode(f, Frame{Version: 9, Op: 300, ID: 1 << 31, Fields: [][]byte{{1}, []byte("stderr\r\n\x1a"), {}, {}}}))
	f.Add([]byte("FVNH"))
	f.Fuzz(func(t *testing.T, b []byte) {
		fr, err := DecodeFrame(b)
		if err == nil {
			again, err := EncodeFrame(fr)
			if err != nil {
				t.Fatalf("an accepted frame does not re-encode: %v", err)
			}
			if !bytes.Equal(again, b) {
				t.Fatalf("an accepted frame re-encodes to other bytes:\n in  %x\n out %x", b, again)
			}
		}
		rf, rerr := ReadFrame(bytes.NewReader(b))
		if (err == nil) && (rerr != nil) {
			t.Fatalf("DecodeFrame accepts what ReadFrame refuses: %v", rerr)
		}
		if err == nil && (rf.ID != fr.ID || len(rf.Fields) != len(fr.Fields)) {
			t.Fatalf("ReadFrame and DecodeFrame disagree: %+v vs %+v", rf, fr)
		}
		_, _ = DecodeFields(b)
	})
}

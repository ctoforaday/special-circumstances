// frame.h — the native helpers' side of the frame protocol. The Go side is
// tools/internal/nativehelper/frame.go; the two speak the same bytes, and a change to one is a
// change to both and to kVersion.
//
// Header, 16 bytes, little-endian: magic "FVNH", uint16 version, uint16 op, uint32 request id,
// uint32 payload length. Payload: fields, each a uint32 little-endian length then its bytes,
// filling the payload exactly.
//
// A response's payload is field 0, one status byte; field 1, the stderr captured during the
// request, its last kStderrKeep bytes after a line naming how many were dropped; fields 2 on, the
// op's result. The handshake's result is the helper's source digest,
// then the sha256 of each extra asset it loaded, in its ASSETS order.
//
// The helper writes frames on a private descriptor that only this codec writes: claim() moves
// fd 1 there and points fd 1 at fd 2 before any library is initialised, so a library's stdout
// write lands in stderr — and, during a request, in that request's capture — never in a frame.
#ifndef FEOV_NATIVE_FRAME_H
#define FEOV_NATIVE_FRAME_H

#include <cstddef>
#include <cstdint>
#include <functional>
#include <string>
#include <vector>

namespace feov {
namespace frame {

constexpr uint32_t kMagic = 0x484E5646u;  // "FVNH" as little-endian bytes
constexpr uint16_t kVersion = 1;
constexpr size_t kHeaderSize = 16;
constexpr uint32_t kMaxPayload = 64u << 20;
constexpr size_t kMaxFields = 1u << 16;
// kStderrKeep bounds a request's captured stderr. nativehelper.StderrKeep is the same number.
constexpr size_t kStderrKeep = 64u << 10;

// Ops 0-15 are the protocol's own and every helper answers them; a helper's own ops start at
// kFirstHelperOp.
enum Op : uint16_t {
  kHandshake = 0,
  kSelftestStdout = 1,
  kSelftestSleep = 2,
  kFirstHelperOp = 16,
};

enum Status : uint8_t {
  kOK = 0,
  kOpError = 1,    // the op refused; the result's first field is its message
  kUnknownOp = 2,  // no such op
};

struct Frame {
  uint16_t version = kVersion;
  uint16_t op = 0;
  uint32_t id = 0;
  std::vector<std::string> fields;
};

struct Header {
  uint16_t version;
  uint16_t op;
  uint32_t id;
  uint32_t length;
};

// decode_header reads a header from exactly kHeaderSize bytes. On a bad magic or an oversized
// length it returns false and sets *err.
bool decode_header(const uint8_t* p, Header* h, std::string* err);

// decode_fields splits a payload into fields, which must fill it exactly.
bool decode_fields(const uint8_t* p, size_t n, std::vector<std::string>* fields, std::string* err);

// decode decodes one whole frame from exactly n bytes.
bool decode(const uint8_t* p, size_t n, Frame* f, std::string* err);

// encode renders a frame; false when it exceeds the bounds the decoder enforces.
bool encode(const Frame& f, std::string* out, std::string* err);

// claim takes the frame descriptor: it duplicates fd 1 for frames, points fd 1 at fd 2, and on
// Windows sets binary mode on stdin and the frame descriptor. Call it first in main, before any
// library is initialised. False with *err set when a descriptor call fails.
bool claim(std::string* err);

// read_frame reads one request from stdin. It returns 1 for a frame, 0 for a clean EOF before a
// frame begins, -1 for an error (*err set).
int read_frame(Frame* f, std::string* err);

// write_frame writes one frame whole on the frame descriptor.
bool write_frame(const Frame& f, std::string* err);

// Capture points fd 1 and fd 2 together at the capture file for one request, and restores both:
// fd 2 to stderr, fd 1 to fd 2. The capture file is NATIVE_HELPER_CAPTURE when set, else a
// temporary file deleted when the helper exits.
class Capture {
 public:
  bool begin(std::string* err);
  // end restores the descriptors and returns the last kStderrKeep bytes written during the
  // request; when more was written, they follow the line "frame: <n> earlier bytes of stderr
  // dropped".
  std::string end();

 private:
  int saved_err_ = -1;
  int file_ = -1;
};

// A handler answers one of the helper's own ops: it sets *result and returns kOK, or sets
// result to one message and returns kOpError. It returns kUnknownOp for an op it lacks.
using Handler = std::function<Status(uint16_t op, const std::vector<std::string>& args,
                                     std::vector<std::string>* result)>;

// serve is a helper's main loop: claim, then answer the handshake (digest, then assets), the
// self-test ops and the handler's ops until stdin ends, each under a Capture. It returns the
// process exit status: 0 on EOF, non-zero on a protocol or descriptor error.
int serve(const std::string& digest, const std::vector<std::string>& assets, const Handler& handler);

}  // namespace frame
}  // namespace feov

#endif

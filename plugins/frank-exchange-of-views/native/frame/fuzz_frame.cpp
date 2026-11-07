// fuzz_frame.cpp — libFuzzer target for the frame decoder. Built with frame.cpp under
// -fsanitize=fuzzer,address,undefined. The property: a frame the decoder accepts re-encodes to
// exactly its input bytes, since the format has one encoding per frame.
#include <cstddef>
#include <cstdint>
#include <string>
#include <vector>

#include "frame.h"

extern "C" int LLVMFuzzerTestOneInput(const uint8_t* data, size_t size) {
  std::string err;
  feov::frame::Frame f;
  if (feov::frame::decode(data, size, &f, &err)) {
    std::string out;
    if (!feov::frame::encode(f, &out, &err) || out.size() != size ||
        out.compare(0, size, reinterpret_cast<const char*>(data), size) != 0) {
      __builtin_trap();
    }
  }
  std::vector<std::string> fields;
  feov::frame::decode_fields(data, size, &fields, &err);
  if (size >= feov::frame::kHeaderSize) {
    feov::frame::Header h;
    feov::frame::decode_header(data, &h, &err);
  }
  return 0;
}

// main.cpp — the feov-ocr native helper binary.
//
// Speaks the frame protocol on the private frame descriptor (fd 1 claimed at startup),
// initializes tesseract and leptonica over the embedded eng.traineddata, and answers
// OCR and morphology requests until stdin EOF.
#include <cstddef>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <memory>
#include <string>
#include <vector>

#ifdef __linux__
#include <sys/mman.h>
#include <unistd.h>
#endif

#include "../frame/frame.h"
#include "shim.h"

#ifndef FEOV_HELPER_DIGEST
#define FEOV_HELPER_DIGEST ""
#endif

extern const unsigned char g_eng_traineddata[];
extern const size_t g_eng_traineddata_len;

namespace {

enum OCROp : uint16_t {
  kPageTSV = 16,
  kRotatedBand = 17,
  kGridStats = 18,
  kGridLines = 19,
  kRepairedRules = 20,
  kInk = 21,
  kSelftestTraineddata = 22,
  kSelftestLibc = 23,
};

uint32_t get_u32(const std::string& s) {
  if (s.size() < 4) return 0;
  const uint8_t* p = reinterpret_cast<const uint8_t*>(s.data());
  return static_cast<uint32_t>(p[0]) |
         (static_cast<uint32_t>(p[1]) << 8) |
         (static_cast<uint32_t>(p[2]) << 16) |
         (static_cast<uint32_t>(p[3]) << 24);
}

void put_u32(std::string* s, uint32_t v) {
  s->push_back(static_cast<char>(v & 0xff));
  s->push_back(static_cast<char>((v >> 8) & 0xff));
  s->push_back(static_cast<char>((v >> 16) & 0xff));
  s->push_back(static_cast<char>((v >> 24) & 0xff));
}

void put_u64(std::string* s, uint64_t v) {
  for (int i = 0; i < 8; i++) {
    s->push_back(static_cast<char>((v >> (8 * i)) & 0xff));
  }
}

std::string sha256_hex(const uint8_t* data, size_t len) {
  uint32_t h[8] = {
      0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
      0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
  };
  static const uint32_t k[64] = {
      0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
      0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
      0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
      0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
      0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
      0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
      0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
      0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
  };
  auto rotr = [](uint32_t x, uint32_t n) { return (x >> n) | (x << (32 - n)); };
  auto ch = [](uint32_t x, uint32_t y, uint32_t z) { return (x & y) ^ (~x & z); };
  auto maj = [](uint32_t x, uint32_t y, uint32_t z) { return (x & y) ^ (x & z) ^ (y & z); };
  auto sig0 = [&](uint32_t x) { return rotr(x, 2) ^ rotr(x, 13) ^ rotr(x, 22); };
  auto sig1 = [&](uint32_t x) { return rotr(x, 6) ^ rotr(x, 11) ^ rotr(x, 25); };
  auto gam0 = [&](uint32_t x) { return rotr(x, 7) ^ rotr(x, 18) ^ (x >> 3); };
  auto gam1 = [&](uint32_t x) { return rotr(x, 17) ^ rotr(x, 19) ^ (x >> 10); };

  std::vector<uint8_t> msg(data, data + len);
  uint64_t total_bits = static_cast<uint64_t>(len) * 8;
  msg.push_back(0x80);
  while ((msg.size() % 64) != 56) msg.push_back(0);
  for (int i = 7; i >= 0; i--) msg.push_back(static_cast<uint8_t>((total_bits >> (i * 8)) & 0xff));

  for (size_t chunk = 0; chunk < msg.size(); chunk += 64) {
    uint32_t w[64];
    for (int i = 0; i < 16; i++) {
      w[i] = (static_cast<uint32_t>(msg[chunk + i * 4]) << 24) |
             (static_cast<uint32_t>(msg[chunk + i * 4 + 1]) << 16) |
             (static_cast<uint32_t>(msg[chunk + i * 4 + 2]) << 8) |
             (static_cast<uint32_t>(msg[chunk + i * 4 + 3]));
    }
    for (int i = 16; i < 64; i++) {
      w[i] = gam1(w[i - 2]) + w[i - 7] + gam0(w[i - 15]) + w[i - 16];
    }
    uint32_t a = h[0], b = h[1], c = h[2], d = h[3], e = h[4], f = h[5], g = h[6], hv = h[7];
    for (int i = 0; i < 64; i++) {
      uint32_t t1 = hv + sig1(e) + ch(e, f, g) + k[i] + w[i];
      uint32_t t2 = sig0(a) + maj(a, b, c);
      hv = g;
      g = f;
      f = e;
      e = d + t1;
      d = c;
      c = b;
      b = a;
      a = t1 + t2;
    }
    h[0] += a;
    h[1] += b;
    h[2] += c;
    h[3] += d;
    h[4] += e;
    h[5] += f;
    h[6] += g;
    h[7] += hv;
  }
  char hex_buf[65];
  for (int i = 0; i < 8; i++) {
    std::snprintf(hex_buf + i * 8, 9, "%08x", h[i]);
  }
  return std::string(hex_buf, 64);
}

#ifdef __linux__
int run_one_libc_probe(const char* which) {
  long pg = sysconf(_SC_PAGESIZE);
  char* m = static_cast<char*>(mmap(NULL, 2 * pg, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0));
  if (m == MAP_FAILED) return -1;
  if (mprotect(m + pg, pg, PROT_NONE) != 0) return -1;
  char* op = m + pg - 5;
  std::memcpy(op, "d7.1", 5);
  volatile size_t sink = 0;
  if (std::strcmp(which, "sscanf") == 0) {
    int a = 0, b = 0;
    sink = static_cast<size_t>(std::sscanf(op + 1, "%d.%d", &a, &b));
  } else if (std::strcmp(which, "memchr") == 0) {
    sink = reinterpret_cast<size_t>(std::memchr(op + 1, 0, 257));
  } else if (std::strcmp(which, "strnlen") == 0) {
    sink = strnlen(op + 1, 257);
  } else if (std::strcmp(which, "strncpy") == 0) {
    char dst[257];
    std::strncpy(dst, op + 1, sizeof dst);
    sink = static_cast<size_t>(dst[0]);
  } else {
    munmap(m, 2 * pg);
    return -2;
  }
  (void)sink;
  munmap(m, 2 * pg);
  return 0;
}

bool run_libc_probes(std::string* err) {
  static const char* kProbes[] = {"sscanf", "memchr", "strnlen", "strncpy"};
  for (const char* p : kProbes) {
    int rc = run_one_libc_probe(p);
    if (rc != 0) {
      *err = std::string("libc probe ") + p + " failed: " + std::to_string(rc);
      return false;
    }
  }
  return true;
}
#endif

}  // namespace

int main(int argc, char** argv) {
  (void)argc;
  (void)argv;

  tessocr_engine* engine = tessocr_new(g_eng_traineddata, static_cast<int>(g_eng_traineddata_len));
  if (engine == nullptr) {
    std::fprintf(stderr, "feov-ocr: tesseract init from embedded traineddata failed\n");
    return 1;
  }

  auto handler = [engine](uint16_t op, const std::vector<std::string>& args,
                          std::vector<std::string>* result) -> feov::frame::Status {
    switch (op) {
      case kPageTSV: {
        if (args.size() != 2 || args[1].size() != 4) {
          result->push_back("tsv takes png and 4-byte psm");
          return feov::frame::kOpError;
        }
        int psm = static_cast<int>(get_u32(args[1]));
        char* text = tessocr_tsv(engine, reinterpret_cast<const unsigned char*>(args[0].data()), args[0].size(), psm);
        if (!text) {
          result->push_back("tsv failed (image did not decode or OCR returned nothing)");
          return feov::frame::kOpError;
        }
        result->emplace_back(text);
        tessocr_free_text(text);
        return feov::frame::kOK;
      }
      case kRotatedBand: {
        if (args.size() != 5 || args[1].size() != 4 || args[2].size() != 4 || args[3].size() != 4 || args[4].size() != 4) {
          result->push_back("rotated band takes png and 4 4-byte coords");
          return feov::frame::kOpError;
        }
        int x = static_cast<int>(get_u32(args[1]));
        int y = static_cast<int>(get_u32(args[2]));
        int w = static_cast<int>(get_u32(args[3]));
        int h = static_cast<int>(get_u32(args[4]));
        char* text = tessocr_rot_band(engine, reinterpret_cast<const unsigned char*>(args[0].data()), args[0].size(), x, y, w, h);
        if (!text) {
          result->push_back("rotated band failed (image did not decode or OCR returned nothing)");
          return feov::frame::kOpError;
        }
        result->emplace_back(text);
        tessocr_free_text(text);
        return feov::frame::kOK;
      }
      case kGridStats: {
        if (args.size() != 2 || args[1].size() != 4) {
          result->push_back("grid stats takes png and 4-byte sel");
          return feov::frame::kOpError;
        }
        int sel = static_cast<int>(get_u32(args[1]));
        int h = 0, v = 0, ix = 0;
        int rc = tessocr_grid_stats(reinterpret_cast<const unsigned char*>(args[0].data()), args[0].size(), sel, &h, &v, &ix);
        if (rc != 0) {
          result->push_back("grid detection failed at stage " + std::to_string(rc));
          return feov::frame::kOpError;
        }
        std::string sh, sv, six;
        put_u32(&sh, static_cast<uint32_t>(h));
        put_u32(&sv, static_cast<uint32_t>(v));
        put_u32(&six, static_cast<uint32_t>(ix));
        result->push_back(std::move(sh));
        result->push_back(std::move(sv));
        result->push_back(std::move(six));
        return feov::frame::kOK;
      }
      case kGridLines: {
        if (args.size() != 2 || args[1].size() != 4) {
          result->push_back("grid lines takes png and 4-byte sel");
          return feov::frame::kOpError;
        }
        int sel = static_cast<int>(get_u32(args[1]));
        char* text = tessocr_grid_lines(reinterpret_cast<const unsigned char*>(args[0].data()), args[0].size(), sel);
        if (!text) {
          result->push_back("grid geometry failed (decode, binarize or open)");
          return feov::frame::kOpError;
        }
        result->emplace_back(text);
        tessocr_free_text(text);
        return feov::frame::kOK;
      }
      case kRepairedRules: {
        if (args.size() != 4 || args[1].size() != 4 || args[2].size() != 4 || args[3].size() != 4) {
          result->push_back("repaired rules takes png, close_h, close_v, sel");
          return feov::frame::kOpError;
        }
        int ch = static_cast<int>(get_u32(args[1]));
        int cv = static_cast<int>(get_u32(args[2]));
        int sel = static_cast<int>(get_u32(args[3]));
        char* text = tessocr_repaired_rules(reinterpret_cast<const unsigned char*>(args[0].data()), args[0].size(), ch, cv, sel);
        if (!text) {
          result->push_back("repaired rule geometry failed (decode, binarize, close or open)");
          return feov::frame::kOpError;
        }
        result->emplace_back(text);
        tessocr_free_text(text);
        return feov::frame::kOK;
      }
      case kInk: {
        if (args.size() != 1) {
          result->push_back("ink takes png");
          return feov::frame::kOpError;
        }
        long long dark = 0, total = 0;
        int rc = tessocr_ink(reinterpret_cast<const unsigned char*>(args[0].data()), args[0].size(), &dark, &total);
        if (rc != 0) {
          result->push_back("ink measurement failed at stage " + std::to_string(rc));
          return feov::frame::kOpError;
        }
        if (total <= 0) {
          result->push_back("ink measurement over an image with no pixels");
          return feov::frame::kOpError;
        }
        std::string sdark, stotal;
        put_u64(&sdark, static_cast<uint64_t>(dark));
        put_u64(&stotal, static_cast<uint64_t>(total));
        result->push_back(std::move(sdark));
        result->push_back(std::move(stotal));
        return feov::frame::kOK;
      }
      case kSelftestTraineddata: {
        result->push_back(sha256_hex(g_eng_traineddata, g_eng_traineddata_len));
        return feov::frame::kOK;
      }
      case kSelftestLibc: {
#ifdef __linux__
        std::string err;
        if (!run_libc_probes(&err)) {
          result->push_back(err);
          return feov::frame::kOpError;
        }
        result->push_back("ok");
        return feov::frame::kOK;
#else
        result->push_back("selftest-libc is Linux only");
        return feov::frame::kOpError;
#endif
      }
      default:
        return feov::frame::kUnknownOp;
    }
  };

  int exit_code = feov::frame::serve(FEOV_HELPER_DIGEST, {}, handler);
  tessocr_free(engine);
  return exit_code;
}

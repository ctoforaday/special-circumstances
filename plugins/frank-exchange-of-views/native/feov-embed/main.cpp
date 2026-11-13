// main.cpp — the feov-embed native helper binary.
//
// Speaks the frame protocol on the private frame descriptor (fd 1 claimed at startup),
// initializes ONNX Runtime 1.31.1 minimal build over gte-small.int8.ort, and answers
// vector embedding requests until stdin EOF.
#include <cmath>
#include <cstddef>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <fstream>
#include <memory>
#include <string>
#include <vector>

#ifdef _WIN32
#include <windows.h>
#else
#include <unistd.h>
#endif

#include "../frame/frame.h"
#include "onnxruntime_c_api.h"

#ifndef FEOV_HELPER_DIGEST
#define FEOV_HELPER_DIGEST ""
#endif

namespace {

enum EmbedOp : uint16_t {
  kEmbed = 16,
};

constexpr int kEmbedDim = 384;

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
    uint32_t a = h[0], b = h[1], c = h[2], d = h[3];
    uint32_t e = h[4], f = h[5], g = h[6], cur_h = h[7];
    for (int i = 0; i < 64; i++) {
      uint32_t t1 = cur_h + sig1(e) + ch(e, f, g) + k[i] + w[i];
      uint32_t t2 = sig0(a) + maj(a, b, c);
      cur_h = g;
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
    h[7] += cur_h;
  }
  char hex_str[65];
  for (int i = 0; i < 8; i++) {
    std::snprintf(hex_str + i * 8, 9, "%08x", h[i]);
  }
  return std::string(hex_str, 64);
}

std::string get_executable_dir(const char* argv0) {
#ifdef _WIN32
  char buf[MAX_PATH];
  DWORD len = GetModuleFileNameA(NULL, buf, MAX_PATH);
  if (len > 0 && len < MAX_PATH) {
    std::string p(buf, len);
    size_t last_slash = p.find_last_of("\\/");
    if (last_slash != std::string::npos) {
      return p.substr(0, last_slash);
    }
  }
#else
  char buf[4096];
  ssize_t len = readlink("/proc/self/exe", buf, sizeof(buf) - 1);
  if (len > 0) {
    buf[len] = '\0';
    std::string p(buf);
    size_t last_slash = p.find_last_of('/');
    if (last_slash != std::string::npos) {
      return p.substr(0, last_slash);
    }
  }
#endif
  if (argv0 != nullptr) {
    std::string p(argv0);
    size_t last_slash = p.find_last_of("\\/");
    if (last_slash != std::string::npos) {
      return p.substr(0, last_slash);
    }
  }
  return ".";
}

bool read_entire_file(const std::string& path, std::vector<uint8_t>* out) {
  std::ifstream f(path, std::ios::binary | std::ios::ate);
  if (!f.is_open()) return false;
  std::streamsize size = f.tellg();
  if (size <= 0) return false;
  f.seekg(0, std::ios::beg);
  out->resize(static_cast<size_t>(size));
  if (!f.read(reinterpret_cast<char*>(out->data()), size)) {
    out->clear();
    return false;
  }
  return true;
}

}  // namespace

int main(int argc, char** argv) {
  const char* argv0 = (argc > 0) ? argv[0] : nullptr;
  std::vector<std::string> candidates;
  const char* env_dir = std::getenv("NATIVE_HELPER_DIR");
  if (env_dir != nullptr && *env_dir != '\0') {
    std::string ed(env_dir);
    if (ed.back() == '/' || ed.back() == '\\') ed.pop_back();
    candidates.push_back(ed + "/gte-small.int8.ort");
  }
  std::string exe_dir = get_executable_dir(argv0);
  candidates.push_back(exe_dir + "/gte-small.int8.ort");
  candidates.push_back("gte-small.int8.ort");

  std::vector<uint8_t> model_data;
  std::string loaded_path;
  for (const auto& path : candidates) {
    if (read_entire_file(path, &model_data)) {
      loaded_path = path;
      break;
    }
  }
  if (model_data.empty()) {
    std::fprintf(stderr, "feov-embed: could not load model gte-small.int8.ort (tried: ");
    for (size_t i = 0; i < candidates.size(); ++i) {
      std::fprintf(stderr, "%s%s", (i > 0 ? ", " : ""), candidates[i].c_str());
    }
    std::fprintf(stderr, ")\n");
    return 1;
  }

  std::string model_sha = sha256_hex(model_data.data(), model_data.size());

  const OrtApi* g_ort = OrtGetApiBase()->GetApi(ORT_API_VERSION);
  if (g_ort == nullptr) {
    std::fprintf(stderr, "feov-embed: OrtGetApiBase()->GetApi returned NULL\n");
    return 1;
  }

  OrtEnv* env = nullptr;
  OrtStatus* st = g_ort->CreateEnv(ORT_LOGGING_LEVEL_WARNING, "feov-embed", &env);
  if (st != nullptr) {
    std::fprintf(stderr, "feov-embed: CreateEnv failed: %s\n", g_ort->GetErrorMessage(st));
    g_ort->ReleaseStatus(st);
    return 1;
  }

  OrtSessionOptions* so = nullptr;
  st = g_ort->CreateSessionOptions(&so);
  if (st != nullptr) {
    std::fprintf(stderr, "feov-embed: CreateSessionOptions failed: %s\n", g_ort->GetErrorMessage(st));
    g_ort->ReleaseStatus(st);
    return 1;
  }

  if (OrtStatus* s = g_ort->SetIntraOpNumThreads(so, 1)) { g_ort->ReleaseStatus(s); }
  if (OrtStatus* s = g_ort->SetInterOpNumThreads(so, 1)) { g_ort->ReleaseStatus(s); }
  if (OrtStatus* s = g_ort->SetSessionExecutionMode(so, ORT_SEQUENTIAL)) { g_ort->ReleaseStatus(s); }
  if (OrtStatus* s = g_ort->SetSessionGraphOptimizationLevel(so, ORT_ENABLE_ALL)) { g_ort->ReleaseStatus(s); }

  OrtSession* sess = nullptr;
  st = g_ort->CreateSessionFromArray(env, model_data.data(), model_data.size(), so, &sess);
  g_ort->ReleaseSessionOptions(so);
  if (st != nullptr) {
    std::fprintf(stderr, "feov-embed: CreateSessionFromArray failed: %s\n", g_ort->GetErrorMessage(st));
    g_ort->ReleaseStatus(st);
    return 1;
  }

  OrtMemoryInfo* mem = nullptr;
  st = g_ort->CreateCpuMemoryInfo(OrtArenaAllocator, OrtMemTypeDefault, &mem);
  if (st != nullptr) {
    std::fprintf(stderr, "feov-embed: CreateCpuMemoryInfo failed: %s\n", g_ort->GetErrorMessage(st));
    g_ort->ReleaseStatus(st);
    return 1;
  }

  auto handler = [g_ort, sess, mem](uint16_t op, const std::vector<std::string>& args,
                                    std::vector<std::string>* result) -> feov::frame::Status {
    if (op != kEmbed) {
      return feov::frame::kUnknownOp;
    }
    if (args.empty() || (args.size() % 2 != 0)) {
      result->push_back("feov-embed: embed request requires non-empty 2N arguments (ids and mask pairs)");
      return feov::frame::kOpError;
    }

    size_t num_texts = args.size() / 2;
    result->reserve(num_texts);
    for (size_t i = 0; i < num_texts; ++i) {
      const std::string& ids_bytes = args[2 * i];
      const std::string& mask_bytes = args[2 * i + 1];
      if (ids_bytes.size() != mask_bytes.size() || ids_bytes.size() % sizeof(int64_t) != 0) {
        result->clear();
        result->push_back("feov-embed: ids and mask must have identical length and be multiples of 8 bytes");
        return feov::frame::kOpError;
      }
      size_t n_tokens = ids_bytes.size() / sizeof(int64_t);
      if (n_tokens == 0) {
        result->clear();
        result->push_back("feov-embed: cannot embed 0 tokens");
        return feov::frame::kOpError;
      }

      int64_t shape[2] = {1, static_cast<int64_t>(n_tokens)};
      OrtValue* in[2] = {nullptr, nullptr};
      OrtStatus* op_st;

      op_st = g_ort->CreateTensorWithDataAsOrtValue(
          mem, const_cast<char*>(ids_bytes.data()), ids_bytes.size(), shape, 2,
          ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &in[0]);
      if (op_st != nullptr) {
        std::string err_msg = std::string("feov-embed: tensor create failed: ") + g_ort->GetErrorMessage(op_st);
        g_ort->ReleaseStatus(op_st);
        result->clear();
        result->push_back(std::move(err_msg));
        return feov::frame::kOpError;
      }

      op_st = g_ort->CreateTensorWithDataAsOrtValue(
          mem, const_cast<char*>(mask_bytes.data()), mask_bytes.size(), shape, 2,
          ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &in[1]);
      if (op_st != nullptr) {
        std::string err_msg = std::string("feov-embed: tensor create failed: ") + g_ort->GetErrorMessage(op_st);
        g_ort->ReleaseStatus(op_st);
        g_ort->ReleaseValue(in[0]);
        result->clear();
        result->push_back(std::move(err_msg));
        return feov::frame::kOpError;
      }

      const char* inames[2] = {"input_ids", "attention_mask"};
      const char* onames[1] = {"last_hidden_state"};
      OrtValue* out_val = nullptr;

      op_st = g_ort->Run(sess, nullptr, inames, const_cast<const OrtValue**>(in), 2, onames, 1, &out_val);
      g_ort->ReleaseValue(in[0]);
      g_ort->ReleaseValue(in[1]);
      if (op_st != nullptr) {
        std::string err_msg = std::string("feov-embed: run failed: ") + g_ort->GetErrorMessage(op_st);
        g_ort->ReleaseStatus(op_st);
        result->clear();
        result->push_back(std::move(err_msg));
        return feov::frame::kOpError;
      }

      float* hidden = nullptr;
      op_st = g_ort->GetTensorMutableData(out_val, reinterpret_cast<void**>(&hidden));
      if (op_st != nullptr) {
        std::string err_msg = std::string("feov-embed: get data failed: ") + g_ort->GetErrorMessage(op_st);
        g_ort->ReleaseStatus(op_st);
        g_ort->ReleaseValue(out_val);
        result->clear();
        result->push_back(std::move(err_msg));
        return feov::frame::kOpError;
      }

      // Mean pool over attention mask:
      const int64_t* mask_arr = reinterpret_cast<const int64_t*>(mask_bytes.data());
      double sum[kEmbedDim] = {0.0};
      int64_t mask_sum = 0;
      for (size_t t = 0; t < n_tokens; ++t) {
        if (mask_arr[t] != 0) {
          mask_sum++;
          for (int d = 0; d < kEmbedDim; ++d) {
            sum[d] += static_cast<double>(hidden[t * kEmbedDim + d]);
          }
        }
      }
      g_ort->ReleaseValue(out_val);

      if (mask_sum == 0) mask_sum = 1;
      double norm_sq = 0.0;
      for (int d = 0; d < kEmbedDim; ++d) {
        sum[d] /= static_cast<double>(mask_sum);
        norm_sq += sum[d] * sum[d];
      }
      double norm = std::sqrt(norm_sq);

      std::string emb_res(kEmbedDim * sizeof(float), '\0');
      float* emb_out = reinterpret_cast<float*>(&emb_res[0]);
      for (int d = 0; d < kEmbedDim; ++d) {
        emb_out[d] = (norm > 0.0) ? static_cast<float>(sum[d] / norm) : 0.0f;
      }
      result->push_back(std::move(emb_res));
    }
    return feov::frame::kOK;
  };

  int exit_code = feov::frame::serve(FEOV_HELPER_DIGEST, {model_sha}, handler);

  g_ort->ReleaseMemoryInfo(mem);
  g_ort->ReleaseSession(sess);
  g_ort->ReleaseEnv(env);
  return exit_code;
}

// embed-probe: a feasibility probe for a static Windows embedding helper over ONNX Runtime's C API.
//
//   embed-probe <model.ort> [threads=1]  < ids.txt  > vectors.f32
//
// stdin: one sequence per line, token ids as decimal integers separated by spaces (the tokenizer runs elsewhere).
// stdout: per sequence, dim float32 values (little-endian), mean-pooled over the attention mask and L2-normalized.
// stderr: one timing line; the first three sequences are embedded once untimed as a warm-up.
#include <onnxruntime_c_api.h>

#include <chrono>
#include <cmath>
#include <cstdio>
#include <cstdlib>
#include <iostream>
#include <sstream>
#include <string>
#include <vector>
#ifdef _WIN32
#include <fcntl.h>
#include <io.h>
#endif

static const OrtApi* g;

static void ck(OrtStatus* st) {
  if (!st) return;
  std::fprintf(stderr, "embed-probe: %s\n", g->GetErrorMessage(st));
  std::exit(1);
}

struct Engine {
  OrtEnv* env = nullptr;
  OrtSession* sess = nullptr;
  OrtMemoryInfo* mem = nullptr;

  Engine(const ORTCHAR_T* model, int threads) {
    ck(g->CreateEnv(ORT_LOGGING_LEVEL_WARNING, "embed-probe", &env));
    OrtSessionOptions* so;
    ck(g->CreateSessionOptions(&so));
    ck(g->SetIntraOpNumThreads(so, threads));
    ck(g->SetInterOpNumThreads(so, 1));
    ck(g->SetSessionExecutionMode(so, ORT_SEQUENTIAL));
    ck(g->SetSessionGraphOptimizationLevel(so, ORT_ENABLE_ALL));
    ck(g->CreateSession(env, model, so, &sess));
    g->ReleaseSessionOptions(so);
    ck(g->CreateCpuMemoryInfo(OrtArenaAllocator, OrtMemTypeDefault, &mem));
  }

  // Returns the pooled, normalized vector.
  std::vector<float> embed(std::vector<int64_t>& ids) {
    int64_t n = (int64_t)ids.size();
    std::vector<int64_t> mask(ids.size(), 1);
    int64_t shape[2] = {1, n};
    OrtValue* in[2] = {nullptr, nullptr};
    OrtValue* out = nullptr;
    ck(g->CreateTensorWithDataAsOrtValue(mem, ids.data(), n * sizeof(int64_t), shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &in[0]));
    ck(g->CreateTensorWithDataAsOrtValue(mem, mask.data(), n * sizeof(int64_t), shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &in[1]));
    const char* inames[2] = {"input_ids", "attention_mask"};
    const char* onames[1] = {"last_hidden_state"};
    ck(g->Run(sess, nullptr, inames, in, 2, onames, 1, &out));
    OrtTensorTypeAndShapeInfo* info;
    ck(g->GetTensorTypeAndShape(out, &info));
    int64_t dims[3] = {0, 0, 0};
    ck(g->GetDimensions(info, dims, 3));
    g->ReleaseTensorTypeAndShapeInfo(info);
    int64_t dim = dims[2];
    float* h;
    ck(g->GetTensorMutableData(out, (void**)&h));
    std::vector<double> acc((size_t)dim, 0.0);
    double cnt = 0;
    for (int64_t i = 0; i < n; i++) {
      if (!mask[(size_t)i]) continue;
      cnt += 1;
      for (int64_t d = 0; d < dim; d++) acc[(size_t)d] += h[i * dim + d];
    }
    double norm = 0;
    for (auto& v : acc) { v /= cnt; norm += v * v; }
    norm = std::sqrt(norm);
    std::vector<float> r((size_t)dim);
    for (int64_t d = 0; d < dim; d++) r[(size_t)d] = norm > 0 ? (float)(acc[(size_t)d] / norm) : 0.f;
    g->ReleaseValue(out);
    g->ReleaseValue(in[0]);
    g->ReleaseValue(in[1]);
    return r;
  }
};

#ifdef _WIN32
int wmain(int argc, wchar_t** argv) {
  _setmode(_fileno(stdout), _O_BINARY);
  int threads = argc > 2 ? _wtoi(argv[2]) : 1;
#else
int main(int argc, char** argv) {
  int threads = argc > 2 ? std::atoi(argv[2]) : 1;
#endif
  if (argc < 2) {
    std::fprintf(stderr, "usage: embed-probe <model.ort> [threads] < ids.txt > vectors.f32\n");
    return 2;
  }
  g = OrtGetApiBase()->GetApi(ORT_API_VERSION);
  if (!g) { std::fprintf(stderr, "embed-probe: GetApi returned NULL\n"); return 1; }
  std::vector<std::vector<int64_t>> seqs;
  std::string line;
  while (std::getline(std::cin, line)) {
    std::istringstream s(line);
    std::vector<int64_t> ids;
    long long v;
    while (s >> v) ids.push_back(v);
    if (!ids.empty()) seqs.push_back(std::move(ids));
  }
  auto t0 = std::chrono::steady_clock::now();
  Engine e(argv[1], threads);
  double load = std::chrono::duration<double>(std::chrono::steady_clock::now() - t0).count();
  for (size_t i = 0; i < 3 && i < seqs.size(); i++) e.embed(seqs[i]);
  std::vector<float> all;
  auto t1 = std::chrono::steady_clock::now();
  size_t dim = 0;
  for (auto& ids : seqs) {
    auto r = e.embed(ids);
    dim = r.size();
    all.insert(all.end(), r.begin(), r.end());
  }
  double secs = std::chrono::duration<double>(std::chrono::steady_clock::now() - t1).count();
  std::fwrite(all.data(), sizeof(float), all.size(), stdout);
  std::fflush(stdout);
  std::fprintf(stderr, "load=%.2fs n=%zu threads=%d secs=%.2f per_s=%.2f ms_each=%.1f dim=%zu\n", load, seqs.size(), threads,
               secs, seqs.size() / secs, secs * 1000 / seqs.size(), dim);
  return 0;
}

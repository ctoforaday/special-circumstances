// frame.cpp — the frame codec, the frame descriptor, the per-request capture and the main loop
// every native helper runs. See frame.h for the wire format.
#include "frame.h"

#include <cerrno>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <iostream>

#ifdef _WIN32
#include <fcntl.h>
#include <io.h>
#include <windows.h>
#else
#include <fcntl.h>
#include <time.h>
#include <unistd.h>
#endif

namespace feov {
namespace frame {

namespace {

uint16_t get16(const uint8_t* p) { return static_cast<uint16_t>(p[0] | (p[1] << 8)); }

uint32_t get32(const uint8_t* p) {
  return static_cast<uint32_t>(p[0]) | (static_cast<uint32_t>(p[1]) << 8) |
         (static_cast<uint32_t>(p[2]) << 16) | (static_cast<uint32_t>(p[3]) << 24);
}

void put16(std::string* s, uint16_t v) {
  s->push_back(static_cast<char>(v & 0xff));
  s->push_back(static_cast<char>(v >> 8));
}

void put32(std::string* s, uint32_t v) {
  for (int i = 0; i < 4; i++) s->push_back(static_cast<char>((v >> (8 * i)) & 0xff));
}

#ifdef _WIN32
HANDLE g_frame = INVALID_HANDLE_VALUE;
#else
int g_frame = -1;
#endif
bool g_claimed = false;

// read_full reads exactly n bytes from stdin. It returns n, 0 on EOF before any byte, or -1.
long read_full(uint8_t* p, size_t n, std::string* err) {
  size_t got = 0;
  while (got < n) {
#ifdef _WIN32
    unsigned want = static_cast<unsigned>(n - got > (1u << 30) ? (1u << 30) : n - got);
    int r = _read(0, p + got, want);
#else
    ssize_t r = read(0, p + got, n - got);
    if (r < 0 && errno == EINTR) continue;
#endif
    if (r < 0) {
      *err = std::string("reading stdin: ") + std::strerror(errno);
      return -1;
    }
    if (r == 0) {
      if (got == 0) return 0;
      *err = "stdin ended inside a frame";
      return -1;
    }
    got += static_cast<size_t>(r);
  }
  return static_cast<long>(got);
}

void flush_all() {
  std::cout.flush();
  std::cerr.flush();
  std::fflush(stdout);
  std::fflush(stderr);
}

void sleep_ms(uint32_t ms) {
#ifdef _WIN32
  Sleep(ms);
#else
  struct timespec ts;
  ts.tv_sec = ms / 1000;
  ts.tv_nsec = static_cast<long>(ms % 1000) * 1000000L;
  while (nanosleep(&ts, &ts) != 0 && errno == EINTR) {
  }
#endif
}

void selftest_stdout(std::vector<std::string>* result) {
  std::printf("selftest-stdout: printf\n");
  std::cout << "selftest-stdout: cout" << std::endl;
  static const char kFwrite[] = "selftest-stdout: fwrite\n";
  std::fwrite(kFwrite, 1, sizeof kFwrite - 1, stdout);
  static const char kWrite[] = "selftest-stdout: write\n";
#ifdef _WIN32
  int ignored = _write(1, kWrite, sizeof kWrite - 1);
#else
  ssize_t ignored = write(1, kWrite, sizeof kWrite - 1);
#endif
  (void)ignored;
  std::string all(256, '\0');
  for (int i = 0; i < 256; i++) all[i] = static_cast<char>(i);
  result->push_back(all);
}

}  // namespace

bool decode_header(const uint8_t* p, Header* h, std::string* err) {
  uint32_t magic = get32(p);
  if (magic != kMagic) {
    *err = "bad magic";
    return false;
  }
  h->version = get16(p + 4);
  h->op = get16(p + 6);
  h->id = get32(p + 8);
  h->length = get32(p + 12);
  if (h->length > kMaxPayload) {
    *err = "payload length over the bound";
    return false;
  }
  return true;
}

bool decode_fields(const uint8_t* p, size_t n, std::vector<std::string>* fields, std::string* err) {
  fields->clear();
  while (n > 0) {
    if (fields->size() == kMaxFields) {
      *err = "too many fields";
      return false;
    }
    if (n < 4) {
      *err = "trailing bytes under a field's length prefix";
      return false;
    }
    uint32_t len = get32(p);
    p += 4;
    n -= 4;
    if (len > n) {
      *err = "field longer than the payload left";
      return false;
    }
    fields->emplace_back(reinterpret_cast<const char*>(p), len);
    p += len;
    n -= len;
  }
  return true;
}

bool decode(const uint8_t* p, size_t n, Frame* f, std::string* err) {
  if (n < kHeaderSize) {
    *err = "shorter than a header";
    return false;
  }
  Header h;
  if (!decode_header(p, &h, err)) return false;
  if (n - kHeaderSize != h.length) {
    *err = "payload length disagrees with the frame";
    return false;
  }
  f->version = h.version;
  f->op = h.op;
  f->id = h.id;
  return decode_fields(p + kHeaderSize, h.length, &f->fields, err);
}

bool encode(const Frame& f, std::string* out, std::string* err) {
  if (f.fields.size() > kMaxFields) {
    *err = "too many fields";
    return false;
  }
  uint64_t n = 0;
  for (const auto& fl : f.fields) {
    n += 4 + static_cast<uint64_t>(fl.size());
    if (n > kMaxPayload) {
      *err = "payload over the bound";
      return false;
    }
  }
  out->clear();
  out->reserve(kHeaderSize + n);
  put32(out, kMagic);
  put16(out, f.version);
  put16(out, f.op);
  put32(out, f.id);
  put32(out, static_cast<uint32_t>(n));
  for (const auto& fl : f.fields) {
    put32(out, static_cast<uint32_t>(fl.size()));
    out->append(fl);
  }
  return true;
}

bool claim(std::string* err) {
  if (g_claimed) return true;
  flush_all();
#ifdef _WIN32
  if (_setmode(0, _O_BINARY) == -1) {
    *err = "setting binary mode on stdin failed";
    return false;
  }
  int fd = _dup(1);
  if (fd < 0) {
    *err = "duplicating fd 1 failed";
    return false;
  }
  // _dup2 closes fd 1's original OS handle, so frames go through the duplicate's handle,
  // taken after the redirect, never a GetStdHandle value read before it.
  if (_dup2(2, 1) != 0) {
    *err = "pointing fd 1 at fd 2 failed";
    return false;
  }
  if (_setmode(fd, _O_BINARY) == -1) {
    *err = "setting binary mode on the frame descriptor failed";
    return false;
  }
  g_frame = reinterpret_cast<HANDLE>(_get_osfhandle(fd));
  if (g_frame == INVALID_HANDLE_VALUE) {
    *err = "the frame descriptor has no OS handle";
    return false;
  }
#else
  int fd = fcntl(1, F_DUPFD_CLOEXEC, 3);
  if (fd < 0) {
    *err = std::string("duplicating fd 1: ") + std::strerror(errno);
    return false;
  }
  if (dup2(2, 1) < 0) {
    *err = std::string("pointing fd 1 at fd 2: ") + std::strerror(errno);
    return false;
  }
  g_frame = fd;
#endif
  std::setvbuf(stdout, nullptr, _IONBF, 0);
  g_claimed = true;
  return true;
}

int read_frame(Frame* f, std::string* err) {
  uint8_t hb[kHeaderSize];
  long r = read_full(hb, kHeaderSize, err);
  if (r <= 0) return static_cast<int>(r);
  Header h;
  if (!decode_header(hb, &h, err)) return -1;
  std::string payload(h.length, '\0');
  if (h.length > 0) {
    r = read_full(reinterpret_cast<uint8_t*>(&payload[0]), h.length, err);
    if (r == 0) {
      *err = "stdin ended inside a frame";
      return -1;
    }
    if (r < 0) return -1;
  }
  f->version = h.version;
  f->op = h.op;
  f->id = h.id;
  if (!decode_fields(reinterpret_cast<const uint8_t*>(payload.data()), payload.size(), &f->fields, err)) return -1;
  return 1;
}

bool write_frame(const Frame& f, std::string* err) {
  if (!g_claimed) {
    *err = "the frame descriptor is not claimed";
    return false;
  }
  std::string b;
  if (!encode(f, &b, err)) return false;
  size_t done = 0;
  while (done < b.size()) {
#ifdef _WIN32
    DWORD wrote = 0;
    DWORD want = static_cast<DWORD>(b.size() - done > (1u << 30) ? (1u << 30) : b.size() - done);
    if (!WriteFile(g_frame, b.data() + done, want, &wrote, nullptr)) {
      *err = "WriteFile on the frame descriptor failed";
      return false;
    }
#else
    ssize_t wrote = write(g_frame, b.data() + done, b.size() - done);
    if (wrote < 0 && errno == EINTR) continue;
    if (wrote < 0) {
      *err = std::string("writing a frame: ") + std::strerror(errno);
      return false;
    }
#endif
    done += static_cast<size_t>(wrote);
  }
  return true;
}

bool Capture::begin(std::string* err) {
  flush_all();
  if (file_ < 0) {
    const char* path = std::getenv("NATIVE_HELPER_CAPTURE");
#ifdef _WIN32
    std::string p;
    if (path != nullptr && *path != '\0') {
      p = path;
    } else {
      char dir[MAX_PATH + 1];
      DWORD n = GetTempPathA(sizeof dir, dir);
      if (n == 0 || n > MAX_PATH) {
        *err = "no temporary directory for the capture file";
        return false;
      }
      p = std::string(dir) + "feov-helper-capture-" + std::to_string(GetCurrentProcessId());
    }
    // The client owns a named capture file and reads it after a crash; the default one is the
    // helper's own and goes when the helper exits.
    int temporary = (path != nullptr && *path != '\0') ? 0 : _O_TEMPORARY;
    file_ = _open(p.c_str(), _O_RDWR | _O_CREAT | _O_TRUNC | _O_BINARY | _O_NOINHERIT | temporary, 0600);
#else
    if (path != nullptr && *path != '\0') {
      file_ = open(path, O_RDWR | O_CREAT | O_TRUNC | O_CLOEXEC, 0600);
    } else {
      std::FILE* t = std::tmpfile();
      if (t != nullptr) {
        file_ = fcntl(fileno(t), F_DUPFD_CLOEXEC, 3);
        std::fclose(t);
      }
    }
#endif
    if (file_ < 0) {
      *err = std::string("opening the capture file: ") + std::strerror(errno);
      return false;
    }
  }
#ifdef _WIN32
  if (_chsize(file_, 0) != 0 || _lseek(file_, 0, SEEK_SET) < 0) {
    *err = "truncating the capture file failed";
    return false;
  }
  saved_err_ = _dup(2);
  if (saved_err_ < 0 || _dup2(file_, 1) != 0 || _dup2(file_, 2) != 0) {
    *err = "pointing fd 1 and fd 2 at the capture file failed";
    return false;
  }
#else
  if (ftruncate(file_, 0) != 0 || lseek(file_, 0, SEEK_SET) < 0) {
    *err = std::string("truncating the capture file: ") + std::strerror(errno);
    return false;
  }
  saved_err_ = fcntl(2, F_DUPFD_CLOEXEC, 3);
  if (saved_err_ < 0 || dup2(file_, 1) < 0 || dup2(file_, 2) < 0) {
    *err = std::string("pointing fd 1 and fd 2 at the capture file: ") + std::strerror(errno);
    return false;
  }
#endif
  return true;
}

std::string Capture::end() {
  flush_all();
  std::string out;
  if (saved_err_ < 0) return out;
  // Only the last kStderrKeep bytes are read: a request that floods stderr must not push its
  // response past kMaxPayload.
#ifdef _WIN32
  _dup2(saved_err_, 2);
  _dup2(2, 1);
  _close(saved_err_);
  long long size = _lseeki64(file_, 0, SEEK_END);
  long long skip = size > static_cast<long long>(kStderrKeep) ? size - static_cast<long long>(kStderrKeep) : 0;
  _lseeki64(file_, skip, SEEK_SET);
  char buf[65536];
  int r;
  while ((r = _read(file_, buf, sizeof buf)) > 0) out.append(buf, static_cast<size_t>(r));
#else
  dup2(saved_err_, 2);
  dup2(2, 1);
  close(saved_err_);
  off_t size = lseek(file_, 0, SEEK_END);
  off_t skip = size > static_cast<off_t>(kStderrKeep) ? size - static_cast<off_t>(kStderrKeep) : 0;
  lseek(file_, skip, SEEK_SET);
  char buf[65536];
  ssize_t r;
  while ((r = read(file_, buf, sizeof buf)) > 0 || (r < 0 && errno == EINTR)) {
    if (r > 0) out.append(buf, static_cast<size_t>(r));
  }
#endif
  saved_err_ = -1;
  if (skip > 0) {
    out.insert(0, "frame: " + std::to_string(static_cast<long long>(skip)) + " earlier bytes of stderr dropped\n");
  }
  return out;
}

int serve(const std::string& digest, const std::vector<std::string>& assets, const Handler& handler) {
  std::string err;
  if (!claim(&err)) {
    std::fprintf(stderr, "frame: %s\n", err.c_str());
    return 2;
  }
  Capture cap;
  bool greeted = false;
  for (;;) {
    Frame req;
    int r = read_frame(&req, &err);
    if (r == 0) return 0;
    if (r < 0) {
      std::fprintf(stderr, "frame: %s\n", err.c_str());
      return 3;
    }
    // A request in another protocol version cannot be answered in a frame its sender reads, and a
    // first request that is not the handshake comes from a client that never checked this
    // helper's digest. Both end the helper with the reason on stderr.
    if (req.version != kVersion) {
      std::fprintf(stderr, "frame: request in protocol version %u, this helper speaks %u\n",
                   static_cast<unsigned>(req.version), static_cast<unsigned>(kVersion));
      return 6;
    }
    if (!greeted && req.op != kHandshake) {
      std::fprintf(stderr, "frame: the first request is op %u, not the handshake\n", static_cast<unsigned>(req.op));
      return 7;
    }
    greeted = true;
    if (!cap.begin(&err)) {
      std::fprintf(stderr, "frame: %s\n", err.c_str());
      return 4;
    }
    std::vector<std::string> result;
    Status st = kOK;
    switch (req.op) {
      case kHandshake:
        result.push_back(digest);
        result.insert(result.end(), assets.begin(), assets.end());
        break;
      case kSelftestStdout:
        selftest_stdout(&result);
        break;
      case kSelftestSleep:
        if (req.fields.size() != 1 || req.fields[0].size() != 4) {
          st = kOpError;
          result.push_back("selftest-sleep takes one 4-byte field");
          break;
        }
        sleep_ms(get32(reinterpret_cast<const uint8_t*>(req.fields[0].data())));
        break;
      default:
        if (req.op >= kFirstHelperOp && handler) {
          st = handler(req.op, req.fields, &result);
        } else {
          st = kUnknownOp;
        }
        if (st == kUnknownOp) result.assign(1, "unknown op " + std::to_string(req.op));
    }
    Frame resp;
    resp.op = req.op;
    resp.id = req.id;
    resp.fields.reserve(2 + result.size());
    resp.fields.emplace_back(1, static_cast<char>(st));
    resp.fields.push_back(cap.end());
    for (auto& x : result) resp.fields.push_back(std::move(x));
    if (!write_frame(resp, &err)) {
      std::fprintf(stderr, "frame: %s\n", err.c_str());
      return 5;
    }
  }
}

}  // namespace frame
}  // namespace feov

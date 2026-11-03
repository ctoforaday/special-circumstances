GO       ?= go
GOFLAGS  ?= -trimpath
PGO_PROFILE ?= cpu.pprof

# ---------- CGo / OpenBLAS detection ----------
# Default: pure Go (CGO_ENABLED=0).  Portable, no C toolchain needed, ~14ms/embed.
# Set CGO_ENABLED=1 to use OpenBLAS via CGo for ~7ms/embed (requires gcc + libopenblas-dev).
#
# Why pure Go is the default:
#   The pure-Go path uses gonum's cache-blocked BLAS, which is already 5× faster
#   than the original baseline.  It produces a single static binary that runs
#   anywhere — no libopenblas.so, no gcc, no pkg-config.  The OpenBLAS CGo path
#   is ~1.9× faster still, but requires a C toolchain and a system library.
#   We prefer portability by default; opt in to CGo when you need maximum speed.
export CGO_ENABLED ?= 0

ifeq (,$(wildcard $(PGO_PROFILE)))
PGO_FLAG :=
else
PGO_FLAG := -pgo=$(PGO_PROFILE)
endif

all: go-build go-test

go-build:
	$(GO) build $(GOFLAGS) $(PGO_FLAG) ./cmd/gte ./cmd/test_gte ./cmd/bench

go-test: gte-small.gtemodel
	GTE_MODEL_PATH=gte-small.gtemodel $(GO) test $(GOFLAGS) $(PGO_FLAG) ./...

run-go: gte-small.gtemodel
	$(GO) run $(GOFLAGS) $(PGO_FLAG) ./cmd/gte --model-path gte-small.gtemodel "I love cats" "I love dogs" "The stock market crashed"

run-bench: gte-small.gtemodel
	$(GO) run $(GOFLAGS) $(PGO_FLAG) ./cmd/bench --model-path gte-small.gtemodel --iterations 100 --warmup 5

run-bench-compare: gte-small.gtemodel
	$(GO) run $(GOFLAGS) $(PGO_FLAG) ./cmd/bench_compare

bench-profile: gte-small.gtemodel
	$(GO) run $(GOFLAGS) ./cmd/bench --model-path gte-small.gtemodel --iterations 200 --warmup 10 --cpuprofile $(PGO_PROFILE)

go-build-pgo: $(PGO_PROFILE)
	$(GO) build $(GOFLAGS) -pgo=$(PGO_PROFILE) ./cmd/gte ./cmd/test_gte ./cmd/bench

# Build with OpenBLAS acceleration (requires gcc + libopenblas-dev)
go-build-cgo:
	CGO_ENABLED=1 $(GO) build $(GOFLAGS) $(PGO_FLAG) ./cmd/gte ./cmd/test_gte ./cmd/bench

gte-small.gtemodel: models/gte-small/model.safetensors
	python3 convert_model.py models/gte-small $@

clean:
	rm -f gte-small.gtemodel $(PGO_PROFILE)

.PHONY: all clean go-build go-test run-go run-bench run-bench-compare bench-profile go-build-pgo go-build-cgo go-bench

go-bench: gte-small.gtemodel
	GTE_MODEL_PATH=gte-small.gtemodel $(GO) test $(GOFLAGS) $(PGO_FLAG) -bench=BenchmarkEmbed -benchmem ./gte

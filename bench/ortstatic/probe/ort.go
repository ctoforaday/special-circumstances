//go:build !noort

package main

/*
#cgo CFLAGS: -DORT_API_MANUAL_INIT
#include <stdlib.h>
#include <string.h>
#include "onnxruntime_c_api.h"

static const OrtApi* g;
static OrtEnv* env;
static OrtSession* sess;
static OrtMemoryInfo* mem;

static char* errstr(OrtStatus* st) {
	if (!st) return NULL;
	const char* m = g->GetErrorMessage(st);
	char* r = strdup(m);
	g->ReleaseStatus(st);
	return r;
}

#define CK(x) do { char* e_ = errstr(x); if (e_) return e_; } while (0)

static char* ortp_init(const ORTCHAR_T* model, int threads) {
	g = OrtGetApiBase()->GetApi(ORT_API_VERSION);
	if (!g) return strdup("OrtGetApiBase()->GetApi returned NULL");
	CK(g->CreateEnv(ORT_LOGGING_LEVEL_WARNING, "ortprobe", &env));
	OrtSessionOptions* so;
	CK(g->CreateSessionOptions(&so));
	CK(g->SetIntraOpNumThreads(so, threads));
	CK(g->SetInterOpNumThreads(so, 1));
	CK(g->SetSessionExecutionMode(so, ORT_SEQUENTIAL));
	CK(g->SetSessionGraphOptimizationLevel(so, ORT_ENABLE_ALL));
	CK(g->CreateSession(env, model, so, &sess));
	g->ReleaseSessionOptions(so);
	CK(g->CreateCpuMemoryInfo(OrtArenaAllocator, OrtMemTypeDefault, &mem));
	return NULL;
}

static char* ortp_run(int64_t* ids, int64_t* mask, int64_t n, float* out) {
	int64_t shape[2] = {1, n};
	OrtValue* in[2] = {NULL, NULL};
	OrtValue* o = NULL;
	CK(g->CreateTensorWithDataAsOrtValue(mem, ids, n * sizeof(int64_t), shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &in[0]));
	CK(g->CreateTensorWithDataAsOrtValue(mem, mask, n * sizeof(int64_t), shape, 2, ONNX_TENSOR_ELEMENT_DATA_TYPE_INT64, &in[1]));
	const char* inames[2] = {"input_ids", "attention_mask"};
	const char* onames[1] = {"last_hidden_state"};
	CK(g->Run(sess, NULL, inames, (const OrtValue* const*)in, 2, onames, 1, &o));
	float* p;
	CK(g->GetTensorMutableData(o, (void**)&p));
	memcpy(out, p, (size_t)n * 384 * sizeof(float));
	g->ReleaseValue(o);
	g->ReleaseValue(in[0]);
	g->ReleaseValue(in[1]);
	return NULL;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

const engineName = "onnxruntime"

type engine struct{}

func newEngine(model string, threads int) (*engine, error) {
	cm := modelPath(model)
	defer C.free(unsafe.Pointer(cm))
	if e := C.ortp_init(cm, C.int(threads)); e != nil {
		defer C.free(unsafe.Pointer(e))
		return nil, errors.New(C.GoString(e))
	}
	return &engine{}, nil
}

func (*engine) run(ids, mask []int64) ([]float32, error) {
	out := make([]float32, len(ids)*dim)
	// The C side copies into out before returning; ids/mask are pinned by cgo for the call.
	if e := C.ortp_run((*C.int64_t)(unsafe.Pointer(&ids[0])), (*C.int64_t)(unsafe.Pointer(&mask[0])), C.int64_t(len(ids)), (*C.float)(unsafe.Pointer(&out[0]))); e != nil {
		defer C.free(unsafe.Pointer(e))
		return nil, errors.New(C.GoString(e))
	}
	return out, nil
}

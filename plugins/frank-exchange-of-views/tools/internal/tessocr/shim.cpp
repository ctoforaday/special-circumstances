//go:build tessocr && cgo

// The package's own shim over tesseract + leptonica — no third-party binding.
//
// It is C++ for exactly one reason: tesseract 5.5.3's C API has no init-from-memory entry
// point, and embedding eng.traineddata in the binary requires the C++ overload
// TessBaseAPI::Init(data, size, ...). Once one C++ translation unit exists, the rest of
// the shim lives here too rather than splitting operations between a cgo C preamble and
// this file — one place to read, one allocation discipline.
//
// Everything is synchronous and the engine is NOT safe for concurrent use: a TessBaseAPI
// carries per-recognition state (image, PSM, result). The Go side documents the same.

#include "shim.h"

#include <tesseract/baseapi.h>
#include <leptonica/allheaders.h>

namespace {

// Leptonica prints harmless per-run stderr noise because TIFF is compiled out
// ("pixReadMemTiff: function not present", bmfCreate font errors). Route it to nothing:
// the messages describe a build decision, not a runtime fault, and they would pollute
// every captured run log.
void discard_stderr(const char *) {}

struct stderr_silencer {
	stderr_silencer() { leptSetStderrHandler(discard_stderr); }
} silencer;

PIX *read_png(const unsigned char *buf, size_t len) {
	return pixReadMem(buf, static_cast<l_uint32>(len));
}

char *ocr_pix(tesseract::TessBaseAPI *api, PIX *pix, int psm, bool tsv) {
	api->SetPageSegMode(static_cast<tesseract::PageSegMode>(psm));
	api->SetImage(pix);
	char *text = tsv ? api->GetTSVText(0) : api->GetUTF8Text();
	pixDestroy(&pix);
	return text;
}

} // namespace

struct tessocr_engine {
	tesseract::TessBaseAPI api;
};

extern "C" {

tessocr_engine *tessocr_new(const unsigned char *traineddata, int len) {
	tessocr_engine *e = new tessocr_engine();
	// OEM_LSTM_ONLY, stated: the C stack is built with DISABLED_LEGACY_ENGINE and
	// tessdata_fast is LSTM-only, so any other mode would be asking for code that is
	// not in the binary.
	if (e->api.Init(reinterpret_cast<const char *>(traineddata), len, "eng",
	                tesseract::OEM_LSTM_ONLY, nullptr, 0, nullptr, nullptr,
	                false, nullptr) != 0) {
		delete e;
		return nullptr;
	}
	return e;
}

void tessocr_free(tessocr_engine *e) {
	if (e == nullptr) return;
	e->api.End();
	delete e;
}

char *tessocr_text(tessocr_engine *e, const unsigned char *png, size_t len, int psm) {
	PIX *pix = read_png(png, len);
	if (pix == nullptr) return nullptr;
	return ocr_pix(&e->api, pix, psm, false);
}

char *tessocr_tsv(tessocr_engine *e, const unsigned char *png, size_t len, int psm) {
	PIX *pix = read_png(png, len);
	if (pix == nullptr) return nullptr;
	return ocr_pix(&e->api, pix, psm, true);
}

char *tessocr_rot_band(tessocr_engine *e, const unsigned char *png, size_t len,
                       int x, int y, int w, int h) {
	PIX *pix = read_png(png, len);
	if (pix == nullptr) return nullptr;
	BOX *box = boxCreate(x, y, w, h);
	PIX *clip = pixClipRectangle(pix, box, nullptr);
	boxDestroy(&box);
	pixDestroy(&pix);
	if (clip == nullptr) return nullptr;
	PIX *rot = pixRotate90(clip, 1); // 1 = clockwise, see shim.h
	pixDestroy(&clip);
	if (rot == nullptr) return nullptr;
	return ocr_pix(&e->api, rot, 3 /* PSM_AUTO */, false);
}

int tessocr_grid_stats(const unsigned char *png, size_t len, int sel,
                       int *hpix, int *vpix, int *inter) {
	PIX *pix = read_png(png, len);
	if (pix == nullptr) return 1;
	// Binarize at 180: the corpus is grayscale scans where rules print well under that
	// threshold; the Wave 0 tune measured everything downstream of this exact value.
	PIX *bin = pixConvertTo1(pix, 180);
	pixDestroy(&pix);
	if (bin == nullptr) return 2;
	PIX *hl = pixOpenBrick(nullptr, bin, sel, 1);
	PIX *vl = pixOpenBrick(nullptr, bin, 1, sel);
	pixDestroy(&bin);
	if (hl == nullptr || vl == nullptr) {
		pixDestroy(&hl);
		pixDestroy(&vl);
		return 3;
	}
	PIX *ix = pixAnd(nullptr, hl, vl);
	l_int32 ch = 0, cv = 0, ci = 0;
	pixCountPixels(hl, &ch, nullptr);
	pixCountPixels(vl, &cv, nullptr);
	pixCountPixels(ix, &ci, nullptr);
	pixDestroy(&hl);
	pixDestroy(&vl);
	pixDestroy(&ix);
	*hpix = ch;
	*vpix = cv;
	*inter = ci;
	return 0;
}

void tessocr_free_text(char *t) {
	delete[] t;
}

} // extern "C"

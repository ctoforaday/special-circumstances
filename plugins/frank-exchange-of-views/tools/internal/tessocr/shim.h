/* The C surface of the tessocr shim — see shim.cpp for why a C++ file exists at all.
 * Every string returned here was allocated by tesseract (new[]) and MUST be released via
 * tessocr_free_text, never free(). */
#ifndef TESSOCR_SHIM_H
#define TESSOCR_SHIM_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct tessocr_engine tessocr_engine;

/* tessocr_new initializes an engine from an in-memory traineddata blob (the go:embed'ed
 * eng.traineddata). NULL means init failed — tesseract reports no more detail through
 * this path, so there is nothing further to surface. */
tessocr_engine *tessocr_new(const unsigned char *traineddata, int len);

void tessocr_free(tessocr_engine *e);

/* OCR a PNG to UTF-8 text under the given page-segmentation mode. NULL on failure. */
char *tessocr_text(tessocr_engine *e, const unsigned char *png, size_t len, int psm);

/* OCR a PNG to level-5 word TSV (geometry + confidence per word). NULL on failure. */
char *tessocr_tsv(tessocr_engine *e, const unsigned char *png, size_t len, int psm);

/* Crop (x,y,w,h), rotate the crop 90 degrees CLOCKWISE, OCR it — the rotated-header
 * band recovery. Clockwise is knowable a priori: bottom-up column headers read top-down
 * after a clockwise turn; counter-clockwise yields upside-down junk. NULL on failure. */
char *tessocr_rot_band(tessocr_engine *e, const unsigned char *png, size_t len,
                       int x, int y, int w, int h);

/* Morphological grid detection: pixels surviving a long-horizontal-run opening, a
 * long-vertical-run opening, and their AND. sel is the minimum run length in px.
 * Returns 0 on success; nonzero names the failing stage (1 decode, 2 binarize, 3 open)
 * so a broken image is an ERROR, never a page measuring zero. */
int tessocr_grid_stats(const unsigned char *png, size_t len, int sel,
                       int *hpix, int *vpix, int *inter);

void tessocr_free_text(char *t);

#ifdef __cplusplus
}
#endif

#endif

# Sources not committed

Pages we would want and may not redistribute, and leads that could not be reached.
Kept as a record so the next round does not re-find them and assume.
GENERATED from references.json.

## chronicling-america

- https://chroniclingamerica.loc.gov/
- Wanted for: faint multi-column newspaper pages with halftones — public domain, and the strongest lead for that class
- Rights: public domain (Library of Congress)
- Blocked because: unreachable from this container: chroniclingamerica.loc.gov redirects to www.loc.gov, which answers 403, as does tile.loc.gov, with and without a browser user agent. Fetch it from a browser and add it.

## commons-typescript-handwriting

- https://upload.wikimedia.org/wikipedia/commons/3/32/MS220_Log_296%2C_Charlotte_Coffin_Gardner_trading_journal_%28IA_ms220log296%29.pdf
- Wanted for: a typescript carrying handwritten interlineations, carets and struck-through words
- Rights: public-domain-expired
- Blocked because: its content IS the PDF's text layer, not pixels: removing the layer leaves 0.53% ink, speckle and a few stamped fragments. It looked like a scan in a thumbnail because the thumbnail rendered the text. corpusadd now refuses a page like this, and the 1% threshold was measured against it.

## funsd

- https://guillaumejaume.github.io/FUNSD/
- Wanted for: the best available source of handwriting in form fields
- Rights: UNKNOWN for redistribution
- Blocked because: 'for non-commercial, research purposes only', and it disclaims having cleared the images at all: 'Licensees are solely responsible for determining what additional licenses, clearances, consents and releases, if any, must be obtained'.

## ieee-1012-table2

- https://people.eecs.ku.edu/~hossein/Teaching/Stds/1012.pdf
- Wanted for: the mark grid every mark and level constant was fitted to (four integrity-level subcolumns per activity), and the page whose marks the engine still drops
- Rights: copyrighted — IEEE Std 1012-1998
- Blocked because: a copyrighted standard. It stays in the measurement runs (~/ocr-runs/932/e2e, ~/ocr-runs/644-real) and out of the repository; usfs-birds-markgrid is the committable mark grid.

## landscape-table

- https://www.govinfo.gov/
- Wanted for: a table printed landscape on a portrait page — the case that defeats the orientation probe, which keys on a confident-word count and does not fire when rotated headers read horizontally
- Rights: us-government-work, once a specific document is chosen
- Blocked because: no verified specimen yet. Leads tried and failed: USGS WSP 365 (its one oversize plate is portrait), a GPO hearing PDF (no inserted exhibit pages). A page whose width exceeds its height is the cheap detector for finding one.

## nara-1950-census-handwriting

- https://nara-1950-census.s3.us-east-2.amazonaws.com/1950census/43290879-Alabama/43290879-Alabama-005563/43290879-Alabama-005563-0005.jpg
- Wanted for: handwriting inside ruled form fields, on microfilm — a class no committed page covers, since commons-typescript-handwriting carries interlineations rather than filled fields
- Rights: us-government-work — AWS Registry of Open Data states 'License: US Government work'
- Blocked because: one frame is 3960x4830 and 5.7 MB, over fetch's own 5242880-byte cap, so the harness could not read it even if it were committed. Cropping a region would re-encode the publisher's pixels; that trade has not been taken.

## rvl-cdip

- https://adamharley.com/rvl-cdip/
- Wanted for: faxes, dot-matrix and handwriting at scale — the widest defect coverage in existence
- Rights: UNKNOWN
- Blocked because: no licence of its own; it defers to IIT-CDIP and the UCSF Legacy Tobacco Document Library, which grant no redistribution.

## skewed-scan

- https://github.com/tesseract-ocr/test
- Wanted for: a genuinely skewed or warped scan, as opposed to a synthetic rotation
- Rights: apache-2.0 for the synthetic rotations in that repository
- Blocked because: the committable files there (phototestrot.tif, deslant.tif) are synthetic controls, not real skew. A real specimen has not been found.

## unlv-isri-full-sets

- https://sourceforge.net/projects/isri-ocr-evaluation-tools-alt/files/
- Wanted for: the 1995 UNLV/ISRI OCR test pages: business letters, magazines, newspapers, DOE reports
- Rights: UNKNOWN
- Blocked because: neither the SourceForge mirror nor tesseract's unlvtests/README.md states a licence. Only the handful vendored into tesseract-ocr/test carries one, and that grant is the tesseract project's Apache-2.0, not UNLV's.


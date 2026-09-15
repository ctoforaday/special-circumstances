## menu

draw one page of a cached PDF, to check a citation of OCR text against the page's pixels

## detail

A citation whose evidence entry lists pages quotes OCR text: a machine's reading of page images, which can misread. Check it against the page image, not against the reading.

Draws ONE page at the engine's DPI and prints the image's path and sha256, with the path of that page's reading beside it. Read-only: it records nothing. The verification you record next names the page image you checked.

`matches_reading: false` means the image was drawn by a different renderer than the reading's; it still shows the page. `reading: none` means the tool holds no OCR reading of the document.

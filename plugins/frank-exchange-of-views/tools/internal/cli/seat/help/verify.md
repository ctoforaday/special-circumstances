## menu

judge a citation blue authored, when the anchor exists and you have read what it points at

## detail

THIS VERB JUDGES A CITATION THAT EXISTS: a claim with NO citation is never verified as `absent`.

ONE LENS OWNS EACH CITATION: the evidence lens. Its work list alone offers a citation nobody has verified, and it is readied for a verified one an edit reopens.

YOU NAME THE CITATION, AND THE TOOL READS ITS CLAIM. The anchor is the id inside the citation's token in the report. The claim recorded is the whole sentence of the report holding that token, whichever part of it the citation backs, as the report stands when you verify, under every outcome — the sentence the citation's source backs, contradicts or does not contain — and the tool prints it back with your verdict. What the SOURCE says goes in your reason.

A `refutes` OR `absent` VERDICT OPENS A CONTRADICTION, which holds a PASS until it is answered. `show evidence` lists it under `unanswered_contradictions` with its `id`, and `raised_by` once a finding has raised it. Only a finding about that contradiction, naming it by its id, raises it, whether or not the report still holds the sentence; a finding whose quote contains the claim raises nothing. One finding raises it, whichever lens files it, and a verdict against the same claim after that finding opens a new contradiction. The gap minted from that finding, crediting it as the finding that surfaced it, answers it, and so does an open gap that already holds its defect, which the finding's lens credits with it (`credit`); the finding alone does not, because a source the report misstates is a citation defect, material by its class whatever reader decision you judge it to change. The citation answers it as well: blue's retire of the citation, or a verdict on the citation's sentence as it now stands, with every lens's latest one backing it. Your verdict stops counting when you read the citation's source again or an edit changes the sentence after your sitting opened; a verdict counts once it comes from a sitting opened after the sentence's latest edit. Another lens's backing verdict while yours counts is a disagreement, and answers nothing.

WHAT THIS VERB PRINTS NAMES THE CONTRADICTION YOUR READING LANDED IN. A reading that opens one ends `This reading opened contradiction <id>`; one against a claim whose contradiction is already open joins it and ends `This reading joined contradiction <id> (opened by <seat>)`; the JSON result carries `contradiction` — `id`, `opened`, `opened_by`. A reading that opens and joins none prints neither. That id is the one the finding that raises it names.

A CITATION THE REPORT HOLDS NO SENTENCE FOR IS REFUSED, and nothing is recorded, because there is no claim to read its source against: one whose anchor is not in the report (blue retired its claim, or the anchor never placed), one whose anchor stands bare (the sentence it backed was cut, and the anchor leaves with blue's retire), and every citation while the report does not render (the board's anomalies say why; record that you could not verify with `log`). Verify the citations whose anchors the report carries.

A CITATION WITH PAGES QUOTES OCR TEXT, and the reading may be what is wrong. Draw one of its pages with `render-page`, read the image, and name that page: the tool refuses a verification of such a citation that names no checked page, and records the image's hash. An unreachable outcome needs no page when no image could be drawn.

Nor is an unevidenced claim automatically a finding: if a source exists and you can reach it, `corroborate` goes and gets it and answers whether the claim is true in the WORLD; `finding` is for when there is nothing to fetch, and answers whether the TEXT stands up.

THE TOOL RECORDS WHICH COPY YOU READ. The citation's url is read through the run's fetch cache, and your verdict is recorded against the copy the cache holds for it: the url's own bytes, or the full text its page pointed to, a copy recovered after a refusal, a fuller copy, or a named backend's answer, with where that copy was read. What this verb prints names a copy that is not the url's own. A source the cache holds no copy of is recorded as `not_cached`, which says nothing about what you read: read the source through `fetch` and the record names the copy.

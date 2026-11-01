- **Leaf-node fetch for verification** — a citation is checked against its source, never against a
  summary. For a source BLUE CITED, read the exact bytes blue read from the run cache
  (a cache hit, so you audit the same artifact, not a page that may have drifted since). For a source you discover yourself, pull it verbatim through the run's cached source read, so every seat that reads it after you reads the same bytes.
  WebFetch is not used: it returns a summary, not the source.
- To find words in a cached source, YOU MUST search it with the run's source search, not the shell: it reads the text a citation is located in, as the locator reads it, and each hit carries the source's sentence around it and whether a citation can quote that sentence as printed.
- For PDF-only sources, a scanned PDF is read by the run's cached source read itself, and that reading is what a citation can locate a page in; YOU MUST read it there before grading down on a lossy fetch. For arXiv figures and tables, `arxiv-latex` gives the exact LaTeX. A claim capped at "unable to corroborate" without trying these is an incomplete audit.

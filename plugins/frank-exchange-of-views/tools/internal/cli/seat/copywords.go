package seat

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

// CopyWords is what a cite or a reading prints about whose bytes it was made against, from the
// record's copy pair: nothing on the url's own bytes, the route and where the copy was read when a
// copy stood in for the url, and not_cached where the run's cache held no copy.
func CopyWords(kind, copyURL string) string {
	switch {
	case kind == "" || kind == recordpb.Word(recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE):
		return ""
	case kind == recordpb.Word(recordpb.CopySubstitution_COPY_SUBSTITUTION_NOT_CACHED):
		return ". The run's fetch cache holds no copy of the source, so the copy read is recorded as not_cached"
	case copyURL == "":
		return ". The bytes read are not the url's: copy_substitution " + kind + ", and no document was read — the cache holds the fetch tool's statement of what it found"
	}
	return ". The bytes read are not the url's: copy_substitution " + kind + ", read at " + copyURL
}

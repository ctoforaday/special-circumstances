package fetchcache

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

// The routes by which the cache answers a url with bytes that are not the url's own. Each is set
// by the code that takes the route — Resolve's pointer walk, its recovery after a refusal, its
// search for a fuller copy, and a fetch that names a backend — and none is ever inferred from the
// entry afterwards.
const (
	// SubstitutionFollowedLink: the url's page named its own full text, and the walk followed it.
	SubstitutionFollowedLink = "followed_link"
	// SubstitutionRecovered: the url refused, and the auto recovery chain answered.
	SubstitutionRecovered = "recovered"
	// SubstitutionFullerCopy: the url's page was not known to carry the body, and a copy carrying
	// more of the work replaced it.
	SubstitutionFullerCopy = "fuller_copy"
	// SubstitutionNamedBackend: a fetch named a backend, and its answer is cached under the url.
	SubstitutionNamedBackend = "named_backend"
)

// CopySubstitution is the record's word for this entry's substitution — what a citation of it
// carries. A word this switch does not know maps to UNSPECIFIED, which the record's write guard
// refuses: a route added without a record word fails at the write rather than reaching a reader
// as the url's own bytes.
func (e Entry) CopySubstitution() recordpb.CopySubstitution {
	switch e.Substitution {
	case "":
		return recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE
	case SubstitutionFollowedLink:
		return recordpb.CopySubstitution_COPY_SUBSTITUTION_FOLLOWED_LINK
	case SubstitutionRecovered:
		return recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED
	case SubstitutionFullerCopy:
		return recordpb.CopySubstitution_COPY_SUBSTITUTION_FULLER_COPY
	case SubstitutionNamedBackend:
		return recordpb.CopySubstitution_COPY_SUBSTITUTION_NAMED_BACKEND
	}
	return recordpb.CopySubstitution_COPY_SUBSTITUTION_UNSPECIFIED
}

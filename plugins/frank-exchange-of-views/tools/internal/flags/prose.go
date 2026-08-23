package flags

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

// Prose is the prose payload channel AS ONE VALUE.
//
// # Why a type and not two flags plus a resolver
//
// `--reason` and `--reason-file` are one argument arriving three ways: inline, from a file, or
// from stdin via `--reason-file -`. Held as two loose string flags and a package-level resolver,
// that fact was true only for verbs that remembered both halves — and three did not:
//
//	line-of-inquiry propose  registered the pair and then filled its `line` key from the raw
//	                         --reason flag, so --reason worked and --reason-file was refused for
//	                         a missing field the seat had supplied. A blue seat wrote a paragraph
//	                         into a heredoc, was refused three times, and filed friction.
//	spot-check               registered --reason by hand. No file form, no stdin form.
//	outcome                  the same — on the field its own help calls "the only evidence the
//	                         determination ever had" on a judged deadlock.
//
// Every one of those is a verb forgetting a convention. A convention that has to be remembered
// at fifty call sites is not a mechanism, and the helper's own doc comment claimed it was one:
// "so a verb cannot register one form and forget the other". It could, and it did.
//
// So the pair becomes a value that owns its own registration and its own reading, the way CSV,
// GradeValue and DateValue already do in this package. Registering one form is not a mistake you
// can make, because there is nothing to register but the whole thing.
//
// # And an unregistered read is LOUD
//
// ReadPayload read both flags with GetString and discarded the errors, so a verb that never
// registered them got "" and no complaint — the resolver returning a plausible zero. That is the
// same defect the comment on Value() twenty lines away already documents for enum flags, still
// live in the function beside it.
type Prose struct {
	inline string
	file   string
	// resolved is the string representation once read: what the seat actually supplied, by
	// whichever spelling. Held here so a caller that has the value does not re-read stdin —
	// `--reason-file -` is a stream, and the second read of it returns nothing.
	resolved string
	read     bool
}

// registry maps a command to the prose channel it registered.
//
// Cobra's Annotations are string-valued, and the alternative — threading the struct through every
// verb constructor — is the convention this type exists to replace. Guarded because the role trees
// are built in parallel by the gates, which is how an unguarded map in the help cache became a
// `concurrent map read and map write` in this package's sibling.
var (
	registryMu sync.RWMutex
	registry   = map[*cobra.Command]*Prose{}
)

// Register attaches the channel to a command: both spellings, the canonical wording, and the
// binding that makes this struct the thing they write to.
func (p *Prose) Register(c *cobra.Command) {
	c.Flags().StringVar(&p.inline, Reason, "", DescReason)
	c.Flags().StringVar(&p.file, ReasonFile, "", DescReasonFile)
	registryMu.Lock()
	registry[c] = p
	registryMu.Unlock()
}

// RegisterRequired attaches the channel and makes it mandatory THROUGH EITHER SPELLING.
//
// MarkFlagRequired names one flag, so a verb that used it refused the file form — which is what
// `spot-check` did with a hand-registered --reason. One argument, one requirement.
func (p *Prose) RegisterRequired(c *cobra.Command) {
	p.Register(c)
	c.MarkFlagsOneRequired(Reason, ReasonFile)
}

// ProseOf returns the channel a command registered, or nil if it registered none.
func ProseOf(c *cobra.Command) *Prose {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[c]
}

// Read resolves the channel to one string.
//
// Both spellings given is refused rather than ranked: a seat that passes both should be told
// which one this verb would have dropped, not discover it in a projection three rounds later.
//
// The trailing newline a shell heredoc leaves behind is trimmed, because `<< 'EOF'` always adds
// one and a seat should not have to think about it.
func (p *Prose) Read(stdin io.Reader) (string, error) {
	if p.read {
		return p.resolved, nil
	}
	s, err := p.resolve(stdin)
	if err != nil {
		return "", err
	}
	p.resolved, p.read = s, true
	return s, nil
}

func (p *Prose) resolve(stdin io.Reader) (string, error) {
	if p.inline != "" && p.file != "" {
		return "", fmt.Errorf("--%s and --%s are two spellings of one payload: pass exactly one", Reason, ReasonFile)
	}
	switch {
	case p.file == "-":
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("--%s -: %w", ReasonFile, err)
		}
		s := strings.TrimRight(string(b), "\n")
		if s == "" {
			return "", fmt.Errorf("--%s - was given but stdin was empty", ReasonFile)
		}
		return s, nil
	case p.file != "":
		b, err := os.ReadFile(p.file)
		if err != nil {
			return "", fmt.Errorf("--%s: %w", ReasonFile, err)
		}
		return strings.TrimRight(string(b), "\n"), nil
	default:
		return p.inline, nil
	}
}

// String is the resolved representation, for a caller that already read it. Empty before Read.
func (p *Prose) String() string { return p.resolved }

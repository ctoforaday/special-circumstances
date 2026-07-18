// Command feov-record is the seat-side record runtime: ONE binary whose role
// subcommands carry bespoke verb sets, because the verb set IS the role boundary.
//
//	feov-record lens  <verb> --run <runDir> --seat-id <SEAT_ID> [...]
//	feov-record merge <verb> ...
//	feov-record blue  <verb> ...
//	feov-record bench <verb> ...
//
// A lens structurally cannot mint or close a board gap: no such verb exists in
// its namespace. Blue has no board verbs at all — additive-only and
// never-touch-the-ledger are topology here, not obedience. The bench rules and
// never originates.
//
// Port of tools/{red-lens,red-merge,blue,bench}.mjs (plans/record-tool.md R2g).
// FAITHFUL-FIRST: semantics are exact to the frozen mjs oracle; the role/seat-id
// cross-check, board.json render, and schema versioning are R2g.2 and land as
// separate commits so the differential gate stays meaningful.
package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

func main() {
	// Arm abort-safety before any write path can run: a seat killed mid-command
	// must lose an event, never leave a torn one or a stuck lock.
	defer record.InstallSignalGuard()()

	root := &cobra.Command{
		Use:   "feov-record",
		Short: "the seat-side record runtime (the verb set IS the role boundary)",
		Long: `feov-record — the seat-side record runtime (the verb set IS the role boundary)

A lens structurally cannot mint or close a board gap: no such verb exists in its
namespace. Blue has no board verbs at all. The bench rules and never originates.`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		record.RoleCommand("lens", "red lens seats — findings, observations, citations. Cannot mint or close.", lensVerbs()),
		record.RoleCommand("merge", "the red merge seat — the board's only writer.", mergeVerbs()),
		record.RoleCommand("blue", "blue seats — revisions, manifest rows, disputes. No board verbs at all.", blueVerbs()),
		record.RoleCommand("bench", "the bench — opinions, petition rulings, halt, certification. Never originates.", benchVerbs()),
	)

	if err := root.Execute(); err != nil {
		// Errors are printed bare, without cobra's usage dump: a validation
		// refusal here is a TEACHING message a seat reads and acts on, and
		// burying it under a flag listing is how it stops being read.
		fmt.Fprintf(os.Stderr, "feov-record: %v\n", err)
		os.Exit(2)
	}
}

// version is stamped on register events at R2g.2; kept here so --version works
// for the setup preflight that fails a skewed binary BEFORE the run-live marker
// is written (a mid-round failure is the one that costs a seat).
const version = "0.1.0"

func register(runDir, seatID string, _ *record.Args) (string, error) {
	nonce, _, err := record.RegisterSeat(runDir, seatID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("registered %s (shard nonce %s)", seatID, nonce), nil
}

var registerVerb = record.Verb{
	Name: "register",
	Help: "FIRST ACTION at the seat: register --run <runDir> --seat-id <SEAT_ID from your prompt>",
	Fn:   register,
}

// ---- lens: findings are lens-scoped observations the MERGE disposes and mints from ----

func lensVerbs() []record.Verb {
	return []record.Verb{
		registerVerb,
		{
			Name:  "finding",
			Help:  `a lens-scoped finding: --label L3-F1 --severity <g> --likelihood <g> --impact <g> [--text "..."|--file f] [--location "..."]`,
			Flags: []string{"label", "severity", "likelihood", "impact", "location"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				text, err := record.PayloadText(a)
				if err != nil {
					return "", err
				}
				p := a.Copy(record.NewPayload(), "label", "severity", "likelihood", "impact", "location")
				p.Set("text", text)
				if _, err := record.Append(runDir, seatID, "finding", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("finding %s recorded", a.Str("label")), nil
			},
		},
		{
			Name:  "observe",
			Help:  "a below-bar observation with a FATE (the merge must dispose it): --kind note|checked-held [--label ...] --text|--file",
			Flags: []string{"kind", "label"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				text, err := record.PayloadText(a)
				if err != nil {
					return "", err
				}
				kind := a.Str("kind")
				if kind == "" {
					kind = "note"
				}
				p := a.Copy(record.NewPayload().Set("kind", kind), "label")
				p.Set("text", text)
				ev, err := record.Append(runDir, seatID, "observe", p)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("observation recorded (%s) — awaiting merge disposition", ev.Key), nil
			},
		},
		{
			Name:  "cite",
			Help:  `citation-ledger row (the cross-round re-fetch gate): --claim "..." --reference "..." --confidence high|medium|low --access-date YYYY-MM-DD`,
			Flags: []string{"claim", "reference", "confidence", "access-date"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				p := a.Copy(record.NewPayload(), "claim", "reference", "confidence", "access_date:access-date")
				if _, err := record.Append(runDir, seatID, "cite", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("citation recorded: %s", a.Str("reference")), nil
			},
		},
		frictionVerb("attributed friction (survives aborts as an event): --text|--file"),
		petitionVerb("petition the bench (never sanctioned; does not pause your duties): --class ethical|safety|integrity|constitutional --basis \"...\" --relief \"...\"",
			" — the bench hears it before the debate continues"),
	}
}

// ---- merge: mint/close/dispose/regrade live here and nowhere else ----

func mergeVerbs() []record.Verb {
	return []record.Verb{
		{Name: "register", Help: "FIRST ACTION at the seat: register --run <runDir> --seat-id <SEAT_ID>", Fn: register},
		{
			Name:  "mint",
			Help:  `mint a board gap (id is TOOL-assigned; --key <stable-label> makes retries idempotent): --class <slug>|--class-new <slug> --definition --neighbor --distinguisher, --location "..." --problem "..."|--file --fix "..." --check "<acceptance check red runs at re-audit>" --severity/--likelihood/--impact/--cx <grade> [--supersedes R1-2,R1-7] [--found-by L5-F3,L6-F2]`,
			Flags: []string{"key", "class", "class-new", "definition", "neighbor", "distinguisher", "location", "problem", "fix", "check", "severity", "likelihood", "impact", "cx", "supersedes", "found-by"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				// Crash-retry idempotency: --key (the stable local label, e.g. the source
				// lens finding) makes a retried mint return the EXISTING id.
				prior, err := record.ExistingMintByKey(runDir, seatID, a.Str("key"))
				if err != nil {
					return "", err
				}
				if prior != "" {
					return fmt.Sprintf("minted %s (idempotent retry — existing id returned)", prior), nil
				}
				gapID, err := record.MintGapID(runDir, record.RoundOf(seatID))
				if err != nil {
					return "", err
				}
				text, err := record.PayloadText(a)
				if err != nil {
					return "", err
				}
				problem := a.Str("problem")
				if problem == "" {
					problem = text // `a.problem || payloadText(a)` — always present, possibly ""
				}
				p := record.NewPayload().Set("gap_id", gapID)
				a.SetInto(p, "mint_key", "key")
				if a.Has("class-new") {
					p.Set("class", a.Val("class-new"))
				} else if a.Has("class") {
					p.Set("class", a.Val("class"))
				}
				p.Set("class_new", a.Has("class-new")) // `!!` — always present
				a.Copy(p, "definition", "neighbor", "distinguisher", "location")
				p.Set("problem", problem)
				a.Copy(p, "required_fix:fix", "acceptance_check:check", "severity", "likelihood", "impact", "complexity_cost:cx")
				p.Set("supersedes", record.List(a.Str("supersedes")))
				p.Set("found_by", record.List(a.Str("found-by")))
				if _, err := record.Append(runDir, seatID, "mint", p); err != nil {
					return "", err
				}
				if a.Has("class-new") {
					cn := a.Copy(record.NewPayload().Set("slug", a.Val("class-new")), "definition", "neighbor", "distinguisher")
					if _, err := record.Append(runDir, seatID, "class-new", cn); err != nil {
						return "", err
					}
				}
				return fmt.Sprintf("minted %s", gapID), nil
			},
		},
		{
			Name:  "close",
			Help:  `close a gap WITH its verification anchor: --id R2-3 --as closed|closed_with_regression|... (--anchor-seat L1 --anchor-tool "git show" --anchor-target "7bc501e:path" | --carried-from <round>) [--successor R3-1] [--prose-file f]`,
			Flags: []string{"id", "as", "anchor-seat", "anchor-tool", "anchor-target", "carried-from", "successor", "prose-file"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				class := a.Str("as")
				if class == "" {
					class = "closed"
				}
				p := a.Copy(record.NewPayload(), "gap_id:id")
				p.Set("closure_class", class)
				a.Copy(p, "anchor_seat:anchor-seat", "anchor_tool:anchor-tool", "anchor_target:anchor-target",
					"carried_from:carried-from", "successor")
				if f := a.Str("prose-file"); f != "" {
					b, err := os.ReadFile(f)
					if err != nil {
						return "", err
					}
					p.Set("prose", string(b))
				}
				if _, err := record.Append(runDir, seatID, "close", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("closed %s (%s)", a.Str("id"), class), nil
			},
		},
		{
			Name:  "dispose",
			Help:  `every lens observation gets a fate: --observation <label-or-key> --as minted-as|folded-into|declined|banked [--into R2-4] [--reason "..."]`,
			Flags: []string{"observation", "as", "into", "reason"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				p := a.Copy(record.NewPayload(), "observation", "disposition:as", "into", "reason")
				if _, err := record.Append(runDir, seatID, "dispose", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("disposed %s: %s", a.Str("observation"), a.Str("as")), nil
			},
		},
		{
			Name:  "regrade",
			Help:  `same-id grade movement, recorded with its reason: --id R2-5 [--severity/--likelihood/--impact/--cx <grade>] --basis "..."`,
			Flags: []string{"id", "severity", "likelihood", "impact", "cx", "basis"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				p := a.Copy(record.NewPayload(), "gap_id:id", "severity", "likelihood", "impact", "complexity_cost:cx", "basis")
				if _, err := record.Append(runDir, seatID, "regrade", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("regraded %s", a.Str("id")), nil
			},
		},
		{
			Name:  "dispute-respond",
			Help:  `red's answer to a blue grade dispute: --id <gap> --as accepted|rejected --rationale "..."`,
			Flags: []string{"id", "as", "rationale"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				p := a.Copy(record.NewPayload(), "gap_id:id", "response:as", "rationale")
				if _, err := record.Append(runDir, seatID, "dispute-respond", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("dispute on %s: %s", a.Str("id"), a.Str("as")), nil
			},
		},
		{
			Name:  "spot-check",
			Help:  `the round archive spot-check record (W1.8 duty): --ids R1-4,R2-7 [--notes "..."]`,
			Flags: []string{"ids", "notes"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				ids := record.List(a.Str("ids"))
				p := a.Copy(record.NewPayload().Set("ids", ids), "notes")
				if _, err := record.Append(runDir, seatID, "spot-check", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("spot-checked %s", strings.Join(ids, ", ")), nil
			},
		},
		positionVerb("the round's ### RED section (prose via --file)"),
		closingVerb("a ### RED CLOSING entry per docketed gap: --id <gap> --file <prose>"),
		{
			Name:  "verdict",
			Help:  "the seat's terminal act: --as PASS|FAIL — renders all projections and checkpoints records/ to the recovery mirror",
			Flags: []string{"as"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				p := a.Copy(record.NewPayload(), "verdict:as")
				if _, err := record.Append(runDir, seatID, "verdict", p); err != nil {
					return "", err
				}
				r, err := record.Render(runDir, "")
				if err != nil {
					return "", err
				}
				mirror, err := checkpoint(runDir)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("verdict %s — rendered (%d open, %d closed) and checkpointed to %s",
					a.Str("as"), r.Open, r.Closed, mirror), nil
			},
		},
		frictionVerb("attributed friction: --text|--file"),
		petitionVerb(`petition the bench: --class ethical|safety|integrity|constitutional --basis "..." --relief "..."`, ""),
	}
}

// ---- blue: NO board verbs at all ----

func blueVerbs() []record.Verb {
	return []record.Verb{
		registerVerb,
		{
			Name:  "revision",
			Help:  "the round-record event (the CHANGELOG entry body via --file) — singleton per seat-round; emit AFTER your report edits land",
			Flags: []string{"claim-count"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				text, err := record.PayloadText(a)
				if err != nil {
					return "", err
				}
				p := record.NewPayload().Set("text", text)
				// `a['claim-count'] ? Number(...) : undefined` — absent and "0" both drop.
				if cc := a.Str("claim-count"); cc != "" && cc != "0" {
					p.Set("claim_count", jsNumber(cc))
				}
				if _, err := record.Append(runDir, seatID, "revision", p); err != nil {
					return "", err
				}
				return "revision recorded — the round is on the record", nil
			},
		},
		{
			Name:  "manifest-row",
			Help:  `one correctness-manifest receipt per repaired gap: --id R2-3 --row "figures recomputed; acceptance check run: pass; sites swept: S2,S4"`,
			Flags: []string{"id", "row"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				text, err := record.PayloadText(a)
				if err != nil {
					return "", err
				}
				row := a.Str("row")
				if row == "" {
					row = text
				}
				p := a.Copy(record.NewPayload(), "gap_id:id")
				p.Set("row", row)
				if _, err := record.Append(runDir, seatID, "manifest-row", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("manifest row recorded for %s", a.Str("id")), nil
			},
		},
		{
			Name:  "dispute",
			Help:  `contest a grade through the accounted channel: --id <gap> --dimension severity|likelihood|impact|complexity_cost --proposed <grade> --evidence "..."`,
			Flags: []string{"id", "dimension", "proposed", "evidence"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				p := a.Copy(record.NewPayload(), "gap_id:id", "dimension", "proposed", "evidence")
				if _, err := record.Append(runDir, seatID, "dispute", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("dispute filed on %s.%s", a.Str("id"), a.Str("dimension")), nil
			},
		},
		{
			Name:  "confidence",
			Help:  "per-claim confidence (the calibration substrate): --label <claim-label> --grade high|medium|low",
			Flags: []string{"label", "grade"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				p := a.Copy(record.NewPayload(), "label", "grade")
				if _, err := record.Append(runDir, seatID, "confidence", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("confidence recorded for %s", a.Str("label")), nil
			},
		},
		positionVerb("the round's ### BLUE section (prose via --file)"),
		closingVerb("a ### BLUE CLOSING entry per docketed gap: --id <gap> --file <prose>"),
		frictionVerb("attributed friction (survives aborts as an event): --text|--file"),
		petitionVerb(`petition the bench (never sanctioned; does not pause your duties): --class ethical|safety|integrity|constitutional --basis "..." --relief "..."`, ""),
	}
}

// ---- bench: rules, never originates ----

func benchVerbs() []record.Verb {
	return []record.Verb{
		{Name: "register", Help: "FIRST ACTION at the sitting: register --run <runDir> --seat-id <SEAT_ID from your prompt>", Fn: register},
		{
			Name:  "opinion",
			Help:  `a ruling as an OPINION: --gap-id R3-2 --disposition carried|closed|... --principle "..." --tension "correctness vs economy" --review-flag "why a human should look" [--file rationale]`,
			Flags: []string{"gap-id", "disposition", "principle", "tension", "review-flag"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				text, err := record.PayloadText(a)
				if err != nil {
					return "", err
				}
				p := a.Copy(record.NewPayload(), "gap_id:gap-id", "disposition", "principle", "tension", "review_flag:review-flag")
				p.Set("rationale", text)
				if _, err := record.Append(runDir, seatID, "opinion", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("opinion recorded: %s %s", a.Str("gap-id"), a.Str("disposition")), nil
			},
		},
		{
			Name:  "petition-rule",
			Help:  "rule on a petition: --petitioner <seatId> --class <class> --as granted|denied --file <opinion> (a halt is its own verb)",
			Flags: []string{"petitioner", "class", "as"},
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				text, err := record.PayloadText(a)
				if err != nil {
					return "", err
				}
				p := a.Copy(record.NewPayload(), "petitioner", "class", "ruling:as")
				p.Set("opinion", text)
				if _, err := record.Append(runDir, seatID, "petition-rule", p); err != nil {
					return "", err
				}
				return fmt.Sprintf("petition %s (%s)", a.Str("as"), a.Str("petitioner")), nil
			},
		},
		{
			Name: "halt",
			Help: "the safety boundary: --file <written opinion — capture relays it like a FAIL, never smoothed>",
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				text, err := record.PayloadText(a)
				if err != nil {
					return "", err
				}
				if _, err := record.Append(runDir, seatID, "halt", record.NewPayload().Set("opinion", text)); err != nil {
					return "", err
				}
				return "JUDICIAL HALT recorded — capture relays this verbatim", nil
			},
		},
		{
			Name: "certify",
			Help: `the run-end certification statement ("what I would want a human to re-examine"): --file <statement>`,
			Fn: func(runDir, seatID string, a *record.Args) (string, error) {
				text, err := record.PayloadText(a)
				if err != nil {
					return "", err
				}
				if _, err := record.Append(runDir, seatID, "certify", record.NewPayload().Set("statement", text)); err != nil {
					return "", err
				}
				return "certification recorded", nil
			},
		},
		frictionVerb("attributed friction (the bench seats had no event path before — the abort-surviving copy): --text|--file"),
	}
}

// ---- verbs shared across roles ----

func positionVerb(help string) record.Verb {
	return record.Verb{Name: "position", Help: help, Fn: func(runDir, seatID string, a *record.Args) (string, error) {
		text, err := record.PayloadText(a)
		if err != nil {
			return "", err
		}
		if _, err := record.Append(runDir, seatID, "position", record.NewPayload().Set("text", text)); err != nil {
			return "", err
		}
		return "position recorded", nil
	}}
}

func closingVerb(help string) record.Verb {
	return record.Verb{Name: "closing", Help: help, Flags: []string{"id"}, Fn: func(runDir, seatID string, a *record.Args) (string, error) {
		text, err := record.PayloadText(a)
		if err != nil {
			return "", err
		}
		p := a.Copy(record.NewPayload(), "gap_id:id")
		p.Set("text", text)
		if _, err := record.Append(runDir, seatID, "closing", p); err != nil {
			return "", err
		}
		return fmt.Sprintf("closing filed for %s", a.Str("id")), nil
	}}
}

func frictionVerb(help string) record.Verb {
	return record.Verb{Name: "friction", Help: help, Fn: func(runDir, seatID string, a *record.Args) (string, error) {
		text, err := record.PayloadText(a)
		if err != nil {
			return "", err
		}
		if _, err := record.Append(runDir, seatID, "friction", record.NewPayload().Set("text", text)); err != nil {
			return "", err
		}
		return "friction recorded", nil
	}}
}

func petitionVerb(help, suffix string) record.Verb {
	return record.Verb{Name: "petition", Help: help, Flags: []string{"class", "basis", "relief"}, Fn: func(runDir, seatID string, a *record.Args) (string, error) {
		p := a.Copy(record.NewPayload(), "class", "basis", "relief")
		if _, err := record.Append(runDir, seatID, "petition", p); err != nil {
			return "", err
		}
		return fmt.Sprintf("petition filed (%s)%s", a.Str("class"), suffix), nil
	}}
}

// jsNumber mirrors Number(str): an unparseable value becomes NaN, which
// JSON.stringify writes as null.
func jsNumber(s string) any {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return nil
	}
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return int64(f)
	}
	return f
}

// checkpoint mirrors the verdict verb's recovery mirror: records/ copied to a
// per-run directory under the user cache, keyed by a hash of the run dir.
func checkpoint(runDir string) (string, error) {
	sum := sha1.Sum([]byte(runDir))
	hash := hex.EncodeToString(sum[:])[:12]
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	mirror := filepath.Join(home, ".cache", "feov", "run-mirror", hash)
	if err := os.MkdirAll(mirror, 0o755); err != nil {
		return "", err
	}
	src := filepath.Join(runDir, "records")
	if _, err := os.Stat(src); err == nil {
		if err := copyDir(src, mirror); err != nil {
			return "", err
		}
	}
	return mirror, nil
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		s, d := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(s, d); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

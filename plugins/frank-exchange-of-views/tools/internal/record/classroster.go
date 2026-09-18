package record

import (
	"database/sql"
	"fmt"
	"sort"
)

// THE ROSTER IS THE CHECK'S OWN SOURCES, READ BACK.
//
// `class new --neighbor` refuses a value the registry does not know, and until now nothing on any
// seat's surface could say what the registry knows. A lens coining `coupled-verification` had to
// leave the record tool entirely and grep the checked-in Go source and feov-memory/class-registry.json
// — outside the run directory, which is the one place a seat is told it should never need to go,
// because every read is supposed to be a projection of the record.
//
// It reads loadRegistry and the class_new rows — the SAME two sources knownClasses consults, in the
// same order — rather than re-deriving the vocabulary from anywhere else. A list assembled from a
// second source is a second hand-kept copy of the registry, and it would drift from the check that
// refuses on it without either one being wrong on its own.
type ClassRow struct {
	Slug string `json:"slug"`
	// MaterialDefault decides whether this class's open gaps hold the PASS gate, so a seat picking
	// a class is picking this too. "" where the staged row carried none — which loadRegistry
	// refuses, so it cannot reach here from the registry, only from an older coined row.
	MaterialDefault string `json:"material_default"`
	// Coined is true for a class this RUN minted with `class new`. The distinction is not
	// decoration: a coined class carries its definition, neighbour and distinguisher on the record,
	// while a shipped one carries only the slug and default setup staged — so the columns a seat
	// can see differ, and it should know why rather than read the blanks as missing data.
	Coined        bool   `json:"coined"`
	Definition    string `json:"definition,omitempty"`
	Neighbor      string `json:"neighbor,omitempty"`
	Distinguisher string `json:"distinguisher,omitempty"`
}

// ClassRoster returns every class a `--class` or `--neighbor` on this run will accept.
//
// The order is the registry's own, then this run's coined classes in slug order. Shipped first
// because that is the vocabulary a seat is choosing WITHIN; coined last because those are this
// run's additions to it and a reader wants to see them as such.
func ClassRoster(run Run) ([]ClassRow, error) {
	reg, err := loadRegistry(run)
	if err != nil {
		return nil, err
	}
	// THE SAME REFUSAL knownClasses MAKES, and for the same reason: with no registry staged every
	// slug would be accepted, so a roster printed here would be a list of what this run happens to
	// have coined, presented as the vocabulary. An empty list and a missing gate look identical.
	if reg == nil {
		return nil, fmt.Errorf("record: no gap-class registry is staged for this run — there is no vocabulary to list, and every --class would be accepted. run-setup stages it from feov-memory/class-registry.json")
	}

	rows := make([]ClassRow, 0, len(reg.Classes))
	seen := map[string]bool{}
	for _, c := range reg.Classes {
		md := ""
		if c.MaterialDefault != nil {
			md = *c.MaterialDefault
		}
		rows = append(rows, ClassRow{Slug: c.Slug, MaterialDefault: md})
		seen[c.Slug] = true
	}

	coined, err := coinedClasses(run)
	if err != nil {
		return nil, err
	}
	sort.Slice(coined, func(i, j int) bool { return coined[i].Slug < coined[j].Slug })
	for _, c := range coined {
		// A COINING CANNOT SHADOW A SHIPPED SLUG — validateClass refuses one that already exists —
		// so a duplicate here would mean the record disagrees with the registry staged beside it.
		// Skipping it silently would hide that; it is reported as the conflict it is.
		if seen[c.Slug] {
			return nil, fmt.Errorf("record: class %s is both staged in this run's registry and coined on its record — the two disagree about the vocabulary, and a roster cannot say which is authoritative", c.Slug)
		}
		rows = append(rows, c)
	}
	return rows, nil
}

// coinedClasses reads the classes this run minted, with the fields the coining recorded.
func coinedClasses(run Run) ([]ClassRow, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nil // a run that has recorded nothing has coined nothing
	}
	// THE COLUMNS ARE THE VIEW'S, CHECKED AGAINST A REAL RECORD. class_new carries no seat_id, and
	// material_default arrives as the bare word the flag takes (`by_grade`), not the protobuf enum
	// name — so nothing is translated on the way out and the column a seat reads is the value it
	// would type.
	sqlRows, err := db.Query(`SELECT "slug", "definition", "neighbor", "distinguisher", "material_default" FROM "class_new"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking the record for coined classes: %w", err)
	}
	defer sqlRows.Close()

	var out []ClassRow
	for sqlRows.Next() {
		var slug, def, nb, dist, md sql.NullString
		if err := sqlRows.Scan(&slug, &def, &nb, &dist, &md); err != nil {
			return nil, err
		}
		out = append(out, ClassRow{
			Slug:            slug.String,
			MaterialDefault: md.String,
			Coined:          true,
			Definition:      def.String,
			Neighbor:        nb.String,
			Distinguisher:   dist.String,
		})
	}
	return out, sqlRows.Err()
}

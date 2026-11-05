package fixture

import "fmt"

const base = `SELECT "a" FROM "t"`

// A constant nobody issues is not a statement the package runs.
const unused = `SELECT "never" FROM "issued"`

func issue(string) {}

func statements(where, table string) bool {
	issue(base + ` WHERE "b" = ?`)
	issue(base + where)
	issue(fmt.Sprintf(`INSERT INTO %q ("x") VALUES (?)`, table))
	issue(`UPDATE "t"
	          SET "a" = 1 WHERE "b" = ?`)
	return where == "WITH" // a verb alone is a word being compared, not a statement
}

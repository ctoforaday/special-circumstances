package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// writeView is what a write's validation reads: the writing transaction, and the run's files.
//
// VALIDATION ASKS THE TRANSACTION IT IS WRITTEN IN. A check made before the transaction opens is a
// check of a record another seat can move: seats write in parallel, and a mint, a close or a motion
// committed between the check and the insert admits the write against a state that no longer holds
// (#1252) — a PASS over a gap minted a moment earlier, a second near-identical gap, a lens's last
// mint taken twice. The transaction takes the write lock at its BEGIN (`_txlock=immediate`), so
// every question asked of q is asked of the record the insert lands on, and no other writer can
// commit until this one has.
//
// IT HOLDS NO Run, AND THAT IS THE POINT. The run's handle is one connection, and the transaction
// holds it: a read through the handle from inside — anything that takes a Run and opens the record —
// waits on that transaction forever. With no Run in reach, such a read does not compile. The same
// holds for the whole of a migration's batch (batch.go), which keeps its transaction open between
// acts: its caller reads through Batch.Querier, never through the run.
type writeView struct {
	// q is the writing transaction. Every question validation asks of the record goes to it.
	q recordsql.Querier
	// files is the run's files the checks read: setup's terms and the staged class registry.
	files runFiles
	// rendering is the write's report, rendered before the lock (landingReport); nil on a batch's
	// write, which renders under its own transaction.
	rendering *landingReport
}

// runFiles is the two files the write path reads besides the record: inputs/run-config.json (the
// run's terms) and the class registry setup staged beside the record. Neither is a record a seat
// writes, so neither can move under a write; they are read from the paths alone, with no Run.
type runFiles struct {
	dir     string // the run directory, which holds inputs/run-config.json
	records string // the records directory, which holds class-registry.json
}

// filesOf is run's files.
func filesOf(run Run) runFiles { return runFiles{dir: run.Dir(), records: run.Records()} }

// params is the run's terms (RunParams).
func (f runFiles) params() (Params, error) { return paramsIn(f.dir) }

// convergenceFraction is the run's convergence fraction (convergenceFractionIn).
func (f runFiles) convergenceFraction() (float64, error) { return convergenceFractionIn(f.dir) }

// registry is the staged class registry (loadRegistry).
func (f runFiles) registry() (*classRegistry, error) { return loadRegistryIn(f.records) }

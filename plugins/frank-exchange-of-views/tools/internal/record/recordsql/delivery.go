package recordsql

import (
	"database/sql"
	"fmt"
)

// AgentSittingRow is one row of the agent_sitting view: a seat's sitting opened under an agent, and
// its ends.
type AgentSittingRow struct {
	SittingID   int64
	SeatID      string
	Sitting     int
	Epoch       int
	AgentID     string
	NextOpening int64 // 0 where the seat has opened no later sitting
	Over        bool
	Closed      bool
	Admitted    bool
	SatThrough  bool
}

// DeliveryRow is one row of the report_delivery view: a completed delivery of report text to a seat.
type DeliveryRow struct {
	EventID    int64
	SeatID     string
	SittingID  int64 // 0 for a row in no sitting
	Head       int64
	SatThrough bool
	// Whole is a lens's delivery being the whole report; Lens is whether the seat is a lens at all.
	Whole, Lens bool
}

const (
	agentSittingsSQL = `SELECT "sitting_id", "seat_id", "sitting", "epoch", "agent_id", COALESCE("next_opening", 0),
  "over", "closed", "admitted", "sat_through" FROM "agent_sitting" ORDER BY "sitting_id"`
	// "whole" is named once: the view is flattened into this statement, so each mention of the
	// column is its whole expression again, a count over the lens's earlier deliveries.
	reportDeliveriesSQL = `SELECT "event_id", "seat_id", COALESCE("sitting_id", 0), "head", "sat_through", "whole"
  FROM "report_delivery" ORDER BY "event_id"`
)

// AgentSittings is the agent_sitting view, in opening order.
func AgentSittings(q Querier) ([]AgentSittingRow, error) {
	rows, err := q.Query(agentSittingsSQL)
	if err != nil {
		return nil, fmt.Errorf("reading agent_sitting: %w", err)
	}
	defer rows.Close()
	var out []AgentSittingRow
	for rows.Next() {
		var r AgentSittingRow
		if err := rows.Scan(&r.SittingID, &r.SeatID, &r.Sitting, &r.Epoch, &r.AgentID, &r.NextOpening,
			&r.Over, &r.Closed, &r.Admitted, &r.SatThrough); err != nil {
			return nil, fmt.Errorf("reading agent_sitting: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ReportDeliveries is the report_delivery view, in record order.
func ReportDeliveries(q Querier) ([]DeliveryRow, error) {
	rows, err := q.Query(reportDeliveriesSQL)
	if err != nil {
		return nil, fmt.Errorf("reading report_delivery: %w", err)
	}
	defer rows.Close()
	var out []DeliveryRow
	for rows.Next() {
		var r DeliveryRow
		var whole sql.NullBool
		if err := rows.Scan(&r.EventID, &r.SeatID, &r.SittingID, &r.Head, &r.SatThrough, &whole); err != nil {
			return nil, fmt.Errorf("reading report_delivery: %w", err)
		}
		r.Whole, r.Lens = whole.Bool, whole.Valid
		out = append(out, r)
	}
	return out, rows.Err()
}

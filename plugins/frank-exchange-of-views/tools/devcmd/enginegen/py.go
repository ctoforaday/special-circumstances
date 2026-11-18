package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

const (
	pyBeginMark  = "# BEGIN GENERATED SEATS — go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen writes this from the record's roster. DO NOT EDIT BY HAND."
	pyEndMark    = "# END GENERATED SEATS"
	pyVocabBegin = "# BEGIN GENERATED ENVELOPE VOCABULARY — go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen writes this from the record's enums, record.Plan and record.LensEnvelope. DO NOT EDIT BY HAND."
	pyVocabEnd   = "# END GENERATED ENVELOPE VOCABULARY"
)

var pyRegions = []region{
	{pyBeginMark, pyEndMark, "Python seat block",
		"the record's roster (seatclass.Seats and Lane, flags.LensAreas, the Occasion enum, record.MassMappingVersion). Edit the roster",
		func() ([]byte, error) { return renderPySeats(fromTheRecord()) }},
	{pyVocabBegin, pyVocabEnd, "Python envelope vocabulary",
		"the record's enums, record.Plan and record.LensEnvelope. Edit the declaration",
		func() ([]byte, error) {
			v, err := vocabularyFromTheRecord()
			if err != nil {
				return nil, err
			}
			return renderPyVocabulary(v)
		}},
}

func renderPySeats(s source) ([]byte, error) {
	ids := make([]string, 0, len(s.seats))
	for id, seat := range s.seats {
		if seat.Role == seatclass.OperatorRole {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)

	seatsMap := make(map[string]map[string]any, len(ids))
	lensOf := map[string]string{}
	for _, id := range ids {
		seat := s.seats[id]
		if seat.AgentType == "" {
			return nil, fmt.Errorf("the roster seats %s under no agent configuration", id)
		}
		tier := s.classOf(seat.Base)
		if tier == "" {
			return nil, fmt.Errorf("the roster gives %s the tier base %q, which has no tier class", id, seat.Base)
		}
		row := map[string]any{
			"role":      seat.Role,
			"tier":      tier,
			"agentType": strings.TrimPrefix(seat.AgentType, runlive.SeatTypePrefix),
		}
		if seat.Area != "" {
			lensOf[seat.Area] = id
			row["area"] = seat.Area
		}
		seatsMap[id] = row
	}
	for _, a := range s.areas {
		if lensOf[a] == "" {
			return nil, fmt.Errorf("the lens area %s has no seat on the roster", a)
		}
		delete(lensOf, a)
	}
	for a, id := range lensOf {
		return nil, fmt.Errorf("the roster seats %s for the area %s, which the record does not declare", id, a)
	}

	laneTier := s.classOf(s.lane.Base)
	if s.lane.AgentType == "" || laneTier == "" {
		return nil, fmt.Errorf("the roster gives lane invalid configuration")
	}
	laneMap := map[string]any{
		"prefix":    s.lanePrefix,
		"role":      s.lane.Role,
		"tier":      laneTier,
		"agentType": strings.TrimPrefix(s.lane.AgentType, runlive.SeatTypePrefix),
	}
	occasions := make(map[string]string, len(s.occasions))
	for _, w := range s.occasions {
		occasions[w] = w
	}

	var b strings.Builder
	writePyJSON(&b, "SEATS", seatsMap)
	writePyJSON(&b, "LANE", laneMap)
	writePyJSON(&b, "OCCASION", occasions)
	writePyJSON(&b, "MASS_MAPPING_VERSION", s.massVersion)
	return []byte(b.String()), nil
}

func renderPyVocabulary(v vocabulary) ([]byte, error) {
	if err := sound("record.Plan", v.plan); err != nil {
		return nil, err
	}
	if err := sound("record.LensEnvelope", v.lens); err != nil {
		return nil, err
	}
	var b strings.Builder
	writePyJSON(&b, "GRADE", map[string]any{"type": "string", "enum": v.grades})
	writePyJSON(&b, "DISPUTE_DIMENSION", map[string]any{"type": "string", "enum": v.dimensions})
	writePyJSON(&b, "RUN_OUTCOME", map[string]any{"type": "string", "enum": v.outcomes})
	writePyJSON(&b, "BENCH_HEARINGS", v.benchHearings)
	writePyJSON(&b, "GAP_ID", map[string]any{"type": "string", "pattern": "^" + anchor.IDPattern("gap") + "$"})

	clauses := make([]map[string]string, 0, len(v.standing))
	for _, c := range v.standing {
		clauses = append(clauses, map[string]string{"kind": string(c.Kind), "clause": c.Clause})
	}
	writePyJSON(&b, "STANDING_CLAUSES", clauses)

	var party []record.PlanField
	for _, f := range v.plan {
		if f.Of == v.party {
			party = f.Items
			break
		}
	}
	if party == nil {
		return nil, fmt.Errorf("record.Plan carries no list of parties")
	}

	benchParty := pyPartySchema(party, map[string]any{"type": "string", "enum": []string{v.benchSeat}}, true)
	lensOrBlueParty := pyPartySchema(party, map[string]any{"type": "string", "enum": v.lensOrBlue}, false)
	writePyJSON(&b, "BENCH_PARTY", benchParty)
	writePyJSON(&b, "LENS_OR_BLUE_PARTY", lensOrBlueParty)

	props := map[string]any{}
	required := []string{}
	fieldTypes := map[string]string{}
	itemTypes := map[string]map[string]string{}
	for _, f := range v.plan {
		props[f.JSON] = v.pySchemaOf(f, benchParty, lensOrBlueParty)
		if f.Optional {
			continue
		}
		required = append(required, f.JSON)
		fieldTypes[f.JSON] = f.Type
		if f.Of != nil {
			of := map[string]string{}
			for _, i := range f.Items {
				of[i.JSON] = i.Type
			}
			itemTypes[f.JSON] = of
		}
	}
	emptyOnlyWith := map[string]string{}
	for _, f := range v.plan {
		if f.EmptyOnlyWith != "" {
			emptyOnlyWith[f.JSON] = f.EmptyOnlyWith
		}
	}

	planSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             required,
		"properties":           props,
	}
	writePyJSON(&b, "PLAN", planSchema)
	writePyJSON(&b, "PLAN_FIELD_TYPES", fieldTypes)
	writePyJSON(&b, "PLAN_EMPTY_ONLY_WITH", emptyOnlyWith)
	writePyJSON(&b, "PLAN_ITEM_TYPES", itemTypes)

	readProps := map[string]any{}
	readReq := []string{}
	for _, f := range v.read {
		readProps[f.JSON] = v.pyScalarOrArraySchema(f)
		if !f.Optional {
			readReq = append(readReq, f.JSON)
		}
	}
	readSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             readReq,
		"properties":           readProps,
	}
	writePyJSON(&b, "RECORD_READ", readSchema)

	lensProps := map[string]any{"record_read": readSchema}
	lensReq := []string{}
	for _, f := range v.lens {
		lensProps[f.JSON] = v.pyScalarOrArraySchema(f)
		if !f.Optional {
			lensReq = append(lensReq, f.JSON)
		}
	}
	lensSchema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             lensReq,
		"properties":           lensProps,
	}
	writePyJSON(&b, "LENS_ENVELOPE", lensSchema)

	return []byte(b.String()), nil
}

func pyPartySchema(fields []record.PlanField, seatSchema map[string]any, bench bool) map[string]any {
	props := map[string]any{}
	req := []string{}
	for _, f := range fields {
		if !f.Optional {
			req = append(req, f.JSON)
		}
		if f.JSON == "seat_id" {
			props[f.JSON] = seatSchema
			continue
		}
		item := map[string]any{"type": "string"}
		if f.Type == "array" {
			arr := map[string]any{"type": "array", "items": item}
			if bench && f.JSON == "gap_ids" {
				arr["minItems"] = 1
			}
			props[f.JSON] = arr
		} else {
			props[f.JSON] = map[string]any{"type": f.Type}
		}
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             req,
		"properties":           props,
	}
}

func (v vocabulary) pyScalarOrArraySchema(f record.PlanField) map[string]any {
	if f.Type == "array" {
		return map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	}
	m := map[string]any{"type": f.Type}
	if len(f.Enum) > 0 {
		m["enum"] = f.Enum
	}
	if f.Desc != "" {
		m["description"] = f.Desc
	}
	return m
}

func (v vocabulary) pySchemaOf(f record.PlanField, benchParty, lensOrBlueParty map[string]any) map[string]any {
	if f.Of == v.party {
		return map[string]any{
			"type":  "array",
			"items": map[string]any{"oneOf": []any{benchParty, lensOrBlueParty}},
		}
	}
	if f.Of != nil {
		props := map[string]any{}
		req := []string{}
		for _, i := range f.Items {
			props[i.JSON] = v.pyScalarOrArraySchema(i)
			if !i.Optional {
				req = append(req, i.JSON)
			}
		}
		return map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             req,
				"properties":           props,
			},
		}
	}
	return v.pyScalarOrArraySchema(f)
}

func writePyJSON(b *strings.Builder, name string, val any) {
	raw, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(b, "%s = json.loads(%q)\n", name, string(raw))
}

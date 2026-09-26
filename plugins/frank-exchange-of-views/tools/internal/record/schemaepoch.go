package record

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

// EventSchema is the event-shape epoch this binary writes, aliased from where it is generated.
//
// AN ALIAS RATHER THAN A SECOND CONSTANT. The epoch is generated into recordpb because recordsql
// needs it — it stamps the epoch into every database it creates and compares it at every open — and
// recordsql cannot import this package, which imports it. Callers that already say
// `record.EventSchema` keep saying it, and an alias cannot drift from what it aliases the way a
// second generated copy could.
const EventSchema = recordpb.EventSchema

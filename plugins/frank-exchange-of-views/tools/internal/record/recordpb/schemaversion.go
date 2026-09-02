package recordpb

// CurrentSchemaVersion is the schema version this binary stamps on every event it writes.
//
// IT SURVIVED THE READER THAT MADE IT LOAD-BEARING. The constant lived beside ClassifyLine, whose
// stage 3 dropped a line carrying no schema_version as an incomplete write — that reader is gone
// with the shard lines it read, so the stamp is now written and stored (a column on `events`,
// like every other envelope field) and read back by nothing.
//
// THAT IS NOT THE EPOCH, and the two are easy to confuse. The compatibility gate a run actually
// passes is `record.EventSchema` — the event-shape epoch derived from the descriptors, which
// `setup` compares against the epoch the plugin beside the binary declares and refuses on
// mismatch. This constant is a field on the record; the epoch is a fact about the binary.
const CurrentSchemaVersion = SchemaVersion_SCHEMA_VERSION_1

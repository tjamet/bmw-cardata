package bmwcardata

// The BMW CarData API and streaming clients are hand-written directly against
// the official OpenAPI specifications (see the auth and cardataapi packages).
// The only generated artifacts left are data tables: the telemetry container
// catalogue and the MQTT reason-code reference, produced by the command below.
//
//go:generate go run ./cmd/generate/

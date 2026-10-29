// Package normalize turns ThaiWater DTOs into the domain types Station and
// Reading.
//
// This is where source quirks are handled explicitly and in one place:
// ICT timestamps without an offset become UTC instants, numeric strings are
// parsed, the bank-full and ground references are chosen per ADR 0003,
// zero-coded or missing references are treated as absent, and readings are
// flagged stale (relative to an injected clock) or suspect (outside plausible
// bounds). Pure functions, no I/O.
package normalize

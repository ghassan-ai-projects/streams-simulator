// Package adapter implements the declarative output adapter engine. An
// adapter is data — a JSON file projecting native sim-event-v0.1 records
// into a consumer's wire format through a closed set of transforms. The
// binary contains no consumer-specific code, no consumer's field names and
// no consumer's framing rules; adding a consumer must never require a
// release.
//
// The package is a facade: the rules live in its private domain layer and
// file reading in its files edge.
package adapter

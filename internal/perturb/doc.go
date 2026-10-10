// Package perturb implements the delivery perturbation layer, which sits
// between the world and the adapter: the world produces what physically
// happened, the perturbation layer produces what the observer got. It
// changes nothing physical and tests a consumer's ingest, time handling and
// deduplication. The delivery ledger records what each perturbation did, so
// a scenario the consumer never saw is scored as a transport miss, not a
// reasoning miss.
//
// The package is a facade: the rules live in its private domain layer.
package perturb

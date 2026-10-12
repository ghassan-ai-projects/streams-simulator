// Package release writes the release manifest: the simulator, domain, adapter,
// suite and consumer digests plus the author and independent review identity,
// optionally signed with stdlib ed25519. One artifact pins exactly what a
// benchmark release claims to be.
//
// The package is a facade: the rules live in its private domain layer, the
// use case in its app layer and the file access in its files edge.
package release

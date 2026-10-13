// Package component loads libraries of reusable equipment, instruments and
// vocabulary, checks them, and composes a plant built from them into an
// ordinary domain. A composed domain never depends on the library at run time.
//
// The package is a facade: the rules live in its private domain layer and
// directory reading in its files edge.
package component

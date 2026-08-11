// Package schemas embeds the simulator's own contract schemas
// (docs/contracts/*.schema.json) so the binary can validate its data
// artifacts without any external files. A test in this package proves each
// embedded copy is byte-identical to the committed original, so the binary
// can never silently validate against a drifted schema.
package schemas

import (
	_ "embed"
)

//go:embed domain-spec-v0.1.schema.json
var domainSpec []byte

//go:embed sim-event-v0.1.schema.json
var simEvent []byte

//go:embed output-adapter-v0.1.schema.json
var outputAdapter []byte

//go:embed consumer-verdict-v0.1.schema.json
var consumerVerdict []byte

//go:embed ground-truth-v0.1.schema.json
var groundTruth []byte

//go:embed run-artifact-v0.1.schema.json
var runArtifact []byte

// DomainSpec returns the embedded domain-spec schema.
func DomainSpec() []byte { return domainSpec }

// SimEvent returns the embedded sim-event schema.
func SimEvent() []byte { return simEvent }

// OutputAdapter returns the embedded output-adapter schema.
func OutputAdapter() []byte { return outputAdapter }

// ConsumerVerdict returns the embedded consumer-verdict schema.
func ConsumerVerdict() []byte { return consumerVerdict }

// GroundTruth returns the embedded ground-truth schema.
func GroundTruth() []byte { return groundTruth }

// RunArtifact returns the embedded run-artifact schema.
func RunArtifact() []byte { return runArtifact }

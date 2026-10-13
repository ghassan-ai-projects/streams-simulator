// Package schemas embeds the simulator's own contract schemas
// (contracts/*.schema.json) so the binary can validate its data
// artifacts without any external files. A test in this package proves each
// embedded copy is byte-identical to the committed original, so the binary
// can never silently validate against a drifted schema.
package schemas

import (
	_ "embed"
)

//go:embed domain-spec-v0.2.schema.json
var domainSpec []byte

//go:embed output-adapter-v0.1.schema.json
var outputAdapter []byte

//go:embed consumer-verdict-v0.1.schema.json
var consumerVerdict []byte

//go:embed ground-truth-v0.1.schema.json
var groundTruth []byte

//go:embed run-artifact-v0.1.schema.json
var runArtifact []byte

//go:embed scenario-v0.1.schema.json
var scenario []byte

//go:embed run-journal-v0.1.schema.json
var runJournal []byte

//go:embed run-report-v0.1.schema.json
var runReport []byte

func DomainSpec() []byte { return domainSpec }

func OutputAdapter() []byte { return outputAdapter }

func ConsumerVerdict() []byte { return consumerVerdict }

func GroundTruth() []byte { return groundTruth }

func RunArtifact() []byte { return runArtifact }

func Scenario() []byte { return scenario }

func RunJournal() []byte { return runJournal }

func RunReport() []byte { return runReport }

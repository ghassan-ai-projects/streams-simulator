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

//go:embed run-comparison-v0.1.schema.json
var runComparison []byte

func RunComparison() []byte { return runComparison }

//go:embed domain-assumptions-v0.1.schema.json
var domainAssumptions []byte

//go:embed domain-lineage-v0.1.schema.json
var domainLineage []byte

func DomainAssumptions() []byte { return domainAssumptions }

func DomainLineage() []byte { return domainLineage }

func DomainSpec() []byte { return domainSpec }

func OutputAdapter() []byte { return outputAdapter }

func ConsumerVerdict() []byte { return consumerVerdict }

func GroundTruth() []byte { return groundTruth }

func RunArtifact() []byte { return runArtifact }

func Scenario() []byte { return scenario }

func RunJournal() []byte { return runJournal }

func RunReport() []byte { return runReport }

//go:embed component-v0.1.schema.json
var component []byte

//go:embed instrument-v0.1.schema.json
var instrument []byte

//go:embed composition-v0.1.schema.json
var composition []byte

//go:embed component-vocabulary-v0.1.schema.json
var componentVocabulary []byte

func Component() []byte { return component }

func Instrument() []byte { return instrument }

func Composition() []byte { return composition }

func ComponentVocabulary() []byte { return componentVocabulary }

//go:embed twin-choices-v0.1.schema.json
var twinChoices []byte

func TwinChoices() []byte { return twinChoices }

//go:embed session-plan-v0.1.schema.json
var sessionPlan []byte

func SessionPlan() []byte { return sessionPlan }

//go:embed node-control-v0.1.schema.json
var nodeControl []byte

func NodeControl() []byte { return nodeControl }

//go:embed interrupt-log-v0.1.schema.json
var interruptLog []byte

func InterruptLog() []byte { return interruptLog }

//go:embed twin-report-v0.1.schema.json
var twinReport []byte

func TwinReport() []byte { return twinReport }

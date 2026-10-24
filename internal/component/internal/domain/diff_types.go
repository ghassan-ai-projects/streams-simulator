package domain

const (
	SourceFile = "file"
	SourceRun  = "run"

	StatusChanged = "changed"
	StatusAdded   = "added"
	StatusRemoved = "removed"

	SectionReadings  = "readings"
	SectionDynamics  = "dynamics"
	SectionCommands  = "commands"
	SectionFaults    = "faults"
	SectionStates    = "states"
	SectionInstances = "instances"
	SectionKind      = "kind"
	SectionLinks     = "links"
	SectionLinkTypes = "link_types"
	SectionDomain    = "domain"

	diffVersion = "0.1"
)

const (
	CodeKindAdded           = "kind_added"
	CodeKindRemoved         = "kind_removed"
	CodeKindChanged         = "kind_changed"
	CodeReadingAdded        = "reading_added"
	CodeReadingRemoved      = "reading_removed"
	CodeReadingChanged      = "reading_changed"
	CodeInstrumentChanged   = "instrument_changed"
	CodeUnitChanged         = "unit_changed"
	CodeRangeChanged        = "range_changed"
	CodeResolutionChanged   = "resolution_changed"
	CodeNoiseChanged        = "noise_changed"
	CodeCadenceChanged      = "cadence_changed"
	CodeAvailabilityChanged = "availability_changed"
	CodeDynamicsAdded       = "dynamics_added"
	CodeDynamicsRemoved     = "dynamics_removed"
	CodeDynamicsChanged     = "dynamics_changed"
	CodeGainChanged         = "gain_changed"
	CodeTimeConstantChanged = "time_constant_changed"
	CodeDeadTimeChanged     = "dead_time_changed"
	CodeClampChanged        = "clamp_changed"
	CodeInputAdded          = "input_added"
	CodeInputRemoved        = "input_removed"
	CodeInputChanged        = "input_changed"
	CodeCommandAdded        = "command_added"
	CodeCommandRemoved      = "command_removed"
	CodeCommandChanged      = "command_changed"
	CodeArgumentAdded       = "argument_added"
	CodeArgumentRemoved     = "argument_removed"
	CodeArgumentChanged     = "argument_changed"
	CodeSafeStopChanged     = "safe_stop_changed"
	CodeRiskClassChanged    = "risk_class_changed"
	CodeSetsChanged         = "sets_changed"
	CodeLeaseChanged        = "lease_changed"
	CodeLatencyChanged      = "latency_changed"
	CodeFailureModesChanged = "failure_modes_changed"
	CodeStateChangesChanged = "state_changes_changed"
	CodeFaultAdded          = "fault_added"
	CodeFaultRemoved        = "fault_removed"
	CodeFaultChanged        = "fault_changed"
	CodeStateAdded          = "state_added"
	CodeStateRemoved        = "state_removed"
	CodeStateChanged        = "state_changed"
	CodeInstancesChanged    = "instances_changed"
	CodeLinkAdded           = "link_added"
	CodeLinkRemoved         = "link_removed"
	CodeLinkTypeAdded       = "link_type_added"
	CodeLinkTypeRemoved     = "link_type_removed"
	CodeLinkTypeChanged     = "link_type_changed"
	CodeSectionChanged      = "section_changed"
)

type Diff struct {
	SchemaVersion string           `json:"schema_version"`
	A             DiffSide         `json:"a"`
	B             DiffSide         `json:"b"`
	Identity      DiffIdentity     `json:"identity"`
	Kinds         []KindDiff       `json:"kinds"`
	DomainChanges []Change         `json:"domain_changes"`
	Summary       DiffSummary      `json:"summary"`
	Unchanged     UnchangedSummary `json:"unchanged_summary"`
}

type DiffSide struct {
	ID      string     `json:"id"`
	Version string     `json:"version"`
	Title   string     `json:"title"`
	Digest  string     `json:"digest"`
	Source  DiffSource `json:"source"`
}

type DiffSource struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type DiffIdentity struct {
	SameID                          bool `json:"same_id"`
	SameVersion                     bool `json:"same_version"`
	SameDigest                      bool `json:"same_digest"`
	SameIDAndVersionDifferentDigest bool `json:"same_id_and_version_different_digest"`
}

type KindDiff struct {
	Kind      string        `json:"kind"`
	Status    string        `json:"status"`
	Instances KindInstances `json:"instances"`
	Relations []Relation    `json:"relations"`
	Changes   []Change      `json:"changes"`
}

type KindInstances struct {
	Before InstanceSummary `json:"before"`
	After  InstanceSummary `json:"after"`
}

type InstanceSummary struct {
	Count       int      `json:"count"`
	Names       []string `json:"names"`
	IDTemplates []string `json:"id_templates"`
}

type Change struct {
	Code     string `json:"code"`
	Subject  string `json:"subject"`
	Section  string `json:"section"`
	Field    string `json:"field"`
	Before   any    `json:"before"`
	After    any    `json:"after"`
	Sentence string `json:"sentence"`
}

type DiffSummary struct {
	Changes      int `json:"changes"`
	KindsChanged int `json:"kinds_changed"`
	KindsAdded   int `json:"kinds_added"`
	KindsRemoved int `json:"kinds_removed"`
}

type UnchangedSummary struct {
	Kinds        int `json:"kinds"`
	Readings     int `json:"readings"`
	Dynamics     int `json:"dynamics"`
	Commands     int `json:"commands"`
	Faults       int `json:"faults"`
	States       int `json:"states"`
	InstanceSets int `json:"instance_sets"`
	Links        int `json:"links"`
	LinkTypes    int `json:"link_types"`
}

type DiffInput struct {
	Raw   []byte
	Label string
	Run   bool
}

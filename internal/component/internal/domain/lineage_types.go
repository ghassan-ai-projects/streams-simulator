package domain

const (
	SourceInSync           = "in_sync"
	SourceOutOfSync        = "out_of_sync"
	SourceDoesNotCompose   = "does_not_compose"
	SourceDoesNotScaffold  = "does_not_scaffold"
	SourceKindFilesMissing = "kind_files_missing"

	KindFound   = "found"
	KindMissing = "missing"

	VerdictEditable = "editable"
	VerdictReadOnly = "read_only"

	ReasonComposition          = "composition"
	ReasonCompositionOutOfSync = "composition_out_of_sync"
	ReasonTwin                 = "twin"
	ReasonTwinOutOfSync        = "twin_out_of_sync"
	ReasonExample              = "example"
	ReasonHandAuthored         = "hand_authored"

	RouteEditComposition    = "edit_composition"
	RouteRecompose          = "recompose"
	RouteOpenDeviceBehavior = "open_device_behavior"
	RouteRescaffoldTwin     = "rescaffold_twin"
	RouteCreateFromExample  = "create_from_example"
	RouteCopyDomainFile     = "copy_domain_file"

	UnreadableKind        = "kind_file_unreadable"
	UnreadableTwin        = "twin_file_unreadable"
	UnreadableComposition = "composition_file_unreadable"
	UnreadableRun         = "run_unreadable"

	lineageVersion = "0.1"
)

type NamedFile struct {
	Name string
	Raw  []byte
	Err  error
}

type RunFolder struct {
	ID       string
	RunJSON  []byte
	Observer []byte
	Err      error
}

type Workspace struct {
	Domains      []NamedFile
	Kinds        []NamedFile
	Twins        []NamedFile
	Compositions []NamedFile
	Examples     []NamedFile
	Runs         []RunFolder
}

type Lineage struct {
	SchemaVersion string              `json:"schema_version"`
	Plants        []PlantLineage      `json:"plants"`
	DeviceKinds   []DeviceKindLineage `json:"device_kinds"`
	Unreadable    []Unreadable        `json:"unreadable"`
}

type PlantDomain struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Title   string `json:"title"`
	Digest  string `json:"digest"`
	File    string `json:"file"`
}

type PlantLineage struct {
	Domain      PlantDomain    `json:"domain"`
	Composition *SourceLineage `json:"composition,omitempty"`
	Twin        *SourceLineage `json:"twin,omitempty"`
	DeviceKinds []PlantKind    `json:"device_kinds"`
	Runs        []RunLineage   `json:"runs"`
	Editable    Editable       `json:"editable"`
}

type SourceLineage struct {
	File           string `json:"file"`
	ID             string `json:"id"`
	Version        string `json:"version"`
	Status         string `json:"status"`
	ProducesDigest string `json:"produces_digest,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type PlantKind struct {
	Kind   string `json:"kind"`
	Status string `json:"status"`
	File   string `json:"file,omitempty"`
	Digest string `json:"digest,omitempty"`
}

type RunLineage struct {
	ID            string `json:"id"`
	Mode          string `json:"mode,omitempty"`
	Recorded      bool   `json:"recorded"`
	DomainVersion string `json:"domain_version"`
	DomainDigest  string `json:"domain_digest"`
	SameAsCurrent bool   `json:"same_as_current"`
}

type Editable struct {
	Verdict string `json:"verdict"`
	Reason  string `json:"reason"`
	Route   Route  `json:"route"`
}

type Route struct {
	Code     string `json:"code"`
	Sentence string `json:"sentence"`
}

type DeviceKindLineage struct {
	Kind           string           `json:"kind"`
	File           string           `json:"file"`
	Digest         string           `json:"digest"`
	TwinsIncluding []TwinMembership `json:"twins_including"`
	HasOwnTwin     bool             `json:"has_own_twin"`
	ImplementedBy  []string         `json:"implemented_by"`
}

type TwinMembership struct {
	Twin  string `json:"twin"`
	File  string `json:"file"`
	Kinds int    `json:"kinds"`
}

type Unreadable struct {
	File    string `json:"file"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

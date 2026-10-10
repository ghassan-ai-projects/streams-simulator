package perturb

import "github.com/ghassan-ai-projects/streams-simulator/internal/perturb/internal/domain"

// Delivered is one delivered record: the (possibly modified) event plus its
// ledger metadata. Drop yields Delivered=false; duplicate yields extra
// records with the same seq.
type Delivered = domain.Delivered

// Perturbation names (the catalog in TECHNICAL_DESIGN §5.2).
const (
	DuplicateBurst = domain.DuplicateBurst
	IDReuse        = domain.IDReuse
	Reorder        = domain.Reorder
	DelayTail      = domain.DelayTail
	GrossBackfill  = domain.GrossBackfill
	Drop           = domain.Drop
	ClockSkew      = domain.ClockSkew
	NonMonotonic   = domain.NonMonotonic
	OutOfEnum      = domain.OutOfEnum
	OutOfRange     = domain.OutOfRange
	UnitMismatch   = domain.UnitMismatch
	Oversize       = domain.Oversize
	Malformed      = domain.Malformed
	NaNInf         = domain.NaNInf
	Storm          = domain.Storm
	ProducerFlap   = domain.ProducerFlap
	TimeEncoding   = domain.TimeEncoding
	PrecisionEdge  = domain.PrecisionEdge
	InjectionProbe = domain.InjectionProbe
)

// Names is the full catalog in a stable order.
var Names = domain.Names

package model

// SimEvent is one observation emitted by a simulated world, matching
// docs/contracts/sim-event-v0.1.schema.json. Deliberately minimal: anything
// a consumer could compute from the nameplate is not carried here, and the
// envelope must never carry perturbation metadata, fault identity, hidden
// state, or scenario labels.
type SimEvent struct {
	Seq          int64  `json:"seq"`
	WorldID      string `json:"world_id"`
	EntityType   string `json:"entity_type"`
	EntityID     string `json:"entity_id"`
	Channel      string `json:"channel"`
	EventTime    string `json:"event_time"`
	ObservedTime string `json:"observed_time"`
	Value        any    `json:"value,omitempty"`
	Unit         string `json:"unit,omitempty"`
	Birth        bool   `json:"birth,omitempty"`
}

// Delivery reasons recorded in the ledger for one native event. They are the
// mechanism that makes a transport miss distinguishable from a reasoning
// miss.
const (
	DeliveryOK               = "ok"
	DeliveryDroppedByPerturb = "dropped_by_perturbation"
	DeliveryDuplicated       = "duplicated"
	DeliveryDelayed          = "delayed"
	DeliveryMangled          = "mangled"
	DeliveryRewritten        = "rewritten"
	DeliveryReordered        = "reordered"
	DeliveryOmitted          = "omitted_by_adapter"
	DeliverySinkError        = "sink_error"
)

// LedgerRecord is one row of the delivery ledger: what the world emitted and
// what the sink actually delivered. Director-side, always.
type LedgerRecord struct {
	DeliveryID     uint64 `json:"delivery_id"`
	Seq            int64  `json:"seq"`
	WorldID        string `json:"world_id"`
	EntityID       string `json:"entity_id"`
	Channel        string `json:"channel"`
	EventTimeNS    int64  `json:"event_time_ns"`
	ObservedTimeNS int64  `json:"observed_time_ns"`
	Delivered      bool   `json:"delivered"`
	DeliveryReason string `json:"delivery_reason"`
	WrittenAtNS    int64  `json:"written_at_ns"`
}

package domain

const (
	TraceGateway = "gateway"
	TraceRelay   = "relay"
	TraceNode    = "node"
	TraceHost    = "host"
)

const (
	relayName        = "relay"
	relayProtocol    = 1
	heartbeatsMissed = 3
)

type Traced struct {
	Direction string
	Record    map[string]any
	Field     bool
	Admitted  bool
}

type Moves struct {
	Up           []map[string]any
	Down         []map[string]any
	ToSim        []map[string]any
	Trace        []Traced
	Notes        []Note
	Err          error
	CloseHost    bool
	CloseGateway bool
	syncRouted   bool
}

func (m *Moves) saw(direction string, record map[string]any) {
	m.Trace = append(m.Trace, Traced{Direction: direction, Record: record})
}

type RelayStats struct {
	GatewaySessions int    `json:"gateway_sessions"`
	HostSessions    int    `json:"host_sessions"`
	Up              int    `json:"up"`
	Down            int    `json:"down"`
	Dropped         int    `json:"dropped"`
	Refused         string `json:"refused,omitempty"`
	Reason          string `json:"reason,omitempty"`
	HostRefused     string `json:"host_refused,omitempty"`
	ToSimulated     int    `json:"to_simulated,omitempty"`
	FromSimulated   int    `json:"from_simulated,omitempty"`
	UnknownDevice   int    `json:"unknown_device,omitempty"`
	Forged          int    `json:"forged,omitempty"`
	SimulatedLost   int    `json:"simulated_lost,omitempty"`
	TraceIncomplete bool   `json:"trace_incomplete,omitempty"`
}

type gatewaySide struct {
	admitted bool
	answered bool
	asked    bool
	session  string
	in       Receiver
	out      *Transmitter
	born     map[string]string
	silent   int64
}

type hostSide struct {
	connected bool
	shaken    bool
	beating   bool
	session   string
	in        Receiver
	out       *Transmitter
	state     map[string]any
	interval  int64
	silent    int64
	table     seqTable
	birthed   bool
	birthSeq  uint64
	simTable  seqTable

	gatewaySyncs syncLedger
	simSyncs     syncLedger
}

type Relay struct {
	admission      Admission
	tickMS         int64
	stats          RelayStats
	gateway        gatewaySide
	host           hostSide
	birth          Inbound
	bdSeq          uint64
	birthBD        uint64
	lastUp         int64
	gatewayDevices map[string]bool
	sim            *simSide
	plan           relayPlan
	routed         routeLedger
}

func NewRelay(tickMS int64) *Relay {
	return &Relay{admission: Admission{Environment: EnvironmentHardware, ProtocolMax: relayProtocol}, tickMS: tickMS, lastUp: -1,
		gateway: gatewaySide{born: map[string]string{}}}
}

func (r *Relay) Stats() RelayStats { return r.stats }

func (r *Relay) GatewayAdmitted() bool { return r.gateway.admitted }

func (r *Relay) WantsHost() bool { return r.gateway.admitted && !r.host.connected }

func (r *Relay) dropped(moves Moves) Moves {
	r.stats.Dropped++
	return moves
}

func (r *Relay) transmitUp(m *Moves, record Record) uint64 {
	r.host.out.Queue(record)
	sent := r.host.out.Drain()[0]
	m.Up = append(m.Up, sent)
	m.saw(TraceNode, sent)
	r.stats.Up++
	return Seq(sent["seq"])
}

func (r *Relay) transmitDown(m *Moves, record Record) {
	r.gateway.out.Queue(record)
	sent := r.gateway.out.Drain()[0]
	m.Down = append(m.Down, sent)
	m.saw(TraceRelay, sent)
	r.stats.Down++
}

func arrivalOf(receiver *Receiver, in Inbound) Arrival {
	if in.Unsequenced {
		return Duplicate
	}
	return receiver.Accept(in.Seq)
}

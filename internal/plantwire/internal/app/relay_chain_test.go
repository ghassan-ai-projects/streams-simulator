package app

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
)

const relaySensorKind = `{"kind":"thermal_sensor","metrics":[
 {"name":"temperature","data_type":"number","unit":"Cel","eu_range":{"min":-20,"max":80},"resolution":0.1,"writable":false},
 {"name":"humidity","data_type":"number","unit":"%","eu_range":{"min":0,"max":100},"resolution":1,"writable":false}],"commands":[]}`

const relayFanKind = `{"kind":"fan","metrics":[
 {"name":"duty_permille","data_type":"integer","unit":"1","eu_range":{"min":0,"max":600},"resolution":1,"writable":true}],
 "commands":[
 {"name":"set_pwm_lease","args_schema":{"type":"object","properties":{"lease_ms":{"type":"integer","minimum":5000,"maximum":10000}},"required":["lease_ms"],"additionalProperties":false},"risk_class":"medium","safe_stop":false},
 {"name":"stop_fan","args_schema":{"type":"object","properties":{},"required":[],"additionalProperties":false},"risk_class":"low","safe_stop":true}]}`

func relayKindFile(t *testing.T, text string) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal([]byte(text), &entry); err != nil {
		t.Fatal(err)
	}
	return entry
}

func relayGateway(t *testing.T, session string) *standin.Node {
	t.Helper()
	node, err := standin.New(standin.Config{Session: session, Board: "mega-01",
		Kinds: []map[string]any{relayKindFile(t, relaySensorKind), relayKindFile(t, relayFanKind)},
		Devices: []standin.Device{{Name: "zone-01", Kind: "thermal_sensor", BootID: "boot-1", Initial: map[string]any{"temperature": 21.5}},
			{Name: "fan-01", Kind: "fan", BootID: "boot-1"}},
		Script: standin.Script{Replies: []standin.Reply{{Command: "set_pwm_lease", Receipt: "accepted", Result: "succeeded"}}}})
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func relayHostOptions(environment string, script ...host.Scripted) host.Options {
	return host.Options{Environment: environment, ProtocolMax: 1, Epoch: 7, Policy: "sha256:" + strings.Repeat("b", 64), Script: script}
}

func relayCommandStep() host.Scripted {
	return host.Scripted{ID: "cmd-1", AtNS: 1, Device: "fan-01", Command: "set_pwm_lease", Args: map[string]any{"lease_ms": 5000}}
}

func decodeRecord(t *testing.T, record map[string]any) domain.Inbound {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	in, ok := domain.Decode(raw)
	if !ok {
		t.Fatalf("not a record: %s", raw)
	}
	return in
}

func exactContent(t *testing.T, record map[string]any, drop ...string) map[string]any {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var fields map[string]any
	if err := decoder.Decode(&fields); err != nil {
		t.Fatal(err)
	}
	for _, member := range append([]string{"type", "lane", "session_id", "protocol_version", "seq"}, drop...) {
		delete(fields, member)
	}
	return fields
}

type chain struct {
	t           *testing.T
	relay       *domain.Relay
	gateway     *standin.Node
	host        *host.Host
	state       map[string]any
	accepted    bool
	lose        func(index int) bool
	fromGateway []map[string]any
	fromHost    []map[string]any
	toHost      []map[string]any
	toGateway   []map[string]any
}

func newChain(t *testing.T, options host.Options) *chain {
	t.Helper()
	return &chain{t: t, relay: domain.NewRelay(250), gateway: relayGateway(t, "s-1"), host: host.New(options)}
}

func (c *chain) connect() {
	c.t.Helper()
	c.accepted = false
	c.relay.GatewayAttached()
	birth, err := c.gateway.Birth()
	if err != nil {
		c.t.Fatal(err)
	}
	c.feedGateway(birth)
	if c.relay.WantsHost() {
		c.pump(c.relay.HostAttached())
	}
}

func (c *chain) feedGateway(record map[string]any) {
	c.fromGateway = append(c.fromGateway, record)
	c.pump(c.relay.FromGateway(decodeRecord(c.t, record)))
}

func (c *chain) feedHost(record map[string]any) {
	c.fromHost = append(c.fromHost, record)
	c.pump(c.relay.FromHost(decodeRecord(c.t, record)))
}

func (c *chain) pump(moves domain.Moves) {
	for _, record := range moves.Up {
		c.toHost = append(c.toHost, record)
		if c.lose != nil && c.lose(len(c.toHost)-1) {
			continue
		}
		for _, reply := range c.host.Receive(decodeRecord(c.t, record)) {
			c.feedHost(reply)
		}
	}
	for _, record := range moves.Down {
		c.toGateway = append(c.toGateway, record)
		for _, answer := range c.react(record) {
			c.feedGateway(answer)
		}
	}
}

func (c *chain) react(record map[string]any) []map[string]any {
	switch record["type"] {
	case "STATE":
		return c.handshake(record)
	case "DCMD":
		return c.gateway.Handle(record)
	case "NCMD":
		return c.rebirth()
	}
	return nil
}

func (c *chain) handshake(state map[string]any) []map[string]any {
	if state["online"] != true {
		return nil
	}
	c.state = state
	if c.accepted {
		return nil
	}
	c.accepted = true
	births, err := c.gateway.Accept(state)
	if err != nil {
		c.t.Fatal(err)
	}
	return births
}

func (c *chain) rebirth() []map[string]any {
	birth, err := c.gateway.Birth()
	if err != nil {
		c.t.Fatal(err)
	}
	births, err := c.gateway.Accept(c.state)
	if err != nil {
		c.t.Fatal(err)
	}
	return append([]map[string]any{birth}, births...)
}

func (c *chain) reading(value float64, quality string) {
	c.t.Helper()
	for _, record := range c.gateway.Emit(domain.DeviceData("zone-01", "boot-1", false, []map[string]any{domain.Sample(1, value, quality, 1767225600123456789)})) {
		c.feedGateway(record)
	}
}

func (c *chain) barrier(throughSeq uint64) {
	c.t.Helper()
	for _, record := range c.gateway.Emit(domain.Sync(throughSeq, 1767225660000000000)) {
		c.feedGateway(record)
	}
}

func ofType(records []map[string]any, kind string) []map[string]any {
	var found []map[string]any
	for _, record := range records {
		if record["type"] == kind {
			found = append(found, record)
		}
	}
	return found
}

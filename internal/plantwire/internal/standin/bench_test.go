package standin

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
)

const sensorKind = `{"kind":"thermal_sensor","metrics":[
 {"name":"temperature","data_type":"number","unit":"Cel","eu_range":{"min":-20,"max":80},"resolution":0.1,"writable":false},
 {"name":"humidity","data_type":"number","unit":"%","eu_range":{"min":0,"max":100},"resolution":1,"writable":false}],"commands":[]}`

const fanKind = `{"kind":"fan","metrics":[
 {"name":"level_percent","data_type":"integer","unit":"%","eu_range":{"min":0,"max":60},"resolution":1,"writable":true}],
 "commands":[
 {"name":"set_cooling","args_schema":{"type":"object","properties":{"level_percent":{"type":"integer","minimum":0,"maximum":60}},"required":["level_percent"],"additionalProperties":false},"risk_class":"medium","safe_stop":false},
 {"name":"stop_fan","args_schema":{"type":"object","properties":{},"required":[],"additionalProperties":false},"risk_class":"low","safe_stop":true}]}`

var policy = "sha256:" + strings.Repeat("b", 64)

func kindFile(t *testing.T, text string) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal([]byte(text), &entry); err != nil {
		t.Fatal(err)
	}
	return entry
}

func boardConfig(t *testing.T) Config {
	t.Helper()
	return Config{Session: "s-1", Board: "mega-01", Kinds: []map[string]any{kindFile(t, sensorKind), kindFile(t, fanKind)},
		Devices: []Device{{Name: "zone-01", Kind: "thermal_sensor", BootID: "boot-1", Initial: map[string]any{"temperature": 21.5}},
			{Name: "fan-01", Kind: "fan", BootID: "boot-1"}}}
}

func boardNode(t *testing.T, script Script) *Node {
	t.Helper()
	config := boardConfig(t)
	config.Script = script
	node, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	return node
}

type bench struct {
	t    *testing.T
	node *Node
	host *host.Host
}

func newBench(t *testing.T, node *Node, options host.Options) *bench {
	t.Helper()
	return &bench{t: t, node: node, host: host.New(options)}
}

func hostOptions(environment string, script ...host.Scripted) host.Options {
	return host.Options{Environment: environment, ProtocolMax: 1, Epoch: 1, Policy: policy, Script: script}
}

func (b *bench) deliver(records ...map[string]any) []map[string]any {
	b.t.Helper()
	var replies []map[string]any
	for _, record := range records {
		raw, err := json.Marshal(record)
		if err != nil {
			b.t.Fatal(err)
		}
		inbound, ok := domain.Decode(raw)
		if !ok || inbound.Invalid != nil {
			b.t.Fatalf("record %s is no valid plant wire record: %v", raw, inbound.Invalid)
		}
		replies = append(replies, b.host.Receive(inbound)...)
	}
	return replies
}

func (b *bench) handshake() []map[string]any {
	b.t.Helper()
	birth, err := b.node.Birth()
	if err != nil {
		b.t.Fatal(err)
	}
	replies := b.deliver(birth)
	if len(replies) != 1 {
		b.t.Fatalf("host answered the NBIRTH with %d records, want one STATE", len(replies))
	}
	births, err := b.node.Accept(replies[0])
	if err != nil {
		b.t.Fatal(err)
	}
	b.deliver(births...)
	return births
}

func wireErr(record map[string]any) error {
	raw, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	_, err = domain.Parse(raw)
	return err
}

func types(records []map[string]any) []string {
	names := make([]string, len(records))
	for i, record := range records {
		names[i], _ = record["type"].(string)
	}
	return names
}

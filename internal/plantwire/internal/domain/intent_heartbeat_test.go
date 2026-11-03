package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestIntentKeyTiesTheCommandIDToItsContent(t *testing.T) {
	t.Parallel()
	command := Command{Device: "pump-1", Name: "stop", Boot: "boot-1", Args: map[string]any{"rpm": 0.0}}
	same := command.IntentKey("c-1")
	if same == "" || same != command.IntentKey("c-1") {
		t.Fatalf("the same command and id gave %q", same)
	}
	other := command
	other.Name = "start"
	if same == command.IntentKey("c-2") || same == other.IntentKey("c-1") {
		t.Fatal("another id or another content kept the same intent key")
	}
}

func TestHeartbeatIntervalNeedsAnOnlineStateThatBeats(t *testing.T) {
	t.Parallel()
	state := func(online bool, record map[string]any) Inbound {
		return Inbound{Message: Message{Type: "STATE", Online: online}, Record: record}
	}
	beating := map[string]any{"heartbeat": uint64(3), "heartbeat_interval_ms": uint64(1000)}
	if got, ok := HeartbeatInterval(state(true, beating)); !ok || got != time.Second {
		t.Fatalf("a beating host: %v, %v", got, ok)
	}
	cases := map[string]Inbound{
		"offline":        state(false, beating),
		"no heartbeat":   state(true, map[string]any{"heartbeat_interval_ms": uint64(1000)}),
		"no interval":    state(true, map[string]any{"heartbeat": uint64(3)}),
		"not a state":    {Message: Message{Type: "DATA", Online: true}, Record: beating},
		"invalid record": {Message: Message{Type: "STATE", Online: true}, Record: beating, Invalid: errors.New("torn")},
	}
	for name, in := range cases {
		if got, ok := HeartbeatInterval(in); ok || got != 0 {
			t.Fatalf("%s: %v, %v, want none", name, got, ok)
		}
	}
}

func TestBoardSourcesAreSelectedByTheBoardTypesOfTheKinds(t *testing.T) {
	t.Parallel()
	spec := &model.DomainSpec{Kinds: []model.Kind{{BoardType: "esp32"}, {}, {BoardType: "uno"}}}
	wanted := BoardTypes(spec)
	if len(wanted) != 2 || !wanted["esp32"] || !wanted["uno"] {
		t.Fatalf("board types = %v", wanted)
	}
	sources := []KindSource{
		{Path: "kinds/esp32.board.json"},
		{Path: "kinds/other.board.json", Raw: []byte(`{"board_type":"uno"}`)},
		{Path: "kinds/mega.board.json", Raw: []byte(`{"board_type":"mega"}`)},
		{Path: "kinds/uno.board.json", Raw: []byte(`{`)},
	}
	var kept []string
	for _, source := range SelectBoardSources(sources, wanted) {
		kept = append(kept, source.Path)
	}
	if len(kept) != 3 || kept[0] != "kinds/esp32.board.json" || kept[1] != "kinds/other.board.json" || kept[2] != "kinds/uno.board.json" {
		t.Fatalf("kept = %v", kept)
	}
}

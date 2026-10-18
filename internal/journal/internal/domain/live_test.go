package domain

import "testing"

func liveFixture() map[string]any {
	return map[string]any{"host":"online","gateway":"online","commands_truncated":false,
		"devices":[]any{map[string]any{"device":"P1","board":"b-1","instance":"pump-1","environment":"hardware",
			"boot_id":"boot-1","health":"online","reboots":0,"readings":[]any{map[string]any{"name":"flow","unit":"m3/h",
				"value":42.0,"quality":"good","state":"good"}}}},
		"commands":[]any{map[string]any{"command_id":"c-1","device":"P1","command":"start","boot_id":"boot-1","sent":true,
			"receipt":"accepted","reason":"","report":"succeeded","effect":"observed","effect_evidence":"Host checked flow: observed",
			"sent_at_ns":100,"updated_at_ns":120}},
		"marks":[]any{map[string]any{"at_ns":120,"kind":"command","label":"Host checked flow: observed","entity":"P1"}}}
}

func TestLiveAndConnectedMetadataAreObserverOnlyAndDoNotAlterReplayableDigests(t *testing.T) {
	t.Parallel()
	j := NewMemory(true)
	before := j.Digests()
	j.Session(0,map[string]any{"mode":"connected","plan":map[string]any{"domain":"d"}})
	j.Live(0,liveFixture())
	after := j.Digests()
	if before.Observer != after.Observer || before.Director != after.Director {
		t.Fatalf("session-only evidence changed replayable digests: %+v => %+v",before,after)
	}
	director,err := j.Bytes(StreamDirector)
	if err != nil || len(director) != 0 {
		t.Fatalf("observer projection leaked to director: %s, %v",director,err)
	}
}

package host

import "testing"

func TestTheHostBeatsWithAnIncreasingHeartbeatOnlyWhileItsSessionIsLive(t *testing.T) {
	t.Parallel()
	h := New(Options{Environment: "simulated", ProtocolMax: 1, Epoch: 3})
	if beats := h.Beat(); beats != nil {
		t.Fatalf("a host without a session beat: %v", beats)
	}
	accepted := h.Receive(birth(t, "simulated"))
	if accepted[0]["heartbeat"] != int64(0) || accepted[0]["heartbeat_interval_ms"] != int64(DefaultHeartbeatMS) {
		t.Fatalf("handshake STATE = %v, want heartbeat 0 at the declared interval", accepted[0])
	}
	for want := int64(1); want <= 3; want++ {
		beat := h.Beat()
		if len(beat) != 1 || beat[0]["type"] != "STATE" || beat[0]["online"] != true || beat[0]["heartbeat"] != want ||
			beat[0]["authority_epoch"] != int64(3) || beat[0]["seq"] != uint64(want) {
			t.Fatalf("beat %d = %v, want an online STATE with the next heartbeat and the next seq", want, beat)
		}
	}
}

func TestARefusingHostNeverBeats(t *testing.T) {
	t.Parallel()
	h := New(Options{Environment: "hardware", ProtocolMax: 1})
	h.Receive(birth(t, "simulated"))
	if beats := h.Beat(); h.Report().Refused == "" || beats != nil {
		t.Fatalf("refused %q, beats %v, want a refusal and no heartbeat", h.Report().Refused, beats)
	}
}

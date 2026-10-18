package app

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
)

func TestEveryRecordTheStandInSendsReachesTheReferenceHostWithTheSameContent(t *testing.T) {
	t.Parallel()
	c := newChain(t, relayHostOptions("hardware"))
	c.connect()
	c.reading(22.5, "")
	c.reading(23.5, "suspect")
	c.reading(0, "stale")
	if len(c.toHost) != len(c.fromGateway) || len(c.fromGateway) != 6 {
		t.Fatalf("the gateway sent %d records and the host received %d, want the same six", len(c.fromGateway), len(c.toHost))
	}
	for i := range c.fromGateway {
		var drop []string
		if c.fromGateway[i]["type"] == "NBIRTH" {
			drop = []string{"bd_seq", "prior_last_seq"}
		}
		if got, want := exactContent(t, c.toHost[i], drop...), exactContent(t, c.fromGateway[i], drop...); !reflect.DeepEqual(got, want) {
			t.Errorf("record %d reached the host as %v, want the content %v", i, got, want)
		}
		if c.toHost[i]["seq"] != uint64(i) || c.toHost[i]["type"] != c.fromGateway[i]["type"] {
			t.Errorf("record %d has seq %v type %v, want the relay's own consecutive seq %d", i, c.toHost[i]["seq"], c.toHost[i]["type"], i)
		}
	}
	if report := c.host.Report(); !c.host.Accepted() || report.Refused != "" || len(report.Quarantined) != 0 {
		t.Fatalf("host report %+v accepted %v, want the session accepted with nothing quarantined", report, c.host.Accepted())
	}
}

func TestAHostCommandReachesTheStandInUnchangedAndItsReceiptAndResultComeBack(t *testing.T) {
	t.Parallel()
	c := newChain(t, relayHostOptions("hardware", relayCommandStep()))
	c.connect()
	c.barrier(2)
	sent, received := ofType(c.fromHost, "DCMD"), ofType(c.toGateway, "DCMD")
	if len(sent) != 1 || len(received) != 1 || !reflect.DeepEqual(exactContent(t, received[0]), exactContent(t, sent[0])) {
		t.Fatalf("the host sent %d commands and the stand-in received %d, want the same one with the same content", len(sent), len(received))
	}
	command := received[0]
	want := map[string]any{"authority_epoch": json.Number("7"), "deadline_ns": json.Number("1767225780000000000"), "not_before_ns": json.Number("0"),
		"expected_boot_id": "boot-1", "fallback_on_lease_loss": "safe_off", "policy_digest": relayHostOptions("hardware").Policy, "device": "fan-01"}
	for member, value := range want {
		if command[member] != value {
			t.Errorf("%s = %v, want %v as the host signed it", member, command[member], value)
		}
	}
	status := c.host.Report().Commands["cmd-1"]
	if status == nil || !reflect.DeepEqual(status.Receipts, []string{"accepted"}) || status.Result != "succeeded" {
		t.Fatalf("host status %+v, want the receipt accepted and the result succeeded to have come back", status)
	}
	for _, kind := range []string{"DRECEIPT", "DRESULT"} {
		up, down := ofType(c.toHost, kind), ofType(c.fromGateway, kind)
		if len(up) != 1 || len(down) != 1 || !reflect.DeepEqual(exactContent(t, up[0]), exactContent(t, down[0])) {
			t.Errorf("%s: host received %v, the stand-in sent %v, want the same one", kind, up, down)
		}
	}
	if len(ofType(c.toGateway, "DCMD")) != len(ofType(c.fromHost, "DCMD")) {
		t.Fatal("the relay sent the gateway a command the host did not issue")
	}
}

func TestTheHostCommissionsTheKindDigestsTheStandInAnnounced(t *testing.T) {
	t.Parallel()
	options := relayHostOptions("hardware")
	options.Inventory = map[string]string{"zone-01": "thermal_sensor", "fan-01": "fan"}
	c := newChain(t, options)
	c.connect()
	sent, received := ofType(c.fromGateway, "NBIRTH")[0], ofType(c.toHost, "NBIRTH")[0]
	files := map[string]string{"thermal_sensor": relaySensorKind, "fan": relayFanKind}
	for _, announced := range received["kinds"].([]any) {
		kind := announced.(map[string]any)
		digest, err := domain.KindDigest(relayKindFile(t, files[kind["kind"].(string)]))
		if err != nil || kind["kind_digest"] != digest {
			t.Errorf("kind %v reached the host with digest %v, its kind file digests to %s (%v)", kind["kind"], kind["kind_digest"], digest, err)
		}
	}
	if sent["capability_digest"] != received["capability_digest"] || !c.host.Accepted() || len(c.host.Report().Quarantined) != 0 {
		t.Fatalf("capability_digest %v vs %v accepted %v quarantined %v, want the commissioning to hold", sent["capability_digest"],
			received["capability_digest"], c.host.Accepted(), c.host.Report().Quarantined)
	}
}

func TestTheBarrierIsMappedBetweenTheGatewaysSequenceAndTheRelaysOwn(t *testing.T) {
	t.Parallel()
	c := newChain(t, relayHostOptions("hardware"))
	c.connect()
	c.reading(22.5, "")
	broken := c.gateway.Emit(domain.DeviceData("zone-01", "boot-1", false, []map[string]any{domain.Sample(1, 1.0, "", 1)}))
	delete(broken[0], "metrics")
	c.feedGateway(broken[0])
	c.reading(23.5, "")
	c.barrier(5)
	gatewaySync, upSync := ofType(c.fromGateway, "NSYNC")[0], ofType(c.toHost, "NSYNC")[0]
	answer := ofType(c.toGateway, "HSYNC")
	if gatewaySync["seq"] != uint64(6) || upSync["seq"] != uint64(5) || upSync["through_seq"] != uint64(4) {
		t.Fatalf("NSYNC seq %v up as seq %v through_seq %v, want gateway 6, relay 5 and through_seq 4 (gateway 5 less the record dropped)", gatewaySync["seq"], upSync["seq"], upSync["through_seq"])
	}
	if len(answer) != 1 || answer[0]["through_seq"] != gatewaySync["seq"] {
		t.Fatalf("HSYNC down = %v, want through_seq %v, the seq the gateway gave its NSYNC", answer, gatewaySync["seq"])
	}
}

func TestARebirthTheHostAsksForReachesTheStandInAndComesBackUp(t *testing.T) {
	t.Parallel()
	c := newChain(t, relayHostOptions("hardware"))
	c.lose = func(index int) bool { return index == 3 }
	c.connect()
	c.reading(1, "")
	c.reading(2, "")
	asked, forwarded := ofType(c.fromHost, "NCMD"), ofType(c.toGateway, "NCMD")
	births := ofType(c.toHost, "NBIRTH")
	if len(asked) != 1 || len(forwarded) != 1 || !reflect.DeepEqual(exactContent(t, forwarded[0]), exactContent(t, asked[0])) {
		t.Fatalf("the host asked %d times and the stand-in was asked %d times, want the one request unchanged", len(asked), len(forwarded))
	}
	if len(births) != 2 || births[1]["bd_seq"] != uint64(1) || c.host.Report().Rebirths != 1 || len(c.host.Report().Quarantined) != 0 {
		t.Fatalf("births up %d, report %+v, want the stand-in's rebirth passed up as the relay's bd_seq 1", len(births), c.host.Report())
	}
}

func TestAGapOnTheGatewayLinkMakesTheRelayAskTheStandInForARebirthItself(t *testing.T) {
	t.Parallel()
	c := newChain(t, relayHostOptions("hardware"))
	c.connect()
	lost := c.gateway.Emit(domain.DeviceData("zone-01", "boot-1", false, []map[string]any{domain.Sample(1, 1.0, "", 1)}))
	c.reading(2, "")
	if len(lost) != 1 || len(ofType(c.toGateway, "NCMD")) != 1 || len(ofType(c.fromHost, "NCMD")) != 0 {
		t.Fatalf("requests down %v, from the host %v, want one request that the relay made as the gateway's host", ofType(c.toGateway, "NCMD"), ofType(c.fromHost, "NCMD"))
	}
	if births := ofType(c.toHost, "NBIRTH"); len(births) != 2 || len(ofType(c.toHost, "DBIRTH")) != 4 {
		t.Fatalf("NBIRTH %d DBIRTH %d up, want the rebirth carried to the host", len(births), len(ofType(c.toHost, "DBIRTH")))
	}
}

func TestAHostThatComesBackIsGivenTheSameBoardAgainThroughARebirth(t *testing.T) {
	t.Parallel()
	c := newChain(t, relayHostOptions("hardware"))
	c.connect()
	c.pump(c.relay.HostLost())
	offline := ofType(c.toGateway, "STATE")
	if last := offline[len(offline)-1]; last["online"] != false || last["authority_epoch"] != json.Number("7") {
		t.Fatalf("STATE = %v, want the host's own epoch with online false", last)
	}
	c.host = host.New(relayHostOptions("hardware"))
	c.pump(c.relay.HostAttached())
	births := ofType(c.toHost, "NBIRTH")
	if len(births) != 3 || births[1]["session_id"] != "relay-h2" || births[1]["bd_seq"] != uint64(1) || births[1]["prior_last_seq"] != int64(2) {
		t.Fatalf("NBIRTH up = %v, want a second host session relay-h2 with bd_seq 1 and prior_last_seq 2 and then the gateway's rebirth", births)
	}
	if report := c.host.Report(); !c.host.Accepted() || report.Refused != "" || len(report.Quarantined) != 0 || len(ofType(c.toHost, "DBIRTH")) != 4 {
		t.Fatalf("new host report %+v accepted %v, want the board born again on it", report, c.host.Accepted())
	}
}

func TestAGatewayThatReturnsIsBornAgainInTheSameHostSession(t *testing.T) {
	t.Parallel()
	c := newChain(t, relayHostOptions("hardware", relayCommandStep()))
	c.connect()
	c.pump(c.relay.GatewayLost())
	deaths := ofType(c.toHost, "DDEATH")
	if len(deaths) != 2 || deaths[0]["reason"] != "link_down" || deaths[1]["reason"] != "link_down" {
		t.Fatalf("DDEATH up = %v, want link_down for both devices", deaths)
	}
	c.gateway = relayGateway(t, "s-2")
	c.connect()
	c.barrier(2)
	if births := ofType(c.toHost, "NBIRTH"); len(births) != 2 || births[1]["session_id"] != "relay-h1" || len(ofType(c.toHost, "DBIRTH")) != 4 {
		t.Fatalf("NBIRTH up = %v, want the same host session reborn with its devices", births)
	}
	if status := c.host.Report().Commands["cmd-1"]; status == nil || status.Result != "succeeded" || len(c.host.Report().Quarantined) != 0 {
		t.Fatalf("report %+v, want the command to reach the returned gateway and nothing quarantined", c.host.Report())
	}
	if len(ofType(c.toGateway, "DCMD")) != len(ofType(c.fromHost, "DCMD")) {
		t.Fatal("the relay sent the gateway a command the host did not issue")
	}
}

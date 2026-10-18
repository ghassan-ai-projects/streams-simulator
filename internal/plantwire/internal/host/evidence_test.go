package host

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func evidenceHost(t *testing.T) *Host {
	t.Helper()
	h := scriptedHost(t, Scripted{ID:"check-1",AtNS:0,Device:"P1",Command:"start",Args:map[string]any{},
		Verify:&Verify{Metric:"flow",Equals:50,Tolerance:0.5,WithinNS:100}})
	h.Receive(sync(t,3,0))
	succeed(t,h,4,"check-1")
	return h
}

func evidenceFrom(t *testing.T, replies []map[string]any) map[string]any {
	t.Helper()
	for _, reply := range replies {
		if list,ok := reply["command_evidence"].([]any); ok && len(list) == 1 {
			if err := domain.Check(reply); err != nil { t.Fatalf("evidence invalid: %v",err) }
			return list[0].(map[string]any)
		}
	}
	t.Fatalf("no one-command evidence in %v",replies)
	return nil
}

func TestHostPublishesIndependentFeedbackEvidenceOnceAfterANewGoodSample(t *testing.T) {
	t.Parallel()
	h := evidenceHost(t)
	h.Receive(sample(t,6,50,false))
	e := evidenceFrom(t,h.Receive(sync(t,7,20)))
	if e["outcome"] != "observed" || e["observed_value"] != 50.0 || e["boot_id"] != "board-01-boot-1" ||
		e["command_id"] != "check-1" || e["device"] != "P1" || e["metric"] != "flow" {
		t.Fatalf("host feedback evidence = %v",e)
	}
	replies := h.Receive(sync(t,8,30))
	if _, repeated := replies[0]["command_evidence"]; repeated {
		t.Fatalf("host repeated a terminal decision: %v",replies)
	}
}

func TestHostDoesNotConfirmHistoricalStaleOrRebootedFeedback(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"historical","stale","reboot","wrong_result_boot"} {
		t.Run(bad,func(t *testing.T) {
			t.Parallel()
			h := evidenceHost(t)
			data := sample(t,6,50,bad == "historical")
			switch bad {
			case "stale":
				data = record(t,"DDATA",6,map[string]any{"metrics":[]any{map[string]any{
					"alias":1,"value":50,"quality":"stale","event_time_ns":20,"time_source":"device"}}})
			case "reboot":
				h.Receive(record(t,"DBIRTH",6,map[string]any{"device":"P1","boot_id":"boot-new"}))
				data = record(t,"DDATA",7,map[string]any{"boot_id":"boot-new","metrics":[]any{map[string]any{
					"alias":1,"value":50,"quality":"good","event_time_ns":20,"time_source":"device"}}})
			case "wrong_result_boot":
				h.report.Commands["check-1"].Result = ""
				h.Receive(record(t,"DRESULT",6,map[string]any{"command_id":"check-1","boot_id":"wrong","reported_outcome":"succeeded"}))
				data = sample(t,7,50,false)
			}
			h.Receive(data)
			replies := h.Receive(sync(t,data.Seq+1,200))
			if bad == "wrong_result_boot" {
				if _,exists := replies[len(replies)-1]["command_evidence"]; exists { t.Fatal("wrong-boot result led to evidence") }
				return
			}
			if e := evidenceFrom(t,replies); e["outcome"] != "not_observed" { t.Fatalf("%s confirmed: %v",bad,e) }
		})
	}
}

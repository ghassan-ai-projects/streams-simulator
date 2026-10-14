package component

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestTheAnnunciatorInstrumentDeclaresTheAlarmStateCodesTheWorldEmits(t *testing.T) {
	t.Parallel()
	detail, err := builtInLibrary(t).Describe("streamsim/annunciator@1")
	if err != nil {
		t.Fatal(err)
	}
	declared, _ := detail.Document["states"].([]any)
	want := []struct {
		value     int
		name      string
		attention any
	}{
		{model.AlarmNormal, "normal", nil},
		{model.AlarmUnacked, "unacknowledged", model.AttentionAlarm},
		{model.AlarmAcked, "acknowledged", model.AttentionAlarm},
		{model.AlarmReturned, "returned unacknowledged", model.AttentionWarn},
		{model.AlarmShelved, "shelved", nil},
	}
	if len(declared) != len(want) {
		t.Fatalf("states = %v, want %d", declared, len(want))
	}
	for i, state := range want {
		entry, _ := declared[i].(map[string]any)
		value, _ := entry["value"].(interface{ Int64() (int64, error) })
		number, _ := value.Int64()
		if int(number) != state.value || entry["name"] != state.name || entry["attention"] != state.attention {
			t.Errorf("state %d = %v, want value %d %q attention %v", i, entry, state.value, state.name, state.attention)
		}
	}
}

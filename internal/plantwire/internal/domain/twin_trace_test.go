package domain

import (
	"testing"
)

func TestAnnouncedDevicesAndMetricsAreSortedByName(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{
		twBirth("zone-02", "thermal_sensor", 100, twAlias{"temperature", 1}),
		twBirth("zone-01", "thermal_sensor", 100, twAlias{"temperature", 2}, twAlias{"humidity", 1}),
	})
	if got := []string{trace.Devices[0].Name, trace.Devices[1].Name}; got[0] != "zone-01" || got[1] != "zone-02" {
		t.Fatalf("devices = %v, want zone-01 then zone-02", got)
	}
	want := []BoardMetric{
		{Device: "zone-01", Kind: "thermal_sensor", Metric: "humidity"},
		{Device: "zone-01", Kind: "thermal_sensor", Metric: "temperature"},
		{Device: "zone-02", Kind: "thermal_sensor", Metric: "temperature"},
	}
	if len(trace.Metrics) != len(want) {
		t.Fatalf("metrics = %+v, want %+v", trace.Metrics, want)
	}
	for i := range want {
		if trace.Metrics[i] != want[i] {
			t.Errorf("metric %d = %+v, want %+v", i, trace.Metrics[i], want[i])
		}
	}
}

func TestARebirthReplacesTheDevicesAliases(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{
		twBirth("zone-01", "thermal_sensor", 100, twAlias{"temperature", 1}),
		twBirth("zone-01", "thermal_sensor", 150, twAlias{"humidity", 1}),
		twData("zone-01", false, twSample(1, twNumber(t, "40"), "good", 200)),
	})
	if len(trace.Devices) != 1 {
		t.Fatalf("devices = %+v, want the rebirth to replace the announced device", trace.Devices)
	}
	if len(trace.Readings) != 1 || trace.Readings[0].Metric != "humidity" {
		t.Fatalf("readings = %+v, want the alias 1 read as humidity after the rebirth", trace.Readings)
	}
}

func TestAReadingIsAttributedByDeviceAndAlias(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{
		twBirth("zone-01", "thermal_sensor", 100, twAlias{"temperature", 1}),
		twData("zone-01", false, twSample(1, twNumber(t, "34.5"), "good", 150)),
	})
	want := BoardReading{Device: "zone-01", Metric: "temperature", Quality: "good", AtNS: 150, Value: 34.5, Numeric: true}
	if len(trace.Readings) != 1 || trace.Readings[0] != want {
		t.Fatalf("readings = %+v, want exactly %+v", trace.Readings, want)
	}
}

func TestAValueThatIsNotANumberIsKeptButNotNumeric(t *testing.T) {
	t.Parallel()
	values := map[string]any{"string": "hot", "bool": true, "null": nil}
	for name, value := range values {
		trace := ReadBoardTrace([]map[string]any{
			twBirth("zone-01", "thermal_sensor", 100, twAlias{"temperature", 1}),
			twData("zone-01", false, twSample(1, value, "good", 150)),
		})
		if len(trace.Readings) != 1 || trace.Readings[0].Numeric || trace.Readings[0].Value != 0 {
			t.Errorf("%s value: readings = %+v, want one reading that is not numeric", name, trace.Readings)
		}
	}
}

func TestAHistoricalBatchIsMarkedHistorical(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{
		twBirth("zone-01", "thermal_sensor", 100, twAlias{"temperature", 1}),
		twData("zone-01", true, twSample(1, twNumber(t, "30"), "good", 90)),
	})
	if len(trace.Readings) != 1 || !trace.Readings[0].Historical {
		t.Fatalf("readings = %+v, want the batch marked historical", trace.Readings)
	}
}

func TestUnknownAliasesAndDevicesAreCountedAndCreateNoReading(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{
		twBirth("zone-01", "thermal_sensor", 100, twAlias{"temperature", 1}),
		twData("zone-01", false, twSample(9, twNumber(t, "30"), "good", 150)),
		twData("ghost", false, twSample(1, twNumber(t, "30"), "good", 150), twSample(1, twNumber(t, "31"), "good", 160)),
	})
	if trace.Unattributed != 3 || len(trace.Readings) != 0 {
		t.Fatalf("unattributed = %d, readings = %+v, want 3 unattributed samples and no reading", trace.Unattributed, trace.Readings)
	}
}

func TestStartNSIsTheEarliestTimeEvenWhenRecordsArriveOutOfOrder(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{
		twBirth("zone-01", "thermal_sensor", 100, twAlias{"temperature", 1}),
		twData("zone-01", false, twSample(1, twNumber(t, "30"), "good", 300)),
		twData("zone-01", false, twSample(1, twNumber(t, "30"), "good", 50)),
	})
	if trace.StartNS != 50 {
		t.Fatalf("StartNS = %d, want the earliest time seen, 50", trace.StartNS)
	}
	if only := ReadBoardTrace([]map[string]any{twBirth("zone-01", "thermal_sensor", 100)}); only.StartNS != 0 {
		t.Fatalf("StartNS of a birth alone = %d, want 0: a birth's time is not a reading time", only.StartNS)
	}
}

func TestAReceiptIsStampedWithTheLatestTimeSeenSoFar(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{
		twBirth("zone-01", "thermal_sensor", 100, twAlias{"temperature", 1}),
		twReceipt("c-1", "accepted", ""),
		twData("zone-01", false, twSample(1, twNumber(t, "30"), "good", 300)),
		twReceipt("c-2", "accepted", ""),
		twData("zone-01", false, twSample(1, twNumber(t, "30"), "good", 200)),
		twReceipt("c-3", "accepted", ""),
	})
	want := map[string]int64{"c-1": 0, "c-2": 300, "c-3": 300}
	for id, at := range want {
		if got := trace.Receipts[id].AtNS; got != at {
			t.Errorf("receipt %s at %d, want the latest time seen so far, %d", id, got, at)
		}
	}
}

func TestAResultBeforeItsReceiptMergesIntoTheSameEntry(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{
		twResult("c-1", "failed"),
		twReceipt("c-1", "accepted", ""),
		twResult("c-2", "succeeded"),
	})
	if got := trace.Receipts["c-1"]; got.Status != "accepted" || got.Outcome != "failed" {
		t.Errorf("c-1 = %+v, want status accepted merged with outcome failed", got)
	}
	if got := trace.Receipts["c-2"]; got.Status != "" || got.Outcome != "succeeded" {
		t.Errorf("c-2 = %+v, want a result alone with no status", got)
	}
}

func TestRecordsThatAreNotBoardRecordsAreIgnored(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{
		{"type": "DSTATE", "device": "zone-01"},
		{"type": "DBIRTH"},
	})
	if len(trace.Devices) != 0 || len(trace.Readings) != 0 || len(trace.Receipts) != 0 || trace.Unattributed != 0 {
		t.Fatalf("trace = %+v, want nothing read from records that are not board records", trace)
	}
}

func TestHasReportsOnlyAnnouncedDevices(t *testing.T) {
	t.Parallel()
	trace := ReadBoardTrace([]map[string]any{twBirth("zone-01", "thermal_sensor", 100)})
	if !trace.Has("zone-01") || trace.Has("zone-09") {
		t.Fatalf("Has = zone-01:%v zone-09:%v, want only zone-01", trace.Has("zone-01"), trace.Has("zone-09"))
	}
}

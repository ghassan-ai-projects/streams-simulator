package domain

import (
	"fmt"
	"slices"
	"testing"
)

// fakeTwin is a scripted TwinModel: it records every Apply and Predict in order.
type fakeTwin struct {
	support map[string]string
	predict map[string]float64
	refuse  map[string]error
	log     []string
}

func newFakeTwin() *fakeTwin {
	return &fakeTwin{support: map[string]string{}, predict: map[string]float64{"zone-01.temperature": 30}, refuse: map[string]error{}}
}

func (f *fakeTwin) Support(device, metric string) string { return f.support[device+"."+metric] }

func (f *fakeTwin) Apply(command TwinCommand) error {
	f.log = append(f.log, "apply "+command.ID)
	return f.refuse[command.ID]
}

func (f *fakeTwin) Predict(device, metric string, atNS int64) float64 {
	f.log = append(f.log, fmt.Sprintf("predict %s.%s@%d", device, metric, atNS))
	return f.predict[device+"."+metric]
}

func (f *fakeTwin) position(entry string) int { return slices.Index(f.log, entry) }

func twTemperature(at int64, value float64) BoardReading {
	return BoardReading{Device: "zone-01", Metric: "temperature", Quality: "good", AtNS: at, Value: value, Numeric: true}
}

func twReadings(n int, value float64) []BoardReading {
	readings := make([]BoardReading, n)
	for i := range readings {
		readings[i] = twTemperature(200+int64(i), value)
	}
	return readings
}

func twTrace(readings ...BoardReading) BoardTrace {
	return BoardTrace{Devices: []BoardDevice{{Name: "zone-01", Kind: "thermal_sensor", Board: "mega-01"}},
		Metrics: []BoardMetric{thermalTemperature}, Readings: readings, Receipts: map[string]BoardReceipt{}}
}

func twKinds(withRange bool) KindFiles {
	metric := map[string]any{"name": "temperature", "unit": "Cel"}
	if withRange {
		metric["eu_range"] = map[string]any{"min": 0.0, "max": 50.0}
	}
	return KindFiles{"thermal_sensor": {"metrics": []any{metric}}}
}

func twApplied(id string, at int64) TwinCommand {
	return TwinCommand{ID: id, Device: "fan-01", Name: "set_pwm_lease", AtNS: at, Status: CommandApplied}
}

func twInput(trace BoardTrace, commands ...TwinCommand) TwinInput {
	return TwinInput{Trace: trace, Commands: commands, Kinds: twKinds(true),
		Run: TwinEvidence{FieldTraceDigest: "sha256:aa", JournalDigest: "sha256:bb"}, Twin: TwinInfo{Domain: "arduino-thermal"}}
}

func twEvaluate(t *testing.T, input TwinInput, twin *fakeTwin) TwinReport {
	t.Helper()
	report, err := EvaluateTwin(input, twin)
	if err != nil {
		t.Fatalf("EvaluateTwin: %v", err)
	}
	return report
}

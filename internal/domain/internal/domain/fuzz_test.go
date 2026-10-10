package domain

// Fuzz targets: the parsers must never panic on adversarial input. The
// shipped domains seed the corpus; the fuzzer explores everything else.

import (
	"testing"
)

func FuzzDomainParse(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(exampleJSON),
		[]byte(`{"id":"x","version":"0.1.0","stresses":"` + repeat("z", 40) + `"}`),
		[]byte(`{}`),
		[]byte(`null`),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		// Parse must never panic; a rejection is a valid outcome.
		_, _ = Parse(raw, "fuzz")
	})
}

const exampleJSON = `{"id":"aquaculture-pond","version":"0.1.0","title":"fuzz","stresses":"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz","axes":{"rate":"low","cardinality":"singleton","value_shape":["scalar"],"cadence":["periodic"],"lateness":"none","absence":"signal","time_reference":"wall","correlation":["independent"],"seasonality":["none"],"actuation":"observe_only","consequence":["cost"],"fidelity":["F0"]},"entities":{"id_template":"e-{n}","count":{"default":1}},"state":[{"name":"x","initial":0}],"channels":[{"name":"ch","value_type":"number","unit":"u","resolution":0.1,"observes":"x","fidelity":"F0","absence":"signal","cadence":{"mode":"periodic","period_s":60},"noise":{"model":"none","sigma":0}}],"faults":[{"id":"f1","onset":{"shape":"step"},"affects":[{"state":"x","delta":1}],"observability":{"detector":{"form":"single_channel_snr","channel":"ch"}}}],"profiles":[{"name":"nominal","description":"x"},{"name":"correlated_cascade","description":"x","not_applicable":"yyyyyyyyyyyyyyyyyyyy"},{"name":"sensor_pathology","description":"x"}],"ground_truth":{"negative_class_fraction":0.4}}`

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

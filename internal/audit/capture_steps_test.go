package audit

import "testing"

func TestAuditCaptureExcludesDroppedRecords(t *testing.T) {
	spec := loadSpec(t)
	panel := NewPanel(spec, 42, 60*1e9)
	grid, log, err := panel.build("site-a/pond-1", 0, ids(spec), nil, nil, []Perturbation{{Name: "drop", Params: map[string]any{"rate": 1.0}}}, 120*1e9)
	if err != nil {
		t.Fatal(err)
	}
	if len(log) != 0 {
		t.Fatalf("dropped records reached audit log: %+v", log)
	}
	for channel, values := range grid {
		for _, v := range values {
			if v != 0 {
				t.Fatalf("dropped record reached grid %s: %v", channel, values)
			}
		}
	}
}

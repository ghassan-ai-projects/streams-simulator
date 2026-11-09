package model

import (
	"encoding/json"
	"testing"
)

func TestADenominatorInputSurvivesARoundTripWithItsCoefficient(t *testing.T) {
	t.Parallel()
	var inputs []F1Input
	if err := json.Unmarshal([]byte(`["demand", {"state": "wsum", "denominator": true}]`), &inputs); err != nil {
		t.Fatal(err)
	}
	if inputs[0].Denominator || !inputs[1].Denominator || inputs[1].Coef != 1 {
		t.Fatalf("inputs = %+v, want the second one dividing with coefficient 1", inputs)
	}
	raw, err := json.Marshal(inputs)
	if err != nil || string(raw) != `[{"state":"demand","coef":1},{"state":"wsum","coef":1,"denominator":true}]` {
		t.Fatalf("marshal = %s, %v", raw, err)
	}
}

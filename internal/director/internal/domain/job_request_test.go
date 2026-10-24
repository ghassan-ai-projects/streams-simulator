package domain

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func validJob() JobRequest {
	return JobRequest{Domain: "aquaculture-pond", Seed: 7, DurationS: 3600, Rate: 120,
		Faults: []JobFault{{EntityID: "pond-1", Fault: "aerator_failure", AtS: 600}},
		Calls:  []JobCall{{EntityID: "pond-1", Effector: "start_aerator", Args: map[string]any{"pond_id": "pond-1"}, AtS: 1200}}}
}

func TestACompleteJobRequestIsAccepted(t *testing.T) {
	t.Parallel()
	if err := CheckJob(validJob()); err != nil {
		t.Fatalf("a valid request was refused: %v", err)
	}
	edge := validJob()
	edge.Faults[0].AtS, edge.Calls[0].AtS, edge.Rate = 0, edge.DurationS, 0
	if err := CheckJob(edge); err != nil {
		t.Fatalf("times at 0 and at the duration are inside the run: %v", err)
	}
}

func TestAJobRequestThatCannotRunIsRefusedByName(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		change func(*JobRequest)
		want   string
	}{
		"no domain":             {func(r *JobRequest) { r.Domain = "" }, "domain"},
		"negative seed":         {func(r *JobRequest) { r.Seed = -1 }, "seed"},
		"zero duration":         {func(r *JobRequest) { r.DurationS = 0 }, "duration_s"},
		"negative duration":     {func(r *JobRequest) { r.DurationS = -5 }, "duration_s"},
		"absurd duration":       {func(r *JobRequest) { r.DurationS = 1 << 40 }, "duration_s"},
		"negative rate":         {func(r *JobRequest) { r.Rate = -1 }, "rate"},
		"not a number rate":     {func(r *JobRequest) { r.Rate = math.NaN() }, "rate"},
		"infinite rate":         {func(r *JobRequest) { r.Rate = math.Inf(1) }, "rate"},
		"fault without entity":  {func(r *JobRequest) { r.Faults[0].EntityID = "" }, "faults[0].entity_id"},
		"fault without name":    {func(r *JobRequest) { r.Faults[0].Fault = "" }, "faults[0].fault"},
		"fault before start":    {func(r *JobRequest) { r.Faults[0].AtS = -1 }, "faults[0].at_s"},
		"fault after the end":   {func(r *JobRequest) { r.Faults[0].AtS = r.DurationS + 1 }, "faults[0].at_s"},
		"call without entity":   {func(r *JobRequest) { r.Calls[0].EntityID = "" }, "calls[0].entity_id"},
		"call without effector": {func(r *JobRequest) { r.Calls[0].Effector = "" }, "calls[0].effector"},
		"call before start":     {func(r *JobRequest) { r.Calls[0].AtS = -1 }, "calls[0].at_s"},
		"call after the end":    {func(r *JobRequest) { r.Calls[0].AtS = r.DurationS + 1 }, "calls[0].at_s"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := validJob()
			tc.change(&request)
			err := CheckJob(request)
			var toolErr *ToolError
			ok := errors.As(err, &toolErr)
			if !ok || toolErr.Code != CodeInvalidArgs || !strings.Contains(toolErr.Msg, tc.want) {
				t.Fatalf("err = %v, want an invalid_args refusal naming %q", err, tc.want)
			}
		})
	}
}

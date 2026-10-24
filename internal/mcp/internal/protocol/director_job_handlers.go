package protocol

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
)

func (d directorTools) handleJobStart(_ context.Context, args map[string]any) (any, error) {
	var request director.JobRequest
	if err := decodeToolRecord(args, &request); err != nil {
		return nil, err
	}
	return d.StartJob(request)
}

func (d directorTools) handleJobStatus(_ context.Context, args map[string]any) (any, error) {
	return d.JobStatus(stringArg(args, "world_id"))
}

func (d directorTools) handleJobStop(_ context.Context, args map[string]any) (any, error) {
	return d.StopJob(stringArg(args, "world_id"))
}

func (d directorTools) handleJobList(_ context.Context, _ map[string]any) (any, error) {
	return map[string]any{"jobs": d.ListJobs()}, nil
}

func (d directorTools) handleJobPause(_ context.Context, args map[string]any) (any, error) {
	return d.PauseJob(stringArg(args, "world_id"))
}

func (d directorTools) handleJobResume(_ context.Context, args map[string]any) (any, error) {
	return d.ResumeJob(stringArg(args, "world_id"))
}

func (d directorTools) handleJobStep(_ context.Context, args map[string]any) (any, error) {
	return d.StepJob(stringArg(args, "world_id"))
}

func (d directorTools) handleJobRate(_ context.Context, args map[string]any) (any, error) {
	var request struct {
		WorldID string  `json:"world_id"`
		Rate    float64 `json:"rate"`
	}
	if err := decodeToolRecord(args, &request); err != nil {
		return nil, err
	}
	return d.SetJobRate(request.WorldID, request.Rate)
}

func (d directorTools) handleJobInjectFault(_ context.Context, args map[string]any) (any, error) {
	return d.JobInjectFault(stringArg(args, "world_id"), stringArg(args, "entity_id"), stringArg(args, "fault"))
}

func (d directorTools) handleJobInvokeEffector(_ context.Context, args map[string]any) (any, error) {
	callArgs, _ := args["args"].(map[string]any)
	return d.JobInvokeEffector(stringArg(args, "world_id"), stringArg(args, "entity_id"), stringArg(args, "effector"), callArgs)
}

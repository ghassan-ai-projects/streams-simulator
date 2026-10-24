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

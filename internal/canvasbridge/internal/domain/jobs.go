package domain

import "encoding/json"

type JobView struct {
	WorldID string `json:"world_id"`
	State   string `json:"state"`
	Name    string `json:"name,omitempty"`
	Domain  string `json:"domain,omitempty"`
	Explore bool   `json:"explore,omitempty"`
}

type jobResult struct {
	JobView
	Jobs []JobView `json:"jobs"`
}

func (j JobView) Active() bool {
	switch j.State {
	case "completed", "stopped", "failed":
		return false
	}
	return true
}

func JobsOf(structured json.RawMessage) []JobView {
	var result jobResult
	if json.Unmarshal(structured, &result) != nil {
		return nil
	}
	jobs := result.Jobs
	if result.WorldID != "" {
		jobs = append(jobs, result.JobView)
	}
	return keepNamed(jobs)
}

func keepNamed(jobs []JobView) []JobView {
	named := make([]JobView, 0, len(jobs))
	for _, job := range jobs {
		if job.WorldID != "" && job.State != "" {
			named = append(named, job)
		}
	}
	return named
}

package domain

const (
	ReplayReproduces = "reproduces"
	ReplayDiverges   = "does_not_reproduce"
	ReplayIncomplete = "incomplete"
)

type Replay struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func WithReplay(report *Report, replay Replay) *Report {
	report.Replay = replay
	report.OpenQuestions = questionsOf(report)
	return report
}

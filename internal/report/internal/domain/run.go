package domain

import (
	"encoding/json"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type opened struct {
	RunID       string            `json:"run_id"`
	SimVersion  string            `json:"sim_version"`
	Domain      model.ArtifactRef `json:"domain"`
	Adapter     model.ArtifactRef `json:"adapter"`
	StartTimeNS int64             `json:"start_time_ns"`
	EntityIDs   []string          `json:"entity_ids"`
	Mode        string            `json:"mode"`
}

func runOf(art *model.RunArtifact, observer []journal.Record) Run {
	if art == nil {
		return openRun(observer)
	}
	return endedRun(art)
}

func endedRun(art *model.RunArtifact) Run {
	mode := modeOf(art.Mode)
	run := Run{RunID: art.RunID, Mode: mode, Ended: true, Unblinded: art.Unblinded || mode == journal.ModeExplore,
		SimVersion: art.SimVersion, Domain: &art.Domain, Adapter: &art.Adapter, Seed: &art.Seed, Sink: art.Sink,
		TimeMode: art.TimeMode, StartTimeNS: art.WorldConfig.StartTimeNS, EntityCount: len(art.WorldConfig.EntityIDs)}
	run.EndTimeNS = lastAdvance(art.CommandLog)
	run.Complete, run.IncompleteReason = completion(art)
	return run
}

func completion(art *model.RunArtifact) (bool, string) {
	switch {
	case !art.Incomplete:
		return true, ""
	case art.Error != "":
		return false, art.Error
	}
	return false, "the run ended incomplete"
}

func openRun(observer []journal.Record) Run {
	var first opened
	if len(observer) > 0 && observer[0].Type == journal.TypeRunOpened {
		decodeFields(observer[0].Fields, &first)
	}
	mode := modeOf(first.Mode)
	return Run{RunID: first.RunID, Mode: mode, Unblinded: mode == journal.ModeExplore,
		IncompleteReason: "run.json is absent: the run is still open or did not end",
		SimVersion:       first.SimVersion, Domain: refOf(first.Domain), Adapter: refOf(first.Adapter),
		StartTimeNS: first.StartTimeNS, EndTimeNS: lastObserved(observer), EntityCount: len(first.EntityIDs)}
}

func modeOf(mode string) string {
	if mode == journal.ModeExplore {
		return journal.ModeExplore
	}
	return journal.ModeEvaluate
}

func refOf(ref model.ArtifactRef) *model.ArtifactRef {
	if ref == (model.ArtifactRef{}) {
		return nil
	}
	return &ref
}

func lastAdvance(commands []model.Command) int64 {
	var last int64
	for _, command := range commands {
		if command.Op == model.OpClockAdvance {
			last = integerOf(command.Args["to_ns"])
		}
	}
	return last
}

func lastObserved(observer []journal.Record) int64 {
	var last int64
	for _, record := range observer {
		if record.Type == journal.TypeClockAdvanced {
			last = integerOf(record.Fields["to_ns"])
		}
	}
	return last
}

func decodeFields(fields map[string]any, dst any) {
	raw, err := json.Marshal(fields)
	if err == nil {
		_ = json.Unmarshal(raw, dst)
	}
}

func integerOf(value any) int64 {
	number, ok := value.(json.Number)
	if !ok {
		return 0
	}
	n, _ := number.Int64()
	return n
}

func directorSeed(director []journal.Record) *uint64 {
	for _, record := range director {
		if record.Type != journal.TypeRunOpened {
			continue
		}
		number, _ := record.Fields["seed"].(json.Number)
		if seed, err := strconv.ParseUint(string(number), 10, 64); err == nil {
			return &seed
		}
	}
	return nil
}

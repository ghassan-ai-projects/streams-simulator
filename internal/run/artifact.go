package run

import (
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"os"
	"sort"
	"time"
)

// artifact assembles the run artifact from the run state.
func (r *Run) artifact() *model.RunArtifact {
	cmdLog := make([]model.Command, len(r.commandLog))
	copy(cmdLog, r.commandLog)
	sort.SliceStable(cmdLog, func(i, j int) bool { return cmdLog[i].Seq < cmdLog[j].Seq })
	counts := model.Counts{
		Emitted:         r.World.EmittedCount(),
		Perturbed:       int64(countLedger(r.ledger, model.DeliveryDuplicated, model.DeliveryDroppedByPerturb, model.DeliveryMangled, model.DeliveryDelayed, model.DeliveryRewritten, model.DeliveryReordered, model.DeliveryOmitted)),
		DroppedByDesign: int64(countLedger(r.ledger, model.DeliveryDroppedByPerturb)),
		EffectorCalls:   int64(len(r.World.EffectorCalls())),
		FaultsInjected:  int64(r.World.ActiveFaultsCount()),
	}
	return &model.RunArtifact{
		SchemaVersion: "0.1",
		SimVersion:    model.SimVersion,
		RunID:         r.ID,
		Label:         r.Config.Label,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		Domain: model.ArtifactRef{
			ID: r.Config.Domain.Spec.ID, Version: r.Config.Domain.Spec.Version, Digest: r.Config.Domain.Digest,
		},
		DomainSpec: json.RawMessage(append([]byte(nil), r.Config.Domain.Raw...)),
		Seed:       r.Config.Seed,
		Sink:       r.Config.SinkName,
		Adapter: model.ArtifactRef{
			ID: r.Config.Adapter.ID, Version: r.Config.Adapter.Version, Digest: adapterDigest(r.Config.Adapter),
		},
		AdapterSpec: json.RawMessage(append([]byte(nil), r.Config.Adapter.Raw...)),
		TimeMode:    r.Config.TimeMode,
		WorldConfig: model.WorldConfig{
			StartTimeNS:     r.worldStartTimeNS,
			EntityIDs:       r.World.InitialEntityIDs(),
			ScenarioProfile: r.Config.ScenarioProfile,
			ClockMultiplier: r.Config.ClockMultiplier,
		},
		WorldDigest:          worldDigest(r),
		CommandLog:           cmdLog,
		ExpectedTraceDigest:  r.traceDigest,
		Reproducible:         r.reproducible,
		Incomplete:           r.incomplete,
		Error:                errorString(r.runErr),
		Unblinded:            r.unblinded,
		UnblindedAt:          r.unblindedAt,
		Platform:             model.CurrentPlatform(),
		Counts:               counts,
		AppliedPerturbations: r.AppliedPerturbations(),
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func countLedger(ledger []model.LedgerRecord, reasons ...string) int {
	n := 0
	for _, l := range ledger {
		for _, r := range reasons {
			if l.DeliveryReason == r {
				n++
				break
			}
		}
	}
	return n
}

func writeJSONL(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	defer func() { _ = f.Close() }()
	switch x := v.(type) {
	case []model.LedgerRecord:
		for _, rec := range x {
			raw, _ := json.Marshal(rec)
			if _, err := f.Write(append(raw, '\n')); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	case []stateSnapshot:
		for _, rec := range x {
			raw, _ := json.Marshal(rec)
			if _, err := f.Write(append(raw, '\n')); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	}
	return nil
}

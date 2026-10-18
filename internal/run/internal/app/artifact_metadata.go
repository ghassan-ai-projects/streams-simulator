package app

import (
	"encoding/json"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/clock"
)

func (r *Run) sortedCommandLog() []model.Command {
	commands := make([]model.Command, len(r.commandLog))
	copy(commands, r.commandLog)
	sort.SliceStable(commands, func(i, j int) bool { return commands[i].Seq < commands[j].Seq })
	return commands
}

func (r *Run) artifactCounts() model.Counts {
	return model.Counts{
		Emitted:         r.World.EmittedCount(),
		Perturbed:       int64(rules.CountLedger(r.ledger, model.DeliveryDuplicated, model.DeliveryDroppedByPerturb, model.DeliveryMangled, model.DeliveryDelayed, model.DeliveryRewritten, model.DeliveryReordered, model.DeliveryOmitted)),
		DroppedByDesign: int64(rules.CountLedger(r.ledger, model.DeliveryDroppedByPerturb)),
		EffectorCalls:   int64(len(r.World.EffectorCalls())),
		FaultsInjected:  r.countCommands(model.OpFaultInject),
	}
}

func (r *Run) countCommands(op string) int64 {
	var n int64
	for _, command := range r.commandLog {
		if command.Op == op {
			n++
		}
	}
	return n
}

func (r *Run) artifactIdentity() *model.RunArtifact {
	return &model.RunArtifact{
		SchemaVersion: "0.1",
		SimVersion:    model.SimVersion,
		RunID:         r.ID,
		Label:         r.Config.Label,
		CreatedAt:     clock.Stamp(),
	}
}

func (r *Run) attachArtifactInputs(artifact *model.RunArtifact) {
	artifact.Domain = model.ArtifactRef{ID: r.Config.Domain.Spec.ID, Version: r.Config.Domain.Spec.Version, Digest: r.Config.Domain.Digest}
	artifact.DomainSpec = json.RawMessage(append([]byte(nil), r.Config.Domain.Raw...))
	artifact.Seed = r.Config.Seed
	artifact.Sink = r.Config.SinkName
	artifact.Adapter = model.ArtifactRef{ID: r.Config.Adapter.ID, Version: r.Config.Adapter.Version, Digest: rules.AdapterDigest(r.Config.Adapter)}
	artifact.AdapterSpec = json.RawMessage(append([]byte(nil), r.Config.Adapter.Raw...))
	artifact.TimeMode = r.Config.TimeMode
	artifact.WorldConfig = r.artifactWorldConfig()
}

func (r *Run) artifactWorldConfig() model.WorldConfig {
	return model.WorldConfig{
		StartTimeNS:     r.worldStartTimeNS,
		EntityIDs:       r.World.InitialEntityIDs(),
		ScenarioProfile: r.Config.ScenarioProfile,
		ClockMultiplier: r.Config.ClockMultiplier,

		Noiseless:        r.Config.Noiseless,
		ForceFailureMode: string(r.Config.ForceFailureMode),
	}
}

func (r *Run) attachArtifactState(artifact *model.RunArtifact, commands []model.Command, counts model.Counts) {
	artifact.WorldDigest = worldDigest(r)
	artifact.CommandLog = commands
	artifact.ExpectedTraceDigest = r.traceDigest
	artifact.Reproducible = r.reproducible
	artifact.Incomplete = r.incomplete
	artifact.Error = r.failureText()
	artifact.Unblinded = r.unblinded
	artifact.UnblindedAt = r.unblindedAt
	artifact.Platform = model.CurrentPlatform()
	artifact.Counts = counts
	artifact.AppliedPerturbations = r.appliedPerturbations()
	r.attachArtifactJournal(artifact)
}

func (r *Run) attachArtifactJournal(artifact *model.RunArtifact) {
	artifact.Mode = r.mode()
	artifact.JournalDigest = r.journalDigests.Observer
	artifact.SealedJournalDigest = r.journalDigests.Director
	artifact.JournalFileDigest, artifact.SealedFileDigest = r.journalDigests.ObserverFile, r.journalDigests.DirectorFile
	artifact.LedgerDigest, artifact.HistoryDigest = r.evidenceDigests()
	artifact.FieldTraceDigest, artifact.InterruptLogDigest = r.fieldTraceDigest, r.interruptLogDigest
}

func (r *Run) evidenceDigests() (string, string) {
	return canonical.DigestBytes(durable.JSONLines(r.ledger)), canonical.DigestBytes(durable.JSONLines(r.history))
}

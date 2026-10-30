package model

const (
	FileRunArtifact = "run.json"
	FileTrace       = "trace.jsonl"
	FileLedger      = "ledger.jsonl"
	FileHistory     = "world_state_history.jsonl"
	FileFieldTrace  = "field-trace.jsonl"

	FileInterruptLog         = "interrupts.jsonl"
	FileInterruptLogRejected = "interrupts.rejected.jsonl"

	FileObserverJournal = "journal.jsonl"
	FileDirectorJournal = "sealed/journal.jsonl"

	FileVerdict = "verdict.json"
	FileLabel   = "label.json"
	FileLabels  = "labels.json"
)

const (
	ModeEvaluate = "evaluate"
	ModeExplore  = "explore"
)

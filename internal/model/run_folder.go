package model

const (
	FileRunArtifact = "run.json"
	FileTrace       = "trace.jsonl"
	FileLedger      = "ledger.jsonl"
	FileHistory     = "world_state_history.jsonl"

	FileObserverJournal = "journal.jsonl"
	FileDirectorJournal = "sealed/journal.jsonl"

	FileVerdict = "verdict.json"
	FileLabel   = "label.json"
)

const (
	ModeEvaluate = "evaluate"
	ModeExplore  = "explore"
)

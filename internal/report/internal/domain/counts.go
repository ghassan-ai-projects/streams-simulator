package domain

func countsOf(folder Folder) Counts {
	var counts Counts
	if folder.Artifact != nil {
		emitted := folder.Artifact.Counts.Emitted
		counts.Emitted = &emitted
	}
	if folder.Observer.Present {
		observed := folder.Observer.scan(ObserverJournal).Delivered
		counts.Observed = &observed
	}
	if folder.Ledger.Present {
		countLedger(&counts, folder.Ledger.scan(LedgerRows).Ledger)
	}
	return counts
}

func countLedger(counts *Counts, tally Tally) {
	total, delivered, dropped := tally.Rows, tally.Delivered, tally.Dropped
	counts.LedgerRows, counts.Delivered, counts.Dropped = &total, &delivered, &dropped
	counts.DeliveredByReason, counts.DroppedByReason = tally.DeliveredByReason, tally.DroppedByReason
}

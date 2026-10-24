package domain

import (
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func runLines(run Run) []string {
	return fieldTable(append(identityFields(run), inputFields(run)...))
}

func identityFields(run Run) []field {
	return []field{
		{"Run ID", code(run.RunID)},
		{"Mode", run.Mode},
		{"Complete", yesNo(run.Complete)},
		{"Incomplete reason", run.IncompleteReason},
		{"Ended", yesNo(run.Ended)},
		{"Unblinded", yesNo(run.Unblinded)},
	}
}

func inputFields(run Run) []field {
	return append(sourceFields(run), spanFields(run)...)
}

func sourceFields(run Run) []field {
	return []field{
		{"Simulator version", run.SimVersion},
		{"Domain", reference(run.Domain)},
		{"Adapter", reference(run.Adapter)},
		{"Seed", seedText(run.Seed)},
		{"Sink", run.Sink},
		{"Time mode", run.TimeMode},
	}
}

func spanFields(run Run) []field {
	return []field{
		{"Start", instant(run.StartTimeNS, 0)},
		{"End", instant(run.EndTimeNS, run.StartTimeNS)},
		{"Requested end", instant(run.RequestedEndNS, run.StartTimeNS)},
		{"Pace", rateText(run.Rate)},
		{"Termination", terminationText(run)},
		{"Entities", countText(run.EntityCount)},
	}
}

func rateText(rate float64) string {
	if rate <= 0 {
		return ""
	}
	return strconv.FormatFloat(rate, 'f', -1, 64) + " simulated s per wall s"
}

func seedText(seed *uint64) string {
	if seed == nil {
		return ""
	}
	return strconv.FormatUint(*seed, 10)
}

func countText(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

func digestLines(digests *Digests) []string {
	if digests == nil {
		return nil
	}
	return fieldTable([]field{
		{"World", code(digests.World)},
		{"Trace", code(digests.Trace)},
		{"Observer journal", code(digests.Journal)},
		{"Director journal", code(digests.SealedJournal)},
	})
}

func verificationLines(verification Verification) []string {
	lines := []string{"Status: " + verification.Status}
	if len(verification.Checks) == 0 {
		return lines
	}
	rows := make([][]string, len(verification.Checks))
	for i, c := range verification.Checks {
		rows[i] = []string{c.Name, c.Status, code(c.Recorded), code(c.Computed)}
	}
	return append(append(lines, ""), table([]string{"Check", "Status", "Recorded", "Computed"}, rows)...)
}

func countLines(counts Counts) []string {
	lines := fieldTable([]field{
		{"Emitted", number(counts.Emitted)},
		{"Observed by the consumer", number(counts.Observed)},
		{"Ledger rows", number(counts.LedgerRows)},
		{"Delivered", number(counts.Delivered)},
		{"Dropped", number(counts.Dropped)},
	})
	lines = reasonTable(lines, "Delivered by reason", counts.DeliveredByReason)
	return reasonTable(lines, "Dropped by reason", counts.DroppedByReason)
}

func reasonTable(lines []string, title string, counts map[string]int64) []string {
	if len(counts) == 0 {
		return lines
	}
	return append(append(lines, "", "### "+title, ""), countTable("Reason", counts)...)
}

func terminationText(run Run) string {
	if run.Termination == nil {
		return ""
	}
	label := "Completed"
	if run.Termination.Reason == model.TerminationStoppedByOperator {
		label = "Stopped by operator"
	}
	return label + " at " + instant(run.Termination.AtNS, run.StartTimeNS)
}

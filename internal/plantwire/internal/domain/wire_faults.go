package domain

const (
	FaultDrop      = "drop"
	FaultDuplicate = "duplicate"
	FaultSwap      = "swap"
	FaultGarble    = "garble"
)

type WireFault struct {
	Seq  uint64
	Kind string
}

func ApplyWireFaults(records []map[string]any, faults []WireFault) []map[string]any {
	if len(faults) == 0 {
		return records
	}
	out := make([]map[string]any, 0, len(records))
	for i := 0; i < len(records); i++ {
		out, i = applyFault(out, records, i, faultAt(faults, Seq(records[i]["seq"])))
	}
	return out
}

func applyFault(out, records []map[string]any, i int, kind string) ([]map[string]any, int) {
	switch kind {
	case FaultDrop:
		return out, i
	case FaultDuplicate:
		return append(out, records[i], records[i]), i
	case FaultSwap:
		return swapped(out, records, i)
	case FaultGarble:
		return append(out, map[string]any{"seq": records[i]["seq"], "type": records[i]["type"]}), i
	}
	return append(out, records[i]), i
}

func swapped(out, records []map[string]any, i int) ([]map[string]any, int) {
	if i+1 >= len(records) {
		return append(out, records[i]), i
	}
	return append(out, records[i+1], records[i]), i + 1
}

func faultAt(faults []WireFault, seq uint64) string {
	for _, fault := range faults {
		if fault.Seq == seq {
			return fault.Kind
		}
	}
	return ""
}

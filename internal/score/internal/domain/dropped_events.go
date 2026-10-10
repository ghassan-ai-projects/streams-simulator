package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

// Each dropped record requires its own detection; retain greedy ledger-order
// matching so one detection cannot satisfy multiple drops.
func detectsEachDrop(detections []model.Detection, ledger []model.LedgerRecord) bool {
	times := detectionTimesByEntity(detections)
	used := map[string]map[int]bool{}
	detected := true
	for _, row := range ledger {
		if row.DeliveryReason == model.DeliveryDroppedByPerturb && !row.Delivered {
			if !matchEntityDrop(row, times, used) {
				detected = false
			}
		}
	}
	return detected
}

func detectionTimesByEntity(detections []model.Detection) map[string][]int64 {
	times := map[string][]int64{}
	for _, detection := range detections {
		if at, err := model.ParseTime(detection.DetectedAt); err == nil {
			times[detection.EntityID] = append(times[detection.EntityID], at)
		}
	}
	return times
}

func matchDropDetection(row model.LedgerRecord, times []int64, used map[int]bool) bool {
	for i, at := range times {
		if !used[i] && at >= row.ObservedTimeNS-30*60*1e9 && at <= row.ObservedTimeNS+30*60*1e9 {
			used[i] = true
			return true
		}
	}
	return false
}

func matchEntityDrop(row model.LedgerRecord, times map[string][]int64, used map[string]map[int]bool) bool {
	if used[row.EntityID] == nil {
		used[row.EntityID] = map[int]bool{}
	}
	return matchDropDetection(row, times[row.EntityID], used[row.EntityID])
}

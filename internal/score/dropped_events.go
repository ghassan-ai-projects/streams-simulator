package score

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

// Each dropped record requires its own detection; retain greedy ledger-order
// matching so one detection cannot satisfy multiple drops.
func detectsEachDrop(detections []model.Detection, ledger []model.LedgerRecord) bool {
	var detected bool
	drops := 0
	detectionTimes := map[string][]int64{}
	for _, d := range detections {
		if t, err := model.ParseTime(d.DetectedAt); err == nil {
			detectionTimes[d.EntityID] = append(detectionTimes[d.EntityID], t)
		}
	}
	for _, l := range ledger {
		if l.DeliveryReason == model.DeliveryDroppedByPerturb && !l.Delivered {
			drops++
		}
	}
	detected = drops == 0
	usedDetections := map[string]map[int]bool{}
	if drops > 0 {
		detected = true
		for _, l := range ledger {
			if l.DeliveryReason != model.DeliveryDroppedByPerturb || l.Delivered {
				continue
			}
			if usedDetections[l.EntityID] == nil {
				usedDetections[l.EntityID] = map[int]bool{}
			}
			matched := false
			for i, t := range detectionTimes[l.EntityID] {
				if !usedDetections[l.EntityID][i] && t >= l.ObservedTimeNS-30*60*1e9 && t <= l.ObservedTimeNS+30*60*1e9 {
					usedDetections[l.EntityID][i] = true
					matched = true
					break
				}
			}
			if !matched {
				detected = false
			}
		}
	}
	return detected
}

package domain

import "time"

func (l *Live) Tick(milliseconds int64) bool {
	l.elapsedMS += milliseconds
	changed := false
	for _, d := range l.sortedDevices() {
		changed = l.ageDevice(d) || changed
	}
	return changed
}

func (l *Live) ageDevice(d *LiveDevice) bool {
	changed := false
	for i := range d.Readings {
		r := &d.Readings[i]
		if r.State == "good" && l.elapsedMS-d.updated[r.Name] >= l.staleAfter(d, r.Name) {
			r.State = "stale"
			changed = true
		}
	}
	return changed
}

func (l *Live) staleAfter(d *LiveDevice, name string) int64 {
	scan := text(l.kinds[d.kind][name], "scan_class")
	interval, err := time.ParseDuration(scan)
	if err == nil && interval >= time.Millisecond && interval <= time.Hour {
		return 3 * interval.Milliseconds()
	}
	return gatewayIdleMS
}

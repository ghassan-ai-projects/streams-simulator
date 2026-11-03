package domain

import "time"

func HeartbeatInterval(in Inbound) (time.Duration, bool) {
	if in.Invalid != nil || in.Type != "STATE" || !in.Online || Seq(in.Record["heartbeat"]) == 0 {
		return 0, false
	}
	interval := time.Duration(saturatingInt64(Seq(in.Record["heartbeat_interval_ms"]))) * time.Millisecond
	return interval, interval > 0
}

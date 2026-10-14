package process

import "time"

func Ticker(every time.Duration) (<-chan time.Time, func()) {
	ticker := time.NewTicker(every)
	return ticker.C, ticker.Stop
}

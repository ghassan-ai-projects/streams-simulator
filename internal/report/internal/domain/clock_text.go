package domain

import "fmt"

const nsPerSecond = int64(1_000_000_000)

func clockText(spanNS int64) string {
	whole := max(0, (spanNS+nsPerSecond/2)/nsPerSecond)
	return fmt.Sprintf("%02d:%02d:%02d", whole/3600, whole%3600/60, whole%60)
}

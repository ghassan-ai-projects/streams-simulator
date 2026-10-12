//go:build !simdet

package wall

import "time"

func Now() time.Time {
	return time.Now()
}

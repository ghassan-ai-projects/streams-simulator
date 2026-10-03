package perturb

// Helpers for the perturbation layer: the reorder window, time-string
// rewrites, and parameter access.

import (
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// reorderWindow buffers records and emits them out of order within a bounded
// displacement window.
type reorderWindow struct {
	maxDisplacement int
	buf             []Delivered
}

func newReorderWindow(disp int) *reorderWindow {
	if disp < 1 {
		disp = 1
	}
	return &reorderWindow{maxDisplacement: disp}
}

func (w *reorderWindow) push(recs []Delivered, atNS int64) []Delivered {
	w.buf = append(w.buf, recs...)
	var out []Delivered
	for len(w.buf) > w.maxDisplacement {
		first, second, paired := w.popDisplaced(atNS)
		if paired {
			out = append(out, first, second)
		} else {
			out = append(out, first)
		}
	}
	return out
}

func (w *reorderWindow) flush(atNS int64) []Delivered {
	out := w.buf
	w.buf = nil
	return out
}

// addSeconds re-renders an RFC 3339 timestamp shifted by a delay.
func addSeconds(ts string, secs float64) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	return model.FormatTime(t.UnixNano() + int64(secs*1e9))
}

// alternateEncoding re-renders a timestamp with a different but equivalent
// RFC 3339 encoding ("Z" vs "+00:00"), testing identical normalization.
func alternateEncoding(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	if strings.HasSuffix(ts, "Z") {
		return t.Format("2006-01-02T15:04:05.999999999+00:00")
	}
	return t.UTC().Format(time.RFC3339Nano)
}

// truncatePrecision drops fractional-second digits below microseconds,
// testing sub-microsecond retention.
func truncatePrecision(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	return t.Format("2006-01-02T15:04:05.000000Z07:00")
}

func (w *reorderWindow) popDisplaced(atNS int64) (Delivered, Delivered, bool) {
	// Emit the later record first when displacing an out-of-order pair.
	head := w.buf[0]
	w.buf = w.buf[1:]
	if len(w.buf) > 0 && atNS%2 == 0 {
		later := w.buf[0]
		w.buf = w.buf[1:]
		return later, head, true
	}
	return head, Delivered{}, false
}

package domain

import (
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

type reorderWindow struct {
	maxDisplacement int
	buf             []Delivered
	rng             *randutil.SplitMix64
}

func newReorderWindow(disp int, rng *randutil.SplitMix64) *reorderWindow {
	if disp < 1 {
		disp = 1
	}
	return &reorderWindow{maxDisplacement: disp, rng: rng}
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

func addSeconds(ts string, secs float64) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	return model.FormatTime(t.UnixNano() + int64(secs*1e9))
}

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

func truncatePrecision(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	return t.Format("2006-01-02T15:04:05.000000Z07:00")
}

func (w *reorderWindow) popDisplaced(_ int64) (Delivered, Delivered, bool) {

	head := w.buf[0]
	w.buf = w.buf[1:]
	if len(w.buf) > 0 && w.rng.Intn(2) == 0 {
		later := w.buf[0]
		w.buf = w.buf[1:]
		return later, head, true
	}
	return head, Delivered{}, false
}

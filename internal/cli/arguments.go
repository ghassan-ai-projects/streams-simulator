package cli

import (
	"fmt"
	"strings"
	"time"
)

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseTriple splits "a=b@c" by two separators into three parts.
func parseTriple(s, sep1, sep2 string) (string, string, float64, error) {
	rest := s
	a := ""
	if i := strings.Index(rest, sep1); i >= 0 {
		a = rest[:i]
		rest = rest[i+len(sep1):]
	} else {
		return "", "", 0, fmt.Errorf("bad triple %q", s)
	}
	b := ""
	offset := 0.0
	if i := strings.Index(rest, sep2); i >= 0 {
		b = rest[:i]
		offset = parseF(rest[i+len(sep2):])
	} else {
		b = rest
	}
	return a, b, offset, nil
}

func parseF(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

func timeNanos() int64 { return time.Now().UnixNano() }

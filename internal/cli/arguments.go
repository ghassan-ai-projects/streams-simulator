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
	index := strings.Index(s, sep1)
	if index < 0 {
		return "", "", 0, fmt.Errorf("bad triple %q", s)
	}
	first := s[:index]
	second, offset := parseOffset(s[index+len(sep1):], sep2)
	return first, second, offset, nil
}

func parseF(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

func timeNanos() int64 { return time.Now().UnixNano() }

func parseOffset(rest, separator string) (string, float64) {
	if index := strings.Index(rest, separator); index >= 0 {
		return rest[:index], parseF(rest[index+len(separator):])
	}
	return rest, 0
}

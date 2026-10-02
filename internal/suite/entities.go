package suite

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"strconv"
)

func defaultEntities(spec *domain.Compiled, prof *model.Profile) []string {
	n := prof.EntityCount
	if n <= 0 {
		n = spec.Spec.Entities.Count.Default
	}
	if n <= 0 {
		n = 1
	}
	var out []string
	for i := 1; i <= n; i++ {
		out = append(out, renderID(spec.Spec.Entities.IDTemplate, i, nil))
	}
	return out
}

func renderID(tmpl string, n int, params map[string]any) string {
	out := tmpl
	out = replaceAll(out, "{n}", strconv.Itoa(n))
	for {
		start := indexOf(out, "{")
		end := indexOf(out, "}")
		if start < 0 || end < 0 || end < start {
			break
		}
		name := out[start+1 : end]
		val := "a"
		if params != nil {
			if v, ok := params[name]; ok {
				val = fmt.Sprint(v)
			}
		}
		out = out[:start] + val + out[end+1:]
	}
	return out
}

func replaceAll(s, old, new string) string {
	out := ""
	for {
		i := indexOf(s, old)
		if i < 0 {
			return out + s
		}
		out += s[:i] + new
		s = s[i+len(old):]
	}
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

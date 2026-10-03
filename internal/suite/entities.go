package suite

import (
	"fmt"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
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
	out := replaceAll(tmpl, "{n}", strconv.Itoa(n))
	for {
		start, end := indexOf(out, "{"), indexOf(out, "}")
		if start < 0 || end < 0 || end < start {
			return out
		}
		out = out[:start] + entityParameter(params, out[start+1:end]) + out[end+1:]
	}
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

func entityParameter(params map[string]any, name string) string {
	if params != nil {
		if value, ok := params[name]; ok {
			return fmt.Sprint(value)
		}
	}
	return "a"
}

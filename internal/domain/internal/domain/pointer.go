package domain

import (
	"fmt"
	"strings"
)

func documentPointer(runtime string) string {
	section, _, _ := strings.Cut(strings.TrimPrefix(runtime, "/"), "/")
	switch section {
	case "state", "dynamics", "channels", "faults", "effectors":
		return "/kinds/0" + runtime
	case "entities":
		return "/instances/0" + strings.TrimPrefix(runtime, "/entities")
	}
	return runtime
}

func itemPointer(section string, index int) string {
	return fmt.Sprintf("/%s/%d", section, index)
}

func memberPointer(name string) string {
	name = strings.ReplaceAll(name, "~", "~0")
	return "/" + strings.ReplaceAll(name, "/", "~1")
}

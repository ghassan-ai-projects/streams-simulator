package domain

import (
	"fmt"
	"strings"
)

func itemPointer(section string, index int) string {
	return fmt.Sprintf("/%s/%d", section, index)
}

func memberPointer(name string) string {
	name = strings.ReplaceAll(name, "~", "~0")
	return "/" + strings.ReplaceAll(name, "/", "~1")
}

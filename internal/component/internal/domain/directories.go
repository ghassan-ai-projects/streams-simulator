package domain

import (
	"path/filepath"
	"strings"
)

func CustomerDirectories(flags []string, environment string) []string {
	named := []string{}
	for _, dir := range append(append([]string{}, flags...), filepath.SplitList(environment)...) {
		if strings.TrimSpace(dir) != "" {
			named = append(named, dir)
		}
	}
	return named
}

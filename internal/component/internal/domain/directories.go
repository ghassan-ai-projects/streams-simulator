package domain

import (
	"slices"
	"strings"
)

const DirectoriesEnv = "STREAMSIM_COMPONENTS_DIRS"

func CustomerDirectories(flags, fromEnvironment []string) []string {
	named := []string{}
	for _, dir := range slices.Concat(flags, fromEnvironment) {
		if strings.TrimSpace(dir) != "" {
			named = append(named, dir)
		}
	}
	return named
}

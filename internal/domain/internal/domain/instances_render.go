package domain

import (
	"slices"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func InstanceRenders(set model.InstanceSet, id string) bool {
	if slices.Contains(set.Names, id) {
		return true
	}
	if set.IDTemplate == "" {
		return false
	}
	match := idTemplatePattern(set.IDTemplate).FindStringSubmatch(id)
	return match != nil && indexWithin(match[1:], countLimit(set.Count))
}

func countLimit(count model.EntityCount) int {
	if count.Max > 0 {
		return count.Max
	}
	return count.Default
}

func indexWithin(captures []string, limit int) bool {
	if len(captures) == 0 {
		return limit >= 1
	}
	index, err := strconv.Atoi(captures[0])
	return err == nil && index >= 1 && index <= limit && sameIndex(captures, index)
}

func sameIndex(captures []string, index int) bool {
	return !slices.ContainsFunc(captures, func(text string) bool { return text != strconv.Itoa(index) })
}

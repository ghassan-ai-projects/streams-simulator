package domain

import (
	"slices"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const unnamedTemplateValue = "a"

func PopulationIDs(c *Compiled, firstSetCount int) []string {
	var ids []string
	counted := firstTemplatedSet(c)
	for i := range c.Spec.Instances {
		set := &c.Spec.Instances[i]
		count := set.Count.Default
		if i == counted && firstSetCount > 0 {
			count = firstSetCount
		}
		ids = append(ids, instanceSetIDs(set, count)...)
	}
	return ids
}

func firstTemplatedSet(c *Compiled) int {
	return slices.IndexFunc(c.Spec.Instances, func(set model.InstanceSet) bool { return set.IDTemplate != "" })
}

func instanceSetIDs(set *model.InstanceSet, count int) []string {
	ids := slices.Clone(set.Names)
	for n := 1; set.IDTemplate != "" && n <= max(count, 1); n++ {
		ids = append(ids, RenderInstanceID(set.IDTemplate, n))
	}
	return ids
}

func RenderInstanceID(template string, index int) string {
	out := strings.ReplaceAll(template, "{n}", strconv.Itoa(index))
	for {
		start, end := strings.Index(out, "{"), strings.Index(out, "}")
		if start < 0 || end < 0 || end < start {
			return out
		}
		out = out[:start] + unnamedTemplateValue + out[end+1:]
	}
}

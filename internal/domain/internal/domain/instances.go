package domain

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var templatePlaceholder = regexp.MustCompile(`\{[^{}]*\}`)

func (checks domainChecks) checkInstances() error {
	return checks.each("instances", len(checks.compiled.Spec.Instances), func(at domainChecks, i int) error {
		return at.checkInstanceCount(&checks.compiled.Spec.Instances[i])
	})
}

func (checks domainChecks) checkInstanceCount(set *model.InstanceSet) error {
	if set.Count.Min > 0 && set.Count.Default < set.Count.Min {
		return checks.bad("/count/default", "instance count default %d is below min %d", set.Count.Default, set.Count.Min)
	}
	if set.Count.Max > 0 && set.Count.Default > set.Count.Max {
		return checks.bad("/count/default", "instance count default %d is above max %d", set.Count.Default, set.Count.Max)
	}
	return nil
}

func (c *Compiled) checkInstanceKinds(src string) error {
	named, err := c.namedKinds(src)
	if err != nil {
		return err
	}
	for i := range c.Spec.Kinds {
		if !named[c.Spec.Kinds[i].Name] {
			return invalidDomain(src, itemPointer("kinds", i)+"/name", fmt.Sprintf("kind %q has no instance set", c.Spec.Kinds[i].Name))
		}
	}
	return nil
}

func (c *Compiled) namedKinds(src string) (map[string]bool, error) {
	named := map[string]bool{}
	for i := range c.Spec.Instances {
		kind := c.Spec.Instances[i].Kind
		if c.Kind(kind) == nil {
			return nil, invalidDomain(src, itemPointer("instances", i)+"/kind", fmt.Sprintf("instances name kind %q, which is not declared", kind))
		}
		named[kind] = true
	}
	return named, nil
}

func (c *Compiled) compileIDPatterns() {
	for i := range c.Spec.Instances {
		c.idPatterns = append(c.idPatterns, idTemplatePattern(c.Spec.Instances[i].IDTemplate))
	}
}

func (c *Compiled) InstanceSetFor(id string) int {
	for i, pattern := range c.idPatterns {
		if pattern.MatchString(id) {
			return i
		}
	}
	return -1
}

func (c *Compiled) CheckEntityIDs(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	counts, err := c.countBySet(ids)
	if err != nil {
		return err
	}
	for i, count := range counts {
		if err := c.checkEntityCount(&c.Spec.Instances[i], count); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiled) countBySet(ids []string) ([]int, error) {
	counts := make([]int, len(c.Spec.Instances))
	for _, id := range ids {
		set := c.InstanceSetFor(id)
		if set < 0 {
			return nil, c.unfitID(id)
		}
		counts[set]++
	}
	return counts, nil
}

func (c *Compiled) unfitID(id string) error {
	templates := make([]string, len(c.Spec.Instances))
	for i := range c.Spec.Instances {
		templates[i] = c.Spec.Instances[i].IDTemplate
	}
	return fmt.Errorf("domain: entity id %q does not fit instance id template %q", id, strings.Join(templates, " or "))
}

func (c *Compiled) checkEntityCount(set *model.InstanceSet, count int) error {
	if count == 0 {
		return nil
	}
	if count < max(set.Count.Min, 1) {
		return fmt.Errorf("domain: %d entity ids is below the instance count minimum %d", count, max(set.Count.Min, 1))
	}
	if set.Count.Max > 0 && count > set.Count.Max {
		return fmt.Errorf("domain: %d entity ids is above the instance count maximum %d", count, set.Count.Max)
	}
	return nil
}

func idTemplatePattern(template string) *regexp.Regexp {
	literals := templatePlaceholder.Split(template, -1)
	placeholders := templatePlaceholder.FindAllString(template, -1)
	var pattern strings.Builder
	pattern.WriteString("^")
	for i, literal := range literals {
		pattern.WriteString(regexp.QuoteMeta(literal))
		if i < len(placeholders) {
			pattern.WriteString(placeholderPattern(placeholders[i]))
		}
	}
	pattern.WriteString("$")
	return regexp.MustCompile(pattern.String())
}

func placeholderPattern(placeholder string) string {
	if placeholder == "{n}" {
		return "[0-9]+"
	}
	return "[A-Za-z0-9_.-]+"
}

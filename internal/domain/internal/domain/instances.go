package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var templatePlaceholder = regexp.MustCompile(`\{[^{}]*\}`)

func (checks domainChecks) checkInstances() error {
	entities := checks.compiled.Spec.Entities
	at := checks.within("/entities")
	if entities.Count.Min > 0 && entities.Count.Default < entities.Count.Min {
		return at.bad("/count/default", "instance count default %d is below min %d", entities.Count.Default, entities.Count.Min)
	}
	if entities.Count.Max > 0 && entities.Count.Default > entities.Count.Max {
		return at.bad("/count/default", "instance count default %d is above max %d", entities.Count.Default, entities.Count.Max)
	}
	return nil
}

func (c *Compiled) CheckEntityIDs(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	if err := c.checkEntityCount(len(ids)); err != nil {
		return err
	}
	pattern := idTemplatePattern(c.Spec.Entities.IDTemplate)
	for _, id := range ids {
		if !pattern.MatchString(id) {
			return fmt.Errorf("domain: entity id %q does not fit instance id template %q", id, c.Spec.Entities.IDTemplate)
		}
	}
	return nil
}

func (c *Compiled) checkEntityCount(count int) error {
	entities := c.Spec.Entities
	if count < max(entities.Count.Min, 1) {
		return fmt.Errorf("domain: %d entity ids is below the instance count minimum %d", count, max(entities.Count.Min, 1))
	}
	if entities.Count.Max > 0 && count > entities.Count.Max {
		return fmt.Errorf("domain: %d entity ids is above the instance count maximum %d", count, entities.Count.Max)
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

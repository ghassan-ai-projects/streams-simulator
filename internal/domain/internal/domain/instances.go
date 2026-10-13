package domain

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var templatePlaceholder = regexp.MustCompile(`\{[^{}]*\}`)

func (checks domainChecks) checkInstances() error {
	err := checks.each("instances", len(checks.compiled.Spec.Instances), func(at domainChecks, i int) error {
		return at.checkInstance(&checks.compiled.Spec.Instances[i])
	})
	if err != nil {
		return err
	}
	return checks.checkTemplateOverlaps()
}

func (checks domainChecks) checkInstance(set *model.InstanceSet) error {
	if err := checks.instanceShape(set); err != nil {
		return err
	}
	if err := checks.instanceNames(set); err != nil {
		return err
	}
	return checks.checkInstanceCount(set)
}

func (checks domainChecks) instanceShape(set *model.InstanceSet) error {
	switch {
	case set.IDTemplate == "" && len(set.Names) == 0:
		return checks.bad("/id_template", "an instance set needs names or an id_template")
	case set.IDTemplate != "" && set.Count.Default == 0:
		return checks.bad("/count", "an instance set with an id_template needs a count")
	case set.IDTemplate == "" && set.Count.Default != 0:
		return checks.bad("/count", "count belongs to an id_template; a set of names has none")
	case set.IDTemplate == "" && set.Churn != nil:
		return checks.bad("/churn", "churn births ids from an id_template; a set of names has none")
	}
	return nil
}

func (checks domainChecks) instanceNames(set *model.InstanceSet) error {
	for i, name := range set.Names {
		if template := checks.compiled.fittingTemplate(name); template != "" {
			return checks.bad(fmt.Sprintf("/names/%d", i), "instance name %q fits id_template %q, so a template could render it too", name, template)
		}
	}
	return nil
}

func (c *Compiled) fittingTemplate(name string) string {
	for i, pattern := range c.idPatterns {
		if pattern != nil && pattern.MatchString(name) {
			return c.Spec.Instances[i].IDTemplate
		}
	}
	return ""
}

func (checks domainChecks) checkInstanceCount(set *model.InstanceSet) error {
	if set.IDTemplate == "" {
		return nil
	}
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

func (c *Compiled) compileInstanceIDs() {
	for i := range c.Spec.Instances {
		set := &c.Spec.Instances[i]
		c.idPatterns = append(c.idPatterns, templatePatternOrNil(set.IDTemplate))
		c.names = append(c.names, nameSet(set.Names))
		c.effective = append(c.effective, c.Kind(set.Kind))
	}
}

func templatePatternOrNil(template string) *regexp.Regexp {
	if template == "" {
		return nil
	}
	return idTemplatePattern(template)
}

func nameSet(names []string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, name := range names {
		out[name] = true
	}
	return out
}

func (c *Compiled) registerInstanceNames(src string) error {
	seen := map[string]bool{}
	for i := range c.Spec.Instances {
		for j, name := range c.Spec.Instances[i].Names {
			at := itemPointer("instances", i) + itemPointer("names", j)
			if err := registerSymbol(seen, name, "instance name", at, src); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Compiled) InstanceSetFor(id string) int {
	for i, names := range c.names {
		if names[id] {
			return i
		}
	}
	for i, pattern := range c.idPatterns {
		if pattern != nil && pattern.MatchString(id) {
			return i
		}
	}
	return -1
}

func (c *Compiled) IsInstanceName(id string) bool {
	set := c.InstanceSetFor(id)
	return set >= 0 && c.names[set][id]
}

func (c *Compiled) InstanceKind(set int) *model.Kind {
	if set < 0 || set >= len(c.effective) {
		return nil
	}
	return c.effective[set]
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
		if !c.names[set][id] {
			counts[set]++
		}
	}
	return counts, nil
}

func (c *Compiled) unfitID(id string) error {
	var fits []string
	for i := range c.Spec.Instances {
		set := &c.Spec.Instances[i]
		fits = append(fits, set.Names...)
		if set.IDTemplate != "" {
			fits = append(fits, set.IDTemplate)
		}
	}
	return fmt.Errorf("domain: entity id %q does not fit instance id template %q", id, strings.Join(fits, " or "))
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

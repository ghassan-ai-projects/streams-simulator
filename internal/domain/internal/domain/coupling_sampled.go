package domain

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) sampledRead(linkType string, sampleS float64, member string) error {
	lt := checks.linkType(linkType)
	if sampleS == 0 || lt == nil {
		return nil
	}
	sampleMs := sampleS * 1000
	if sampleMs != math.Trunc(sampleMs) {
		return checks.bad(member, "a sample period of %v s is not a whole number of milliseconds", sampleS)
	}
	if err := checks.sampleOnSteps(checks.compiled.Kind(lt.From), int64(sampleMs), member); err != nil {
		return err
	}
	return checks.sampledSourcesStay(lt.From, member)
}

func (checks domainChecks) sampleOnSteps(source *model.Kind, sampleMs int64, member string) error {
	if source == nil {
		return nil
	}
	for i := range source.Dynamics {
		if dyn := &source.Dynamics[i]; dyn.Integrated() && sampleMs%dyn.StepMs() != 0 {
			return checks.bad(member, "a sample period of %d ms is not a multiple of the %d ms step of %s state %q", sampleMs, dyn.StepMs(), source.Name, dyn.Target)
		}
	}
	return nil
}

func (checks domainChecks) sampledSourcesStay(kind, member string) error {
	for _, set := range checks.setsOf(kind) {
		if checks.compiled.Spec.Instances[set].Churn != nil {
			return checks.bad(member, "kind %q is read by sample, so its instances may not churn", kind)
		}
	}
	return nil
}

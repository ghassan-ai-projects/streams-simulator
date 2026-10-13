package domain

import "slices"

func (checks domainChecks) checkKPIs() error {
	kind := checks.subject
	return checks.each("kpis", len(kind.KPIs), func(at domainChecks, i int) error {
		state := kind.KPIs[i]
		if slices.Contains(kind.KPIs[:i], state) {
			return at.bad("", "KPI %q is declared twice", state)
		}
		if !checks.hasState(state) {
			return at.bad("", "KPI %q is not a declared state", state)
		}
		return nil
	})
}

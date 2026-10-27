package domain

import (
	"cmp"
	"maps"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func (p *CampaignPhase) check(compiled *domain.Compiled) error {
	switch {
	case p.Name == "":
		return specErrorf("a phase needs a name")
	case p.Runs < 1:
		return specErrorf("phase %q needs at least one run", p.Name)
	case p.NegativeShare < 0 || p.NegativeShare > 1:
		return specErrorf("phase %q: negative_share must be within 0 and 1", p.Name)
	case compiled == nil:
		return specErrorf("phase %q: domain %q is not available", p.Name, p.Domain)
	}
	return p.checkFaults(compiled)
}

func (p *CampaignPhase) checkFaults(compiled *domain.Compiled) error {
	negative, hasNegative := negativeFaultOf(compiled)
	if p.negatives() > 0 && !hasNegative {
		return specErrorf("phase %q asks for negative runs, which %q declares no negative-class fault for", p.Name, p.Domain)
	}
	if len(p.Faults) == 0 && p.negatives() < p.Runs {
		return specErrorf("phase %q lists no fault for its positive runs", p.Name)
	}
	if err := p.checkPositive(compiled); err != nil {
		return err
	}
	return p.checkScriptable(compiled, negative)
}

func (p *CampaignPhase) checkPositive(compiled *domain.Compiled) error {
	for _, id := range slices.Sorted(maps.Keys(p.Faults)) {
		if p.Faults[id] <= 0 || p.Faults[id] > maxFaultWeight || isNegativeClass(compiled, id) || len(faultEntities(compiled, id)) == 0 {
			return specErrorf("phase %q: fault %q needs a weight above 0 and at most %g and must be a positive-class fault %q declares", p.Name, id, maxFaultWeight, p.Domain)
		}
	}
	return nil
}

func (p *CampaignPhase) checkScriptable(compiled *domain.Compiled, negative string) error {
	used := slices.Sorted(maps.Keys(p.Faults))
	if p.negatives() > 0 {
		used = append(used, negative)
	}
	for _, id := range used {
		if commandedOn(compiled, id) {
			return specErrorf("phase %q: fault %q acts on a state a command sets, which a session cannot script yet (D-116)", p.Name, id)
		}
	}
	return nil
}

func (p *CampaignPhase) negatives() int {
	return int(math.Round(float64(p.Runs) * p.NegativeShare))
}

func (p *CampaignPhase) schedule(negative string, rng *randutil.SplitMix64) []string {
	faults := slices.Repeat([]string{negative}, p.negatives())
	faults = append(faults, apportion(p.Faults, p.Runs-p.negatives())...)
	for i := len(faults) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		faults[i], faults[j] = faults[j], faults[i]
	}
	return faults
}

type faultShare struct {
	id        string
	count     int
	remainder float64
}

func apportion(weights map[string]float64, total int) []string {
	shares := floorShares(weights, total)
	grantLeftovers(shares, total)
	var faults []string
	for _, share := range shares {
		faults = append(faults, slices.Repeat([]string{share.id}, share.count)...)
	}
	return faults
}

func floorShares(weights map[string]float64, total int) []*faultShare {
	ids := slices.Sorted(maps.Keys(weights))
	sum := 0.0
	for _, id := range ids {
		sum += weights[id]
	}
	var shares []*faultShare
	for _, id := range ids {
		exact := weights[id] / sum * float64(total)
		shares = append(shares, &faultShare{id: id, count: int(exact), remainder: exact - math.Floor(exact)})
	}
	return shares
}

func grantLeftovers(shares []*faultShare, total int) {
	left := total
	for _, share := range shares {
		left -= share.count
	}
	byRemainder := slices.Clone(shares)
	slices.SortStableFunc(byRemainder, func(a, b *faultShare) int { return cmp.Compare(b.remainder, a.remainder) })
	for _, share := range byRemainder[:left] {
		share.count++
	}
}

func faultEntities(compiled *domain.Compiled, fault string) []string {
	var ids []string
	for _, id := range world.PopulationIDs(compiled, 0) {
		if kindOfEntity(compiled, id).FaultDecl(fault) != nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func negativeFaultOf(compiled *domain.Compiled) (string, bool) {
	for _, fault := range declaredFaults(compiled) {
		if fault.IsNegativeClass {
			return fault.ID, true
		}
	}
	return "", false
}

func commandedOn(compiled *domain.Compiled, fault string) bool {
	return slices.ContainsFunc(faultEntities(compiled, fault), func(id string) bool {
		kind := kindOfEntity(compiled, id)
		return kind.FaultCommanded(kind.FaultDecl(fault))
	})
}

func isPulse(compiled *domain.Compiled, entity, fault string) bool {
	return kindOfEntity(compiled, entity).FaultDecl(fault).IsPulse()
}

func isNegativeClass(compiled *domain.Compiled, fault string) bool {
	return slices.ContainsFunc(declaredFaults(compiled), func(f *model.Fault) bool { return f.ID == fault && f.IsNegativeClass })
}

// Package suite generates graded scenario suites: declared negative-class
// fraction, randomized onset (including pre-degraded starts), perturbation
// coverage, and a generate-audit-regenerate loop that keeps only
// non-trivial scenarios. A domain that cannot produce non-trivial scenarios
// at the declared prevalence reports a terminal state rather than an empty
// directory (G-06).
package suite

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

// Perturbation is one delivery perturbation in a scenario.
type Perturbation struct {
	Name    string         `json:"name"`
	Params  map[string]any `json:"params,omitempty"`
	FromNS  int64          `json:"from_ns,omitempty"`
	UntilNS int64          `json:"until_ns,omitempty"`
}

// Scenario is one executable scenario recipe.
type Scenario struct {
	ID            string            `json:"id"`
	Seed          uint64            `json:"seed"`
	Profile       string            `json:"profile"`
	EntityID      string            `json:"entity_id"`
	Fault         string            `json:"fault"`
	StartNS       int64             `json:"start_ns"`
	OnsetNS       int64             `json:"onset_ns"`
	DurationNS    int64             `json:"duration_ns"`
	PreDegraded   bool              `json:"pre_degraded,omitempty"`
	Perturbations []Perturbation    `json:"perturbations,omitempty"`
	Setup         []truth.SetupCall `json:"setup,omitempty"`
	CommandLog    []model.Command   `json:"command_log"`
}

// TrivialCase is a scenario the audit rejected, kept as a mechanism
// regression fixture.
type TrivialCase struct {
	ScenarioID string             `json:"scenario_id"`
	Fault      string             `json:"fault"`
	Scores     map[string]float64 `json:"scores"`
	Best       string             `json:"best"`
}

// Suite is a generated graded suite.
type Suite struct {
	DomainID        string                    `json:"domain"`
	Profile         string                    `json:"profile"`
	Seed            uint64                    `json:"seed"`
	Scenarios       []Scenario                `json:"scenarios"`
	Labels          []model.GroundTruthRecord `json:"labels"`
	TrivialExcluded []TrivialCase             `json:"trivial_excluded,omitempty"`
	Attempts        int                       `json:"attempts"`
	TerminalState   string                    `json:"terminal_state,omitempty"`
}

// Config pins suite generation.
type Config struct {
	Domain      *domain.Compiled
	Profile     string
	N           int
	Seed        uint64
	SampleNS    int64
	MaxAttempts int
}

// Generate builds a suite with the generate-audit-regenerate loop.
func Generate(cfg Config) (*Suite, error) {
	if cfg.N <= 0 {
		cfg.N = 100
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3 * cfg.N
	}
	prof := cfg.Domain.Profile(cfg.Profile)
	if prof == nil {
		return nil, fmt.Errorf("suite: unknown profile %q", cfg.Profile)
	}
	rng := randutil.NewSplitMix64(cfg.Seed ^ randutil.Fnv1a64(cfg.Domain.Spec.ID+"/"+cfg.Profile))
	gt := cfg.Domain.Spec.GroundTruth
	negativeFrac := gt.NegativeClassFraction
	if negativeFrac <= 0 {
		negativeFrac = 0.4
	}
	solver := truth.NewSolver(cfg.Domain, cfg.Seed, cfg.SampleNS, 72*3600*1e9)
	panel := audit.NewPanel(cfg.Domain, cfg.Seed^0x5eed, cfg.SampleNS)

	s := &Suite{
		DomainID: cfg.Domain.Spec.ID,
		Profile:  cfg.Profile,
		Seed:     cfg.Seed,
	}
	perturbCount := map[string]int{}
	negatives := 0
	preDegraded := 0
	cascadeCount := 0
	pathologyCount := 0

	startNS := model.DefaultStartTimeNS + 4*3600*1e9
	durationNS := int64(prof.DurationS * 1e9)
	if durationNS <= 0 {
		durationNS = 72 * 3600 * 1e9
	}
	entities := defaultEntities(cfg.Domain, prof)

	for len(s.Scenarios) < cfg.N && s.Attempts < cfg.MaxAttempts {
		s.Attempts++
		idx := len(s.Scenarios)
		scenarioSeed := cfg.Seed + uint64(idx)*0x9e3779b97f4a7c15
		// Adaptive negative sampling: the audit rejects positive scenarios
		// aggressively, so the admission fraction drifts above the declared
		// band; proportional control on the sampling probability holds the
		// band when the domain can sustain it, and the terminal state
		// reports when it cannot (G-06).
		admitted := 0.0
		if len(s.Scenarios) > 0 {
			admitted = float64(negatives) / float64(len(s.Scenarios))
		}
		sampleFrac := negativeFrac + (negativeFrac-admitted)*2.0
		if sampleFrac < 0.1 {
			sampleFrac = 0.1
		}
		if sampleFrac > 0.7 {
			sampleFrac = 0.7
		}
		sc, label, verdict, err := s.buildScenario(cfg, rng, solver, panel, prof, entities,
			startNS, durationNS, scenarioSeed, idx, &negatives, &preDegraded, &cascadeCount, &pathologyCount, perturbCount, sampleFrac)
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		if verdict != nil && verdict.Trivial {
			// Kept as a mechanism fixture, excluded from the graded set;
			// the loop keeps going for a replacement.
			s.TrivialExcluded = append(s.TrivialExcluded, TrivialCase{
				ScenarioID: sc.ID, Fault: sc.Fault, Scores: verdict.Scores, Best: verdict.Best,
			})
			continue
		}
		s.Scenarios = append(s.Scenarios, *sc)
		s.Labels = append(s.Labels, *label)
	}

	// Perturbation coverage: every catalog perturbation in >= 5 scenarios.
	var shortfalls []string
	if missing := uncovered(perturbCount); len(missing) > 0 {
		shortfalls = append(shortfalls, fmt.Sprintf("perturbation coverage not met at the declared size: %v", missing))
	}
	// Composition checks. All shortfalls are aggregated into one reportable
	// terminal state (G-06): a domain that cannot sustain the declared
	// composition says so, rather than silently degrading.
	negShare := float64(negatives) / float64(len(s.Scenarios))
	if len(s.Scenarios) > 0 && (negShare < 0.35 || negShare > 0.45) {
		shortfalls = append(shortfalls, fmt.Sprintf("negative-class fraction %.2f outside [0.35, 0.45]", negShare))
	}
	if len(s.Scenarios) > 0 && preDegraded*100/len(s.Scenarios) < 10 {
		shortfalls = append(shortfalls, "pre-degraded fraction below 10%")
	}
	if cascadeCount < 3 {
		shortfalls = append(shortfalls, fmt.Sprintf("correlated_cascade scenarios: %d < 3", cascadeCount))
	}
	if pathologyCount < 10 {
		shortfalls = append(shortfalls, fmt.Sprintf("sensor_pathology scenarios: %d < 10", pathologyCount))
	}
	if s.Attempts >= cfg.MaxAttempts && len(s.Scenarios) < cfg.N {
		shortfalls = append(shortfalls, fmt.Sprintf("audit rejected too many scenarios: %d non-trivial of %d attempts", len(s.Scenarios), s.Attempts))
	}
	s.TerminalState = joinShortfalls(shortfalls)
	return s, nil
}

func joinShortfalls(shortfalls []string) string {
	out := ""
	for i, sf := range shortfalls {
		if i > 0 {
			out += "; "
		}
		out += sf
	}
	return out
}

func defaultEntities(spec *domain.Compiled, prof *model.Profile) []string {
	n := prof.EntityCount
	if n <= 0 {
		n = spec.Spec.Entities.Count.Default
	}
	if n <= 0 {
		n = 1
	}
	var out []string
	for i := 1; i <= n; i++ {
		out = append(out, renderID(spec.Spec.Entities.IDTemplate, i, nil))
	}
	return out
}

func renderID(tmpl string, n int, params map[string]any) string {
	out := tmpl
	out = replaceAll(out, "{n}", strconv.Itoa(n))
	for {
		start := indexOf(out, "{")
		end := indexOf(out, "}")
		if start < 0 || end < 0 || end < start {
			break
		}
		name := out[start+1 : end]
		val := "a"
		if params != nil {
			if v, ok := params[name]; ok {
				val = fmt.Sprint(v)
			}
		}
		out = out[:start] + val + out[end+1:]
	}
	return out
}

func replaceAll(s, old, new string) string {
	out := ""
	for {
		i := indexOf(s, old)
		if i < 0 {
			return out + s
		}
		out += s[:i] + new
		s = s[i+len(old):]
	}
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// buildScenario constructs one candidate scenario and its audit verdict.
func (s *Suite) buildScenario(cfg Config, rng *randutil.SplitMix64, solver *truth.Solver, panel *audit.Panel,
	prof *model.Profile, entities []string, startNS, durationNS int64, seed uint64, idx int,
	negatives, preDegraded, cascadeCount, pathologyCount *int, perturbCount map[string]int, sampleFrac float64) (*Scenario, *model.GroundTruthRecord, *audit.Verdict, error) {

	gt := cfg.Domain.Spec.GroundTruth
	isNegative := rng.Float64() < sampleFrac
	faultID := s.pickFault(cfg, prof, rng, isNegative, idx)

	entity := entities[rng.Intn(len(entities))]
	profileName := cfg.Profile
	preDeg := false
	var onsetNS int64
	// Randomized onset: uniform across the first half of the trace, so the
	// fault has room to develop and run-to-failure bias is countered. 10-15%
	// of scenarios begin with the asset already degraded (fault before t0).
	if rng.Float64() < gt.PreDegradedFraction {
		preDeg = true
		onsetNS = startNS - int64(rng.Float64()*float64(durationNS/4))
	} else {
		onsetNS = startNS + int64(rng.Float64()*float64(durationNS/2))
	}
	if onsetNS < 0 {
		onsetNS = 0
	}

	// Perturbations: sample 0-2, with forced coverage for thin ones.
	scPerts := []Perturbation{}
	for _, name := range pickPerturbations(rng, perturbCount) {
		params := samplePerturbParams(rng, name)
		from := startNS + int64(rng.Float64()*float64(durationNS/2))
		until := from + int64((0.25+rng.Float64()*0.5)*float64(durationNS/2))
		scPerts = append(scPerts, Perturbation{Name: name, Params: params, FromNS: from, UntilNS: until})
		perturbCount[name]++
	}

	scenario := &Scenario{
		ID:            fmt.Sprintf("%s/%04d", cfg.Domain.Spec.ID, idx),
		Seed:          seed,
		Profile:       profileName,
		EntityID:      entity,
		Fault:         faultID,
		StartNS:       startNS,
		OnsetNS:       onsetNS,
		DurationNS:    durationNS,
		PreDegraded:   preDeg,
		Perturbations: scPerts,
	}

	// Setup: aquaculture scenarios start the aerator for the night so
	// aerator_failure is meaningful — except pre-degraded starts, where the
	// fault predates t0 and starting the aerator would mask it.
	var setup []truth.SetupCall
	if !preDeg {
		setup = s.defaultSetup(cfg.Domain, entity, startNS)
	}
	scenario.Setup = setup

	// Seal the label first: unobservable scenarios (no signal crosses the
	// noise floor within the horizon) cannot be graded and are regenerated.
	var pertStrs []string
	for _, p := range scPerts {
		pertStrs = append(pertStrs, p.Name+"@"+fmt.Sprint(p.Params["rate"]))
	}
	label, err := truth.BuildRecord(cfg.Domain, solver, scenario.ID, seed, entity, faultID,
		onsetNS, startNS, entities, preDeg, pertStrs, setup)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("suite: build label: %w", err)
	}
	if label.FirstObservableTimeNS == 0 {
		return scenario, label, &audit.Verdict{Trivial: true}, nil
	}

	// Audit the candidate: trivial scenarios are excluded from the graded
	// set and the loop regenerates.
	verdict, err := panel.Audit(entity, faultID, onsetNS, startNS, entities, durationNS, setup)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("suite: audit: %w", err)
	}

	label.TrivialBaselineVerdict = verdictTrivialString(verdict)
	if verdict != nil {
		label.TrivialBaselineDetail = verdict.Scores
	}
	if isNegative {
		*negatives++
	}
	if preDeg {
		*preDegraded++
	}
	if profileName == "correlated_cascade" {
		*cascadeCount++
	}
	if profileName == "sensor_pathology" {
		*pathologyCount++
	}

	// The executable command log.
	scenario.CommandLog = buildCommandLog(cfg.Domain, scenario)
	return scenario, label, verdict, nil
}

func verdictTrivialString(v *audit.Verdict) string {
	if v == nil {
		return model.TrivialNonTrivial
	}
	if v.Trivial {
		return model.TrivialTrivial
	}
	return model.TrivialNonTrivial
}

// pickFault samples a fault id per the profile weights and the negative
// fraction.
func (s *Suite) pickFault(cfg Config, prof *model.Profile, rng *randutil.SplitMix64, isNegative bool, idx int) string {
	if isNegative {
		for i := range cfg.Domain.Spec.Faults {
			if cfg.Domain.Spec.Faults[i].IsNegativeClass {
				return cfg.Domain.Spec.Faults[i].ID
			}
		}
	}
	if len(prof.FaultWeights) > 0 {
		total := 0.0
		ids := make([]string, 0, len(prof.FaultWeights))
		for id := range prof.FaultWeights {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			w := prof.FaultWeights[id]
			total += w
		}
		if total > 0 {
			r := rng.Float64() * total
			for _, id := range ids {
				w := prof.FaultWeights[id]
				if r < w {
					return id
				}
				r -= w
			}
		}
	}
	// Uniform over positive faults.
	var positives []string
	for i := range cfg.Domain.Spec.Faults {
		if !cfg.Domain.Spec.Faults[i].IsNegativeClass {
			positives = append(positives, cfg.Domain.Spec.Faults[i].ID)
		}
	}
	if len(positives) == 0 {
		return ""
	}
	sort.Strings(positives)
	return positives[rng.Intn(len(positives))]
}

// defaultSetup translates profile data into scenario context calls. The
// simulator never branches on a domain id or effector name here.
func (s *Suite) defaultSetup(spec *domain.Compiled, entity string, startNS int64) []truth.SetupCall {
	prof := spec.Profile(s.Profile)
	if prof == nil {
		return nil
	}
	var out []truth.SetupCall
	for i, setup := range prof.Setup {
		if setup.Effector == "" || spec.Effector(setup.Effector) == nil {
			continue
		}
		args := substituteEntity(setup.Args, entity)
		commandID := setup.CommandID
		if commandID == "" {
			commandID = fmt.Sprintf("setup-%d", i)
		}
		out = append(out, truth.SetupCall{Effector: setup.Effector, EntityID: entity, CommandID: commandID, Args: args, AtNS: startNS})
	}
	return out
}

func substituteEntity(in map[string]any, entity string) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		if s, ok := v.(string); ok && s == "{entity_id}" {
			out[k] = entity
		} else {
			out[k] = v
		}
	}
	return out
}

// buildCommandLog renders the scenario as replayable commands.
func buildCommandLog(spec *domain.Compiled, sc *Scenario) []model.Command {
	var cmds []model.Command
	seq := 0
	appendCmd := func(op string, args map[string]any, atNS int64) {
		cmds = append(cmds, model.Command{Seq: int64(seq), AtNS: atNS, Op: op, Args: args})
		seq++
	}
	appendCmd(model.OpWorldCreate, map[string]any{
		"domain": spec.Spec.ID, "seed": sc.Seed, "start_time_ns": sc.StartNS,
		"time_mode": model.TimeStepped,
	}, sc.StartNS)
	for _, call := range sc.Setup {
		appendCmd(model.OpEffectorInvoke, map[string]any{
			"effector": call.Effector, "entity_id": call.EntityID,
			"command_id": call.CommandID, "args": call.Args, "at_ns": call.AtNS,
		}, call.AtNS)
	}
	appendCmd(model.OpFaultInject, map[string]any{
		"entity_id": sc.EntityID, "fault": sc.Fault, "onset_ns": sc.OnsetNS,
	}, sc.OnsetNS)
	for _, p := range sc.Perturbations {
		appendCmd(model.OpPerturbApply, map[string]any{
			"perturbation": p.Name, "params": p.Params,
			"from_ns": p.FromNS, "until_ns": p.UntilNS,
		}, p.FromNS)
	}
	appendCmd(model.OpClockAdvance, map[string]any{
		"to_ns": sc.StartNS + sc.DurationNS, "await_consumer": false,
	}, sc.StartNS+sc.DurationNS)
	return cmds
}

func pickPerturbations(rng *randutil.SplitMix64, counts map[string]int) []string {
	// Force coverage: perturbations below 5 scenarios ride along.
	var forced []string
	for _, name := range perturb.Names {
		if counts[name] < 5 {
			forced = append(forced, name)
		}
	}
	if len(forced) > 0 {
		return forced[:min(2, len(forced))]
	}
	var out []string
	n := rng.Intn(3) // 0..2
	names := perturb.Names
	for i := 0; i < n; i++ {
		out = append(out, names[rng.Intn(len(names))])
	}
	return out
}

func samplePerturbParams(rng *randutil.SplitMix64, name string) map[string]any {
	p := map[string]any{}
	switch name {
	case "drop", "duplicate_burst", "id_reuse", "out_of_enum", "out_of_range":
		p["rate"] = round3(0.005 + rng.Float64()*0.04)
	case "delay_tail":
		p["mean_s"] = round1(30 + rng.Float64()*300)
		p["sigma_s"] = round1(60 + rng.Float64()*300)
	case "storm":
		p["multiplier"] = 2 + rng.Intn(4)
	case "clock_skew":
		p["offset_s"] = round1(60 + rng.Float64()*300)
	case "oversize":
		p["bytes"] = 4096
	}
	return p
}

func round3(f float64) float64 { return float64(int(f*1000+0.5)) / 1000 }
func round1(f float64) float64 { return float64(int(f*10+0.5)) / 10 }

func uncovered(counts map[string]int) []string {
	var out []string
	for _, name := range perturb.Names {
		if counts[name] < 5 {
			out = append(out, name)
		}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

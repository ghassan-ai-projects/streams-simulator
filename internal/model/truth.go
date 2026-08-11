package model

// GroundTruthRecord is the sealed label for one scenario, matching
// docs/contracts/ground-truth-v0.1.schema.json. Reachable only under the
// director role.
type GroundTruthRecord struct {
	ScenarioID             string             `json:"scenario_id"`
	Domain                 string             `json:"domain"`
	Seed                   uint64             `json:"seed"`
	EntityID               string             `json:"entity_id"`
	Label                  string             `json:"label"`
	IsNegativeClass        bool               `json:"is_negative_class,omitempty"`
	ExpectedEpisode        bool               `json:"expected_episode"`
	InjectionTimeNS        int64              `json:"injection_time_ns"`
	FirstObservableTimeNS  int64              `json:"first_observable_time_ns"`
	UnavoidableTimeNS      int64              `json:"unavoidable_time_ns"`
	Observability          ObservabilityInfo  `json:"observability"`
	DeadlineNS             int64              `json:"deadline_ns,omitempty"`
	Counterfactual         *Counterfactual    `json:"counterfactual,omitempty"`
	ExpectedEffector       string             `json:"expected_effector,omitempty"`
	TrivialBaselineVerdict string             `json:"trivial_baseline_verdict"`
	TrivialBaselineDetail  map[string]float64 `json:"trivial_baseline_detail,omitempty"`
	PreDegraded            bool               `json:"pre_degraded,omitempty"`
	Perturbations          []string           `json:"perturbations,omitempty"`
	WorldStateReference    string             `json:"world_state_reference,omitempty"`
}

// Trivial-baseline verdicts.
const (
	TrivialNonTrivial = "non_trivial"
	TrivialTrivial    = "trivial"
)

// ObservabilityInfo records how the onset timestamps were computed.
type ObservabilityInfo struct {
	DetectorForm       string   `json:"detector_form"`
	Channels           []string `json:"channels"`
	EffectiveSigma     float64  `json:"effective_sigma"`
	FirstObservableSNR float64  `json:"first_observable_snr"`
	UnavoidableSNR     float64  `json:"unavoidable_snr"`
	SolutionMethod     string   `json:"solution_method,omitempty"`
}

// Detector forms.
const (
	DetectorSingleChannel = "single_channel_snr"
	DetectorDivergence    = "channel_divergence"
	DetectorPeerResidual  = "peer_residual"
	DetectorConservation  = "conservation_residual"
)

// Solution methods for the onset solver.
const (
	SolveAnalytic = "analytic"
	SolveNumeric  = "numeric"
)

// Epithet for the ground truth: "expected episode" is true for the positive
// class; a missing episode on a positive scenario is a mechanism miss.
func (g *GroundTruthRecord) IsPositive() bool {
	return !g.IsNegativeClass && g.ExpectedEpisode
}

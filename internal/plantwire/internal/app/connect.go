package app

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

var ErrPlanRefused = errors.New("plantwire: the session plan does not hold")

const (
	connectAdapter = "native-jsonl"
	maxPlanBytes   = 1 << 20
)

type ConnectOptions struct {
	Plan          io.Reader
	Source        string
	DomainsDir    string
	AdaptersDir   string
	KindsDirs     []string
	Out           string
	SliceNS       int64
	ControlSocket string
	BarrierCap    time.Duration
	Ready         func()
}

type ConnectCheck struct {
	Domain            string                    `json:"domain"`
	Environment       string                    `json:"environment"`
	Gateway           bool                      `json:"gateway"`
	Instances         []domain.PlannedInstance  `json:"instances"`
	ImplementsChecked []model.ImplementsChecked `json:"implements_checked"`
	Policies          []string                  `json:"policies"`
}

type ConnectResult struct {
	Out     string             `json:"out"`
	Stopped bool               `json:"stopped"`
	Relay   *domain.RelayStats `json:"relay,omitempty"`
}

type connectPlan struct {
	session *model.SessionPlan
	spec    *simdomain.Compiled
	adapter *model.Adapter
	plan    domain.ConnectPlan
	checked []model.ImplementsChecked
	pacer   *uds.WallPacer
	faces   domain.Interfaces
}

func CheckConnect(options ConnectOptions) (ConnectCheck, error) {
	planned, err := planConnect(options)
	if err != nil {
		return ConnectCheck{}, err
	}
	return planned.result(), nil
}

func (p *connectPlan) result() ConnectCheck {
	return ConnectCheck{Domain: p.spec.Spec.ID, Environment: p.plan.Environment, Gateway: p.plan.GatewaySocket != "", Instances: p.plan.Instances,
		ImplementsChecked: p.checked, Policies: p.plan.Policies.Digests()}
}

func planConnect(options ConnectOptions) (connectPlan, error) {
	session, err := readPlan(options.Plan)
	if err != nil {
		return connectPlan{}, fmt.Errorf("plantwire connect: %w", model.WithSource(err, options.Source))
	}
	planned, err := loadConnectInputs(options, session)
	if err != nil {
		return connectPlan{}, err
	}
	err = planned.check(options.Source)
	return planned, err
}

func (p *connectPlan) check(source string) error {
	plan, issues := domain.CheckPlan(p.session, planInstances(p.spec, p.checked))
	p.plan = plan
	if len(issues) > 0 {
		return planViolation(source, issues)
	}
	return nil
}

func readPlan(reader io.Reader) (*model.SessionPlan, error) {
	raw, err := io.ReadAll(io.LimitReader(reader, maxPlanBytes+1))
	switch {
	case err != nil:
		return nil, fmt.Errorf("read the session plan: %w", err)
	case len(raw) > maxPlanBytes:
		return nil, fmt.Errorf("%w: the session plan holds more than %d bytes", ErrBadOptions, maxPlanBytes)
	}
	plan, err := model.DecodeSessionPlan(raw)
	if err != nil {
		return nil, fmt.Errorf("decode the session plan: %w", err)
	}
	return plan, nil
}

func loadConnectInputs(options ConnectOptions, session *model.SessionPlan) (connectPlan, error) {
	spec, adapterSpec, err := run.LoadInputs(options.DomainsDir, options.AdaptersDir, session.Domain, connectAdapter)
	if err != nil {
		return connectPlan{}, fmt.Errorf("plantwire connect: %w", err)
	}
	checked, faces, err := loadInterfaces(spec.Spec, options.KindsDirs)
	if err != nil {
		return connectPlan{}, fmt.Errorf("plantwire connect: %w", model.WithSource(err, spec.Spec.ID))
	}
	return connectPlan{session: session, spec: spec, adapter: adapterSpec, checked: checked, faces: faces}, nil
}

func checkedByKind(checked []model.ImplementsChecked) map[string]model.ImplementsChecked {
	byKind := map[string]model.ImplementsChecked{}
	for _, entry := range checked {
		byKind[entry.Kind] = entry
	}
	return byKind
}

func planInstances(spec *simdomain.Compiled, checked []model.ImplementsChecked) []domain.PlanInstance {
	byKind := checkedByKind(checked)
	ids := world.PopulationIDs(spec, 0)
	instances := make([]domain.PlanInstance, len(ids))
	for i, id := range ids {
		kind := spec.InstanceKind(spec.InstanceSetFor(id))
		entry := byKind[kind.Implements]
		instances[i] = domain.PlanInstance{ID: id, Kind: kind.Name, Implements: kind.Implements, KindDigest: entry.KindDigest,
			BoardType: entry.BoardType, BoardDigest: entry.BoardDigest, Faults: planFaults(kind)}
	}
	return instances
}

func planFaults(kind *model.Kind) []domain.PlanFault {
	commanded := commandedStates(kind)
	faults := make([]domain.PlanFault, len(kind.Faults))
	for i, fault := range kind.Faults {
		faults[i] = domain.PlanFault{ID: fault.ID, Pulse: fault.Onset.Shape == model.ShapePulse,
			Commanded: slices.ContainsFunc(fault.Affects, func(e model.FaultEffect) bool { return commanded[e.State] })}
	}
	return faults
}

func commandedStates(kind *model.Kind) map[string]bool {
	states := map[string]bool{}
	for _, effector := range kind.Effectors {
		for _, delta := range effector.Effect.StateDeltas {
			states[delta.State] = true
		}
	}
	return states
}

func planViolation(source string, issues []model.Issue) error {
	return &model.ValidationError{Contract: model.ContractSessionPlan, Source: source, Issues: issues, Cause: ErrPlanRefused,
		Summary: fmt.Sprintf("plantwire connect: the session plan does not hold: %s", issues[0].Message)}
}

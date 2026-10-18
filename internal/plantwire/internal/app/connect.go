package app

import (
	"errors"
	"fmt"
	"io"
	"time"

	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

var ErrPlanRefused = errors.New("plantwire: the session plan does not hold")

const (
	connectAdapter = "native-jsonl"
	maxPlanBytes   = 1 << 20
)

type ConnectOptions struct {
	Plan        io.Reader
	Source      string
	DomainsDir  string
	AdaptersDir string
	KindsDirs   []string
	Policies    []string
	Out         string
	SliceNS     int64
	BarrierCap  time.Duration
	Ready       func()
}

type ConnectCheck struct {
	Domain            string                    `json:"domain"`
	Environment       string                    `json:"environment"`
	Instances         []domain.PlannedInstance  `json:"instances"`
	ImplementsChecked []model.ImplementsChecked `json:"implements_checked"`
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
}

func CheckConnect(options ConnectOptions) (ConnectCheck, error) {
	planned, err := planConnect(options)
	if err != nil {
		return ConnectCheck{}, err
	}
	return ConnectCheck{Domain: planned.spec.Spec.ID, Environment: planned.plan.Environment, Instances: planned.plan.Instances,
		ImplementsChecked: planned.checked}, nil
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
	checked, err := CheckImplements(spec.Spec, options.KindsDirs)
	if err != nil {
		return connectPlan{}, fmt.Errorf("plantwire connect: %w", model.WithSource(err, spec.Spec.ID))
	}
	return connectPlan{session: session, spec: spec, adapter: adapterSpec, checked: checked}, nil
}

func planInstances(spec *simdomain.Compiled, checked []model.ImplementsChecked) []domain.PlanInstance {
	digests := map[string]string{}
	for _, entry := range checked {
		digests[entry.Kind] = entry.KindDigest
	}
	ids := world.PopulationIDs(spec, 0)
	instances := make([]domain.PlanInstance, len(ids))
	for i, id := range ids {
		kind := spec.InstanceKind(spec.InstanceSetFor(id))
		instances[i] = domain.PlanInstance{ID: id, Kind: kind.Name, Implements: kind.Implements, KindDigest: digests[kind.Implements]}
	}
	return instances
}

func planViolation(source string, issues []model.Issue) error {
	return &model.ValidationError{Contract: model.ContractSessionPlan, Source: source, Issues: issues, Cause: ErrPlanRefused,
		Summary: fmt.Sprintf("plantwire connect: the session plan does not hold: %s", issues[0].Message)}
}

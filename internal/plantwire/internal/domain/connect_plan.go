package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	PlannedReal      = "real"
	PlannedSimulated = "simulated"
)

type PlanInstance struct {
	ID, Kind, Implements, KindDigest string
}

type PlannedInstance struct {
	Instance   string `json:"instance"`
	Origin     string `json:"origin"`
	Kind       string `json:"kind"`
	KindDigest string `json:"kind_digest,omitempty"`
	Board      string `json:"board,omitempty"`
	Device     string `json:"device,omitempty"`
}

type Bound struct {
	Index                                     int
	Instance, Board, Device, Kind, KindDigest string
}

type ConnectPlan struct {
	Environment string
	Instances   []PlannedInstance
	Bound       []Bound
}

type planCheck struct {
	known     map[string]PlanInstance
	instances map[string]bool
	devices   map[string]bool
	bound     []Bound
}

func CheckPlan(plan *model.SessionPlan, instances []PlanInstance) (ConnectPlan, []model.Issue) {
	check := planCheck{known: map[string]PlanInstance{}, instances: map[string]bool{}, devices: map[string]bool{}}
	for _, instance := range instances {
		check.known[instance.ID] = instance
	}
	issues := socketIssues(plan)
	for i, binding := range plan.Real {
		issues = append(issues, check.binding(i, binding)...)
	}
	return check.outcome(plan, instances), issues
}

func socketIssues(plan *model.SessionPlan) []model.Issue {
	switch {
	case len(plan.Real) > 0 && plan.GatewaySocket == "":
		return []model.Issue{model.RuleIssue("/real", "gateway_required", "gateway_socket",
			"a real board is met through a gateway: the plan needs gateway_socket")}
	case plan.GatewaySocket != "" && plan.GatewaySocket == plan.HostSocket:
		return []model.Issue{model.RuleIssue("/gateway_socket", "same_socket", "gateway_socket",
			"the gateway and the host need two different sockets")}
	}
	return nil
}

func (c *planCheck) binding(i int, b model.RealBinding) []model.Issue {
	at := fmt.Sprintf("/real/%d", i)
	var issues []model.Issue
	for _, found := range []func(string, model.RealBinding) (model.Issue, bool){c.instanceIssue, c.deviceIssue} {
		if issue, broken := found(at, b); broken {
			issues = append(issues, issue)
		}
	}
	if len(issues) == 0 {
		c.bind(i, b)
	}
	return issues
}

func (c *planCheck) deviceIssue(at string, b model.RealBinding) (model.Issue, bool) {
	device := b.Board + "/" + b.Device
	if c.devices[device] {
		return model.RuleIssue(at+"/device", "device_bound_twice", "device",
			fmt.Sprintf("device %q of board %q is already bound to another instance", clip(b.Device), clip(b.Board))), true
	}
	c.devices[device] = true
	return model.Issue{}, false
}

func (c *planCheck) instanceIssue(at string, b model.RealBinding) (model.Issue, bool) {
	instance, known := c.known[b.Instance]
	switch {
	case !known:
		return model.RuleIssue(at+"/instance", "unknown_instance", "instance", fmt.Sprintf("the plant has no instance %q", clip(b.Instance))), true
	case instance.Implements == "":
		return model.RuleIssue(at+"/instance", "no_kind_file", "instance", fmt.Sprintf(
			"%s is of kind %q, which implements no kind file: a real board can stand only for an instance whose interface is a kind file",
			clip(b.Instance), clip(instance.Kind))), true
	case c.instances[b.Instance]:
		return model.RuleIssue(at+"/instance", "instance_bound_twice", "instance", fmt.Sprintf("%s is bound to a real board twice", clip(b.Instance))), true
	}
	c.instances[b.Instance] = true
	return model.Issue{}, false
}

func (c *planCheck) bind(i int, b model.RealBinding) {
	instance := c.known[b.Instance]
	c.bound = append(c.bound, Bound{Index: i, Instance: b.Instance, Board: b.Board, Device: b.Device,
		Kind: instance.Implements, KindDigest: instance.KindDigest})
}

func (c *planCheck) outcome(plan *model.SessionPlan, instances []PlanInstance) ConnectPlan {
	environment := EnvironmentSimulated
	if plan.GatewaySocket != "" {
		environment = EnvironmentMixed
	}
	planned := make([]PlannedInstance, len(instances))
	for i, instance := range instances {
		planned[i] = c.planned(instance)
	}
	return ConnectPlan{Environment: environment, Instances: planned, Bound: c.bound}
}

func (c *planCheck) planned(instance PlanInstance) PlannedInstance {
	for _, b := range c.bound {
		if b.Instance == instance.ID {
			return PlannedInstance{Instance: instance.ID, Origin: PlannedReal, Kind: instance.Kind, KindDigest: b.KindDigest, Board: b.Board, Device: b.Device}
		}
	}
	return PlannedInstance{Instance: instance.ID, Origin: PlannedSimulated, Kind: instance.Kind}
}

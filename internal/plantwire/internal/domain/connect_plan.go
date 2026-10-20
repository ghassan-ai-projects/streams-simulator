package domain

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"unicode"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	PlannedReal      = "real"
	PlannedSimulated = "simulated"
)

type PlanInstance struct {
	ID, Kind, Implements, KindDigest string
	Faults                           []PlanFault
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
	Environment   string
	Instances     []PlannedInstance
	Bound         []Bound
	HostSocket    string
	GatewaySocket string
	Policies      AllowList
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
	issues = append(issues, check.disturbanceIssues(plan)...)
	allowed, listed := NewAllowList(plan.Policies, "/policies")
	connect := check.outcome(plan, instances)
	connect.Policies = allowed
	return connect, append(issues, listed...)
}

func socketIssues(plan *model.SessionPlan) []model.Issue {
	issues := socketPathIssues("/host_socket", "host_socket", plan.HostSocket)
	issues = append(issues, socketPathIssues("/gateway_socket", "gateway_socket", plan.GatewaySocket)...)
	return append(issues, socketPairIssues(plan)...)
}

func cleanSocket(socket string) string {
	if socket == "" {
		return ""
	}
	return path.Clean(socket)
}

func socketPathIssues(pointer, member, socket string) []model.Issue {
	switch {
	case socket == "":
		return nil
	case !path.IsAbs(socket):
		return []model.Issue{model.RuleIssue(pointer, "socket_not_absolute", member, "a socket is an absolute path")}
	case strings.IndexFunc(socket, unicode.IsControl) >= 0:
		return []model.Issue{model.RuleIssue(pointer, "socket_control_character", member, "a socket path holds no control character")}
	case slices.Contains(strings.Split(socket, "/"), ".."):
		return []model.Issue{model.RuleIssue(pointer, "socket_parent_component", member, "a socket path holds no .. component")}
	}
	return nil
}

func socketPairIssues(plan *model.SessionPlan) []model.Issue {
	switch {
	case len(plan.Real) > 0 && plan.GatewaySocket == "":
		return []model.Issue{model.RuleIssue("/real", "gateway_required", "gateway_socket",
			"a real board is met through a gateway: the plan needs gateway_socket")}
	case plan.GatewaySocket != "" && cleanSocket(plan.GatewaySocket) == cleanSocket(plan.HostSocket):
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
	environment := c.environment(plan, len(instances))
	planned := make([]PlannedInstance, len(instances))
	for i, instance := range instances {
		planned[i] = c.planned(instance)
	}
	return ConnectPlan{Environment: environment, Instances: planned, Bound: c.bound,
		HostSocket: cleanSocket(plan.HostSocket), GatewaySocket: cleanSocket(plan.GatewaySocket)}
}

func (c *planCheck) environment(plan *model.SessionPlan, instances int) string {
	if plan.GatewaySocket == "" {
		return EnvironmentSimulated
	}
	return AnnouncedEnvironment(map[string]bool{EnvironmentHardware: true, EnvironmentSimulated: len(c.bound) < instances})
}

func (c *planCheck) planned(instance PlanInstance) PlannedInstance {
	for _, b := range c.bound {
		if b.Instance == instance.ID {
			return PlannedInstance{Instance: instance.ID, Origin: PlannedReal, Kind: instance.Kind, KindDigest: b.KindDigest, Board: b.Board, Device: b.Device}
		}
	}
	return PlannedInstance{Instance: instance.ID, Origin: PlannedSimulated, Kind: instance.Kind}
}

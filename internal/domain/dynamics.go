package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"sort"
	"strings"
)

func rejectDynamicsCycles(spec *model.DomainSpec, c *Compiled, src string) error {
	deps := make(map[string][]string)
	for _, d := range spec.Dynamics {
		if d.F1 == nil {
			continue
		}
		for _, in := range d.F1.Inputs {
			deps[d.Target] = append(deps[d.Target], in.State)
		}
	}
	state := make(map[string]uint8)
	var stack []string
	var visit func(string) error
	visit = func(name string) error {
		switch state[name] {
		case 1:
			return fmt.Errorf("domain: %s: dynamics dependency cycle at state %q (path %v)", src, name, append(stack, name))
		case 2:
			return nil
		}
		state[name] = 1
		stack = append(stack, name)
		for _, dep := range deps[name] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		stack = stack[:len(stack)-1]
		state[name] = 2
		return nil
	}
	for _, name := range c.StateNames() {
		if err := visit(name); err != nil {
			return err
		}
	}
	return nil
}

func checkF1(f1 *model.F1Dyn, c *Compiled, src string) error {
	bad := func(format string, args ...any) error {
		return fmt.Errorf("domain: %s: %s", src, fmt.Sprintf(format, args...))
	}
	for _, in := range f1.Inputs {
		if !c.HasState(in.State) {
			return bad("f1 form %q input %q is not a declared state", f1.Form, in.State)
		}
	}
	switch f1.Form {
	case "first_order_lag", "dead_time":
		if f1.TimeConstantS == 0 {
			return bad("f1 form %q requires time_constant_s", f1.Form)
		}
	case "rc_network":
		if f1.TimeConstantS == 0 || f1.TimeConstant2S == 0 {
			return bad("rc_network requires time_constant_s and time_constant_2_s")
		}
	case "hysteresis":
		if f1.ThresholdLow >= f1.ThresholdHigh {
			return bad("hysteresis requires threshold_low < threshold_high")
		}
	case "threshold_integrator":
		if f1.Direction != "below" && f1.Direction != "above" {
			return bad("threshold_integrator requires direction below|above")
		}
	case "saturation":
		if f1.Clamp == nil || (f1.Clamp.Min == nil && f1.Clamp.Max == nil) {
			return bad("saturation requires a clamp interval")
		}
	}
	return nil
}

func checkDetector(d *model.Detector, c *Compiled, src, who string) error {
	bad := func(format string, args ...any) error {
		return fmt.Errorf("domain: %s: %s", src, fmt.Sprintf(format, args...))
	}
	switch d.Form {
	case model.DetectorSingleChannel, model.DetectorPeerResidual:
		if d.Channel == "" || !c.HasChannel(d.Channel) {
			return bad("%s: detector %s requires a declared channel", who, d.Form)
		}
	case model.DetectorDivergence:
		if d.ChannelA == "" || d.ChannelB == "" || !c.HasChannel(d.ChannelA) || !c.HasChannel(d.ChannelB) {
			return bad("%s: channel_divergence requires declared channel_a and channel_b", who)
		}
	case model.DetectorConservation:
		for _, ch := range append(append([]string{}, d.Inputs...), d.Outputs...) {
			if !c.HasChannel(ch) {
				return bad("%s: conservation_residual references undeclared channel %q", who, ch)
			}
		}
	default:
		return bad("%s: unknown detector form %q", who, d.Form)
	}
	return nil
}

func dynamicsFor(c *Compiled, state string) *model.Dynamics {
	for i := range c.Spec.Dynamics {
		if c.Spec.Dynamics[i].Target == state {
			return &c.Spec.Dynamics[i]
		}
	}
	return nil
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func formatErrs(errs []jsonschema.Error) string {
	var b strings.Builder
	for i, e := range errs {
		if i == 10 {
			fmt.Fprintf(&b, "  ... and %d more\n", len(errs)-10)
			break
		}
		fmt.Fprintf(&b, "  %s\n", e.Error())
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func mustAny(b []byte) any {
	var v any
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		panic(err)
	}
	return v
}

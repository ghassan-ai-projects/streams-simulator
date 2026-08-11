// Package domain loads, validates and compiles domain specs — the data
// files that define whole simulated worlds. The binary contains no domain
// behavior: every domain in the catalog loads through this one path, and a
// domain that needs a code branch in the binary is a bug in the simulator's
// design, not a missing feature.
package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

// Load reads, schema-validates and structurally validates a domain spec from
// a path. The returned Compiled carries the spec, its canonical digest, and
// resolved defaults.
func Load(path string) (*Compiled, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("domain: read %s: %w", path, err)
	}
	return Parse(raw, path)
}

// Parse is Load over an in-memory document. src names the source for error
// messages.
func Parse(raw []byte, src string) (*Compiled, error) {
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return nil, fmt.Errorf("domain: %s: not valid JSON: %w", src, err)
	}
	sch, err := jsonschema.Compile(mustAny(schemas.DomainSpec()))
	if err != nil {
		return nil, fmt.Errorf("domain: compile contract schema: %w", err)
	}
	if errs := sch.Validate(doc); len(errs) > 0 {
		return nil, fmt.Errorf("domain: %s fails domain-spec-v0.1 validation:\n  %s", src, formatErrs(errs))
	}
	var spec model.DomainSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("domain: %s: decode: %w", src, err)
	}
	digest, err := canonical.Digest(doc)
	if err != nil {
		return nil, fmt.Errorf("domain: %s: digest: %w", src, err)
	}
	c, err := compile(&spec, digest, src)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	c.Raw = append([]byte(nil), raw...)
	return c, nil
}

// LoadAll loads every domain spec in a directory (non-recursive), sorted by
// id. A single unloadable domain fails the whole load: the catalog is a
// fixed, reviewed set, and a silent skip would make coverage reports lie.
func LoadAll(dir string) ([]*Compiled, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("domain: list %s: %w", dir, err)
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
	}
	sort.Strings(paths)
	var out []*Compiled
	for _, p := range paths {
		c, err := Load(p)
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		out = append(out, c)
	}
	return out, nil
}

// Compiled is the typed, validated, digest-carrying form of a domain spec.
type Compiled struct {
	Spec   *model.DomainSpec
	Digest string
	// Raw is the validated source document embedded in run artifacts.
	Raw []byte

	states    map[string]bool
	channels  map[string]bool
	faults    map[string]bool
	effectors map[string]bool
	profiles  map[string]bool

	// Resolved defaults (schema defaults applied).
	channelGain map[string]float64 // channel -> observation_gain
}

// HasState reports whether the state name is declared.
func (c *Compiled) HasState(name string) bool { return c.states[name] }

// HasChannel reports whether the channel name is declared.
func (c *Compiled) HasChannel(name string) bool { return c.channels[name] }

// HasFault reports whether the fault id is declared.
func (c *Compiled) HasFault(id string) bool { return c.faults[id] }

// HasEffector reports whether the effector name is declared.
func (c *Compiled) HasEffector(name string) bool { return c.effectors[name] }

// HasProfile reports whether the profile name is declared.
func (c *Compiled) HasProfile(name string) bool { return c.profiles[name] }

// ChannelGain returns the resolved observation gain for a channel.
func (c *Compiled) ChannelGain(ch string) float64 {
	if g, ok := c.channelGain[ch]; ok {
		return g
	}
	return 1
}

// StateNames returns the declared state names, sorted.
func (c *Compiled) StateNames() []string { return sortedKeys(c.states) }

// ChannelNames returns the declared channel names, sorted.
func (c *Compiled) ChannelNames() []string { return sortedKeys(c.channels) }

// FaultNames returns the declared fault ids, sorted.
func (c *Compiled) FaultNames() []string { return sortedKeys(c.faults) }

// EffectorNames returns the declared effector names, sorted.
func (c *Compiled) EffectorNames() []string { return sortedKeys(c.effectors) }

// ProfileNames returns the declared profile names, sorted.
func (c *Compiled) ProfileNames() []string { return sortedKeys(c.profiles) }

// Fault returns the fault by id, or nil.
func (c *Compiled) Fault(id string) *model.Fault {
	for i := range c.Spec.Faults {
		if c.Spec.Faults[i].ID == id {
			return &c.Spec.Faults[i]
		}
	}
	return nil
}

// Effector returns the effector by name, or nil.
func (c *Compiled) Effector(name string) *model.Effector {
	for i := range c.Spec.Effectors {
		if c.Spec.Effectors[i].Name == name {
			return &c.Spec.Effectors[i]
		}
	}
	return nil
}

// Channel returns the channel by name, or nil.
func (c *Compiled) Channel(name string) *model.Channel {
	for i := range c.Spec.Channels {
		if c.Spec.Channels[i].Name == name {
			return &c.Spec.Channels[i]
		}
	}
	return nil
}

// Profile returns the profile by name, or nil.
func (c *Compiled) Profile(name string) *model.Profile {
	for i := range c.Spec.Profiles {
		if c.Spec.Profiles[i].Name == name {
			return &c.Spec.Profiles[i]
		}
	}
	return nil
}

func compile(spec *model.DomainSpec, digest, src string) (*Compiled, error) {
	c := &Compiled{
		Spec:        spec,
		Digest:      digest,
		states:      map[string]bool{},
		channels:    map[string]bool{},
		faults:      map[string]bool{},
		effectors:   map[string]bool{},
		profiles:    map[string]bool{},
		channelGain: map[string]float64{},
	}
	for i := range spec.State {
		s := &spec.State[i]
		if c.states[s.Name] {
			return nil, fmt.Errorf("domain: %s: duplicate state %q", src, s.Name)
		}
		c.states[s.Name] = true
	}
	for i := range spec.Channels {
		ch := &spec.Channels[i]
		if c.channels[ch.Name] {
			return nil, fmt.Errorf("domain: %s: duplicate channel %q", src, ch.Name)
		}
		c.channels[ch.Name] = true
		gain := ch.ObservationGain
		if gain == 0 {
			gain = 1
		}
		c.channelGain[ch.Name] = gain
	}
	for i := range spec.Faults {
		f := &spec.Faults[i]
		if c.faults[f.ID] {
			return nil, fmt.Errorf("domain: %s: duplicate fault %q", src, f.ID)
		}
		c.faults[f.ID] = true
	}
	for i := range spec.Effectors {
		e := &spec.Effectors[i]
		if c.effectors[e.Name] {
			return nil, fmt.Errorf("domain: %s: duplicate effector %q", src, e.Name)
		}
		c.effectors[e.Name] = true
	}
	for i := range spec.Profiles {
		p := &spec.Profiles[i]
		if c.profiles[p.Name] {
			return nil, fmt.Errorf("domain: %s: duplicate profile %q", src, p.Name)
		}
		c.profiles[p.Name] = true
	}
	if err := crossCheck(c, src); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return c, nil
}

// crossCheck validates every cross-reference in the spec: channels observe
// declared states, faults affect declared states, effectors touch declared
// states, detector expressions name declared channels, and so on. The JSON
// schema cannot express these.
func crossCheck(c *Compiled, src string) error {
	spec := c.Spec
	bad := func(format string, args ...any) error {
		return fmt.Errorf("domain: %s: %s", src, fmt.Sprintf(format, args...))
	}
	for i := range spec.Channels {
		ch := &spec.Channels[i]
		if ch.Observes != "" && !c.HasState(ch.Observes) {
			return bad("channel %q observes undeclared state %q", ch.Name, ch.Observes)
		}
		if ch.ObservationBias != nil && !c.HasState(ch.ObservationBias.State) {
			return bad("channel %q bias references undeclared state %q", ch.Name, ch.ObservationBias.State)
		}
		if ch.Cadence.Mode == "event_driven" && ch.Cadence.TriggerState != "" && !c.HasState(ch.Cadence.TriggerState) {
			return bad("channel %q trigger references undeclared state %q", ch.Name, ch.Cadence.TriggerState)
		}
		if ch.Cadence.Mode == "batch" {
			return bad("channel %q uses batch cadence, which is not implemented", ch.Name)
		}
		if ch.Noise.Model == "pink" {
			return bad("channel %q uses pink noise, which is not implemented", ch.Name)
		}
		if ch.Noise.Model == "none" && ch.Noise.Sigma > 0 {
			// Fail closed: a "no noise" declaration with a nonzero sigma
			// would otherwise be silently approximated by the gaussian
			// branch in observe.go.
			return bad("channel %q declares noise model none with sigma %v; sigma must be 0", ch.Name, ch.Noise.Sigma)
		}
		if ch.Availability != nil && ch.Availability.MTTRS <= 0 {
			return bad("channel %q availability requires mttr_s > 0", ch.Name)
		}
	}
	for i := range spec.Dynamics {
		d := &spec.Dynamics[i]
		if !c.HasState(d.Target) {
			return bad("dynamics target %q is not a declared state", d.Target)
		}
		if d.Tier == "F3" {
			return fmt.Errorf("domain: %s: F3/FMU dynamics are the specified-but-unbuilt seam; not_implemented", src)
		}
		if d.Tier == "F2" {
			// F2 reference models are not built. The integrator must fail
			// loudly rather than silently leave the state static: a domain
			// that used F2 would emit confident wrong values.
			return fmt.Errorf("domain: %s: F2 reference models are the specified-but-unbuilt seam; not_implemented", src)
		}
		if d.F1 != nil {
			if err := checkF1(d.F1, c, src); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
		if d.F2 != nil {
			for name := range d.F2.Inputs {
				if !c.HasState(name) {
					return bad("f2 model %q input %q is not a declared state", d.F2.Model, name)
				}
			}
		}
	}
	if err := rejectDynamicsCycles(spec, c, src); err != nil {
		return err
	}
	for i := range spec.Faults {
		f := &spec.Faults[i]
		if f.Onset.RatePerHour < 0 {
			return bad("fault %q onset rate_per_hour must be non-negative", f.ID)
		}
		for _, a := range f.Affects {
			if !c.HasState(a.State) {
				return bad("fault %q affects undeclared state %q", f.ID, a.State)
			}
		}
		if err := checkDetector(&f.Observability.Detector, c, src, "fault "+f.ID); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		if f.ExpectedEffector != "" && !c.HasEffector(f.ExpectedEffector) {
			return bad("fault %q expects undeclared effector %q", f.ID, f.ExpectedEffector)
		}
	}
	for i := range spec.Effectors {
		e := &spec.Effectors[i]
		for _, d := range e.Effect.StateDeltas {
			if !c.HasState(d.State) {
				return bad("effector %q effect targets undeclared state %q", e.Name, d.State)
			}
		}
		if e.Interlock != nil && !c.HasState(e.Interlock.State) {
			return bad("effector %q interlock references undeclared state %q", e.Name, e.Interlock.State)
		}
		if e.Interlock != nil && e.Interlock.AutonomousAction != nil {
			if !c.HasState(e.Interlock.AutonomousAction.State) {
				return bad("effector %q autonomous action targets undeclared state %q", e.Name, e.Interlock.AutonomousAction.State)
			}
		}
		for _, ch := range e.ConfirmationChannels {
			if !c.HasChannel(ch) {
				return bad("effector %q confirmation channel %q is not declared", e.Name, ch)
			}
		}
		// The dead time is modeled once: an effector that declares its own
		// dead_time_s must not drive a dead_time state, or the delay applies
		// twice.
		if e.Effect.DeadTimeS > 0 {
			for _, d := range e.Effect.StateDeltas {
				if dyn := dynamicsFor(c, d.State); dyn != nil && dyn.F1 != nil && dyn.F1.Form == "dead_time" {
					return bad("effector %q has dead_time_s and drives dead_time state %q: delay would apply twice", e.Name, d.State)
				}
			}
		}
	}
	for i := range spec.Profiles {
		p := &spec.Profiles[i]
		for name, w := range p.FaultWeights {
			if !c.HasFault(name) {
				return bad("profile %q weights undeclared fault %q", p.Name, name)
			}
			if w < 0 {
				return bad("profile %q has negative weight for %q", p.Name, name)
			}
		}
	}
	return nil
}

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

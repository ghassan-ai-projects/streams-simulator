// Package host is the plant wire reference host: handshake checks, inventory, ordering, scripted commands and verification.
package host

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

type Options struct {
	Environment string
	ProtocolMax int
	Epoch       int64
	Policy      string
	Inventory   map[string]string
	Script      []Scripted
	Ready       func()
}

type Scripted struct {
	AtNS    int64          `json:"at_ns"`
	Device  string         `json:"device"`
	Command string         `json:"command"`
	Args    map[string]any `json:"args"`
	Verify  *Verify        `json:"verify,omitempty"`
}

type Verify struct {
	Metric    string  `json:"metric"`
	Equals    float64 `json:"equals"`
	Tolerance float64 `json:"tolerance"`
	WithinNS  int64   `json:"within_ns"`
}

type Report struct {
	Refused     string             `json:"refused,omitempty"`
	Quarantined []string           `json:"quarantined"`
	Rebirths    int                `json:"rebirths"`
	Commands    map[string]*Status `json:"commands"`
}

type Status struct {
	Receipts []string `json:"receipts"`
	Result   string   `json:"result,omitempty"`
	Denied   string   `json:"denied,omitempty"`
	Verified *bool    `json:"verified,omitempty"`
}

type Host struct {
	options  Options
	out      *domain.Transmitter
	in       domain.Receiver
	devices  map[string]*device
	script   []Scripted
	issued   map[string]Scripted
	issuedAt map[string]int64
	order    []string
	resultAt map[string]int64
	samples  int64
	report   Report
	asked    bool
	accepted bool
}

type device struct {
	boot    string
	aliases map[int64]string
	values  map[string]float64
	updated map[string]int64
}

func New(options Options) *Host {
	script := slices.Clone(options.Script)
	slices.SortStableFunc(script, func(a, b Scripted) int { return cmp.Compare(a.AtNS, b.AtNS) })
	return &Host{options: options, out: domain.NewTransmitter(""), devices: map[string]*device{}, script: script,
		issued: map[string]Scripted{}, issuedAt: map[string]int64{}, resultAt: map[string]int64{},
		report: Report{Quarantined: []string{}, Commands: map[string]*Status{}}}
}

func (h *Host) Report() Report { return h.report }

func ParseScript(raw []byte) ([]Scripted, error) {
	var script []Scripted
	if err := decodeStrict(raw, &script); err != nil {
		return nil, fmt.Errorf("plantwire host script: %w", err)
	}
	for i := range script {
		if err := checkStep(&script[i]); err != nil {
			return nil, fmt.Errorf("plantwire host script: step %d: %w", i, err)
		}
	}
	return script, nil
}

var ErrInvalidStep = errors.New("a step needs a device, a command, and a verification with a positive window and no negative tolerance")

func checkStep(step *Scripted) error {
	if step.Args == nil {
		step.Args = map[string]any{}
	}
	if step.Device == "" || step.Command == "" || (step.Verify != nil && (step.Verify.WithinNS <= 0 || step.Verify.Tolerance < 0)) {
		return ErrInvalidStep
	}
	return nil
}

func ParseInventory(raw []byte) (map[string]string, error) {
	var inventory map[string]string
	if err := decodeStrict(raw, &inventory); err != nil {
		return nil, fmt.Errorf("plantwire host inventory: %w", err)
	}
	return inventory, nil
}

func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

// Package domain holds the component library rules: loading and checking
// equipment, instruments and vocabulary, and composing a plant into an
// ordinary domain. It reads no file; bytes arrive from the files edge.
package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

type Range struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type MeasurementPoint struct {
	Name              string `json:"name"`
	State             string `json:"state"`
	UnitFamily        string `json:"unit_family"`
	Range             Range  `json:"range"`
	Tag               string `json:"tag"`
	DefaultInstrument string `json:"default_instrument"`
	Optional          bool   `json:"optional"`
}

type Port struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Direction string `json:"direction"`
	State     string `json:"state"`
}

type Preset struct {
	Title  string         `json:"title"`
	Values map[string]any `json:"values"`
}

type Parameter struct {
	Name    string                    `json:"name"`
	Title   string                    `json:"title"`
	Unit    string                    `json:"unit"`
	Type    string                    `json:"type"`
	Min     *float64                  `json:"min"`
	Max     *float64                  `json:"max"`
	Values  []string                  `json:"values"`
	Default any                       `json:"default"`
	Targets []string                  `json:"targets"`
	Cases   map[string]map[string]any `json:"cases"`
}

type Equipment struct {
	ID                string             `json:"id"`
	Version           int                `json:"version"`
	Title             string             `json:"title"`
	Class             string             `json:"class"`
	Tags              []string           `json:"tags"`
	MeasurementPoints []MeasurementPoint `json:"measurement_points"`
	Ports             []Port             `json:"ports"`
	DefaultPreset     string             `json:"default_preset"`
	Presets           map[string]Preset  `json:"presets"`
	Parameters        []Parameter        `json:"parameters"`
	Document          map[string]any     `json:"-"`
	Digest            string             `json:"-"`
	File              string             `json:"-"`
}

type InstrumentFault struct {
	ID              string  `json:"id"`
	Form            string  `json:"form"`
	Magnitude       float64 `json:"magnitude"`
	DurationS       float64 `json:"duration_s"`
	IsNegativeClass bool    `json:"is_negative_class"`
	StuckAt         string  `json:"stuck_at"`
	Factor          float64 `json:"factor"`
	Quality         string  `json:"quality"`
}

type Instrument struct {
	ID         string            `json:"id"`
	Version    int               `json:"version"`
	Title      string            `json:"title"`
	Tag        string            `json:"tag"`
	UnitFamily string            `json:"unit_family"`
	Unit       string            `json:"unit"`
	Range      *Range            `json:"range"`
	Channel    map[string]any    `json:"channel"`
	Faults     []InstrumentFault `json:"faults"`
	Scaling    string            `json:"scaling"`
	Document   map[string]any    `json:"-"`
	Digest     string            `json:"-"`
	File       string            `json:"-"`
}

type InstrumentChanges struct {
	Replace map[string]string `json:"replace"`
	Remove  []string          `json:"remove"`
	Add     []string          `json:"add"`
}

type Entry struct {
	Use         string            `json:"use"`
	As          string            `json:"as"`
	Preset      string            `json:"preset"`
	Parameters  map[string]any    `json:"parameters"`
	Names       []string          `json:"names"`
	IDTemplate  string            `json:"id_template"`
	Count       map[string]any    `json:"count"`
	Area        string            `json:"area"`
	Instruments InstrumentChanges `json:"instruments"`
}

type Connection struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Composition struct {
	ID          string         `json:"id"`
	Version     string         `json:"version"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Stresses    string         `json:"stresses"`
	Axes        map[string]any `json:"axes"`
	Profiles    []any          `json:"profiles"`
	GroundTruth map[string]any `json:"ground_truth"`
	Site        *model.Site    `json:"site"`
	Areas       []model.Area   `json:"areas"`
	Equipment   []Entry        `json:"equipment"`
	Connections []Connection   `json:"connections"`
}

type Kind string

const (
	KindEquipment  Kind = "equipment"
	KindInstrument Kind = "instrument"
)

type Summary struct {
	ID      string   `json:"id"`
	Version int      `json:"version"`
	Kind    Kind     `json:"kind"`
	Title   string   `json:"title"`
	Class   string   `json:"class"`
	Tags    []string `json:"tags"`
	Digest  string   `json:"digest"`
}

type Detail struct {
	Entry    Summary        `json:"entry"`
	Document map[string]any `json:"document"`
}

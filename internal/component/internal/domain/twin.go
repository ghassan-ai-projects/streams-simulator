package domain

type twinBuild struct {
	lib     *Library
	source  string
	choices *twinChoices
	kinds   []*twinKind
	byName  map[string]*twinKind
}

type twinKind struct {
	index    int
	choice   kindChoice
	file     *kindFile
	metrics  []*metricPlan
	commands []*commandPlan
	actuated map[string]bool
}

type metricPlan struct {
	file       kindMetric
	choice     metricChoice
	instrument *Instrument
	driven     []coupling
	tauS       float64
}

type commandPlan struct {
	file   kindCommand
	choice commandChoice
	at     string
}

type coupling struct {
	driver *twinKind
	state  string
	gain   float64
}

func ScaffoldTwin(lib *Library, files map[string]map[string]any, raw []byte, source string) (*Composed, error) {
	choices, err := parseTwinChoices(lib.contracts.twin, raw, source)
	if err != nil {
		return nil, err
	}
	t := &twinBuild{lib: lib, source: source, choices: choices, byName: map[string]*twinKind{}}
	return t.compose(files)
}

func (t *twinBuild) compose(files map[string]map[string]any) (*Composed, error) {
	if err := t.resolve(files); err != nil {
		return nil, err
	}
	composed, err := render(t.document(), t.source)
	if err != nil {
		return nil, err
	}
	composed.Implements = t.implemented()
	return composed, nil
}

func (t *twinBuild) resolve(files map[string]map[string]any) error {
	if err := t.loadKinds(files); err != nil {
		return err
	}
	for _, step := range []func() error{t.checkInstances, t.planKinds, t.planDrives} {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (t *twinBuild) implemented() []string {
	names := make([]string, len(t.kinds))
	for i, k := range t.kinds {
		names[i] = k.file.Name
	}
	return names
}

func (f *kindFile) metricNames() []string {
	names := make([]string, len(f.Metrics))
	for i, metric := range f.Metrics {
		names[i] = metric.Name
	}
	return names
}

func (f *kindFile) askedCommands() []string {
	var names []string
	for _, command := range f.Commands {
		if !command.SafeStop {
			names = append(names, command.Name)
		}
	}
	return names
}

func (f *kindFile) metric(name string) (kindMetric, bool) {
	for _, metric := range f.Metrics {
		if metric.Name == name {
			return metric, true
		}
	}
	return kindMetric{}, false
}

func (k *twinKind) plan(name string) *metricPlan {
	for _, plan := range k.metrics {
		if plan.file.Name == name {
			return plan
		}
	}
	return nil
}

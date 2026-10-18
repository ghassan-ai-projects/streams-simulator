package domain

const (
	responseDirect     = "direct"
	responseFirstOrder = "first_order"
	responseNone       = "none"
	instrumentNone     = "none"
)

type twinChoices struct {
	ID          string       `json:"id"`
	Version     string       `json:"version"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Site        *siteChoice  `json:"site"`
	Board       *boardChoice `json:"board"`
	Kinds       []kindChoice `json:"kinds"`
}

type siteChoice struct {
	Name string `json:"name"`
	Area string `json:"area"`
}

type boardChoice struct {
	Name string `json:"name"`
}

type kindChoice struct {
	Kind     string                   `json:"kind"`
	Instance string                   `json:"instance"`
	Metrics  map[string]metricChoice  `json:"metrics"`
	Commands map[string]commandChoice `json:"commands"`
}

type metricChoice struct {
	Instrument string   `json:"instrument"`
	Initial    *float64 `json:"initial"`
}

type commandChoice struct {
	Response string            `json:"response"`
	Sets     map[string]string `json:"sets"`
	Drives   []driveChoice     `json:"drives"`
}

type driveChoice struct {
	From          string  `json:"from"`
	Kind          string  `json:"kind"`
	Metric        string  `json:"metric"`
	Gain          float64 `json:"gain"`
	TimeConstantS float64 `json:"time_constant_s"`
}

func parseTwinChoices(c contractSchema, raw []byte, source string) (*twinChoices, error) {
	parsed, err := parseFile[twinChoices](c, raw, source)
	if err != nil {
		return nil, err
	}
	return parsed.value, nil
}

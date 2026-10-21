package domain

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func boardSource(t *testing.T, path string) KindSource {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return KindSource{Path: path, Raw: raw}
}

func boardFixtures(t *testing.T, folder string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", folder, "BOARD*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no BOARD fixtures in %s: %v", folder, err)
	}
	return files
}

func TestTheBoardDigestIsTheDigestOfTheWholeFile(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, raw, want string }{
		{"empty mapping", `{"board_type": "b", "kinds": {}}`, "sha256:a51d34d87b6a08c93bbc8bb7703b28da014ae39886c905dd83e4c94f198e9830"},
		{"valid/BOARD.json", fixtureText(t, "valid", "BOARD.json"), "sha256:045620b117bbd9ec7bdfc0ea16893f4878597b2b510a7033cfdec325b34afc4e"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			mapping, err := ParseBoardMapping(KindSource{Path: c.name, Raw: []byte(c.raw)})
			if err != nil || mapping.Digest != c.want {
				t.Fatalf("board digest %q, %v, want %s", mapping.Digest, err, c.want)
			}
		})
	}
}

func fixtureText(t *testing.T, folder, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", folder, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestTheThermalBoardMappingGivesTheBoardValuesOfEveryWord(t *testing.T) {
	t.Parallel()
	path := filepath.Join(testsupport.ThermalBoardKindsDir(), "arduino-thermal.board.json")
	files, summaries, err := ParseBoardFiles([]KindSource{boardSource(t, path)})
	if err != nil {
		t.Fatal(err)
	}
	want := BoardSummary{Path: path, BoardType: "arduino-thermal", Kinds: []string{"fan", "indicator"},
		BoardDigest: "sha256:25c7a642d0d7af2ca692e080ddf4314b5aeeaca86adcad2c8de9a34426c08582"}
	if got := summaries[0]; got.Path != want.Path || got.BoardType != want.BoardType || got.BoardDigest != want.BoardDigest || !slices.Equal(got.Kinds, want.Kinds) {
		t.Fatalf("summary %+v, want %+v", got, want)
	}
	cooling := files["arduino-thermal"].Kinds["fan"].Commands["set_cooling"]
	if cooling.Raw != "set_pwm_lease" || cooling.Args["duty_permille"].From != "level_percent" || cooling.Args["duty_permille"].Scale != "10" ||
		cooling.Lease == nil || *cooling.Lease != (MappedLease{Arg: "lease_ms", MS: 5000}) {
		t.Errorf("set_cooling maps to %+v, want set_pwm_lease with duty 10 x level_percent and a 5000 ms lease_ms", cooling)
	}
	show := files["arduino-thermal"].Kinds["indicator"].Commands["show_state"]
	rows := map[string][2]any{"off": {"0", "off"}, "watch": {"500", "slow_blink"}, "alert": {"1000", "solid"}}
	for word, values := range rows {
		row := show.Table[word]
		if show.By != "state" || numberText(row["brightness_permille"]) != values[0] || row["pattern"] != values[1] {
			t.Errorf("show_state %s maps to %v, want brightness %v pattern %v", word, row, values[0], values[1])
		}
	}
	if state := files["arduino-thermal"].Kinds["indicator"].Metrics["state"]; state.Raw != "pattern" || state.Table["slow_blink"] != "watch" ||
		!slices.Equal(files["arduino-thermal"].Kinds["indicator"].RawEnums["pattern"], []string{"off", "slow_blink", "solid"}) {
		t.Errorf("the state metric reads %+v with raw words %v", state, files["arduino-thermal"].Kinds["indicator"].RawEnums)
	}
}

func numberText(value any) string {
	if number, ok := value.(interface{ String() string }); ok {
		return number.String()
	}
	return ""
}

func TestEveryInvalidBoardFixtureFailsItsSchemaAtTheMemberItNames(t *testing.T) {
	t.Parallel()
	for _, file := range boardFixtures(t, "invalid") {
		var fixture struct {
			FailsAt string         `json:"fails_at"`
			Record  map[string]any `json:"record"`
		}
		if err := decode([]byte(fixtureText(t, "invalid", filepath.Base(file))), &fixture); err != nil {
			t.Fatal(err)
		}
		raw, _ := marshalCanonical(t, fixture.Record)
		_, err := ParseBoardMapping(KindSource{Path: file, Raw: raw})
		assertBoardIssue(t, filepath.Base(file), err, "", fixture.FailsAt)
	}
}

func TestEveryBoardRuleFixtureFailsWithItsKeywordAtTheMemberItNames(t *testing.T) {
	t.Parallel()
	for _, file := range boardFixtures(t, "rules") {
		var fixture struct {
			Keyword  string         `json:"keyword"`
			FailsAt  string         `json:"fails_at"`
			Document map[string]any `json:"document"`
		}
		if err := decode([]byte(fixtureText(t, "rules", filepath.Base(file))), &fixture); err != nil {
			t.Fatal(err)
		}
		raw, _ := marshalCanonical(t, fixture.Document)
		_, err := ParseBoardMapping(KindSource{Path: file, Raw: raw})
		assertBoardIssue(t, filepath.Base(file), err, fixture.Keyword, fixture.FailsAt)
		if !errors.Is(err, ErrBoardRule) {
			t.Errorf("%s: %v, want ErrBoardRule", filepath.Base(file), err)
		}
	}
}

func assertBoardIssue(t *testing.T, name string, err error, keyword, pointer string) {
	t.Helper()
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Contract != model.ContractBoardFile {
		t.Fatalf("%s: %v, want a board mapping validation error", name, err)
	}
	if !slices.ContainsFunc(invalid.Issues, func(issue model.Issue) bool {
		return strings.HasPrefix(issue.Pointer, pointer) && (keyword == "" || issue.Keyword == keyword)
	}) {
		t.Errorf("%s: issues %v, want %q at %s", name, invalid.Issues, keyword, pointer)
	}
}

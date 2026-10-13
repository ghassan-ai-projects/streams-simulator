package contract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type library struct {
	files map[string]map[string]any
	vocab map[string]any
}

func loadLibrary(t *testing.T) library {
	t.Helper()
	lib := library{files: map[string]map[string]any{}, vocab: readJSON(t, filepath.Join(componentsDir(), "vocabulary.json")).(map[string]any)}
	for _, pattern := range []string{"equipment/*.component.json", "instruments/*.instrument.json"} {
		paths, _ := filepath.Glob(filepath.Join(componentsDir(), pattern))
		for _, path := range paths {
			file := readJSON(t, path).(map[string]any)
			lib.files[file["id"].(string)+"@"+jsonString(file["version"])] = file
		}
	}
	return lib
}

func jsonString(v any) string {
	raw, _ := canonical.MarshalString(v)
	return raw
}

func (l library) has(section, name string) bool {
	_, ok := l.vocab[section].(map[string]any)[name]
	return ok
}

func TestTheVocabularyIsConsistent(t *testing.T) {
	t.Parallel()
	lib := loadLibrary(t)
	for family, raw := range lib.vocab["unit_families"].(map[string]any) {
		def := raw.(map[string]any)
		base, ok := def["units"].(map[string]any)[def["base"].(string)].(map[string]any)
		if !ok || jsonString(base["scale"]) != "1" || (base["offset"] != nil && jsonString(base["offset"]) != "0") {
			t.Errorf("%s: the base unit must be a unit with scale 1 and no offset", family)
		}
	}
	for port, raw := range lib.vocab["port_types"].(map[string]any) {
		if !lib.has("unit_families", raw.(map[string]any)["unit_family"].(string)) {
			t.Errorf("port type %s names an unknown family", port)
		}
	}
	for tag, raw := range lib.vocab["tag_letters"].(map[string]any) {
		for _, family := range raw.(map[string]any)["measures"].([]any) {
			if !lib.has("unit_families", family.(string)) {
				t.Errorf("tag %s measures an unknown family %s", tag, family)
			}
		}
	}
}

func TestEveryMeasurementPointHasAMatchingDefaultInstrument(t *testing.T) {
	t.Parallel()
	lib := loadLibrary(t)
	for id, file := range lib.files {
		points, _ := file["measurement_points"].([]any)
		for _, raw := range points {
			point := raw.(map[string]any)
			instrument := lib.files[point["default_instrument"].(string)]
			switch {
			case instrument == nil:
				t.Errorf("%s.%s: unknown instrument %s", id, point["name"], point["default_instrument"])
			case instrument["tag"] != point["tag"] || instrument["unit_family"] != point["unit_family"]:
				t.Errorf("%s.%s: instrument %s measures %s/%s, the point %s/%s", id, point["name"], point["default_instrument"],
					instrument["tag"], instrument["unit_family"], point["tag"], point["unit_family"])
			case !lib.has("tag_letters", point["tag"].(string)) || !lib.has("unit_families", point["unit_family"].(string)):
				t.Errorf("%s.%s: tag or family absent from the vocabulary", id, point["name"])
			}
		}
	}
}

func TestThePilotRecordsTheDigestsOfTheFilesItCameFrom(t *testing.T) {
	t.Parallel()
	lib := loadLibrary(t)
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "docs", "examples", "chiller-pilot.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec model.DomainSpec
	if err := model.DecodeBytes(raw, &spec); err != nil {
		t.Fatal(err)
	}
	for _, set := range spec.Instances {
		want := map[string]string{set.Provenance.Component + "@" + jsonString(set.Provenance.Version): set.Provenance.Digest}
		for ref, digest := range set.Provenance.InstrumentDigests {
			want[ref] = digest
		}
		for ref, digest := range want {
			if got, _ := canonical.Digest(lib.files[ref]); got != digest {
				t.Errorf("%s: provenance records %s for %s, the file is %s", set.Kind, digest, ref, got)
			}
		}
	}
}

func TestComposedInstrumentFaultsKeepTheChannelNoise(t *testing.T) {
	t.Parallel()
	spec := readJSON(t, filepath.Join(testsupport.RepositoryRoot(), "docs", "examples", "chiller-pilot.domain.json")).(map[string]any)
	for _, kind := range spec["kinds"].([]any) {
		for _, raw := range kind.(map[string]any)["channels"].([]any) {
			bias, ok := raw.(map[string]any)["observation_bias"].(map[string]any)
			if ok && jsonString(bias["noise_scale_at_full"]) != "1" {
				t.Errorf("%s: an instrument fault would scale the noise to %v", raw.(map[string]any)["name"], bias["noise_scale_at_full"])
			}
		}
	}
}

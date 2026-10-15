package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const basePlant = `{"schema_version":"0.1","id":"base","version":"1.0.0","title":"Base","description":"old",
"stresses":"a base plant that the new one starts from","axes":{},"ground_truth":{},
"equipment":[{"use":"lib/pump@1","names":["P-1"]}],"connections":[],
"profiles":[{"name":"nominal","duration_s":60,"fault_weights":{"p__f":1}},{"name":"x","not_applicable":"this family is fiction in the base plant"}]}`

func TestAnEmptyPlantKeepsItsFrameAndDropsWhatNamedTheOldEquipment(t *testing.T) {
	t.Parallel()
	doc, err := NewComposition([]byte(basePlant), NewPlant{ID: "fresh", Title: "Fresh", Empty: true}, "base")
	if err != nil {
		t.Fatal(err)
	}
	profiles, _ := doc["profiles"].([]any)
	nominal, _ := profiles[0].(map[string]any)
	if doc["id"] != "fresh" || doc["equipment"] != nil || doc["description"] != nil || len(profiles) != 3 || nominal["fault_weights"] != nil {
		t.Fatalf("empty plant = %v", doc)
	}
	copied, err := NewComposition([]byte(basePlant), NewPlant{ID: "copy", Title: "Copy"}, "base")
	if err != nil || copied["equipment"] == nil || copied["description"] != "old" {
		t.Fatalf("copy = %v, %v", copied, err)
	}
}

func TestANewPlantWithABadIDSaysWhatAnIDMustBe(t *testing.T) {
	t.Parallel()
	_, err := NewComposition([]byte(basePlant), NewPlant{ID: "ab", Title: "Short", Empty: true}, "base")
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != "/id" ||
		!strings.Contains(invalid.Issues[0].Message, "is not a plant id") {
		t.Fatalf("bad id = %v", err)
	}
	if _, err := NewComposition([]byte(`{"id":1}`), NewPlant{ID: "abc", Title: "T"}, "base"); err == nil || !strings.Contains(err.Error(), "base") {
		t.Fatalf("a broken base = %v", err)
	}
	empty := NewPlant{ID: "solo", Title: "Solo", Empty: true}
	doc, err := NewComposition([]byte(strings.Replace(basePlant, `"name":"nominal","duration_s":60,`, `"not_applicable":"no nominal day in this base plant","name":"n",`, 1)), empty, "base")
	if profiles, _ := doc["profiles"].([]any); err != nil || len(profiles) != 3 {
		t.Fatalf("all profiles not applicable = %v, %v", doc, err)
	}
}

func TestAnEmptyPlantAlwaysCarriesTheThreeProfilesTheDomainSpecRequires(t *testing.T) {
	t.Parallel()
	if got := requiredProfileNames(); len(got) != 3 || got[0] != "nominal" || got[1] != "correlated_cascade" || got[2] != "sensor_pathology" {
		t.Fatalf("required profiles = %v", got)
	}
	doc, err := NewComposition([]byte(basePlant), NewPlant{ID: "fresh", Title: "Fresh", Empty: true}, "base")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range doc["profiles"].([]any) {
		names = append(names, entry.(map[string]any)["name"].(string))
	}
	if strings.Join(names, ",") != "nominal,correlated_cascade,sensor_pathology" {
		t.Fatalf("profile names = %v", names)
	}
}

func TestAnEmptyPlantKeepsItsSiteAndStartsWithOneAreaNamedAfterIt(t *testing.T) {
	t.Parallel()
	sited := strings.Replace(basePlant, `"axes":{}`, `"site":{"name":"Pilot plant"},"areas":[{"name":"Utilities"},{"name":"Chiller hall","parent":"Utilities"}],"axes":{}`, 1)
	empty, err := NewComposition([]byte(sited), NewPlant{ID: "fresh", Title: "Fresh", Empty: true}, "base")
	if err != nil || areaLayout(empty) != "Pilot plant<" || empty["site"] == nil {
		t.Fatalf("empty plant areas = %v, site = %v, %v", areaLayout(empty), empty["site"], err)
	}
	copied, err := NewComposition([]byte(sited), NewPlant{ID: "copy", Title: "Copy"}, "base")
	if err != nil || areaLayout(copied) != "Utilities<,Chiller hall<Utilities" {
		t.Fatalf("copied areas = %v, %v", areaLayout(copied), err)
	}
}

func TestAnEmptyPlantWithoutASiteNameStartsWithASiteArea(t *testing.T) {
	t.Parallel()
	doc, err := NewComposition([]byte(basePlant), NewPlant{ID: "fresh", Title: "Fresh", Empty: true}, "base")
	if err != nil || areaLayout(doc) != "Site<" {
		t.Fatalf("areas = %v, %v", areaLayout(doc), err)
	}
}

func areaLayout(doc map[string]any) string {
	entries, _ := doc["areas"].([]any)
	var out []string
	for _, entry := range entries {
		area, _ := entry.(map[string]any)
		name, _ := area["name"].(string)
		parent, _ := area["parent"].(string)
		out = append(out, name+"<"+parent)
	}
	return strings.Join(out, ",")
}

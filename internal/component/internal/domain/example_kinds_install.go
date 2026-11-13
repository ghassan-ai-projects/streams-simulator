package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const kindFileDiffers = "kind_file_differs"

type KindFilesInstalled struct {
	Example   string   `json:"example"`
	Written   []string `json:"written"`
	Unchanged []string `json:"unchanged"`
}

type kindFilePlan struct {
	installed KindFilesInstalled
	writes    []KindFile
	differing []string
}

func PlanKindFileInstall(id string, files []KindFile, present map[string][]byte) (KindFilesInstalled, []KindFile, error) {
	plan := kindFilePlan{installed: KindFilesInstalled{Example: id, Written: []string{}, Unchanged: []string{}}, writes: []KindFile{}}
	for _, file := range files {
		plan.place(file, present)
	}
	if len(plan.differing) > 0 {
		return KindFilesInstalled{}, nil, kindFilesDiffer(id, plan.differing)
	}
	return plan.installed, plan.writes, nil
}

func (p *kindFilePlan) place(file KindFile, present map[string][]byte) {
	existing, found := present[file.Name]
	switch {
	case !found:
		p.installed.Written, p.writes = append(p.installed.Written, file.Name), append(p.writes, file)
	case sameDocument(existing, file.Document):
		p.installed.Unchanged = append(p.installed.Unchanged, file.Name)
	default:
		p.differing = append(p.differing, file.Name)
	}
}

func kindFilesDiffer(id string, names []string) error {
	var refused issueList
	for _, name := range names {
		refused.rule(pointer(name), kindFileDiffers, "",
			"The workspace already has a %s that differs from the one example %s implements; it was left as it is and no kind file was written.", name, id)
	}
	summary := fmt.Sprintf("component: the kind files of example %s were not written: the kinds folder already has a different %s", id, strings.Join(names, ", "))
	return &model.ValidationError{Contract: model.ContractKindFile, Source: "example " + id, Summary: summary, Issues: refused.items}
}

func sameDocument(a, b []byte) bool {
	left, leftOK := canonicalText(a)
	right, rightOK := canonicalText(b)
	return leftOK && rightOK && left == right
}

func canonicalText(raw []byte) (string, bool) {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return "", false
	}
	text, err := canonical.MarshalString(value)
	return text, err == nil
}

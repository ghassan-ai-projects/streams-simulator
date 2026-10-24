package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	foldersKinds        = "kinds/"
	foldersTwins        = "twins/"
	foldersCompositions = "compositions/"
	foldersRuns         = "runs/"
)

type kindInfo struct {
	file, name, digest string
	doc                map[string]any
}

type twinInfo struct {
	file, id, version string
	raw               []byte
	kinds             []string
}

type compositionInfo struct {
	file, id, version string
	raw               []byte
}

type runInfo struct {
	RunLineage
	domain string
}

type index struct {
	kinds        []kindInfo
	twins        []twinInfo
	compositions []compositionInfo
	runs         []runInfo
	unreadable   []Unreadable
}

func (lib *Library) indexWorkspace(workspace Workspace) *index {
	found := &index{unreadable: []Unreadable{}}
	found.indexKinds(workspace.Kinds)
	found.indexTwins(lib, workspace.Twins)
	found.indexCompositions(workspace.Compositions)
	found.indexRuns(workspace.Runs)
	return found
}

func (x *index) unreadableFile(code, file string, err error) {
	x.unreadable = append(x.unreadable, Unreadable{File: file, Code: code, Message: oneLine(err.Error())})
}

func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func (x *index) indexKinds(files []NamedFile) {
	for _, file := range byName(files) {
		info, err := kindInfoOf(file)
		if err != nil {
			x.unreadableFile(UnreadableKind, foldersKinds+file.Name, err)
			continue
		}
		x.kinds = append(x.kinds, info)
	}
}

func byName(files []NamedFile) []NamedFile {
	return slices.SortedFunc(slices.Values(files), func(a, b NamedFile) int { return strings.Compare(a.Name, b.Name) })
}

func kindInfoOf(file NamedFile) (kindInfo, error) {
	if file.Err != nil {
		return kindInfo{}, file.Err
	}
	var doc map[string]any
	if err := model.DecodeBytes(file.Raw, &doc); err != nil {
		return kindInfo{}, fmt.Errorf("not valid JSON: %w", err)
	}
	name, _ := doc["kind"].(string)
	digest := kindFileDigest(doc)
	if digest == "" || name == "" {
		return kindInfo{}, errors.New("a kind file names its kind")
	}
	return kindInfo{file: foldersKinds + file.Name, name: name, digest: digest, doc: doc}, nil
}

func kindFileDigest(doc map[string]any) string {
	body := maps.Clone(doc)
	delete(body, "kind_digest")
	digest, _ := canonical.Digest(body)
	return digest
}

func (x *index) indexTwins(lib *Library, files []NamedFile) {
	for _, file := range byName(files) {
		info, err := twinInfoOf(lib, file)
		if err != nil {
			x.unreadableFile(UnreadableTwin, foldersTwins+file.Name, err)
			continue
		}
		x.twins = append(x.twins, info)
	}
}

func twinInfoOf(lib *Library, file NamedFile) (twinInfo, error) {
	if file.Err != nil {
		return twinInfo{}, file.Err
	}
	choices, err := parseTwinChoices(lib.contracts.twin, file.Raw, foldersTwins+file.Name)
	if err != nil {
		return twinInfo{}, err
	}
	info := twinInfo{file: foldersTwins + file.Name, id: choices.ID, version: choices.Version, raw: file.Raw}
	for _, kind := range choices.Kinds {
		info.kinds = append(info.kinds, kind.Kind)
	}
	return info, nil
}

func (x *index) indexCompositions(files []NamedFile) {
	for _, file := range byName(files) {
		info, err := compositionInfoOf(file)
		if err != nil {
			x.unreadableFile(UnreadableComposition, foldersCompositions+file.Name, err)
			continue
		}
		x.compositions = append(x.compositions, info)
	}
}

func compositionInfoOf(file NamedFile) (compositionInfo, error) {
	var peek struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	}
	if file.Err != nil {
		return compositionInfo{}, file.Err
	}
	if err := json.Unmarshal(file.Raw, &peek); err != nil || peek.ID == "" {
		return compositionInfo{}, orError(err, "a composition names its id")
	}
	return compositionInfo{file: foldersCompositions + file.Name, id: peek.ID, version: peek.Version, raw: file.Raw}, nil
}

func orError(err error, otherwise string) error {
	if err != nil {
		return err
	}
	return errors.New(otherwise)
}

func (x *index) indexRuns(folders []RunFolder) {
	for _, folder := range slices.SortedFunc(slices.Values(folders), func(a, b RunFolder) int { return strings.Compare(a.ID, b.ID) }) {
		info, err := runInfoOf(folder)
		if err != nil {
			x.unreadableFile(UnreadableRun, foldersRuns+folder.ID, err)
			continue
		}
		x.runs = append(x.runs, info)
	}
}

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type recordedDomain struct {
	Mode   string            `json:"mode"`
	Domain model.ArtifactRef `json:"domain"`
}

func runInfoOf(folder RunFolder) (runInfo, error) {
	if folder.Err != nil {
		return runInfo{}, folder.Err
	}
	opened, recorded, err := recordedDomainOf(folder)
	if err != nil {
		return runInfo{}, err
	}
	return runInfo{RunLineage: RunLineage{ID: folder.ID, Mode: opened.Mode, Recorded: recorded, DomainVersion: opened.Domain.Version, DomainDigest: opened.Domain.Digest},
		domain: opened.Domain.ID}, nil
}

func recordedDomainOf(folder RunFolder) (recordedDomain, bool, error) {
	raw, recorded := folder.RunJSON, true
	if raw == nil {
		raw, recorded = firstLine(folder.Observer), false
	}
	var opened recordedDomain
	if err := json.Unmarshal(raw, &opened); err != nil || !digestPattern.MatchString(opened.Domain.Digest) {
		return recordedDomain{}, false, errors.New("run.json, or the first line of journal.jsonl when there is none, does not name the domain and digest the run used")
	}
	return opened, recorded, nil
}

func firstLine(raw []byte) []byte {
	line, _, _ := strings.Cut(string(raw), "\n")
	return []byte(line)
}

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
)

var ErrExampleKindFileMissing = errors.New("component: an example implements a kind file or board it does not carry")

const (
	kindFileSuffix  = ".kind.json"
	boardFileSuffix = ".board.json"
)

type KindFile struct {
	Name     string          `json:"name"`
	Document json.RawMessage `json:"document"`
}

type exampleKindFiles struct {
	kinds  map[string]KindFile
	boards map[string]KindFile
}

type kindFileHeader struct {
	Kind      string `json:"kind"`
	BoardType string `json:"board_type"`
}

type implementingEntry struct {
	Implements string `json:"implements"`
	BoardType  string `json:"board_type"`
}

func loadExampleKindFiles(files []File) (exampleKindFiles, error) {
	pool := exampleKindFiles{kinds: map[string]KindFile{}, boards: map[string]KindFile{}}
	for _, file := range files {
		if err := pool.add(file); err != nil {
			return exampleKindFiles{}, err
		}
	}
	return pool, nil
}

func (p exampleKindFiles) add(file File) error {
	var header kindFileHeader
	if err := json.Unmarshal(file.Raw, &header); err != nil {
		return fmt.Errorf("component: example kind file %s: %w", file.Path, err)
	}
	entry := KindFile{Name: path.Base(file.Path), Document: json.RawMessage(file.Raw)}
	files, key := p.shelfOf(entry.Name, header)
	if key == "" {
		return fmt.Errorf("component: example kind file %s names no kind or board type", file.Path)
	}
	files[key] = entry
	return nil
}

func (p exampleKindFiles) shelfOf(name string, header kindFileHeader) (map[string]KindFile, string) {
	if strings.HasSuffix(name, boardFileSuffix) {
		return p.boards, header.BoardType
	}
	if strings.HasSuffix(name, kindFileSuffix) {
		return p.kinds, header.Kind
	}
	return nil, ""
}

func (p exampleKindFiles) of(id string, composition []byte) ([]KindFile, error) {
	var plant struct {
		Equipment []implementingEntry `json:"equipment"`
	}
	if err := json.Unmarshal(composition, &plant); err != nil {
		return nil, fmt.Errorf("component: example %s: %w", id, err)
	}
	chosen := map[string]KindFile{}
	for _, entry := range plant.Equipment {
		if err := p.choose(chosen, id, entry); err != nil {
			return nil, err
		}
	}
	return sortedKindFiles(chosen), nil
}

func (p exampleKindFiles) choose(chosen map[string]KindFile, id string, entry implementingEntry) error {
	if err := pick(chosen, p.kinds, id, entry.Implements); err != nil {
		return err
	}
	return pick(chosen, p.boards, id, entry.BoardType)
}

func pick(chosen, carried map[string]KindFile, id, name string) error {
	if name == "" {
		return nil
	}
	file, found := carried[name]
	if !found {
		return fmt.Errorf("%w: example %s names %q", ErrExampleKindFileMissing, id, name)
	}
	chosen[file.Name] = file
	return nil
}

func sortedKindFiles(chosen map[string]KindFile) []KindFile {
	files := make([]KindFile, 0, len(chosen))
	for _, name := range slices.Sorted(maps.Keys(chosen)) {
		files = append(files, chosen[name])
	}
	return files
}

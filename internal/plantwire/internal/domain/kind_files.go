package domain

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const KindFileSuffix = ".kind.json"

var ErrKindFileDuplicate = errors.New("plantwire: two kind files define one kind")

type KindSource struct {
	Path string
	Raw  []byte
}

type KindFiles map[string]map[string]any

type KindSummary struct {
	Path       string `json:"path"`
	Kind       string `json:"kind"`
	KindDigest string `json:"kind_digest"`
	Metrics    int    `json:"metrics"`
	Commands   int    `json:"commands"`
}

func ParseKindFiles(sources []KindSource) (KindFiles, []KindSummary, error) {
	index := kindIndex{files: KindFiles{}, origin: map[string]string{}}
	summaries := make([]KindSummary, 0, len(sources))
	for _, source := range sources {
		summary, err := index.parse(source)
		if err != nil {
			return nil, nil, err
		}
		summaries = append(summaries, summary)
	}
	return index.files, summaries, nil
}

type kindIndex struct {
	files  KindFiles
	origin map[string]string
}

func (i kindIndex) parse(source KindSource) (KindSummary, error) {
	entry, err := parseKindFile(source)
	if err != nil {
		return KindSummary{}, err
	}
	return summarizeKind(source.Path, entry), i.add(source.Path, entry)
}

func (i kindIndex) add(path string, entry map[string]any) error {
	name := kindNameOf(entry)
	if first, taken := i.origin[name]; taken {
		return fmt.Errorf("%w: %q is defined by %s and %s", ErrKindFileDuplicate, name, first, path)
	}
	i.files[name], i.origin[name] = entry, path
	return nil
}

func parseKindFile(source KindSource) (map[string]any, error) {
	var entry map[string]any
	if err := decode(source.Raw, &entry); err != nil {
		return nil, model.NotJSONError(model.ContractKindFile, source.Path, "plantwire: kind file "+source.Path, err)
	}
	if err := CheckKind(entry); err != nil {
		return nil, fmt.Errorf("%s: %w", source.Path, model.WithSource(err, source.Path))
	}
	return entry, nil
}

func summarizeKind(path string, entry map[string]any) KindSummary {
	metrics, _ := entry["metrics"].([]any)
	commands, _ := entry["commands"].([]any)
	return KindSummary{Path: path, Kind: kindNameOf(entry), KindDigest: KindDigest(entry), Metrics: len(metrics), Commands: len(commands)}
}

func kindNameOf(entry map[string]any) string {
	name, _ := entry["kind"].(string)
	return name
}

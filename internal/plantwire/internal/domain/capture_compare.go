package domain

import "slices"

const (
	CompareNew     = "new"
	CompareSame    = "same"
	CompareDiffers = "differs"
)

type KindComparison struct {
	Status     string `json:"status"`
	File       string `json:"file,omitempty"`
	FileDigest string `json:"file_digest,omitempty"`
	DiffersAt  string `json:"differs_at,omitempty"`
}

func KindNames(capture Capture) map[string]bool {
	names := make(map[string]bool, len(capture.Kinds))
	for _, kind := range capture.Kinds {
		names[kind.Kind] = true
	}
	return names
}

func CompareWithFiles(capture Capture, files KindFiles, summaries []KindSummary) (Capture, error) {
	located := locateKinds(summaries)
	compared := slices.Clone(capture.Kinds)
	for i := range compared {
		comparison, err := compareKind(compared[i], files[compared[i].Kind], located[compared[i].Kind])
		if err != nil {
			return Capture{}, err
		}
		compared[i].Comparison = &comparison
	}
	capture.Kinds = compared
	return capture, nil
}

func locateKinds(summaries []KindSummary) map[string]KindSummary {
	located := make(map[string]KindSummary, len(summaries))
	for _, summary := range summaries {
		located[summary.Kind] = summary
	}
	return located
}

func compareKind(kind CapturedKind, file map[string]any, located KindSummary) (KindComparison, error) {
	if file == nil {
		return KindComparison{Status: CompareNew}, nil
	}
	comparison := KindComparison{Status: CompareSame, File: located.Path, FileDigest: located.KindDigest}
	if located.KindDigest == kind.KindDigest {
		return comparison, nil
	}
	comparison.Status = CompareDiffers
	var err error
	comparison.DiffersAt, err = differenceAt(file, kind.Draft)
	return comparison, err
}

func differenceAt(file, draft map[string]any) (string, error) {
	want, err := canonicalTree(file)
	if err != nil {
		return "", err
	}
	got, err := canonicalTree(draft)
	if err != nil {
		return "", err
	}
	at, _ := firstDifference(want, got, "")
	return at, nil
}

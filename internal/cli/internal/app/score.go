package app

import (
	"context"
	"errors"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading"
)

type scoreOptions struct{ artifact, label, domainsDir, adaptersDir string }

func cmdScore(s *session, args []string) (any, error) {
	options, err := parseScoreOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	graded, err := grading.GradeRun(context.Background(), grading.Options{Artifact: options.artifact, Label: options.label,
		DomainsDir: options.domainsDir, AdaptersDir: options.adaptersDir})
	if err != nil {
		return nil, gradingFailure(err)
	}
	return grading.Document(graded), nil
}

func gradingFailure(err error) error {
	var mismatch *grading.ReplayMismatchError
	if !errors.As(err, &mismatch) {
		return err
	}
	return &outcomeError{code: codeReplayMismatch, message: mismatch.Error(), result: replayResult(mismatch.Result)}
}

func parseScoreOptions(args []string, stderr io.Writer) (scoreOptions, error) {
	var options scoreOptions
	fs := newFlagSet("score", stderr)
	fs.StringVar(&options.artifact, "run", "", "run artifact (run.json)")
	fs.StringVar(&options.label, "label", "", "ground truth label (JSON; default label.json or labels.json beside the run)")
	registerInputDirs(fs, &options.domainsDir, &options.adaptersDir)
	if err := parseOnlyFlags(fs, args); err != nil {
		return scoreOptions{}, err
	}
	if options.artifact == "" {
		return scoreOptions{}, usagef("score requires --run")
	}
	return options, nil
}

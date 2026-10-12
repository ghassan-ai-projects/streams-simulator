package app

import (
	"context"
	"errors"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading"
)

// scoreOptions are the inputs of the offline score command.
type scoreOptions struct{ artifact, label, domainsDir, adaptersDir string }

// cmdScore grades a finished run from its artifacts (CLI_SURFACE §1). A run
// whose replay does not reproduce its recorded trace is not graded: the
// command fails with the replay report as its result.
func cmdScore(s *session, args []string) (any, error) {
	options, err := parseScoreOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	card, err := grading.Grade(context.Background(), grading.Options{Artifact: options.artifact, Label: options.label,
		DomainsDir: options.domainsDir, AdaptersDir: options.adaptersDir})
	if err != nil {
		return nil, gradingFailure(err)
	}
	return card, nil
}

// gradingFailure restates the refusal to grade a run that does not replay as
// a failed check carrying the replay report; any other failure is passed on.
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
	fs.StringVar(&options.label, "label", "", "ground truth label (JSON)")
	registerInputDirs(fs, &options.domainsDir, &options.adaptersDir)
	if err := parseOnlyFlags(fs, args); err != nil {
		return scoreOptions{}, err
	}
	if options.artifact == "" {
		return scoreOptions{}, usagef("score requires --run")
	}
	return options, nil
}

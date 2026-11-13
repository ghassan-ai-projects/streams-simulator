package app

import (
	"context"
	"errors"
	"flag"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading"
)

type scoreOptions struct{ run, label, domainsDir, adaptersDir string }

func cmdScore(s *session, args []string) (any, error) {
	options, err := parseScoreOptions("score", "run artifact (run.json)", args, s.stderr)
	if err != nil {
		return nil, err
	}
	graded, err := grading.GradeRun(context.Background(), grading.Options{Artifact: options.run, Label: options.label,
		DomainsDir: options.domainsDir, AdaptersDir: options.adaptersDir})
	if err != nil {
		return nil, gradingFailure(err)
	}
	return grading.Document(graded), nil
}

func cmdOutcome(s *session, args []string) (any, error) {
	options, err := parseScoreOptions("outcome", "run folder", args, s.stderr)
	if err != nil {
		return nil, err
	}
	graded, err := grading.GradeOutcome(context.Background(), grading.OutcomeOptions{RunDir: options.run,
		DomainsDir: options.domainsDir, AdaptersDir: options.adaptersDir})
	if err != nil {
		return nil, gradingFailure(err)
	}
	return graded, nil
}

func gradingFailure(err error) error {
	var mismatch *grading.ReplayMismatchError
	if !errors.As(err, &mismatch) {
		return err
	}
	return &outcomeError{code: codeReplayMismatch, message: mismatch.Error(), result: replayResult(mismatch.Result)}
}

func parseScoreOptions(command, runUsage string, args []string, stderr io.Writer) (scoreOptions, error) {
	var options scoreOptions
	if err := parseOnlyFlags(scoreFlags(command, runUsage, stderr, &options), args); err != nil {
		return scoreOptions{}, err
	}
	if options.run == "" {
		return scoreOptions{}, usagef("%s requires --run", command)
	}
	return options, nil
}

func scoreFlags(command, runUsage string, stderr io.Writer, options *scoreOptions) *flag.FlagSet {
	fs := newFlagSet(command, stderr)
	fs.StringVar(&options.run, "run", "", runUsage)
	if command == "score" {
		fs.StringVar(&options.label, "label", "", "ground truth label (JSON; default label.json or labels.json beside the run)")
	}
	registerInputDirs(fs, &options.domainsDir, &options.adaptersDir)
	return fs
}

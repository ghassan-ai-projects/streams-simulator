package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report"
)

type reportOptions struct {
	report.Request
	format, out string
	compare     bool
}

func cmdReport(s *session, args []string) (any, error) {
	options, err := parseReportOptions(s, args)
	if err != nil {
		return nil, err
	}
	document, err := report.Produce(context.Background(), options.Request)
	if errors.Is(err, report.ErrRunNotEnded) {
		return nil, &outcomeError{code: codeRunNotEnded, message: err.Error()}
	}
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return deliverReport(document, options)
}

func deliverReport(document any, options reportOptions) (any, error) {
	if options.out == "" && options.format == report.FormatJSON {
		return document, nil
	}
	content, err := report.Render(document, options.format)
	switch {
	case err != nil:
		return nil, fmt.Errorf("%w", err)
	case options.out == "":
		return map[string]any{"format": options.format, "markdown": string(content)}, nil
	}
	return map[string]any{"report_written": options.out, "format": options.format}, files.Write(options.out, content)
}

func parseReportOptions(s *session, args []string) (reportOptions, error) {
	options := reportOptions{compare: len(args) > 0 && args[0] == "compare"}
	if options.compare {
		args = args[1:]
	}
	fs := newFlagSet("report", s.stderr)
	registerReportFlags(fs, &options.format, &options.out, &options.A, &options.B, &options.Replay)
	registerInputDirs(fs, &options.Directories.Domains, &options.Directories.Adapters)
	if err := parseFlags(fs, args); err != nil {
		return reportOptions{}, err
	}
	options.Dir = fs.Arg(0)
	return options, admitReportOptions(options, fs.NArg())
}

func admitReportOptions(options reportOptions, positional int) error {
	switch {
	case options.format != report.FormatJSON && options.format != report.FormatMarkdown:
		return usagef("report: --format must be json or markdown, got %q", options.format)
	case options.compare && (positional != 0 || options.A == "" || options.B == "" || options.Replay):
		return usagef("usage: streamsim report compare --a <run-dir> --b <run-dir> [--format json|markdown] [--out <file>]")
	case !options.compare && (positional != 1 || options.A != "" || options.B != ""):
		return usagef("usage: streamsim report <run-dir> [--format json|markdown] [--out <file>] [--replay]")
	}
	return nil
}

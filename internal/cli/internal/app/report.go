package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report"
)

type reportOptions struct{ dir, format, out string }

func cmdReport(s *session, args []string) (any, error) {
	options, err := parseReportOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	built, err := report.Build(context.Background(), options.dir)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if options.out == "" {
		return built, nil
	}
	return writeReport(built, options)
}

func writeReport(built *report.Report, options reportOptions) (any, error) {
	content, err := report.Render(built, options.format)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if err := files.Write(options.out, content); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"report_written": options.out, "format": options.format}, nil
}

func parseReportOptions(args []string, stderr io.Writer) (reportOptions, error) {
	var options reportOptions
	fs := newFlagSet("report", stderr)
	fs.StringVar(&options.format, "format", report.FormatJSON, "report format: json|markdown")
	fs.StringVar(&options.out, "out", "", "file to write the report to (required for markdown)")
	if err := parseFlags(fs, args); err != nil {
		return reportOptions{}, err
	}
	if fs.NArg() != 1 {
		return reportOptions{}, usagef("usage: streamsim report <run-dir> [--format json|markdown] [--out <file>]")
	}
	options.dir = fs.Arg(0)
	return options, admitReportFormat(options)
}

func admitReportFormat(options reportOptions) error {
	if options.format != report.FormatJSON && options.format != report.FormatMarkdown {
		return usagef("report: --format must be json or markdown, got %q", options.format)
	}
	if options.format == report.FormatMarkdown && options.out == "" {
		return usagef("report: --format markdown requires --out: stdout carries one JSON document")
	}
	return nil
}

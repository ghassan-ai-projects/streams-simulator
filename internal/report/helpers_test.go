package report_test

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/report"
)

func build(ctx context.Context, dir string) (*report.Report, error) {
	document, err := report.Produce(ctx, report.Request{Dir: dir})
	if err != nil {
		return nil, fmt.Errorf("report: %w", err)
	}
	return document.(*report.Report), nil
}

func buildVerified(ctx context.Context, dir string, dirs report.Directories) (*report.Report, error) {
	document, err := report.Produce(ctx, report.Request{Dir: dir, Replay: true, Directories: dirs})
	if err != nil {
		return nil, fmt.Errorf("report: %w", err)
	}
	return document.(*report.Report), nil
}

func compare(ctx context.Context, dirA, dirB string) (*report.Comparison, error) {
	document, err := report.Produce(ctx, report.Request{A: dirA, B: dirB})
	if err != nil {
		return nil, fmt.Errorf("report: %w", err)
	}
	return document.(*report.Comparison), nil
}

package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
)

func Compare(ctx context.Context, dirA, dirB string) (*domain.Comparison, error) {
	a, err := comparableReport(ctx, "A", dirA)
	if err != nil {
		return nil, err
	}
	b, err := comparableReport(ctx, "B", dirB)
	if err != nil {
		return nil, err
	}
	return validated(domain.Compare(a, b))
}

func validated(comparison *domain.Comparison) (*domain.Comparison, error) {
	raw, err := domain.ComparisonJSON(comparison)
	if err != nil {
		return nil, err
	}
	if err := model.ValidateRunComparison(raw); err != nil {
		return nil, fmt.Errorf("report: %w", err)
	}
	return comparison, nil
}

func comparableReport(ctx context.Context, side, dir string) (*domain.Report, error) {
	built, err := Build(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("report: run %s: %w", side, err)
	}
	if reason := domain.NotComparable(built); reason != "" {
		return nil, fmt.Errorf("%w: run %s: %s", domain.ErrNotEnded, side, reason)
	}
	return built, nil
}

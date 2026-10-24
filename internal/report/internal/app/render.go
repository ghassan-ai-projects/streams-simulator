package app

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
)

var (
	ErrUnknownFormat   = errors.New("report: unknown format")
	ErrUnknownDocument = errors.New("report: unknown document")
)

const (
	JSONFormat = "json"
	Markdown   = "markdown"
)

func Render(document any, format string) ([]byte, error) {
	if format != JSONFormat && format != Markdown {
		return nil, fmt.Errorf("%w %q", ErrUnknownFormat, format)
	}
	switch built := document.(type) {
	case *domain.Report:
		return renderReport(built, format)
	case *domain.Comparison:
		return renderComparison(built, format)
	}
	return nil, fmt.Errorf("%w %T", ErrUnknownDocument, document)
}

func renderReport(report *domain.Report, format string) ([]byte, error) {
	if format == JSONFormat {
		return domain.JSON(report)
	}
	return domain.Markdown(report), nil
}

func renderComparison(comparison *domain.Comparison, format string) ([]byte, error) {
	if format == JSONFormat {
		return domain.ComparisonJSON(comparison)
	}
	return domain.ComparisonMarkdown(comparison), nil
}

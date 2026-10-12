package report

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/app"
)

func Build(ctx context.Context, dir string) (*Report, error) {
	return app.Build(ctx, dir)
}

func Render(report *Report, format string) ([]byte, error) {
	return app.Render(report, format)
}

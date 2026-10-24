package report

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/app"
)

func Produce(ctx context.Context, request Request) (any, error) {
	return app.Produce(ctx, request)
}

func Render(document any, format string) ([]byte, error) {
	return app.Render(document, format)
}

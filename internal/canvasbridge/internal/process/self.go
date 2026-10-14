package process

import (
	"context"
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func Self() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("canvasbridge: locate the streamsim executable: %w", err)
	}
	return executable, nil
}

func UntilInterrupted(ctx context.Context) (context.Context, context.CancelFunc) {
	return kernel.UntilInterrupted(ctx)
}

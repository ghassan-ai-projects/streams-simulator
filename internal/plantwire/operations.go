package plantwire

import (
	"context"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func Serve(ctx context.Context, r *run.Run, socketPath string, barrierCap time.Duration, options Options) error {
	return app.Serve(ctx, r, socketPath, barrierCap, options)
}

func Host(ctx context.Context, socketPath string, delay time.Duration, options HostOptions) (HostReport, error) {
	return app.ServeHost(ctx, socketPath, delay, options)
}

func ParseScript(raw []byte) ([]Scripted, error) {
	return host.ParseScript(raw)
}

func ParseInventory(raw []byte) (map[string]string, error) {
	return host.ParseInventory(raw)
}

func ContractDigest() string {
	return domain.ContractDigest()
}

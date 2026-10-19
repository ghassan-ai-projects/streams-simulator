package plantwire

import (
	"context"
	"io"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func Serve(ctx context.Context, r *run.Run, socketPath string, barrierCap time.Duration, options Options) error {
	return app.Serve(ctx, r, socketPath, barrierCap, options)
}

func Host(ctx context.Context, socketPath string, delay time.Duration, options HostOptions) (HostReport, error) {
	return app.ServeHost(ctx, socketPath, delay, options)
}

func LoadHost(options HostOptions, scriptPath, inventoryPath string) (HostOptions, error) {
	return app.LoadHost(options, scriptPath, inventoryPath)
}

func Relay(ctx context.Context, options RelayOptions) (RelayStats, error) {
	return app.Relay(ctx, options)
}

func ReplayTrace(ctx context.Context, options ReplayOptions) (ReplayStats, error) {
	return app.ReplayTrace(ctx, options)
}

func Connect(ctx context.Context, options ConnectOptions) (ConnectResult, error) {
	return app.Connect(ctx, options)
}

func CheckConnect(options ConnectOptions) (ConnectCheck, error) {
	return app.CheckConnect(options)
}

func CheckImplements(spec *model.DomainSpec, kindsDirs []string) ([]model.ImplementsChecked, error) {
	return app.CheckImplements(spec, kindsDirs)
}

func Scaffold[T any](paths, dirs []string, source string, scaffold func(map[string]map[string]any) (T, *model.DomainSpec, error)) (T, error) {
	return app.Scaffold(paths, dirs, source, scaffold)
}

func WithKindFiles[T any](paths, dirs []string, use func(map[string]map[string]any) (T, error)) (T, error) {
	return app.WithKindFiles(paths, dirs, use)
}

func ValidateKindFilesWith(paths []string, rules TwinRules) ([]KindSummary, error) {
	return app.ValidateKindFilesWith(paths, rules)
}

func ValidateKindStream(reader io.Reader, rules TwinRules) (KindSummary, error) {
	return app.ValidateKindStream(reader, rules)
}

func CaptureFile(path string, kindsDirs []string) (Capture, error) {
	return app.CaptureFile(path, kindsDirs)
}

func CaptureStream(reader io.Reader, kindsDirs []string) (Capture, error) {
	return app.CaptureStream(reader, kindsDirs)
}

func TwinReport(ctx context.Context, options TwinReportOptions) (TwinReportDocument, error) {
	return app.TwinReport(ctx, options)
}

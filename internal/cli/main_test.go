package cli_test

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli"
)

func TestMainReportsHelpAndUsageErrorsAsExitStatuses(t *testing.T) {
	t.Parallel()
	build := cli.Build{Version: "v-test", Commit: "c-test"}
	if code := cli.Main([]string{"streamsim", "help"}, build); code != 0 {
		t.Fatalf("help exit = %d", code)
	}
	if code := cli.Main([]string{"streamsim"}, build); code != 2 {
		t.Fatalf("no command exit = %d", code)
	}
	if code := cli.Main([]string{"streamsim", "frobnicate"}, build); code != 2 {
		t.Fatalf("unknown command exit = %d", code)
	}
}

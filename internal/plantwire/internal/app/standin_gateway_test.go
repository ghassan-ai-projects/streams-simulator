package app

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
)

// The canvas end-to-end tests run this test binary as a stand-in for a gateway with a real board behind it:
// go test -c -o <file> ./internal/plantwire/internal/app, then <file> --standin-gateway --socket ... --device name:kind:boot.
func TestMain(m *testing.M) {
	if index := slices.Index(os.Args, "--standin-gateway"); index >= 0 {
		os.Exit(playStandInGateway(os.Args[index+1:]))
	}
	os.Exit(m.Run())
}

type deviceFlags []string

func (d *deviceFlags) String() string { return strings.Join(*d, ",") }

func (d *deviceFlags) Set(value string) error {
	*d = append(*d, value)
	return nil
}

func playStandInGateway(args []string) int {
	flags := flag.NewFlagSet("standin-gateway", flag.ContinueOnError)
	socket, kinds, board := flags.String("socket", "", "the gateway socket the session publishes"), flags.String("kinds-dir", "", "kind files"),
		flags.String("board", "uno-07", "the board the gateway announces")
	var devices deviceFlags
	flags.Var(&devices, "device", "name:kind:boot, repeatable")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	node, err := standInNode(*kinds, *board, devices)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if err := playOn(*socket, node); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func standInNode(kindsDir, board string, devices []string) (*standin.Node, error) {
	config := standin.Config{Session: "standin-1", Board: board}
	for _, device := range devices {
		parts := strings.SplitN(device, ":", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("device %q is not name:kind:boot", device)
		}
		raw, err := os.ReadFile(filepath.Join(kindsDir, parts[1]+".kind.json"))
		if err != nil {
			return nil, fmt.Errorf("read kind %q: %w", parts[1], err)
		}
		var kind map[string]any
		if err := json.Unmarshal(raw, &kind); err != nil {
			return nil, fmt.Errorf("decode kind %q: %w", parts[1], err)
		}
		config.Kinds = append(config.Kinds, kind)
		config.Devices = append(config.Devices, standin.Device{Name: parts[0], Kind: parts[1], BootID: parts[2]})
	}
	return standin.New(config)
}

func playOn(socket string, node *standin.Node) error {
	link, err := dialWhenPublished(socket)
	if err != nil {
		return err
	}
	defer func() { _ = link.Close() }()
	birth, err := node.Birth()
	if err != nil {
		return err
	}
	if err := link.Send([]map[string]any{birth}); err != nil {
		return err
	}
	return answer(link, node)
}

func dialWhenPublished(socket string) (*uds.Link, error) {
	ticks := time.NewTicker(100 * time.Millisecond)
	defer ticks.Stop()
	for attempt := 0; ; attempt++ {
		link, err := uds.Dial(context.Background(), socket, 30*time.Second)
		if err == nil || attempt == 300 {
			return link, err
		}
		<-ticks.C
	}
}

func answer(link *uds.Link, node *standin.Node) error {
	var state map[string]any
	for in := range link.Incoming() {
		replies, err := reply(node, in, &state)
		if err != nil {
			return err
		}
		if err := link.Send(replies); err != nil {
			return nil
		}
	}
	return nil
}

func reply(node *standin.Node, in domain.Inbound, state *map[string]any) ([]map[string]any, error) {
	switch {
	case in.Type == "STATE" && in.Online && *state == nil:
		*state = in.Record
		return node.Accept(in.Record)
	case in.Type == "NCMD" && *state != nil:
		birth, err := node.Birth()
		if err != nil {
			return nil, err
		}
		births, err := node.Accept(*state)
		return append([]map[string]any{birth}, births...), err
	case in.Type == "DCMD":
		return node.Handle(in.Record), nil
	}
	return nil, nil
}

package app

import (
	"bufio"
	"errors"
	"io/fs"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const hostStateTemplate = `{"type":"STATE","lane":"control","session_id":"h","protocol_version":1,"seq":%SEQ%,"host_id":"h","online":true,"heartbeat":0,` +
	`"heartbeat_interval_ms":1000,"authority_epoch":7,"protocol_min":%MIN%,"protocol_max":%MAX%}`

func stateLine(seq, minVersion, maxVersion string) string {
	return strings.NewReplacer("%SEQ%", seq, "%MIN%", minVersion, "%MAX%", maxVersion).Replace(hostStateTemplate)
}

func scriptedReplayHost(t *testing.T, path, answer, hangUpAfter string) {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		reader := bufio.NewReader(conn)
		if _, err := reader.ReadString('\n'); err != nil {
			return
		}
		if answer != "" {
			_, _ = conn.Write([]byte(answer + "\n"))
		}
		for hangUpAfter != "" {
			if line, err := reader.ReadString('\n'); err != nil || strings.Contains(line, hangUpAfter) {
				return
			}
		}
	}()
}

func TestAReplayRefusesAHostThatDoesNotOpenWithAUsableSTATE(t *testing.T) {
	t.Parallel()
	folder := recordedHardwareRun(t, 22.5)
	cases := map[string]struct {
		answer string
		want   error
		text   string
	}{
		"a host that hangs up":                      {"", ErrHostLost, ""},
		"a STATE that is not the first record":      {stateLine("4", "1", "1"), ErrHostGap, ""},
		"a host with no protocol version in common": {stateLine("0", "2", "3"), nil, "handshake"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			hostPath := testsupport.SocketPath(t)
			scriptedReplayHost(t, hostPath, tc.answer, "")
			_, err := ReplayTrace(t.Context(), replayOptions(folder, hostPath))
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || !strings.Contains(err.Error(), tc.text) {
				t.Fatalf("ReplayTrace = %v, want a refusal %v mentioning %q", err, tc.want, tc.text)
			}
		})
	}
}

func TestAReplayWhoseHostHangsUpBeforeAnsweringTheBarrierIsLost(t *testing.T) {
	t.Parallel()
	folder := recordedHardwareRun(t, 22.5)
	hostPath := testsupport.SocketPath(t)
	scriptedReplayHost(t, hostPath, stateLine("0", "1", "1"), `"type":"NSYNC"`)
	if _, err := ReplayTrace(t.Context(), replayOptions(folder, hostPath)); !errors.Is(err, ErrHostLost) {
		t.Fatalf("ReplayTrace = %v, want %v", err, ErrHostLost)
	}
}

func TestAReplayNeedsItsOptionsAndAHostToDial(t *testing.T) {
	t.Parallel()
	folder := recordedHardwareRun(t, 22.5)
	good := replayOptions(folder, testsupport.SocketPath(t))
	cases := map[string]func(*ReplayOptions){
		"no run folder":     func(o *ReplayOptions) { o.Run = "" },
		"no host socket":    func(o *ReplayOptions) { o.HostSocket = "" },
		"no barrier cap":    func(o *ReplayOptions) { o.BarrierCap = 0 },
		"an unsafe session": func(o *ReplayOptions) { o.SessionID = "not wire safe!" },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options := good
			change(&options)
			if _, err := ReplayTrace(t.Context(), options); !errors.Is(err, ErrBadOptions) {
				t.Fatalf("ReplayTrace = %v, want %v", err, ErrBadOptions)
			}
		})
	}
	if _, err := ReplayTrace(t.Context(), good); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReplayTrace with no host listening = %v, want %v", err, fs.ErrNotExist)
	}
}

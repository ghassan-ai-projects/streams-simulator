package app

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
)

func digestOf(char string) string { return "sha256:" + strings.Repeat(char, 64) }

func meterCommand(digest string) func(h *scriptedHost, syncs int) {
	return func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", set(commandFields("c-1", "meter-1", "shed_load", map[string]any{"meter_id": "meter-1"}, sliceEnd(1)), "policy_digest", digest))
		}
	}
}

func TestTheNodeExecutesACommandOnlyWhenItsPolicyDigestIsOnTheList(t *testing.T) {
	t.Parallel()
	accepted := []string{"DRECEIPT:accepted", "DRESULT:succeeded"}
	refusal := []string{"DRECEIPT:refusedpolicy_not_allowed"}
	cases := []struct {
		name string
		list []string
		sent string
		want []string
	}{
		{"a listed digest", []string{policy}, policy, accepted},
		{"one of several, in any order", []string{digestOf("f"), policy, digestOf("a")}, policy, accepted},
		{"a digest that is not listed", []string{policy}, digestOf("e"), refusal},
		{"an empty list", nil, policy, refusal},
		{"the reserved stop digest, which no list can hold", []string{policy}, digestOf("0"), refusal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, answers := commandSession(t, 2, meterCommand(tc.sent), func(o *Options) { o.Policies = tc.list })
			if got := answers["c-1"]; !slices.Equal(got, tc.want) {
				t.Fatalf("answers = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAnApprovedCommandIsStillCheckedAgainstTheList(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		sent  string
		want  []string
		calls int
	}{
		{"a listed digest runs once approved", policy, []string{"DRECEIPT:queued", "DRECEIPT:accepted", "DRESULT:succeeded"}, 1},
		{"an unlisted digest is refused once approved", digestOf("e"), []string{"DRECEIPT:queued", "DRECEIPT:refusedpolicy_not_allowed"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := newControlled(t, nil)
			c.host.afterSync = func(h *scriptedHost, syncs int) {
				switch syncs {
				case 1:
					h.reply("DCMD", set(fanCommand("c-1", sliceEnd(1)), "policy_digest", tc.sent))
				case 2:
					c.submit(t, `{"op":"approve","command_id":"c-1"}`)
				}
			}
			c.submit(t, approveFan)
			c.play(t, 3)
			if got := c.answers()["c-1"]; !slices.Equal(got, tc.want) {
				t.Fatalf("answers = %v, want %v", got, tc.want)
			}
			if calls := c.run.World.EffectorCalls(); len(calls) != tc.calls {
				t.Fatalf("effector calls = %d, want %d", len(calls), tc.calls)
			}
		})
	}
}

func TestAnAllowListThatBreaksTheRuleIsBadNodeOptionsAndTheSessionNeverStarts(t *testing.T) {
	t.Parallel()
	nine := make([]string, 9)
	for i := range nine {
		nine[i] = digestOf(string("abcdef012"[i]))
	}
	cases := []struct {
		name string
		list []string
		want string
	}{
		{"a malformed digest", []string{"nope"}, "64 lower-case hexadecimal digits"},
		{"a repeat", []string{policy, policy}, "already listed"},
		{"the reserved stop digest", []string{digestOf("0")}, "reserved"},
		{"nine digests", nine, "at most 8"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			options := sliceOptions(1)
			options.Policies = tc.list
			link := &scriptedHost{t: t}
			err := NewSession(utilityRun(t), link, options).Run(t.Context())
			if !errors.Is(err, ErrBadOptions) || !strings.Contains(err.Error(), tc.want) || len(link.sent) != 0 {
				t.Fatalf("Run = %v after %d records, want ErrBadOptions saying %q before anything is sent", err, len(link.sent), tc.want)
			}
		})
	}
}

func TestAHostThatIssuesCommandsMustCarryAWellFormedPolicyDigest(t *testing.T) {
	t.Parallel()
	script := []host.Scripted{{ID: "c-1", AtNS: 1, Device: "fan-01", Command: "set_pwm_lease"}}
	cases := []struct {
		name    string
		options host.Options
		refused bool
	}{
		{"no script and no policy", host.Options{Environment: "simulated"}, false},
		{"a script with a policy", host.Options{Environment: "simulated", Script: script, Policy: policy}, false},
		{"a script and no policy", host.Options{Environment: "simulated", Script: script}, true},
		{"a script and a malformed policy", host.Options{Environment: "simulated", Script: script, Policy: "sha256:00"}, true},
		{"a script and the reserved stop digest", host.Options{Environment: "simulated", Script: script, Policy: digestOf("0")}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := CheckHost(tc.options)
			if tc.refused != errors.Is(err, ErrHostPolicy) || (!tc.refused && err != nil) {
				t.Fatalf("CheckHost = %v, want a refusal: %v", err, tc.refused)
			}
		})
	}
}

func TestServingAScriptedHostWithoutAPolicyIsRefusedBeforeItListens(t *testing.T) {
	t.Parallel()
	options := host.Options{Environment: "simulated", Script: []host.Scripted{{ID: "c-1", AtNS: 1, Device: "fan-01", Command: "set_pwm_lease"}}}
	report, err := ServeHost(t.Context(), t.TempDir()+"/h.sock", 0, options)
	if !errors.Is(err, ErrHostPolicy) || len(report.Commands) != 0 {
		t.Fatalf("ServeHost = %+v, %v, want ErrHostPolicy and a host that never listened", report, err)
	}
}

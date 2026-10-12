package app

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestReplayNamesAnArtifactThatCannotBeRead(t *testing.T) {
	t.Parallel()
	got := invoke(t, "replay", filepath.Join(t.TempDir(), "absent.json"))
	if got.code != 1 || !strings.Contains(got.stderr, "run: read artifact") {
		t.Fatalf("replay of a missing artifact: %+v", got)
	}
}

func TestRunRefusesAnUnknownAdapterAndANonexistentEffector(t *testing.T) {
	t.Parallel()
	base := []string{"run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--domain", pondDomain, "--duration", "60"}
	unknown := invoke(t, append(base, "--adapter", "no-such")...)
	if unknown.code != 1 || !strings.Contains(unknown.stderr, `unknown adapter "no-such"`) {
		t.Fatalf("unknown adapter: %+v", unknown)
	}
	effector := invoke(t, append(base, "--effector", "no_such@site-a/pond-1@1")...)
	want := `--effector "no_such@site-a/pond-1@1": the domain declares no effector "no_such"`
	if effector.code != 2 || !strings.Contains(effector.stderr, want) {
		t.Fatalf("unknown effector: %+v", effector)
	}
	issues, _ := effector.failure(t)["issues"].([]any)
	issue, _ := issues[0].(map[string]any)
	if _, located := issue["pointer"]; located || issue["message"] != want {
		t.Fatalf("the issue must name the flag, not point into a document the user never wrote: %v", issues)
	}
}

// Flag-built scenario issues name the flag behind them: the effector's
// arguments, a perturbation's window, the profile.
func TestFlagIssuesNameTheFlagBehindThem(t *testing.T) {
	t.Parallel()
	base := []string{"run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--domain", pondDomain, "--duration", "60"}
	cases := []struct{ name, flag, value, want string }{
		{"effector args", "--effector", `start_aerator@site-a/pond-1@1@{}`, `--effector "start_aerator@site-a/pond-1@1@{}": effector "start_aerator" refuses these args:`},
		{"perturbation window", "--perturb", "drop@30@10", `--perturb "drop@30@10": drop window would end at`},
		{"profile", "--profile", "no-such-profile", `--profile: the domain has no profile "no-such-profile"`},
		{"fault", "--fault", "site-a/pond-1=no_such@1", `--fault "site-a/pond-1=no_such@1": the domain declares no fault "no_such"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, append(base, tc.flag, tc.value)...)
			issues, _ := got.failure(t)["issues"].([]any)
			if got.code != 2 || len(issues) != 1 || !strings.Contains(issues[0].(map[string]any)["message"].(string), tc.want) {
				t.Fatalf("%s: %+v, want one issue containing %q", tc.flag, got, tc.want)
			}
		})
	}
}

func TestRefconsumerAndScoreNameWhatTheyCannotRead(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "absent")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"missing trace", []string{"refconsumer", "--trace", missing}, "no such file or directory"},
		{"missing run artifact", []string{"score", "--run", filepath.Join(missing, "run.json")}, "run: read artifact " + filepath.Join(missing, "run.json")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if got.code != 1 || !strings.Contains(got.stderr, tc.want) {
				t.Fatalf("%v: %+v, want %q", tc.args, got, tc.want)
			}
		})
	}
}

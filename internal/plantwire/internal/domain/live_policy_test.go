package domain

import "testing"

func policyCommand(digest string) map[string]any {
	record := liveRecord(hostCommand())
	record["policy_digest"] = digest
	return record
}

func TestARefusalForPolicyCarriesTheDigestTheHostSentAndNothingElseDoes(t *testing.T) {
	t.Parallel()
	sent := digestOf("e")
	cases := []struct {
		name, status, reason, want string
	}{
		{"a policy refusal", "refused", "policy_not_allowed", sent},
		{"another refusal", "refused", "stale_epoch", ""},
		{"an acceptance", "accepted", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			l := liveBoard(t)
			l.Observe(TraceHost, policyCommand(sent), 2)
			answer := liveAnswer("DRECEIPT", "c-1", "boot-1", tc.status)
			answer["reason"] = tc.reason
			l.Observe(TraceNode, answer, 3)
			if c := l.command("c-1"); c.PolicyDigest != tc.want {
				t.Fatalf("policy digest = %q, want %q", c.PolicyDigest, tc.want)
			}
		})
	}
}

func TestAPendingCommandShowsNoPolicyDigestAndAMalformedOneIsNeverCarried(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Observe(TraceHost, policyCommand(digestOf("e")), 2)
	if c := l.command("c-1"); c.PolicyDigest != "" {
		t.Fatalf("a command without a receipt carries %q", c.PolicyDigest)
	}
	forged := liveBoard(t)
	forged.Observe(TraceHost, policyCommand("the host's own words"), 2)
	answer := liveAnswer("DRECEIPT", "c-1", "boot-1", "refused")
	answer["reason"] = "policy_not_allowed"
	forged.Observe(TraceNode, answer, 3)
	if c := forged.command("c-1"); c.PolicyDigest != "" || c.Reason != "policy_not_allowed" {
		t.Fatalf("command = %+v, want the refusal without a digest that is no digest", c)
	}
}

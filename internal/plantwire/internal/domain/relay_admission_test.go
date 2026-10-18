package domain

import "testing"

func TestAnAdmissionRefusesWithTheHostsReasonInTheHostsOrder(t *testing.T) {
	t.Parallel()
	hardware := Admission{Environment: EnvironmentHardware, ProtocolMax: 1}
	cases := []struct {
		name           string
		admission      Admission
		message        func() Message
		uncommissioned bool
		want           string
	}{
		{"a birth that holds", hardware, relayBirthMessage, false, ""},
		{"a commissioned device missing from the birth", hardware, relayBirthMessage, true, "inventory_mismatch"},
		{"a host set to another environment", Admission{Environment: EnvironmentSimulated, ProtocolMax: 1}, relayBirthMessage, false, "environment_mismatch"},
		{"a host that speaks no version the node does", Admission{Environment: EnvironmentHardware, ProtocolMax: 0}, relayBirthMessage, false, "version_unsupported"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			message := tc.message()
			if got := tc.admission.Refusal(message, tc.uncommissioned); got != tc.want {
				t.Fatalf("Refusal = %q, want %q", got, tc.want)
			}
			if reason := tc.admission.Reason(tc.want, message); (tc.want == "environment_mismatch") != (reason != "") {
				t.Fatalf("Reason = %q for %q, want a reason for an environment refusal only", reason, tc.want)
			}
		})
	}
}

func relayBirthMessage() Message {
	return Message{Type: "NBIRTH", ProtocolMin: 1, ProtocolMax: 1, Environment: EnvironmentHardware, ContractVersion: ContractVersion,
		ContractDigest: ContractDigest(), Boards: []Board{{Board: "b-1", Environment: EnvironmentHardware}}}
}

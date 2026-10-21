package domain

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func boardsOf(environments ...string) []Board {
	boards := make([]Board, len(environments))
	for i, environment := range environments {
		boards[i] = Board{Board: fmt.Sprintf("b-%d", i+1), Environment: environment}
	}
	return boards
}

func TestAHostAcceptsOnlyTheNodesOfItsEnvironmentAndAMixedHostAcceptsAllThree(t *testing.T) {
	t.Parallel()
	const sim, hw, mixed = EnvironmentSimulated, EnvironmentHardware, EnvironmentMixed
	cases := []struct {
		name    string
		host    string
		node    string
		boards  []Board
		pointer string
		reason  string
	}{
		{"a simulated host and a simulated node", sim, sim, boardsOf(sim, sim), "", ""},
		{"a hardware host and a hardware node", hw, hw, boardsOf(hw), "", ""},
		{"a mixed host and a simulated node", mixed, sim, boardsOf(sim), "", ""},
		{"a mixed host and a hardware node", mixed, hw, boardsOf(hw, hw), "", ""},
		{"a mixed host and a mixed node", mixed, mixed, boardsOf(sim, hw), "", ""},
		{"a node without boards claims simulated vacuously", sim, sim, nil, "", ""},
		{"a simulated host refuses a hardware node", sim, hw, boardsOf(hw), "/environment", "the host accepts simulated nodes only, the node announces hardware"},
		{"a hardware host refuses a simulated node", hw, sim, boardsOf(sim), "/environment", "the host accepts hardware nodes only, the node announces simulated"},
		{"a simulated host refuses a mixed node", sim, mixed, boardsOf(sim, hw), "/environment", "the host accepts simulated nodes only, the node announces mixed"},
		{"a hardware host refuses a mixed node", hw, mixed, boardsOf(sim, hw), "/environment", "the host accepts hardware nodes only, the node announces mixed"},
		{"a host set to nothing known accepts nothing", "lab", sim, boardsOf(sim), "/environment", "the host accepts lab nodes only, the node announces simulated"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertFault(t, EnvironmentFaultFor(tc.host, Message{Environment: tc.node, Boards: tc.boards}), tc.pointer, tc.reason)
		})
	}
}

func TestANodeThatContradictsItsBoardsIsRefusedByEveryHostEvenAMixedOne(t *testing.T) {
	t.Parallel()
	const sim, hw, mixed = EnvironmentSimulated, EnvironmentHardware, EnvironmentMixed
	cases := []struct {
		name    string
		host    string
		node    string
		boards  []Board
		pointer string
		reason  string
	}{
		{"a simulated node with a hardware board", mixed, sim, boardsOf(sim, hw), "/boards/1/environment", `node environment simulated needs every board simulated, board "b-2" is "hardware"`},
		{"a hardware node with a simulated board", mixed, hw, boardsOf(sim), "/boards/0/environment", `node environment hardware needs every board hardware, board "b-1" is "simulated"`},
		{"a mixed node with only simulated boards", mixed, mixed, boardsOf(sim, sim), "/boards", "node environment mixed needs a simulated and a hardware board, none of its 2 boards is hardware"},
		{"a mixed node with only hardware boards", mixed, mixed, boardsOf(hw), "/boards", "node environment mixed needs a simulated and a hardware board, none of its 1 boards is simulated"},
		{"a mixed node without boards", mixed, mixed, nil, "/boards", "none of its 0 boards is simulated"},
		{"a node of an environment that does not exist", mixed, "staging", boardsOf(sim), "/environment", `environment "staging" is none of simulated, hardware, mixed`},
		{"the contradiction is named before the host's setting", sim, hw, boardsOf(sim), "/boards/0/environment", `node environment hardware needs every board hardware`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertFault(t, EnvironmentFaultFor(tc.host, Message{Environment: tc.node, Boards: tc.boards}), tc.pointer, tc.reason)
		})
	}
}

func assertFault(t *testing.T, fault *EnvironmentFault, pointer, reason string) {
	t.Helper()
	switch {
	case reason == "" && fault != nil:
		t.Fatalf("fault = %+v, want the environment accepted", *fault)
	case reason != "" && (fault == nil || fault.Pointer != pointer || !strings.Contains(fault.Reason, reason)):
		t.Fatalf("fault = %+v, want %s saying %q", fault, pointer, reason)
	}
}

func TestACaptureAppliesTheNodesOwnEnvironmentRulesWithoutAHost(t *testing.T) {
	t.Parallel()
	message := Message{Environment: EnvironmentHardware, Boards: boardsOf(EnvironmentSimulated), ContractVersion: ContractVersion, ContractDigest: ContractDigest()}
	issues := claimIssues(message)
	if len(issues) != 1 || issues[0].Keyword != "environment_mismatch" || issues[0].Pointer != "/boards/0/environment" || issues[0].Member != "environment" {
		t.Fatalf("claimIssues = %+v, want one environment_mismatch at /boards/0/environment", issues)
	}
	message.Boards = boardsOf(EnvironmentHardware)
	if issues := claimIssues(message); len(issues) != 0 {
		t.Fatalf("claimIssues of a consistent node = %+v, want none", issues)
	}
	message.ContractVersion = ""
	if issues := claimIssues(message); len(issues) != 1 || issues[0].Keyword != "contract_unsupported" || issues[0].Pointer != "/contract_version" {
		t.Fatalf("claimIssues without a contract version = %+v, want one contract_unsupported at /contract_version", issues)
	}
}

func TestTheEnvironmentAHostIsSetToIsOneOfThree(t *testing.T) {
	t.Parallel()
	for _, environment := range []string{"simulated", "hardware", "mixed"} {
		if err := CheckEnvironment(environment); err != nil {
			t.Errorf("CheckEnvironment(%q) = %v", environment, err)
		}
	}
	for _, environment := range []string{"", "lab", "Simulated", "simulated,hardware"} {
		if err := CheckEnvironment(environment); !errors.Is(err, ErrEnvironment) || !strings.Contains(err.Error(), "none of simulated, hardware, mixed") {
			t.Errorf("CheckEnvironment(%q) = %v, want ErrEnvironment naming the three environments", environment, err)
		}
	}
}

func TestABirthNamesTheEnvironmentOfEveryBoardAndADomainBoardIsSimulated(t *testing.T) {
	t.Parallel()
	boards := []model.Board{{Name: "from-a-domain", Members: []string{"P1"}}, {Name: "behind-a-gateway", Members: []string{"P2"}, Environment: EnvironmentHardware}}
	birth, err := NodeBirthOf(NodeIdentity{Environment: EnvironmentMixed, ProtocolMax: 1, PriorLastSeq: -1}, birthKinds(), boards)
	if err != nil {
		t.Fatal(err)
	}
	record := stamped(t, birth)[0]
	listed, _ := record["boards"].([]any)
	if len(listed) != 2 || listed[0].(map[string]any)["environment"] != "simulated" || listed[1].(map[string]any)["environment"] != "hardware" || record["environment"] != "mixed" {
		t.Fatalf("NBIRTH = %v, want a mixed node whose domain board is simulated and whose gateway board is hardware", record)
	}
}

func TestTheAnnouncedEnvironmentFollowsTheBoardsActuallyAnnounced(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		boards map[string]bool
		want   string
	}{
		{map[string]bool{EnvironmentSimulated: true}, EnvironmentSimulated},
		{map[string]bool{EnvironmentHardware: true}, EnvironmentHardware},
		{map[string]bool{EnvironmentSimulated: true, EnvironmentHardware: true}, EnvironmentMixed},
		{map[string]bool{EnvironmentSimulated: false, EnvironmentHardware: true}, EnvironmentHardware},
	} {
		if got := AnnouncedEnvironment(tc.boards); got != tc.want {
			t.Errorf("AnnouncedEnvironment(%v) = %q, want %q", tc.boards, got, tc.want)
		}
	}
}

func TestAMergedBirthClaimsOnlyTheEnvironmentsOfItsBoards(t *testing.T) {
	t.Parallel()
	hardware := map[string]any{"environment": EnvironmentHardware, "boards": []any{map[string]any{"board": "uno-07", "environment": EnvironmentHardware}}}
	for _, tc := range []struct {
		simulated []any
		want      string
	}{
		{nil, EnvironmentHardware},
		{[]any{map[string]any{"board": "mega-01", "environment": EnvironmentSimulated}}, EnvironmentMixed},
	} {
		merged := mergedBirth(hardware, map[string]any{"environment": EnvironmentSimulated, "boards": tc.simulated})
		if merged["environment"] != tc.want || len(merged["boards"].([]any)) != 1+len(tc.simulated) {
			t.Errorf("merged birth with simulated boards %v = %v, want a %s birth over every board", tc.simulated, merged, tc.want)
		}
	}
}

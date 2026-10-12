package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestClassificationFollowsTheDocumentedPrecedence(t *testing.T) {
	t.Parallel()
	missing := &fs.PathError{Op: "open", Path: "x.json", Err: fs.ErrNotExist}
	denied := &fs.PathError{Op: "open", Path: "x.json", Err: fs.ErrPermission}
	invalid := &model.ValidationError{Contract: model.ContractRunArtifact, Summary: "s", Cause: missing}
	cases := []struct {
		name string
		err  error
		want failureCode
		exit int
	}{
		{"flag usage over everything", fmt.Errorf("x: %w", usageExit{code: 2, message: "bad flag"}), codeUsageInvalid, 2},
		{"argument usage", usagef("needs --domain"), codeUsageInvalid, 2},
		{"usage over an outcome", fmt.Errorf("%w: %w", usagef("u"), &outcomeError{code: codeReplayMismatch}), codeUsageInvalid, 2},
		{"outcome over a contract violation", fmt.Errorf("%w: %w", &outcomeError{code: codeAdapterVerifyFailed}, invalid), codeAdapterVerifyFailed, 1},
		{"digest mismatch is a replay mismatch", fmt.Errorf("run: world %w", run.ErrDigestMismatch), codeReplayMismatch, 1},
		{"contract violation over its missing-file cause", invalid, codeArtifactInvalid, 1},
		{"unknown domain", fmt.Errorf("catalog: %w", domain.ErrUnknownDomain), codeDomainNotFound, 1},
		{"unknown adapter", fmt.Errorf("%w", adapter.ErrUnknownAdapter), codeAdapterNotFound, 1},
		{"missing path", missing, codeFileNotFound, 1},
		{"other file failure", denied, codeIOFailed, 1},
		{"anything else", errors.New("boom"), "score_failed", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifyFailure("score", tc.err); got != tc.want {
				t.Fatalf("classifyFailure = %s, want %s", got, tc.want)
			}
			if got := exitStatusOf(tc.err); got != tc.exit {
				t.Fatalf("exitStatusOf = %d, want %d", got, tc.exit)
			}
		})
	}
}

func TestEveryCommandFailureCodeIsInTheContract(t *testing.T) {
	t.Parallel()
	declared := cliOutputCodes(t)
	codes := []failureCode{codeUsageInvalid, codeFileNotFound, codeDomainNotFound, codeAdapterNotFound, codeDomainInvalid,
		codeAdapterInvalid, codeArtifactInvalid, codeReplayMismatch, codeAdapterVerifyFailed, codeIOFailed}
	for name, registered := range commands {
		if !registered.ownsStdout {
			codes = append(codes, commandFailed(name))
		}
	}
	for _, code := range codes {
		if !slices.Contains(declared, string(code)) {
			t.Errorf("code %s is not in the cli-output error.code enum", code)
		}
	}
	if len(codes) != len(declared) {
		t.Errorf("the CLI produces %d codes and the contract declares %d: %v vs %v", len(codes), len(declared), codes, declared)
	}
}

func TestEveryContractMapsToAFailureCode(t *testing.T) {
	t.Parallel()
	for _, contract := range declaredContracts(t) {
		if _, ok := contractFailureCode(contract); !ok {
			t.Errorf("contract %s has no failure code", contract)
		}
	}
}

// cliOutputCodes is the error.code enum of cli-output-v0.1.
func cliOutputCodes(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.Contract("cli-output-v0.1"))
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		Defs struct {
			Error struct {
				Properties struct {
					Error struct {
						Properties struct {
							Code struct {
								Enum []string `json:"enum"`
							} `json:"code"`
						} `json:"properties"`
					} `json:"error"`
				} `json:"properties"`
			} `json:"error"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatal(err)
	}
	return contract.Defs.Error.Properties.Error.Properties.Code.Enum
}

// declaredContracts parses every model.Contract constant from the model
// package source, so a new contract cannot be added without a code.
func declaredContracts(t *testing.T) []model.Contract {
	t.Helper()
	source := filepath.Join(testsupport.RepositoryRoot(), "internal", "model", "validation.go")
	file, err := parser.ParseFile(token.NewFileSet(), source, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var contracts []model.Contract
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || len(spec.Values) != 1 {
			return true
		}
		ident, typed := spec.Type.(*ast.Ident)
		literal, isLiteral := spec.Values[0].(*ast.BasicLit)
		if typed && isLiteral && ident.Name == "Contract" {
			contracts = append(contracts, model.Contract(literal.Value[1:len(literal.Value)-1]))
		}
		return true
	})
	if len(contracts) < 6 {
		t.Fatalf("found only %d Contract constants in %s", len(contracts), source)
	}
	return contracts
}

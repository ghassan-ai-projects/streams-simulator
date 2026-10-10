package architecture

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// snippetFacade parses a facade source for the fictional module internal/m
// whose layer is internal/m/internal/domain.
func snippetFacade(t *testing.T, path, source string) productionFile {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, source, parser.ParseComments)
	if err != nil {
		t.Fatalf("snippet does not parse: %v", err)
	}
	return productionFile{path: path, pkgDir: "internal/m", source: parsed, imports: importNames(parsed), paths: importPaths(parsed)}
}

const snippetHeader = `package m

import layer "` + modulePrefix + `internal/m/internal/domain"
`

func TestFacadeRulesAcceptADelegatingFacade(t *testing.T) {
	t.Parallel()
	source := snippetHeader + `
type Record = layer.Record

const Kind = layer.Kind

var ErrGone = layer.ErrGone

type Service struct{ impl *layer.Service }

func New(dependency *Dependency) (*Service, error) {
	if dependency == nil {
		return nil, ErrGone
	}
	return &Service{impl: layer.New(dependency)}, nil
}

func (s *Service) Do(id string) error { return s.impl.Do(id) }

func (s *Service) Guarded(id string) error {
	if s == nil {
		return ErrGone
	}
	return s.impl.Do(id)
}

func Parse(raw []byte) (Record, error) { return layer.Parse(raw, string(raw)) }
`
	noMethods := func(string, string) bool { return false }
	if got := facadeViolations(snippetFacade(t, "internal/m/api.go", source), "internal/m", noMethods); len(got) != 0 {
		t.Fatalf("violations = %v", got)
	}
}

func TestFacadeRulesRejectEveryKnownBypass(t *testing.T) {
	t.Parallel()
	hasMethods := func(path, typeName string) bool { return typeName == "Aggregate" }
	for name, tc := range map[string]struct {
		source string
		want   string
	}{
		"loop in an exported method": {
			`type S struct{ impl *layer.S }
func (s *S) Count() int { n := 0; for range []int{1} { n++ }; return n }`,
			"does more than delegate"},
		"unexported helper with logic": {
			`type S struct{ impl *layer.S }
func (s *S) Count() int { return s.count() }
func (s *S) count() int { n := 0; for range []int{1} { n++ }; return n }`,
			"does more than delegate"},
		"function literal argument": {
			`type S struct{ impl *layer.S }
func (s *S) Do() string { return s.impl.Do(func() string { return "x" }()) }`,
			"does more than delegate"},
		"nested call argument": {
			`type S struct{ impl *layer.S }
func (s *S) Do(a string) string { return s.impl.Do(compute(a)) }
func compute(a string) string { return a }`,
			"does more than delegate"},
		"guarded delegation that repairs instead of refusing": {
			`type S struct{ impl *layer.S }
func (s *S) Do(id string) error {
	if id == "" {
		id = "default"
	}
	return s.impl.Do(id)
}`,
			"does more than delegate"},
		"guard whose condition calls a function": {
			`type S struct{ impl *layer.S }
func New(id string) (*S, error) {
	if lookup(id) != nil {
		return nil, ErrGone
	}
	return &S{impl: layer.New(id)}, nil
}
func lookup(id string) *S { return nil }`,
			"does more than delegate"},
		"constructor returning a computed field": {
			`import "strings"
type S struct{ ID string; impl *layer.S }
func New(inner *layer.S) *S { return &S{ID: strings.ToUpper(inner.ID), impl: inner} }`,
			"does more than delegate"},
		"constructor with two definitions": {
			`type S struct{ impl *layer.S }
func New() *S {
	a := layer.New()
	b := layer.New()
	return &S{impl: a, other: b}
}`,
			"does more than delegate"},
		"delegation with an arithmetic argument": {
			`type S struct{ impl *layer.S }
func (s *S) Advance(to int64) int64 { return s.impl.Advance(to + 1) }`,
			"does more than delegate"},
		"delegation with an indexed argument": {
			`type S struct{ impl *layer.S }
func (s *S) Do(m map[string]string, k string) string { return s.impl.Do(m[k]) }`,
			"does more than delegate"},
		"constructor with a loop": {
			`type S struct{ impl *layer.S }
func New() *S { for { break }; return &S{impl: layer.New()} }`,
			"does more than delegate"},
		"constructor guard that repairs instead of refusing": {
			`type S struct{ impl *layer.S }
func New(check func(string) bool) *S {
	if check == nil {
		check = func(string) bool { return true }
	}
	return &S{impl: layer.New(check)}
}`,
			"does more than delegate"},
		"constructor returning computed arguments": {
			`type S struct{ impl *layer.S }
func New(seed uint64) *S { return &S{impl: layer.New(seed + 1)} }`,
			"does more than delegate"},
		"delegating through a receiver field that holds no layer object": {
			`type S struct{ other *Other; impl *layer.S }
type Other struct{}
func (o *Other) Stop() {}
func (s *S) Stop() { s.other.Stop() }`,
			"does more than delegate"},
		"exported variable with inferred layer type": {
			`var Raw = layer.New`,
			"exported variable Raw"},
		"alias of an aggregate with methods": {
			`type Whole = layer.Aggregate`,
			"would export the methods of Aggregate"},
		"unexported alias of a layer type": {
			`type inner = layer.Record
func Make() *inner { return layer.Make() }`,
			"must be exported and declared in api.go"},
		"exported type defined over a layer type": {
			`type Public layer.Record`,
			"exposes an internal layer type"},
		"embedding a layer type in an exported struct": {
			`type S struct{ layer.Record }`,
			"exposes an internal layer type"},
		"layer type in an exported signature": {
			`type S struct{ impl *layer.S }
func (s *S) Inner() *layer.S { return s.impl }`,
			"names an internal layer in its signature"},
		"init in a facade": {
			`func init() { layer.Register() }`,
			"init is not allowed"},
		"dot import of a layer": {
			``,
			"dot or blank name"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source := snippetHeader + tc.source
			path := "internal/m/api.go"
			if name == "unexported alias of a layer type" {
				source = snippetHeader + tc.source
			}
			if name == "dot import of a layer" {
				source = "package m\nimport . \"" + modulePrefix + "internal/m/internal/domain\"\n"
			}
			violations := facadeViolations(snippetFacade(t, path, source), "internal/m", hasMethods)
			if !strings.Contains(strings.Join(violations, "\n"), tc.want) {
				t.Fatalf("violations = %v, want one containing %q", violations, tc.want)
			}
		})
	}
}

func TestFacadeRulesKeepAliasesInAPIFile(t *testing.T) {
	t.Parallel()
	source := snippetHeader + "type Record = layer.Record\n"
	noMethods := func(string, string) bool { return false }
	got := facadeViolations(snippetFacade(t, "internal/m/service.go", source), "internal/m", noMethods)
	if !strings.Contains(strings.Join(got, "\n"), "declared in api.go") {
		t.Fatalf("violations = %v", got)
	}
}

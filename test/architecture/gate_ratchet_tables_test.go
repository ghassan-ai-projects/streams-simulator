package architecture

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// tableEntry is what one entry of a gate table lets through: its numeric
// limits (line caps) and its allowed items (imports, packages).
type tableEntry struct {
	limits []int
	items  []string
}

type goSource struct {
	file   *ast.File
	consts map[string]string
}

func parseGoSource(source []byte) (goSource, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "gate.go", source, 0)
	if err != nil {
		return goSource{}, fmt.Errorf("parse gate source: %w", err)
	}
	return goSource{file: file, consts: stringConstants(file)}, nil
}

func stringConstants(file *ast.File) map[string]string {
	consts := map[string]string{}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST {
			continue
		}
		for _, spec := range general.Specs {
			value := spec.(*ast.ValueSpec)
			for i, name := range value.Names {
				if i < len(value.Values) {
					consts[name.Name] = literalString(value.Values[i])
				}
			}
		}
	}
	return consts
}

func literalString(expr ast.Expr) string {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return ""
	}
	text, err := strconv.Unquote(literal.Value)
	if err != nil {
		return ""
	}
	return text
}

func (s goSource) variable(name string) (*ast.CompositeLit, bool) {
	for _, declaration := range s.file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			value := spec.(*ast.ValueSpec)
			if len(value.Names) == 1 && value.Names[0].Name == name && len(value.Values) == 1 {
				literal, isLiteral := value.Values[0].(*ast.CompositeLit)
				return literal, isLiteral
			}
		}
	}
	return nil, false
}

func (s goSource) text(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return cmp.Or(s.consts[ident.Name], ident.Name)
	}
	return literalString(expr)
}

func (s goSource) table(name string) (map[string]tableEntry, bool) {
	literal, found := s.variable(name)
	if !found {
		return nil, false
	}
	entries := map[string]tableEntry{}
	for _, element := range literal.Elts {
		pair := element.(*ast.KeyValueExpr)
		entries[s.text(pair.Key)] = s.entry(pair.Value)
	}
	return entries, true
}

func (s goSource) entry(value ast.Expr) tableEntry {
	if text := literalString(value); text != "" {
		return tableEntry{items: strings.Fields(text)}
	}
	var entry tableEntry
	record, ok := value.(*ast.CompositeLit)
	if !ok {
		return entry
	}
	for _, element := range record.Elts {
		switch e := element.(type) {
		case *ast.BasicLit:
			if limit, err := strconv.Atoi(e.Value); err == nil && e.Kind == token.INT {
				entry.limits = append(entry.limits, limit)
			}
		case *ast.CompositeLit:
			entry.items = append(entry.items, s.items(e)...)
		}
	}
	return entry
}

func (s goSource) items(list *ast.CompositeLit) []string {
	var items []string
	for _, element := range list.Elts {
		items = append(items, s.text(element))
	}
	return items
}

func (s goSource) list(name string) ([]string, bool) {
	literal, found := s.variable(name)
	if !found {
		return nil, false
	}
	return s.items(literal), true
}

func tableViolations(label string, base, current map[string]tableEntry) []string {
	var violations []string
	for _, key := range sortedKeys(current) {
		before, existed := base[key]
		if existed {
			violations = append(violations, entryViolations(label+" "+key, before, current[key])...)
		}
	}
	return violations
}

func entryViolations(label string, before, now tableEntry) []string {
	var violations []string
	for i := 0; i < min(len(before.limits), len(now.limits)); i++ {
		if now.limits[i] > before.limits[i] {
			violations = append(violations, fmt.Sprintf("%s: cap raised from %d to %d", label, before.limits[i], now.limits[i]))
		}
	}
	for _, item := range now.items {
		if !slices.Contains(before.items, item) {
			violations = append(violations, fmt.Sprintf("%s: allowed import %q added", label, item))
		}
	}
	return violations
}

func listViolations(label string, base, current []string) []string {
	var violations []string
	for _, item := range current {
		if !slices.Contains(base, item) {
			violations = append(violations, fmt.Sprintf("%s: allowed entry %q added", label, item))
		}
	}
	return violations
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	// determinism-safe: the collected keys are sorted below.
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// goGateViolations compares one Go gate file with its base version: the named
// tables and lists must exist in both and must not have loosened.
func goGateViolations(file string, base, current []byte, tables, lists []string) ([]string, error) {
	if len(base) == 0 {
		return nil, nil
	}
	before, err := parseGoSource(base)
	if err != nil {
		return nil, err
	}
	now, err := parseGoSource(current)
	if err != nil {
		return nil, err
	}
	var violations []string
	for _, name := range tables {
		violations = append(violations, compareTable(file, name, before, now)...)
	}
	for _, name := range lists {
		violations = append(violations, compareList(file, name, before, now)...)
	}
	return violations, nil
}

func compareTable(file, name string, before, now goSource) []string {
	old, existed := before.table(name)
	if !existed {
		return nil
	}
	current, found := now.table(name)
	if !found {
		return []string{file + ": table " + name + " no longer found; update the gate ratchet with it"}
	}
	return tableViolations(file+" "+name, old, current)
}

func compareList(file, name string, before, now goSource) []string {
	old, existed := before.list(name)
	if !existed {
		return nil
	}
	current, found := now.list(name)
	if !found {
		return []string{file + ": list " + name + " no longer found; update the gate ratchet with it"}
	}
	return listViolations(file+" "+name, old, current)
}

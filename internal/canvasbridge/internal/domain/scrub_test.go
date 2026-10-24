package domain

import "testing"

func TestScrubKeepsPathBoundaries(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		text  string
		roots []string
		want  string
	}{
		"prefix stripped":              {"open /ws/a/notes/x.json: denied", []string{"/ws/a"}, "open notes/x.json: denied"},
		"root alone at end":            {"in /ws/a", []string{"/ws/a"}, "in ."},
		"root alone before a space":    {"in /ws/a is bad", []string{"/ws/a"}, "in . is bad"},
		"root with a trailing period":  {"failed in /ws/a.", []string{"/ws/a"}, "failed in .."},
		"sibling sharing the prefix":   {"/ws/ab/x and /ws/a-b", []string{"/ws/a"}, "/ws/ab/x and /ws/a-b"},
		"root as a file stem":          {"/ws/a.sock", []string{"/ws/a"}, "/ws/a.sock"},
		"root inside a longer path":    {"/other/ws/a/x", []string{"/ws/a"}, "/other/ws/a/x"},
		"empty roots":                  {"/ws/a/x", nil, "/ws/a/x"},
		"empty root":                   {"/ws/a/x", []string{""}, "/ws/a/x"},
		"after a quote":                {`{"composed":"/ws/a/d.json"}`, []string{"/ws/a"}, `{"composed":"d.json"}`},
		"after a colon and space":      {"error: /ws/a/d.json", []string{"/ws/a"}, "error: d.json"},
		"after an equals sign":         {"dir=/ws/a/d.json", []string{"/ws/a"}, "dir=d.json"},
		"after a bracket and comma":    {"[/ws/a/x,/ws/a/y]", []string{"/ws/a"}, "[x,y]"},
		"every occurrence":             {"/ws/a/x /ws/a", []string{"/ws/a"}, "x ."},
		"not the root of the sibling":  {"/ws/ab/x /ws/a/y", []string{"/ws/a"}, "/ws/ab/x y"},
		"symlinked form alongside":     {"/private/tmp/ws/sockets/a.sock", []string{"/tmp/ws", "/private/tmp/ws"}, "sockets/a.sock"},
		"symlinked form, short first":  {"/private/tmp/ws/sockets/a.sock", []string{"/private/tmp/ws", "/tmp/ws"}, "sockets/a.sock"},
		"configured form still scrubs": {"/tmp/ws/sockets/a.sock", []string{"/tmp/ws", "/private/tmp/ws"}, "sockets/a.sock"},
		"resolved form not configured": {"/private/tmp/ws/sockets/a.sock", []string{"/tmp/ws"}, "/private/tmp/ws/sockets/a.sock"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := Scrub(tc.text, tc.roots); got != tc.want {
				t.Errorf("Scrub(%q, %q) = %q, want %q", tc.text, tc.roots, got, tc.want)
			}
		})
	}
}

func TestScrubAuditFScrubDoesNotEatTheMiddleOfAResolvedPath(t *testing.T) {
	t.Parallel()
	const root = "/tmp/streamsim-ui-audit-20261024"
	text := "dial unix /private" + root + "/sockets/absent.sock: connect: no such file or directory"
	want := "dial unix sockets/absent.sock: connect: no such file or directory"
	if got := Scrub(text, []string{root, "/private" + root}); got != want {
		t.Errorf("audit F-scrub: Scrub = %q, want %q", got, want)
	}
	if got := Scrub(text, []string{root}); got != "dial unix /private"+root+"/sockets/absent.sock: connect: no such file or directory" {
		t.Errorf("audit F-scrub: with only the configured root the text is left alone, got %q", got)
	}
}

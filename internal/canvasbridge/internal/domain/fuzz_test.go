package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
)

// FuzzBridgeInput holds every parser of untrusted browser input to its
// promises on arbitrary bytes: none panics, an accepted workspace path stays
// inside the workspace folders, an admitted request names the bridge, an
// action's arguments become argv entries that are never flags or escapes,
// and a cookie value never carries a separator.
func FuzzBridgeInput(f *testing.F) {
	f.Add(inDomains("plant.domain.json"), "127.0.0.1:4100", "http://127.0.0.1:4100", "streamsim_canvas=s-9; a=b", []byte(`{"path":"`+inDomains("p.domain.json")+`"}`))
	f.Add("runs/w-1/sealed/journal.jsonl", "evil.example", "null", "a==b;;", []byte(`{"run":"w-1","out":"scenarios/s.scenario.json"}`))
	f.Add("../../etc/passwd", "localhost:4100", "", "", []byte(`{"hours":-1e400,"ref":"--out"}`))
	f.Add("runs/w\x00/run.json", "", "x", ";", []byte(`[]`))
	f.Fuzz(func(t *testing.T, path, host, origin string, cookies string, args []byte) {
		checkPath(t, path)
		checkAdmission(t, host, origin)
		checkCookie(t, cookies)
		checkArguments(t, args)
		checkCode(t, path)
	})
}

func checkPath(t *testing.T, rel string) {
	t.Helper()
	if journal, err := RunJournal(rel); err == nil && (strings.Contains(rel, "/") || journal.RunFile != "journal.jsonl") {
		t.Fatalf("RunJournal(%q) = %+v", rel, journal)
	}
	_ = SummarizeRun([]byte(rel), []byte(rel))
	_ = Scrub(rel, "/work")
	parsed, err := ParsePath(rel)
	if err != nil {
		if !errors.Is(err, ErrBadPath) {
			t.Fatalf("ParsePath(%q) failed with %v", rel, err)
		}
		return
	}
	if parsed.Rel != rel || strings.HasPrefix(rel, "/") || strings.ContainsAny(rel, "\\\x00") {
		t.Fatalf("ParsePath(%q) = %+v", rel, parsed)
	}
	for _, part := range strings.Split(rel, "/") {
		if part == ".." || part == "." || part == "" {
			t.Fatalf("ParsePath(%q) accepted the segment %q", rel, part)
		}
	}
	folder, _, _ := strings.Cut(rel, "/")
	if _, _, known := ClassOfFolder(folder); !known && folder != "runs" {
		t.Fatalf("ParsePath(%q) accepted a path outside the workspace folders", rel)
	}
}

func checkAdmission(t *testing.T, host, origin string) {
	t.Helper()
	const bound = "127.0.0.1:4100"
	err := Admit(Request{Host: host, Origin: origin}, bound)
	if err != nil && !errors.Is(err, ErrNotAdmitted) {
		t.Fatalf("Admit(%q, %q) failed with %v", host, origin, err)
	}
	if err == nil && host != bound && host != "localhost:4100" {
		t.Fatalf("Admit admitted the host %q", host)
	}
	if err == nil && origin != "" && origin != "http://"+host {
		t.Fatalf("Admit admitted the origin %q for %q", origin, host)
	}
	_ = AdmitListenAddress(host)
}

func checkCookie(t *testing.T, header string) {
	t.Helper()
	value := CookieValue(header, "streamsim_canvas")
	if strings.Contains(value, ";") || (value != "" && !strings.Contains(header, "streamsim_canvas="+value)) {
		t.Fatalf("CookieValue(%q) = %q", header, value)
	}
}

func checkCode(t *testing.T, offered string) {
	t.Helper()
	normal := NormalizeCode(offered)
	if NormalizeCode(normal) != normal || strings.ReplaceAll(FormatCode(normal), "-", "") != normal {
		t.Fatalf("NormalizeCode(%q) = %q is not stable", offered, normal)
	}
	_ = NewPairing("ABCDEFGHJKLM").Pair(offered, "s")
}

func checkArguments(t *testing.T, raw []byte) {
	t.Helper()
	var args map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&args) != nil {
		_, _ = Lookup(string(raw))
		_ = ToolFailure(string(raw))
		_ = CommandFailure(raw, 1, string(raw))
		_ = SummarizeRun(raw, raw)
		return
	}
	for _, action := range Actions() {
		checkAction(t, action, args)
		_, _, _ = action.Guarded(args)
	}
}

func checkAction(t *testing.T, action Action, args map[string]any) {
	t.Helper()
	if action.Kind == KindMCP {
		checkToolArguments(t, action, args)
		return
	}
	parts, err := action.Argv(args)
	if err != nil {
		if !errors.Is(err, ErrBadArguments) {
			t.Fatalf("%s: %v", action.Name, err)
		}
		return
	}
	for _, part := range parts[len(action.Command):] {
		checkPart(t, action, part)
	}
}

func checkPart(t *testing.T, action Action, part ArgvPart) {
	t.Helper()
	if part.Path != "" {
		if _, err := ParsePath(part.Path); err != nil && !runFolder(part.Path) {
			t.Fatalf("%s produced the path %q", action.Name, part.Path)
		}
		return
	}
	if _, err := strconv.ParseFloat(part.Literal, 64); err != nil && strings.HasPrefix(part.Literal, "-") && !isFlag(action, part.Literal) {
		t.Fatalf("%s turned an argument into the flag %q", action.Name, part.Literal)
	}
}

func runFolder(rel string) bool {
	parts := strings.Split(rel, "/")
	return len(parts) == 2 && parts[0] == "runs" && !badSegment(parts[1])
}

func isFlag(action Action, literal string) bool {
	for _, spec := range action.CLIArgs {
		if spec.Flag == literal {
			return true
		}
	}
	return false
}

func checkToolArguments(t *testing.T, action Action, args map[string]any) {
	t.Helper()
	out, err := action.ToolArguments(args)
	if err != nil {
		if !errors.Is(err, ErrBadArguments) {
			t.Fatalf("%s: %v", action.Name, err)
		}
		return
	}
	for key := range out {
		if !isKey(action, key) {
			t.Fatalf("%s passed the undeclared argument %q to the director", action.Name, key)
		}
	}
}

func isKey(action Action, key string) bool {
	for _, declared := range action.Keys {
		if declared == key {
			return true
		}
	}
	return false
}

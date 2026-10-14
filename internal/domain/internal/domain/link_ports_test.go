package domain

import (
	"strings"
	"testing"
)

func withPorts(doc map[string]any) {
	put(doc, "out_flow", "links", 0, "from_port")
	put(doc, "in_flow", "links", 0, "to_port")
}

func TestALinkKeepsItsPortsThroughLoadAndFormat(t *testing.T) {
	t.Parallel()
	loaded, err := Parse(structuredDocument(t, withPorts), "ports.json")
	if err != nil {
		t.Fatal(err)
	}
	if first := loaded.Spec.Links[0]; first.FromPort != "out_flow" || first.ToPort != "in_flow" {
		t.Fatalf("link 0 = %+v, want both ports", first)
	}
	formatted, err := Format(loaded)
	if err != nil {
		t.Fatal(err)
	}
	again, err := Parse(formatted, "ports.json")
	if err != nil || again.Digest != loaded.Digest || again.Spec.Links[0] != loaded.Spec.Links[0] {
		t.Fatalf("round trip = %+v (%v), want the same digest and the same link", again, err)
	}
}

func TestALinkWithoutPortsStaysValidAndWritesNone(t *testing.T) {
	t.Parallel()
	loaded := loadStructured(t)
	for _, link := range loaded.Spec.Links {
		if link.FromPort != "" || link.ToPort != "" {
			t.Fatalf("link %+v carries a port the file never named", link)
		}
	}
	formatted, err := Format(loaded)
	if err != nil || strings.Contains(string(formatted), "_port") {
		t.Fatalf("Format wrote a port nobody named: %v", err)
	}
}

func TestTheLinkObjectStaysClosedAndPortsDoNotMakeALinkNew(t *testing.T) {
	t.Parallel()
	structuredViolation(t, func(doc map[string]any) { put(doc, "x", "links", 0, "via") }, "/links/0/via")
	structuredViolation(t, func(doc map[string]any) { put(doc, "Bad Port", "links", 0, "to_port") }, "/links/0/to_port")
	structuredViolation(t, func(doc map[string]any) {
		withPorts(doc)
		twice := map[string]any{"type": "feeds", "from": "tank-1", "to": "pump-1", "from_port": "other"}
		put(doc, append(doc["links"].([]any), twice), "links")
	}, "/links/3")
}

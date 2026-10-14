package domain

import "testing"

func TestEveryComposedLinkNamesThePortsItsConnectionUsed(t *testing.T) {
	t.Parallel()
	raw := connected(t, []any{link("S-1.out_load", "K-1.load_in"), link("S-2.out_load", "K-1.load_in")},
		entryOf("acme/source@1", "source", "S-1", "S-2"), entryOf("acme/kiln@1", "kiln", "K-1"))
	links := member(t, composeOK(t, plantLibrary(t), raw), "links").([]any)
	for i, one := range links {
		if member(t, one, "from_port") != "out_load" || member(t, one, "to_port") != "load_in" {
			t.Errorf("link %d = %v, want from_port out_load and to_port load_in", i, one)
		}
	}
}

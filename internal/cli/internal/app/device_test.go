package app

import "testing"

func TestFaultSpecFlagIsRepeatable(t *testing.T) {
	t.Parallel()
	var flag faultSpecFlag
	if err := flag.Set("stuck@2"); err != nil {
		t.Fatal(err)
	}
	if err := flag.Set("ack_lost"); err != nil {
		t.Fatal(err)
	}
	if len(flag.entries) != 2 || flag.entries[0].AcceptedCommand != 2 || flag.entries[1].AcceptedCommand != 1 {
		t.Fatalf("fault entries = %+v", flag.entries)
	}
}

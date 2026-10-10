package domain

import "testing"

func TestAdmissionPreservesBootBeforeFreshnessBeforeTarget(t *testing.T) {
	d := New(Config{Capabilities: testCaps(t), Clock: func() int64 { return 1000 }})
	command := validCommand(t, nil)
	command["expected_boot_id"] = "wrong"
	command["not_before_mono_us"] = float64(2000)
	command["target"] = "wrong"
	if got := d.admit(command, 1000); got != "wrong_boot" {
		t.Fatalf("first rejection=%q", got)
	}
	command["expected_boot_id"] = d.BootID()
	if got := d.admit(command, 1000); got != "not_ready" {
		t.Fatalf("freshness rejection=%q", got)
	}
	command["not_before_mono_us"] = float64(0)
	if got := d.admit(command, 1000); got != "wrong_target" {
		t.Fatalf("target rejection=%q", got)
	}
}

package domain

import "testing"

func TestTheWallScheduleReleasesEachWorldInstantAtItsOwnWallInstant(t *testing.T) {
	t.Parallel()
	started := WallScheduleFrom(1_767_225_600_123_456_789)
	if started.WorldStartNS != 1_767_225_600_123_000_000 || started.WallStartNS != started.WorldStartNS {
		t.Fatalf("a paced world starts at the wall time floored to the millisecond, got %+v", started)
	}
	shifted := WallSchedule{WorldStartNS: 1_000, WallStartNS: 5_000}
	cases := []struct {
		name                     string
		schedule                 WallSchedule
		world, now, wallAt, wait int64
	}{
		{"before its time", started, started.WorldStartNS + 1e9, started.WallStartNS + 250e6, started.WallStartNS + 1e9, 750e6},
		{"on its time", started, started.WorldStartNS + 1e9, started.WallStartNS + 1e9, started.WallStartNS + 1e9, 0},
		{"late: no wait, no catch-up debt", started, started.WorldStartNS + 1e9, started.WallStartNS + 4e9, started.WallStartNS + 1e9, 0},
		{"a world that starts elsewhere", shifted, 3_000, 6_000, 7_000, 1_000},
		{"far ahead of a 2026 start without overflow", started, started.WorldStartNS + 30*24*3600e9, started.WallStartNS, started.WallStartNS + 30*24*3600e9, 30 * 24 * 3600e9},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.schedule.WallAt(c.world); got != c.wallAt {
				t.Errorf("WallAt = %d, want %d", got, c.wallAt)
			}
			if got := c.schedule.WaitNS(c.world, c.now); got != c.wait {
				t.Errorf("WaitNS = %d, want %d", got, c.wait)
			}
		})
	}
}

func TestAMergedBirthIsOnTheWallClockWhenItsSimulatedNodeIsPaced(t *testing.T) {
	t.Parallel()
	gateway := map[string]any{"node_metrics": map[string]any{"clock_mode": "virtual", "run_state": "running"}}
	cases := []struct {
		name               string
		gateway, simulated map[string]any
		want               map[string]any
	}{
		{"paced", gateway, map[string]any{"node_metrics": map[string]any{"clock_mode": "wall"}}, map[string]any{"clock_mode": "wall", "run_state": "running"}},
		{"paced, gateway without metrics", map[string]any{}, map[string]any{"node_metrics": map[string]any{"clock_mode": "wall"}}, map[string]any{"clock_mode": "wall"}},
		{"unpaced keeps the gateway's", gateway, map[string]any{"node_metrics": map[string]any{"clock_mode": "virtual"}}, gateway["node_metrics"].(map[string]any)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, _ := mergedBirth(c.gateway, c.simulated)["node_metrics"].(map[string]any)
			if len(got) != len(c.want) || got["clock_mode"] != c.want["clock_mode"] || got["run_state"] != c.want["run_state"] {
				t.Fatalf("node_metrics %v, want %v", got, c.want)
			}
		})
	}
	if gateway["node_metrics"].(map[string]any)["clock_mode"] != "virtual" {
		t.Error("merging changed the gateway's own birth")
	}
}

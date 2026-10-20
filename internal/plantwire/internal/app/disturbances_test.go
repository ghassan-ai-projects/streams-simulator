package app

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type reading struct {
	atNS  int64
	value float64
}

func numberOf(value any) float64 {
	number, _ := strconv.ParseFloat(fmt.Sprint(value), 64)
	return number
}

func readingsOf(records []map[string]any, device, metric string) []reading {
	alias := ""
	var found []reading
	for _, record := range records {
		if record["device"] != device {
			continue
		}
		if record["type"] == "DBIRTH" {
			alias = aliasOf(record, metric)
		}
		if record["type"] == "DDATA" {
			found = append(found, samplesOf(record, alias)...)
		}
	}
	return found
}

func aliasOf(birth map[string]any, metric string) string {
	metrics, _ := birth["metrics"].([]any)
	for _, entry := range metrics {
		if fields, _ := entry.(map[string]any); fields["name"] == metric {
			return fmt.Sprint(fields["alias"])
		}
	}
	return ""
}

func samplesOf(data map[string]any, alias string) []reading {
	var found []reading
	samples, _ := data["metrics"].([]any)
	for _, sample := range samples {
		if fields, _ := sample.(map[string]any); fmt.Sprint(fields["alias"]) == alias {
			found = append(found, reading{atNS: int64(numberOf(fields["event_time_ns"])), value: numberOf(fields["value"])})
		}
	}
	return found
}

func meanBetween(readings []reading, fromNS, toNS int64) float64 {
	sum, n := 0.0, 0
	for _, r := range readings {
		if r.atNS >= fromNS && r.atNS < toNS {
			sum, n = sum+r.value, n+1
		}
	}
	return sum / float64(max(n, 1))
}

func thermalSession(t *testing.T, disturbances []domain.ScheduledFault) []map[string]any {
	t.Helper()
	host := &scriptedHost{t: t}
	options := sliceOptions(60)
	options.Disturbances = disturbances
	if err := NewSession(runOf(t, testsupport.ThermalBoardDomain()), host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	return host.sent
}

func zoneShift(disturbed, quiet []map[string]any, fromNS, toNS int64) float64 {
	return meanBetween(readingsOf(disturbed, "zone-01", "temperature"), fromNS, toNS) - meanBetween(readingsOf(quiet, "zone-01", "temperature"), fromNS, toNS)
}

func TestAScheduledHeatLoadRaisesTheZoneFromItsOnsetAndTheHostSeesOnlyTheReadings(t *testing.T) {
	t.Parallel()
	onset, hour := model.DefaultStartTimeNS+600e9, model.DefaultStartTimeNS+3600e9
	quiet := thermalSession(t, nil)
	lasting := thermalSession(t, []domain.ScheduledFault{{Entity: "zone-01", Fault: "zone_heat_load", OnsetNS: onset}})
	cleared := thermalSession(t, []domain.ScheduledFault{{Entity: "zone-01", Fault: "zone_heat_load", OnsetNS: onset, ClearNS: onset + 600e9}})
	if before, after := zoneShift(lasting, quiet, 0, onset), zoneShift(lasting, quiet, hour-1200e9, hour); before != 0 || after < 6 {
		t.Fatalf("a lasting heat load moves the zone %.2f before its onset and %.2f in the last 20 minutes, want 0 then at least 6 degC", before, after)
	}
	if during, after := zoneShift(cleared, quiet, onset+300e9, onset+600e9), zoneShift(cleared, quiet, hour-1200e9, hour); during < 3 || after > 1 {
		t.Fatalf("a heat load cleared after 10 minutes moves the zone %.2f while it lasts and %.2f at the end, want warmer, then back", during, after)
	}
	for _, record := range slices.Concat(lasting, cleared) {
		if text := fmt.Sprint(record); strings.Contains(text, "heat_load") || strings.Contains(text, "fault") {
			t.Fatalf("a %v record names the disturbance: %s", record["type"], text)
		}
	}
}

func TestADisturbanceTheWorldRefusesEndsTheSessionBeforeItsBirth(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t}
	options := sliceOptions(2)
	options.Disturbances = []domain.ScheduledFault{{Entity: "zone-01", Fault: "zone_heat_load", OnsetNS: model.DefaultStartTimeNS, ClearNS: model.DefaultStartTimeNS + 1}}
	err := NewSession(runOf(t, testsupport.UtilityDomain()), host, options).Run(t.Context())
	if err == nil || !strings.Contains(err.Error(), "disturbance zone_heat_load on zone-01") || len(host.sent) != 0 {
		t.Fatalf("a disturbance on an entity the world lacks: %v after %d records, want the session refused before any record", err, len(host.sent))
	}
}

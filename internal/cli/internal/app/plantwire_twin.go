package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

func cmdPlantWireTwinReport(s *session, args []string) (any, error) {
	o, err := parsePlantWireTwinOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	report, err := plantwire.TwinReport(context.Background(), plantwire.TwinReportOptions{Run: o.run, DomainsDir: o.domainsDir,
		Domain: o.domain, KindsDirs: o.kinds.items(), Tolerances: o.tolerances.items(), SpanFraction: o.spanFraction})
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return report, nil
}

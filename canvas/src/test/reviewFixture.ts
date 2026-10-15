import type { ReviewState } from "../app/reviewStore";
import { parseReport } from "../model/report";
import { report } from "./runFixture";

const parsed = parseReport(report);

export function reviewState(overrides: Partial<ReviewState>): ReviewState {
  return {
    worldId: "w-1",
    report: { phase: "ready", value: parsed },
    verified: { phase: "ready", value: { matches: true, domainChanged: false, detail: "", reason: "", file: "" } },
    playback: { phase: "ready", records: parsed.counts.delivered ?? 0, adapter: "native-jsonl" },
    positionNs: 0,
    selected: null,
    runDomain: { phase: "loading" },
    tab: "report",
    picked: null,
    playing: false,
    speed: 60,
    openedAt: 0,
    ...overrides,
  };
}

import type { ReviewState } from "../app/review";
import { report } from "./runFixture";

export function reviewState(overrides: Partial<ReviewState>): ReviewState {
  return {
    worldId: "w-1",
    report: { phase: "ready", value: report },
    verified: { phase: "ready", value: { matches: true, domainChanged: false, detail: "", reason: "", file: "" } },
    playback: { phase: "ready", records: report.counts.delivered ?? 0 },
    positionNs: 0,
    selected: null,
    runDomain: { phase: "loading" },
    tab: "report",
    picked: null,
    playing: false,
    speed: 60,
    openedAt: 0,
    origins: new Map(),
    recordedOrigins: new Map(),
    connection: "none",
    simulated: [],
    sessionDomain: "",
    sessionPolicies: null,
    binding: null,
    following: false,
    newestNs: 0,
    live: null,
    liveHistory: null,
    liveMarks: [],
    liveUnreadableNs: null,
    interruptUnreadableNs: null,
    lastLiveAtMs: null,
    caughtUp: true,
    ...overrides,
  };
}

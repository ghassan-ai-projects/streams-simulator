import type { InstanceOrigin, PlaybackIndex, SessionLive } from "../../../model/review";
import type { ReviewState } from "./reviewState";

export interface LiveHistory {
  readonly previous: LiveHistory | null;
  readonly entries: readonly SessionLive[];
  readonly startNs: number;
  readonly endNs: number;
}

function snapshotAt(history: LiveHistory | null, positionNs: number): SessionLive | null {
  for (let batch = history; batch !== null; batch = batch.previous) {
    const found = batch.entries.findLast((snapshot) => snapshot.atNs <= positionNs);
    if (found !== undefined) return found;
  }
  return null;
}

function cursorOrigins(state: ReviewState, live: SessionLive | null): ReviewState["origins"] {
  if (state.liveHistory === null) return state.origins;
  const origins = new Map<string, InstanceOrigin>();
  for (const device of live?.devices ?? []) {
    if (device.instance === "") continue;
    const kind = device.environment === "hardware" ? "real" : device.environment;
    origins.set(device.instance, {
      instance: device.instance,
      kind,
      board: device.board,
      device: device.device,
      bootId: device.bootId,
    });
  }
  return origins;
}

export function atPosition(state: ReviewState, positionNs: number): ReviewState {
  const live = snapshotAt(state.liveHistory, positionNs);
  return { ...state, positionNs, live, origins: cursorOrigins(state, live) };
}

export function playbackBounds(state: ReviewState, index: PlaybackIndex) {
  const observations = index.bounds();
  const history = state.liveHistory;
  if (history === null) return observations;
  return {
    startNs: Math.min(history.startNs, observations?.startNs ?? history.startNs),
    endNs: Math.max(history.endNs, observations?.endNs ?? history.endNs),
  };
}

export function withLiveBatch(state: ReviewState, entries: readonly SessionLive[]): ReviewState {
  const first = entries[0];
  const last = entries.at(-1);
  if (first === undefined || last === undefined) return state;
  const liveHistory = {
    previous: state.liveHistory,
    entries,
    startNs: state.liveHistory?.startNs ?? first.atNs,
    endNs: last.atNs,
  };
  const marks = entries.flatMap((entry) =>
    entry.marks.map((mark) => ({ ...mark, raw: mark.label, tone: "other" as const, untilNs: null })),
  );
  const liveMarks = marks.length === 0 ? state.liveMarks : [...state.liveMarks, ...marks];
  const updated = { ...state, liveHistory, liveMarks, newestNs: last.atNs };
  return atPosition(updated, state.following && state.connection === "running" ? last.atNs : state.positionNs);
}

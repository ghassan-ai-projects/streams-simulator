import type { EventMark, InstanceOrigin, JournalEntry, PlaybackIndex, SessionLive } from "../../../model/review";
import { originOf } from "../../../model/review";
import type { LiveHistory, ReviewState } from "./reviewState";

export type LiveItem =
  { readonly kind: "snapshot"; readonly live: SessionLive } | { readonly kind: "unreadable"; readonly atNs: number };

const maxLiveSnapshots = 2000;
const maxLiveMarks = 2000;

type Origins = ReadonlyMap<string, InstanceOrigin>;

const originCache = new WeakMap<SessionLive, { readonly recorded: Origins; readonly origins: Origins }>();

function snapshotAt(history: LiveHistory | null, positionNs: number): SessionLive | null {
  const entries = history?.entries ?? [];
  let low = 0;
  let high = entries.length;
  while (low < high) {
    const middle = Math.floor((low + high) / 2);
    if ((entries[middle]?.atNs ?? Infinity) <= positionNs) low = middle + 1;
    else high = middle;
  }
  return entries[low - 1] ?? null;
}

function cursorOrigins(state: ReviewState, live: SessionLive | null): Origins {
  if (state.liveHistory === null) return state.origins;
  if (live === null) return state.recordedOrigins;
  const cached = originCache.get(live);
  if (cached?.recorded === state.recordedOrigins) return cached.origins;
  const origins = new Map(state.recordedOrigins);
  for (const device of live.devices) {
    const origin = originOf(device);
    if (origin !== null) origins.set(origin.instance, origin);
  }
  originCache.set(live, { recorded: state.recordedOrigins, origins });
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

function playStartNs(state: ReviewState, index: PlaybackIndex): number | null {
  const settled = index.settled();
  const history = state.liveHistory;
  if (history === null) return settled;
  return Math.min(history.startNs, settled ?? history.startNs);
}

function appendDistinct(entries: readonly SessionLive[], fresh: readonly SessionLive[]): SessionLive[] {
  const kept = [...entries];
  for (const snapshot of fresh) {
    if (kept.at(-1)?.key !== snapshot.key) kept.push(snapshot);
  }
  return kept;
}

function liveStartNs(
  history: LiveHistory | null,
  kept: readonly SessionLive[],
  trimmed: boolean,
  newest: SessionLive,
): number {
  if (trimmed) return kept.at(-maxLiveSnapshots)?.atNs ?? newest.atNs;
  return history?.startNs ?? kept[0]?.atNs ?? newest.atNs;
}

function appendSnapshots(history: LiveHistory | null, fresh: readonly SessionLive[]): LiveHistory | null {
  const newest = fresh.at(-1);
  if (newest === undefined) return history;
  const kept = appendDistinct(history?.entries ?? [], fresh);
  const trimmed = kept.length > maxLiveSnapshots;
  const entries = trimmed ? kept.slice(-maxLiveSnapshots) : kept;
  const startNs = liveStartNs(history, kept, trimmed, newest);
  return {
    entries,
    startNs,
    endNs: Math.max(history?.endNs ?? newest.atNs, newest.atNs),
    trimmedBeforeNs: trimmed ? startNs : (history?.trimmedBeforeNs ?? null),
  };
}

function withMarks(current: readonly EventMark[], fresh: readonly SessionLive[]): readonly EventMark[] {
  const marks = fresh.flatMap((entry) =>
    entry.marks.map((mark) => ({
      atNs: mark.atNs,
      label: mark.label,
      raw: mark.label,
      kind: mark.kind,
      tone: "other" as const,
      entity: mark.entity,
      untilNs: null,
    })),
  );
  if (marks.length === 0) return current;
  return [...current, ...marks].slice(-maxLiveMarks);
}

function unreadableAfter(current: number | null, items: readonly LiveItem[]): number | null {
  let unreadable = current;
  for (const item of items) unreadable = item.kind === "unreadable" ? item.atNs : null;
  return unreadable;
}

function followsNewest(state: ReviewState): boolean {
  return state.following && state.connection === "running";
}

export function withLiveBatch(state: ReviewState, items: readonly LiveItem[], nowMs: number): ReviewState {
  if (items.length === 0) return state;
  const fresh = items.flatMap((item) => (item.kind === "snapshot" ? [item.live] : []));
  const liveHistory = appendSnapshots(state.liveHistory, fresh);
  const updated: ReviewState = {
    ...state,
    liveHistory,
    liveMarks: withMarks(state.liveMarks, fresh),
    liveUnreadableNs: unreadableAfter(state.liveUnreadableNs, items),
    newestNs: liveHistory?.endNs ?? state.newestNs,
    lastLiveAtMs: fresh.length > 0 ? nowMs : state.lastLiveAtMs,
  };
  return atPosition(updated, followsNewest(state) ? updated.newestNs : state.positionNs);
}

export function liveItemOf(entry: JournalEntry): LiveItem | null {
  if (entry.kind === "live") return { kind: "snapshot", live: entry.live };
  return entry.kind === "live-unreadable" ? { kind: "unreadable", atNs: entry.atNs } : null;
}

export function caughtUpAt(state: ReviewState): ReviewState {
  return { ...atPosition(state, followsNewest(state) ? state.newestNs : state.positionNs), caughtUp: true };
}

export function loadStart(state: ReviewState, index: PlaybackIndex): number | null {
  const start = playStartNs(state, index);
  const before = (playbackBounds(state, index)?.startNs ?? 0) > state.positionNs;
  return start !== null && before && !followsNewest(state) ? start : null;
}

export function playStart(state: ReviewState, index: PlaybackIndex): ReviewState | null {
  const bounds = playbackBounds(state, index);
  const start = playStartNs(state, index);
  if (bounds === null || start === null) return null;
  const fresh = state.positionNs >= bounds.endNs || state.positionNs < bounds.startNs;
  return { ...atPosition(state, fresh ? start : state.positionNs), playing: true, following: false };
}

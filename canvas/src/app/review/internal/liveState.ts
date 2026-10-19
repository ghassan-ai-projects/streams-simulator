import type {
  EventMark,
  InstanceOrigin,
  InterruptRecord,
  JournalEntry,
  PlaybackIndex,
  SessionLive,
} from "../../../model/review";
import { interruptMarks, originOf } from "../../../model/review";
import type { RunEntry } from "../../../model/workspace";
import type { LiveHistory, ReviewState } from "./reviewState";

export type LiveItem =
  | { readonly kind: "snapshot"; readonly live: SessionLive }
  | { readonly kind: "unreadable"; readonly atNs: number }
  | { readonly kind: "interrupt-unreadable"; readonly atNs: number }
  | { readonly kind: "interrupt"; readonly atNs: number; readonly record: InterruptRecord };

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

function snapshotMarks(live: SessionLive): readonly EventMark[] {
  return live.marks.map((mark) => ({
    atNs: mark.atNs,
    label: mark.label,
    raw: mark.label,
    kind: mark.kind,
    tone: "other" as const,
    entity: mark.entity,
    untilNs: null,
  }));
}

function itemMarks(item: LiveItem): readonly EventMark[] {
  if (item.kind === "snapshot") return snapshotMarks(item.live);
  return item.kind === "interrupt" ? interruptMarks(item.record, item.atNs) : [];
}

function withMarks(current: readonly EventMark[], items: readonly LiveItem[]): readonly EventMark[] {
  const marks = items.flatMap(itemMarks);
  if (marks.length === 0) return current;
  return [...current, ...marks].slice(-maxLiveMarks);
}

function unreadableAfter(current: number | null, items: readonly LiveItem[]): number | null {
  let unreadable = current;
  for (const item of items) {
    if (item.kind === "unreadable") unreadable = item.atNs;
    else if (item.kind === "snapshot") unreadable = null;
  }
  return unreadable;
}

function interruptUnreadableAfter(current: number | null, items: readonly LiveItem[]): number | null {
  return items.reduce((newest, item) => (item.kind === "interrupt-unreadable" ? item.atNs : newest), current);
}

function followsNewest(state: ReviewState): boolean {
  return state.following && state.connection === "running";
}

function withLiveBatch(state: ReviewState, items: readonly LiveItem[], nowMs: number): ReviewState {
  if (items.length === 0) return state;
  const fresh = items.flatMap((item) => (item.kind === "snapshot" ? [item.live] : []));
  const liveHistory = appendSnapshots(state.liveHistory, fresh);
  const updated: ReviewState = {
    ...state,
    liveHistory,
    liveMarks: withMarks(state.liveMarks, items),
    liveUnreadableNs: unreadableAfter(state.liveUnreadableNs, items),
    interruptUnreadableNs: interruptUnreadableAfter(state.interruptUnreadableNs, items),
    newestNs: liveHistory?.endNs ?? state.newestNs,
    lastLiveAtMs: fresh.length > 0 ? nowMs : state.lastLiveAtMs,
  };
  return atPosition(updated, followsNewest(state) ? updated.newestNs : state.positionNs);
}

export function liveItemOf(entry: JournalEntry): LiveItem | null {
  if (entry.kind === "live") return { kind: "snapshot", live: entry.live };
  if (entry.kind === "interrupt") return { kind: "interrupt", atNs: entry.atNs, record: entry.record };
  if (entry.kind === "interrupt-unreadable") return { kind: "interrupt-unreadable", atNs: entry.atNs };
  return entry.kind === "live-unreadable" ? { kind: "unreadable", atNs: entry.atNs } : null;
}

function newestOf(state: ReviewState, index: PlaybackIndex): number {
  return Math.max(state.newestNs, index.bounds()?.endNs ?? state.newestNs);
}

export function drained(
  state: ReviewState,
  items: readonly LiveItem[],
  nowMs: number,
  index: PlaybackIndex,
): ReviewState {
  const next = withLiveBatch(state, items, nowMs);
  const newest = newestOf(next, index);
  return followsNewest(next) && next.positionNs !== newest ? atPosition(next, newest) : next;
}

export function toLive(state: ReviewState, index: PlaybackIndex): ReviewState {
  return { ...atPosition(state, newestOf(state, index)), playing: false, following: true };
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

export function liveEntry(worldId: string, known: RunEntry | undefined): RunEntry {
  const base = { incompleteReason: "", explore: false, seed: null, plant: "", durationS: 0, recordedAt: "", faults: 0 };
  return { ...base, ...known, worldId, ended: false, complete: false, connected: "running" };
}

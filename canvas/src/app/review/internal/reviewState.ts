import type { RunEntry } from "../../../model/workspace";
import type {
  Connection,
  EventMark,
  InstanceOrigin,
  JournalEntry,
  Report,
  Replay,
  SessionLive,
  TimedBinding,
} from "../../../model/review";
import { withOrigin } from "../../../model/review";
import { unfinishedReason } from "../../../model/workspace";
import type { DomainDocument, Outcome, ReviewFailure } from "../../ports";

export type Load<T> =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: ReviewFailure }
  | { readonly phase: "ready"; readonly value: T };

export function toLoad<T>(outcome: Outcome<T, ReviewFailure>): Load<T> {
  return outcome.ok ? { phase: "ready", value: outcome.value } : { phase: "failed", failure: outcome.failure };
}

export interface LiveHistory {
  readonly entries: readonly SessionLive[];
  readonly startNs: number;
  readonly endNs: number;
  readonly trimmedBeforeNs: number | null;
}

export type OpenMode = "review" | "watch";

export interface ReviewState {
  readonly following: boolean;
  readonly watching: boolean;
  readonly newestNs: number;
  readonly live: SessionLive | null;
  readonly liveHistory: LiveHistory | null;
  readonly liveMarks: readonly EventMark[];
  readonly liveUnreadableNs: number | null;
  readonly interruptUnreadableNs: number | null;
  readonly lastLiveAtMs: number | null;
  readonly caughtUp: boolean;
  readonly worldId: string | null;
  readonly report: Load<Report>;
  readonly verified: Load<Replay>;
  readonly playback: {
    readonly phase: "loading" | "ready" | "failed";
    readonly records: number;
  };
  readonly positionNs: number;
  readonly selected: string | null;
  readonly runDomain: Load<DomainDocument>;
  readonly tab: "report" | "inspect";
  readonly picked: { readonly atNs: number; readonly entity: string } | null;
  readonly openedAt: number;
  readonly playing: boolean;
  readonly speed: number;
  readonly origins: ReadonlyMap<string, InstanceOrigin>;
  readonly recordedOrigins: ReadonlyMap<string, InstanceOrigin>;
  readonly connection: Connection;
  readonly running: boolean;
  readonly simulated: readonly string[];
  readonly sessionDomain: string;
  readonly sessionPolicies: readonly string[] | null;
  readonly binding: TimedBinding | null;
}

export const idle: ReviewState = {
  following: true,
  watching: false,
  newestNs: 0,
  live: null,
  liveHistory: null,
  liveMarks: [],
  liveUnreadableNs: null,
  interruptUnreadableNs: null,
  lastLiveAtMs: null,
  caughtUp: false,
  worldId: null,
  report: { phase: "loading" },
  verified: { phase: "loading" },
  playback: { phase: "loading", records: 0 },
  positionNs: 0,
  selected: null,
  runDomain: { phase: "loading" },
  tab: "report",
  picked: null,
  openedAt: 0,
  playing: false,
  speed: 60,
  origins: new Map(),
  recordedOrigins: new Map(),
  connection: "none",
  running: false,
  simulated: [],
  sessionDomain: "",
  sessionPolicies: null,
  binding: null,
};

export function runGoing(state: ReviewState): boolean {
  return state.watching || state.running || state.connection === "running";
}

export function endedSinceReport(state: ReviewState, worldId: string, known: RunEntry | undefined): boolean {
  const { report } = state;
  return known?.ended === true && state.worldId === worldId && report.phase === "ready" && !report.value.ended;
}

function connectionOf(known: RunEntry | undefined): Connection {
  return known?.connected ?? "none";
}

export function withListedConnection(state: ReviewState, known: RunEntry | undefined): ReviewState {
  const running = known?.running ?? state.running;
  const listed = known?.connected;
  if (listed === undefined || listed === state.connection)
    return running === state.running ? state : { ...state, running };
  return { ...state, running, connection: listed, simulated: known?.simulated ?? state.simulated };
}

export function watchEnded(state: ReviewState): ReviewState {
  return state.watching ? { ...state, watching: false, running: false, following: false } : state;
}

function closedState(state: ReviewState): ReviewState {
  if (state.watching) return watchEnded(state);
  return state.connection === "running" ? { ...state, connection: "saved", following: false } : state;
}

export type SessionEntry = Extract<JournalEntry, { kind: "origin" | "binding" | "closed" | "session" }>;

export function withSessionEntry(state: ReviewState, entry: SessionEntry): ReviewState {
  if (entry.kind === "closed") return closedState(state);
  const connection = state.connection === "none" ? "saved" : state.connection;
  if (entry.kind === "session")
    return {
      ...state,
      connection,
      simulated: entry.simulated,
      sessionDomain: entry.domain,
      sessionPolicies: entry.policies,
    };
  if (entry.kind === "origin")
    return {
      ...state,
      connection,
      origins: withOrigin(state.origins, entry.origin),
      recordedOrigins: withOrigin(state.recordedOrigins, entry.origin),
    };
  return { ...state, connection, binding: { atNs: entry.atNs, binding: entry.binding } };
}

const incomplete: ReviewFailure = {
  kind: "run-incomplete",
  diagnostic: {
    displayCode: "run_incomplete",
    message: "run.json is absent: the run is still open or did not end",
  },
};

type FinalLoad = "domain" | "verify";

export interface FinalLoads {
  readonly refreshed: boolean;
  readonly deferred: Set<FinalLoad>;
}

export function finalLoadsFor(known: RunEntry | undefined): FinalLoads {
  return { refreshed: known?.ended === true, deferred: new Set(knownUnfinished(known) ? ["domain"] : []) };
}

export function knownUnfinished(known: RunEntry | undefined): known is RunEntry {
  return known !== undefined && !known.complete;
}

function absentLoads(known: RunEntry | undefined): Pick<ReviewState, "verified" | "runDomain"> | Record<string, never> {
  if (!knownUnfinished(known)) return {};
  return {
    verified: {
      phase: "failed",
      failure: {
        ...incomplete,
        diagnostic: { ...incomplete.diagnostic, message: `the run is incomplete: ${unfinishedReason(known)}` },
      },
    },
    runDomain: { phase: "failed", failure: incomplete },
  };
}

export function visitState(
  worldId: string,
  known: RunEntry | undefined,
  visit: { readonly openedAt: number; readonly mode: OpenMode },
): ReviewState {
  const simulated = known?.simulated ?? [];
  return {
    ...idle,
    worldId,
    openedAt: visit.openedAt,
    watching: visit.mode === "watch",
    connection: connectionOf(known),
    running: known?.running ?? false,
    simulated,
    ...absentLoads(known),
  };
}

import type { RunEntry } from "../../../model/workspace";
import type {
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
import type { LiveHistory } from "./liveState";

export type Load<T> =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: ReviewFailure }
  | { readonly phase: "ready"; readonly value: T };

export function toLoad<T>(outcome: Outcome<T, ReviewFailure>): Load<T> {
  return outcome.ok ? { phase: "ready", value: outcome.value } : { phase: "failed", failure: outcome.failure };
}

export interface ReviewState {
  readonly following: boolean;
  readonly newestNs: number;
  readonly live: SessionLive | null;
  readonly liveHistory: LiveHistory | null;
  readonly liveMarks: readonly EventMark[];
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
  readonly connection: "none" | "running" | "saved";
  readonly simulated: readonly string[];
  readonly binding: TimedBinding | null;
}

export const idle: ReviewState = {
  following: true,
  newestNs: 0,
  live: null,
  liveHistory: null,
  liveMarks: [],
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
  connection: "none",
  simulated: [],
  binding: null,
};

export function connectionOf(known: RunEntry | undefined): ReviewState["connection"] {
  return known?.connected ?? "none";
}

export function withListedConnection(state: ReviewState, known: RunEntry | undefined): ReviewState {
  const listed = known?.connected;
  if (listed === undefined || listed === state.connection) return state;
  return { ...state, connection: listed, simulated: known?.simulated ?? state.simulated };
}

type SessionEntry = Extract<JournalEntry, { kind: "origin" | "binding" | "closed" | "session" }>;

export function withSessionEntry(state: ReviewState, entry: SessionEntry): ReviewState {
  if (entry.kind === "closed")
    return state.connection === "running" ? { ...state, connection: "saved", following: false } : state;
  const connection = state.connection === "none" ? "saved" : state.connection;
  if (entry.kind === "session") return { ...state, connection, simulated: entry.simulated };
  if (entry.kind === "origin") return { ...state, connection, origins: withOrigin(state.origins, entry.origin) };
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

export function absentLoads(
  known: RunEntry | undefined,
): Pick<ReviewState, "verified" | "runDomain"> | Record<string, never> {
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

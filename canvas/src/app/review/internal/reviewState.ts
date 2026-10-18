import type { RunEntry } from "../../../model/workspace";
import type { InstanceOrigin, Report, Replay } from "../../../model/review";
import { unfinishedReason } from "../../../model/workspace";
import type { DomainDocument, Outcome, ReviewFailure } from "../../ports";

export type Load<T> =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: ReviewFailure }
  | { readonly phase: "ready"; readonly value: T };

export function toLoad<T>(outcome: Outcome<T, ReviewFailure>): Load<T> {
  return outcome.ok ? { phase: "ready", value: outcome.value } : { phase: "failed", failure: outcome.failure };
}

export interface ReviewState {
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
}

export const idle: ReviewState = {
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
};

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

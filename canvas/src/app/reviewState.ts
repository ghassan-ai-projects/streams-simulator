import type { BridgeFailure, RunEntry } from "../model/bridge";
import type { Report, Replay } from "../model/report";
import { unfinishedReason } from "../model/workspace";

export type Load<T> =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: BridgeFailure }
  | { readonly phase: "ready"; readonly value: T };

export interface ReviewState {
  readonly worldId: string | null;
  readonly report: Load<Report>;
  readonly verified: Load<Replay>;
  readonly playback: {
    readonly phase: "loading" | "ready" | "failed";
    readonly records: number;
    readonly adapter: string;
  };
  readonly positionNs: number;
  readonly selected: string | null;
  readonly runDomain: Load<unknown>;
  readonly tab: "report" | "inspect";
  readonly picked: { readonly atNs: number; readonly entity: string } | null;
  readonly openedAt: number;
  readonly playing: boolean;
  readonly speed: number;
}

export const idle: ReviewState = {
  worldId: null,
  report: { phase: "loading" },
  verified: { phase: "loading" },
  playback: { phase: "loading", records: 0, adapter: "" },
  positionNs: 0,
  selected: null,
  runDomain: { phase: "loading" },
  tab: "report",
  picked: null,
  openedAt: 0,
  playing: false,
  speed: 60,
};

const absentFile: BridgeFailure = {
  status: 404,
  code: "not_found",
  message: "run.json is absent: the run is still open or did not end",
  issues: [],
};

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
        status: 422,
        code: "run_incomplete",
        message: `the run is incomplete: ${unfinishedReason(known)}`,
        issues: [],
      },
    },
    runDomain: { phase: "failed", failure: absentFile },
  };
}

import type { CallOutcome, LaunchPlan, PlannedFault, RunControl } from "../../../model/launch";
import type { LaunchFailure } from "../../ports";
import type { LaunchOrigin } from "./trackedRun";

export type LaunchState =
  | { readonly phase: "idle" }
  | {
      readonly phase: "running";
      readonly origin: LaunchOrigin;
      readonly worldId: string | null;
      readonly atS: number;
      readonly durationS: number;
      readonly plan: LaunchPlan;
      readonly startedMs: number;
      readonly atMs: number;
      readonly stopping: boolean;
      readonly paused: boolean;
      readonly control: RunControl;
      readonly injected: readonly PlannedFault[];
      readonly rate: number | null;
      readonly calls: readonly CallOutcome[];
    }
  | {
      readonly phase: "done";
      readonly origin: LaunchOrigin;
      readonly worldId: string;
      readonly early: boolean;
      readonly plan: LaunchPlan;
      readonly reachedS: number;
      readonly calls: readonly CallOutcome[];
    }
  | {
      readonly phase: "failed";
      readonly origin: LaunchOrigin;
      readonly failure: LaunchFailure;
      readonly words: string;
      readonly kept: string | null;
    };

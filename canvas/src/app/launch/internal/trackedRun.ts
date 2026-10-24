import type { CallOutcome, LaunchPlan, PlannedFault, RunControl } from "../../../model/launch";
import { idleControl } from "../../../model/launch";

export interface LaunchOrigin {
  readonly domain: string;
  readonly title: string;
  readonly path: string;
}

export class TrackedRun {
  worldId: string | null = null;
  calls: readonly CallOutcome[] = [];
  stopping = false;
  missed = 0;
  epoch = 0;
  control: RunControl = idleControl;
  injected: readonly PlannedFault[] = [];

  constructor(
    readonly origin: LaunchOrigin,
    readonly plan: LaunchPlan,
    readonly startedMs: number,
  ) {}
}

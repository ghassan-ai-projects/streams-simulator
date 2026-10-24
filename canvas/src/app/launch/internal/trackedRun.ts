import type { CallOutcome, LaunchPlan } from "../../../model/launch";

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

  constructor(
    readonly origin: LaunchOrigin,
    readonly plan: LaunchPlan,
    readonly startedMs: number,
  ) {}
}

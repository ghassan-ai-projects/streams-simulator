import type { Scheduler } from "../../ports";
import type { LaunchState } from "./launchState";

type Done = Extract<LaunchState, { phase: "done" }>;

export interface Launching {
  snapshot(): LaunchState;
  subscribe(listener: () => void): () => void;
  reset(): void;
}

const finishedHoldMs = 2000;

export class RunWatch {
  private watched = false;

  constructor(
    private readonly launch: Launching,
    private readonly scheduler: Scheduler,
    private readonly plant: string,
  ) {
    this.notice(launch.snapshot());
  }

  shows(state: LaunchState): state is Done {
    return state.phase === "done" && this.owns(state) && (state.calls.length > 0 || this.watched);
  }

  holdFinished(then: (worldId: string) => void): () => void {
    let cancel = this.holdOf(this.launch.snapshot(), then);
    const unsubscribe = this.launch.subscribe(() => {
      cancel();
      cancel = this.holdOf(this.launch.snapshot(), then);
    });
    return () => {
      unsubscribe();
      cancel();
    };
  }

  private holdOf(state: LaunchState, then: (worldId: string) => void): () => void {
    this.notice(state);
    if (!this.shows(state) || state.calls.length > 0) return () => undefined;
    return this.scheduler.schedule(finishedHoldMs, () => {
      this.launch.reset();
      then(state.worldId);
    });
  }

  private notice(state: LaunchState): void {
    if (state.phase === "running" && this.owns(state)) this.watched = true;
  }

  private owns(state: LaunchState): boolean {
    return state.phase !== "idle" && state.origin.domain === this.plant;
  }
}

import type { BridgeFailure } from "../model/bridge";
import type { LaunchPlan, LaunchStep } from "../model/launch";
import { launchSteps } from "../model/launch";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

export interface LaunchOrigin {
  readonly domain: string;
  readonly title: string;
  readonly path: string;
}

export type LaunchState =
  | { readonly phase: "idle" }
  | {
      readonly phase: "running";
      readonly origin: LaunchOrigin;
      readonly worldId: string | null;
      readonly atS: number;
      readonly durationS: number;
      readonly canceling: boolean;
    }
  | { readonly phase: "done"; readonly origin: LaunchOrigin; readonly worldId: string }
  | { readonly phase: "failed"; readonly origin: LaunchOrigin; readonly failure: BridgeFailure };

const nowhere: LaunchOrigin = { domain: "", title: "", path: "" };

class Stopped extends Error {
  constructor(readonly failure: BridgeFailure) {
    super(failure.message);
  }
}

const canceled: BridgeFailure = { status: 0, code: "canceled", message: "The run was canceled.", issues: [] };

function worldOf(value: unknown): string {
  const record = typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};
  return typeof record.world_id === "string" ? record.world_id : "";
}

export class LaunchStore extends Observable<LaunchState> {
  private stop = false;
  private origin: LaunchOrigin = nowhere;
  private proposed: { readonly domain: string; readonly plan: LaunchPlan } | null = null;

  propose(domain: string, plan: LaunchPlan): void {
    this.proposed = { domain, plan };
  }

  proposal(domain: string): LaunchPlan | null {
    return this.proposed?.domain === domain ? this.proposed.plan : null;
  }

  constructor(private readonly bridge: Bridge) {
    super({ phase: "idle" });
  }

  async start(origin: LaunchOrigin, plan: LaunchPlan): Promise<void> {
    if (this.snapshot().phase === "running") return;
    const domain = origin.domain;
    this.origin = origin;
    this.proposed = null;
    this.stop = false;
    this.progress(null, 0, plan.durationS);
    let worldId = "";
    try {
      worldId = worldOf(await this.call("world.create", { domain, seed: plan.seed, explore: plan.explore }));
      this.progress(worldId, 0, plan.durationS);
      await this.steps(worldId, plan);
      await this.call("run.end", { world_id: worldId });
      this.set({ phase: "done", origin, worldId });
    } catch (error) {
      await this.abandon(worldId, error);
    }
  }

  cancel(): void {
    const state = this.snapshot();
    if (state.phase !== "running") return;
    this.stop = true;
    this.set({ ...state, canceling: true });
  }

  private progress(worldId: string | null, atS: number, durationS: number): void {
    this.set({ phase: "running", origin: this.origin, worldId, atS, durationS, canceling: this.stop });
  }

  reset(): void {
    if (this.snapshot().phase !== "running") this.set({ phase: "idle" });
  }

  private async steps(worldId: string, plan: LaunchPlan): Promise<void> {
    for (const step of launchSteps(plan)) {
      if (this.stop) throw new Stopped(canceled);
      await this.step(worldId, step);
      if (step.kind === "advance") this.progress(worldId, step.toS, plan.durationS);
    }
  }

  private async step(worldId: string, step: LaunchStep): Promise<void> {
    if (step.kind === "advance") await this.call("clock.advance", { world_id: worldId, by_ns: step.byNs });
    else await this.call("fault.inject", { world_id: worldId, entity_id: step.entity, fault: step.fault });
  }

  private async call(name: string, args: Readonly<Record<string, unknown>>): Promise<unknown> {
    const outcome = await this.bridge.action(name, args);
    if (!outcome.ok) throw new Stopped(outcome.failure);
    return outcome.value;
  }

  private async abandon(worldId: string, error: unknown): Promise<void> {
    if (worldId !== "") await this.bridge.action("world.destroy", { world_id: worldId });
    const failure =
      error instanceof Stopped ? error.failure : { ...canceled, code: "launch_failed", message: String(error) };
    this.set({ phase: "failed", origin: this.origin, failure });
  }
}

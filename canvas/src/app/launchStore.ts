import type { BridgeFailure } from "../model/bridge";
import type { LaunchPlan, LaunchStep } from "../model/launch";
import { launchSteps } from "../model/launch";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

export type LaunchState =
  | { readonly phase: "idle" }
  | { readonly phase: "running"; readonly worldId: string | null; readonly atS: number; readonly durationS: number }
  | { readonly phase: "done"; readonly worldId: string }
  | { readonly phase: "failed"; readonly failure: BridgeFailure };

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

  constructor(private readonly bridge: Bridge) {
    super({ phase: "idle" });
  }

  async start(domain: string, plan: LaunchPlan): Promise<void> {
    this.stop = false;
    this.set({ phase: "running", worldId: null, atS: 0, durationS: plan.durationS });
    let worldId = "";
    try {
      worldId = worldOf(await this.call("world.create", { domain, seed: plan.seed, explore: plan.explore }));
      this.set({ phase: "running", worldId, atS: 0, durationS: plan.durationS });
      await this.steps(worldId, plan);
      await this.call("run.end", { world_id: worldId });
      this.set({ phase: "done", worldId });
    } catch (error) {
      await this.abandon(worldId, error);
    }
  }

  cancel(): void {
    this.stop = true;
  }

  reset(): void {
    if (this.snapshot().phase !== "running") this.set({ phase: "idle" });
  }

  private async steps(worldId: string, plan: LaunchPlan): Promise<void> {
    for (const step of launchSteps(plan)) {
      if (this.stop) throw new Stopped(canceled);
      await this.step(worldId, step);
      if (step.kind === "advance") this.set({ phase: "running", worldId, atS: step.toS, durationS: plan.durationS });
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
    this.set({ phase: "failed", failure });
  }
}

import type { BridgeFailure } from "../model/bridge";
import type { CallOutcome, LaunchPlan, LaunchStep } from "../model/launch";
import { failureWords, launchSteps, outcomeOf, paceOf } from "../model/launch";
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
      readonly plan: LaunchPlan;
      readonly startedMs: number;
      readonly atMs: number;
      readonly stopping: boolean;
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
      readonly failure: BridgeFailure;
      readonly words: string;
      readonly kept: string | null;
    };

const nowhere: LaunchOrigin = { domain: "", title: "", path: "" };

class Stopped extends Error {
  constructor(readonly failure: BridgeFailure) {
    super(failure.message);
  }
}

const broken: BridgeFailure = {
  status: 0,
  code: "launch_failed",
  message: "The run stopped unexpectedly.",
  issues: [],
};

function worldOf(value: unknown): string {
  const record = typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};
  return typeof record.world_id === "string" ? record.world_id : "";
}

export interface Proposal {
  readonly plan: LaunchPlan;
  readonly left: readonly string[];
  readonly from: string;
}

export class LaunchStore extends Observable<LaunchState> {
  private stop = false;
  private origin: LaunchOrigin = nowhere;
  private plan: LaunchPlan = { seed: 0, durationS: 0, explore: true, faults: [], calls: [] };
  private outcomes: readonly CallOutcome[] = [];
  private startedMs = 0;
  private reachedS = 0;
  private samples: readonly { readonly ms: number; readonly s: number }[] = [];
  private interrupt: (() => void) | null = null;
  private proposed: (Proposal & { readonly domain: string }) | null = null;
  private current: LaunchStep | null = null;

  propose(domain: string, plan: LaunchPlan, left: readonly string[] = [], from = ""): void {
    this.proposed = { domain, plan, left, from };
  }

  forgetProposal(): void {
    this.proposed = null;
  }

  proposal(domain: string): Proposal | null {
    return this.proposed?.domain === domain
      ? { plan: this.proposed.plan, left: this.proposed.left, from: this.proposed.from }
      : null;
  }

  constructor(
    private readonly bridge: Bridge,
    private readonly now: () => number = () => Date.now(),
    private readonly sleep: (ms: number) => Promise<void> = (ms) =>
      new Promise((resolve) => {
        window.setTimeout(resolve, ms);
      }),
  ) {
    super({ phase: "idle" });
  }

  async start(origin: LaunchOrigin, plan: LaunchPlan): Promise<void> {
    if (this.snapshot().phase === "running") return;
    const domain = origin.domain;
    this.origin = origin;
    this.plan = plan;
    this.startedMs = this.now();
    this.proposed = null;
    this.stop = false;
    this.outcomes = [];
    this.reachedS = 0;
    this.samples = [];
    this.progress(null, 0);
    let worldId = "";
    try {
      worldId = worldOf(await this.call("world.create", { domain, seed: plan.seed, explore: plan.explore }));
      this.progress(worldId, 0);
      await this.steps(worldId, plan);
      await this.call("run.end", { world_id: worldId });
      const { stop: early, reachedS, outcomes: calls } = this;
      this.set({ phase: "done", origin, worldId, early, plan, reachedS, calls });
    } catch (error) {
      await this.abandon(worldId, error);
    }
  }

  stopEarly(): void {
    const state = this.snapshot();
    if (state.phase !== "running") return;
    this.stop = true;
    this.set({ ...state, stopping: true });
    this.interrupt?.();
  }

  private progress(worldId: string | null, atS: number): void {
    const { origin, plan, startedMs } = this;
    const atMs = this.now();
    this.reachedS = atS;
    this.set({
      phase: "running",
      origin,
      worldId,
      atS,
      durationS: plan.durationS,
      plan,
      startedMs,
      atMs,
      stopping: this.stop,
      rate: this.rateAt(atMs, atS),
      calls: this.outcomes,
    });
  }

  reset(): void {
    if (this.snapshot().phase !== "running") this.set({ phase: "idle" });
  }

  private async steps(worldId: string, plan: LaunchPlan): Promise<void> {
    let advanced = false;
    for (const step of launchSteps(plan, plan.pace === "watch" ? 30 : 20)) {
      if (this.stop && advanced) return;
      await this.step(worldId, step);
      advanced ||= step.kind === "advance";
      if (step.kind === "advance") {
        this.progress(worldId, step.toS);
        await this.pace(step.toS);
      }
    }
  }

  private rateAt(atMs: number, atS: number): number | null {
    this.samples = [...this.samples.filter((sample) => atMs - sample.ms <= 5000), { ms: atMs, s: atS }];
    const first = this.samples[0];
    if (first === undefined || atMs - first.ms < 1000) return null;
    return (atS - first.s) / ((atMs - first.ms) / 1000);
  }

  private async pace(toS: number): Promise<void> {
    const rate = paceOf(this.plan);
    const wait = rate === null ? 0 : this.startedMs + (toS / rate) * 1000 - this.now();
    if (wait <= 0 || this.stop) return;
    const woken = new Promise<void>((resolve) => {
      this.interrupt = resolve;
    });
    await Promise.race([this.sleep(wait), woken]);
    this.interrupt = null;
  }

  private async step(worldId: string, step: LaunchStep): Promise<void> {
    this.current = step;
    if (step.kind === "advance") await this.call("clock.advance", { world_id: worldId, by_ns: step.byNs });
    else if (step.kind === "inject")
      await this.call("fault.inject", { world_id: worldId, entity_id: step.entity, fault: step.fault });
    else await this.invoke(worldId, step);
    this.current = null;
  }

  private async invoke(worldId: string, step: Extract<LaunchStep, { kind: "invoke" }>): Promise<void> {
    const { entity, effector, args } = step;
    const acknowledged = await this.call("effector.invoke", {
      world_id: worldId,
      entity_id: entity,
      effector,
      args,
    });
    this.outcomes = [...this.outcomes, outcomeOf(step, acknowledged)];
  }

  private async call(name: string, args: Readonly<Record<string, unknown>>): Promise<unknown> {
    const outcome = await this.bridge.action(name, args);
    if (!outcome.ok) throw new Stopped(outcome.failure);
    return outcome.value;
  }

  private async abandon(worldId: string, error: unknown): Promise<void> {
    if (worldId !== "") await this.bridge.action("world.destroy", { world_id: worldId });
    const failure = error instanceof Stopped ? error.failure : { ...broken, message: String(error) };
    const words = failureWords(this.current);
    this.current = null;
    this.set({ phase: "failed", origin: this.origin, failure, words, kept: worldId === "" ? null : worldId });
  }
}

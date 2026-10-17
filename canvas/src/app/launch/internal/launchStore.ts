import type { LaunchPlan, LaunchStep } from "../../../model/launch";
import { failureWords, outcomeOf, paceOf, stepsFor } from "../../../model/launch";
import { Observable } from "../../shared";
import type { Clock, LaunchFailure, LaunchGateway, Outcome, Scheduler } from "../../ports";
import type { Proposal } from "./proposalSlot";
import { ProposalSlot } from "./proposalSlot";
import type { LaunchState } from "./launchState";
import type { LaunchOrigin } from "./runCursor";
import { RunCursor } from "./runCursor";
import { RunWatch } from "./runWatch";

export type { Proposal } from "./proposalSlot";
export type { LaunchState } from "./launchState";

const nowhere: LaunchOrigin = { domain: "", title: "", path: "" };
const noPlan: LaunchPlan = { seed: 0, durationS: 0, explore: true, faults: [], calls: [] };

class Stopped extends Error {
  constructor(readonly failure: LaunchFailure) {
    super(failure.diagnostic.message);
  }
}

const broken: LaunchFailure = {
  kind: "unavailable",
  diagnostic: { displayCode: "launch_failed", message: "The run stopped unexpectedly." },
};

export class LaunchStore extends Observable<LaunchState> {
  private cursor = new RunCursor(nowhere, noPlan, 0);
  private readonly proposals = new ProposalSlot();

  constructor(
    private readonly gateway: LaunchGateway,
    private readonly clock: Clock,
    private readonly scheduler: Scheduler,
  ) {
    super({ phase: "idle" });
  }

  propose(domain: string, plan: LaunchPlan, left: readonly string[] = [], from = ""): void {
    this.proposals.propose(domain, plan, left, from);
  }

  forgetProposal(): void {
    this.proposals.forget();
  }

  proposal(domain: string): Proposal | null {
    return this.proposals.proposalFor(domain);
  }

  watch(plant: string): RunWatch {
    return new RunWatch(this, this.scheduler, plant);
  }

  async start(origin: LaunchOrigin, plan: LaunchPlan): Promise<void> {
    if (this.snapshot().phase === "running") return;
    this.cursor = new RunCursor(origin, plan, this.clock.nowMs());
    this.proposals.forget();
    this.progress(null, 0);
    let worldId = "";
    try {
      worldId = this.require(
        await this.gateway.createWorld({ domain: origin.domain, seed: plan.seed, explore: plan.explore }),
      ).worldId;
      this.progress(worldId, 0);
      await this.steps(worldId, plan);
      this.require(await this.gateway.endRun(worldId));
      const { stop: early, reachedS, outcomes: calls } = this.cursor;
      this.set({ phase: "done", origin, worldId, early, plan, reachedS, calls });
    } catch (error) {
      await this.abandon(worldId, error);
    }
  }

  stopEarly(): void {
    const state = this.snapshot();
    if (state.phase !== "running") return;
    this.cursor.stop = true;
    this.set({ ...state, stopping: true });
    this.cursor.interrupt?.();
  }

  reset(): void {
    if (this.snapshot().phase !== "running") this.set({ phase: "idle" });
  }

  private progress(worldId: string | null, atS: number): void {
    const { origin, plan, startedMs, stop, outcomes } = this.cursor;
    const atMs = this.clock.nowMs();
    const rate = this.cursor.sampleRate(atMs, atS);
    this.cursor.reachedS = atS;
    this.set({
      phase: "running",
      origin,
      worldId,
      atS,
      durationS: plan.durationS,
      plan,
      startedMs,
      atMs,
      stopping: stop,
      rate,
      calls: outcomes,
    });
  }

  private async steps(worldId: string, plan: LaunchPlan): Promise<void> {
    let advanced = false;
    for (const step of stepsFor(plan)) {
      if (this.cursor.stop && advanced) return;
      await this.step(worldId, step);
      advanced ||= step.kind === "advance";
      if (step.kind === "advance") {
        this.progress(worldId, step.toS);
        await this.pace(step.toS);
      }
    }
  }

  private async pace(toS: number): Promise<void> {
    const cursor = this.cursor;
    const rate = paceOf(cursor.plan);
    const wait = rate === null ? 0 : cursor.startedMs + (toS / rate) * 1000 - this.clock.nowMs();
    if (wait <= 0 || cursor.stop) return;
    const woken = new Promise<void>((resolve) => {
      cursor.interrupt = resolve;
    });
    const controller = new AbortController();
    try {
      await Promise.race([this.scheduler.delay(wait, controller.signal), woken]);
    } finally {
      controller.abort();
      cursor.interrupt = null;
    }
  }

  private async step(worldId: string, step: LaunchStep): Promise<void> {
    this.cursor.current = step;
    if (step.kind === "advance") this.require(await this.gateway.advance(worldId, step.byNs));
    else if (step.kind === "inject") this.require(await this.gateway.inject(worldId, step.entity, step.fault));
    else await this.invoke(worldId, step);
    this.cursor.current = null;
  }

  private async invoke(worldId: string, step: Extract<LaunchStep, { kind: "invoke" }>): Promise<void> {
    const { entity, effector, args } = step;
    const acknowledged = this.require(await this.gateway.invoke(worldId, entity, effector, args));
    this.cursor.outcomes = [...this.cursor.outcomes, outcomeOf(step, acknowledged, this.clock.nowMs())];
    this.progress(worldId, this.cursor.reachedS);
  }

  private require<T>(outcome: Outcome<T, LaunchFailure>): T {
    if (!outcome.ok) throw new Stopped(outcome.failure);
    return outcome.value;
  }

  private async abandon(worldId: string, error: unknown): Promise<void> {
    if (worldId !== "") await this.gateway.destroyWorld(worldId);
    const failure =
      error instanceof Stopped
        ? error.failure
        : { ...broken, diagnostic: { ...broken.diagnostic, message: String(error) } };
    const words = failureWords(this.cursor.current);
    this.cursor.current = null;
    this.set({ phase: "failed", origin: this.cursor.origin, failure, words, kept: worldId === "" ? null : worldId });
  }
}

import type { JobStatus, LaunchPlan } from "../../../model/launch";
import {
  idleControl,
  jobActive,
  jobRequestOf,
  observedCalls,
  reachedS,
  recoveredPlanOf,
  refusalWords,
  unobserved,
} from "../../../model/launch";
import type { ArgValue, ControlName, RunControl } from "../../../model/launch";
import { Observable } from "../../shared";
import type { Clock, LaunchFailure, LaunchGateway, Outcome, Scheduler } from "../../ports";
import { success } from "../../ports";
import type { Proposal } from "./proposalSlot";
import { ProposalSlot } from "./proposalSlot";
import type { LaunchState } from "./launchState";
import { followedState, lostState, refusedStart, startingState } from "./jobView";
import { StatusPoll } from "./statusPoll";
import type { LaunchOrigin } from "./trackedRun";
import { TrackedRun } from "./trackedRun";
import { RunWatch } from "./runWatch";
import type { LaunchStore as LaunchStoreApi } from "./api";
import { ConnectStore } from "./connectStore";
import { PlanCatalog } from "./planCatalog";

const nowhere: LaunchOrigin = { domain: "", title: "", path: "" };
const noPlan: LaunchPlan = { seed: 0, durationS: 0, explore: true, faults: [], perturbations: [], calls: [] };
const patientPolls = 3;

type Reading = Outcome<JobStatus, LaunchFailure>;

export class LaunchStore extends Observable<LaunchState> {
  private run = new TrackedRun(nowhere, noPlan, 0);
  private poll: StatusPoll | null = null;
  private readonly proposals = new ProposalSlot();
  readonly connect: ConnectStore;
  readonly catalog: PlanCatalog;

  constructor(
    private readonly gateway: LaunchGateway,
    private readonly clock: Clock,
    private readonly scheduler: Scheduler,
  ) {
    super({ phase: "idle" });
    this.connect = new ConnectStore(gateway, scheduler);
    this.catalog = new PlanCatalog(gateway);
  }

  propose(domain: string, plan: LaunchPlan, left: readonly string[] = [], from = ""): void {
    this.proposals.propose(domain, { plan, left, from, fresh: false });
  }

  proposeFresh(domain: string, plan: LaunchPlan, from: string): void {
    this.proposals.propose(domain, { plan, left: [], from, fresh: true });
  }

  nowMs(): number {
    return this.clock.nowMs();
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
    const run = this.track(new TrackedRun(origin, plan, this.clock.nowMs()));
    this.proposals.forget();
    this.set(startingState(run));
    const started = await this.gateway.startJob(jobRequestOf(origin.domain, plan));
    if (!started.ok) this.set(refusedStart(run, started.failure));
    else await this.follow(run, started.value);
  }

  async resume(originOf: (domain: string) => LaunchOrigin | null): Promise<void> {
    if (this.snapshot().phase !== "idle") return;
    const listed = await this.gateway.listJobs();
    if (!listed.ok || this.snapshot().phase !== "idle") return;
    const status = listed.value.find(jobActive);
    if (status === undefined) return;
    const origin = originOf(status.domain) ?? { domain: status.domain, title: status.domain, path: "" };
    const run = this.track(new TrackedRun(origin, recoveredPlanOf(status), this.clock.nowMs()));
    run.calls = observedCalls([], status, unobserved);
    await this.follow(run, status);
  }

  pause(): Promise<void> {
    return this.control("pause", (worldId) => this.gateway.pauseJob(worldId));
  }

  continueRun(): Promise<void> {
    return this.control("resume", (worldId) => this.gateway.resumeJob(worldId));
  }

  step(): Promise<void> {
    const state = this.snapshot();
    if (state.phase !== "running" || !state.paused) return Promise.resolve();
    return this.control("step", (worldId) => this.gateway.stepJob(worldId));
  }

  setRate(rate: number): Promise<void> {
    return this.control("rate", (worldId) => this.gateway.setJobRate(worldId, rate));
  }

  invokeNow(entity: string, effector: string, args: Readonly<Record<string, ArgValue>>): Promise<void> {
    return this.control("call", async (worldId) => {
      const done = await this.gateway.invokeEffectorNow(worldId, entity, effector, args);
      return done.ok ? success(done.value.status) : done;
    });
  }

  injectNow(entity: string, fault: string): Promise<void> {
    const run = this.run;
    if (!run.plan.explore) return Promise.resolve();
    return this.control("fault", async (worldId) => {
      const done = await this.gateway.injectFaultNow(worldId, entity, fault);
      if (!done.ok) return done;
      run.injected = [...run.injected, { entity, fault, atS: reachedS(done.value.status) }];
      return success(done.value.status);
    });
  }

  stopEarly(): void {
    const state = this.snapshot();
    if (state.phase !== "running" || state.stopping) return;
    this.run.stopping = true;
    this.set({ ...state, stopping: true });
    if (state.worldId !== null) void this.requestStop(this.run, state.worldId);
  }

  reset(): void {
    if (this.snapshot().phase !== "running") this.set({ phase: "idle" });
  }

  private async control(name: ControlName, call: (worldId: string) => Promise<Reading>): Promise<void> {
    const run = this.run;
    const state = this.snapshot();
    if (state.phase !== "running" || state.worldId === null || state.stopping || state.control.pending !== null) return;
    this.setControl(run, { pending: name, refusal: "" });
    const done = await call(state.worldId);
    if (this.run !== run) return;
    if (!done.ok) this.setControl(run, { pending: null, refusal: refusalWords(done.failure.diagnostic.message) });
    else this.controlAnswered(run, done.value);
  }

  private setControl(run: TrackedRun, control: RunControl): void {
    run.control = control;
    run.epoch += 1;
    const state = this.snapshot();
    if (state.phase === "running") this.set({ ...state, control });
  }

  private controlAnswered(run: TrackedRun, status: JobStatus): void {
    run.control = idleControl;
    run.epoch += 1;
    if (!this.show(run, status)) this.poll?.end();
  }

  private track(run: TrackedRun): TrackedRun {
    this.poll?.end();
    this.poll = null;
    this.run = run;
    return run;
  }

  private async follow(run: TrackedRun, status: JobStatus): Promise<void> {
    if (!this.show(run, status)) return;
    const worldId = status.worldId;
    this.poll = new StatusPoll(
      this.scheduler,
      () => this.read(run, worldId),
      (reading) => this.took(run, reading),
    );
    this.poll.begin();
    if (run.stopping && (status.state === "running" || status.state === "paused")) await this.requestStop(run, worldId);
  }

  private async read(run: TrackedRun, worldId: string): Promise<Reading | null> {
    const epoch = run.epoch;
    const reading = await this.gateway.jobStatus(worldId);
    return run.epoch === epoch ? reading : null;
  }

  private show(run: TrackedRun, status: JobStatus): boolean {
    const atMs = this.clock.nowMs();
    run.worldId = status.worldId;
    run.calls = observedCalls(run.calls, status, atMs);
    this.set(followedState(run, status, atMs));
    return jobActive(status);
  }

  private took(run: TrackedRun, reading: Reading): boolean {
    if (this.run !== run) return false;
    if (!reading.ok) return this.missed(run, reading.failure);
    run.missed = 0;
    return this.show(run, reading.value);
  }

  private missed(run: TrackedRun, failure: LaunchFailure): boolean {
    run.missed += 1;
    if (failure.kind === "unavailable" && run.missed < patientPolls) return true;
    this.set(lostState(run, failure));
    return false;
  }

  private async requestStop(run: TrackedRun, worldId: string): Promise<void> {
    const stopped = await this.gateway.stopJob(worldId);
    if (this.run !== run) return;
    if (stopped.ok) this.stopAnswered(run, stopped.value);
    else this.stopRefused(run);
  }

  private stopAnswered(run: TrackedRun, status: JobStatus): void {
    if (!this.show(run, status)) this.poll?.end();
  }

  private stopRefused(run: TrackedRun): void {
    run.stopping = false;
    const state = this.snapshot();
    if (state.phase === "running") this.set({ ...state, stopping: false });
  }
}

export function createLaunchStore(gateway: LaunchGateway, clock: Clock, scheduler: Scheduler): LaunchStoreApi {
  return new LaunchStore(gateway, clock, scheduler);
}

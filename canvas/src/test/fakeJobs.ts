import { failure as failed, success } from "../app/ports";
import type { Clock } from "../app/ports";
import type { Canned } from "./fakeFailures";
import { cannedFailure } from "./fakeFailures";

type Args = Readonly<Record<string, unknown>>;

interface WireCall {
  readonly entity_id: string;
  readonly effector: string;
  readonly at_s: number;
}

interface Sim {
  readonly worldId: string;
  readonly request: Args;
  readonly startedMs: number;
  readonly durationS: number;
  readonly rate: number;
  halted: "stopped" | "failed" | null;
  stopping: boolean;
  polled: boolean;
  frozenS: number;
  failure: string;
}

const jobStartNs = 1_767_225_600e9;

const noClock: Clock = { nowMs: () => 0 };

export class FakeJobs {
  worldId = "w-7";
  requests: Args[] = [];
  statusReads = 0;
  private clock: Clock = noClock;
  private sim: Sim | null = null;
  private readonly refused = new Map<string, string>();

  drive(clock: Clock): void {
    this.clock = clock;
  }

  refuse(effector: string, reason: string): void {
    this.refused.set(effector, reason);
  }

  fail(message: string): void {
    const sim = this.running();
    if (sim === undefined) return;
    sim.frozenS = this.reachedS(sim);
    sim.halted = "failed";
    sim.failure = message;
  }

  adopt(reachedS: number, request: Args): void {
    const durationS = Number(request.duration_s);
    const startedMs = this.clock.nowMs() - (reachedS * 1000) / Number(request.rate);
    this.sim = { ...this.fresh(request, durationS), startedMs };
  }

  answer(name: string, args: Args): Canned<unknown> | undefined {
    const handlers: Record<string, () => Canned<unknown>> = {
      "job.start": () => this.start(args),
      "job.status": () => this.read(args, false),
      "job.stop": () => this.read(args, true),
      "job.list": () => success({ jobs: this.sim === null ? [] : [this.wire(this.sim)] }),
    };
    return handlers[name]?.();
  }

  private start(args: Args): Canned<unknown> {
    if (this.running() !== undefined)
      return failed(cannedFailure("job_active", 409, "a job is already running or stopping"));
    this.requests.push(args);
    this.sim = this.fresh(args, Number(args.duration_s));
    return success(this.wire(this.sim));
  }

  private fresh(request: Args, durationS: number): Sim {
    return {
      worldId: this.worldId,
      request,
      startedMs: this.clock.nowMs(),
      durationS,
      rate: Number(request.rate ?? 0),
      halted: null,
      stopping: false,
      polled: false,
      frozenS: 0,
      failure: "",
    };
  }

  private read(args: Args, stop: boolean): Canned<unknown> {
    const sim = this.sim;
    if (sim === null || sim.worldId !== args.world_id)
      return failed(cannedFailure("world_not_found", 404, `no job for world "${String(args.world_id)}"`));
    if (stop) return success(this.wire(this.stopped(sim)));
    this.statusReads += 1;
    sim.polled = true;
    return success(this.wire(this.settled(sim)));
  }

  private stopped(sim: Sim): Sim {
    if (sim.halted === null && !sim.stopping && this.reachedS(sim) < sim.durationS) {
      sim.frozenS = this.reachedS(sim);
      sim.stopping = true;
    }
    return sim;
  }

  private settled(sim: Sim): Sim {
    if (sim.stopping && sim.halted === null) sim.halted = "stopped";
    return sim;
  }

  private running(): Sim | undefined {
    const sim = this.sim;
    const live = sim !== null && sim.halted === null && (sim.stopping || this.reachedS(sim) < sim.durationS);
    return live ? sim : undefined;
  }

  private reachedS(sim: Sim): number {
    if (sim.halted !== null || sim.stopping) return sim.frozenS;
    if (sim.rate === 0) return sim.polled ? sim.durationS : 0;
    return Math.min(sim.durationS, ((this.clock.nowMs() - sim.startedMs) / 1000) * sim.rate);
  }

  private stateOf(sim: Sim): string {
    if (sim.halted !== null) return sim.halted;
    if (sim.stopping) return "stopping";
    return this.reachedS(sim) >= sim.durationS ? "completed" : "running";
  }

  private calls(sim: Sim): readonly object[] {
    const reached = this.reachedS(sim);
    const planned = (Array.isArray(sim.request.calls) ? sim.request.calls : []) as WireCall[];
    return [...planned]
      .sort((a, b) => a.at_s - b.at_s)
      .filter((call) => call.at_s <= reached)
      .map((call) => this.outcome(call));
  }

  private outcome(call: WireCall): object {
    const reason = this.refused.get(call.effector);
    const { at_s, entity_id, effector } = call;
    return reason === undefined
      ? { at_s, entity_id, effector, accepted: true, effect_eta_ns: 30e9 }
      : { at_s, entity_id, effector, accepted: false, reason };
  }

  private wire(sim: Sim): object {
    return {
      world_id: sim.worldId,
      run_id: `${sim.worldId}-run`,
      domain: sim.request.domain,
      seed: sim.request.seed,
      state: this.stateOf(sim),
      start_ns: jobStartNs,
      reached_ns: jobStartNs + this.reachedS(sim) * 1e9,
      requested_end_ns: jobStartNs + sim.durationS * 1e9,
      rate: sim.rate,
      explore: sim.request.explore === true,
      calls: this.calls(sim),
      ...(sim.failure === "" ? {} : { failure: sim.failure }),
    };
  }
}

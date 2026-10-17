import type { CallOutcome, LaunchPlan, LaunchStep } from "../../../model/launch";

export interface LaunchOrigin {
  readonly domain: string;
  readonly title: string;
  readonly path: string;
}

interface Sample {
  readonly ms: number;
  readonly s: number;
}

const sampledMs = 5000;
const settledMs = 1000;

export class RunCursor {
  stop = false;
  outcomes: readonly CallOutcome[] = [];
  reachedS = 0;
  interrupt: (() => void) | null = null;
  current: LaunchStep | null = null;
  private samples: readonly Sample[] = [];

  constructor(
    readonly origin: LaunchOrigin,
    readonly plan: LaunchPlan,
    readonly startedMs: number,
  ) {}

  sampleRate(atMs: number, atS: number): number | null {
    this.samples = [...this.samples.filter((sample) => atMs - sample.ms <= sampledMs), { ms: atMs, s: atS }];
    const first = this.samples[0];
    if (first === undefined || atMs - first.ms < settledMs) return null;
    return (atS - first.s) / ((atMs - first.ms) / 1000);
  }
}

import type { BridgeFailure } from "../model/bridge";
import { PlaybackIndex, parseJournalRecord } from "../model/playback";
import type { Report } from "../model/report";
import { parseReport, verifyMatches } from "../model/report";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

type Load<T> =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: BridgeFailure }
  | { readonly phase: "ready"; readonly value: T };

export interface ReviewState {
  readonly worldId: string | null;
  readonly report: Load<Report>;
  readonly verified: Load<boolean>;
  readonly playback: {
    readonly phase: "loading" | "ready" | "failed";
    readonly records: number;
    readonly adapter: string;
  };
  readonly positionNs: number;
  readonly selected: string | null;
}

const idle: ReviewState = {
  worldId: null,
  report: { phase: "loading" },
  verified: { phase: "loading" },
  playback: { phase: "loading", records: 0, adapter: "" },
  positionNs: 0,
  selected: null,
};

export class ReviewStore extends Observable<ReviewState> {
  index = new PlaybackIndex();
  private stop: (() => void) | null = null;
  private pending = 0;

  constructor(
    private readonly bridge: Bridge,
    private readonly schedule: (flush: () => void) => void = (flush) => {
      setTimeout(flush, 250);
    },
  ) {
    super(idle);
  }

  open(worldId: string): void {
    if (this.snapshot().worldId === worldId) return;
    this.close();
    this.index = new PlaybackIndex();
    this.set({ ...idle, worldId });
    void this.loadReport(worldId);
    void this.loadVerify(worldId);
    this.stop = this.bridge.followRun(worldId, {
      record: (data) => {
        this.take(worldId, data);
      },
      closed: () => {
        this.finish(worldId, "ready");
      },
      failed: () => {
        this.finish(worldId, "failed");
      },
    });
  }

  close(): void {
    this.stop?.();
    this.stop = null;
  }

  seek(positionNs: number): void {
    this.set({ ...this.snapshot(), positionNs });
  }

  select(entity: string | null): void {
    this.set({ ...this.snapshot(), selected: entity });
  }

  private take(worldId: string, data: string): void {
    const entry = parseJournalRecord(data);
    if (entry.kind === "observation") this.index.add(entry.observation);
    if (entry.kind === "opened") this.update(worldId, { adapter: entry.adapter });
    this.pending += 1;
    if (this.pending === 1) {
      this.schedule(() => {
        this.pending = 0;
        this.update(worldId, {});
      });
    }
  }

  private finish(worldId: string, phase: "ready" | "failed"): void {
    this.update(worldId, { phase });
    const bounds = this.index.bounds();
    const state = this.snapshot();
    if (bounds !== null && state.worldId === worldId && state.positionNs === 0) this.seek(bounds.endNs);
  }

  private update(worldId: string, playback: Partial<ReviewState["playback"]>): void {
    const state = this.snapshot();
    if (state.worldId !== worldId) return;
    this.set({ ...state, playback: { ...state.playback, records: this.index.size, ...playback } });
  }

  private async loadReport(worldId: string): Promise<void> {
    const outcome = await this.bridge.action("run.report", { run: worldId });
    if (this.snapshot().worldId !== worldId) return;
    this.set({
      ...this.snapshot(),
      report: outcome.ok
        ? { phase: "ready", value: parseReport(outcome.value) }
        : { phase: "failed", failure: outcome.failure },
    });
  }

  private async loadVerify(worldId: string): Promise<void> {
    const outcome = await this.bridge.action("run.verify", { run: worldId });
    if (this.snapshot().worldId !== worldId) return;
    this.set({
      ...this.snapshot(),
      verified: outcome.ok
        ? { phase: "ready", value: verifyMatches(outcome.value) }
        : { phase: "failed", failure: outcome.failure },
    });
  }
}

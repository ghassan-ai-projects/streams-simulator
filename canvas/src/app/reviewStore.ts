import type { BridgeFailure } from "../model/bridge";
import { PlaybackIndex, parseJournalRecord } from "../model/playback";
import type { Badge, Report } from "../model/report";
import type { Replay } from "../model/report";
import { badgeOf, parseReplay, parseReport } from "../model/report";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

type Load<T> =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: BridgeFailure }
  | { readonly phase: "ready"; readonly value: T };

export interface ReviewState {
  readonly worldId: string | null;
  readonly report: Load<Report>;
  readonly verified: Load<Replay>;
  readonly playback: {
    readonly phase: "loading" | "ready" | "failed";
    readonly records: number;
    readonly adapter: string;
  };
  readonly positionNs: number;
  readonly selected: string | null;
  readonly runDomain: Load<unknown>;
  readonly tab: "report" | "inspect";
  readonly openedAt: number;
}

const idle: ReviewState = {
  worldId: null,
  report: { phase: "loading" },
  verified: { phase: "loading" },
  playback: { phase: "loading", records: 0, adapter: "" },
  positionNs: 0,
  selected: null,
  runDomain: { phase: "loading" },
  tab: "report",
  openedAt: 0,
};

export class ReviewStore extends Observable<ReviewState> {
  index = new PlaybackIndex();
  private readonly badges = new Map<string, Badge>();
  private stop: (() => void) | null = null;
  private pending = 0;

  constructor(
    private readonly bridge: Bridge,
    private readonly schedule: (flush: () => void) => void = (flush) => {
      setTimeout(flush, 250);
    },
    private readonly now: () => number = () => Date.now(),
  ) {
    super(idle);
  }

  open(worldId: string): void {
    if (this.snapshot().worldId === worldId) return;
    this.close();
    this.index = new PlaybackIndex();
    this.set({ ...idle, worldId, openedAt: this.now() });
    void this.loadReport(worldId);
    void this.loadVerify(worldId);
    void this.loadRunDomain(worldId);
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

  badgeFor(worldId: string): Badge | null {
    return this.badges.get(worldId) ?? null;
  }

  protected override set(next: ReviewState): void {
    if (next.worldId !== null && next.report.phase === "ready" && next.verified.phase !== "loading") {
      this.badges.set(next.worldId, badgeOf(next.report.value, next.verified));
    }
    super.set(next);
  }

  seek(positionNs: number): void {
    this.set({ ...this.snapshot(), positionNs });
  }

  select(entity: string | null): void {
    this.set({ ...this.snapshot(), selected: entity, tab: entity === null ? this.snapshot().tab : "inspect" });
  }

  showTab(tab: ReviewState["tab"]): void {
    this.set({ ...this.snapshot(), tab });
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

  private async loadRunDomain(worldId: string): Promise<void> {
    const file = await this.bridge.readFile(`runs/${worldId}/run.json`);
    if (this.snapshot().worldId !== worldId) return;
    const spec = file.ok ? domainSpec(file.value.content) : null;
    const runDomain: Load<unknown> =
      spec !== null
        ? { phase: "ready", value: spec }
        : {
            phase: "failed",
            failure: file.ok
              ? { status: 0, code: "no_domain_spec", message: "run.json has no domain_spec", issues: [] }
              : file.failure,
          };
    this.set({ ...this.snapshot(), runDomain });
  }

  private async loadVerify(worldId: string): Promise<void> {
    const outcome = await this.bridge.action("run.verify", { run: worldId });
    if (this.snapshot().worldId !== worldId) return;
    const replay = parseReplay(outcome.ok ? outcome.value : outcome.failure.result);
    const verified: Load<Replay> =
      replay !== null
        ? { phase: "ready", value: replay }
        : outcome.ok
          ? {
              phase: "failed",
              failure: { status: 0, code: "verify_unreadable", message: "verify gave no verdict", issues: [] },
            }
          : { phase: "failed", failure: outcome.failure };
    this.set({ ...this.snapshot(), verified });
  }
}

function domainSpec(content: string): unknown {
  try {
    const parsed = JSON.parse(content) as unknown;
    return typeof parsed === "object" && parsed !== null && "domain_spec" in parsed ? parsed.domain_spec : null;
  } catch {
    return null;
  }
}

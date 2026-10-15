import type { BridgeFailure, RunEntry } from "../model/bridge";
import { PlaybackIndex, advance, parseJournalRecord } from "../model/playback";
import type { Badge, Report } from "../model/report";
import type { Replay } from "../model/report";
import { badgeOf, parseReplay, parseReport } from "../model/report";
import { unfinishedReason } from "../model/workspace";
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
  readonly picked: { readonly atNs: number; readonly entity: string } | null;
  readonly openedAt: number;
  readonly playing: boolean;
  readonly speed: number;
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
  picked: null,
  openedAt: 0,
  playing: false,
  speed: 60,
};

const absentFile: BridgeFailure = {
  status: 404,
  code: "not_found",
  message: "run.json is absent: the run is still open or did not end",
  issues: [],
};

function absentLoads(known: RunEntry | undefined): Pick<ReviewState, "verified" | "runDomain"> | Record<string, never> {
  if (known === undefined || !known.ended || known.complete) return {};
  return {
    verified: {
      phase: "failed",
      failure: {
        status: 422,
        code: "run_incomplete",
        message: `the run is incomplete: ${unfinishedReason(known)}`,
        issues: [],
      },
    },
    runDomain: { phase: "failed", failure: absentFile },
  };
}

export class ReviewStore extends Observable<ReviewState> {
  index = new PlaybackIndex();
  private readonly badges = new Map<string, Badge>();
  private stop: (() => void) | null = null;
  private pending = 0;
  private refreshed = false;
  private knownIncomplete = false;

  constructor(
    private readonly bridge: Bridge,
    private readonly schedule: (flush: () => void) => void = (flush) => {
      setTimeout(flush, 250);
    },
    private readonly now: () => number = () => Date.now(),
  ) {
    super(idle);
  }

  open(worldId: string, known?: RunEntry): void {
    const state = this.snapshot();
    if (state.worldId === worldId && (this.stop !== null || state.playback.phase !== "loading")) return;
    this.close();
    this.refreshed = false;
    this.knownIncomplete = known !== undefined && known.ended && !known.complete;
    this.index = new PlaybackIndex();
    this.set({ ...idle, worldId, openedAt: this.now(), ...absentLoads(known) });
    void this.loadReport(worldId);
    if (!this.knownIncomplete) {
      void this.loadVerify(worldId);
      void this.loadRunDomain(worldId);
    }
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

  play(): void {
    const bounds = this.index.bounds();
    const settled = this.index.settled();
    if (bounds === null || settled === null) return;
    const state = this.snapshot();
    const fresh = state.positionNs >= bounds.endNs || state.positionNs < bounds.startNs;
    this.set({ ...state, playing: true, positionNs: fresh ? settled : state.positionNs });
  }

  pause(): void {
    if (this.snapshot().playing) this.set({ ...this.snapshot(), playing: false });
  }

  setSpeed(speed: number): void {
    this.set({ ...this.snapshot(), speed });
  }

  tick(elapsedMs: number): void {
    const state = this.snapshot();
    const bounds = this.index.bounds();
    if (!state.playing || bounds === null) return;
    const positionNs = advance(state.positionNs, elapsedMs, state.speed, bounds.endNs);
    this.set({ ...state, positionNs, playing: positionNs < bounds.endNs });
  }

  select(entity: string | null): void {
    this.set({ ...this.snapshot(), selected: entity, tab: entity === null ? this.snapshot().tab : "inspect" });
  }

  reveal(positionNs: number, entity: string): void {
    this.set({ ...this.snapshot(), positionNs, selected: entity, picked: { atNs: positionNs, entity } });
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
    if (phase === "ready") this.refreshIfLive(worldId);
    const state = this.snapshot();
    const settled = this.index.settled();
    const before = (this.index.bounds()?.startNs ?? 0) > state.positionNs;
    if (settled !== null && state.worldId === worldId && before) this.seek(settled);
  }

  private refreshIfLive(worldId: string): void {
    if (this.knownIncomplete) return;
    const state = this.snapshot();
    if (this.refreshed || state.worldId !== worldId || state.report.phase !== "ready" || state.report.value.ended)
      return;
    if (state.playback.phase !== "ready") return;
    this.refreshed = true;
    this.set({ ...state, verified: { phase: "loading" } });
    void this.loadReport(worldId);
    void this.loadVerify(worldId);
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
    this.refreshIfLive(worldId);
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

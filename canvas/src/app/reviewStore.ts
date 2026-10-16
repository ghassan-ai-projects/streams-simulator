import type { RunEntry } from "../model/bridge";
import { PlaybackIndex, advance, parseJournalRecord } from "../model/playback";
import type { Badge, Replay } from "../model/report";
import { badgeOf, parseReplay, parseReport } from "../model/report";
import { domainSpec } from "./reviewDomain";
import { Observable } from "./observable";
import type { Bridge } from "./ports";
import type { Load, ReviewState } from "./reviewState";
import { absentLoads, idle, knownUnfinished } from "./reviewState";

export type { ReviewState } from "./reviewState";
export class ReviewStore extends Observable<ReviewState> {
  index = new PlaybackIndex();
  private readonly badges = new Map<string, Badge>();
  private stop: (() => void) | null = null;
  private session = 0;
  private pending = 0;
  private refreshed = false;
  private deferredDomain = false;
  private deferredVerify = false;

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
    const session = (this.session += 1);
    this.refreshed = known?.ended === true;
    this.deferredVerify = false;
    this.deferredDomain = knownUnfinished(known);
    this.pending = 0;
    this.index = new PlaybackIndex();
    this.set({ ...idle, worldId, openedAt: this.now(), ...absentLoads(known) });
    void this.loadReport(worldId, session);
    if (!knownUnfinished(known)) {
      void this.loadVerify(worldId, session);
      void this.loadRunDomain(worldId, session);
    }
    this.stop = this.bridge.followRun(worldId, {
      record: (data) => {
        this.take(worldId, session, data);
      },
      closed: () => {
        this.finish(worldId, session, "ready");
      },
      failed: () => {
        this.finish(worldId, session, "failed");
      },
    });
  }

  refreshCompleted(worldId: string, known: RunEntry | undefined): void {
    const state = this.snapshot();
    if (
      this.stop === null ||
      known?.ended !== true ||
      state.worldId !== worldId ||
      state.report.phase !== "ready" ||
      state.report.value.ended
    )
      return;
    const session = this.session;
    this.set({ ...state, report: { phase: "loading" } });
    void this.loadReport(worldId, session);
  }

  close(): void {
    this.session += 1;
    this.stop?.();
    this.stop = null;
  }

  private live(session: number): boolean {
    return session === this.session;
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

  private take(worldId: string, session: number, data: string): void {
    if (!this.live(session)) return;
    const entry = parseJournalRecord(data);
    if (entry.kind === "observation") this.index.add(entry.observation);
    if (entry.kind === "opened") this.update(worldId, session, { adapter: entry.adapter });
    this.pending += 1;
    if (this.pending === 1) {
      this.schedule(() => {
        if (!this.live(session)) return;
        this.pending = 0;
        this.update(worldId, session, {});
      });
    }
  }

  private finish(worldId: string, session: number, phase: "ready" | "failed"): void {
    if (!this.live(session)) return;
    this.update(worldId, session, { phase });
    if (phase === "ready") this.refreshIfLive(worldId, session);
    const state = this.snapshot();
    const settled = this.index.settled();
    const before = (this.index.bounds()?.startNs ?? 0) > state.positionNs;
    if (settled !== null && state.worldId === worldId && before) this.seek(settled);
  }

  private refreshIfLive(worldId: string, session: number): void {
    const state = this.snapshot();
    if (this.refreshed || state.worldId !== worldId || state.report.phase !== "ready" || state.report.value.ended)
      return;
    if (state.playback.phase !== "ready") return;
    this.refreshed = true;
    this.deferredDomain = true;
    this.deferredVerify = true;
    void this.loadReport(worldId, session);
  }

  private update(worldId: string, session: number, playback: Partial<ReviewState["playback"]>): void {
    const state = this.snapshot();
    if (state.worldId !== worldId || !this.live(session)) return;
    this.set({ ...state, playback: { ...state.playback, records: this.index.size, ...playback } });
  }

  private async loadReport(worldId: string, session: number): Promise<void> {
    const outcome = await this.bridge.action("run.report", { run: worldId });
    if (this.snapshot().worldId !== worldId || !this.live(session)) return;
    this.set({
      ...this.snapshot(),
      report: outcome.ok
        ? { phase: "ready", value: parseReport(outcome.value) }
        : { phase: "failed", failure: outcome.failure },
    });
    this.resumeFinalLoads(worldId, session);
    this.refreshIfLive(worldId, session);
  }

  private resumeFinalLoads(worldId: string, session: number): void {
    const state = this.snapshot();
    if (state.report.phase !== "ready" || !state.report.value.ended) return;
    if (this.deferredDomain) {
      this.deferredDomain = false;
      this.set({ ...state, runDomain: { phase: "loading" } });
      void this.loadRunDomain(worldId, session);
    }
    if (
      state.report.value.complete &&
      (this.deferredVerify || (state.verified.phase === "failed" && state.verified.failure.code === "run_incomplete"))
    ) {
      this.deferredVerify = false;
      this.set({ ...this.snapshot(), verified: { phase: "loading" } });
      void this.loadVerify(worldId, session);
    }
  }

  private async loadRunDomain(worldId: string, session: number): Promise<void> {
    const file = await this.bridge.readFile(`runs/${worldId}/run.json`);
    if (this.snapshot().worldId !== worldId || !this.live(session)) return;
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

  private async loadVerify(worldId: string, session: number): Promise<void> {
    const outcome = await this.bridge.action("run.verify", { run: worldId });
    if (this.snapshot().worldId !== worldId || !this.live(session)) return;
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

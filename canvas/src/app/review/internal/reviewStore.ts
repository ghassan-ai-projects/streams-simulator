import type { RunEntry } from "../../../model/workspace";
import { PlaybackIndex, advance } from "../../../model/review";
import type { JournalEntry } from "../../../model/review";
import type { Badge } from "../../../model/review";
import { badgeOf } from "../../../model/review";
import { Observable, createOperationIdentity } from "../../shared";
import type { OperationIdentity, OperationId } from "../../shared";
import type { Clock, Outcome, ReviewFailure, ReviewGateway, Scheduler } from "../../ports";
import type { Load, ReviewState } from "./reviewState";
import { absentLoads, idle, knownUnfinished, toLoad } from "./reviewState";

export type { ReviewState } from "./reviewState";
export class ReviewStore extends Observable<ReviewState> {
  index = new PlaybackIndex();
  private readonly badges = new Map<string, Badge>();
  private stop: (() => void) | null = null;
  private readonly operations: OperationIdentity;
  private activeSession: OperationId | null = null;
  private pending = 0;
  private refreshed = false;
  private deferredDomain = false;
  private deferredVerify = false;
  private cancelFlush: (() => void) | null = null;

  constructor(
    private readonly gateway: ReviewGateway,
    private readonly clock: Clock,
    private readonly scheduler: Scheduler,
    operations: OperationIdentity = createOperationIdentity(),
  ) {
    super(idle);
    this.operations = operations;
  }

  open(worldId: string, known?: RunEntry): void {
    if (this.snapshot().worldId === worldId && this.activeSession !== null && this.live(this.activeSession)) return;
    this.close();
    const session = this.beginVisit(worldId, known);
    void this.loadReport(worldId, session);
    if (!knownUnfinished(known)) {
      void this.loadVerify(worldId, session);
      void this.loadRunDomain(worldId, session);
    }
    this.stop = this.gateway.follow(worldId, {
      entry: (entry) => {
        this.take(session, entry);
      },
      closed: () => {
        this.finish(worldId, session, "ready");
      },
      failed: () => {
        this.finish(worldId, session, "failed");
      },
    });
  }

  private beginVisit(worldId: string, known: RunEntry | undefined): OperationId {
    const session = this.operations.begin();
    this.activeSession = session;
    this.refreshed = known?.ended === true;
    this.deferredVerify = false;
    this.deferredDomain = knownUnfinished(known);
    this.pending = 0;
    this.index = new PlaybackIndex();
    this.set({ ...idle, worldId, openedAt: this.clock.nowMs(), ...absentLoads(known) });
    return session;
  }

  refreshCompleted(worldId: string, known: RunEntry | undefined): void {
    const state = this.snapshot();
    if (
      known?.ended !== true ||
      state.worldId !== worldId ||
      state.report.phase !== "ready" ||
      state.report.value.ended
    )
      return;
    const session = this.activeSession;
    if (session === null || !this.live(session)) return;
    this.set({ ...state, report: { phase: "loading" } });
    void this.loadReport(worldId, session);
  }

  close(): void {
    this.operations.invalidate();
    this.activeSession = null;
    this.stop?.();
    this.stop = null;
    this.cancelFlush?.();
    this.cancelFlush = null;
  }

  private live(session: OperationId): boolean {
    return this.operations.isCurrent(session);
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

  private take(session: OperationId, entry: JournalEntry): void {
    if (!this.live(session)) return;
    if (entry.kind === "observation") this.index.add(entry.observation);
    if (entry.kind === "opened") this.update(session, { adapter: entry.adapter });
    this.pending += 1;
    if (this.pending === 1) {
      this.cancelFlush = this.scheduler.schedule(250, () => {
        if (!this.live(session)) return;
        this.cancelFlush = null;
        this.pending = 0;
        this.update(session, {});
      });
    }
  }

  private finish(worldId: string, session: OperationId, phase: "ready" | "failed"): void {
    if (!this.live(session)) return;
    this.update(session, { phase });
    if (phase === "ready") this.refreshIfLive(worldId, session);
    const state = this.snapshot();
    const settled = this.index.settled();
    const before = (this.index.bounds()?.startNs ?? 0) > state.positionNs;
    if (settled !== null && before) this.seek(settled);
  }

  private refreshIfLive(worldId: string, session: OperationId): void {
    const state = this.snapshot();
    if (this.refreshed || state.report.phase !== "ready" || state.report.value.ended) return;
    if (state.playback.phase !== "ready") return;
    this.refreshed = true;
    this.deferredDomain = true;
    this.deferredVerify = true;
    void this.loadReport(worldId, session);
  }

  private update(session: OperationId, playback: Partial<ReviewState["playback"]>): void {
    if (!this.live(session)) return;
    const state = this.snapshot();
    this.set({ ...state, playback: { ...state.playback, records: this.index.size, ...playback } });
  }

  private async loadReport(worldId: string, session: OperationId): Promise<void> {
    const applied = await this.loadInto(session, this.gateway.report(worldId), (state, report) => ({
      ...state,
      report,
    }));
    if (!applied) return;
    this.resumeFinalLoads(worldId, session);
    this.refreshIfLive(worldId, session);
  }

  private resumeFinalLoads(worldId: string, session: OperationId): void {
    const state = this.snapshot();
    if (state.report.phase !== "ready" || !state.report.value.ended) return;
    if (this.deferredDomain) {
      this.deferredDomain = false;
      this.set({ ...state, runDomain: { phase: "loading" } });
      void this.loadRunDomain(worldId, session);
    }
    if (
      state.report.value.complete &&
      (this.deferredVerify || (state.verified.phase === "failed" && state.verified.failure.kind === "run-incomplete"))
    ) {
      this.deferredVerify = false;
      this.set({ ...this.snapshot(), verified: { phase: "loading" } });
      void this.loadVerify(worldId, session);
    }
  }

  private async loadRunDomain(worldId: string, session: OperationId): Promise<void> {
    await this.loadInto(session, this.gateway.runDomain(worldId), (state, runDomain) => ({ ...state, runDomain }));
  }

  private async loadVerify(worldId: string, session: OperationId): Promise<void> {
    await this.loadInto(session, this.gateway.verify(worldId), (state, verified) => ({ ...state, verified }));
  }

  private async loadInto<T>(
    session: OperationId,
    pending: Promise<Outcome<T, ReviewFailure>>,
    place: (state: ReviewState, load: Load<T>) => ReviewState,
  ): Promise<boolean> {
    const load = toLoad(await pending);
    if (!this.live(session)) return false;
    this.set(place(this.snapshot(), load));
    return true;
  }
}

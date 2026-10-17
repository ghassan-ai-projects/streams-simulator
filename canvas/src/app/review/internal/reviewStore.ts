import type { RunEntry } from "../../../model/workspace";
import { PlaybackIndex, advance } from "../../../model/review";
import type { JournalEntry } from "../../../model/review";
import type { Badge } from "../../../model/review";
import { badgeOf } from "../../../model/review";
import { Observable, createOperationIdentity } from "../../shared";
import type { OperationIdentity, Operation } from "../../shared";
import type { Clock, Outcome, ReviewFailure, ReviewGateway, Scheduler } from "../../ports";
import { FrameLoop } from "./frameLoop";
import type { FinalLoads, Load, ReviewState } from "./reviewState";
import { absentLoads, finalLoadsFor, idle, knownUnfinished, toLoad } from "./reviewState";

export type { ReviewState } from "./reviewState";
export class ReviewStore extends Observable<ReviewState> {
  index = new PlaybackIndex();
  private readonly badges = new Map<string, Badge>();
  private stop: (() => void) | null = null;
  private readonly operations: OperationIdentity;
  private readonly reports = createOperationIdentity();
  private pending = 0;
  private finals: FinalLoads = finalLoadsFor(undefined);
  private cancelFlush: (() => void) | null = null;
  private readonly frames: FrameLoop;

  constructor(
    private readonly gateway: ReviewGateway,
    private readonly clock: Clock,
    private readonly scheduler: Scheduler,
    operations: OperationIdentity = createOperationIdentity(),
  ) {
    super(idle);
    this.operations = operations;
    this.frames = new FrameLoop(clock, scheduler);
  }

  open(worldId: string, known?: RunEntry): void {
    if (this.snapshot().worldId === worldId && this.operations.current() !== null) return;
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

  private beginVisit(worldId: string, known: RunEntry | undefined): Operation {
    const session = this.operations.begin();
    this.finals = finalLoadsFor(known);
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
    const session = this.operations.current();
    if (session === null) return;
    this.set({ ...state, report: { phase: "loading" } });
    void this.loadReport(worldId, session);
  }

  close(): void {
    this.pause();
    this.operations.invalidate();
    this.stop?.();
    this.stop = null;
    this.cancelFlush?.();
    this.cancelFlush = null;
  }

  badgeFor(worldId: string): Badge | null {
    return this.badges.get(worldId) ?? null;
  }

  private recordBadge(): void {
    const { worldId, report, verified } = this.snapshot();
    if (worldId !== null && report.phase === "ready" && verified.phase !== "loading") {
      this.badges.set(worldId, badgeOf(report.value, verified));
    }
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
    this.frames.start((elapsedMs) => {
      this.tick(elapsedMs);
    });
  }

  pause(): void {
    this.frames.stop();
    if (this.snapshot().playing) this.set({ ...this.snapshot(), playing: false });
  }

  setSpeed(speed: number): void {
    this.set({ ...this.snapshot(), speed });
  }

  private tick(elapsedMs: number): void {
    const state = this.snapshot();
    const bounds = this.index.bounds();
    if (!state.playing || bounds === null) return;
    const positionNs = advance(state.positionNs, elapsedMs, state.speed, bounds.endNs);
    if (positionNs >= bounds.endNs) this.frames.stop();
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

  private take(session: Operation, entry: JournalEntry): void {
    if (!session.live()) return;
    if (entry.kind === "observation") this.index.add(entry.observation);
    this.pending += 1;
    if (this.pending === 1) {
      this.cancelFlush = this.scheduler.schedule(250, () => {
        if (!session.live()) return;
        this.cancelFlush = null;
        this.pending = 0;
        this.update(session, {});
      });
    }
  }

  private finish(worldId: string, session: Operation, phase: "ready" | "failed"): void {
    if (!session.live()) return;
    this.update(session, { phase });
    if (phase === "ready") this.refreshIfLive(worldId, session);
    const state = this.snapshot();
    const settled = this.index.settled();
    const before = (this.index.bounds()?.startNs ?? 0) > state.positionNs;
    if (settled !== null && before) this.seek(settled);
  }

  private refreshIfLive(worldId: string, session: Operation): void {
    const state = this.snapshot();
    if (this.finals.refreshed || state.report.phase !== "ready" || state.report.value.ended) return;
    if (state.playback.phase !== "ready") return;
    this.finals = { refreshed: true, deferred: new Set(["domain", "verify"]) };
    void this.loadReport(worldId, session);
  }

  private update(session: Operation, playback: Partial<ReviewState["playback"]>): void {
    if (!session.live()) return;
    const state = this.snapshot();
    this.set({ ...state, playback: { ...state.playback, records: this.index.size, ...playback } });
  }

  private async loadReport(worldId: string, session: Operation): Promise<void> {
    const request = this.reports.begin();
    const applied = await this.loadInto(
      () => session.live() && request.live(),
      this.gateway.report(worldId),
      (state, report) => ({ ...state, report }),
    );
    if (!applied) return;
    this.recordBadge();
    this.resumeFinalLoads(worldId, session);
    this.refreshIfLive(worldId, session);
  }

  private resumeFinalLoads(worldId: string, session: Operation): void {
    const state = this.snapshot();
    if (state.report.phase !== "ready" || !state.report.value.ended) return;
    if (this.finals.deferred.delete("domain")) {
      this.set({ ...state, runDomain: { phase: "loading" } });
      void this.loadRunDomain(worldId, session);
    }
    if (
      state.report.value.complete &&
      (this.finals.deferred.delete("verify") ||
        (state.verified.phase === "failed" && state.verified.failure.kind === "run-incomplete"))
    ) {
      this.set({ ...this.snapshot(), verified: { phase: "loading" } });
      void this.loadVerify(worldId, session);
    }
  }

  private async loadRunDomain(worldId: string, session: Operation): Promise<void> {
    await this.loadInto(
      () => session.live(),
      this.gateway.runDomain(worldId),
      (state, runDomain) => ({ ...state, runDomain }),
    );
  }

  private async loadVerify(worldId: string, session: Operation): Promise<void> {
    const applied = await this.loadInto(
      () => session.live(),
      this.gateway.verify(worldId),
      (state, verified) => ({ ...state, verified }),
    );
    if (applied) this.recordBadge();
  }

  private async loadInto<T>(
    current: () => boolean,
    pending: Promise<Outcome<T, ReviewFailure>>,
    place: (state: ReviewState, load: Load<T>) => ReviewState,
  ): Promise<boolean> {
    const load = toLoad(await pending);
    if (!current()) return false;
    this.set(place(this.snapshot(), load));
    return true;
  }
}

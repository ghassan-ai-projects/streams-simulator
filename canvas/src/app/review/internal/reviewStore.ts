import type { RunEntry } from "../../../model/workspace";
import { PlaybackIndex, advance } from "../../../model/review";
import type { Badge, JournalEntry } from "../../../model/review";
import { Observable, createOperationIdentity } from "../../shared";
import type { OperationIdentity, Operation } from "../../shared";
import type { Clock, ReviewGateway, Scheduler } from "../../ports";
import { BadgeBook } from "./badgeBook";
import { FrameLoop } from "./frameLoop";
import type { LiveItem } from "./liveState";
import {
  atPosition,
  drained,
  liveEntry,
  liveItemOf,
  loadStart,
  playbackBounds,
  playStart,
  toLive,
  watchEntry,
} from "./liveState";
import { indexFor, ingest, isSessionEntry } from "./reviewIngest";
import { RunLoads } from "./runLoads";
import type { OpenMode, ReviewState } from "./reviewState";
import { endedSinceReport, idle, watchEnded, withListedConnection, visitState, withSessionEntry } from "./reviewState";

export class ReviewStore extends Observable<ReviewState> {
  index = new PlaybackIndex();
  private readonly badges = new BadgeBook();
  private stop: (() => void) | null = null;
  private readonly operations: OperationIdentity;
  private pending = 0;
  private livePending: LiveItem[] = [];
  private holders = 0;
  private mode: OpenMode = "review";
  private cancelFlush: (() => void) | null = null;
  private readonly frames: FrameLoop;
  private readonly loads: RunLoads;

  constructor(
    private readonly gateway: ReviewGateway,
    private readonly clock: Clock,
    private readonly scheduler: Scheduler,
    operations: OperationIdentity = createOperationIdentity(),
  ) {
    super(idle);
    this.operations = operations;
    this.frames = new FrameLoop(clock, scheduler);
    this.loads = new RunLoads(gateway, this.badges, {
      snapshot: () => this.snapshot(),
      set: (state) => {
        this.set(state);
      },
    });
  }

  open(worldId: string, known?: RunEntry): void {
    this.enter(worldId, known, "review");
  }

  openLive(worldId: string, known?: RunEntry): void {
    this.enter(worldId, liveEntry(worldId, known), "review");
  }

  watch(worldId: string, known?: RunEntry): void {
    this.enter(worldId, watchEntry(worldId, known), "watch");
  }

  private enter(worldId: string, known: RunEntry | undefined, mode: OpenMode): void {
    if (this.snapshot().worldId === worldId && this.operations.current() !== null) {
      this.holders += 1;
      return;
    }
    this.release();
    this.holders = 1;
    this.mode = mode;
    const session = this.beginVisit(worldId, known);
    if (mode === "review") this.loads.all(worldId, session, known);
    this.stop = this.follow(worldId, session, known);
  }

  private follow(worldId: string, session: Operation, known: RunEntry | undefined): () => void {
    return this.gateway.follow(worldId, {
      entry: (entry) => {
        this.take(session, entry);
      },
      caughtUp: () => {
        this.markCaughtUp(worldId, session);
      },
      closed: () => {
        this.finish(worldId, session, "ready");
      },
      failed: () => {
        this.finish(worldId, session, "failed");
        if (this.snapshot().connection === "running" || this.snapshot().watching) {
          this.retryLater(worldId, session, known);
        }
      },
    });
  }

  private retryLater(worldId: string, session: Operation, known: RunEntry | undefined): void {
    this.scheduler.schedule(1000, () => {
      if (!session.live()) return;
      const holders = this.holders;
      this.release();
      this.enter(worldId, known, this.mode);
      this.holders = holders;
    });
  }

  private beginVisit(worldId: string, known: RunEntry | undefined): Operation {
    const session = this.operations.begin();
    this.loads.begin(known);
    this.pending = 0;
    this.livePending = [];
    this.index = indexFor(this.mode);
    this.set(visitState(worldId, known, { openedAt: this.clock.nowMs(), mode: this.mode }));
    return session;
  }

  refreshCompleted(worldId: string, known: RunEntry | undefined): void {
    const listed = withListedConnection(this.snapshot(), known);
    if (listed.worldId === worldId && listed !== this.snapshot()) this.set(listed);
    const state = this.snapshot();
    if (!endedSinceReport(state, worldId, known)) return;
    const session = this.operations.current();
    if (session === null) return;
    this.set({ ...state, report: { phase: "loading" } });
    void this.loads.report(worldId, session);
  }

  close(): void {
    this.holders = Math.max(0, this.holders - 1);
    if (this.holders === 0) this.release();
  }

  private release(): void {
    this.pause();
    this.operations.invalidate();
    this.stop?.();
    this.stop = null;
    this.cancelFlush?.();
    this.cancelFlush = null;
  }

  badgeFor(worldId: string): Badge | null {
    return this.badges.of(worldId);
  }

  seek(positionNs: number): void {
    this.set({ ...atPosition(this.drainLive(), positionNs), following: false });
  }

  backToLive(): void {
    const state = this.drainLive();
    if (state.connection !== "running") return;
    this.pause();
    this.set(toLive(state, this.index));
  }

  private markCaughtUp(worldId: string, session: Operation): void {
    if (!session.live()) return;
    this.set({ ...this.drainLive(), caughtUp: true });
    if (this.snapshot().report.phase === "failed") void this.loads.report(worldId, session);
  }

  play(): void {
    const started = playStart(this.drainLive(), this.index);
    if (started === null) return;
    this.set(started);
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
    const bounds = playbackBounds(state, this.index);
    if (!state.playing || bounds === null) return;
    const positionNs = advance(state.positionNs, elapsedMs, state.speed, bounds.endNs);
    if (positionNs >= bounds.endNs) this.frames.stop();
    this.set({ ...atPosition(state, positionNs), playing: positionNs < bounds.endNs });
  }

  select(entity: string | null): void {
    this.set({ ...this.snapshot(), selected: entity, tab: entity === null ? this.snapshot().tab : "inspect" });
  }

  reveal(positionNs: number, entity: string): void {
    this.set({
      ...atPosition(this.drainLive(), positionNs),
      following: false,
      selected: entity,
      picked: { atNs: positionNs, entity },
    });
  }

  showTab(tab: ReviewState["tab"]): void {
    this.set({ ...this.snapshot(), tab });
  }

  private take(session: Operation, entry: JournalEntry): void {
    if (!session.live()) return;
    const item = liveItemOf(entry);
    if (item !== null) this.livePending.push(item);
    ingest(this.index, entry);
    if (isSessionEntry(entry)) {
      const next = withSessionEntry(this.drainLive(), entry);
      if (next !== this.snapshot()) this.set(next);
    }
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
    if (phase === "ready") this.streamEnded(worldId, session);
    const start = this.mode === "watch" ? null : loadStart(this.snapshot(), this.index);
    if (start !== null) this.seek(start);
  }

  private streamEnded(worldId: string, session: Operation): void {
    this.set(watchEnded(this.snapshot()));
    this.loads.refreshIfLive(worldId, session);
  }

  private update(session: Operation, playback: Partial<ReviewState["playback"]>): void {
    if (!session.live()) return;
    const state = this.drainLive();
    this.set({ ...state, playback: { ...state.playback, records: this.index.size, ...playback } });
  }

  private drainLive(): ReviewState {
    const batch = this.livePending;
    this.livePending = [];
    return drained(this.snapshot(), batch, this.clock.nowMs(), this.index);
  }
}

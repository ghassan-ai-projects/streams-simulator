import type { SessionCheck, SessionDraft } from "../../../model/launch";
import { sessionPlanText } from "../../../model/launch";
import { Observable } from "../../shared";
import type { ConnectFailure, ConnectGateway, ConnectStatus, Outcome, Scheduler } from "../../ports";
import type { ConnectState, Live } from "./connectState";
import { isBusy, refusedState, stateOf } from "./connectState";

const pollMs = 2000;

export class ConnectStore extends Observable<ConnectState> {
  private unfollow: (() => void) | null = null;
  private unpoll: (() => void) | null = null;

  constructor(
    private readonly gateway: ConnectGateway,
    private readonly scheduler: Scheduler,
  ) {
    super({ phase: "idle" });
  }

  async check(domain: string, draft: SessionDraft): Promise<SessionCheck | null> {
    if (isBusy(this.snapshot())) return null;
    this.set({ phase: "checking", domain });
    const outcome = await this.gateway.checkSession(sessionPlanText(domain, draft));
    this.set(outcome.ok ? { phase: "checked", domain, check: outcome.value } : refusedState(domain, outcome.failure));
    return outcome.ok ? outcome.value : null;
  }

  async start(domain: string, draft: SessionDraft): Promise<void> {
    const check = await this.check(domain, draft);
    if (check === null) return;
    this.set({ phase: "starting", domain, check });
    this.follow(domain, check, await this.gateway.startSession(sessionPlanText(domain, draft)));
  }

  async resume(domain: string): Promise<void> {
    if (this.snapshot().phase !== "idle") return;
    const outcome = await this.gateway.sessionStatus();
    if (outcome.ok && outcome.value.state === "running" && this.snapshot().phase === "idle") {
      this.follow(domain, null, outcome);
    }
  }

  async stop(): Promise<void> {
    const state = this.snapshot();
    if (state.phase !== "live" || state.stopping) return;
    this.set({ ...state, stopping: true });
    this.settle(state, await this.gateway.stopSession());
  }

  reset(): void {
    if (isBusy(this.snapshot())) return;
    this.set({ phase: "idle" });
  }

  private follow(domain: string, check: SessionCheck | null, outcome: Outcome<ConnectStatus, ConnectFailure>): void {
    if (!outcome.ok) {
      this.set(refusedState(domain, outcome.failure));
      return;
    }
    const next = stateOf(domain, check, outcome.value);
    this.set(next);
    if (next.phase !== "live") return;
    this.unfollow = this.gateway.followSession(next.run, {
      entry: (entry) => {
        const live = this.snapshot();
        if (entry.kind === "binding" && live.phase === "live") this.set({ ...live, binding: entry.binding });
      },
      closed: () => undefined,
      failed: () => undefined,
    });
    this.poll();
  }

  private poll(): void {
    this.unpoll = this.scheduler.schedule(pollMs, () => {
      void this.checkAlive();
    });
  }

  private async checkAlive(): Promise<void> {
    const state = this.snapshot();
    if (state.phase !== "live" || state.stopping) return;
    const outcome = await this.gateway.sessionStatus();
    const now = this.snapshot();
    if (now.phase !== "live" || now.stopping) return;
    if (outcome.ok && outcome.value.state !== "running") this.settle(now, outcome);
    else this.poll();
  }

  private settle(live: Live, outcome: Outcome<ConnectStatus, ConnectFailure>): void {
    this.quiet();
    this.set(outcome.ok ? stateOf(live.domain, live.check, outcome.value) : refusedState(live.domain, outcome.failure));
  }

  private quiet(): void {
    this.unfollow?.();
    this.unpoll?.();
    this.unfollow = null;
    this.unpoll = null;
  }
}

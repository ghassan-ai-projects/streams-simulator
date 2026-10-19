import type { SessionCheck, SessionDraft } from "../../../model/launch";
import { filledFromDefault, firstSessionDraft, sessionPlanText, withPolicyListed } from "../../../model/launch";
import { Observable } from "../../shared";
import type { ConnectFailure, ConnectGateway, ConnectStatus, Outcome, Scheduler } from "../../ports";
import type { ConnectState, LiveSession } from "./connectState";
import { isBusy, markedStale, refusedState, stateOf } from "./connectState";
import { PolicyDefault } from "./policyDefault";

const pollMs = 2000;
const lastingStates: readonly ConnectStatus["state"][] = ["starting", "running", "stopping"];

export class ConnectStore extends Observable<ConnectState> {
  private unfollow: (() => void) | null = null;
  private unpoll: (() => void) | null = null;
  private readonly drafts = new Map<string, SessionDraft>();
  readonly defaults: PolicyDefault;

  constructor(
    private readonly gateway: ConnectGateway,
    private readonly scheduler: Scheduler,
  ) {
    super({ phase: "idle" });
    this.defaults = new PolicyDefault(gateway);
  }

  draftFor(domain: string): SessionDraft {
    return this.drafts.get(domain) ?? firstSessionDraft;
  }

  keepDraft(domain: string, draft: SessionDraft): void {
    this.drafts.set(domain, draft);
    const state = this.snapshot();
    const next = markedStale(state, domain);
    if (next !== state) this.set(next);
  }

  async prefill(domain: string): Promise<SessionDraft | null> {
    const policies = await this.defaults.load();
    if (this.drafts.has(domain) || policies === null || policies.length === 0) return null;
    const draft = filledFromDefault(firstSessionDraft, policies);
    this.drafts.set(domain, draft);
    return draft;
  }

  addPolicy(domain: string, digest: string): SessionDraft {
    const draft = withPolicyListed(this.draftFor(domain), digest);
    this.keepDraft(domain, draft);
    return draft;
  }

  async saveDefault(domain: string): Promise<void> {
    const state = this.snapshot();
    if (state.phase !== "checked" || state.stale || state.domain !== domain) return;
    await this.defaults.save(state.check.policies);
  }

  async check(domain: string, draft: SessionDraft): Promise<SessionCheck | null> {
    if (isBusy(this.snapshot())) return null;
    this.set({ phase: "checking", domain });
    const outcome = await this.gateway.checkSession(sessionPlanText(domain, draft));
    this.set(
      outcome.ok
        ? { phase: "checked", domain, check: outcome.value, stale: false }
        : refusedState(domain, outcome.failure),
    );
    return outcome.ok ? outcome.value : null;
  }

  async start(domain: string, draft: SessionDraft): Promise<void> {
    const check = await this.check(domain, draft);
    if (check === null) return;
    this.set({ phase: "starting", domain, check });
    this.take(domain, check, await this.gateway.startSession(sessionPlanText(domain, draft)));
  }

  async resume(): Promise<void> {
    if (this.snapshot().phase !== "idle") return;
    const outcome = await this.gateway.sessionStatus();
    if (!outcome.ok || !lastingStates.includes(outcome.value.state) || this.snapshot().phase !== "idle") return;
    this.take("", null, outcome);
  }

  async stop(): Promise<void> {
    const state = this.snapshot();
    if (state.phase !== "live" || state.stopping) return;
    this.set({ ...state, stopping: true, note: null });
    this.settle(await this.gateway.stopSession());
  }

  reset(): void {
    if (isBusy(this.snapshot())) return;
    this.set({ phase: "idle" });
  }

  private take(domain: string, check: SessionCheck | null, outcome: Outcome<ConnectStatus, ConnectFailure>): void {
    if (!outcome.ok) {
      this.set(refusedState(domain, outcome.failure));
      return;
    }
    const next = stateOf(domain, check, outcome.value);
    this.set(next);
    if (next.phase !== "live" && next.phase !== "starting") return;
    if (next.phase === "live") this.followJournal(next.run);
    this.poll();
  }

  private followJournal(run: string): void {
    this.unfollow = this.gateway.followSession(run, {
      entry: (entry) => {
        const live = this.snapshot();
        if (entry.kind !== "binding" || live.phase !== "live") return;
        this.set({ ...live, binding: entry.binding });
        this.stopFollowing();
      },
      closed: () => {
        this.lostJournal("The run journal stopped streaming; following it again.");
      },
      failed: () => {
        this.lostJournal("The run journal could not be followed; trying again.");
      },
    });
  }

  private stopFollowing(): void {
    this.unfollow?.();
    this.unfollow = null;
  }

  private lostJournal(note: string): void {
    this.stopFollowing();
    const live = this.snapshot();
    if (live.phase === "live") this.set({ ...live, note });
  }

  private poll(): void {
    this.unpoll = this.scheduler.schedule(pollMs, () => {
      this.unpoll = null;
      void this.checkAlive();
    });
  }

  private async checkAlive(): Promise<void> {
    if (!watched(this.snapshot())) return;
    const outcome = await this.gateway.sessionStatus();
    const now = this.snapshot();
    if (!watched(now) || this.unpoll !== null) return;
    if (outcome.ok && !lastingStates.includes(outcome.value.state)) this.settle(outcome);
    else this.carryOn(now, outcome);
  }

  private carryOn(now: Watched, outcome: Outcome<ConnectStatus, ConnectFailure>): void {
    if (outcome.ok && now.phase === "starting") {
      this.take(now.domain, now.check, outcome);
      return;
    }
    if (outcome.ok && now.phase === "live") this.refreshLive(now, outcome.value);
    this.poll();
    const live = this.snapshot();
    if (live.phase === "live" && live.binding === null && this.unfollow === null) this.followJournal(live.run);
  }

  private refreshLive(now: LiveSession, status: ConnectStatus): void {
    if (status.state !== "running" && status.state !== "stopping") return;
    const stopping = now.stopping || status.state === "stopping";
    if (stopping !== now.stopping || status.controls !== now.controls) {
      this.set({ ...now, stopping, controls: status.controls });
    }
  }

  private settle(outcome: Outcome<ConnectStatus, ConnectFailure>): void {
    const state = this.snapshot();
    if (!watched(state)) return;
    if (!outcome.ok) {
      if (state.phase === "live") this.set(unstopped(state, outcome.failure));
      this.keepPolling();
      return;
    }
    const next = stateOf(state.domain, state.check, outcome.value);
    if (next.phase === "ended" || next.phase === "failed") {
      this.quiet();
      this.set(next);
      return;
    }
    this.set(
      next.phase === "live" && state.phase === "live" ? { ...next, binding: state.binding, note: state.note } : next,
    );
    this.keepPolling();
  }

  private keepPolling(): void {
    if (this.unpoll === null) this.poll();
  }

  private quiet(): void {
    this.unfollow?.();
    this.unpoll?.();
    this.unfollow = null;
    this.unpoll = null;
  }
}

type Watched = LiveSession | Extract<ConnectState, { phase: "starting" }>;

function watched(state: ConnectState): state is Watched {
  return state.phase === "live" || state.phase === "starting";
}

function unstopped(state: LiveSession, failure: ConnectFailure): LiveSession {
  const note = `The session could not be stopped: ${failure.diagnostic.message}. It is still running; try Stop again.`;
  return { ...state, stopping: false, note };
}

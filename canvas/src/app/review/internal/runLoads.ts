import { createOperationIdentity } from "../../shared";
import type { Operation } from "../../shared";
import type { Outcome, ReviewFailure, ReviewGateway } from "../../ports";
import type { RunEntry } from "../../../model/workspace";
import type { BadgeBook } from "./badgeBook";
import type { FinalLoads, Load, ReviewState } from "./reviewState";
import { finalLoadsFor, knownUnfinished, toLoad } from "./reviewState";

interface LoadHost {
  snapshot(): ReviewState;
  set(state: ReviewState): void;
}

export class RunLoads {
  private finals: FinalLoads = finalLoadsFor(undefined);
  private readonly reports = createOperationIdentity();

  constructor(
    private readonly gateway: ReviewGateway,
    private readonly badges: BadgeBook,
    private readonly host: LoadHost,
  ) {}

  begin(known: RunEntry | undefined): void {
    this.finals = finalLoadsFor(known);
  }

  all(worldId: string, session: Operation, known: RunEntry | undefined): void {
    void this.report(worldId, session);
    if (knownUnfinished(known)) return;
    void this.verify(worldId, session);
    void this.runDomain(worldId, session);
  }

  async report(worldId: string, session: Operation): Promise<void> {
    const request = this.reports.begin();
    const applied = await this.loadInto(
      () => session.live() && request.live(),
      this.gateway.report(worldId),
      (state, report) => ({ ...state, report }),
    );
    if (!applied) return;
    this.badges.record(this.host.snapshot());
    this.resumeFinal(worldId, session);
    this.refreshIfLive(worldId, session);
  }

  async checkReplay(worldId: string, session: Operation): Promise<void> {
    const outcome = await this.gateway.report(worldId, { replay: true });
    if (!session.live()) return;
    const state = this.host.snapshot();
    if (!outcome.ok) {
      this.host.set({ ...state, replayCheck: { phase: "failed", message: outcome.failure.diagnostic.message } });
      return;
    }
    this.host.set({ ...state, replayCheck: { phase: "idle" }, report: { phase: "ready", value: outcome.value } });
  }

  refreshIfLive(worldId: string, session: Operation): void {
    const state = this.host.snapshot();
    if (this.finals.refreshed || state.report.phase !== "ready" || state.report.value.ended) return;
    if (state.playback.phase !== "ready") return;
    this.finals = { refreshed: true, deferred: new Set(["domain", "verify"]) };
    void this.report(worldId, session);
  }

  private resumeFinal(worldId: string, session: Operation): void {
    const state = this.host.snapshot();
    if (state.connection === "running" || state.report.phase !== "ready" || !state.report.value.ended) return;
    if (this.finals.deferred.delete("domain")) {
      this.host.set({ ...state, runDomain: { phase: "loading" } });
      void this.runDomain(worldId, session);
    }
    if (
      state.report.value.complete &&
      (this.finals.deferred.delete("verify") ||
        (state.verified.phase === "failed" && state.verified.failure.kind === "run-incomplete"))
    ) {
      this.host.set({ ...this.host.snapshot(), verified: { phase: "loading" } });
      void this.verify(worldId, session);
    }
  }

  private async runDomain(worldId: string, session: Operation): Promise<void> {
    await this.loadInto(
      () => session.live(),
      this.gateway.runDomain(worldId),
      (state, runDomain) => ({ ...state, runDomain }),
    );
  }

  private async verify(worldId: string, session: Operation): Promise<void> {
    const applied = await this.loadInto(
      () => session.live(),
      this.gateway.verify(worldId),
      (state, verified) => ({ ...state, verified }),
    );
    if (applied) this.badges.record(this.host.snapshot());
  }

  private async loadInto<T>(
    current: () => boolean,
    pending: Promise<Outcome<T, ReviewFailure>>,
    place: (state: ReviewState, load: Load<T>) => ReviewState,
  ): Promise<boolean> {
    const load = toLoad(await pending);
    if (!current()) return false;
    this.host.set(place(this.host.snapshot(), load));
    return true;
  }
}

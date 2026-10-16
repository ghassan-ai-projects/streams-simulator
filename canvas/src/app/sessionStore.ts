import { normalisePairingCode, pairingCodeComplete } from "../model/bridge";
import { Observable } from "./observable";
import type { SessionFailure, SessionGateway } from "./ports/session";

export type SessionState =
  | { readonly phase: "checking" }
  | { readonly phase: "unreachable"; readonly failure: SessionFailure }
  | {
      readonly phase: "unpaired";
      readonly code: string;
      readonly busy: boolean;
      readonly failure: SessionFailure | null;
    }
  | { readonly phase: "paired" };

export function pairingBlocked(failure: SessionFailure | null): boolean {
  return failure?.kind === "pairing-closed";
}

export class SessionStore extends Observable<SessionState> {
  constructor(private readonly gateway: SessionGateway) {
    super({ phase: "checking" });
  }

  async check(): Promise<void> {
    const outcome = await this.gateway.check();
    if (!outcome.ok) {
      this.set({ phase: "unreachable", failure: outcome.failure });
      return;
    }
    if (outcome.value) {
      this.set({ phase: "paired" });
    } else if (this.snapshot().phase !== "unpaired") {
      this.set({ phase: "unpaired", code: "", busy: false, failure: null });
    }
  }

  type(raw: string): void {
    const state = this.snapshot();
    if (state.phase !== "unpaired" || pairingBlocked(state.failure)) {
      return;
    }
    this.set({ ...state, code: normalisePairingCode(raw), failure: null });
  }

  canPair(): boolean {
    const state = this.snapshot();
    return (
      state.phase === "unpaired" && !state.busy && !pairingBlocked(state.failure) && pairingCodeComplete(state.code)
    );
  }

  async pair(): Promise<void> {
    const state = this.snapshot();
    if (state.phase !== "unpaired" || !this.canPair()) {
      return;
    }
    this.set({ ...state, busy: true });
    const outcome = await this.gateway.pair(state.code);
    if (outcome.ok) {
      this.set({ phase: "paired" });
      return;
    }
    this.set({
      phase: "unpaired",
      code: pairingBlocked(outcome.failure) ? "" : state.code,
      busy: false,
      failure: outcome.failure,
    });
  }
}

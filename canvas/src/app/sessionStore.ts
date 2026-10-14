import type { BridgeFailure } from "../model/bridge";
import { normalisePairingCode, pairingBlocked, pairingCodeComplete } from "../model/bridge";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

export type SessionState =
  | { readonly phase: "checking" }
  | { readonly phase: "unreachable"; readonly failure: BridgeFailure }
  | {
      readonly phase: "unpaired";
      readonly code: string;
      readonly busy: boolean;
      readonly failure: BridgeFailure | null;
    }
  | { readonly phase: "paired" };

export class SessionStore extends Observable<SessionState> {
  constructor(private readonly bridge: Bridge) {
    super({ phase: "checking" });
  }

  async check(): Promise<void> {
    const outcome = await this.bridge.session();
    if (!outcome.ok) {
      this.set({ phase: "unreachable", failure: outcome.failure });
      return;
    }
    this.set(outcome.value ? { phase: "paired" } : { phase: "unpaired", code: "", busy: false, failure: null });
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
    const outcome = await this.bridge.pair(state.code);
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

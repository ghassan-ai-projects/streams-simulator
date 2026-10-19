import type {
  ControlReply,
  ControlRequest,
  ControlResult,
  ControlScope,
  SeenState,
  StateCheck,
} from "../../../model/review";
import type { ControlFailure, Outcome, ReviewGateway } from "../../ports";
import type { Store } from "../../shared";
import { Observable } from "../../shared";

interface ControlOutcome {
  readonly request: ControlRequest;
  readonly result: ControlResult;
  readonly check: StateCheck;
}

interface ControlState {
  readonly sending: ControlRequest | null;
  readonly last: ControlOutcome | null;
  readonly seen: SeenState | null;
}

export interface ControlSession extends Store<ControlState> {
  hold(scope: ControlScope): Promise<void>;
  release(scope: ControlScope): Promise<void>;
  approveEach(scope: ControlScope): Promise<void>;
  stopApproving(scope: ControlScope): Promise<void>;
  emergencyStop(scope: ControlScope): Promise<void>;
  approve(commandId: string): Promise<void>;
  deny(commandId: string): Promise<void>;
}

const nothingToRead: readonly ControlFailure["kind"][] = ["no-session", "no-controls"];

export class ControlStore extends Observable<ControlState> implements ControlSession {
  constructor(
    private readonly gateway: Pick<ReviewGateway, "control">,
    private readonly newestKey: () => string,
  ) {
    super({ sending: null, last: null, seen: null });
  }

  hold(scope: ControlScope): Promise<void> {
    return this.send({ op: "hold", scope });
  }

  release(scope: ControlScope): Promise<void> {
    return this.send({ op: "release", scope });
  }

  approveEach(scope: ControlScope): Promise<void> {
    return this.send({ op: "approve_mode", scope, enabled: true });
  }

  stopApproving(scope: ControlScope): Promise<void> {
    return this.send({ op: "approve_mode", scope, enabled: false });
  }

  emergencyStop(scope: ControlScope): Promise<void> {
    return this.send({ op: "estop", scope });
  }

  approve(commandId: string): Promise<void> {
    return this.send({ op: "approve", commandId });
  }

  deny(commandId: string): Promise<void> {
    return this.send({ op: "deny", commandId });
  }

  private async send(request: ControlRequest): Promise<void> {
    const state = this.snapshot();
    if (state.sending !== null) return;
    this.set({ ...state, sending: request });
    this.settle(request, await this.gateway.control(request));
  }

  private settle(request: ControlRequest, outcome: Outcome<ControlReply, ControlFailure>): void {
    const state = this.snapshot();
    if (!outcome.ok) {
      const result: ControlResult = { kind: "failed", failure: outcome.failure.kind };
      this.set({ ...state, sending: null, last: { request, result, check: "none" } });
      if (!nothingToRead.includes(outcome.failure.kind)) void this.recheck(request);
      return;
    }
    const seen = { state: outcome.value.state, seenKey: this.newestKey() };
    const busy = outcome.value.reason === "node_busy";
    const result: ControlResult = { kind: "reply", reply: outcome.value };
    this.set({ sending: null, seen, last: { request, result, check: busy ? "pending" : "none" } });
    if (busy) void this.recheck(request);
  }

  private async recheck(request: ControlRequest): Promise<void> {
    this.markCheck(request, "pending");
    const outcome = await this.gateway.control({ op: "state" });
    if (outcome.ok) this.set({ ...this.snapshot(), seen: { state: outcome.value.state, seenKey: this.newestKey() } });
    this.markCheck(request, outcome.ok ? "read" : "failed");
  }

  private markCheck(request: ControlRequest, check: StateCheck): void {
    const state = this.snapshot();
    if (state.last?.request === request) this.set({ ...state, last: { ...state.last, check } });
  }
}

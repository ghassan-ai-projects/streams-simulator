import type {
  ActionRequest,
  ControlReply,
  ControlResult,
  ControlScope,
  SeenState,
  SessionLive,
  StateCheck,
  Waiting,
} from "../../../model/review";
import { allDevices, newestInterrupts } from "../../../model/review";
import type { ControlFailure, Outcome, ReviewGateway } from "../../ports";
import type { Store } from "../../shared";
import { Observable } from "../../shared";

interface ControlOutcome {
  readonly request: ActionRequest;
  readonly result: ControlResult;
  readonly check: StateCheck;
  readonly order: number;
}

interface Sending {
  readonly request: ActionRequest;
  readonly order: number;
}

export interface ControlState {
  readonly run: string | null;
  readonly sending: readonly Sending[];
  readonly blocked: ActionRequest | null;
  readonly last: ControlOutcome | null;
  readonly seen: SeenState | null;
  readonly picked: ControlScope;
}

type Sent = "sent" | "not-sent";

interface ControlSource {
  readonly run: () => string | null;
  readonly newest: () => SessionLive | null;
}

export interface ControlSession extends Store<ControlState> {
  bind(run: string | null): void;
  pick(scope: ControlScope): void;
  hold(scope: ControlScope): Promise<Sent>;
  release(scope: ControlScope): Promise<Sent>;
  approveEach(scope: ControlScope): Promise<Sent>;
  stopApproving(scope: ControlScope): Promise<Sent>;
  emergencyStop(scope: ControlScope): Promise<Sent>;
  approve(waiting: Waiting): Promise<Sent>;
  deny(waiting: Waiting): Promise<Sent>;
}

interface Flight {
  readonly request: ActionRequest;
  readonly run: string | null;
  readonly order: number;
}

const nothingToRead: readonly ControlFailure["kind"][] = ["no-session", "no-controls", "rejected"];
const remembered = 4;

function fresh(run: string | null): ControlState {
  return { run, sending: [], blocked: null, last: null, seen: null, picked: allDevices };
}

function without(sending: readonly Sending[], order: number): readonly Sending[] {
  return sending.filter((entry) => entry.order !== order);
}

export class ControlStore extends Observable<ControlState> implements ControlSession {
  private ordinal = 0;
  private generation = 0;

  constructor(
    private readonly gateway: Pick<ReviewGateway, "control">,
    private readonly source: ControlSource,
  ) {
    super(fresh(source.run()));
  }

  bind(run: string | null): void {
    if (this.snapshot().run === run) return;
    this.generation += 1;
    this.set(fresh(run));
  }

  pick(scope: ControlScope): void {
    this.set({ ...this.snapshot(), picked: scope });
  }

  hold(scope: ControlScope): Promise<Sent> {
    return this.send({ op: "hold", scope });
  }

  release(scope: ControlScope): Promise<Sent> {
    return this.send({ op: "release", scope });
  }

  approveEach(scope: ControlScope): Promise<Sent> {
    return this.send({ op: "approve_mode", scope, enabled: true });
  }

  stopApproving(scope: ControlScope): Promise<Sent> {
    return this.send({ op: "approve_mode", scope, enabled: false });
  }

  emergencyStop(scope: ControlScope): Promise<Sent> {
    return this.send({ op: "estop", scope });
  }

  approve({ commandId, contentDigest }: Waiting): Promise<Sent> {
    return this.send({ op: "approve", commandId, contentDigest });
  }

  deny({ commandId, contentDigest }: Waiting): Promise<Sent> {
    return this.send({ op: "deny", commandId, contentDigest });
  }

  private async send(request: ActionRequest): Promise<Sent> {
    const state = this.snapshot();
    if (request.op !== "estop" && state.sending.length > 0) {
      this.set({ ...state, blocked: request });
      return "not-sent";
    }
    this.ordinal += 1;
    const flight: Flight = { request, run: state.run, order: this.ordinal };
    this.set({ ...state, blocked: null, sending: [...state.sending, { request, order: this.ordinal }] });
    this.settle(flight, await this.gateway.control(request));
    return "sent";
  }

  private settle(flight: Flight, outcome: Outcome<ControlReply, ControlFailure>): void {
    const state = this.snapshot();
    if (state.run !== flight.run) return;
    const sending = without(state.sending, flight.order);
    const blocked = sending.length === 0 ? null : state.blocked;
    if (!outcome.ok) {
      this.failed({ ...state, sending, blocked }, flight, outcome.failure);
      return;
    }
    const busy = outcome.value.reason === "node_busy";
    const result: ControlResult = { kind: "reply", reply: outcome.value };
    const next = this.adopted({ ...state, sending, blocked }, outcome.value);
    this.set(this.withLast(next, flight, result, busy ? "pending" : "none"));
    if (busy) void this.recheck(flight);
  }

  private failed(state: ControlState, flight: Flight, failure: ControlFailure): void {
    const result: ControlResult = { kind: "failed", failure: failure.kind, detail: failure.diagnostic.message };
    this.set(this.withLast(state, flight, result, "none"));
    if (!nothingToRead.includes(failure.kind)) void this.recheck(flight);
  }

  private withLast(state: ControlState, flight: Flight, result: ControlResult, check: StateCheck): ControlState {
    if (state.last !== null && state.last.order > flight.order) return state;
    return { ...state, last: { request: flight.request, result, check, order: flight.order } };
  }

  private adopted(state: ControlState, reply: ControlReply): ControlState {
    this.generation += 1;
    const shown = newestInterrupts(this.source.newest(), state.seen, state.run);
    const older = [shown, ...(state.seen?.older ?? [])].flatMap((entry) => (entry === null ? [] : [entry]));
    const seen: SeenState = {
      run: state.run ?? "",
      state: reply.state,
      seenKey: this.source.newest()?.key ?? "",
      older: older.slice(0, remembered),
    };
    return { ...state, seen };
  }

  private async recheck(flight: Flight): Promise<void> {
    const generation = this.generation + 1;
    this.generation = generation;
    this.markCheck(flight, "pending");
    const outcome = await this.gateway.control({ op: "state" });
    const state = this.snapshot();
    if (state.run !== flight.run || this.generation !== generation) {
      this.markCheck(flight, "none");
      return;
    }
    if (outcome.ok) this.set({ ...state, seen: this.read(state, outcome.value) });
    this.markCheck(flight, outcome.ok ? "read" : "failed");
  }

  private read(state: ControlState, reply: ControlReply): SeenState {
    const seenKey = this.source.newest()?.key ?? "";
    return { run: state.run ?? "", state: reply.state, seenKey, older: [] };
  }

  private markCheck(flight: Flight, check: StateCheck): void {
    const state = this.snapshot();
    if (state.last?.order === flight.order) this.set({ ...state, last: { ...state.last, check } });
  }
}

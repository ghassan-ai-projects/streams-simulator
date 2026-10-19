import type { TwinReport, TwinRequest } from "../../../model/review";
import type { ReviewGateway, TwinFailure } from "../../ports";
import type { Store } from "../../shared";
import { Observable, createOperationIdentity } from "../../shared";

export type TwinState =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly request: TwinRequest }
  | { readonly phase: "ready"; readonly request: TwinRequest; readonly report: TwinReport }
  | { readonly phase: "failed"; readonly request: TwinRequest; readonly failure: TwinFailure };

export interface TwinSession extends Store<TwinState> {
  compare(request: TwinRequest): Promise<void>;
}

const idle: TwinState = { phase: "idle" };

export function twinFor(state: TwinState, run: string | null): TwinState {
  return state.phase !== "idle" && state.request.run === run ? state : idle;
}

export class TwinStore extends Observable<TwinState> implements TwinSession {
  private readonly operations = createOperationIdentity();

  constructor(private readonly gateway: Pick<ReviewGateway, "twinReport">) {
    super(idle);
  }

  async compare(request: TwinRequest): Promise<void> {
    const state = this.snapshot();
    if (state.phase === "loading" && state.request.run === request.run) return;
    const operation = this.operations.begin();
    this.set({ phase: "loading", request });
    const outcome = await this.gateway.twinReport(request);
    if (!operation.live()) return;
    this.set(
      outcome.ok
        ? { phase: "ready", request, report: outcome.value }
        : { phase: "failed", request, failure: outcome.failure },
    );
  }
}

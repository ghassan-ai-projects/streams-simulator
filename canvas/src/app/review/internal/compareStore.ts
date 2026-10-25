import type { Comparison } from "../../../model/review";
import type { DocumentFailure, ReviewGateway } from "../../ports";
import type { Store } from "../../shared";
import { Observable, createOperationIdentity } from "../../shared";

interface Pairing {
  readonly run: string;
  readonly reference: string;
}

export type CompareState =
  | { readonly phase: "idle" }
  | ({ readonly phase: "loading" } & Pairing)
  | ({ readonly phase: "ready"; readonly comparison: Comparison } & Pairing)
  | ({ readonly phase: "failed"; readonly failure: DocumentFailure } & Pairing);

export interface CompareSession extends Store<CompareState> {
  choose(run: string, reference: string): Promise<void>;
  clear(): void;
}

const idle: CompareState = { phase: "idle" };

export function compareFor(state: CompareState, run: string | null): CompareState {
  return state.phase !== "idle" && state.run === run ? state : idle;
}

export class CompareStore extends Observable<CompareState> implements CompareSession {
  private readonly operations = createOperationIdentity();

  constructor(private readonly gateway: Pick<ReviewGateway, "compare">) {
    super(idle);
  }

  async choose(run: string, reference: string): Promise<void> {
    const operation = this.operations.begin();
    this.set({ phase: "loading", run, reference });
    const outcome = await this.gateway.compare(reference, run);
    if (!operation.live()) return;
    this.set(
      outcome.ok
        ? { phase: "ready", run, reference, comparison: outcome.value }
        : { phase: "failed", run, reference, failure: outcome.failure },
    );
  }

  clear(): void {
    this.operations.invalidate();
    this.set(idle);
  }
}

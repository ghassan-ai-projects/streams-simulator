import type { Outcome, ProvenanceFailure } from "../../ports";
import { Observable, createOperationIdentity } from "../../shared";

export type Fetched<T> =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly key: string }
  | { readonly phase: "ready"; readonly key: string; readonly value: T }
  | { readonly phase: "failed"; readonly key: string; readonly failure: ProvenanceFailure };

const idle = { phase: "idle" } as const;

export class FetchStore<T> extends Observable<Fetched<T>> {
  private readonly operations = createOperationIdentity();

  constructor() {
    super(idle);
  }

  async run(key: string, request: () => Promise<Outcome<T, ProvenanceFailure>>): Promise<void> {
    const operation = this.operations.begin();
    this.set({ phase: "loading", key });
    const outcome = await request();
    if (!operation.live()) return;
    this.set(
      outcome.ok ? { phase: "ready", key, value: outcome.value } : { phase: "failed", key, failure: outcome.failure },
    );
  }

  clear(): void {
    this.operations.invalidate();
    this.set(idle);
  }
}

export function fetchedFor<T>(state: Fetched<T>, key: string): Fetched<T> {
  return state.phase !== "idle" && state.key === key ? state : idle;
}

import type { FileSaver, ReviewGateway } from "../../ports";
import type { Store } from "../../shared";
import { Observable } from "../../shared";

export type HandoverState =
  | { readonly phase: "idle" }
  | { readonly phase: "working"; readonly run: string }
  | { readonly phase: "saved"; readonly run: string; readonly name: string }
  | { readonly phase: "failed"; readonly run: string; readonly message: string };

interface SaveOptions {
  readonly replay: boolean;
  readonly fileStem: string;
}

export interface HandoverSession extends Store<HandoverState> {
  save(run: string, options: SaveOptions): Promise<void>;
}

const idle: HandoverState = { phase: "idle" };

export function handoverFor(state: HandoverState, run: string | null): HandoverState {
  return state.phase !== "idle" && state.run === run ? state : idle;
}

export class HandoverStore extends Observable<HandoverState> implements HandoverSession {
  constructor(
    private readonly gateway: Pick<ReviewGateway, "handover">,
    private readonly saver: FileSaver,
  ) {
    super(idle);
  }

  async save(run: string, options: SaveOptions): Promise<void> {
    if (this.snapshot().phase === "working") return;
    this.set({ phase: "working", run });
    const outcome = await this.gateway.handover(run, { replay: options.replay });
    if (!outcome.ok) {
      this.set({ phase: "failed", run, message: outcome.failure.diagnostic.message });
      return;
    }
    const name = `${options.fileStem}-handover.md`;
    this.set(
      this.saver.save(name, outcome.value)
        ? { phase: "saved", run, name }
        : { phase: "failed", run, message: "The browser could not save the file." },
    );
  }
}

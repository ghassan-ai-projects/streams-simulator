import type { BridgeFailure, Workspace } from "../model/bridge";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

export type WorkspaceState =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: BridgeFailure }
  | { readonly phase: "ready"; readonly workspace: Workspace };

export class WorkspaceStore extends Observable<WorkspaceState> {
  private stopWatching: (() => void) | null = null;

  constructor(private readonly bridge: Bridge) {
    super({ phase: "loading" });
  }

  async open(): Promise<void> {
    await this.reload();
    this.stopWatching ??= this.bridge.watchWorkspace(() => {
      void this.reload();
    });
  }

  close(): void {
    this.stopWatching?.();
    this.stopWatching = null;
  }

  async reload(): Promise<void> {
    const outcome = await this.bridge.workspace();
    this.set(outcome.ok ? { phase: "ready", workspace: outcome.value } : { phase: "failed", failure: outcome.failure });
  }
}

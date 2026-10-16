import type { BridgeFailure, Workspace } from "../model/bridge";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

export type WorkspaceState =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: BridgeFailure }
  | { readonly phase: "ready"; readonly workspace: Workspace };

export class WorkspaceStore extends Observable<WorkspaceState> {
  private stopWatching: (() => void) | null = null;
  private session = 0;
  private request = 0;

  constructor(private readonly bridge: Bridge) {
    super({ phase: "loading" });
  }

  async open(): Promise<void> {
    const session = (this.session += 1);
    await this.reload(session);
    if (session !== this.session) return;
    this.stopWatching?.();
    this.stopWatching = this.bridge.watchWorkspace(() => {
      void this.reload(session);
    });
  }

  close(): void {
    this.session += 1;
    this.stopWatching?.();
    this.stopWatching = null;
  }

  async reload(session = this.session): Promise<void> {
    const request = (this.request += 1);
    const outcome = await this.bridge.workspace();
    if (session !== this.session || request !== this.request) return;
    this.set(outcome.ok ? { phase: "ready", workspace: outcome.value } : { phase: "failed", failure: outcome.failure });
  }
}

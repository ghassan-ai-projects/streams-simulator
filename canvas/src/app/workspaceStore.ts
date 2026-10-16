import type { BridgeFailure, Workspace } from "../model/bridge";
import { Observable } from "./observable";
import type { Bridge } from "./ports";
import { createOperationIdentity } from "./shared/operationIdentity";
import type { OperationIdentity, OperationId } from "./shared/operationIdentity";

export type WorkspaceState =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: BridgeFailure }
  | { readonly phase: "ready"; readonly workspace: Workspace };

export class WorkspaceStore extends Observable<WorkspaceState> {
  private stopWatching: (() => void) | null = null;
  private readonly sessions: OperationIdentity;
  private readonly requests: OperationIdentity;

  constructor(private readonly bridge: Bridge, operations = createOperationIdentity()) {
    super({ phase: "loading" });
    this.sessions = operations;
    this.requests = createOperationIdentity();
  }

  async open(): Promise<void> {
    const session = this.sessions.begin();
    await this.reload(session);
    if (!this.sessions.isCurrent(session)) return;
    this.stopWatching?.();
    this.stopWatching = this.bridge.watchWorkspace(() => {
      void this.reload(session);
    });
  }

  close(): void {
    this.sessions.invalidate();
    this.requests.invalidate();
    this.stopWatching?.();
    this.stopWatching = null;
  }

  async reload(session: OperationId | null = null): Promise<void> {
    const activeSession = session ?? this.sessions.begin();
    const request = this.requests.begin();
    const outcome = await this.bridge.workspace();
    if (!this.sessions.isCurrent(activeSession) || !this.requests.isCurrent(request)) return;
    this.set(outcome.ok ? { phase: "ready", workspace: outcome.value } : { phase: "failed", failure: outcome.failure });
  }
}

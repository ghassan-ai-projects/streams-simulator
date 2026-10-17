import { Observable, createOperationIdentity } from "../../shared";
import type { WorkspaceGateway } from "../../ports";
import type { OperationIdentity, Operation } from "../../shared";
import type { WorkspaceState, WorkspaceStore as WorkspaceStoreApi } from "./api";

export class WorkspaceStore extends Observable<WorkspaceState> {
  private stopWatching: (() => void) | null = null;
  private readonly sessions: OperationIdentity;
  private readonly requests: OperationIdentity;

  constructor(
    private readonly gateway: WorkspaceGateway,
    operations = createOperationIdentity(),
  ) {
    super({ phase: "loading" });
    this.sessions = operations;
    this.requests = createOperationIdentity();
  }

  async open(): Promise<void> {
    const session = this.sessions.begin();
    this.stopWatching?.();
    this.stopWatching = this.gateway.watch(() => {
      void this.reload(session);
    });
    await this.reload(session);
  }

  close(): void {
    this.sessions.invalidate();
    this.requests.invalidate();
    this.stopWatching?.();
    this.stopWatching = null;
  }

  private async reload(session: Operation): Promise<void> {
    const request = this.requests.begin();
    const outcome = await this.gateway.load();
    if (!session.live() || !request.live()) return;
    this.set(outcome.ok ? { phase: "ready", workspace: outcome.value } : { phase: "failed", failure: outcome.failure });
  }
}

export function createWorkspaceStore(gateway: WorkspaceGateway): WorkspaceStoreApi {
  return new WorkspaceStore(gateway);
}

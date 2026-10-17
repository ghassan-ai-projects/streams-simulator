import type { Session } from "../../session";
import type { WorkspaceStore } from "../../workspace";

export function sessionGate(session: Session, workspace: WorkspaceStore): () => void {
  let opened = false;
  const syncWorkspace = () => {
    const paired = session.snapshot().phase === "paired";
    if (paired && !opened) void workspace.open();
    if (!paired && opened) workspace.close();
    opened = paired;
  };
  const recheckUnpaired = () => {
    const listing = workspace.snapshot();
    if (listing.phase === "failed" && listing.failure.kind === "unpaired") void session.check();
  };
  const stops = [session.subscribe(syncWorkspace), workspace.subscribe(recheckUnpaired)];
  recheckUnpaired();
  void session.check();
  syncWorkspace();
  return () => {
    stops.forEach((stop) => {
      stop();
    });
    if (opened) workspace.close();
    opened = false;
  };
}

import type { ConnectState } from "../../launch";
import type { WorkspaceState } from "../../workspace";

function listedRunning(listing: WorkspaceState, worldId: string): boolean {
  if (listing.phase !== "ready") return false;
  return listing.workspace.runs.some((run) => run.worldId === worldId && run.connected === "running");
}

export function runIsLive(connected: ConnectState, listing: WorkspaceState, worldId: string): boolean {
  return (connected.phase === "live" && connected.run === worldId) || listedRunning(listing, worldId);
}

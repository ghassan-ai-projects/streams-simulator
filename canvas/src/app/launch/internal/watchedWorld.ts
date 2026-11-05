import type { ConnectState } from "./connectState";
import type { LaunchState } from "./launchState";

export interface Following {
  readonly kind: "session" | "watch";
  readonly run: string | null;
}

export function watchedWorld(state: LaunchState, plantId: string): string | null {
  if (state.phase !== "running" || state.origin.domain !== plantId) return null;
  return state.worldId;
}

export function followingOf(connected: ConnectState, launched: LaunchState, plantId: string | undefined): Following {
  const session = connected.phase === "live" && connected.domain === plantId ? connected.run : null;
  if (session !== null || plantId === undefined) return { kind: "session", run: session };
  return { kind: "watch", run: watchedWorld(launched, plantId) };
}

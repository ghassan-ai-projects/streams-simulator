import type { LaunchState } from "./launchState";

export function watchedWorld(state: LaunchState, plantId: string): string | null {
  if (state.phase !== "running" || state.origin.domain !== plantId) return null;
  return state.worldId;
}

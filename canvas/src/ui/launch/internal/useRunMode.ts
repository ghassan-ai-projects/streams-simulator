import { useState } from "react";
import type { ConnectSession, ConnectState } from "../../../app/launch";
import { useStore } from "../../primitives";

export type RunKind = "complete" | "session";

const heldPhases: readonly ConnectState["phase"][] = ["starting", "live", "ended", "failed"];

function holdsSession(state: ConnectState, plant: string): boolean {
  return state.phase !== "idle" && heldPhases.includes(state.phase) && state.domain === plant;
}

export function useRunMode(connect: ConnectSession, plant: string) {
  const state = useStore(connect);
  const [chosen, choose] = useState<RunKind>("complete");
  const held = holdsSession(state, plant);
  return { kind: held ? ("session" as const) : chosen, held, choose };
}

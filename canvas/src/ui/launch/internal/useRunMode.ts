import { useState } from "react";
import type { ConnectSession, ConnectState } from "../../../app/launch";
import { useStore } from "../../primitives";

export type RunKind = "complete" | "session";

type Held = Extract<ConnectState, { phase: "starting" | "live" | "ended" | "failed" }>;

const heldPhases: readonly ConnectState["phase"][] = ["starting", "live", "ended", "failed"];

function isHeld(state: ConnectState): state is Held {
  return heldPhases.includes(state.phase);
}

export function useRunMode(connect: ConnectSession, plant: string) {
  const state = useStore(connect);
  const [chosen, choose] = useState<RunKind>("complete");
  const held = isHeld(state) && state.domain === plant;
  return { kind: held ? ("session" as const) : chosen, held, choose };
}

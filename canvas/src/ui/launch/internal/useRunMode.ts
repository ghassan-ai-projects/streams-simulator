import { useState } from "react";
import type { ConnectSession, ConnectState, RunKind } from "../../../app/launch";
import { useStore } from "../../primitives";

export type { RunKind };

type Held = Extract<ConnectState, { phase: "starting" | "live" | "ended" | "failed" }>;

const heldPhases: readonly ConnectState["phase"][] = ["starting", "live", "ended", "failed"];

function isHeld(state: ConnectState): state is Held {
  return heldPhases.includes(state.phase);
}

export function useRunMode(connect: ConnectSession, plant: string) {
  const state = useStore(connect);
  const [chosen, setChosen] = useState<RunKind>(() => connect.kindFor(plant));
  const held = isHeld(state) && state.domain === plant;
  const choose = (kind: RunKind) => {
    connect.chooseKind(plant, kind);
    setChosen(kind);
  };
  return { kind: held ? ("session" as const) : chosen, held, choose };
}

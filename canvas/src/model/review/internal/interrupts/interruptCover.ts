import { sameScope } from "./interruptScope";
import type { ControlScope, InterruptState } from "./interrupts";

interface Cover {
  readonly narrower: readonly ControlScope[];
  readonly wider: ControlScope | null;
}

function boardOfDevice(state: InterruptState, device: string): string | null {
  return state.devices.find((covered) => covered.device === device)?.board ?? null;
}

function contains(outer: ControlScope, inner: ControlScope, state: InterruptState): boolean {
  if (outer.kind === "all") return inner.kind !== "all";
  if (outer.kind === "board") return inner.kind === "device" && boardOfDevice(state, inner.id) === outer.id;
  return false;
}

export function coverAround(scope: ControlScope, scopes: readonly ControlScope[], state: InterruptState): Cover {
  const others = scopes.filter((other) => !sameScope(other, scope));
  return {
    narrower: others.filter((other) => contains(scope, other, state)),
    wider: others.find((other) => contains(other, scope, state)) ?? null,
  };
}

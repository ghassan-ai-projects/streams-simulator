import type { Operation } from "./operationIdentity";

interface Placed {
  readonly phase: string;
  readonly path?: string;
}

export function holdsOperation(operation: Operation, state: Placed, path: string): boolean {
  return operation.live() && state.phase !== "idle" && state.path === path;
}

import type { Acknowledgement } from "../../model/launch";
import type { Diagnostic, Outcome } from "./shared";

export type { Acknowledgement } from "../../model/launch";

export type LaunchFailure =
  | {
      readonly kind: "rejected";
      readonly step: "create" | "advance" | "inject" | "invoke" | "end";
      readonly diagnostic: Diagnostic;
    }
  | { readonly kind: "unavailable" | "unreadable"; readonly diagnostic: Diagnostic };

export interface LaunchGateway {
  createWorld(input: {
    readonly domain: string;
    readonly seed: number;
    readonly explore: boolean;
  }): Promise<Outcome<{ readonly worldId: string }, LaunchFailure>>;
  advance(worldId: string, byNs: number): Promise<Outcome<void, LaunchFailure>>;
  inject(worldId: string, entity: string, fault: string): Promise<Outcome<void, LaunchFailure>>;
  invoke(
    worldId: string,
    entity: string,
    effector: string,
    args: Readonly<Record<string, unknown>>,
  ): Promise<Outcome<Acknowledgement, LaunchFailure>>;
  endRun(worldId: string): Promise<Outcome<void, LaunchFailure>>;
  destroyWorld(worldId: string): Promise<void>;
}

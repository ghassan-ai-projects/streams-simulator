import type { Diagnostic, Outcome } from "./shared";

export type SessionFailure =
  | { readonly kind: "unreachable"; readonly diagnostic: Diagnostic }
  | { readonly kind: "code-rejected"; readonly diagnostic: Diagnostic }
  | {
      readonly kind: "pairing-closed";
      readonly reason: "locked" | "already-used";
      readonly diagnostic: Diagnostic;
    };

export interface SessionGateway {
  check(): Promise<Outcome<boolean, SessionFailure>>;
  pair(code: string): Promise<Outcome<boolean, SessionFailure>>;
}

import type { SessionFailure } from "../ports";

export type { SessionFailure };

export const PAIRING_CODE_LENGTH = 12;

export type SessionState =
  | { readonly phase: "checking" }
  | { readonly phase: "unreachable"; readonly failure: SessionFailure }
  | {
      readonly phase: "unpaired";
      readonly code: string;
      readonly busy: boolean;
      readonly failure: SessionFailure | null;
    }
  | { readonly phase: "paired" };

export interface Session {
  readonly snapshot: () => SessionState;
  readonly subscribe: (listener: () => void) => () => void;
  check(): Promise<void>;
  type(raw: string): void;
  canPair(): boolean;
  pair(): Promise<void>;
}

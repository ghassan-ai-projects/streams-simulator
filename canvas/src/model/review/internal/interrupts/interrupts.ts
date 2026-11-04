type ScopeKind = "all" | "board" | "device";

export interface ControlScope {
  readonly kind: ScopeKind;
  readonly id: string;
}

export type ControlOp = "hold" | "release" | "approve_mode" | "approve" | "deny" | "estop" | "state";

export type ActionRequest =
  | { readonly op: "hold" | "release" | "estop"; readonly scope: ControlScope }
  | { readonly op: "approve_mode"; readonly scope: ControlScope; readonly enabled: boolean }
  | { readonly op: "approve" | "deny"; readonly commandId: string; readonly contentDigest: string };

export type ControlRequest = ActionRequest | { readonly op: "state" };

export type WordedRequest =
  | Exclude<ActionRequest, { readonly op: "approve" | "deny" }>
  | { readonly op: "approve" | "deny"; readonly commandId: string };

export type ControlFailureKind =
  "no-session" | "no-controls" | "gone" | "timeout" | "unavailable" | "unreadable" | "rejected";

export type InterruptMode = "hold" | "approve";
type NodeClock = "world" | "wall";

export interface InterruptHold {
  readonly scope: ControlScope;
  readonly latched: boolean;
}

export interface ArgPair {
  readonly name: string;
  readonly value: string;
}

export interface PendingApproval {
  readonly commandId: string;
  readonly device: string;
  readonly command: string;
  readonly deadlineNs: number;
  readonly queuedAtNs: number;
  readonly clock: NodeClock;
  readonly contentDigest: string;
  readonly args: readonly ArgPair[];
  readonly argsTruncated: boolean;
}

export type Waiting = Pick<PendingApproval, "commandId" | "contentDigest">;

export interface CoveredDevice {
  readonly device: string;
  readonly board: string;
  readonly mode: InterruptMode;
  readonly latched: boolean;
}

export interface InterruptState {
  readonly holds: readonly InterruptHold[];
  readonly approve: readonly ControlScope[];
  readonly pending: readonly PendingApproval[];
  readonly devices: readonly CoveredDevice[];
}

export interface ControlReply {
  readonly result: "applied" | "refused";
  readonly reason: string;
  readonly op: ControlOp | null;
  readonly commandId: string;
  readonly stopped: readonly string[];
  readonly skipped: readonly string[];
  readonly unreachable: readonly string[];
  readonly state: InterruptState;
}

export type InterruptEvent =
  | "inventory"
  | "control"
  | "command"
  | "clock"
  | "flush"
  | "reply"
  | "verdict"
  | "resolve"
  | "stop"
  | "effect"
  | "purge";

export interface InterruptRecord {
  readonly event: InterruptEvent;
  readonly nodeNs: number;
  readonly clock: NodeClock;
  readonly op: ControlOp | null;
  readonly scope: ControlScope | null;
  readonly enabled: boolean | null;
  readonly commandId: string;
  readonly device: string;
  readonly command: string;
  readonly verdict: "pass" | "refuse" | "queue" | null;
  readonly resolution: "forward" | "refuse" | null;
  readonly result: "applied" | "refused" | null;
  readonly reason: string;
}

import type { FieldIssue, KindFile, Selection } from "../../../model/devices";
import type { DeviceFailure, KindChecked } from "../../ports";

type SaveState =
  | { readonly kind: "idle" }
  | { readonly kind: "saving" }
  | { readonly kind: "write-failed"; readonly failure: DeviceFailure };

type CheckState =
  | { readonly kind: "unchecked" }
  | { readonly kind: "checking" }
  | { readonly kind: "valid"; readonly summary: KindChecked }
  | { readonly kind: "refused"; readonly failure: DeviceFailure }
  | { readonly kind: "failed"; readonly failure: DeviceFailure };

export interface Draft {
  readonly phase: "ready";
  readonly path: string;
  readonly saved: KindFile;
  readonly doc: KindFile;
  readonly selected: Selection;
  readonly save: SaveState;
  readonly check: CheckState;
  readonly issues: readonly FieldIssue[];
}

export type DeviceState =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly path: string }
  | { readonly phase: "failed"; readonly path: string; readonly failure: DeviceFailure }
  | Draft;

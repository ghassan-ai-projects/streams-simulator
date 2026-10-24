import type { HeldIssue, Instrument, Member, TwinChoices, TwinIssue } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { DeviceFailure, DomainChecked, LintReport, TwinMade } from "../../ports";

export type Catalogue = Readonly<Record<string, readonly Instrument[] | null>>;

type ChoiceSave =
  | { readonly kind: "idle" }
  | { readonly kind: "saving" }
  | { readonly kind: "write-failed"; readonly failure: DeviceFailure };

type Implemented =
  | { readonly kind: "checking" }
  | { readonly kind: "passed"; readonly checked: DomainChecked }
  | { readonly kind: "failed"; readonly failure: DeviceFailure };

type Plausibility =
  | { readonly kind: "running" }
  | { readonly kind: "done"; readonly report: LintReport }
  | { readonly kind: "unavailable" };

export interface Built {
  readonly kind: "built";
  readonly made: TwinMade | null;
  readonly path: string;
  readonly fingerprint: string;
  readonly implemented: Implemented;
  readonly lint: Plausibility;
}

type Build =
  | { readonly kind: "idle" }
  | { readonly kind: "building" }
  | { readonly kind: "exists" }
  | { readonly kind: "refused"; readonly failure: DeviceFailure }
  | { readonly kind: "failed"; readonly failure: DeviceFailure }
  | Built;

export interface Held {
  readonly path: string;
  readonly name: string;
  readonly words: string;
  readonly issues: readonly HeldIssue[];
  readonly adding: boolean;
}

export interface TwinDraft {
  readonly phase: "ready";
  readonly path: string;
  readonly choicesPath: string;
  readonly members: readonly Member[];
  readonly known: readonly KindEntry[];
  readonly instruments: Catalogue;
  readonly suggested: Readonly<Record<string, unknown>> | null;
  readonly onDisk: boolean;
  readonly touched: boolean;
  readonly edited: boolean;
  readonly saved: TwinChoices;
  readonly written: string;
  readonly doc: TwinChoices;
  readonly save: ChoiceSave;
  readonly build: Build;
  readonly issues: readonly TwinIssue[];
  readonly held: readonly Held[];
  readonly reviewed: readonly string[];
}

export type TwinState =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly path: string }
  | {
      readonly phase: "failed";
      readonly path: string;
      readonly failure: DeviceFailure;
      readonly held: readonly Held[];
    }
  | TwinDraft;

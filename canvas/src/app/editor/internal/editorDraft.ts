import type { Composition, IssueTargets } from "../../../model/editor";
import type { ComponentSpec, LibraryEntry } from "../../../model/plant";
import type { EditorFailure } from "../../ports";

type Outcome =
  | { readonly kind: "composed"; readonly summary: string; readonly revision: number }
  | { readonly kind: "refused"; readonly message: string | null; readonly revision: number }
  | null;

export interface Draft {
  readonly phase: "ready";
  readonly path: string;
  readonly etag: string;
  readonly domainPath: string;
  readonly domainEtag: string;
  readonly saved: Composition;
  readonly doc: Composition;
  readonly revision: number;
  readonly savedRevision: number;
  readonly saveStatus: "idle" | "saved-compose-failed";
  readonly composeFailure: EditorFailure | null;
  readonly past: readonly Composition[];
  readonly future: readonly Composition[];
  readonly specs: ReadonlyMap<string, ComponentSpec>;
  readonly library: readonly LibraryEntry[];
  readonly selected: string | null;
  readonly saving: boolean;
  readonly issues: IssueTargets | null;
  readonly outcome: Outcome;
  readonly failure: EditorFailure | null;
}

export type EditorState =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly path: string }
  | { readonly phase: "failed"; readonly path: string; readonly failure: EditorFailure }
  | Draft;

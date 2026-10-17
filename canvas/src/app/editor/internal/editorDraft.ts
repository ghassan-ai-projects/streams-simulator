import type { Composition, IssueTargets } from "../../../model/editor";
import type { ComponentSpec, LibraryEntry } from "../../../model/plant";
import type { EditorFailure } from "../../ports";

export type SaveState =
  | { readonly kind: "idle" }
  | { readonly kind: "saving" }
  | { readonly kind: "write-failed"; readonly failure: EditorFailure }
  | { readonly kind: "compose-failed"; readonly failure: EditorFailure }
  | { readonly kind: "composed"; readonly summary: string; readonly doc: Composition };

export interface Draft {
  readonly phase: "ready";
  readonly path: string;
  readonly domain: string | null;
  readonly saved: Composition;
  readonly doc: Composition;
  readonly save: SaveState;
  readonly issues: IssueTargets | null;
  readonly past: readonly Composition[];
  readonly future: readonly Composition[];
  readonly specs: ReadonlyMap<string, ComponentSpec>;
  readonly library: readonly LibraryEntry[];
  readonly selected: string | null;
}

export type EditorState =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly path: string }
  | { readonly phase: "failed"; readonly path: string; readonly failure: EditorFailure }
  | Draft;

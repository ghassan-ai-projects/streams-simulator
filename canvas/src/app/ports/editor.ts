import type { ComponentSpec } from "../../model/plant";
import type { LibraryEntry, Presets } from "../../model/plant";
import type { LintReport } from "./devices";
import type { ComposeSummary, Diagnostic, FileContent, Outcome, CompositionIssue } from "./shared";

export type EditorFailure =
  | { readonly kind: "missing" | "conflict"; readonly diagnostic: Diagnostic }
  | {
      readonly kind: "invalid-composition" | "compose-refused" | "unavailable" | "unreadable";
      readonly diagnostic: Diagnostic;
      readonly issues?: readonly CompositionIssue[];
    };

export interface EditorGateway {
  readFile(path: string): Promise<Outcome<FileContent, EditorFailure>>;
  library(): Promise<Outcome<readonly LibraryEntry[], EditorFailure>>;
  component(ref: string): Promise<Outcome<ComponentSpec, EditorFailure>>;
  presets(ref: string): Promise<Outcome<Presets, EditorFailure>>;
  lint(path: string): Promise<Outcome<LintReport, EditorFailure>>;
  writeFile(
    path: string,
    content: string,
    etag: string,
  ): Promise<Outcome<{ readonly path: string; readonly etag: string }, EditorFailure>>;
  compose(input: {
    readonly composition: string;
    readonly output: string;
    readonly etag: string;
  }): Promise<Outcome<ComposeSummary, EditorFailure>>;
}

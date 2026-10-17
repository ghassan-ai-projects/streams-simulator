import type { ComponentSpec } from "../../model/plant";
import type { Composition } from "../../model/editor";
import type { ExamplePlant } from "../../model/plant";
import type { ComposeSummary, CompositionIssue, Diagnostic, FileContent, Outcome, PlantSource } from "./shared";

export type PlantFailure =
  | { readonly kind: "missing" | "conflict"; readonly diagnostic: Diagnostic }
  | {
      readonly kind: "invalid-composition" | "compose-refused" | "unavailable" | "unreadable";
      readonly diagnostic: Diagnostic;
      readonly issues?: readonly CompositionIssue[];
    };

export interface PlantGateway {
  readFile(path: string): Promise<Outcome<FileContent, PlantFailure>>;
  component(ref: string): Promise<Outcome<ComponentSpec, PlantFailure>>;
  examples(): Promise<Outcome<readonly ExamplePlant[], PlantFailure>>;
  generateComposition(input: {
    readonly source: PlantSource;
    readonly id: string;
    readonly title: string;
    readonly mode: "copy" | "empty";
  }): Promise<Outcome<Composition, PlantFailure>>;
  writeFile(
    path: string,
    content: string,
    etag: string,
  ): Promise<Outcome<{ readonly path: string; readonly etag: string }, PlantFailure>>;
  compose(input: {
    readonly composition: string;
    readonly output: string;
    readonly etag: string;
  }): Promise<Outcome<ComposeSummary, PlantFailure>>;
}

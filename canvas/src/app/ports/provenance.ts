import type { Assumptions, Lineage, ModelDiff } from "../../model/workspace";
import type { DocumentFailure } from "./review";
import type { ComposeSummary, Outcome } from "./shared";

export type ProvenanceFailure = DocumentFailure;

export interface ProvenanceGateway {
  lineage(): Promise<Outcome<Lineage, ProvenanceFailure>>;
  assumptions(input: {
    readonly path: string;
    readonly twin: string | null;
  }): Promise<Outcome<Assumptions, ProvenanceFailure>>;
  diff(a: string, b: string): Promise<Outcome<ModelDiff, ProvenanceFailure>>;
  compose(input: {
    readonly composition: string;
    readonly output: string;
  }): Promise<Outcome<ComposeSummary, ProvenanceFailure>>;
}

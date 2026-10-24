import type { Assumptions, Lineage, ModelDiff } from "../../../model/workspace";
import type { ComposeSummary, Outcome, ProvenanceFailure } from "../../ports";
import type { Store } from "../../shared";
import type { Fetched } from "./fetchStore";

interface LineageSession extends Store<Lineage | null> {
  refresh(): Promise<void>;
}

export interface AssumptionsSession extends Store<Fetched<Assumptions>> {
  load(path: string, twin: string | null): Promise<void>;
  clear(): void;
}

export interface DiffSession extends Store<Fetched<ModelDiff>> {
  load(a: string, b: string): Promise<void>;
  clear(): void;
}

export interface ProvenanceSession {
  readonly lineage: LineageSession;
  readonly assumptions: AssumptionsSession;
  readonly diff: DiffSession;
  recompose(composition: string, domain: string): Promise<Outcome<ComposeSummary, ProvenanceFailure>>;
}

import type { Assumptions, ModelDiff } from "../../../model/workspace";
import type { ProvenanceGateway } from "../../ports";
import type { WorkspaceStore } from "./api";
import type { AssumptionsSession, DiffSession, ProvenanceSession } from "./provenanceApi";
import { FetchStore } from "./fetchStore";
import { LineageStore } from "./lineageStore";

class Assumed extends FetchStore<Assumptions> implements AssumptionsSession {
  constructor(private readonly gateway: ProvenanceGateway) {
    super();
  }

  load(path: string, twin: string | null): Promise<void> {
    return this.run(`${path}|${twin ?? ""}`, () => this.gateway.assumptions({ path, twin }));
  }
}

class Differed extends FetchStore<ModelDiff> implements DiffSession {
  constructor(private readonly gateway: ProvenanceGateway) {
    super();
  }

  load(a: string, b: string): Promise<void> {
    return this.run(`${a}|${b}`, () => this.gateway.diff(a, b));
  }
}

function followWorkspace(workspace: WorkspaceStore, lineage: LineageStore): void {
  let seen: unknown = null;
  workspace.subscribe(() => {
    const state = workspace.snapshot();
    if (state.phase !== "ready" || state.workspace === seen) return;
    seen = state.workspace;
    void lineage.refresh();
  });
}

export function createProvenanceSession(gateway: ProvenanceGateway, workspace: WorkspaceStore): ProvenanceSession {
  const lineage = new LineageStore(gateway);
  followWorkspace(workspace, lineage);
  return {
    lineage,
    assumptions: new Assumed(gateway),
    diff: new Differed(gateway),
    recompose: async (composition, output) => {
      const composed = await gateway.compose({ composition, output });
      if (composed.ok) void lineage.refresh();
      return composed;
    },
  };
}

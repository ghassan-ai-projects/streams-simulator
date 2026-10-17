import type { Workspace } from "../../model/workspace";
import type { Diagnostic, Outcome } from "./shared";

export type WorkspaceFailure =
  | { readonly kind: "unpaired"; readonly diagnostic: Diagnostic }
  | { readonly kind: "unavailable" | "unreadable"; readonly diagnostic: Diagnostic };

export interface WorkspaceGateway {
  load(): Promise<Outcome<Workspace, WorkspaceFailure>>;
  watch(onChange: () => void): () => void;
}

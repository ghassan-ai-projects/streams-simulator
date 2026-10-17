import type { Workspace } from "../../model/workspace";
import type { WorkspaceFailure } from "../ports";
import type { Store } from "../shared";

export type WorkspaceState =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: WorkspaceFailure }
  | { readonly phase: "ready"; readonly workspace: Workspace };

export interface WorkspaceStore extends Store<WorkspaceState> {
  open(): Promise<void>;
  close(): void;
}

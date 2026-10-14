import type { Workspace } from "./bridge";

function counted(count: number, one: string, many: string): string {
  return `${String(count)} ${count === 1 ? one : many}`;
}

export function workspaceSummary(workspace: Workspace): string {
  return [
    counted(workspace.domains.length, "plant", "plants"),
    counted(workspace.compositions.length, "composition", "compositions"),
    counted(workspace.scenarios.length, "scenario", "scenarios"),
    counted(workspace.runs.length, "run", "runs"),
  ].join(" · ");
}

export function workspaceEmpty(workspace: Workspace): boolean {
  return (
    workspace.domains.length + workspace.compositions.length + workspace.scenarios.length + workspace.runs.length === 0
  );
}

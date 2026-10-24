import type { LaunchStore } from "../../launch";
import type { WorkspaceStore } from "../../workspace";

export function resumeRun(launch: LaunchStore, workspace: WorkspaceStore): () => void {
  let asked = false;
  const attempt = () => {
    const listing = workspace.snapshot();
    if (asked || listing.phase !== "ready") return;
    asked = true;
    const { domains } = listing.workspace;
    void launch.resume((id) => {
      const plant = domains.find((domain) => domain.id === id);
      return plant === undefined ? null : { domain: id, title: plant.title, path: plant.path };
    });
  };
  const stop = workspace.subscribe(attempt);
  attempt();
  return () => {
    asked = true;
    stop();
  };
}

import type { LaunchStore } from "../../launch";
import type { ReviewState } from "../../review";
import { rerunProposalOf } from "../../review";
import type { NavigationStore } from "./api";

export function proposeRerun(
  launch: LaunchStore,
  navigation: NavigationStore,
  state: ReviewState,
  worldId: string,
  plantPath: string,
): void {
  const rerun = rerunProposalOf(state);
  if (rerun === null) return;
  if (state.connection !== "none") launch.connect.setUpNext(rerun.domainId, state.sessionPolicies ?? []);
  else if (rerun.fresh) launch.proposeFresh(rerun.domainId, rerun.plan, worldId);
  else launch.propose(rerun.domainId, rerun.plan, rerun.left, worldId);
  navigation.go({ screen: "plant", domain: plantPath, hall: null, launch: true });
}

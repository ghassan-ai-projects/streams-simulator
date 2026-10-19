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
  if (state.connection === "none") launch.propose(rerun.domainId, rerun.plan, rerun.left, worldId);
  else launch.connect.setUpNext(rerun.domainId, state.sessionPolicies ?? []);
  navigation.go({ screen: "plant", domain: plantPath, hall: null, launch: true });
}

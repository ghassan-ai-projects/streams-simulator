import type { LaunchStore } from "../../launch";
import type { ReviewState } from "../../review";
import { rerunProposalOf } from "../../review";
import type { NavigationStore } from "./api";

type RerunProposal = NonNullable<ReturnType<typeof rerunProposalOf>>;

export function proposeRerun(
  launch: LaunchStore,
  navigation: NavigationStore,
  state: ReviewState,
  worldId: string,
  plantPath: string,
): void {
  const rerun = rerunProposalOf(state);
  if (rerun === null) return;
  launch.reset();
  launch.connect.reset();
  if (state.connection !== "none") setUpSession(launch, rerun, state.sessionPolicies ?? []);
  else setUpSimulation(launch, rerun, worldId);
  navigation.go({ screen: "plant", domain: plantPath, hall: null, launch: true });
}

function setUpSession(launch: LaunchStore, rerun: RerunProposal, policies: readonly string[]): void {
  if (rerun.fresh) launch.connect.setUpFresh(rerun.domainId);
  else launch.connect.setUpNext(rerun.domainId, policies);
}

function setUpSimulation(launch: LaunchStore, rerun: RerunProposal, worldId: string): void {
  if (rerun.fresh) launch.proposeFresh(rerun.domainId, rerun.plan, worldId);
  else launch.propose(rerun.domainId, rerun.plan, rerun.left, worldId);
}

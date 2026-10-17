import { useEffect, useState } from "react";
import type { LaunchStore, Proposal } from "../../../app/launch";
import type { Draft } from "../../../model/launch";
import { draftOf, firstDraft, keepKnownFaults, pickedOnGraph } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { faultsOf } from "../../../model/plant";

function prefillOf(proposal: Proposal | null, plant: Plant): Proposal | null {
  if (proposal === null) return null;
  const kept = keepKnownFaults(proposal.plan, (entity) => faultsOf(plant, entity));
  return { ...proposal, plan: kept.plan, left: [...proposal.left, ...kept.dropped] };
}

export function useLaunchDraft(launch: LaunchStore, plant: Plant, selected: string | null) {
  const [proposed] = useState(() => prefillOf(launch.proposal(plant.id), plant));
  useEffect(
    () => () => {
      launch.forgetProposal();
    },
    [launch],
  );
  const [draft, setDraft] = useState<Draft>(() => (proposed === null ? firstDraft : draftOf(proposed.plan)));
  const [seen, setSeen] = useState(selected);
  if (seen !== selected) {
    setSeen(selected);
    if (selected !== null) setDraft((current) => pickedOnGraph(current, selected));
  }
  return { proposed, draft, setDraft };
}

import { useCallback, useMemo } from "react";
import type { PlantState } from "../../app/plantStore";
import type { ReviewState, ReviewStore } from "../../app/reviewStore";
import { channelOf } from "../../model/plant";
import { hallOf } from "../../model/plantView";
import { attentionOf, showReading, worstOf } from "../../model/readout";
import { effectorAt, faultedAt } from "../../model/timeline";
import type { NodeLive } from "../plant/flow";

type ReadyPlant = Extract<PlantState, { phase: "ready" }>;

export function useLive(review: ReviewStore, state: ReviewState, plant: ReadyPlant | null) {
  const report = state.report.phase === "ready" ? state.report.value : null;
  const faulted = useMemo(() => faultedAt(report, state.positionNs), [report, state.positionNs]);
  const effectors = useMemo(() => effectorAt(report, state.positionNs), [report, state.positionNs]);
  const one = useCallback(
    (id: string): NodeLive => {
      const shown = review.index
        .readings(id, state.positionNs)
        .map((reading) => showReading(reading, plant === null ? null : channelOf(plant.plant, id, reading.channel)));
      const attention = attentionOf(shown);
      return {
        shown,
        observed: review.index.observed(id),
        status: attention === "alarm" ? "alarm" : faulted.has(id) ? "fault" : attention,
        count: 0,
        fault: (faulted.get(id) ?? []).join(", "),
        effector: effectors.get(id) ?? "",
      };
    },
    [review, state.positionNs, plant, faulted, effectors],
  );
  return useCallback(
    (id: string): NodeLive => {
      const members = plant?.plant.instances.filter((instance) => hallOf(instance.id) === id) ?? [];
      if (members.length === 0) return one(id);
      const worst = worstOf(members.map((member) => one(member.id).status));
      return {
        shown: [],
        observed: new Set<string>(),
        status: worst.status,
        count: worst.count,
        fault: "",
        effector: "",
      };
    },
    [one, plant],
  );
}

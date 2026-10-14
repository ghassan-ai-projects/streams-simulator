import { useCallback, useMemo } from "react";
import type { PlantState } from "../../app/plantStore";
import type { ReviewState, ReviewStore } from "../../app/reviewStore";
import { statesOf } from "../../model/plant";
import { hallOf } from "../../model/plantView";
import { abnormalFirst, attentionOf, showReading, withFault, worstOf } from "../../model/readout";
import { faultedAt } from "../../model/timeline";
import type { NodeLive } from "../plant/flow";

type ReadyPlant = Extract<PlantState, { phase: "ready" }>;

export function useLive(review: ReviewStore, state: ReviewState, plant: ReadyPlant | null) {
  const report = state.report.phase === "ready" ? state.report.value : null;
  const faulted = useMemo(() => faultedAt(report, state.positionNs), [report, state.positionNs]);
  const one = useCallback(
    (id: string): NodeLive => {
      const shown = review.index
        .readings(id, state.positionNs)
        .map((reading) => showReading(reading, plant === null ? [] : statesOf(plant.plant, id, reading.channel)));
      const attention = attentionOf(shown);
      return {
        shown: withFault(abnormalFirst(shown), (faulted.get(id) ?? []).join(", ")),
        status: attention === "alarm" ? "alarm" : faulted.has(id) ? "fault" : attention,
        count: 0,
        fault: (faulted.get(id) ?? []).join(", "),
      };
    },
    [review, state.positionNs, plant, faulted],
  );
  return useCallback(
    (id: string): NodeLive => {
      const members = plant?.plant.instances.filter((instance) => hallOf(instance.id) === id) ?? [];
      if (members.length === 0) return one(id);
      const worst = worstOf(members.map((member) => one(member.id).status));
      return { shown: [], status: worst.status, count: worst.count, fault: "" };
    },
    [one, plant],
  );
}

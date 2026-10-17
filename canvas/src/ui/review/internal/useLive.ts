import { useCallback, useMemo } from "react";
import type { PlantState } from "../../../app/plant";
import type { ReviewState, ReviewSession } from "../../../app/review";
import { channelOf } from "../../../model/plant";
import type { NodeLive } from "../../graph";
import type { NodeStatus, PlantView, ReadoutBody } from "../../../model/graph";
import { attentionOf, hallOf, showReading, worstOf } from "../../../model/graph";
import { readoutBody } from "../../../model/review";
import { effectorAt, faultedAt } from "../../../model/review";

type ReadyPlant = Extract<PlantState, { phase: "ready" }>;

const noReadout: ReadoutBody = { rows: [], hidden: 0, hiddenAttention: null };

function hallLive(worst: { readonly status: NodeStatus; readonly count: number }): NodeLive {
  return {
    shown: [],
    observed: new Set<string>(),
    readout: noReadout,
    status: worst.status,
    count: worst.count,
    fault: "",
    effector: "",
  };
}

export function useLive(review: ReviewSession, state: ReviewState, plant: ReadyPlant | null, view: PlantView | null) {
  const report = state.report.phase === "ready" ? state.report.value : null;
  const faulted = useMemo(() => faultedAt(report, state.positionNs), [report, state.positionNs]);
  const effectors = useMemo(() => effectorAt(report, state.positionNs), [report, state.positionNs]);
  const one = useCallback(
    (id: string): NodeLive => {
      const shown = review.index
        .readings(id, state.positionNs)
        .map((reading) => showReading(reading, plant === null ? null : channelOf(plant.plant, id, reading.channel)));
      const observed = review.index.observed(id);
      const node = view === null ? undefined : view.nodes.find((entry) => entry.id === id);
      const attention = attentionOf(shown);
      return {
        shown,
        observed,
        readout: node === undefined ? noReadout : readoutBody(node.pins, node.channels, shown, observed),
        status: attention === "alarm" ? "alarm" : faulted.has(id) ? "fault" : attention,
        count: 0,
        fault: (faulted.get(id) ?? []).join(", "),
        effector: effectors.get(id) ?? "",
      };
    },
    [review, state.positionNs, plant, view, faulted, effectors],
  );
  return useCallback(
    (id: string): NodeLive => {
      const members = plant?.plant.instances.filter((instance) => hallOf(instance.id) === id) ?? [];
      if (members.length === 0) return one(id);
      return hallLive(worstOf(members.map((member) => one(member.id).status)));
    },
    [one, plant],
  );
}

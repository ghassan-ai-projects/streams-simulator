import { useMemo } from "react";
import type { LaunchStore } from "../../app/launchStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantStore } from "../../app/plantStore";
import type { PlanMark } from "../../model/planMarks";
import { planMarks } from "../../model/planMarks";
import { plantView } from "../../model/plantView";
import { useStore } from "../useStore";
import { useCanvasAspect } from "./canvasAspect";
import { PlantGraph } from "./PlantGraph";
import styles from "./PlantScreen.module.css";

type Ready = Extract<PlantState, { phase: "ready" }>;

function useRunMarks(launch: LaunchStore, plant: string): ReadonlyMap<string, PlanMark> | undefined {
  const run = useStore(launch);
  return useMemo(() => {
    if (run.phase !== "running" || run.origin.domain !== plant) return undefined;
    return planMarks(run.plan, run.calls, run.atS);
  }, [run, plant]);
}

export function PlantCanvas({
  state,
  hall,
  plants,
  navigation,
  launch,
}: {
  readonly state: Ready;
  readonly hall: string | null;
  readonly plants: PlantStore;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
}) {
  const { ref, aspect } = useCanvasAspect<HTMLDivElement>();
  const marks = useRunMarks(launch, state.plant.id);
  const view = useMemo(
    () => plantView(state.plant, state.specs, hall, false, aspect),
    [state.plant, state.specs, hall, aspect],
  );
  return (
    <div ref={ref} className={styles.graphHost}>
      <PlantGraph
        view={view}
        viewKey={hall ?? "plant"}
        fitKey={`${hall ?? "plant"}:${String(aspect)}`}
        selected={state.selected}
        {...(marks === undefined ? {} : { marks })}
        onSelect={(id) => {
          plants.select(id);
        }}
        onOpenHall={(name) => {
          navigation.go({ screen: "plant", domain: state.path, hall: name });
        }}
        onEscape={() => {
          if (hall !== null) navigation.go({ screen: "plant", domain: state.path, hall: null });
        }}
      />
    </div>
  );
}

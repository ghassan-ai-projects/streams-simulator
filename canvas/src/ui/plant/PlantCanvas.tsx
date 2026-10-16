import { useMemo } from "react";
import type { LaunchStore } from "../../app/launchStore";
import type { NavigationStore } from "../../app/navigationStore";
import type { PlantState, PlantStore } from "../../app/plantStore";
import { showingMarks } from "../../model/launch";
import { planMarks, plainView, plantView } from "../../model/graph";
import type { PlanMark, PlantView } from "../../model/graph";
import { PlantGraph, useCanvasAspect } from "../graph";
import { useStore } from "../useStore";
import styles from "./PlantScreen.module.css";

type Ready = Extract<PlantState, { phase: "ready" }>;

function useRunMarks(launch: LaunchStore, plant: string): ReadonlyMap<string, PlanMark> | undefined {
  const run = useStore(launch);
  return useMemo(() => {
    if (run.phase !== "running" || run.origin.domain !== plant) return undefined;
    return planMarks(showingMarks(run.plan, run.calls, run.atS, run.atMs));
  }, [run, plant]);
}

function PlantViewGraph({
  view,
  state,
  hall,
  aspect,
  plants,
  navigation,
  marks,
}: {
  readonly view: PlantView;
  readonly state: Ready;
  readonly hall: string | null;
  readonly aspect: number | undefined;
  readonly plants: PlantStore;
  readonly navigation: NavigationStore;
  readonly marks: ReadonlyMap<string, PlanMark> | undefined;
}) {
  return (
    <PlantGraph
      view={view}
      viewKey={hall ?? "plant"}
      fitKey={`${hall ?? "plant"}:${String(aspect)}`}
      panCue
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
  );
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
    () => plantView(state.plant, state.specs, hall, plainView, aspect),
    [state.plant, state.specs, hall, aspect],
  );
  return (
    <div ref={ref} className={styles.graphHost}>
      <PlantViewGraph
        view={view}
        state={state}
        hall={hall}
        aspect={aspect}
        plants={plants}
        navigation={navigation}
        marks={marks}
      />
    </div>
  );
}

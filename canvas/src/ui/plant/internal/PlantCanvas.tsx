import { screen } from "../../primitives";
import { useMemo } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import type { PlantSession, ReadyPlant } from "../../../app/plant";
import { showingMarks } from "../../../model/launch";
import { planMarks, plantView } from "../../../model/graph";
import type { PlanMark, PlantView } from "../../../model/graph";
import { PlantGraph, useCanvasAspect } from "../../graph";
import { useStore } from "../../primitives";
import type { Connected } from "./connected";

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
  connected,
}: {
  readonly view: PlantView;
  readonly state: ReadyPlant;
  readonly hall: string | null;
  readonly aspect: number | undefined;
  readonly plants: PlantSession;
  readonly navigation: NavigationStore;
  readonly marks: ReadonlyMap<string, PlanMark> | undefined;
  readonly connected?: Connected;
}) {
  return (
    <PlantGraph
      view={view}
      viewKey={hall ?? "plant"}
      fitKey={`${hall ?? "plant"}:${String(aspect)}`}
      panCue
      selected={state.selected}
      {...(connected === undefined ? {} : { live: connected.live })}
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

function useLaidOutPlant(
  state: ReadyPlant,
  hall: string | null,
  aspect: number | undefined,
  connected: Connected | undefined,
): PlantView {
  const presentation = connected?.view;
  const notes = connected?.notes;
  return useMemo(
    () =>
      plantView(state.plant, state.specs, {
        open: hall,
        aspect,
        ...(presentation === undefined ? {} : { view: presentation }),
        ...(notes === undefined ? {} : { notes }),
      }),
    [state.plant, state.specs, hall, aspect, presentation, notes],
  );
}

export function PlantCanvas({
  state,
  hall,
  plants,
  navigation,
  launch,
  connected,
}: {
  readonly state: ReadyPlant;
  readonly hall: string | null;
  readonly plants: PlantSession;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly connected?: Connected;
}) {
  const { ref, aspect } = useCanvasAspect<HTMLDivElement>();
  const marks = useRunMarks(launch, state.plant.id);
  const view = useLaidOutPlant(state, hall, aspect, connected);
  return (
    <div ref={ref} className={screen.graphHost}>
      <PlantViewGraph
        view={view}
        state={state}
        hall={hall}
        aspect={aspect}
        plants={plants}
        navigation={navigation}
        marks={marks}
        {...(connected === undefined ? {} : { connected })}
      />
    </div>
  );
}

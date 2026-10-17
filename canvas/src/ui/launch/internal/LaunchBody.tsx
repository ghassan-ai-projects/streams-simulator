import type { LaunchState, LaunchStore, RunWatch } from "../../../app/launch";
import type { Draft } from "../../../model/launch";
import type { Plant } from "../../../model/plant";
import { plantTitle } from "../../../model/plant";
import type { Change } from "./FaultRows";
import { Failure } from "./LaunchFailure";
import { LaunchForm } from "./LaunchForm";
import { Result } from "./LaunchResult";
import { Elsewhere, Progress } from "./LaunchProgress";

interface BodyProps {
  readonly state: LaunchState;
  readonly launch: LaunchStore;
  readonly plant: Plant;
  readonly path: string;
  readonly draft: Draft;
  readonly selected: string | null;
  readonly change: (change: Change) => void;
  readonly onFinished: (worldId: string) => void;
  readonly watch: RunWatch;
}

export function Body({ state, launch, plant, path, draft, selected, change, onFinished, watch }: BodyProps) {
  const own = state.phase !== "idle" && state.origin.domain === plant.id;
  if (state.phase === "running") {
    return own ? <Progress state={state} launch={launch} title={plantTitle(plant)} /> : <Elsewhere state={state} />;
  }
  if (state.phase === "failed" && own) return <Failure state={state} launch={launch} />;
  if (watch.shows(state))
    return (
      <Result
        state={state}
        onOpen={() => {
          launch.reset();
          onFinished(state.worldId);
        }}
      />
    );
  return (
    <LaunchForm
      plant={plant}
      draft={draft}
      selected={selected}
      change={change}
      onStart={(plan) => {
        void launch.start({ domain: plant.id, title: plantTitle(plant), path }, plan);
      }}
    />
  );
}

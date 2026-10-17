import { panel } from "../../primitives";
import { Play, X } from "lucide-react";
import { useEffect, useMemo } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { Plant } from "../../../model/plant";
import { plantTitle } from "../../../model/plant";
import { useStore } from "../../primitives";
import { Body } from "./LaunchBody";
import { ProposalNote } from "./ProposalNote";
import { RunningMark } from "./RunningMark";
import { useLaunchDraft } from "./useLaunchDraft";
import styles from "./LaunchPanel.module.css";

interface LaunchPanelProps {
  readonly launch: LaunchStore;
  readonly plant: Plant;
  readonly path: string;
  readonly selected?: string | null;
  readonly onClose: () => void;
  readonly onFinished: (worldId: string) => void;
}

function useLaunch(launch: LaunchStore, plant: string, onFinished: (worldId: string) => void) {
  const state = useStore(launch);
  const watch = useMemo(() => launch.watch(plant), [launch, plant]);
  useEffect(() => watch.holdFinished(onFinished), [watch, onFinished]);
  return { state, watch };
}

export function LaunchPanel({ launch, plant, path, selected = null, onClose, onFinished }: LaunchPanelProps) {
  const { state, watch } = useLaunch(launch, plant.id, onFinished);
  const { proposed, draft, setDraft } = useLaunchDraft(launch, plant, selected);
  return (
    <aside className={panel.panel} aria-labelledby="launch-title">
      <header className={panel.head}>
        <Play size={16} aria-hidden="true" />
        <h2 id="launch-title" className={styles.title}>{`Run ${plantTitle(plant)}`}</h2>
        {state.phase === "running" && state.origin.domain === plant.id && <RunningMark state={state} launch={launch} />}
        <button type="button" className={panel.icon} aria-label="Close the run panel" onClick={onClose}>
          <X size={16} aria-hidden="true" />
        </button>
      </header>
      {state.phase !== "running" && <ProposalNote proposed={proposed} />}
      <Body
        state={state}
        launch={launch}
        plant={plant}
        path={path}
        draft={draft}
        selected={selected}
        change={setDraft}
        onFinished={onFinished}
        watch={watch}
      />
    </aside>
  );
}

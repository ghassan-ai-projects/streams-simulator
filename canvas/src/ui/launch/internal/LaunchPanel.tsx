import { panel } from "../../primitives";
import { Play, X } from "lucide-react";
import { useEffect, useMemo } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { Plant } from "../../../model/plant";
import { plantTitle } from "../../../model/plant";
import { useStore } from "../../primitives";
import { Body } from "./LaunchBody";
import { ProposalNote } from "./ProposalNote";
import { RunMode } from "./RunMode";
import { RunningMark } from "./RunningMark";
import { Session } from "./Session";
import { useLaunchDraft } from "./useLaunchDraft";
import { useRunMode } from "./useRunMode";
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

type CompleteProps = Omit<LaunchPanelProps, "onClose" | "selected"> & { readonly selected: string | null };

function ToCompletion({ launch, plant, path, selected, onFinished }: CompleteProps) {
  const { state, watch } = useLaunch(launch, plant.id, onFinished);
  const { proposed, draft, setDraft } = useLaunchDraft(launch, plant, selected);
  return (
    <>
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
    </>
  );
}

function lockWords(running: boolean, held: boolean): string {
  if (held) return "A connected session holds this plant until you stop it.";
  return running ? "A run is in progress." : "";
}

export function LaunchPanel({ launch, plant, path, selected = null, onClose, onFinished }: LaunchPanelProps) {
  const state = useStore(launch);
  const mode = useRunMode(launch.connect, plant.id);
  const running = state.phase === "running";
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
      {!mode.held && <RunMode kind={mode.kind} locked={lockWords(running, mode.held)} onKind={mode.choose} />}
      {mode.kind === "session" ? (
        <Session launch={launch} plant={plant} onFinished={onFinished} />
      ) : (
        <ToCompletion launch={launch} plant={plant} path={path} selected={selected} onFinished={onFinished} />
      )}
    </aside>
  );
}

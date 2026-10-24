import { cx, shared } from "../../primitives";
import { Loader, Pause, PenLine, Play } from "lucide-react";
import type { RefObject } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import { runStageWords } from "../../../model/launch";
import { clockOf } from "../../../model/shared/duration";
import { useStore } from "../../primitives";
import styles from "./PlantBar.module.css";

export interface RunningHere {
  readonly words: string;
  readonly paused: boolean;
}

export function useRunningHere(launch: LaunchStore, plant: string): RunningHere | null {
  const state = useStore(launch);
  if (state.phase !== "running" || state.origin.domain !== plant) return null;
  return { words: `${runStageWords(state)} · ${clockOf(state.atS)}`, paused: state.paused };
}

function RunIcon({ running }: { readonly running: RunningHere | null }) {
  if (running === null) return <Play size={14} aria-hidden="true" />;
  if (running.paused) return <Pause size={14} aria-hidden="true" />;
  return <Loader size={14} aria-hidden="true" className={shared.spin} />;
}

interface RunButtonProps {
  readonly launching: boolean;
  readonly running: RunningHere | null;
  readonly button: RefObject<HTMLButtonElement | null>;
  readonly onToggle: () => void;
  readonly secondary?: boolean;
}

export function RunButton({ launching, running, button, onToggle, secondary = false }: RunButtonProps) {
  return (
    <button
      type="button"
      ref={button}
      className={cx(secondary ? shared.secondary : shared.primary, styles.run)}
      aria-pressed={launching}
      onClick={onToggle}
    >
      <RunIcon running={running} />
      {running === null ? "Run plant" : running.words}
    </button>
  );
}

interface EditButtonProps {
  readonly composition: string | null;
  readonly running: RunningHere | null;
  readonly navigation: NavigationStore;
}

export function EditButton({ composition, running, navigation }: EditButtonProps) {
  if (composition === null) return null;
  const busy = running !== null;
  return (
    <button
      type="button"
      className={shared.secondary}
      aria-disabled={busy}
      title={
        busy ? "Editing waits for the run: saving would change the plant file the run is checked against" : undefined
      }
      onClick={() => {
        if (!busy) navigation.go({ screen: "edit", composition });
      }}
    >
      <PenLine size={14} aria-hidden="true" />
      Edit plant
    </button>
  );
}

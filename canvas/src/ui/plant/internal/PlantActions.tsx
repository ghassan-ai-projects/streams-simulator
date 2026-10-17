import { shared } from "../../primitives";
import { Loader, PenLine, Play } from "lucide-react";
import type { RefObject } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import { clockOf } from "../../../model/shared/duration";
import { useStore } from "../../primitives";
import styles from "./PlantScreen.module.css";

export function useRunningHere(launch: LaunchStore, plant: string): string | null {
  const state = useStore(launch);
  if (state.phase !== "running" || state.origin.domain !== plant) return null;
  return clockOf(state.atS);
}

interface RunButtonProps {
  readonly launching: boolean;
  readonly running: string | null;
  readonly button: RefObject<HTMLButtonElement | null>;
  readonly onToggle: () => void;
}

export function RunButton({ launching, running, button, onToggle }: RunButtonProps) {
  return (
    <button type="button" ref={button} className={styles.run} aria-pressed={launching} onClick={onToggle}>
      {running === null ? (
        <Play size={14} aria-hidden="true" />
      ) : (
        <Loader size={14} aria-hidden="true" className={shared.spin} />
      )}
      {running === null ? "Run plant" : `Running · ${running}`}
    </button>
  );
}

interface EditButtonProps {
  readonly composition: string | null;
  readonly running: string | null;
  readonly navigation: NavigationStore;
}

export function EditButton({ composition, running, navigation }: EditButtonProps) {
  if (composition === null) return null;
  const busy = running !== null;
  return (
    <button
      type="button"
      className={shared.button}
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

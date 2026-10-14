import { CircleAlert, CircleCheck, Loader, X } from "lucide-react";
import type { LaunchState, LaunchStore } from "../../app/launchStore";
import type { NavigationStore } from "../../app/navigationStore";
import { clockOf } from "../../model/launch";
import { useStore } from "../useStore";
import styles from "./TopBar.module.css";

interface IndicatorProps {
  readonly launch: LaunchStore;
  readonly navigation: NavigationStore;
}

type Shown = Exclude<LaunchState, { phase: "idle" }>;

function words(state: Shown): string {
  if (state.phase === "running") {
    const at = `${clockOf(state.atS)} of ${clockOf(state.durationS)}`;
    return state.canceling ? `Canceling the run of ${state.origin.title}` : `Running ${state.origin.title} · ${at}`;
  }
  if (state.phase === "done") return `Run ${state.worldId} of ${state.origin.title} finished · Review`;
  return `The run of ${state.origin.title} did not finish`;
}

const icons = { running: Loader, done: CircleCheck, failed: CircleAlert } as const;

export function RunIndicator({ launch, navigation }: IndicatorProps) {
  const state = useStore(launch);
  if (state.phase === "idle") return null;
  const Icon = icons[state.phase];
  const open = () => {
    if (state.phase === "done") {
      launch.reset();
      navigation.go({ screen: "run", worldId: state.worldId });
    } else navigation.go({ screen: "plant", domain: state.origin.path, hall: null, launch: true });
  };
  return (
    <div className={styles[`run-${state.phase}`]} role="status">
      <button type="button" className={styles.runOpen} onClick={open}>
        <Icon size={14} aria-hidden="true" className={state.phase === "running" ? styles.spin : undefined} />
        {words(state)}
      </button>
      {state.phase !== "running" && (
        <button
          type="button"
          className={styles.runDismiss}
          aria-label="Dismiss"
          onClick={() => {
            launch.reset();
          }}
        >
          <X size={14} aria-hidden="true" />
        </button>
      )}
    </div>
  );
}

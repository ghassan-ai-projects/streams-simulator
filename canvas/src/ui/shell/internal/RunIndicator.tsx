import { CircleAlert, CircleCheck, Loader, X } from "lucide-react";
import type { LaunchState, LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import { rateWords } from "../../../model/launch";
import { clockOf } from "../../../model/shared/duration";
import { shared, useStore } from "../../primitives";
import styles from "./TopBar.module.css";

interface IndicatorProps {
  readonly launch: LaunchStore;
  readonly navigation: NavigationStore;
}

type Shown = Exclude<LaunchState, { phase: "idle" }>;

function words(state: Shown): string {
  if (state.phase === "running") {
    const speed = rateWords(state.rate);
    const at = `${clockOf(state.atS)} of ${clockOf(state.durationS)}${speed === "" ? "" : ` · ${speed}`}`;
    return state.stopping ? `Stopping the run of ${state.origin.title}` : `Running ${state.origin.title} · ${at}`;
  }
  if (state.phase === "done") {
    return `Run ${state.worldId} ${state.early ? "stopped early" : "finished"} · Open review`;
  }
  return `The run of ${state.origin.title} did not finish`;
}

function announcement(state: LaunchState): string {
  if (state.phase === "idle") return "";
  if (state.phase === "running") return `Run of ${state.origin.title} started`;
  if (state.phase === "done") return `The review of run ${state.worldId} is ready`;
  return words(state);
}

const icons = { running: Loader, done: CircleCheck, failed: CircleAlert } as const;

function dismiss(launch: LaunchStore) {
  launch.reset();
  document.querySelector<HTMLElement>("h1")?.focus();
}

export function RunIndicator(props: IndicatorProps) {
  const state = useStore(props.launch);
  return (
    <>
      <p className="sr-only" role="status">
        {announcement(state)}
      </p>
      {state.phase !== "idle" && <Pill {...props} state={state} />}
    </>
  );
}

function Pill({ launch, navigation, state }: IndicatorProps & { readonly state: Shown }) {
  const Icon = icons[state.phase];
  const open = () => {
    if (state.phase === "done") {
      launch.reset();
      navigation.go({ screen: "run", worldId: state.worldId });
    } else navigation.go({ screen: "plant", domain: state.origin.path, hall: null, launch: true });
  };
  return (
    <div className={styles[`run-${state.phase}`]}>
      <button type="button" className={styles.runOpen} onClick={open}>
        <Icon size={14} aria-hidden="true" className={state.phase === "running" ? shared.spin : undefined} />
        {words(state)}
      </button>
      {state.phase !== "running" && (
        <button
          type="button"
          className={styles.runDismiss}
          aria-label="Dismiss"
          onClick={() => {
            dismiss(launch);
          }}
        >
          <X size={14} aria-hidden="true" />
        </button>
      )}
    </div>
  );
}

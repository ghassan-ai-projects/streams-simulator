import { CircleAlert, CircleCheck, Loader, Pause, X } from "lucide-react";
import { useEffect } from "react";
import type { LaunchState, LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import { runProgressWords, runStageWords, runTitle } from "../../../model/launch";
import { shared, useStore } from "../../primitives";
import styles from "./RunIndicator.module.css";

interface IndicatorProps {
  readonly launch: LaunchStore;
  readonly navigation: NavigationStore;
}

type Shown = Exclude<LaunchState, { phase: "idle" }>;

interface Words {
  readonly head: string;
  readonly name: string;
  readonly tail: string;
}

function parts(state: Shown): Words {
  if (state.phase === "running") {
    const name = state.plan.name ?? state.origin.title;
    if (state.stopping) return { head: "Stopping the run of", name, tail: "" };
    return { head: runStageWords(state), name, tail: `· ${runProgressWords(state)}` };
  }
  if (state.phase === "done") {
    const named = state.plan.name === undefined ? "" : ` (${state.plan.name})`;
    const how = state.early ? "stopped early" : "finished";
    return { head: `Run ${state.worldId}${named} ${how}`, name: "", tail: "· Open review" };
  }
  return { head: `The run of ${state.origin.title} did not finish`, name: "", tail: "" };
}

function words(state: Shown): string {
  const { head, name, tail } = parts(state);
  return [head, name, tail].filter((part) => part !== "").join(" ");
}

function announcement(state: LaunchState): string {
  if (state.phase === "idle") return "";
  if (state.phase === "running")
    return `Run of ${runTitle(state.plan, state.origin.title)} ${state.paused ? "paused" : "started"}`;
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
  const connected = useStore(props.launch.connect);
  const route = useStore(props.navigation);
  const watching = connected.phase === "live" && route.screen === "review" && route.worldId === connected.run;
  useEffect(() => {
    void props.launch.connect.resume();
  }, [props.launch]);
  return (
    <>
      <p className="sr-only" role="status">
        {announcement(state)}
      </p>
      {state.phase !== "idle" && <Pill {...props} state={state} />}
      {connected.phase === "live" && !watching && (
        <div className={styles["run-running"]}>
          <button
            type="button"
            className={styles.runOpen}
            onClick={() => {
              props.navigation.go({ screen: "review", worldId: connected.run });
            }}
          >
            <Loader size={14} aria-hidden="true" />
            Connected session · running · Open live view
          </button>
        </div>
      )}
    </>
  );
}

function Label({ head, name, tail }: Words) {
  return (
    <>
      {name === "" ? <span className={styles.runName}>{head}</span> : head}
      {name !== "" && (
        <>
          {" "}
          <span className={styles.runName} title={name}>
            {name}
          </span>
        </>
      )}
      {tail !== "" && (
        <>
          {" "}
          <span className={styles.runTail}>{tail}</span>
        </>
      )}
    </>
  );
}

function Pill({ launch, navigation, state }: IndicatorProps & { readonly state: Shown }) {
  const Icon = state.phase === "running" && state.paused ? Pause : icons[state.phase];
  const open = () => {
    if (state.phase === "done") {
      launch.reset();
      navigation.go({ screen: "review", worldId: state.worldId });
    } else navigation.go({ screen: "plant", domain: state.origin.path, hall: null, launch: true });
  };
  return (
    <div className={styles[`run-${state.phase}`]}>
      <button type="button" className={styles.runOpen} onClick={open}>
        <Icon
          size={14}
          aria-hidden="true"
          className={state.phase === "running" && !state.paused ? shared.spin : undefined}
        />
        <Label {...parts(state)} />
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

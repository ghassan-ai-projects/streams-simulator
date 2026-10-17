import { useState } from "react";
import type { LaunchState } from "../../../app/launch";
import { clockOf } from "../../../model/shared/duration";
import { useInterval } from "../../primitives";
import styles from "./LaunchPanel.module.css";

type Running = Extract<LaunchState, { phase: "running" }>;

export function RunningMark({ state }: { readonly state: Running }) {
  const [now, setNow] = useState(state.startedMs);
  useInterval(() => {
    setNow(Date.now());
  }, 1000);
  const elapsed = clockOf((now - state.startedMs) / 1000);
  return <span className={styles.mark}>{`${elapsed} elapsed`}</span>;
}

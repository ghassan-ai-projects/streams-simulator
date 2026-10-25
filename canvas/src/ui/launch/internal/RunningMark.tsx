import { useState } from "react";
import type { LaunchState, LaunchStore } from "../../../app/launch";
import { clockOf } from "../../../model/shared/duration";
import { useInterval } from "../../primitives";
import styles from "./LaunchPanel.module.css";

type Running = Extract<LaunchState, { phase: "running" }>;

export function RunningMark({ state, launch }: { readonly state: Running; readonly launch: LaunchStore }) {
  const [now, setNow] = useState(() => launch.nowMs());
  useInterval(() => {
    setNow(launch.nowMs());
  }, 1000);
  if (state.plan.recovered === true) return null;
  const elapsed = clockOf((now - state.startedMs) / 1000);
  return <span className={styles.mark}>{`${elapsed} elapsed`}</span>;
}

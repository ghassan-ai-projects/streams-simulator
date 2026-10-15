import { useEffect, useState } from "react";
import type { LaunchState } from "../../app/launchStore";
import { clockOf } from "../../model/launch";
import styles from "./LaunchPanel.module.css";

type Running = Extract<LaunchState, { phase: "running" }>;

export function RunningMark({ state }: { readonly state: Running }) {
  const [now, setNow] = useState(state.startedMs);
  useEffect(() => {
    const timer = window.setInterval(() => {
      setNow(Date.now());
    }, 1000);
    return () => {
      window.clearInterval(timer);
    };
  }, []);
  const elapsed = clockOf((now - state.startedMs) / 1000);
  return <span className={styles.mark}>{`${elapsed} elapsed`}</span>;
}

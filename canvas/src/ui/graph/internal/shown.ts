import type { NodeStatus, PlanMark } from "../../../model/graph";
import type { NodeLive } from "./types";
import styles from "./nodes.module.css";

export interface Shown {
  readonly status: NodeStatus;
  readonly count: number;
  readonly fault: string;
  readonly effector: string;
}

export function shownOf(data: { readonly live: NodeLive | null; readonly plan: PlanMark | null }): Shown {
  const { live, plan } = data;
  if (live !== null) return live;
  return plan ?? { status: null, count: 0, fault: "", effector: "" };
}

export function statusClass(status: NodeStatus): string | undefined {
  return status === null ? undefined : styles[`status-${status}`];
}

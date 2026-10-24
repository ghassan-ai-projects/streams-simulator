import type { NodeStatus, PlanMark } from "../../../model/graph";
import { mergeMarks } from "../../../model/graph";
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
  if (live === null) return plan ?? { status: null, count: 0, fault: "", effector: "" };
  return plan === null ? live : mergeMarks(live, plan);
}

export function statusClass(status: NodeStatus): string | undefined {
  return status === null ? undefined : styles[`status-${status}`];
}

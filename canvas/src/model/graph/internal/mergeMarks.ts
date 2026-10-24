import type { NodeStatus } from "./types";

interface Marks {
  readonly status: NodeStatus;
  readonly count: number;
  readonly fault: string;
  readonly effector: string;
}

const rank = { alarm: 0, fault: 1, warn: 2 } as const;

function worse(a: NodeStatus, b: NodeStatus): NodeStatus {
  if (a === null || b === null) return a ?? b;
  return rank[a] <= rank[b] ? a : b;
}

export function mergeMarks(live: Marks, plan: Marks): Marks {
  const status = worse(live.status, plan.status);
  return {
    status,
    count: status === plan.status && plan.status !== null ? plan.count : live.count,
    fault: live.fault === "" ? plan.fault : live.fault,
    effector: live.effector === "" ? plan.effector : live.effector,
  };
}

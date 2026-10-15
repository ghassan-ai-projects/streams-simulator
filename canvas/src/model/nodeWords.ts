import type { ViewNode } from "./plantView";
import type { NodeStatus } from "./readout";
import { statusWords } from "./readout";

function counted(n: number, noun: string): string {
  return `${String(n)} ${noun}${n === 1 ? "" : "s"}`;
}

function stateOf(status: NodeStatus, count: number, fault: string): string {
  if (status === null) return "";
  const where = count > 0 ? ` in ${String(count)}` : "";
  if (fault === "") return `${statusWords[status]}${where}, `;
  return status === "fault" ? `${statusWords.fault}: ${fault}, ` : `${statusWords[status]}${where}, Fault: ${fault}, `;
}

export function describeNode(
  node: ViewNode,
  selected: boolean,
  status: NodeStatus = null,
  count = 0,
  fault = "",
  effector = "",
): string {
  const ports =
    node.instance === null
      ? "Space selects, Enter opens the hall"
      : `${counted(node.pins.ins.length, "input")}, ${counted(node.pins.outs.length, "output")}`;
  const state = stateOf(status, count, fault);
  const call = effector === "" ? "" : `, effector call: ${effector}`;
  return `${node.label}, ${state}${node.detail}${call}, ${ports}${selected ? ", selected" : ""}`;
}

import type { NodeProps } from "@xyflow/react";
import { Handle, Position } from "@xyflow/react";
import { ChevronRight } from "lucide-react";
import { describeNode } from "../../../model/graph";
import { cx } from "../../primitives";
import type { HallNodeType } from "./flow";
import { StatusTag } from "./Marks";
import { shownOf, statusClass } from "./shown";
import styles from "./nodes.module.css";
import pins from "./pins.module.css";

export function HallNode({ data, selected }: NodeProps<HallNodeType>) {
  const { view } = data;
  const { status, count } = shownOf(data);
  return (
    <div
      className={cx(
        styles.node,
        styles.hall,
        selected && styles.selected,
        data.active && styles.active,
        statusClass(status),
      )}
      data-node-hall="true"
      data-active={data.active ? "true" : undefined}
      role="group"
      aria-label={describeNode(view, { selected, status, count })}
    >
      <Handle type="target" position={Position.Left} id="in" className={pins.plain} isConnectable={false} />
      <div className={styles.header}>
        <span className={styles.id}>{view.label}</span>
        <StatusTag status={status} count={count} />
        <ChevronRight size={14} aria-hidden="true" className={styles.opens} />
      </div>
      <div className={styles.members} title="Double-click, or press Enter, to open the hall">
        {view.detail}
      </div>
      <Handle type="source" position={Position.Right} id="out" className={pins.plain} isConnectable={false} />
    </div>
  );
}

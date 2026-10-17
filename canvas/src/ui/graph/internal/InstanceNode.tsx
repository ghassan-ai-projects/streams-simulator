import type { NodeProps } from "@xyflow/react";
import { Handle, Position, useStore } from "@xyflow/react";
import type { CSSProperties } from "react";
import { describeNode, headerHeight, idExtra } from "../../../model/graph";
import type { ViewNode } from "../../../model/graph";
import { cx } from "../../primitives";
import type { InstanceNodeType } from "./flow";
import { InstanceHeader } from "./InstanceHeader";
import { Pins } from "./Pins";
import { Readout } from "./Readout";
import { shownOf, statusClass } from "./shown";
import styles from "./nodes.module.css";
import pins from "./pins.module.css";

function instanceStyle(view: ViewNode): CSSProperties {
  return {
    "--cat": `var(--cat-${view.category})`,
    "--id-extra": `${String(idExtra(view.label))}px`,
    "--header-height": `${String(headerHeight(view.label))}px`,
    "--note-lines": String(view.noteLines),
  } as CSSProperties;
}

function instanceLabel(data: InstanceNodeType["data"], selected: boolean): string {
  const { status, fault, effector } = shownOf(data);
  return describeNode(data.view, { selected, status, fault, effector }) + (data.flagged ? ", has issues" : "");
}

export function InstanceNode({ data, selected }: NodeProps<InstanceNodeType>) {
  const { view } = data;
  const { status } = shownOf(data);
  const zoom = useStore((state) => state.transform[2]);
  return (
    <div
      className={cx(
        styles.node,
        selected && styles.selected,
        data.active && styles.active,
        data.flagged && styles.flagged,
        statusClass(status),
      )}
      data-node-root="true"
      data-active={data.active ? "true" : undefined}
      role="group"
      aria-label={instanceLabel(data, selected)}
      style={instanceStyle(view)}
    >
      <Handle type="target" position={Position.Left} id="in" className={pins.anchor} isConnectable={false} />
      <Handle type="source" position={Position.Right} id="out" className={pins.anchor} isConnectable={false} />
      <InstanceHeader data={data} status={status} zoom={zoom} />
      {data.live !== null && <Readout body={data.live.readout} fault={data.live.fault} effector={data.live.effector} />}
      {data.live === null && <Pins ins={view.pins.ins} outs={view.pins.outs} editable={data.editable} zoom={zoom} />}
    </div>
  );
}

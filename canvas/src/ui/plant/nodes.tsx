import type { NodeProps } from "@xyflow/react";
import { Handle, Position } from "@xyflow/react";
import { ChevronRight, OctagonAlert, TriangleAlert, Wrench } from "lucide-react";
import type { CSSProperties } from "react";
import { describeNode } from "../../model/plantView";
import { familyNames, shapeOf } from "../../model/ports";
import type { Pin } from "../../model/ports";
import type { NodeStatus, Shown } from "../../model/readout";
import { statusWords } from "../../model/readout";
import type { FrameNodeType, HallNodeType, InstanceNodeType } from "./flow";
import { cx } from "../cx";
import styles from "./nodes.module.css";

function pinLook(pin: Pin): { readonly family: string; readonly shape: string; readonly typeName: string } {
  if (pin.family === null) {
    return { family: "none", shape: "circle", typeName: "untyped" };
  }
  return { family: pin.family, shape: shapeOf(pin.family), typeName: familyNames[pin.family] };
}

const connectionWords = {
  connected: "connected",
  open: "not connected",
  unknown: "link without a recorded port",
} as const;

function PinRow({ pin, side }: { readonly pin: Pin; readonly side: "in" | "out" }) {
  const look = pinLook(pin);
  return (
    <div
      className={cx(styles.pin, styles[side])}
      title={`${pin.name} · ${pin.type || look.typeName} · ${look.typeName}`}
    >
      <Handle
        type={side === "in" ? "target" : "source"}
        position={side === "in" ? Position.Left : Position.Right}
        id={pin.name}
        className={cx(styles.dot, styles[look.shape], styles[pin.connection], `family-${look.family}`)}
        isConnectable={false}
      />
      <span className={styles.pinName}>{pin.name}</span>
      <span className="sr-only">{`, ${look.typeName}, ${connectionWords[pin.connection]}`}</span>
    </div>
  );
}

function Pins({ ins, outs }: { readonly ins: readonly Pin[]; readonly outs: readonly Pin[] }) {
  if (ins.length + outs.length === 0) {
    return <div className={styles.none}>no port data</div>;
  }
  return (
    <div className={styles.pins}>
      <div>
        {ins.map((pin) => (
          <PinRow key={pin.name} pin={pin} side="in" />
        ))}
      </div>
      <div>
        {outs.map((pin) => (
          <PinRow key={pin.name} pin={pin} side="out" />
        ))}
      </div>
    </div>
  );
}

export const readoutHeight = 84;

function Readout({ shown }: { readonly shown: readonly Shown[] }) {
  if (shown.length === 0) {
    return (
      <div className={styles.readout}>
        <span className={styles.none}>no observation yet</span>
      </div>
    );
  }
  const more = shown.length - 4;
  return (
    <dl className={styles.readout}>
      {shown.slice(0, more > 0 ? 3 : 4).map((entry) => (
        <div key={entry.channel} className={entry.attention === null ? undefined : styles[entry.attention]}>
          <dt>{entry.channel}</dt>
          <dd>
            {entry.value}
            {entry.unit !== "" && <span className={styles.unit}>{entry.unit}</span>}
          </dd>
        </div>
      ))}
      {more > 0 && (
        <div className={styles.more}>
          <dt>{`+${String(more + 1)} more`}</dt>
          <dd>in Inspect</dd>
        </div>
      )}
    </dl>
  );
}

const statusIcons = { alarm: OctagonAlert, fault: Wrench, warn: TriangleAlert } as const;

function StatusTag({ status }: { readonly status: NodeStatus }) {
  if (status === null) return null;
  const Icon = statusIcons[status];
  return (
    <span className={cx(styles.tag, styles[`tag-${status}`])} aria-hidden="true">
      <Icon size={12} />
      {statusWords[status]}
    </span>
  );
}

function statusClass(status: NodeStatus): string | undefined {
  return status === null ? undefined : styles[`status-${status}`];
}

export function InstanceNode({ data, selected }: NodeProps<InstanceNodeType>) {
  const { view } = data;
  const status = data.live?.status ?? null;
  return (
    <div
      className={cx(styles.node, selected && styles.selected, data.active && styles.active, statusClass(status))}
      role="group"
      aria-label={describeNode(view, selected, status)}
      style={{ "--cat": `var(--cat-${view.category})` } as CSSProperties}
    >
      <Handle type="target" position={Position.Left} id="in" className={styles.anchor} isConnectable={false} />
      <Handle type="source" position={Position.Right} id="out" className={styles.anchor} isConnectable={false} />
      <div className={styles.header}>
        <span className={styles.id}>{view.label}</span>
        <StatusTag status={status} />
        <span className={styles.kind} title={view.detail}>
          {view.detail}
        </span>
      </div>
      {data.live !== null && <Readout shown={data.live.shown} />}
      <Pins ins={view.pins.ins} outs={view.pins.outs} />
    </div>
  );
}

export function HallNode({ data, selected }: NodeProps<HallNodeType>) {
  const { view } = data;
  return (
    <div
      className={cx(styles.node, styles.hall, selected && styles.selected, data.active && styles.active)}
      role="group"
      aria-label={describeNode(view, selected)}
    >
      <Handle type="target" position={Position.Left} id="in" className={styles.plain} isConnectable={false} />
      <div className={styles.header}>
        <span className={styles.id}>{view.label}</span>
        <ChevronRight size={14} aria-hidden="true" className={styles.open} />
      </div>
      <div className={styles.members}>{view.detail}</div>
      <Handle type="source" position={Position.Right} id="out" className={styles.plain} isConnectable={false} />
    </div>
  );
}

export function FrameNode({ data }: NodeProps<FrameNodeType>) {
  return (
    <div className={styles.frame}>
      <div className={styles.frameTitle}>
        {data.title}
        <span className={styles.frameCount}>{`${String(data.count)} ${data.count === 1 ? "node" : "nodes"}`}</span>
      </div>
    </div>
  );
}

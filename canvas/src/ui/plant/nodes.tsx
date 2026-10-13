import type { NodeProps } from "@xyflow/react";
import { Handle, Position } from "@xyflow/react";
import { ChevronRight } from "lucide-react";
import type { Reading } from "../../model/playback";
import { formatReading, shortChannel } from "../../model/playback";
import { familyNames, shapeOf } from "../../model/ports";
import type { Pin } from "../../model/ports";
import type { FrameNodeType, HallNodeType, InstanceNodeType } from "./flow";
import { cx } from "../cx";
import styles from "./nodes.module.css";

function pinLook(pin: Pin): { readonly family: string; readonly shape: string; readonly typeName: string } {
  if (pin.family === null) {
    return { family: "none", shape: "circle", typeName: "untyped" };
  }
  return { family: pin.family, shape: shapeOf(pin.family), typeName: familyNames[pin.family] };
}

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
        className={cx(styles.dot, styles[look.shape], pin.connected && styles.connected, `family-${look.family}`)}
        isConnectable={false}
      />
      <span className={styles.pinName}>{pin.name}</span>
      <span className="sr-only">{`, ${look.typeName}, ${pin.connected ? "connected" : "not connected"}`}</span>
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

function Readout({ readings }: { readonly readings: readonly Reading[] }) {
  if (readings.length === 0) {
    return (
      <div className={styles.readout}>
        <span className={styles.none}>no observation yet</span>
      </div>
    );
  }
  return (
    <dl className={styles.readout}>
      {readings.slice(0, 4).map((reading) => (
        <div key={reading.channel}>
          <dt>{shortChannel(reading.channel)}</dt>
          <dd>{formatReading(reading)}</dd>
        </div>
      ))}
    </dl>
  );
}

export function InstanceNode({ data, selected }: NodeProps<InstanceNodeType>) {
  const { view } = data;
  return (
    <div className={cx(styles.node, selected && styles.selected)}>
      <Handle type="target" position={Position.Left} id="in" className={styles.anchor} isConnectable={false} />
      <Handle type="source" position={Position.Right} id="out" className={styles.anchor} isConnectable={false} />
      <div className={styles.header}>
        <span className={styles.id}>{view.label}</span>
        <span className={styles.kind} title={view.detail}>
          {view.detail}
        </span>
      </div>
      {data.readings !== null && <Readout readings={data.readings} />}
      <Pins ins={view.pins.ins} outs={view.pins.outs} />
    </div>
  );
}

export function HallNode({ data, selected }: NodeProps<HallNodeType>) {
  const { view } = data;
  return (
    <div className={cx(styles.node, styles.hall, selected && styles.selected)}>
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

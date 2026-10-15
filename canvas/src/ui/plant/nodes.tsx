import type { NodeProps } from "@xyflow/react";
import { Handle, Position, useStore } from "@xyflow/react";
import { ChevronRight, OctagonAlert, TriangleAlert, Wrench } from "lucide-react";
import type { CSSProperties } from "react";
import { useRef } from "react";
import { countWords } from "../../model/frameTitle";
import type { ViewNode } from "../../model/plantView";
import { describeNode, headerHeight, idExtra } from "../../model/plantView";
import { familyNames, shapeOf } from "../../model/ports";
import type { Pin } from "../../model/ports";
import type { NodeStatus } from "../../model/readout";
import { channelWords, statusWords } from "../../model/readout";
import type { FrameTitleNodeType, HallNodeType, InstanceNodeType } from "./flow";
import { cx } from "../cx";
import { Readout } from "./Readout";
import { useLabelFit } from "./labelFit";
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

function PinRow({
  pin,
  side,
  editable,
  zoom,
}: {
  readonly pin: Pin;
  readonly side: "in" | "out";
  readonly editable: boolean;
  readonly zoom: number;
}) {
  const look = pinLook(pin);
  const label = useRef<HTMLSpanElement>(null);
  useLabelFit(label, zoom);
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
        isConnectable={editable}
      />
      <span ref={label} className={styles.pinName}>
        {channelWords(pin.name)}
      </span>
      <span className="sr-only">{`, ${look.typeName}, ${connectionWords[pin.connection]}`}</span>
    </div>
  );
}

function Pins({
  ins,
  outs,
  editable,
  zoom,
}: {
  readonly ins: readonly Pin[];
  readonly outs: readonly Pin[];
  readonly editable: boolean;
  readonly zoom: number;
}) {
  if (ins.length + outs.length === 0) {
    return <div className={styles.none}>no port data</div>;
  }
  return (
    <div className={styles.pins}>
      <div>
        {ins.map((pin) => (
          <PinRow key={pin.name} pin={pin} side="in" editable={editable} zoom={zoom} />
        ))}
      </div>
      <div>
        {outs.map((pin) => (
          <PinRow key={pin.name} pin={pin} side="out" editable={editable} zoom={zoom} />
        ))}
      </div>
    </div>
  );
}

const statusIcons = { alarm: OctagonAlert, fault: Wrench, warn: TriangleAlert } as const;

function StatusTag({ status, count = 0 }: { readonly status: NodeStatus; readonly count?: number }) {
  if (status === null) return null;
  const Icon = statusIcons[status];
  const word = count > 0 ? `${statusWords[status]} ${String(count)}` : statusWords[status];
  return (
    <span className={cx(styles.tag, styles[`tag-${status}`])} aria-hidden="true">
      <Icon size={12} />
      {word}
    </span>
  );
}

function EffectorTag({ effector }: { readonly effector: string }) {
  if (effector === "") return null;
  return (
    <span className={styles.effectorTag} title={`Effector call · ${effector}`}>
      <span className={styles.effectorMark} aria-hidden="true" />
      <span className="sr-only">{`Effector call: ${effector}`}</span>
    </span>
  );
}

function issueWords(flagged: boolean): string {
  return flagged ? ", has issues" : "";
}

function IssueTag({ flagged }: { readonly flagged: boolean }) {
  if (!flagged) return null;
  return (
    <span className={cx(styles.tag, styles["tag-alarm"])} aria-hidden="true">
      <OctagonAlert size={12} />
      Issue
    </span>
  );
}

function statusClass(status: NodeStatus): string | undefined {
  return status === null ? undefined : styles[`status-${status}`];
}

function instanceStyle(view: ViewNode): CSSProperties {
  return {
    "--cat": `var(--cat-${view.category})`,
    "--id-extra": `${String(idExtra(view.label))}px`,
    "--header-height": `${String(headerHeight(view.label))}px`,
  } as CSSProperties;
}

function InstanceHeader({ data, status }: { readonly data: InstanceNodeType["data"]; readonly status: NodeStatus }) {
  const { view } = data;
  return (
    <div className={styles.header}>
      <span className={styles.id} title={view.label}>
        {view.label}
      </span>
      <StatusTag status={status} />
      <EffectorTag effector={data.live?.effector ?? ""} />
      <IssueTag flagged={data.flagged} />
      <span className={styles.kind} title={view.detail}>
        {view.detail}
      </span>
    </div>
  );
}

function instanceLabel(data: InstanceNodeType["data"], selected: boolean): string {
  const { view, live } = data;
  const words = describeNode(view, selected, live?.status ?? null, 0, live?.fault ?? "", live?.effector ?? "");
  return words + issueWords(data.flagged);
}

export function InstanceNode({ data, selected }: NodeProps<InstanceNodeType>) {
  const { view } = data;
  const status = data.live?.status ?? null;
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
      role="group"
      aria-label={instanceLabel(data, selected)}
      style={instanceStyle(view)}
    >
      <Handle type="target" position={Position.Left} id="in" className={styles.anchor} isConnectable={false} />
      <Handle type="source" position={Position.Right} id="out" className={styles.anchor} isConnectable={false} />
      <InstanceHeader data={data} status={status} />
      {data.live !== null && (
        <Readout
          shown={data.live.shown}
          pins={view.pins}
          channels={view.channels}
          fault={data.live.fault}
          effector={data.live.effector}
        />
      )}
      {data.live === null && <Pins ins={view.pins.ins} outs={view.pins.outs} editable={data.editable} zoom={zoom} />}
    </div>
  );
}

export function HallNode({ data, selected }: NodeProps<HallNodeType>) {
  const { view } = data;
  const status = data.live?.status ?? null;
  const count = data.live?.count ?? 0;
  return (
    <div
      className={cx(
        styles.node,
        styles.hall,
        selected && styles.selected,
        data.active && styles.active,
        statusClass(status),
      )}
      role="group"
      aria-label={describeNode(view, selected, status, count)}
    >
      <Handle type="target" position={Position.Left} id="in" className={styles.plain} isConnectable={false} />
      <div className={styles.header}>
        <span className={styles.id}>{view.label}</span>
        <StatusTag status={status} count={count} />
        <ChevronRight size={14} aria-hidden="true" className={styles.opens} />
      </div>
      <div className={styles.members} title="Double-click, or press Enter, to open the hall">
        {view.detail}
      </div>
      <Handle type="source" position={Position.Right} id="out" className={styles.plain} isConnectable={false} />
    </div>
  );
}

export function FrameNode() {
  return <div className={styles.frame} />;
}

export function FrameTitleNode({ data }: NodeProps<FrameTitleNodeType>) {
  return (
    <div className={styles.frameTitleNode}>
      <div className={styles.frameTitle}>
        {data.title}
        <span className={styles.frameCount}>{countWords(data.count)}</span>
      </div>
    </div>
  );
}

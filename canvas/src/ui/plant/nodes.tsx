import type { NodeProps } from "@xyflow/react";
import { Handle, Position, useStore } from "@xyflow/react";
import { ChevronRight } from "lucide-react";
import type { CSSProperties } from "react";
import { useRef } from "react";
import { countWords, splitName } from "../../model/frameTitle";
import { fitId, idCap, idRoom } from "../../model/headerFit";
import type { ViewNode } from "../../model/plantView";
import { describeNode } from "../../model/nodeWords";
import { headerHeight, idExtra } from "../../model/plantView";
import { familyNames, shapeOf } from "../../model/ports";
import type { Pin } from "../../model/ports";
import { tierOf } from "../../model/viewport";
import type { NodeStatus } from "../../model/readout";
import { pinWords } from "../../model/readout";
import type { PlanMark } from "../../model/planMarks";
import type { FrameTitleNodeType, HallNodeType, InstanceNodeType, NodeLive } from "./flow";
import { cx } from "../cx";
import { EffectorTag, IssueTag, StatusTag, tagWords } from "./Marks";
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
  const words = pinWords(pin.name, side);
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
        <span className={styles.pinFull}>{words.full}</span>
        <span className={styles.pinShort}>{words.short}</span>
      </span>
      <span className="sr-only">{`, ${side === "in" ? "input" : "output"}, ${look.typeName}, ${connectionWords[pin.connection]}`}</span>
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

function issueWords(flagged: boolean): string {
  return flagged ? ", has issues" : "";
}

function statusClass(status: NodeStatus): string | undefined {
  return status === null ? undefined : styles[`status-${status}`];
}

function instanceStyle(view: ViewNode): CSSProperties {
  return {
    "--cat": `var(--cat-${view.category})`,
    "--id-extra": `${String(idExtra(view.label))}px`,
    "--header-height": `${String(headerHeight(view.label))}px`,
    "--note-lines": String(view.noteLines),
  } as CSSProperties;
}

function markWords(data: InstanceNodeType["data"], status: NodeStatus): number | null {
  const { effector } = shownOf(data);
  if (status === null && effector === "" && !data.flagged) return null;
  const issue = data.flagged ? 6 : 0;
  if (status === null || data.live !== null) return issue;
  return issue + tagWords(status, 0, data.plan?.fault ?? "").length;
}

function IdText({
  label,
  zoom,
  words,
}: {
  readonly label: string;
  readonly zoom: number;
  readonly words: number | null;
}) {
  const tier = tierOf(zoom);
  if (tier !== "compact" && tier !== "overview") return label;
  const fitted = fitId(label, idRoom(zoom, words, tier), zoom, idCap[tier]);
  if (fitted.text === label) return <span style={{ fontSize: `${String(fitted.size)}px` }}>{label}</span>;
  return (
    <span style={{ fontSize: `${String(fitted.size)}px` }}>
      <span aria-hidden="true">{fitted.text}</span>
      <span className="sr-only">{label}</span>
    </span>
  );
}

function InstanceHeader({
  data,
  status,
  zoom,
}: {
  readonly data: InstanceNodeType["data"];
  readonly status: NodeStatus;
  readonly zoom: number;
}) {
  const { view } = data;
  const planned = data.live === null ? data.plan : null;
  const words = markWords(data, status);
  return (
    <div className={styles.header}>
      <span className={styles.id} title={view.label}>
        <IdText label={view.label} zoom={zoom} words={words} />
      </span>
      {words !== null && (
        <span className={styles.marks}>
          <StatusTag status={status} cause={planned?.fault ?? ""} brief={data.live !== null} />
          {data.live !== null && data.live.fault !== "" && (
            <span className="sr-only">{`Fault: ${data.live.fault}`}</span>
          )}
          <EffectorTag effector={shownOf(data).effector} />
          <IssueTag flagged={data.flagged} />
        </span>
      )}
      <span className={styles.kind} title={view.detail}>
        {view.detail}
      </span>
    </div>
  );
}

interface Shown {
  readonly status: NodeStatus;
  readonly count: number;
  readonly fault: string;
  readonly effector: string;
}

function shownOf(data: { readonly live: NodeLive | null; readonly plan: PlanMark | null }): Shown {
  const { live, plan } = data;
  if (live !== null) return live;
  return plan ?? { status: null, count: 0, fault: "", effector: "" };
}

function instanceLabel(data: InstanceNodeType["data"], selected: boolean): string {
  const { status, fault, effector } = shownOf(data);
  return describeNode(data.view, selected, status, 0, fault, effector) + issueWords(data.flagged);
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
      role="group"
      aria-label={instanceLabel(data, selected)}
      style={instanceStyle(view)}
    >
      <Handle type="target" position={Position.Left} id="in" className={styles.anchor} isConnectable={false} />
      <Handle type="source" position={Position.Right} id="out" className={styles.anchor} isConnectable={false} />
      <InstanceHeader data={data} status={status} zoom={zoom} />
      {data.live !== null && (
        <Readout
          shown={data.live.shown}
          observed={data.live.observed}
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
  const count = countWords(data.count);
  const { head, tail } = splitName(data.title);
  return (
    <div className={styles.frameTitleNode}>
      <div className={styles.frameTitle} title={`${data.title} · ${count}`}>
        <span className={styles.frameName}>
          <span className={styles.nameHead}>{head}</span>
          <span className={styles.nameTail}>{tail}</span>
        </span>
        <span className={styles.frameCount}>{count}</span>
      </div>
    </div>
  );
}

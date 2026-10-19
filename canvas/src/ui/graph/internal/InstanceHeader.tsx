import type { NodeStatus } from "../../../model/graph";
import type { InstanceNodeType } from "./flow";
import { EffectorTag, InterruptTag, IssueTag, StatusTag, tagWords } from "./Marks";
import { IdText } from "./IdText";
import { OriginMark } from "./OriginMark";
import { shownOf } from "./shown";
import styles from "./nodes.module.css";

const originIconRoom = 2;
const interruptTagRoom = 8;

function originRoom(data: InstanceNodeType["data"]): number {
  const origin = (data.live?.origin ?? null) === null ? 0 : originIconRoom;
  return origin + ((data.live?.interrupt ?? null) === null ? 0 : interruptTagRoom);
}

function markWords(data: InstanceNodeType["data"], status: NodeStatus): number | null {
  const { effector } = shownOf(data);
  const origin = originRoom(data);
  if (status === null && effector === "" && !data.flagged && origin === 0) return null;
  const marks = (data.flagged ? 6 : 0) + origin;
  if (status === null || data.live !== null) return marks;
  return marks + tagWords(status, 0, data.plan?.fault ?? "").length;
}

function InstanceMarks({ data, status }: { readonly data: InstanceNodeType["data"]; readonly status: NodeStatus }) {
  const planned = data.live === null ? data.plan : null;
  return (
    <span>
      <span className={styles.origin}>
        <OriginMark origin={data.live?.origin ?? null} voice="silent" />
      </span>
      <InterruptTag interrupt={data.live?.interrupt ?? null} />
      <StatusTag status={status} cause={planned?.fault ?? ""} brief={data.live !== null} />
      {data.live !== null && data.live.fault !== "" && <span className="sr-only">{`Fault: ${data.live.fault}`}</span>}
      <EffectorTag effector={shownOf(data).effector} />
      <IssueTag flagged={data.flagged} />
    </span>
  );
}

export function InstanceHeader({
  data,
  status,
  zoom,
}: {
  readonly data: InstanceNodeType["data"];
  readonly status: NodeStatus;
  readonly zoom: number;
}) {
  const { view } = data;
  const words = markWords(data, status);
  return (
    <div className={styles.header}>
      <span className={styles.id} data-part="id" title={view.label}>
        <IdText label={view.label} zoom={zoom} words={words} />
      </span>
      {words !== null && <InstanceMarks data={data} status={status} />}
      <span className={styles.kind} data-part="kind" title={view.detail}>
        {view.detail}
      </span>
    </div>
  );
}

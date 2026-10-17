import type { NodeStatus } from "../../../model/graph";
import type { InstanceNodeType } from "./flow";
import { EffectorTag, IssueTag, StatusTag, tagWords } from "./Marks";
import { IdText } from "./IdText";
import { shownOf } from "./shown";
import styles from "./nodes.module.css";

function markWords(data: InstanceNodeType["data"], status: NodeStatus): number | null {
  const { effector } = shownOf(data);
  if (status === null && effector === "" && !data.flagged) return null;
  const issue = data.flagged ? 6 : 0;
  if (status === null || data.live !== null) return issue;
  return issue + tagWords(status, 0, data.plan?.fault ?? "").length;
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
  const planned = data.live === null ? data.plan : null;
  const words = markWords(data, status);
  return (
    <div className={styles.header}>
      <span className={styles.id} data-part="id" title={view.label}>
        <IdText label={view.label} zoom={zoom} words={words} />
      </span>
      {words !== null && (
        <span>
          <StatusTag status={status} cause={planned?.fault ?? ""} brief={data.live !== null} />
          {data.live !== null && data.live.fault !== "" && (
            <span className="sr-only">{`Fault: ${data.live.fault}`}</span>
          )}
          <EffectorTag effector={shownOf(data).effector} />
          <IssueTag flagged={data.flagged} />
        </span>
      )}
      <span className={styles.kind} data-part="kind" title={view.detail}>
        {view.detail}
      </span>
    </div>
  );
}

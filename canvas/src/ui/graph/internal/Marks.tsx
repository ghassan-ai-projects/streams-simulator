import { OctagonAlert, TriangleAlert, Wrench } from "lucide-react";
import type { NodeStatus } from "../../../model/graph";
import { statusWords } from "../../../model/graph";
import { cx } from "../../cx";
import styles from "./marks.module.css";

const statusIcons = { alarm: OctagonAlert, fault: Wrench, warn: TriangleAlert } as const;

export function tagWords(status: NonNullable<NodeStatus>, count: number, cause: string): string {
  const word = count > 0 ? `${statusWords[status]} ${String(count)}` : statusWords[status];
  return cause === "" ? word : `${word} · ${cause}`;
}

export function StatusTag({
  status,
  count = 0,
  cause = "",
  brief = false,
}: {
  readonly status: NodeStatus;
  readonly count?: number;
  readonly cause?: string;
  readonly brief?: boolean;
}) {
  if (status === null) return null;
  const Icon = statusIcons[status];
  const words = tagWords(status, count, cause);
  return (
    <span className={cx(styles.tag, styles[`tag-${status}`], brief && styles.brief)} aria-hidden="true" title={words}>
      <Icon size={12} />
      <span className={styles.tagWords}>{words}</span>
    </span>
  );
}

export function EffectorTag({ effector }: { readonly effector: string }) {
  if (effector === "") return null;
  return (
    <span className={styles.effectorTag} title={`Effector call · ${effector}`}>
      <span className={styles.effectorMark} aria-hidden="true" />
      <span className="sr-only">{`Effector call: ${effector}`}</span>
    </span>
  );
}

export function IssueTag({ flagged }: { readonly flagged: boolean }) {
  if (!flagged) return null;
  return (
    <span className={cx(styles.tag, styles["tag-alarm"])} aria-hidden="true">
      <OctagonAlert size={12} />
      Issue
    </span>
  );
}

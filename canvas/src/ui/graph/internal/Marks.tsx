import { ClipboardCheck, Hand, OctagonAlert, OctagonX, TriangleAlert, Wrench } from "lucide-react";
import type { NodeInterrupt, NodeStatus } from "../../../model/graph";
import { statusWords } from "../../../model/graph";
import { cx } from "../../primitives";
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

const interruptIcons = { hold: Hand, approve: ClipboardCheck } as const;

function interruptWord(interrupt: NodeInterrupt): string {
  if (interrupt.latched) return "Latched";
  return interrupt.mode === "hold" ? "Held" : "Approving";
}

export function InterruptTag({ interrupt }: { readonly interrupt: NodeInterrupt | null }) {
  if (interrupt === null) return null;
  const Icon = interrupt.latched ? OctagonX : interruptIcons[interrupt.mode];
  return (
    <span
      className={cx(styles.tag, interrupt.latched ? styles["tag-alarm"] : styles["tag-warn"])}
      aria-hidden="true"
      title={interrupt.words}
      data-interrupt={interrupt.latched ? "latched" : interrupt.mode}
    >
      <Icon size={12} />
      <span className={styles.tagWords}>{interruptWord(interrupt)}</span>
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

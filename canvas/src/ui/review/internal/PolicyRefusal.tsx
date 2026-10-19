import { useId, useState } from "react";
import type { ConnectSession } from "../../../app/launch";
import type { ReviewSession } from "../../../app/review";
import type { SessionCommand } from "../../../model/review";
import { copiedWords, copyFailedWords, digestListedWords } from "../../../model/review";
import { shortDigest } from "../../../model/shared/words";
import { cx, shared } from "../../primitives";
import styles from "./SessionEvidence.module.css";

export interface PolicyActions {
  readonly review: ReviewSession;
  readonly connect: ConnectSession;
  readonly domain: string;
}

interface PolicyRefusalProps extends PolicyActions {
  readonly command: SessionCommand;
}

interface Feedback {
  readonly status: string;
  readonly full: boolean;
}

const quiet: Feedback = { status: "", full: false };

function DigestLine({ id, digest }: { readonly id: string; readonly digest: string }) {
  return (
    <span id={id}>
      {"Host's policy "}
      <code className={styles.digest} aria-hidden="true" title={digest}>
        {shortDigest(digest)}
      </code>
      <span className="sr-only">{digest}</span>
    </span>
  );
}

interface ActionProps {
  readonly describedBy: string;
  readonly label: string;
  readonly onPress: () => void;
}

function Action({ describedBy, label, onPress }: ActionProps) {
  return (
    <button
      type="button"
      className={cx(shared.secondary, styles.action)}
      aria-describedby={describedBy}
      onClick={onPress}
    >
      {label}
    </button>
  );
}

export function PolicyRefusal({ command, review, connect, domain }: PolicyRefusalProps) {
  const line = useId();
  const [feedback, setFeedback] = useState<Feedback>(quiet);
  const digest = command.policyDigest;
  const copy = async () => {
    const copied = await review.copyText(digest);
    setFeedback(
      copied ? { status: copiedWords(command.commandId), full: false } : { status: copyFailedWords, full: true },
    );
  };
  const use = () => {
    connect.addPolicy(domain, digest);
    setFeedback({ status: digestListedWords(domain), full: false });
  };
  return (
    <li className={styles.stage}>
      <span className={styles.stageName}>Policy</span>
      <DigestLine id={line} digest={digest} />
      <Action
        describedBy={line}
        label="Copy policy digest"
        onPress={() => {
          void copy();
        }}
      />
      {domain !== "" && <Action describedBy={line} label="Use this digest" onPress={use} />}
      {feedback.full && <code className={styles.fullDigest}>{digest}</code>}
      <span role="status" className={feedback.status === "" ? "sr-only" : styles.detail}>
        {feedback.status}
      </span>
    </li>
  );
}

import { panel } from "../../primitives";
import type { ReactNode } from "react";
import styles from "./Editor.module.css";
import { Reason } from "./Reason";

interface SubmitRowProps {
  readonly id: string;
  readonly reason: string | null;
  readonly children: ReactNode;
}

export function SubmitRow({ id, reason, children }: SubmitRowProps) {
  return (
    <div className={styles.submitRow}>
      <button
        type="submit"
        className={panel.secondary}
        aria-disabled={reason !== null}
        aria-describedby={reason === null ? undefined : id}
      >
        {children}
      </button>
      <Reason id={id} text={reason} />
    </div>
  );
}

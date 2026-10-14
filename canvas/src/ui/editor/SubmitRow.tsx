import type { ReactNode } from "react";
import panel from "../panel.module.css";
import styles from "./Editor.module.css";

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
      {reason !== null && (
        <p className={styles.reason} id={id}>
          {reason}
        </p>
      )}
    </div>
  );
}

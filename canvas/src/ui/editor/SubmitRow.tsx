import type { ReactNode } from "react";
import panel from "../panel.module.css";

interface SubmitRowProps {
  readonly id: string;
  readonly reason: string | null;
  readonly children: ReactNode;
}

export function SubmitRow({ id, reason, children }: SubmitRowProps) {
  return (
    <>
      {reason !== null && (
        <p className={panel.hint} id={id}>
          {reason}
        </p>
      )}
      <button
        type="submit"
        className={panel.secondary}
        aria-disabled={reason !== null}
        aria-describedby={reason === null ? undefined : id}
      >
        {children}
      </button>
    </>
  );
}

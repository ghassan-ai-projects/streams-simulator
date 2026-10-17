import type { ReactNode } from "react";
import styles from "./Shell.module.css";

interface ShellProps {
  readonly topBar: ReactNode;
  readonly children: ReactNode;
  readonly left: boolean;
  readonly right: boolean;
}

export function Shell({ topBar, children, left, right }: ShellProps) {
  return (
    <div className={styles.shell} data-left={left ? "open" : "closed"} data-right={right ? "open" : "closed"}>
      {topBar}
      <main id="main" className={styles.main} aria-label="Workspace" tabIndex={-1}>
        {children}
      </main>
    </div>
  );
}

import type { ReactNode } from "react";
import styles from "./Shell.module.css";

interface ShellProps {
  readonly topBar: ReactNode;
  readonly sidebar: ReactNode;
  readonly children: ReactNode;
}

export function Shell({ topBar, sidebar, children }: ShellProps) {
  return (
    <div className={styles.shell}>
      {topBar}
      {sidebar}
      <main className={styles.main} aria-label="Workspace">
        {children}
      </main>
    </div>
  );
}

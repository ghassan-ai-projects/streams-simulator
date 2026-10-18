import type { CaptureSession } from "../../../app/devices";
import type { Workspace } from "../../../model/workspace";
import { runRows } from "../../../model/workspace";
import styles from "./Capture.module.css";
import { CaptureRun } from "./CaptureRun";
import { CaptureText } from "./CaptureText";

type Source = Parameters<CaptureSession["read"]>[0];

interface SourceProps {
  readonly workspace: Workspace;
  readonly busy: boolean;
  readonly onRead: (source: Source) => void;
}

export function CaptureSource({ workspace, busy, onRead }: SourceProps) {
  return (
    <section className={styles.panel} aria-labelledby="capture-source">
      <h2 id="capture-source" className={styles.panelTitle}>
        Where is the board's announcement?
      </h2>
      <p className={styles.lead}>
        A board announces the devices it carries when it connects. Paste that announcement or load it from a file, or
        read it from a run that was recorded with a board.
      </p>
      <CaptureText
        busy={busy}
        onRead={(content) => {
          onRead({ content });
        }}
      />
      <p className={styles.note}>or</p>
      <CaptureRun
        runs={runRows(workspace, null)}
        busy={busy}
        onRead={(run) => {
          onRead({ run });
        }}
      />
    </section>
  );
}

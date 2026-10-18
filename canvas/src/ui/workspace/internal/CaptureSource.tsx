import { useEffect, useRef, useState } from "react";
import type { CaptureSession } from "../../../app/devices";
import { sourceWords } from "../../../model/devices";
import type { Workspace } from "../../../model/workspace";
import { runRows } from "../../../model/workspace";
import { shared } from "../../primitives";
import styles from "./Capture.module.css";
import { CaptureRun } from "./CaptureRun";
import { CaptureText } from "./CaptureText";

type Source = Parameters<CaptureSession["read"]>[0];

interface SourceProps {
  readonly workspace: Workspace;
  readonly busy: boolean;
  readonly read: Source | null;
  readonly onRead: (source: Source) => void;
  readonly onClear: () => void;
}

function Form({ workspace, busy, onRead, onClear }: Omit<SourceProps, "read">) {
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
        onClear={onClear}
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

function Collapsed({ read, onOpen }: { readonly read: Source; readonly onOpen: () => void }) {
  return (
    <div className={styles.collapsed}>
      <span className={styles.facts}>Read from {sourceWords(read)}.</span>
      <button type="button" className={shared.secondary} onClick={onOpen}>
        Read another announcement
      </button>
    </div>
  );
}

export function CaptureSource({ read, ...form }: SourceProps) {
  const [open, setOpen] = useState(false);
  const shell = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (open) shell.current?.querySelector("textarea")?.focus();
  }, [open]);
  return (
    <div ref={shell}>
      {read === null || open ? (
        <Form
          {...form}
          onRead={(source) => {
            setOpen(false);
            form.onRead(source);
          }}
        />
      ) : (
        <Collapsed
          read={read}
          onOpen={() => {
            setOpen(true);
          }}
        />
      )}
    </div>
  );
}

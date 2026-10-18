import type { SyntheticEvent } from "react";
import { useId, useState } from "react";
import type { RunRow } from "../../../model/workspace";
import { whenOf } from "../../../model/workspace";
import { shared } from "../../primitives";
import styles from "./Capture.module.css";

interface RunProps {
  readonly runs: readonly RunRow[];
  readonly busy: boolean;
  readonly onRead: (run: string) => void;
}

interface SelectProps {
  readonly id: string;
  readonly runs: readonly RunRow[];
  readonly run: string;
  readonly onPick: (run: string) => void;
}

function label({ run, plantTitle }: RunRow): string {
  return [run.worldId, plantTitle, whenOf(run.recordedAt)].filter((part) => part !== "").join(" · ");
}

function RunSelect({ id, runs, run, onPick }: SelectProps) {
  return (
    <div className={styles.field}>
      <label htmlFor={`${id}-run`}>Recorded run</label>
      <select
        id={`${id}-run`}
        className={styles.select}
        value={run}
        onChange={(event) => {
          onPick(event.target.value);
        }}
      >
        {runs.map((row) => (
          <option key={row.run.worldId} value={row.run.worldId}>
            {label(row)}
          </option>
        ))}
      </select>
    </div>
  );
}

export function CaptureRun({ runs, busy, onRead }: RunProps) {
  const [picked, setPicked] = useState("");
  const id = useId();
  const run = runs.some((row) => row.run.worldId === picked) ? picked : (runs[0]?.run.worldId ?? "");
  if (runs.length === 0) {
    return (
      <p className={styles.note}>No recorded runs yet. A run recorded with a board holds that board's announcement.</p>
    );
  }
  const submit = (event: SyntheticEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!busy) onRead(run);
  };
  return (
    <form className={styles.form} aria-label="Read a recorded run" onSubmit={submit}>
      <RunSelect id={id} runs={runs} run={run} onPick={setPicked} />
      <span className={styles.actions}>
        <button type="submit" className={shared.secondary} aria-disabled={busy}>
          Read this run
        </button>
      </span>
    </form>
  );
}

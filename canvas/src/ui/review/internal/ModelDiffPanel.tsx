import { useEffect } from "react";
import type { DiffSession } from "../../../app/workspace";
import { fetchedFor } from "../../../app/workspace";
import { diffView } from "../../../model/workspace";
import { failureLine, useStore } from "../../primitives";
import { ModelDiffView } from "./ModelDiffView";
import styles from "./ModelDiff.module.css";

interface PanelProps {
  readonly session: DiffSession;
  readonly a: string;
  readonly b: string;
}

export function ModelDiffPanel({ session, a, b }: PanelProps) {
  useEffect(() => {
    void session.load(a, b);
  }, [session, a, b]);
  const state = fetchedFor(useStore(session), `${a}|${b}`);
  return (
    <section aria-label="What changed in the model">
      {state.phase === "loading" && <p role="status">Comparing the two models…</p>}
      {state.phase === "failed" && (
        <p role="alert" className={styles.failure}>
          {failureLine(state.failure.diagnostic)}
        </p>
      )}
      {state.phase === "ready" && <ModelDiffView view={diffView(state.value)} />}
    </section>
  );
}

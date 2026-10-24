import { useEffect } from "react";
import type { AssumptionsSession } from "../../../app/workspace";
import { fetchedFor } from "../../../app/workspace";
import type { Assumptions } from "../../../model/workspace";
import { kindAssumptions, originWords } from "../../../model/workspace";
import { failureLine, useStore } from "../../primitives";
import { KindBlock } from "./AssumptionRows";
import styles from "./Provenance.module.css";
import rows from "./Assumptions.module.css";

interface PanelProps {
  readonly session: AssumptionsSession;
  readonly path: string;
  readonly twin: string | null;
}

function NotModeled({ doc }: { readonly doc: Assumptions }) {
  return (
    <section aria-label="Not modelled">
      <h3 className={styles.sectionTitle}>Not modelled</h3>
      <ul className={rows.limits}>
        {doc.notModeled.map((limit) => (
          <li key={`${limit.code}:${limit.scope}`}>
            {limit.sentence} <span className={rows.scope}>{limit.scope}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}

function Document({ doc }: { readonly doc: Assumptions }) {
  return (
    <>
      <p className={styles.muted}>{originWords(doc.origin)}</p>
      {doc.kinds.map((kind) => (
        <KindBlock key={kind.kind} kind={kindAssumptions(kind)} />
      ))}
      <NotModeled doc={doc} />
    </>
  );
}

export function AssumptionsPanel({ session, path, twin }: PanelProps) {
  useEffect(() => {
    void session.load(path, twin);
  }, [session, path, twin]);
  const state = fetchedFor(useStore(session), `${path}|${twin ?? ""}`);
  return (
    <section aria-label="Model assumptions">
      {state.phase === "loading" && <p role="status">Reading the assumptions of the model…</p>}
      {state.phase === "failed" && (
        <p role="alert" className={styles.refusal}>
          {failureLine(state.failure.diagnostic)}
        </p>
      )}
      {state.phase === "ready" && <Document doc={state.value} />}
    </section>
  );
}

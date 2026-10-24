import { useState } from "react";
import type { ProvenanceSession } from "../../../app/workspace";
import type { NavigationStore } from "../../../app/shell";
import type { PlantRoute } from "../../../model/workspace";
import { cx, failureLine, shared } from "../../primitives";
import styles from "./Provenance.module.css";

export interface NoticeProps {
  readonly route: PlantRoute;
  readonly provenance: ProvenanceSession;
  readonly navigation: NavigationStore;
  readonly onCopy: () => void;
}

function Recompose({ route, provenance }: Pick<NoticeProps, "route" | "provenance">) {
  const [note, setNote] = useState("");
  const [failed, setFailed] = useState(false);
  const { action } = route;
  if (action.kind !== "recompose") return null;
  const { composition, domain } = action;
  const compose = () => {
    setNote("Composing…");
    setFailed(false);
    void provenance.recompose(composition, domain).then((composed) => {
      setFailed(!composed.ok);
      setNote(composed.ok ? "Composed again." : failureLine(composed.failure.diagnostic));
    });
  };
  return (
    <>
      <button type="button" className={cx(shared.primary, styles.action)} onClick={compose}>
        {route.label}
      </button>
      {note !== "" && (
        <span role={failed ? "alert" : "status"} className={failed ? styles.refusal : styles.actionNote} title={note}>
          {note}
        </span>
      )}
    </>
  );
}

function Action({ route, navigation, provenance, onCopy }: NoticeProps) {
  const { action } = route;
  if (action.kind === "behaviour") {
    return (
      <button
        type="button"
        className={cx(shared.primary, styles.action)}
        onClick={() => {
          navigation.go({ screen: "device", path: action.device, step: "behaviour" });
        }}
      >
        {route.label}
      </button>
    );
  }
  if (action.kind === "example") {
    return (
      <button type="button" className={cx(shared.primary, styles.action)} onClick={onCopy}>
        {route.label}
      </button>
    );
  }
  return <Recompose route={route} provenance={provenance} />;
}

export function RouteNotice(props: NoticeProps) {
  const { route } = props;
  return (
    <section className={styles.notice} aria-label="Why this plant cannot be edited here">
      <strong className={styles.headline} title={`${route.sentence} ${route.origin}`.trim()}>
        {route.headline}
      </strong>
      <Action {...props} />
    </section>
  );
}

export function RouteWhy({ route }: { readonly route: PlantRoute }) {
  return (
    <section aria-label="Why this plant is read-only">
      <p className={styles.sentence}>{route.sentence}</p>
      {route.origin !== "" && <p className={styles.sentence}>{route.origin}</p>}
    </section>
  );
}

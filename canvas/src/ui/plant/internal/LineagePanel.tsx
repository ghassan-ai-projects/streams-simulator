import type { NavigationStore } from "../../../app/shell";
import type { ChainStep, LineagePlant, Unreadable } from "../../../model/workspace";
import { chainOf, digestOf, runLines } from "../../../model/workspace";
import { shared } from "../../primitives";
import styles from "./Provenance.module.css";

interface PanelProps {
  readonly plant: LineagePlant;
  readonly unreadable: readonly Unreadable[];
  readonly navigation: NavigationStore;
}

function Step({ step }: { readonly step: ChainStep }) {
  return (
    <li className={styles.step} data-tone={step.tone === "" ? undefined : step.tone} title={step.title}>
      <span className={styles.stepLabel}>{step.label}</span>
      <span className={styles.mono}>{step.value}</span>
      {step.detail !== "" && <span className={styles.detail}>{step.detail}</span>}
    </li>
  );
}

function Runs({ plant, navigation }: Pick<PanelProps, "plant" | "navigation">) {
  const lines = runLines(plant);
  if (lines.length === 0) return <p className={styles.muted}>No run has used this plant yet.</p>;
  return (
    <ul className={styles.runs}>
      {lines.map((line) => (
        <li key={line.id} className={styles.run}>
          <button
            type="button"
            className={shared.secondary}
            aria-label={`Open the review of ${line.id}`}
            onClick={() => {
              navigation.go({ screen: "review", worldId: line.id });
            }}
          >
            {line.id}
          </button>
          <span>{line.facts}</span>
          <span className={styles.verdict} data-current={line.current ? "true" : "false"}>
            {line.verdict}
          </span>
          {!line.current && line.verdict !== "no run record to compare" && (
            <button
              type="button"
              className={shared.secondary}
              aria-label={`What changed since ${line.id}`}
              onClick={() => {
                navigation.go({ screen: "review", worldId: line.id, model: true });
              }}
            >
              What changed
            </button>
          )}
        </li>
      ))}
    </ul>
  );
}

function Unreadables({ files }: { readonly files: readonly Unreadable[] }) {
  if (files.length === 0) return null;
  return (
    <section aria-label="Files that could not be read">
      <h3 className={styles.sectionTitle}>Files that could not be read</h3>
      <ul className={styles.unreadable}>
        {files.map((file) => (
          <li key={`${file.file}:${file.code}`}>
            <span className={styles.mono}>{file.file}</span>
            {` — ${file.message}`}
          </li>
        ))}
      </ul>
    </section>
  );
}

export function LineagePanel({ plant, unreadable, navigation }: PanelProps) {
  const digest = digestOf(plant);
  return (
    <section aria-label="Lineage of this plant">
      <p className={styles.muted}>
        {`${plant.id} ${plant.version} · `}
        <span className={styles.digest} title={digest.full}>
          {digest.short}
        </span>
      </p>
      <ol className={styles.chain} aria-label="Where this plant comes from">
        {chainOf(plant).map((step) => (
          <Step key={step.label} step={step} />
        ))}
      </ol>
      <h3 className={styles.sectionTitle}>Runs that used this plant</h3>
      <Runs plant={plant} navigation={navigation} />
      <Unreadables files={unreadable} />
    </section>
  );
}

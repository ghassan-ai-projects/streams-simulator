import type { Comparison } from "../../../model/review";
import { CompareDiff } from "./CompareDiff";
import { CompareResults } from "./CompareResults";
import styles from "./Compare.module.css";

function Limits({ comparison }: { readonly comparison: Comparison }) {
  return (
    <section className={styles.block} aria-label="Comparability">
      <h3 className={styles.blockTitle}>Comparability</h3>
      <p className={styles.muted}>In these messages run A is the reference and run B is this run.</p>
      <ul className={styles.limits}>
        {comparison.limits.map((limit) => (
          <li key={`${limit.code}-${limit.side ?? ""}`} data-code={limit.code}>
            {limit.message}
          </li>
        ))}
      </ul>
    </section>
  );
}

export function ComparisonView({
  comparison,
  unit,
}: {
  readonly comparison: Comparison;
  readonly unit: (entity: string, state: string) => string;
}) {
  return (
    <div className={styles.result}>
      <CompareDiff comparison={comparison} />
      <Limits comparison={comparison} />
      <CompareResults comparison={comparison} unit={unit} />
    </div>
  );
}

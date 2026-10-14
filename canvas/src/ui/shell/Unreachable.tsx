import type { BridgeFailure } from "../../model/bridge";
import styles from "./Unreachable.module.css";

export function Unreachable({ failure }: { readonly failure: BridgeFailure }) {
  return (
    <div className={styles.wrap}>
      <section className={styles.card} aria-labelledby="unreachable-title">
        <h1 id="unreachable-title" className={styles.title}>
          The bridge is not reachable
        </h1>
        <p className={styles.body}>
          {failure.message} The canvas reads plants, compositions and runs from a workspace folder through the local
          bridge. Start it from a terminal and open the address it prints.
        </p>
        <code className={styles.command}>
          streamsim canvas --workspace &lt;folder&gt; --ui-dir &lt;built canvas&gt;
        </code>
      </section>
    </div>
  );
}

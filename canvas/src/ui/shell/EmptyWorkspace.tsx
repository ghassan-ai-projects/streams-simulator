import styles from "./EmptyWorkspace.module.css";

export function EmptyWorkspace() {
  return (
    <div className={styles.wrap}>
      <section className={styles.card} aria-labelledby="empty-title">
        <h1 id="empty-title" className={styles.title}>
          No workspace connected
        </h1>
        <p className={styles.body}>
          The canvas reads plants, compositions and runs from a workspace folder through the local bridge. Start it from
          a terminal and open the address it prints.
        </p>
        <code className={styles.command}>streamsim canvas --workspace &lt;folder&gt;</code>
      </section>
    </div>
  );
}

import styles from "./Sidebar.module.css";

const sections = ["Plants", "Compositions", "Runs"] as const;

export function Sidebar() {
  return (
    <nav className={styles.sidebar} aria-label="Workspace contents">
      {sections.map((section) => (
        <section key={section} aria-labelledby={`section-${section}`}>
          <h2 id={`section-${section}`} className={styles.heading}>
            {section}
          </h2>
          <p className={styles.empty}>Nothing yet</p>
        </section>
      ))}
    </nav>
  );
}

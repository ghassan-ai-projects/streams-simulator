import type { ReactNode } from "react";
import { cx, shared } from "../../primitives";
import styles from "./Inspector.module.css";

interface ListProps {
  readonly title: string;
  readonly items: readonly string[];
  readonly mono?: boolean | undefined;
  readonly show?: ((item: string) => string) | undefined;
  readonly count?: number;
  readonly folded?: boolean;
  readonly rows?: ReactNode;
}

function Items({ items, mono = true, show }: Pick<ListProps, "items" | "mono" | "show">) {
  if (items.length === 0) return <p className={styles.empty}>None</p>;
  return (
    <ul className={styles.items}>
      {items.map((item) => (
        <li key={item} className={mono ? styles.mono : undefined} title={show && item}>
          {show ? show(item) : item}
        </li>
      ))}
    </ul>
  );
}

export function List({ title, items, mono, show, count = items.length, folded = false, rows }: ListProps) {
  const heading = (
    <>
      {title}
      <span className={styles.count}>{count}</span>
    </>
  );
  if (folded && items.length > 0) {
    return (
      <section className={styles.section} aria-label={title}>
        <details>
          <summary className={cx(shared.sectionTitle, shared.disclosure)}>{heading}</summary>
          {rows ?? <Items items={items} mono={mono} show={show} />}
        </details>
      </section>
    );
  }
  return (
    <section className={styles.section} aria-label={title}>
      <h3 className={shared.sectionTitle}>{heading}</h3>
      {rows ?? <Items items={items} mono={mono} show={show} />}
    </section>
  );
}

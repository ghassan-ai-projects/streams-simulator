import { shared } from "../../primitives";
import type { ReactNode } from "react";
import cells from "./ReportTable.module.css";
import styles from "./ReportParts.module.css";

type Column = "clock" | "id" | "label" | "effector" | "outcome" | "text" | "number";

export type Cell = string | { readonly text: string; readonly title: string };

export function Section({
  title,
  note,
  children,
}: {
  readonly title: string;
  readonly note?: string | undefined;
  readonly children: ReactNode;
}) {
  return (
    <section className={styles.section} aria-label={title}>
      <div className={shared.sectionTitle}>
        <h2 className={styles.sectionName}>{title}</h2>
        {note !== undefined && <span className={styles.note}>{note}</span>}
      </div>
      {children}
    </section>
  );
}

export interface RowsProps {
  readonly rows: readonly (readonly Cell[])[];
  readonly columns: readonly Column[];
  readonly headers: readonly string[];
  readonly empty?: string;
  readonly current?: number;
  readonly onSelect?: (index: number) => void;
}

interface RowProps {
  readonly row: readonly Cell[];
  readonly columns: readonly Column[];
  readonly current: boolean;
  readonly onSelect: (() => void) | undefined;
}

const textOf = (cell: Cell) => (typeof cell === "string" ? cell : cell.text);

function Shown({ row, onSelect }: { readonly row: readonly Cell[]; readonly onSelect: (() => void) | undefined }) {
  const text = textOf(row[0] ?? "");
  if (onSelect === undefined) return text;
  return (
    <button type="button" className={styles.pick} aria-label={`Open ${row.slice(0, 3).map(textOf).join(" ")}`}>
      {text}
    </button>
  );
}

function Row({ row, columns, current, onSelect }: RowProps) {
  return (
    <tr
      className={onSelect === undefined ? undefined : styles.pickable}
      aria-current={current ? "true" : undefined}
      ref={current ? (element) => element?.scrollIntoView({ block: "nearest" }) : undefined}
      onClick={onSelect}
    >
      {row.map((cell, c) => {
        const Tag = c === 0 ? "th" : "td";
        return (
          <Tag
            key={c}
            scope={c === 0 ? "row" : undefined}
            className={cells[columns[c] ?? "text"]}
            title={typeof cell === "string" ? undefined : cell.title}
          >
            {c === 0 ? <Shown row={row} onSelect={onSelect} /> : textOf(cell)}
          </Tag>
        );
      })}
    </tr>
  );
}

export function Rows({ rows, columns, headers, empty = "None", current = -1, onSelect }: RowsProps) {
  if (rows.length === 0) return <p className={styles.muted}>{empty}</p>;
  return (
    <table className={cells.table}>
      <thead className={cells.head}>
        <tr>
          {headers.map((header) => (
            <th key={header} scope="col">
              {header}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((row, r) => (
          <Row
            key={r}
            row={row}
            columns={columns}
            current={r === current}
            onSelect={
              onSelect === undefined
                ? undefined
                : () => {
                    onSelect(r);
                  }
            }
          />
        ))}
      </tbody>
    </table>
  );
}

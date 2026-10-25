import { CircleCheck, Loader, Minus, TriangleAlert } from "lucide-react";
import type { Draft, ReportRow } from "../../../app/editor";
import { saveReportOf } from "../../../app/editor";
import { cx, panel, shared } from "../../primitives";
import styles from "./Parameters.module.css";

const icons = {
  ok: <CircleCheck size={16} aria-hidden="true" />,
  warn: <TriangleAlert size={16} aria-hidden="true" />,
  bad: <TriangleAlert size={16} aria-hidden="true" />,
  pending: <Loader size={16} className={cx(shared.spin)} aria-hidden="true" />,
  neutral: <Minus size={16} aria-hidden="true" />,
} as const;

function Row({ row }: { readonly row: ReportRow }) {
  return (
    <li className={styles.reportRow} data-state={row.state}>
      {icons[row.state]}
      <span className={styles.reportText}>
        <span className={styles.reportLabel}>{`${row.label}: `}</span>
        <span title={row.title === "" ? undefined : row.title}>{row.text}</span>
        {row.notes.length > 0 && (
          <ul className={styles.notes}>
            {row.notes.map((note) => (
              <li key={note}>{note}</li>
            ))}
          </ul>
        )}
      </span>
    </li>
  );
}

export function SaveReport({ draft }: { readonly draft: Draft }) {
  const rows = saveReportOf(draft);
  if (rows === null) return null;
  return (
    <section className={panel.group} aria-labelledby="save-report-title">
      <h3 id="save-report-title">Last save</h3>
      <ul className={styles.report}>
        {rows.map((row) => (
          <Row key={row.id} row={row} />
        ))}
      </ul>
    </section>
  );
}

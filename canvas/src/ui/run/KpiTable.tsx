import { Fragment } from "react";
import type { KpiRow } from "../../model/report";
import { kpiGroups, kpiWords } from "../../model/reportViews";
import { displayUnit } from "../../model/units";
import { cx } from "../cx";
import styles from "./RunScreen.module.css";

const fixed = (value: number) =>
  Math.abs(value) >= 100 ? value.toFixed(0) : Math.abs(value) >= 10 ? value.toFixed(1) : value.toFixed(2);

interface KpiTableProps {
  readonly rows: readonly KpiRow[];
  readonly grouped: boolean;
  readonly unit: (entity: string, state: string) => string;
}

function KpiLine({ row, unit }: { readonly row: KpiRow; readonly unit: KpiTableProps["unit"] }) {
  const shown = displayUnit(unit(row.entity, row.state));
  const steady = fixed(row.min) === fixed(row.max);
  return (
    <tr>
      <th scope="row" className={styles.kpiName} title={`${row.entity} ${row.state}`}>
        {kpiWords(row.state)}
      </th>
      <td className={cx(styles.number, styles.kpiValue)}>{fixed(row.mean)}</td>
      <td className={styles.unitCell}>{shown}</td>
      <td className={cx(steady ? styles.steady : styles.number, styles.kpiValue)}>
        {steady ? "steady" : `${fixed(row.min)} – ${fixed(row.max)}`}
      </td>
    </tr>
  );
}

export function KpiTable({ rows, grouped, unit }: KpiTableProps) {
  return (
    <table className={styles.table}>
      <thead className={styles.head}>
        <tr>
          <th scope="col">KPI</th>
          <th scope="col" className={styles.numberHead}>
            Mean
          </th>
          <th scope="col">
            <span className="sr-only">Unit</span>
          </th>
          <th scope="col" className={styles.numberHead}>
            Range
          </th>
        </tr>
      </thead>
      <tbody>
        {kpiGroups(rows).map((group) => (
          <Fragment key={group.entity}>
            {grouped && (
              <tr className={styles.groupRow}>
                <th scope="rowgroup" colSpan={4}>
                  {group.entity}
                </th>
              </tr>
            )}
            {group.rows.map((row) => (
              <KpiLine key={row.state} row={row} unit={unit} />
            ))}
          </Fragment>
        ))}
      </tbody>
    </table>
  );
}

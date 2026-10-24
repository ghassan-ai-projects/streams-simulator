import type { ObservedRow } from "../../../model/review";
import { cx } from "../../primitives";
import { OutsideMark } from "../../primitives";
import { MiniTrend } from "./MiniTrend";
import cells from "./ReportTable.module.css";
import styles from "./ObservedTable.module.css";

function Value({ row }: { readonly row: ObservedRow }) {
  return (
    <td className={cx(cells.number, styles.value, row.attention !== null && styles[`value-${row.attention}`])}>
      <span data-part="value">{row.value}</span>
      {row.unit !== "" && (
        <span className={styles.unit} title={row.unitTitle === "" ? undefined : row.unitTitle}>
          {row.unit}
        </span>
      )}
      <OutsideMark text={row.outside} />
      {row.quality !== "" && (
        <span
          className={styles.quality}
          data-part="quality"
          title={`The instrument reported this reading as ${row.quality}`}
        >
          {row.quality}
        </span>
      )}
    </td>
  );
}

function When({ row }: { readonly row: ObservedRow }) {
  return (
    <td className={styles.when}>
      <span data-part="observed-at" title="run time of this observation">
        {row.observedAt}
      </span>
      <span className={styles.age} data-part="age">
        {row.age}
      </span>
      {row.measuredAt !== "" && (
        <span className={styles.age} title="when the instrument took the reading">{`measured ${row.measuredAt}`}</span>
      )}
    </td>
  );
}

function Row({ row }: { readonly row: ObservedRow }) {
  return (
    <tr data-part="observed-row" data-quality={row.quality === "" ? undefined : row.quality}>
      <th scope="row" className={cells.label} title={row.key}>
        {row.label}
      </th>
      {row.present ? (
        <>
          <Value row={row} />
          <When row={row} />
          <td className={styles.trendCell}>
            {row.trend !== null && <MiniTrend trend={row.trend} label={row.label} note={row.trendNote} />}
          </td>
        </>
      ) : (
        <td colSpan={3} className={styles.absent} data-part="absent">
          {row.absent}
        </td>
      )}
    </tr>
  );
}

export function ObservedTable({ rows, clock }: { readonly rows: readonly ObservedRow[]; readonly clock: string }) {
  return (
    <table className={cells.table} data-part="observed">
      <caption className={styles.caption}>{clock}</caption>
      <thead className="sr-only">
        <tr>
          <th scope="col">Channel</th>
          <th scope="col">Value</th>
          <th scope="col">Observed at</th>
          <th scope="col">Trend</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((row) => (
          <Row key={row.key} row={row} />
        ))}
      </tbody>
    </table>
  );
}

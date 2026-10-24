import type { Plot, Readout } from "../../../model/review";
import type { PlotCursor, PlotHandlers } from "./TrendPlotSvg";
import { TrendPlotSvg } from "./TrendPlotSvg";
import styles from "./Trends.module.css";

interface RowProps {
  readonly plot: Plot;
  readonly cursor: PlotCursor;
  readonly readout: Readout | null;
  readonly handlers: PlotHandlers;
  readonly onRemove: () => void;
}

function Legend({ plot, readout, onRemove }: Pick<RowProps, "plot" | "readout" | "onRemove">) {
  return (
    <div className={styles.legend}>
      <span className={styles.name} title={plot.channel}>
        {plot.label}
      </span>
      {readout === null ? (
        <span className={styles.when}>No reading yet</span>
      ) : (
        <>
          <span className={styles.reading}>
            <span data-part="trend-value">{readout.value}</span>
            {readout.unit !== "" && (
              <span className={styles.unit} title={readout.unitTitle === "" ? undefined : readout.unitTitle}>
                {readout.unit}
              </span>
            )}
            {readout.quality !== "" && <span className={styles.quality}>{readout.quality}</span>}
          </span>
          <span className={styles.when} data-part="trend-when">{`at ${readout.observedAt} · ${readout.age}`}</span>
        </>
      )}
      <button type="button" className={styles.remove} aria-label={`Stop watching ${plot.label}`} onClick={onRemove}>
        Remove
      </button>
    </div>
  );
}

export function TrendRow({ plot, cursor, readout, handlers, onRemove }: RowProps) {
  return (
    <li className={styles.row} data-state={plot.state} data-gaps={plot.gaps}>
      <Legend plot={plot} readout={readout} onRemove={onRemove} />
      <TrendPlotSvg plot={plot} cursor={cursor} readout={readout} handlers={handlers} />
    </li>
  );
}

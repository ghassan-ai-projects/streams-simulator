import styles from "./TrendPlot.module.css";

export function TrendsHelp() {
  return (
    <details className={styles.help}>
      <summary>How to read these plots</summary>
      <p>
        Each plot has its own scale and its own unit; channels with different units are never drawn on one axis. All
        plots share one time axis and one cursor. A pin on a channel in the Observed table adds it; Remove drops it.
      </p>
      <p>
        A gap is a pause longer than 3 times the median interval between a channel&apos;s readings, inferred from the
        observed intervals. Nothing is drawn across a gap, and a missing reading is never drawn as zero. Long windows
        keep the lowest and highest reading of each stretch. Coloured marks along the bottom of a plot are readings the
        instrument flagged (suspect, out of range, stale or disconnected).
      </p>
      <p>
        Triangles on the event row are your effector calls (filled accepted, outlined refused). Diamonds are injected
        faults, shown only where the run&apos;s truth is available. Left and Right move the probe to the previous or
        next reading of the plot in focus; Shift or Page keys move it by a twentieth of the window; Home and End jump to
        the ends; Enter selects the instance at that time; Escape clears the probe.
      </p>
    </details>
  );
}

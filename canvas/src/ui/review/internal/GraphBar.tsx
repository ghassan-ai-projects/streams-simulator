import type { ReviewState } from "../../../app/review";
import type { Opening } from "../../../model/review";
import { runClock } from "../../../model/review";
import { grouped } from "../../../model/shared/units";
import { ZoomBar } from "../../graph";
import styles from "./GraphBar.module.css";

function openedWords(opening: Opening, startNs: number): string {
  const at = runClock(opening.atNs, startNs);
  return opening.full ? ` · readings start at ${at}` : ` · opened at ${at}, some channels report later`;
}

function graphWords(state: ReviewState, opening: Opening | null, startNs: number): string {
  const loading = state.playback.phase === "loading" ? " · loading" : "";
  const shown = opening !== null && state.positionNs === opening.atNs ? openedWords(opening, startNs) : "";
  return `${grouped(state.playback.records)} observations${loading}${shown}`;
}

export function GraphBar({
  state,
  opening,
  startNs,
}: {
  readonly state: ReviewState;
  readonly opening: Opening | null;
  readonly startNs: number;
}) {
  const words = graphWords(state, opening, startNs);
  return (
    <div className={styles.graphBar}>
      <span className={styles.barWords} title={words}>
        {words}
      </span>
      <ZoomBar review />
    </div>
  );
}

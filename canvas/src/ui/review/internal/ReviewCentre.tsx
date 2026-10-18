import type { NavigationStore } from "../../../app/shell";
import type { ReviewSession, ReviewState } from "../../../app/review";
import { readyReport } from "../../../app/review";
import { runSpan } from "../../../model/review";
import { MissingRun } from "./MissingRun";
import { Graph } from "./ReviewGraph";
import type { GraphProps } from "./ReviewGraph";
import { Scrubber } from "./Scrubber";
import styles from "./ReviewScreen.module.css";
import { LiveControls } from "./LiveControls";

function Timeline({ review, state }: { readonly review: ReviewSession; readonly state: ReviewState }) {
  const report = readyReport(state);
  const span = runSpan(report, review.index.bounds());
  const startNs = Math.min(span.startNs, state.liveHistory?.startNs ?? span.startNs);
  return (
    <div>
      {state.connection === "running" && <LiveControls review={review} state={state} />}
      <Scrubber
        startNs={startNs}
        endNs={Math.max(span.endNs, state.newestNs)}
        liveMarks={state.liveMarks}
        positionNs={state.positionNs === 0 ? span.startNs : state.positionNs}
        report={report}
        origins={state.origins}
        binding={state.binding}
        playing={state.playing}
        speed={state.speed}
        onPlay={() => {
          review.play();
        }}
        onPause={() => {
          review.pause();
        }}
        onSpeed={(speed) => {
          review.setSpeed(speed);
        }}
        onSeek={(ns) => {
          review.seek(ns);
        }}
        onEvent={(mark) => {
          if (mark.entity === "") review.seek(mark.atNs);
          else review.reveal(mark.atNs, mark.entity);
        }}
      />
    </div>
  );
}

interface CentreProps extends GraphProps {
  readonly worldId: string;
  readonly navigation: NavigationStore;
  readonly review: ReviewSession;
}

export function Centre({ review, state, plant, hall, onHall, worldId, navigation }: CentreProps) {
  return (
    <section className={styles.center} aria-label={`Run ${worldId}`}>
      <Timeline review={review} state={state} />
      {state.report.phase === "failed" && state.report.failure.kind === "missing" ? (
        <MissingRun worldId={worldId} navigation={navigation} />
      ) : (
        <Graph review={review} state={state} plant={plant} hall={hall} onHall={onHall} />
      )}
    </section>
  );
}

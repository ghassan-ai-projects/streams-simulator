import type { NavigationStore } from "../../../app/shell";
import type { ReviewSession, ReviewState } from "../../../app/review";
import { readyReport } from "../../../app/review";
import { runSpan } from "../../../model/review";
import type { EventMark } from "../../../model/review";
import { MissingRun } from "./MissingRun";
import { Graph } from "./ReviewGraph";
import type { GraphProps } from "./ReviewGraph";
import { Scrubber } from "./Scrubber";
import styles from "./ReviewScreen.module.css";
import { LiveControls } from "./LiveControls";

function handlers(review: ReviewSession, running: boolean) {
  return {
    onPlay: () => {
      review.play();
    },
    onPause: () => {
      review.pause();
    },
    onSpeed: (speed: number) => {
      review.setSpeed(speed);
    },
    onSeek: (ns: number) => {
      review.seek(ns);
    },
    onEvent: (mark: EventMark) => {
      if (mark.entity === "") review.seek(mark.atNs);
      else review.reveal(mark.atNs, mark.entity);
    },
    ...(running
      ? {
          onLatest: () => {
            review.backToLive();
          },
        }
      : {}),
  };
}

interface TimelineProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly now: () => number;
}

function Timeline({ review, state, now }: TimelineProps) {
  const report = readyReport(state);
  const span = runSpan(report, review.index.bounds());
  const startNs = Math.min(span.startNs, state.liveHistory?.startNs ?? span.startNs);
  const running = state.connection === "running";
  return (
    <div>
      {running && <LiveControls review={review} state={state} startNs={startNs} now={now} />}
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
        {...handlers(review, running)}
      />
    </div>
  );
}

interface CentreProps extends GraphProps {
  readonly now: () => number;
  readonly worldId: string;
  readonly navigation: NavigationStore;
  readonly review: ReviewSession;
}

export function Centre({ review, state, plant, hall, onHall, worldId, navigation, now }: CentreProps) {
  return (
    <section className={styles.center} aria-label={`Run ${worldId}`}>
      <Timeline review={review} state={state} now={now} />
      {state.report.phase === "failed" && state.report.failure.kind === "missing" ? (
        <MissingRun worldId={worldId} navigation={navigation} />
      ) : (
        <Graph review={review} state={state} plant={plant} hall={hall} onHall={onHall} />
      )}
    </section>
  );
}

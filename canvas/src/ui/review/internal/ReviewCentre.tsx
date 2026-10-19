import type { ConnectSession } from "../../../app/launch";
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
import { Interrupts } from "./Interrupts";
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

interface CentreProps extends GraphProps {
  readonly connect: ConnectSession;
  readonly now: () => number;
  readonly worldId: string;
  readonly navigation: NavigationStore;
  readonly review: ReviewSession;
}

function startOf(review: ReviewSession, state: ReviewState): number {
  const span = runSpan(readyReport(state), review.index.bounds());
  return Math.min(span.startNs, state.liveHistory?.startNs ?? span.startNs);
}

function Timeline({ review, state }: Pick<CentreProps, "review" | "state">) {
  const report = readyReport(state);
  const span = runSpan(report, review.index.bounds());
  return (
    <Scrubber
      startNs={startOf(review, state)}
      endNs={Math.max(span.endNs, state.newestNs)}
      liveMarks={state.liveMarks}
      positionNs={state.positionNs === 0 ? span.startNs : state.positionNs}
      report={report}
      origins={state.origins}
      binding={state.binding}
      playing={state.playing}
      speed={state.speed}
      {...handlers(review, state.connection === "running")}
    />
  );
}

function LiveTop({ review, state, connect, now }: Omit<CentreProps, "worldId" | "navigation">) {
  return (
    <div className={styles.liveTop}>
      <LiveControls review={review} state={state} startNs={startOf(review, state)} now={now} />
      <Interrupts review={review} state={state} connect={connect} now={now} />
    </div>
  );
}

export function Centre(props: CentreProps) {
  const { review, state, plant, hall, onHall, worldId, navigation } = props;
  const missing = state.report.phase === "failed" && state.report.failure.kind === "missing";
  return (
    <section className={styles.center} aria-label={`Run ${worldId}`}>
      {state.connection === "running" && <LiveTop {...props} />}
      <div className={styles.body}>
        <Timeline review={review} state={state} />
        {missing ? (
          <MissingRun worldId={worldId} navigation={navigation} />
        ) : (
          <Graph review={review} state={state} plant={plant} hall={hall} onHall={onHall} />
        )}
      </div>
    </section>
  );
}

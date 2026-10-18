import { useRef, useState } from "react";
import type { ReviewSession, ReviewState } from "../../../app/review";
import { runClock, updateAgeWords } from "../../../model/review";
import { shared, useInterval } from "../../primitives";
import styles from "./LiveControls.module.css";

interface LiveControlsProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly startNs: number;
  readonly now: () => number;
}

function followWords(state: ReviewState): string {
  if (!state.caughtUp) return "Catching up to live…";
  return state.following ? "Live · following newest record" : "Following paused · viewing history";
}

function useNow(now: () => number): number {
  const [current, setCurrent] = useState(now);
  useInterval(() => {
    setCurrent(now());
  }, 1000);
  return current;
}

function Age({ state, now }: Pick<LiveControlsProps, "state" | "now">) {
  const current = useNow(now);
  if (state.lastLiveAtMs === null) return null;
  return <span className={styles.age}>{updateAgeWords(current - state.lastLiveAtMs)}</span>;
}

function Notes({ state, startNs }: Pick<LiveControlsProps, "state" | "startNs">) {
  const trimmed = state.liveHistory?.trimmedBeforeNs ?? null;
  return (
    <>
      {state.liveUnreadableNs !== null && (
        <span className={styles.alert} role="alert">
          {`A live snapshot could not be read at ${runClock(state.liveUnreadableNs, startNs)}; the view is not current.`}
        </span>
      )}
      {trimmed !== null && (
        <span className={styles.note}>{`History trimmed before ${runClock(trimmed, startNs)}`}</span>
      )}
    </>
  );
}

export function LiveControls({ review, state, startNs, now }: LiveControlsProps) {
  const status = useRef<HTMLSpanElement>(null);
  return (
    <div role="group" aria-label="Live following" className={styles.bar}>
      <span role="status" ref={status} tabIndex={-1} className={styles.state}>
        {followWords(state)}
      </span>
      <Age state={state} now={now} />
      <Notes state={state} startNs={startNs} />
      {state.caughtUp && !state.following && (
        <button
          type="button"
          className={shared.primary}
          onClick={() => {
            review.backToLive();
            status.current?.focus();
          }}
        >
          Back to live
        </button>
      )}
    </div>
  );
}

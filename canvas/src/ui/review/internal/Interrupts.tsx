import { useId, useState } from "react";
import type { ConnectSession } from "../../../app/launch";
import type { ReviewSession, ReviewState } from "../../../app/review";
import type { InterruptState, NodeNow } from "../../../model/review";
import {
  activeScopes,
  allDevices,
  chosenScope,
  newestInterrupts,
  scopeChoices,
  scopeKey,
  scopeStatus,
} from "../../../model/review";
import { useInterval, useStore } from "../../primitives";
import { ActiveInterrupts } from "./ActiveInterrupts";
import { InterruptButtons, ScopeField } from "./InterruptButtons";
import { InterruptOutcome } from "./InterruptOutcome";
import { PendingApprovals } from "./PendingApprovals";
import styles from "./Interrupts.module.css";

interface InterruptsProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly connect: ConnectSession;
  readonly now: () => number;
}

function useWallMs(now: () => number): number {
  const [current, setCurrent] = useState(now);
  useInterval(() => {
    setCurrent(now());
  }, 1000);
  return current;
}

function reasonOf(shown: InterruptState | null, offered: boolean): "waiting" | "not-offered" | null {
  if (shown !== null) return null;
  return offered ? "waiting" : "not-offered";
}

export function Interrupts({ review, state, connect, now }: InterruptsProps) {
  const noteId = useId();
  const [picked, setPicked] = useState(scopeKey(allDevices));
  const controls = useStore(review.controls);
  const connection = useStore(connect);
  const wallMs = useWallMs(now);
  const newest = state.liveHistory?.entries.at(-1) ?? null;
  const shown = newestInterrupts(newest, controls.seen);
  const choices = scopeChoices(newest);
  const scope = chosenScope(choices, picked);
  const offered = connection.phase !== "live" || connection.controls;
  const nodeNow: NodeNow = { worldNs: newest?.atNs ?? 0, wallMs };
  return (
    <section className={styles.bar} aria-label="Interrupts">
      <div className={styles.row}>
        <h2 className={styles.title}>Interrupts</h2>
        <ScopeField choices={choices} picked={scopeKey(scope)} onPick={setPicked} />
        <InterruptButtons
          controls={review.controls}
          scope={scope}
          status={scopeStatus(shown, scope)}
          ready={shown !== null}
          noteId={noteId}
        />
      </div>
      <ActiveInterrupts active={activeScopes(shown)} noteId={noteId} />
      <InterruptOutcome review={review} reason={reasonOf(shown, offered)} />
      <PendingApprovals controls={review.controls} pending={shown?.pending ?? []} now={nodeNow} />
    </section>
  );
}

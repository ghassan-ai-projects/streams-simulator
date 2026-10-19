import { useId } from "react";
import type { ConnectSession, ConnectState } from "../../../app/launch";
import type { ReviewSession, ReviewState } from "../../../app/review";
import type { InterruptState, ScopeChoice, ScopeStatus } from "../../../model/review";
import {
  activeScopes,
  chosenChoice,
  newestInterrupts,
  scopeChoices,
  scopeLabel,
  scopeStatus,
} from "../../../model/review";
import { useStore } from "../../primitives";
import { ActiveInterrupts } from "./ActiveInterrupts";
import type { Enabled } from "./InterruptButtons";
import { InterruptButtons, ScopeField } from "./InterruptButtons";
import type { Reason } from "./InterruptOutcome";
import { InterruptOutcome } from "./InterruptOutcome";
import { PendingApprovals } from "./PendingApprovals";
import styles from "./Interrupts.module.css";

interface InterruptsProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly connect: ConnectSession;
  readonly now: () => number;
}

interface Facts {
  readonly shown: InterruptState | null;
  readonly offered: boolean;
  readonly stopping: boolean;
  readonly present: boolean;
}

function reasonOf({ shown, offered, stopping, present }: Facts): Reason | null {
  if (!offered) return "not-offered";
  if (stopping) return "stopping";
  if (shown === null) return "waiting";
  return present ? null : "absent-scope";
}

function enabledOf({ shown, offered, stopping, present }: Facts): Enabled {
  const reachable = offered && !stopping;
  return { actions: reachable && shown !== null && present, stop: reachable };
}

function factsOf(shown: InterruptState | null, connection: ConnectState, choice: ScopeChoice): Facts {
  const offered = connection.phase !== "live" || connection.controls;
  return { shown, offered, stopping: connection.phase === "live" && connection.stopping, present: choice.present };
}

interface BarView {
  readonly world: number;
  readonly shown: InterruptState | null;
  readonly choices: readonly ScopeChoice[];
  readonly choice: ScopeChoice;
  readonly status: ScopeStatus;
  readonly enabled: Enabled;
  readonly reason: Reason | null;
}

function useBarView({ review, state, connect }: InterruptsProps): BarView {
  const controls = useStore(review.controls);
  const connection = useStore(connect);
  const newest = state.liveHistory?.entries.at(-1) ?? null;
  const shown = newestInterrupts(newest, controls.seen, state.worldId);
  const choices = scopeChoices(newest, controls.picked);
  const choice = chosenChoice(choices, controls.picked);
  const facts = factsOf(shown, connection, choice);
  const status = scopeStatus(shown, choice.scope);
  return {
    world: newest?.atNs ?? 0,
    shown,
    choices,
    choice,
    status,
    enabled: enabledOf(facts),
    reason: reasonOf(facts),
  };
}

interface RowProps {
  readonly review: ReviewSession;
  readonly view: BarView;
  readonly noteId: string;
}

function ControlRow({ review, view, noteId }: RowProps) {
  const { choices, choice } = view;
  return (
    <div className={styles.row}>
      <ScopeField
        choices={choices}
        picked={choice.key}
        onPick={(key) => {
          const picked = choices.find((entry) => entry.key === key);
          if (picked !== undefined) review.controls.pick(picked.scope);
        }}
      />
      <InterruptButtons
        controls={review.controls}
        scope={choice.scope}
        status={view.status}
        enabled={view.enabled}
        noteId={noteId}
      />
    </div>
  );
}

function StatusRow({ review, view }: Omit<RowProps, "noteId">) {
  const { status, choice } = view;
  const covered = status.coveredBy === "" ? "" : `${scopeLabel(choice.scope)} is also covered by ${status.coveredBy}.`;
  return (
    <div className={styles.statusRow}>
      <ActiveInterrupts active={activeScopes(view.shown)} covered={covered} />
      <InterruptOutcome review={review} reason={view.reason} />
    </div>
  );
}

export function Interrupts(props: InterruptsProps) {
  const noteId = useId();
  const { review, now } = props;
  const view = useBarView(props);
  return (
    <section className={styles.bar} aria-label="Interrupts">
      <h2 className="sr-only">Interrupts</h2>
      <ControlRow review={review} view={view} noteId={noteId} />
      <StatusRow review={review} view={view} />
      <p className="sr-only" id={noteId}>
        Emergency stop sends each device in scope its safe stop and holds the scope, latched, until you release it.
      </p>
      <PendingApprovals
        controls={review.controls}
        pending={view.shown?.pending ?? []}
        world={view.world}
        now={now}
        enabled={view.enabled.stop}
      />
    </section>
  );
}

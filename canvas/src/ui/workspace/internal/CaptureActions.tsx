import { Check } from "lucide-react";
import type { KeyboardEvent } from "react";
import { useEffect, useRef, useState } from "react";
import type { Card, CaptureSession } from "../../../app/devices";
import { canSkip, canWrite } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { shared, useFocusOnMount } from "../../primitives";
import styles from "./Capture.module.css";
import cardStyles from "./CaptureCard.module.css";
import type { CardFocus } from "./useCardFocus";

interface ActionProps {
  readonly card: Card;
  readonly index: number;
  readonly session: CaptureSession;
  readonly navigation: NavigationStore;
  readonly focus: CardFocus;
}

type Part = Pick<ActionProps, "card" | "index" | "session">;

function deviceName(card: Card): string {
  return card.kind.kind;
}

function Saved({ card, path, navigation }: Pick<ActionProps, "card" | "navigation"> & { readonly path: string }) {
  const open = useFocusOnMount<HTMLButtonElement>();
  return (
    <div className={cardStyles.outcome}>
      <span className={cardStyles.done}>
        <Check size={14} aria-hidden="true" /> Saved as {path}
      </span>
      <button
        type="button"
        ref={open}
        className={shared.primary}
        aria-label={`Open ${deviceName(card)} in the device editor`}
        onClick={() => {
          navigation.go({ screen: "device", path });
        }}
      >
        Open in the device editor
      </button>
    </div>
  );
}

function Skipped({ card, index, session, focus }: Pick<ActionProps, "card" | "index" | "session" | "focus">) {
  return (
    <p className={cardStyles.outcome}>
      Skipped. Nothing will be saved for {deviceName(card)}.
      <button
        type="button"
        className={shared.secondary}
        aria-label={`Undo skipping ${deviceName(card)}`}
        onClick={() => {
          session.unskip(index);
          focus.afterUndo(index);
        }}
      >
        Undo
      </button>
    </p>
  );
}

function useReturnFocus(asking: boolean) {
  const trigger = useRef<HTMLButtonElement>(null);
  const was = useRef(false);
  useEffect(() => {
    if (was.current && !asking) trigger.current?.focus();
    was.current = asking;
  }, [asking]);
  return trigger;
}

function Question({ card, onAnswer }: { readonly card: Card; readonly onAnswer: (replace: boolean) => void }) {
  const keep = useFocusOnMount<HTMLButtonElement>();
  const escape = (event: KeyboardEvent) => {
    if (event.key !== "Escape") return;
    event.stopPropagation();
    onAnswer(false);
  };
  return (
    <span
      className={cardStyles.confirm}
      role="group"
      aria-label={`Replace the saved ${deviceName(card)}?`}
      onKeyDown={escape}
    >
      The saved file {card.kind.comparison.target} is overwritten with what the board announced.
      <button
        type="button"
        className={shared.secondary}
        onClick={() => {
          onAnswer(true);
        }}
      >
        Yes, replace it
      </button>
      <button
        type="button"
        ref={keep}
        className={shared.primary}
        onClick={() => {
          onAnswer(false);
        }}
      >
        Keep the saved device
      </button>
    </span>
  );
}

function Replace({ card, index, session }: Part) {
  const [asking, setAsking] = useState(false);
  const trigger = useReturnFocus(asking);
  const allowed = canWrite(card, "replace");
  if (asking) {
    return (
      <Question
        card={card}
        onAnswer={(replace) => {
          setAsking(false);
          if (replace) void session.replace(index);
        }}
      />
    );
  }
  return (
    <button
      type="button"
      ref={trigger}
      className={shared.primary}
      aria-label={`Replace the saved device: ${deviceName(card)}`}
      aria-disabled={!allowed}
      onClick={() => {
        if (allowed) setAsking(true);
      }}
    >
      Replace the saved device
    </button>
  );
}

function SaveNew({ card, index, session }: Part) {
  const allowed = canWrite(card, "create");
  return (
    <button
      type="button"
      className={shared.primary}
      aria-label={`Save as device: ${deviceName(card)}`}
      aria-disabled={!allowed}
      onClick={() => {
        if (allowed) void session.save(index);
      }}
    >
      Save as device
    </button>
  );
}

function OpenSaved({ card, navigation }: Pick<ActionProps, "card" | "navigation">) {
  const { comparison } = card.kind;
  return (
    <button
      type="button"
      className={shared.primary}
      aria-label={`Open the saved device: ${deviceName(card)}`}
      onClick={() => {
        if (comparison.status !== "new") navigation.go({ screen: "device", path: comparison.file });
      }}
    >
      Open the saved device
    </button>
  );
}

function Skip({ card, index, session, focus }: Pick<ActionProps, "card" | "index" | "session" | "focus">) {
  const allowed = canSkip(card);
  return (
    <button
      type="button"
      className={shared.secondary}
      aria-label={`Skip ${deviceName(card)}`}
      aria-disabled={!allowed}
      onClick={() => {
        if (!allowed) return;
        session.skip(index);
        focus.afterSkip(index);
      }}
    >
      Skip
    </button>
  );
}

function Primary(props: ActionProps) {
  const { status } = props.card.kind.comparison;
  if (status === "new") return <SaveNew {...props} />;
  if (status === "differs") return <Replace {...props} />;
  return status === "same" ? <OpenSaved {...props} /> : null;
}

function Choices(props: ActionProps) {
  const { card, index } = props;
  return (
    <div className={styles.actions} data-actions-of={index}>
      <Primary {...props} />
      {card.kind.comparison.status !== "same" && <Skip {...props} />}
      {card.decision.kind === "saving" && <span className={styles.facts}>Saving…</span>}
      {card.decision.kind === "failed" && (
        <p role="alert" className={styles.problem}>
          {card.decision.failure.diagnostic.message}
        </p>
      )}
    </div>
  );
}

export function CaptureActions(props: ActionProps) {
  const { decision } = props.card;
  if (decision.kind === "saved") return <Saved card={props.card} path={decision.path} navigation={props.navigation} />;
  if (decision.kind === "skipped") return <Skipped {...props} />;
  return <Choices {...props} />;
}

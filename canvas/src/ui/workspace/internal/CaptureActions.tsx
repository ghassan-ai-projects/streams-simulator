import { Check } from "lucide-react";
import type { KeyboardEvent } from "react";
import { useEffect, useRef, useState } from "react";
import type { Card, CaptureSession } from "../../../app/devices";
import { canWrite } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { shared, useFocusOnMount } from "../../primitives";
import styles from "./Capture.module.css";

interface ActionProps {
  readonly card: Card;
  readonly session: CaptureSession;
  readonly navigation: NavigationStore;
}

type CardAndSession = Pick<ActionProps, "card" | "session">;

function deviceName(card: Card): string {
  return card.kind.kind;
}

function Saved({ card, path, navigation }: Pick<ActionProps, "card" | "navigation"> & { readonly path: string }) {
  const open = useFocusOnMount<HTMLButtonElement>();
  return (
    <div className={styles.outcome}>
      <span className={styles.done}>
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
      className={styles.confirm}
      role="group"
      aria-label={`Replace the saved ${deviceName(card)}?`}
      onKeyDown={escape}
    >
      The saved file {card.path} is overwritten with what the board announced.
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

function Replace({ card, session }: CardAndSession) {
  const [asking, setAsking] = useState(false);
  const trigger = useReturnFocus(asking);
  const allowed = canWrite(card, "replace");
  if (asking) {
    return (
      <Question
        card={card}
        onAnswer={(replace) => {
          setAsking(false);
          if (replace) void session.replace(deviceName(card));
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

function SaveNew({ card, session }: CardAndSession) {
  const allowed = canWrite(card, "create");
  return (
    <button
      type="button"
      className={shared.primary}
      aria-label={`Save as device: ${deviceName(card)}`}
      aria-disabled={!allowed}
      onClick={() => {
        if (allowed) void session.save(deviceName(card));
      }}
    >
      Save as device
    </button>
  );
}

function OpenSaved({ card, navigation }: Pick<ActionProps, "card" | "navigation">) {
  return (
    <button
      type="button"
      className={shared.primary}
      aria-label={`Open the saved device: ${deviceName(card)}`}
      onClick={() => {
        navigation.go({ screen: "device", path: card.path });
      }}
    >
      Open the saved device
    </button>
  );
}

function Skip({ card, session }: CardAndSession) {
  const saving = card.decision.kind === "saving";
  return (
    <button
      type="button"
      className={shared.secondary}
      aria-label={`Skip ${deviceName(card)}`}
      aria-disabled={saving}
      onClick={() => {
        if (!saving) session.skip(deviceName(card));
      }}
    >
      Skip
    </button>
  );
}

function Primary({ card, session, navigation }: ActionProps) {
  const { status } = card.kind.comparison;
  if (status === "new") return <SaveNew card={card} session={session} />;
  if (status === "differs") return <Replace card={card} session={session} />;
  return <OpenSaved card={card} navigation={navigation} />;
}

function Choices(props: ActionProps) {
  const { card, session } = props;
  return (
    <div className={styles.actions}>
      <Primary {...props} />
      <Skip card={card} session={session} />
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
  if (decision.kind === "skipped") {
    return <p className={styles.outcome}>Skipped. Nothing was saved for {deviceName(props.card)}.</p>;
  }
  return <Choices {...props} />;
}

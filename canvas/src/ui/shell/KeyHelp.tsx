import { useEffect, useRef, useState } from "react";
import styles from "./KeyHelp.module.css";

const keys: readonly (readonly [string, string])[] = [
  ["Tab, Shift Tab", "move between the top bar, the outline, the graph, the inspector and the report"],
  ["Arrow keys (graph)", "move to the nearest instance in that direction"],
  ["Enter or Space (graph)", "select the instance, or open the hall"],
  ["Escape (graph)", "clear the selection, then leave the hall"],
  ["Arrow keys (playback)", "step the playback time; Page Up and Page Down jump further"],
  ["Previous and next event", "jump between disturbances and effector calls"],
  ["?", "show or hide this list"],
];

function typing(target: EventTarget | null): boolean {
  return (
    target instanceof HTMLElement &&
    (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable)
  );
}

function useToggle(): readonly [boolean, (open: boolean) => void] {
  const [open, setOpen] = useState(false);
  useEffect(() => {
    const onKey = (event: globalThis.KeyboardEvent) => {
      if (event.key === "?" && !typing(event.target)) setOpen((current) => !current);
      if (event.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("keydown", onKey);
    };
  }, []);
  return [open, setOpen];
}

function KeyList() {
  return (
    <dl className={styles.list}>
      {keys.map(([key, does]) => (
        <div key={key}>
          <dt>
            <kbd>{key}</kbd>
          </dt>
          <dd>{does}</dd>
        </div>
      ))}
    </dl>
  );
}

export function KeyHelp() {
  const [open, setOpen] = useToggle();
  const close = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (open) close.current?.focus();
  }, [open]);
  if (!open) return null;
  return (
    <dialog open className={styles.dialog} aria-labelledby="keys-title" aria-modal="false">
      <h2 id="keys-title" className={styles.title}>
        Keyboard
      </h2>
      <KeyList />
      <button
        ref={close}
        type="button"
        className={styles.close}
        onClick={() => {
          setOpen(false);
        }}
      >
        Close
      </button>
    </dialog>
  );
}

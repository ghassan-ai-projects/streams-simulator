import { useEffect, useRef, useState } from "react";
import styles from "./KeyHelp.module.css";

const keys: readonly (readonly [string, string])[] = [
  ["Tab, Shift Tab", "move between the top bar, the outline, the graph, the inspector and the report"],
  ["Arrow keys (graph)", "move to the nearest instance in that direction"],
  ["Space (graph)", "select the instance or hall"],
  ["Enter (graph)", "select the instance, or open the hall"],
  ["Escape (graph)", "clear the selection, then leave the hall"],
  ["+ and − (graph)", "zoom in and out"],
  ["0 (graph)", "fit the plant in view"],
  ["Arrow keys (playback)", "step one second; Shift and arrow one minute; Page Up and Page Down further"],
  ["Space (review)", "play or pause the playback"],
  ["Previous and next event", "jump between disturbances and effector calls"],
  ["[ and ]", "hide or show the outline panel and the details panel"],
  ["Ctrl or ⌘ Z, Shift Z (editor)", "undo and redo; Ctrl or ⌘ S saves"],
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

function useFocusReturn(open: boolean) {
  const close = useRef<HTMLButtonElement>(null);
  const previous = useRef<Element | null>(null);
  useEffect(() => {
    if (open) {
      previous.current = document.activeElement;
      close.current?.focus();
    } else if (previous.current instanceof HTMLElement) {
      previous.current.focus();
      previous.current = null;
    }
  }, [open]);
  const forget = () => {
    previous.current = null;
  };
  return { close, forget };
}

export function KeyHelp() {
  const [open, setOpen] = useToggle();
  const { close, forget } = useFocusReturn(open);
  if (!open) return null;
  return (
    <dialog
      open
      className={styles.dialog}
      aria-labelledby="keys-title"
      aria-modal="false"
      onBlur={(event) => {
        if (event.currentTarget.contains(event.relatedTarget)) return;
        forget();
        setOpen(false);
      }}
    >
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

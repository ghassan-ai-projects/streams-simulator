import { Keyboard } from "lucide-react";
import type { RefObject } from "react";
import { Fragment, useEffect, useRef, useState } from "react";
import { inTextField } from "../keys";
import styles from "./KeyHelp.module.css";
import { keyGroups } from "./keyList";
import topBar from "./TopBar.module.css";

function useToggle(): readonly [boolean, (open: boolean) => void] {
  const [open, setOpen] = useState(false);
  useEffect(() => {
    const onKey = (event: globalThis.KeyboardEvent) => {
      if (event.key === "?" && !inTextField(event.target)) setOpen((current) => !current);
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
  return keyGroups.map((group, index) => (
    <section key={group.scope} className={styles.group} aria-labelledby={`keys-${String(index)}`}>
      <h3 id={`keys-${String(index)}`}>{group.scope}</h3>
      <dl className={styles.list}>
        {group.rows.map((row) => (
          <div key={row.does}>
            <dt>
              {row.keys.map((key, index) => (
                <Fragment key={key}>
                  {index > 0 && <span className={styles.or}>, </span>}
                  <kbd>{key}</kbd>
                </Fragment>
              ))}
            </dt>
            <dd>{row.does}</dd>
          </div>
        ))}
      </dl>
    </section>
  ));
}

function useFocusReturn(open: boolean) {
  const title = useRef<HTMLHeadingElement>(null);
  const previous = useRef<Element | null>(null);
  useEffect(() => {
    if (open) {
      previous.current = document.activeElement;
      title.current?.focus();
    } else if (previous.current instanceof HTMLElement) {
      previous.current.focus();
      previous.current = null;
    }
  }, [open]);
  const forget = () => {
    previous.current = null;
  };
  return { title, forget };
}

function HelpButton({ open, onOpen }: { readonly open: boolean; readonly onOpen: () => void }) {
  return (
    <button
      type="button"
      className={topBar.labelled}
      aria-expanded={open}
      aria-keyshortcuts="?"
      title="Keyboard (?)"
      onClick={onOpen}
    >
      <Keyboard size={16} aria-hidden="true" />
      Keyboard
    </button>
  );
}

interface DialogProps {
  readonly title: RefObject<HTMLHeadingElement | null>;
  readonly forget: () => void;
  readonly close: () => void;
}

function KeyDialog({ title, forget, close }: DialogProps) {
  return (
    <dialog
      open
      tabIndex={-1}
      className={styles.dialog}
      aria-labelledby="keys-title"
      aria-modal="false"
      onBlur={(event) => {
        if (event.currentTarget.contains(event.relatedTarget)) return;
        forget();
        close();
      }}
    >
      <h2 id="keys-title" className={styles.title} ref={title} tabIndex={-1}>
        Keyboard
      </h2>
      <div className={styles.keys}>
        <KeyList />
      </div>
      <button type="button" className={styles.close} onClick={close}>
        Close
      </button>
    </dialog>
  );
}

export function KeyHelp() {
  const [open, setOpen] = useToggle();
  const { title, forget } = useFocusReturn(open);
  const close = () => {
    setOpen(false);
  };
  return (
    <>
      <HelpButton
        open={open}
        onOpen={() => {
          setOpen(true);
        }}
      />
      {open && <div className={styles.scrim} aria-hidden="true" onClick={close} />}
      {open && <KeyDialog title={title} forget={forget} close={close} />}
    </>
  );
}

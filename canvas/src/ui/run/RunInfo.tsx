import { Info } from "lucide-react";
import { useId, useRef, useState } from "react";
import styles from "./RunScreen.module.css";

export function RunInfo({ text }: { readonly text: string }) {
  const [open, setOpen] = useState(false);
  const button = useRef<HTMLButtonElement>(null);
  const id = useId();
  return (
    <span
      className={styles.info}
      onKeyDown={(event) => {
        if (event.key !== "Escape" || !open) return;
        event.stopPropagation();
        setOpen(false);
        button.current?.focus();
      }}
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) setOpen(false);
      }}
    >
      <button
        ref={button}
        type="button"
        aria-expanded={open}
        aria-controls={id}
        aria-label="About this run"
        title="About this run"
        onClick={() => {
          setOpen(!open);
        }}
      >
        <Info size={14} aria-hidden="true" />
      </button>
      <p id={id} className={styles.popover} hidden={!open}>
        {text}
      </p>
    </span>
  );
}

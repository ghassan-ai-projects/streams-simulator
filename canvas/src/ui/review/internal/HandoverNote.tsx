import { Check } from "lucide-react";
import { useEffect, useState } from "react";
import type { HandoverState } from "../../../app/review";
import styles from "./HandoverNote.module.css";

const savedShownMs = 8000;

function Saved({ name }: { readonly name: string }) {
  const [gone, setGone] = useState(false);
  useEffect(() => {
    const timer = window.setTimeout(() => {
      setGone(true);
    }, savedShownMs);
    return () => {
      window.clearTimeout(timer);
    };
  }, []);
  if (gone) return null;
  return (
    <span className={styles.saved} title={`Saved ${name}`}>
      <Check size={14} aria-hidden="true" />
      {`Saved ${name}`}
    </span>
  );
}

export function HandoverNote({ handover }: { readonly handover: HandoverState }) {
  return (
    <>
      <span role="status" className={styles.live}>
        {handover.phase === "saved" && <Saved name={handover.name} />}
      </span>
      {handover.phase === "failed" && (
        <span role="alert" className={styles.failed} title={handover.message}>
          {handover.message}
        </span>
      )}
    </>
  );
}

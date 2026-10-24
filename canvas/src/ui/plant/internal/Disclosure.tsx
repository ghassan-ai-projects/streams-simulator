import { ChevronRight } from "lucide-react";
import type { ReactNode } from "react";
import { cx, shared } from "../../primitives";
import styles from "./Provenance.module.css";

interface DisclosureProps {
  readonly label: string;
  readonly open: boolean;
  readonly onToggle: () => void;
  readonly children?: ReactNode;
}

export function Disclosure({ label, open, onToggle, children }: DisclosureProps) {
  return (
    <div>
      <button type="button" className={cx(shared.secondary)} aria-expanded={open} onClick={onToggle}>
        <ChevronRight size={14} aria-hidden="true" data-open={open ? "true" : undefined} className={styles.chevron} />
        {label}
      </button>
      {open && children}
    </div>
  );
}

import { GitCompare } from "lucide-react";
import { useState } from "react";
import type { DiffSession } from "../../../app/workspace";
import type { Comparison } from "../../../model/review";
import { shared } from "../../primitives";
import { ModelDiffPanel } from "./ModelDiffPanel";
import styles from "./ModelDiff.module.css";

const differentModels = "different_domain_digest";

export function modelsDiffer(comparison: Comparison): boolean {
  return comparison.limits.some((limit) => limit.code === differentModels);
}

interface LinkProps {
  readonly session: DiffSession;
  readonly reference: string;
  readonly worldId: string;
}

export function ModelLink({ session, reference, worldId }: LinkProps) {
  const [open, setOpen] = useState(false);
  return (
    <section className={styles.link} aria-label="The model of the two runs">
      <p>The two runs used different versions of the plant model.</p>
      <button
        type="button"
        className={shared.secondary}
        aria-expanded={open}
        onClick={() => {
          setOpen(!open);
        }}
      >
        <GitCompare size={14} aria-hidden="true" />
        See what changed in the model
      </button>
      {open && <ModelDiffPanel session={session} a={reference} b={worldId} />}
    </section>
  );
}

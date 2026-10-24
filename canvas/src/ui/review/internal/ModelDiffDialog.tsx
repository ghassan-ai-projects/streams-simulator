import { X } from "lucide-react";
import { useState } from "react";
import type { ProvenanceSession, WorkspaceStore } from "../../../app/workspace";
import type { Report } from "../../../model/review";
import { referenceChoices } from "../../../model/review";
import { runRows } from "../../../model/workspace";
import { cx, shared, useStore } from "../../primitives";
import { ModelDiffPanel } from "./ModelDiffPanel";
import { useModal } from "./useModal";
import compare from "./Compare.module.css";
import styles from "./ModelDiff.module.css";

interface DialogProps {
  readonly provenance: ProvenanceSession;
  readonly workspace: WorkspaceStore;
  readonly report: Report;
  readonly worldId: string;
  readonly onClose: () => void;
}

interface Target {
  readonly value: string;
  readonly label: string;
}

function useTargets(workspace: WorkspaceStore, report: Report, worldId: string): readonly Target[] {
  const listing = useStore(workspace);
  if (listing.phase !== "ready") return [];
  const current = listing.workspace.domains.find((domain) => domain.id === report.domainId);
  const runs = runRows(listing.workspace, report.domainId).map((row) => row.run);
  const others = referenceChoices(runs, { current: worldId, plant: report.domainId });
  return [
    ...(current === undefined ? [] : [{ value: current.path, label: "The current plant" }]),
    ...others.map((choice) => ({ value: choice.worldId, label: `Run ${choice.title}` })),
  ];
}

function Picker({
  targets,
  value,
  onPick,
}: {
  readonly targets: readonly Target[];
  readonly value: string;
  readonly onPick: (value: string) => void;
}) {
  return (
    <label className={styles.picker}>
      Compare the model of this run with
      <select
        value={value}
        onChange={(event) => {
          onPick(event.target.value);
        }}
      >
        {targets.map((target) => (
          <option key={target.value} value={target.value}>
            {target.label}
          </option>
        ))}
      </select>
    </label>
  );
}

export function ModelDiffDialog({ provenance, workspace, report, worldId, onClose }: DialogProps) {
  const targets = useTargets(workspace, report, worldId);
  const [picked, setPicked] = useState<string | null>(null);
  const chosen = picked ?? targets[0]?.value ?? null;
  const dialog = useModal<HTMLDialogElement>(onClose);
  return (
    <div className={shared.scrim}>
      <dialog
        open
        ref={dialog}
        tabIndex={-1}
        className={cx(shared.dialog, compare.dialog)}
        aria-modal="true"
        aria-labelledby="model-diff-title"
      >
        <header className={compare.head}>
          <h2 id="model-diff-title" className={compare.title}>
            What changed in the model
          </h2>
          <button type="button" className={shared.secondary} aria-label="Close the model comparison" onClick={onClose}>
            <X size={14} aria-hidden="true" />
            Close
          </button>
        </header>
        <div className={compare.body}>
          {chosen === null ? (
            <p>There is no other version of this plant, or another run of it, to compare the model with.</p>
          ) : (
            <>
              <Picker targets={targets} value={chosen} onPick={setPicked} />
              <ModelDiffPanel session={provenance.diff} a={worldId} b={chosen} />
            </>
          )}
        </div>
      </dialog>
    </div>
  );
}

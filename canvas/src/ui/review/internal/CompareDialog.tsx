import { X } from "lucide-react";
import { useEffect } from "react";
import type { CompareState, ReviewSession } from "../../../app/review";
import { compareFor } from "../../../app/review";
import type { ProvenanceSession, WorkspaceStore } from "../../../app/workspace";
import type { Report } from "../../../model/review";
import { referenceChoices, referenceRole, thisRunRole } from "../../../model/review";
import { runRows } from "../../../model/workspace";
import { cx, failureLine, shared, useStore } from "../../primitives";
import { ComparisonView } from "./ComparisonView";
import { ModelLink, modelsDiffer } from "./ModelLink";
import { useModal } from "./useModal";
import styles from "./Compare.module.css";

interface DialogProps {
  readonly review: ReviewSession;
  readonly provenance: ProvenanceSession;
  readonly workspace: WorkspaceStore;
  readonly report: Report;
  readonly worldId: string;
  readonly unit: (entity: string, state: string) => string;
  readonly onClose: () => void;
}

function useChoices(workspace: WorkspaceStore, report: Report, worldId: string) {
  const listing = useStore(workspace);
  if (listing.phase !== "ready") return [];
  const runs = runRows(listing.workspace, report.domainId).map((row) => row.run);
  return referenceChoices(runs, { current: worldId, plant: report.domainId });
}

function Roles({
  reference,
  worldId,
  name,
}: {
  readonly reference: string | null;
  readonly worldId: string;
  readonly name: string;
}) {
  return (
    <dl className={styles.roles}>
      <div>
        <dt>{referenceRole}</dt>
        <dd>{reference ?? "choose below"}</dd>
      </div>
      <div>
        <dt>{thisRunRole}</dt>
        <dd>{name === "" ? worldId : `${name} · ${worldId}`}</dd>
      </div>
    </dl>
  );
}

function Choices({ props, state }: { readonly props: DialogProps; readonly state: CompareState }) {
  const choices = useChoices(props.workspace, props.report, props.worldId);
  if (state.phase !== "idle")
    return (
      <div>
        <button
          type="button"
          className={shared.secondary}
          onClick={() => {
            props.review.compare.clear();
          }}
        >
          Choose another reference
        </button>
      </div>
    );
  if (choices.length === 0) return <p className={styles.muted}>No other ended run of this plant yet.</p>;
  return (
    <ul className={styles.choices} aria-label="Runs to use as the reference">
      {choices.map((choice) => (
        <li key={choice.worldId}>
          <button
            type="button"
            className={cx(shared.secondary, styles.choice)}
            onClick={() => {
              void props.review.compare.choose(props.worldId, choice.worldId);
            }}
          >
            <span className={styles.choiceTitle}>{choice.title}</span>
            <span className={styles.choiceFacts}>{choice.facts}</span>
          </button>
        </li>
      ))}
    </ul>
  );
}

function Outcome({ props, state }: { readonly props: DialogProps; readonly state: CompareState }) {
  if (state.phase === "loading")
    return (
      <p role="status" className={styles.muted}>
        Comparing the two runs…
      </p>
    );
  if (state.phase === "failed")
    return (
      <p role="alert" className={styles.failure}>
        {failureLine(state.failure.diagnostic)}
      </p>
    );
  if (state.phase !== "ready") return null;
  return (
    <>
      {modelsDiffer(state.comparison) && (
        <ModelLink session={props.provenance.diff} reference={state.reference} worldId={props.worldId} />
      )}
      <ComparisonView comparison={state.comparison} unit={props.unit} />
    </>
  );
}

function DialogHead({ onClose }: { readonly onClose: () => void }) {
  return (
    <header className={styles.head}>
      <h2 id="compare-title" className={styles.title}>
        Compare with another run
      </h2>
      <button type="button" className={shared.secondary} aria-label="Close comparison" onClick={onClose}>
        <X size={14} aria-hidden="true" />
        Close
      </button>
    </header>
  );
}

export function CompareDialog(props: DialogProps) {
  const { review, report, worldId, onClose } = props;
  const state = compareFor(useStore(review.compare), worldId);
  const close = () => {
    review.compare.clear();
    onClose();
  };
  const dialog = useModal<HTMLDialogElement>(close);
  const choosing = state.phase === "idle";
  useEffect(() => {
    dialog.current?.focus();
  }, [choosing, dialog]);
  return (
    <div className={shared.scrim}>
      <dialog
        open
        ref={dialog}
        tabIndex={-1}
        className={cx(shared.dialog, styles.dialog)}
        aria-modal="true"
        aria-labelledby="compare-title"
      >
        <DialogHead onClose={close} />
        <div className={styles.body}>
          <Roles reference={state.phase === "idle" ? null : state.reference} worldId={worldId} name={report.name} />
          <Choices props={props} state={state} />
          <Outcome props={props} state={state} />
        </div>
      </dialog>
    </div>
  );
}

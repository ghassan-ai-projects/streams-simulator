import { Download, GitCompare, GitPullRequestArrow } from "lucide-react";
import { useState } from "react";
import type { ReviewSession } from "../../../app/review";
import { handoverFor, readyReport } from "../../../app/review";
import type { ReviewState } from "../../../app/review";
import type { ProvenanceSession, WorkspaceStore } from "../../../app/workspace";
import type { Report } from "../../../model/review";
import { endedGate } from "../../../model/review";
import { shared, useStore } from "../../primitives";
import { CompareDialog } from "./CompareDialog";
import { HandoverNote } from "./HandoverNote";
import { ModelDiffDialog } from "./ModelDiffDialog";

export interface ActionsProps {
  readonly review: ReviewSession;
  readonly workspace: WorkspaceStore;
  readonly provenance: ProvenanceSession;
  readonly modelOpen: boolean;
  readonly onModelClosed: () => void;
  readonly state: ReviewState;
  readonly worldId: string;
  readonly unit: (entity: string, state: string) => string;
}

function Why({ id, text }: { readonly id: string; readonly text: string }) {
  return (
    <span id={id} className="sr-only">
      {text}
    </span>
  );
}

function ExportHandover({
  review,
  report,
  worldId,
}: Pick<ActionsProps, "review" | "worldId"> & { readonly report: Report }) {
  const gate = endedGate(report);
  const handover = handoverFor(useStore(review.handover), worldId);
  const blocked = !gate.open || handover.phase === "working";
  return (
    <>
      <HandoverNote handover={handover} />
      <button
        type="button"
        className={shared.secondary}
        aria-disabled={blocked}
        aria-describedby={gate.open ? undefined : "handover-why"}
        title={gate.open ? "Download the binary's handover report as a Markdown file" : gate.why}
        onClick={() => {
          if (blocked) return;
          void review.handover.save(worldId, {
            replay: report.replay.status !== "not_checked",
            fileStem: report.runId === "" ? worldId : report.runId,
          });
        }}
      >
        <Download size={14} aria-hidden="true" />
        Export handover
      </button>
      {!gate.open && <Why id="handover-why" text={gate.why} />}
    </>
  );
}

function CompareButton({ report, onOpen }: { readonly report: Report; readonly onOpen: () => void }) {
  const gate = endedGate(report);
  return (
    <>
      <button
        type="button"
        className={shared.secondary}
        aria-disabled={!gate.open}
        aria-describedby={gate.open ? undefined : "compare-why"}
        title={gate.open ? "Compare this run with another ended run of the same plant" : gate.why}
        onClick={() => {
          if (gate.open) onOpen();
        }}
      >
        <GitCompare size={14} aria-hidden="true" />
        Compare with…
      </button>
      {!gate.open && <Why id="compare-why" text={gate.why} />}
    </>
  );
}

function ModelButton({ report, onOpen }: { readonly report: Report; readonly onOpen: () => void }) {
  const gate = endedGate(report);
  return (
    <button
      type="button"
      className={shared.secondary}
      aria-disabled={!gate.open}
      title={gate.open ? "Show what changed between the model this run used and the current plant" : gate.why}
      onClick={() => {
        if (gate.open) onOpen();
      }}
    >
      <GitPullRequestArrow size={14} aria-hidden="true" />
      Compare model with the current plant
    </button>
  );
}

type Open = "model" | "compare" | null;

function Dialogs({
  props,
  report,
  open,
  onClose,
}: {
  readonly props: ActionsProps;
  readonly report: Report;
  readonly open: Open;
  readonly onClose: () => void;
}) {
  const shared = { provenance: props.provenance, workspace: props.workspace, report, worldId: props.worldId };
  if (open === "model")
    return (
      <ModelDiffDialog
        {...shared}
        onClose={() => {
          onClose();
          props.onModelClosed();
        }}
      />
    );
  if (open === "compare")
    return <CompareDialog {...shared} review={props.review} unit={props.unit} onClose={onClose} />;
  return null;
}

export function ReviewActions(props: ActionsProps) {
  const [open, setOpen] = useState<Open>(props.modelOpen ? "model" : null);
  const report = readyReport(props.state);
  if (report === null) return null;
  return (
    <>
      <ModelButton
        report={report}
        onOpen={() => {
          setOpen("model");
        }}
      />
      <CompareButton
        report={report}
        onOpen={() => {
          setOpen("compare");
        }}
      />
      <ExportHandover review={props.review} report={report} worldId={props.worldId} />
      <Dialogs
        props={props}
        report={report}
        open={open}
        onClose={() => {
          setOpen(null);
        }}
      />
    </>
  );
}

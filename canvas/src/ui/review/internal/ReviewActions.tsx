import { Download, GitCompare } from "lucide-react";
import { useState } from "react";
import type { ReviewSession } from "../../../app/review";
import { handoverFor, readyReport } from "../../../app/review";
import type { ReviewState } from "../../../app/review";
import type { WorkspaceStore } from "../../../app/workspace";
import type { Report } from "../../../model/review";
import { endedGate } from "../../../model/review";
import { shared, useStore } from "../../primitives";
import { CompareDialog } from "./CompareDialog";

interface ActionsProps {
  readonly review: ReviewSession;
  readonly workspace: WorkspaceStore;
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

export function ReviewActions(props: ActionsProps) {
  const [comparing, setComparing] = useState(false);
  const report = readyReport(props.state);
  if (report === null) return null;
  return (
    <>
      <CompareButton
        report={report}
        onOpen={() => {
          setComparing(true);
        }}
      />
      <ExportHandover review={props.review} report={report} worldId={props.worldId} />
      {comparing && (
        <CompareDialog
          review={props.review}
          workspace={props.workspace}
          report={report}
          worldId={props.worldId}
          unit={props.unit}
          onClose={() => {
            setComparing(false);
          }}
        />
      )}
    </>
  );
}

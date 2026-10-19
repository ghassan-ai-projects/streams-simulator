import type { ReactNode } from "react";
import { ClipboardCheck, Hand, OctagonX } from "lucide-react";
import type { ReviewState } from "../../../app/review";
import type { CoveredDevice } from "../../../model/review";
import { interruptOf, modeLine } from "../../../model/review";
import parts from "./ReportParts.module.css";

function iconOf(covered: CoveredDevice): ReactNode {
  if (covered.latched) return <OctagonX size={13} aria-hidden="true" />;
  return covered.mode === "hold" ? (
    <Hand size={13} aria-hidden="true" />
  ) : (
    <ClipboardCheck size={13} aria-hidden="true" />
  );
}

export function InterruptLine({ state, entity }: { readonly state: ReviewState; readonly entity: string }) {
  const covered = interruptOf(state.live, entity);
  if (covered === null) return null;
  return (
    <p className={parts.origin} data-interrupt={covered.latched ? "latched" : covered.mode}>
      {iconOf(covered)}
      {modeLine(covered)}
    </p>
  );
}

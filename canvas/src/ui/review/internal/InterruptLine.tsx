import type { ReviewState } from "../../../app/review";
import { interruptOf } from "../../../model/review";
import { ModeBadge } from "./ModeBadge";
import parts from "./ReportParts.module.css";

export function InterruptLine({ state, entity }: { readonly state: ReviewState; readonly entity: string }) {
  const covered = interruptOf(state.live, entity);
  if (covered === null) return null;
  return (
    <p className={parts.origin} data-interrupt={covered.latched ? "latched" : covered.mode}>
      <ModeBadge mode={covered.mode} latched={covered.latched} size={13} />
    </p>
  );
}

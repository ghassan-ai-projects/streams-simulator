import type { ReviewSession, ReviewState } from "../../../app/review";
import { interruptOf } from "../../../model/review";
import { ModeBadge } from "./ModeBadge";
import { useShownLive } from "./useShownLive";
import parts from "./ReportParts.module.css";

interface LineProps {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly entity: string;
}

export function InterruptLine({ review, state, entity }: LineProps) {
  const covered = interruptOf(useShownLive(review, state), entity);
  if (covered === null) return null;
  return (
    <p className={parts.origin} data-interrupt={covered.latched ? "latched" : covered.mode}>
      <ModeBadge mode={covered.mode} latched={covered.latched} size={13} />
    </p>
  );
}

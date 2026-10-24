import type { ReadyPlant } from "../../../app/plant";
import type { ReviewSession, ReviewState } from "../../../app/review";
import { runGoing, startNsOf } from "../../../app/review";
import type { ObservedRow } from "../../../model/review";
import { observedRows, runClock } from "../../../model/review";
import { shared } from "../../primitives";
import { ObservedTable } from "./ObservedTable";

interface ObservedView {
  readonly rows: readonly ObservedRow[];
  readonly clock: string;
  readonly live: boolean;
}

export function observedView(
  review: ReviewSession,
  state: ReviewState,
  plant: ReadyPlant | null,
  entity: string,
): ObservedView {
  const live = runGoing(state);
  const following = state.following && (state.watching || state.connection === "running");
  const referenceNs = following ? Math.max(state.positionNs, review.index.clockNs) : state.positionNs;
  const startNs = startNsOf(state, review);
  const rows = observedRows({
    index: review.index,
    plant: plant?.plant ?? null,
    entity,
    atNs: state.positionNs,
    referenceNs,
    startNs,
    live,
  });
  const reached = `${following ? "Run clock" : "Cursor"} ${runClock(referenceNs, startNs)} · ages in simulated time`;
  return { rows, clock: reached, live };
}

export function ObservedReadings({
  review,
  state,
  plant,
  entity,
}: {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
  readonly entity: string;
}) {
  const view = observedView(review, state, plant, entity);
  const none = view.live ? "Awaiting first reading" : "Never observed in this run";
  return (
    <section aria-label="Observed readings">
      <h3 className={shared.sectionTitle}>Observed now</h3>
      {view.rows.length === 0 ? <p>{none}</p> : <ObservedTable rows={view.rows} clock={view.clock} />}
    </section>
  );
}

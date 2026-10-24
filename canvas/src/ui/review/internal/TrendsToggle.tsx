import { ChartSpline } from "lucide-react";
import type { ReadyPlant } from "../../../app/plant";
import type { ReviewSession } from "../../../app/review";
import { useStore } from "../../primitives";
import styles from "./Trends.module.css";

export function TrendsToggle({ review, plant }: { readonly review: ReviewSession; readonly plant: ReadyPlant | null }) {
  const watch = useStore(review.watchList);
  const count = watch.lists.get(plant?.plant.id ?? "")?.length ?? 0;
  return (
    <button
      type="button"
      className={styles.toggle}
      aria-pressed={watch.open}
      aria-label={count === 0 ? "Trends" : `Trends · ${String(count)}`}
      title="Trends: watched channels on one time axis"
      onClick={() => {
        review.watchList.toggleOpen();
      }}
    >
      <ChartSpline size={14} aria-hidden="true" />
      {count > 0 && <span aria-hidden="true">{count}</span>}
    </button>
  );
}

import type { ReadyPlant } from "../../../app/plant";
import type { ReviewSession, ReviewState } from "../../../app/review";
import { Navigator, Outline } from "../../graph";
import type { HallProps } from "./RunHeader";
import styles from "./ReviewScreen.module.css";

export function Side({
  review,
  state,
  plant,
  hall,
  onHall,
}: {
  readonly review: ReviewSession;
  readonly state: ReviewState;
  readonly plant: ReadyPlant | null;
} & HallProps) {
  return (
    <div className={styles.side}>
      {plant === null ? (
        <div className={styles.outlineSlot} />
      ) : (
        <Outline
          plant={plant.plant}
          hall={hall}
          selected={state.selected}
          onSelect={(id) => {
            review.select(id);
          }}
          onOpenHall={onHall}
          origins={state.origins}
        />
      )}
      <Navigator />
    </div>
  );
}

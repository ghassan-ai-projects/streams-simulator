import { ReactFlowProvider } from "@xyflow/react";
import { useState } from "react";
import type { LaunchStore } from "../../../app/launch";
import type { NavigationStore } from "../../../app/shell";
import type { PlantSession } from "../../../app/plant";
import type { ReviewSession } from "../../../app/review";
import type { WorkspaceStore } from "../../../app/workspace";
import { cx, shared, useStore } from "../../primitives";
import { Centre } from "./ReviewCentre";
import { Side } from "./ReviewSide";
import { Header } from "./RunHeader";
import { RunPanel } from "./RunPanel";
import { useReviewOpen } from "./useReviewOpen";
import { useClosePlant, useRunPlant, useUnitOf } from "./useRunPlant";
import { usePlaybackKeys } from "./usePlaybackKeys";
import { useShownOrigins } from "./useShownOrigins";
import styles from "./ReviewScreen.module.css";

interface ReviewScreenProps {
  readonly review: ReviewSession;
  readonly plants: PlantSession;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly launch: LaunchStore;
  readonly worldId: string;
  readonly live?: boolean;
}

export function ReviewScreen(props: ReviewScreenProps) {
  const { review, plants, workspace, navigation, launch, worldId, live = false } = props;
  const stored = useStore(review);
  const view = useRunPlant(stored, plants, workspace);
  const { plant } = view;
  const state = useShownOrigins(stored, plant);
  const [hall, setHall] = useState<string | null>(null);
  useReviewOpen(review, workspace, worldId, live ? "live" : "review");
  useClosePlant(plants, worldId);
  usePlaybackKeys(review);
  const unit = useUnitOf(plant);
  return (
    <ReactFlowProvider>
      <div className={cx(styles.screen, shared.panels)}>
        <Header
          state={state}
          navigation={navigation}
          launch={launch}
          worldId={worldId}
          hall={hall}
          onHall={setHall}
          plant={view}
        />
        <Side review={review} state={state} plant={plant} hall={hall} onHall={setHall} />
        <Centre
          review={review}
          state={state}
          plant={plant}
          hall={hall}
          onHall={setHall}
          worldId={worldId}
          navigation={navigation}
          now={() => launch.nowMs()}
        />
        <RunPanel review={review} state={state} plant={plant} unit={unit} />
      </div>
    </ReactFlowProvider>
  );
}

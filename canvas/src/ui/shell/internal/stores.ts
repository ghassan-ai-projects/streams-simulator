import type { DeviceDesk, DeviceSession } from "../../../app/devices";
import type { EditorSession } from "../../../app/editor";
import type { LaunchStore } from "../../../app/launch";
import type { PlantMaker, PlantSession } from "../../../app/plant";
import type { Scheduler } from "../../../app/ports";
import type { ReviewSession } from "../../../app/review";
import type { Session } from "../../../app/session";
import type { NavigationStore, PanelStore, ThemeStore } from "../../../app/shell";
import type { WorkspaceStore } from "../../../app/workspace";

export interface Stores {
  readonly themes: ThemeStore;
  readonly session: Session;
  readonly workspace: WorkspaceStore;
  readonly navigation: NavigationStore;
  readonly plants: PlantSession;
  readonly review: ReviewSession;
  readonly liveReview: ReviewSession;
  readonly watchReview: ReviewSession;
  readonly launch: LaunchStore;
  readonly scheduler: Scheduler;
  readonly editor: EditorSession;
  readonly devices: DeviceSession;
  readonly desk: DeviceDesk;
  readonly panels: PanelStore;
  readonly maker: PlantMaker;
  readonly showDetails: () => void;
}

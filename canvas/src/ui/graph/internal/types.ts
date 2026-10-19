import type {
  InstanceOrigin,
  NodeInterrupt,
  NodeStatus,
  PlantView,
  PlanMark,
  ReadoutBody,
  Shown,
} from "../../../model/graph";
import type { Plant } from "../../../model/plant";

export interface PlantGraphProps {
  readonly view: PlantView;
  readonly viewKey: string;
  readonly selected: string | null;
  readonly live?: (id: string) => NodeLive;
  readonly onSelect: (id: string | null) => void;
  readonly onOpenHall?: (hall: string) => void;
  readonly onEscape?: () => void;
  readonly anchor?: string | null;
  readonly focusZoom?: number;
  readonly flagged?: ReadonlySet<string>;
  readonly fitKey?: string;
  readonly panCue?: boolean;
  readonly notice?: string | null;
  readonly onConnect?: (from: string, to: string) => void;
  readonly marks?: ReadonlyMap<string, PlanMark>;
}

export interface OutlineProps {
  readonly flagged?: ReadonlySet<string>;
  readonly plant: Plant;
  readonly hall: string | null;
  readonly selected: string | null;
  readonly onSelect: (id: string) => void;
  readonly onOpenHall?: (hall: string) => void;
  readonly showEmptyAreas?: boolean;
  readonly origins?: ReadonlyMap<string, InstanceOrigin>;
}

export interface NodeLive {
  readonly shown: readonly Shown[];
  readonly observed: ReadonlySet<string>;
  readonly readout: ReadoutBody;
  readonly status: NodeStatus;
  readonly count: number;
  readonly fault: string;
  readonly effector: string;
  readonly origin: InstanceOrigin | null;
  readonly interrupt: NodeInterrupt | null;
}

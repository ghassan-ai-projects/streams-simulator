import type { Plant, PlantView, PlanMark, Shown, ReadoutBody, NodeStatus } from "../../../model/graph";

export interface PlantGraphProps {
  readonly view: PlantView;
  readonly viewKey: string;
  readonly selected: string | null;
  readonly live?: (id: string) => NodeLive;
  readonly onSelect: (id: string | null) => void;
  readonly onOpenHall?: (hall: string) => void;
  readonly onEscape?: () => void;
  readonly minFitZoom?: number;
  readonly anchor?: string | null;
  readonly focusZoom?: number;
  readonly flagged?: ReadonlySet<string>;
  readonly fitKey?: string;
  readonly panCue?: boolean;
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
}

export interface NodeLive {
  readonly shown: readonly Shown[];
  readonly observed: ReadonlySet<string>;
  readonly readout: ReadoutBody;
  readonly status: NodeStatus;
  readonly count: number;
  readonly fault: string;
  readonly effector: string;
}
